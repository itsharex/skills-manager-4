package lifecycle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/agent"
	"github.com/skillsmanager/skillsmanager/backend/internal/registry"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Lifecycle 封装生命周期管理操作
type Lifecycle struct {
	store    *storage.Storage
	reg      *registry.Registry
	agentMgr *agent.Manager
}

// New 创建 Lifecycle 管理器
func New(store *storage.Storage, reg *registry.Registry, agentMgr *agent.Manager) *Lifecycle {
	return &Lifecycle{
		store:    store,
		reg:      reg,
		agentMgr: agentMgr,
	}
}

// --- 清理操作 ---

// UninstallFromProject 从项目中移除技能（仅移除软链，不删除全局库）
func (l *Lifecycle) UninstallFromProject(skillName, agentID, projectPath string) error {
	// 构建项目中的技能软链路径
	skillLink := filepath.Join(projectPath, ".agent", "skills", skillName)
	
	// 检查软链是否存在
	info, err := os.Lstat(skillLink)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q is not installed in project %q", skillName, projectPath)
		}
		return fmt.Errorf("lstat %q: %w", skillLink, err)
	}
	
	// 确认是软链
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%q is not a symlink", skillLink)
	}
	
	// 删除软链
	if err := os.Remove(skillLink); err != nil {
		return fmt.Errorf("remove symlink: %w", err)
	}
	
	return nil
}

// CleanOrphanedSymlinks 清理指定 Agent 的孤立软链
func (l *Lifecycle) CleanOrphanedSymlinks(agentID string) (int, error) {
	var skillDir string
	
	if agentID == "" {
		// 清理所有已检测的 Agent
		cleaned := 0
		for _, ag := range l.agentMgr.List() {
			if !ag.Detected {
				continue
			}
			dir := filepath.Join(ag.GlobalLocation, "skills")
			c, err := l.cleanDirOrphanedSymlinks(dir, agentID)
			if err != nil {
				return cleaned, err
			}
			cleaned += c
		}
		return cleaned, nil
	}
	
	ag, ok := l.agentMgr.List()[agentID]
	if !ok {
		return 0, fmt.Errorf("unknown agent: %s", agentID)
	}
	
	skillDir = filepath.Join(ag.GlobalLocation, "skills")
	return l.cleanDirOrphanedSymlinks(skillDir, agentID)
}

func (l *Lifecycle) cleanDirOrphanedSymlinks(dir, agentID string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("readdir %q: %w", dir, err)
	}
	
	cleaned := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// 只处理软链
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		
		// 检查软链目标是否存在
		target, err := os.Readlink(path)
		if err != nil {
			continue
		}
		
		// 解析相对路径
		if !filepath.IsAbs(target) {
			target = filepath.Join(dir, target)
		}
		
		if _, err := os.Stat(target); err != nil {
			// 目标不存在，删除软链
			if err := os.Remove(path); err != nil {
				continue
			}
			cleaned++
		}
	}
	
	return cleaned, nil
}

// CleanGlobalLibrary 清理全局库中未使用的技能和版本
func (l *Lifecycle) CleanGlobalLibrary(dryRun bool) (*models.CleanResult, error) {
	result := &models.CleanResult{
		Items: []models.CleanItemResult{},
	}

	// 1. 找出未安装到任何 Agent 的技能
	for _, skill := range l.reg.List() {
		if len(skill.Versions) == 0 {
			continue
		}

		// 检查是否安装到任何 Agent
		installedAgents := 0
		for _, v := range skill.Versions {
			installedAgents += len(v.Agents)
		}

		if installedAgents == 0 {
			// 找出最新版本用于展示
			latest := skill.LatestVersion
			if latest == "" {
				for _, v := range skill.Versions {
					latest = v.Version
					break
				}
			}

			if dryRun {
				result.Items = append(result.Items, models.CleanItemResult{
					SkillName: skill.Name,
					Version:   latest,
					Action:   "unused_skill",
					Success:  true,
				})
			} else {
				// 删除所有版本
				for version, ver := range skill.Versions {
					if err := os.RemoveAll(ver.Path); err != nil {
						result.Items = append(result.Items, models.CleanItemResult{
							SkillName: skill.Name,
							Version:   version,
							Action:   "deleted",
							Success:  false,
							Error:    err.Error(),
						})
						result.Failed++
					} else {
						result.Items = append(result.Items, models.CleanItemResult{
							SkillName: skill.Name,
							Version:   version,
							Action:   "deleted",
							Success:  true,
						})
						result.Succeeded++
					}
					result.TotalProcessed++
				}
				// 从注册表中移除
				l.reg.Remove(skill.Name, "")
			}
		}
	}

	if !dryRun {
		l.reg.Save()
	}

	return result, nil
}

// --- 版本管理 ---

// ListSkillVersions 列出技能的所有版本
func (l *Lifecycle) ListSkillVersions(skillName string) ([]models.VersionInfo, error) {
	skill, ok := l.reg.Get(skillName)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}
	
	var versions []models.VersionInfo
	for version, ver := range skill.Versions {
		// 获取版本目录大小
		size, _ := l.dirSize(ver.Path)
		
		versions = append(versions, models.VersionInfo{
			Version:    version,
			Installed:  parseTime(ver.InstalledAt),
			SizeBytes: size,
			Source:     skill.Source.URL,
			IsLatest:   version == skill.LatestVersion,
			AgentCount: len(ver.Agents),
		})
	}
	
	// 按版本排序（新版在前）
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Installed.After(versions[j].Installed)
	})
	
	return versions, nil
}

// SwitchSkillVersion 切换技能版本
func (l *Lifecycle) SwitchSkillVersion(skillName, version string) error {
	skill, ok := l.reg.Get(skillName)
	if !ok {
		return fmt.Errorf("skill not found: %s", skillName)
	}

	targetVer, ok := skill.Versions[version]
	if !ok {
		return fmt.Errorf("version not found: %s", version)
	}

	// 1. 更新所有 Agent 的软链：先移除旧版本，再安装新版本
	for _, agentID := range targetVer.Agents {
		// 获取当前版本路径
		currentVersion := ""
		for v, ver := range skill.Versions {
			if v == version {
				continue
			}
			for _, aid := range ver.Agents {
				if aid == agentID {
					currentVersion = v
					break
				}
			}
			if currentVersion != "" {
				break
			}
		}

		// 移除旧版本软链
		if currentVersion != "" {
			if err := l.agentMgr.RemoveFromAgent(skillName, agentID, "global", ""); err != nil {
				// 记录错误但继续
				fmt.Fprintf(os.Stderr, "warn: remove old version for agent %s: %v\n", agentID, err)
			}
		}

		// 安装新版本软链
		if _, err := l.agentMgr.InstallToAgent(skillName, targetVer.Path, agentID, "global", ""); err != nil {
			// 记录错误但继续
			fmt.Fprintf(os.Stderr, "warn: install new version for agent %s: %v\n", agentID, err)
		}
	}

	// 2. 更新 latest 文件
	latestPath := filepath.Join(l.store.Root(), skillName, "latest")
	if err := os.WriteFile(latestPath, []byte(version), 0644); err != nil {
		return fmt.Errorf("update latest: %w", err)
	}

	// 3. 更新注册表
	skill.LatestVersion = version
	if err := l.reg.Save(); err != nil {
		return fmt.Errorf("save registry: %w", err)
	}

	return nil
}

// DeleteSkillVersion 删除技能版本
func (l *Lifecycle) DeleteSkillVersion(skillName, version string) error {
	skill, ok := l.reg.Get(skillName)
	if !ok {
		return fmt.Errorf("skill not found: %s", skillName)
	}
	
	if len(skill.Versions) <= 1 {
		return fmt.Errorf("cannot delete the only version of skill %q", skillName)
	}
	
	targetVer, ok := skill.Versions[version]
	if !ok {
		return fmt.Errorf("version not found: %s", version)
	}
	
	// 如果有 Agent 正在使用此版本，先切换到 latest
	if len(targetVer.Agents) > 0 {
		// 找到另一个版本
		var alternateVersion string
		for v := range skill.Versions {
			if v != version {
				alternateVersion = v
				break
			}
		}
		if alternateVersion != "" {
			if err := l.SwitchSkillVersion(skillName, alternateVersion); err != nil {
				return fmt.Errorf("switch away from version: %w", err)
			}
		}
	}
	
	// 删除版本目录
	if err := os.RemoveAll(targetVer.Path); err != nil {
		return fmt.Errorf("remove version dir: %w", err)
	}
	
	// 从注册表中移除
	delete(skill.Versions, version)
	if err := l.reg.Save(); err != nil {
		return fmt.Errorf("save registry: %w", err)
	}
	
	return nil
}

// --- 健康检查 ---

// CheckHealth 执行完整健康检查
func (l *Lifecycle) CheckHealth() (*models.HealthReport, error) {
	report := &models.HealthReport{
		GeneratedAt: time.Now(),
		Status:      "healthy",
		Summary:     models.HealthSummary{},
		Issues:      []models.HealthIssue{},
		Symlinks:    []models.SymlinkIssue{},
		Files:       []models.FileIssue{},
		Versions:    []models.VersionIssue{},
	}
	
	// 统计
	report.Summary.TotalSkills = len(l.reg.List())
	agents := l.agentMgr.List()
	for _, ag := range agents {
		if ag.Detected {
			report.Summary.TotalAgents++
		}
	}
	
	// 1. 检查孤立软链
	for agentID, ag := range agents {
		if !ag.Detected {
			continue
		}
		skillDir := filepath.Join(ag.GlobalLocation, "skills")
		entries, err := os.ReadDir(skillDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			path := filepath.Join(skillDir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			if info.Mode()&os.ModeSymlink == 0 {
				continue
			}
			
			target, _ := os.Readlink(path)
			if !filepath.IsAbs(target) {
				target = filepath.Join(skillDir, target)
			}
			
			if _, err := os.Stat(target); err != nil {
				report.Summary.BrokenSymlinks++
				report.Symlinks = append(report.Symlinks, models.SymlinkIssue{
					AgentID:     agentID,
					Path:        path,
					Target:      target,
					TargetExists: false,
				})
				report.Issues = append(report.Issues, models.HealthIssue{
					Type:        "broken_symlink",
					Severity:    "error",
					AgentID:     agentID,
					Path:        path,
					Message:     fmt.Sprintf("broken symlink: %s -> %s", entry.Name(), target),
					Remediation: fmt.Sprintf("Run 'skills clean --agent %s' to remove", agentID),
				})
			}
		}
	}
	
	// 2. 检查不可达的技能（注册表中有但目录不存在）
	for _, skill := range l.reg.List() {
		if len(skill.Versions) == 0 {
			continue
		}

		skillPath := filepath.Join(l.store.Root(), skill.Name)
		if _, err := os.Stat(skillPath); err != nil {
			report.Summary.UnreachableSkills++
			report.Issues = append(report.Issues, models.HealthIssue{
				Type:        "unreachable",
				Severity:    "error",
				SkillName:   skill.Name,
				Path:        skillPath,
				Message:     fmt.Sprintf("skill %q in registry but directory not found", skill.Name),
				Remediation: "Reinstall the skill or remove from registry",
			})
		}
		
		// 3. 检查 latest 指针是否有效
		latestPath := filepath.Join(skillPath, "latest")
		if data, err := os.ReadFile(latestPath); err == nil {
			latestVer := strings.TrimSpace(string(data))
			if _, ok := skill.Versions[latestVer]; !ok {
				report.Versions = append(report.Versions, models.VersionIssue{
					SkillName: skill.Name,
					Version:   latestVer,
					Issue:     "latest_broken",
				})
				report.Issues = append(report.Issues, models.HealthIssue{
					Type:        "broken_latest",
					Severity:    "warning",
					SkillName:   skill.Name,
					Path:        latestPath,
					Message:     fmt.Sprintf("latest points to non-existent version: %s", latestVer),
					Remediation: fmt.Sprintf("Run 'skills version fix-latest %s'", skill.Name),
				})
			}
		}
	}
	
	// 确定总体状态
	if report.Summary.BrokenSymlinks > 0 || report.Summary.UnreachableSkills > 0 {
		report.Status = "error"
	} else if len(report.Issues) > 0 {
		report.Status = "warning"
	}
	
	return report, nil
}

// FixBrokenSymlinks 修复孤立软链
func (l *Lifecycle) FixBrokenSymlinks(agentID string) (int, error) {
	return l.CleanOrphanedSymlinks(agentID)
}

// --- 统计 ---

// GetSkillStats 获取技能统计
func (l *Lifecycle) GetSkillStats(skillName string) (*models.SkillStats, error) {
	skill, ok := l.reg.Get(skillName)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}
	
	stats := &models.SkillStats{
		Name:         skillName,
		VersionCount: len(skill.Versions),
		InstalledBy:  []models.AgentInstall{},
	}
	
	// 找出最新版本
	stats.CurrentVersion = skill.LatestVersion
	if stats.CurrentVersion == "" {
		for _, v := range skill.Versions {
			stats.CurrentVersion = v.Version
			break
		}
	}
	
	// 计算总大小
	var totalSize int64
	for _, v := range skill.Versions {
		size, _ := l.dirSize(v.Path)
		totalSize += size
	}
	stats.SizeBytes = totalSize
	
	// 收集安装信息
	seen := make(map[string]bool)
	for _, v := range skill.Versions {
		for _, agentID := range v.Agents {
			if seen[agentID] {
				continue
			}
			seen[agentID] = true
			stats.InstalledBy = append(stats.InstalledBy, models.AgentInstall{
				AgentID: agentID,
				Version: v.Version,
			})
		}
	}
	
	return stats, nil
}

// GetAgentStats 获取 Agent 统计
func (l *Lifecycle) GetAgentStats(agentID string) (*models.AgentStats, error) {
	ag, ok := l.agentMgr.List()[agentID]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	
	stats := &models.AgentStats{
		ID:               agentID,
		Name:             ag.Name,
		InstalledSkills:  []models.SkillInstall{},
	}
	
	skillDir := filepath.Join(ag.GlobalLocation, "skills")
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return nil, err
	}
	
	orphaned := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(skillDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// 检查是否是软链且目标存在
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(path)
			if !filepath.IsAbs(target) {
				target = filepath.Join(skillDir, target)
			}
			if _, err := os.Stat(target); err != nil {
				orphaned++
				continue
			}
		}
		
		stats.SkillCount++
		
		// 尝试获取版本信息
		version := ""
		isLatest := false
		if skill, ok := l.reg.Get(entry.Name()); ok {
			version = skill.LatestVersion
			isLatest = true
		}
		
		stats.InstalledSkills = append(stats.InstalledSkills, models.SkillInstall{
			SkillName: entry.Name(),
			Version:   version,
			IsLatest:  isLatest,
			Status:    "current",
		})
	}
	
	stats.OrphanedCount = orphaned
	
	return stats, nil
}

// GetUsageDashboard 获取使用统计仪表盘
func (l *Lifecycle) GetUsageDashboard() (*models.UsageDashboard, error) {
	dashboard := &models.UsageDashboard{}
	
	// 统计技能
	skills := l.reg.List()
	dashboard.TotalSkills = len(skills)

	// 统计安装总数和大小
	var totalInstallations int
	var totalSize int64
	agentCounts := make(map[string]int) // skillName -> agentCount

	for _, skill := range skills {
		for _, v := range skill.Versions {
			totalInstallations += len(v.Agents)
			size, _ := l.dirSize(v.Path)
			totalSize += size
		}
		agentCounts[skill.Name] = 0
		for _, v := range skill.Versions {
			agentCounts[skill.Name] += len(v.Agents)
		}
	}
	
	dashboard.TotalInstallations = totalInstallations
	dashboard.TotalSizeBytes = totalSize
	
	// 计算平均每 Agent
	agentCount := 0
	for _, ag := range l.agentMgr.List() {
		if ag.Detected {
			agentCount++
		}
	}
	if agentCount > 0 {
		dashboard.AveragePerAgent = float64(totalInstallations) / float64(agentCount)
	}
	
	// 最常用技能
	type scoredSkill struct {
		name  string
		count int
	}
	var scored []scoredSkill
	for name, count := range agentCounts {
		scored = append(scored, scoredSkill{name, count})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].count > scored[j].count
	})
	
	// 取前5和后5
	if len(scored) > 5 {
		for i := 0; i < 5; i++ {
			dashboard.MostPopular = append(dashboard.MostPopular, models.SkillCount{
				SkillName: scored[i].name,
				Count:     scored[i].count,
			})
		}
		for i := len(scored) - 5; i < len(scored); i++ {
			dashboard.LeastUsed = append(dashboard.LeastUsed, models.SkillCount{
				SkillName: scored[i].name,
				Count:     scored[i].count,
			})
		}
	} else {
		dashboard.MostPopular = make([]models.SkillCount, 0)
		dashboard.LeastUsed = make([]models.SkillCount, 0)
		for _, s := range scored {
			dashboard.MostPopular = append(dashboard.MostPopular, models.SkillCount{
				SkillName: s.name,
				Count:     s.count,
			})
			dashboard.LeastUsed = append(dashboard.LeastUsed, models.SkillCount{
				SkillName: s.name,
				Count:     s.count,
			})
		}
	}
	
	return dashboard, nil
}

// --- 辅助函数 ---

func (l *Lifecycle) dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
