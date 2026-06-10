package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/skillsmanager/skillsmanager/backend/internal/agent"
	"github.com/skillsmanager/skillsmanager/backend/internal/clawhub"
	"github.com/skillsmanager/skillsmanager/backend/internal/config"
	"github.com/skillsmanager/skillsmanager/backend/internal/installer"
	"github.com/skillsmanager/skillsmanager/backend/internal/lifecycle"
	"github.com/skillsmanager/skillsmanager/backend/internal/registry"
	"github.com/skillsmanager/skillsmanager/backend/internal/skill"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// API 统一对外接口（Wails GUI + CLI 共用）
type API struct {
	cfg       *config.Config
	store     *storage.Storage
	reg       *registry.Registry
	agentMgr  *agent.Manager
	install   *installer.Installer
	lifecycle *lifecycle.Lifecycle
	clawhub   *clawhub.Manager
}

// New 创建 API
func New(configPath string) (*API, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	store, err := storage.New(cfg.Data.Skillspool.Root)
	if err != nil {
		return nil, err
	}
	reg, err := registry.Load(store.Root())
	if err != nil {
		return nil, err
	}
	agentMgr := agent.New(cfg.Data.Agents)
	inst := installer.New(store, reg, agentMgr)
	lc := lifecycle.New(store, reg, agentMgr)
	ch := clawhub.New(store.Root())

	return &API{
		cfg:       cfg,
		store:     store,
		reg:       reg,
		agentMgr:  agentMgr,
		install:   inst,
		lifecycle: lc,
		clawhub:   ch,
	}, nil
}

// Config 返回配置
func (a *API) Config() *models.Config {
	return a.cfg.Data
}

// SkillspoolRoot 返回 skillspool 根目录
func (a *API) SkillspoolRoot() string {
	return a.store.Root()
}

// ListSkills 列出所有已安装技能
func (a *API) ListSkills() []*models.Skill {
	return a.reg.List()
}

// GetSkill 获取单个技能
func (a *API) GetSkill(name string) (*models.Skill, bool) {
	return a.reg.Get(name)
}

// ListAgentIDs 返回已检测的 Agent ID 列表
func (a *API) ListAgentIDs() []string {
	return a.agentMgr.ListDetected()
}

// ListAgents 返回所有 Agent 配置
func (a *API) ListAgents() map[string]models.Agent {
	return a.agentMgr.List()
}

// ListAgentGroups 返回按目录分组的 Agent 列表
// scope: "global" 或 "project"
// 用于前端界面按目录组展示 Agent
func (a *API) ListAgentGroups(scope string) []models.AgentGroup {
	agents := a.agentMgr.List()

	// 按目录分组
	groupsMap := make(map[string]*models.AgentGroup)

	for id, ag := range agents {
		var dirKey, dirPath string

		switch scope {
		case "project":
			// 不支持项目级的跳过
			if !ag.SupportsProject {
				continue
			}
			dirKey = ag.ProjectDirectoryKey
			if dirKey == "" {
				dirKey = ag.SkillLocation // 兜底
			}
			dirPath = ag.SkillLocation
		default: // global
			dirKey = ag.GlobalDirectoryKey
			if dirKey == "" {
				dirKey = ag.GlobalLocation // 兜底
			}
			dirPath = ag.GlobalLocation
		}

		if dirKey == "" {
			// 没有目录信息的 Agent，单独作为一个组
			dirKey = "__" + id + "__"
			dirPath = id + "(未配置目录)"
		}

		g, ok := groupsMap[dirKey]
		if !ok {
			g = &models.AgentGroup{
				ID:          dirKey,
				Directory:   dirPath,
				Scope:       scope,
				AgentIDs:    []string{},
				AgentNames:  []string{},
				DetectedIDs: []string{},
			}
			groupsMap[dirKey] = g
		}
		g.AgentIDs = append(g.AgentIDs, id)
		g.AgentNames = append(g.AgentNames, ag.Name)
		if ag.Detected {
			g.DetectedIDs = append(g.DetectedIDs, id)
		}
	}

	// 构建结果，标记共享风险
	result := make([]models.AgentGroup, 0, len(groupsMap))
	for _, g := range groupsMap {
		g.SharedRisk = len(g.AgentIDs) >= 2
		result = append(result, *g)
	}
	return result
}

// InstallRequest 安装请求参数（方便 Wails 调用）
type InstallRequest struct {
	Source     string   `json:"source"`
	SubPath    string   `json:"sub_path,omitempty"`
	Version    string   `json:"version,omitempty"`
	Ref        string   `json:"ref,omitempty"`
	Agents     []string `json:"agents"`
	Scope      string   `json:"scope"` // "global" or "project"
	ProjectDir string   `json:"project_dir,omitempty"`
}

// InstallLink 是安装到单个 Agent 的链接结果
type InstallLink struct {
	AgentID string `json:"agent_id"`
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// InstallResult 是安装单个技能的完整结果
type InstallResult struct {
	SkillName string                 `json:"skill_name"`
	Version   string                 `json:"version"`
	Source    models.Source          `json:"source"`
	Agents    map[string]InstallLink `json:"agents"`
}

// Install 安装技能
func (a *API) Install(req InstallRequest) ([]InstallResult, error) {
	if len(req.Agents) == 0 {
		req.Agents = a.agentMgr.ListDetected()
	}
	if len(req.Agents) == 0 {
		return nil, fmt.Errorf("no agent detected, please configure agents first")
	}
	if req.Scope == "" {
		req.Scope = "global"
	}
	internalResults, err := a.install.Install(installer.InstallOptions{
		Source:     req.Source,
		SubPath:    req.SubPath,
		Version:    req.Version,
		Ref:        req.Ref,
		AgentIDs:   req.Agents,
		Scope:      req.Scope,
		ProjectDir: req.ProjectDir,
	})
	if err != nil {
		return nil, err
	}
	// 转换结果类型
	var results []InstallResult
	for _, ir := range internalResults {
		agentsMap := make(map[string]InstallLink)
		for agentID, link := range ir.Agents {
			agentsMap[agentID] = InstallLink{
				AgentID: link.AgentID,
				Path:    link.Path,
				Success: link.Success,
				Error:   link.Error,
			}
		}
		results = append(results, InstallResult{
			SkillName: ir.SkillName,
			Version:   ir.Version,
			Source:    ir.Source,
			Agents:    agentsMap,
		})
	}
	return results, nil
}

// SaveConfig 保存配置
func (a *API) SaveConfig() error {
	return a.cfg.Save()
}

// SetSkillspoolRoot 修改 skillspool 根目录并迁移所有技能。
// 流程：1) 验证新路径可写；2) 复制旧 skillspool 全部内容到新路径；3) 删除旧 skillspool；4) 更新配置。
// 失败时不会修改配置，可安全重试。
func (a *API) SetSkillspoolRoot(newRoot string) (*models.SkillspoolMigrationResult, error) {
	// 展开 ~ 和环境变量
	newRoot = expandPath(newRoot)
	if newRoot == "" {
		return nil, fmt.Errorf("新路径不能为空")
	}
	oldRoot := a.store.Root()
	// 归一化路径
	oldNorm := filepath.Clean(oldRoot)
	newNorm := filepath.Clean(newRoot)
	if oldNorm == newNorm {
		// 路径未变化
		_ = a.cfg.Data.Skillspool.Root
		return &models.SkillspoolMigrationResult{
			Success:    true,
			OldRoot:    oldRoot,
			NewRoot:    newRoot,
			MovedFiles: 0,
			Message:    "路径未变化",
		}, nil
	}

	// 1. 验证新路径父目录存在且可写
	parent := filepath.Dir(newRoot)
	if _, err := os.Stat(parent); err != nil {
		return nil, fmt.Errorf("新路径的父目录不存在: %w", err)
	}

	// 2. 如果新路径已存在且非空，校验是空目录或者就是当前 skillspool
	if info, err := os.Stat(newRoot); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("新路径已存在但不是目录: %s", newRoot)
		}
		// 检查是否非空
		entries, _ := os.ReadDir(newRoot)
		if len(entries) > 0 {
			return nil, fmt.Errorf("新路径已存在且非空: %s，请先清空或选择其他路径", newRoot)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("检查新路径失败: %w", err)
	}

	// 3. 创建新目录
	if err := os.MkdirAll(newRoot, 0o755); err != nil {
		return nil, fmt.Errorf("创建新 skillspool 失败: %w", err)
	}

	// 4. 复制旧 skillspool 内容到新路径
	moved, err := copyDirContents(oldRoot, newRoot)
	if err != nil {
		// 回滚：删除已创建的新目录
		_ = os.RemoveAll(newRoot)
		return nil, fmt.Errorf("复制失败: %w（已回滚）", err)
	}

	// 5. 删除旧 skillspool
	if err := os.RemoveAll(oldRoot); err != nil {
		// 复制成功但删除失败：保留新目录，配置改为新目录，旧目录残留
		// 仍然继续：更新配置以让用户可以使用新目录
		_ = a.cfg.Data.Skillspool.Root
		a.cfg.Data.Skillspool.Root = newRoot
		_ = a.cfg.Save()
		return &models.SkillspoolMigrationResult{
			Success:    true,
			OldRoot:    oldRoot,
			NewRoot:    newRoot,
			MovedFiles: moved,
			Message:    fmt.Sprintf("复制成功但删除旧目录失败：%v。请手动清理：%s", err, oldRoot),
		}, nil
	}

	// 6. 更新配置
	a.cfg.Data.Skillspool.Root = newRoot
	if err := a.cfg.Save(); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w（请检查配置权限）", err)
	}

	return &models.SkillspoolMigrationResult{
		Success:    true,
		OldRoot:    oldRoot,
		NewRoot:    newRoot,
		MovedFiles: moved,
		Message:    fmt.Sprintf("已迁移 %d 个文件/目录", moved),
	}, nil
}

// GetSkillspoolRoot 返回当前 skillspool 根目录
func (a *API) GetSkillspoolRoot() string {
	return a.store.Root()
}

// expandPath 展开 ~ 和环境变量
func expandPath(p string) string {
	p = os.ExpandEnv(p)
	if len(p) > 0 && p[0] == '~' {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[1:])
		}
	}
	return p
}

// copyDirContents 复制目录全部内容到目标位置，返回复制的文件/目录数量
func copyDirContents(src, dst string) (int, error) {
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if rel == "." {
			// 根目录本身
			return nil
		}
		if info.IsDir() {
			if err := os.MkdirAll(target, info.Mode()); err != nil {
				return err
			}
			count++
			return nil
		}
		// 处理软链接：直接复制内容
		if info.Mode()&os.ModeSymlink != 0 {
			// 读取软链接目标
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			// 创建新软链接
			if err := os.Symlink(linkTarget, target); err != nil {
				// Windows 下软链接可能需要权限，降级为文件复制
				return copyFile(path, target)
			}
			count++
			return nil
		}
		// 普通文件
		if err := copyFile(path, target); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	// 确保目标父目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// AddAgent 手动添加 Agent
func (a *API) AddAgent(id string, ag models.Agent) {
	a.cfg.AddAgent(id, ag)
	_ = a.cfg.Save()
	// 重新创建 Manager
	a.agentMgr = agent.New(a.cfg.Data.Agents)
	a.install = installer.New(a.store, a.reg, a.agentMgr)
	a.lifecycle = lifecycle.New(a.store, a.reg, a.agentMgr)
}

// --- 清理功能 API ---

// UninstallFromProject 从项目中移除技能（仅移除软链）
func (a *API) UninstallFromProject(skillName, agentID, projectPath string) error {
	return a.lifecycle.UninstallFromProject(skillName, agentID, projectPath)
}

// CleanOrphanedSymlinks 清理孤立软链
func (a *API) CleanOrphanedSymlinks(agentID string) (int, error) {
	return a.lifecycle.CleanOrphanedSymlinks(agentID)
}

// CleanGlobalLibrary 清理全局库
func (a *API) CleanGlobalLibrary(dryRun bool) (*models.CleanResult, error) {
	return a.lifecycle.CleanGlobalLibrary(dryRun)
}

// BatchClean 批量清理
func (a *API) BatchClean(criteria models.BatchCleanCriteria) (*models.BatchCleanResult, error) {
	// TODO: 实现批量清理逻辑
	result := &models.BatchCleanResult{
		DryRun: false,
		Results: []models.CleanItemResult{},
	}
	return result, nil
}

// --- 版本管理 API ---

// ListSkillVersions 列出技能版本
func (a *API) ListSkillVersions(skillName string) ([]models.VersionInfo, error) {
	return a.lifecycle.ListSkillVersions(skillName)
}

// SwitchSkillVersion 切换技能版本
func (a *API) SwitchSkillVersion(skillName, version string) error {
	return a.lifecycle.SwitchSkillVersion(skillName, version)
}

// DeleteSkillVersion 删除技能版本
func (a *API) DeleteSkillVersion(skillName, version string) error {
	return a.lifecycle.DeleteSkillVersion(skillName, version)
}

// --- 健康检查 API ---

// CheckHealth 执行健康检查
func (a *API) CheckHealth() (*models.HealthReport, error) {
	return a.lifecycle.CheckHealth()
}

// FixBrokenSymlinks 修复孤立软链
func (a *API) FixBrokenSymlinks(agentID string) (int, error) {
	return a.lifecycle.FixBrokenSymlinks(agentID)
}

// --- 统计 API ---

// GetSkillStats 获取技能统计
func (a *API) GetSkillStats(skillName string) (*models.SkillStats, error) {
	return a.lifecycle.GetSkillStats(skillName)
}

// GetAgentStats 获取 Agent 统计
func (a *API) GetAgentStats(agentID string) (*models.AgentStats, error) {
	return a.lifecycle.GetAgentStats(agentID)
}

// GetUsageDashboard 获取使用统计仪表盘
func (a *API) GetUsageDashboard() (*models.UsageDashboard, error) {
	return a.lifecycle.GetUsageDashboard()
}

// SearchMarket 搜索技能市场（简单版本：目前从 GitHub 搜索）
// 完整的市场搜索可以后续对接 trae skills market API
func (a *API) SearchMarket(keyword string) []MarketSkill {
	// 目前返回空，GUI 层会直接让用户粘贴 GitHub URL
	// TODO: 对接 trae 技能市场 API
	return nil
}

type MarketSkill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Stars       int      `json:"stars"`
	Tags        []string `json:"tags"`
}

// --- 新增：技能市场配置管理 ---

// SetMarketConfig 设置技能市场配置
func (a *API) SetMarketConfig(config models.SkillMarketConfig) error {
	a.cfg.Data.SkillMarket = config
	return a.cfg.Save()
}

// --- 新增：技能市场扫描与获取 ---

// ScanMarket 扫描技能市场
func (a *API) ScanMarket() (*models.ScanMarketResult, error) {
	// 简单实现：扫描配置的市场 URL
	if a.cfg.Data.SkillMarket.URL == "" {
		return &models.ScanMarketResult{
			TotalSkills: 0,
			Categories:  []string{},
			Skills:      []models.MarketSkill{},
		}, nil
	}
	// TODO: 完整的市场扫描实现
	return &models.ScanMarketResult{
		TotalSkills: 0,
		Categories:  []string{},
		Skills:      []models.MarketSkill{},
	}, nil
}

// ListMarketSkills 列出市场技能
func (a *API) ListMarketSkills(category string) ([]models.MarketSkill, error) {
	// TODO: 完整实现
	return []models.MarketSkill{}, nil
}

// --- 新增：全局技能库扫描 ---

// ListGlobalSkillsWithAgents 扫描并列出所有 Agent 的全局技能
func (a *API) ListGlobalSkillsWithAgents() ([]models.GlobalSkillWithAgents, error) {
	// 简单实现框架
	skillMap := make(map[string]*models.GlobalSkillWithAgents)
	agents := a.agentMgr.List()

	for _, agent := range agents {
		if !agent.Installed {
			continue
		}
		// TODO: 扫描 agent.GlobalLocation 中的技能目录
	}

	var result []models.GlobalSkillWithAgents
	for _, skill := range skillMap {
		result = append(result, *skill)
	}
	return result, nil
}

// --- 新增：项目技能管理 ---

// MigrateProjectSkillToLibrary 迁移项目技能到全局 skillspool 库
// skillPath: 项目中某个技能目录的绝对路径（该目录下有 SKILL.md）
func (a *API) MigrateProjectSkillToLibrary(skillPath string, projectPath string) (*models.MigrateResult, error) {
	agentIDs := a.agentMgr.ListDetected()
	targetPath, err := a.install.MigrateProjectSkill(skillPath, agentIDs)
	if err != nil {
		return &models.MigrateResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &models.MigrateResult{
		Success:     true,
		LibraryPath: targetPath,
	}, nil
}

// --- 新增：按状态/标签列出技能（带安装状态计算） ---

// ListSkillsWithStatus 返回带安装状态的技能列表（按已安装优先排序）
func (a *API) ListSkillsWithStatus() []models.SkillWithStatus {
	detectedAgents := a.agentMgr.ListDetected()
	totalAgents := len(detectedAgents)

	skills := a.reg.List()
	result := make([]models.SkillWithStatus, 0, len(skills))
	installedPool := make([]models.SkillWithStatus, 0)
	otherPool := make([]models.SkillWithStatus, 0)

	for _, s := range skills {
		agents := a.reg.CollectAgents(s.Name)
		installedCount := len(agents)
		var status models.SkillInstallStatus
		if installedCount == 0 {
			status = "not_installed"
		} else if installedCount >= totalAgents && totalAgents > 0 {
			status = "installed"
		} else {
			status = "partially_installed"
		}

		// 合并 SKILL.md 标签与用户自定义标签
		allTags := make([]string, 0, len(s.UserTags))
		seen := make(map[string]bool)
		for _, t := range s.UserTags {
			if !seen[t] {
				seen[t] = true
				allTags = append(allTags, t)
			}
		}

		item := models.SkillWithStatus{
			Name:            s.Name,
			Description:     s.Description,
			Tags:            allTags,
			LatestVersion:   s.LatestVersion,
			InstallStatus:   status,
			InstalledAgents: agents,
			TotalAgents:     totalAgents,
			Source:          s.Source,
		}
		if status == "installed" {
			installedPool = append(installedPool, item)
		} else {
			otherPool = append(otherPool, item)
		}
	}
	// 已安装优先，按名称排序
	sortByName := func(items []models.SkillWithStatus) {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
	}
	sortByName(installedPool)
	sortByName(otherPool)
	result = append(result, installedPool...)
	result = append(result, otherPool...)
	return result
}

// --- 新增：标签管理 API ---

// AddSkillTag 给技能添加标签
func (a *API) AddSkillTag(skillName, tag string) bool {
	return a.reg.AddUserTag(skillName, tag)
}

// RemoveSkillTag 移除技能标签
func (a *API) RemoveSkillTag(skillName, tag string) bool {
	return a.reg.RemoveUserTag(skillName, tag)
}

// GetSkillTags 获取技能的标签
func (a *API) GetSkillTags(skillName string) []string {
	return a.reg.GetUserTags(skillName)
}

// GetAllTags 获取所有标签使用情况
func (a *API) GetAllTags() []models.TagUsage {
	return a.reg.GetAllTags()
}

// --- 新增：单技能安装/卸载到单个 Agent ---

// InstallSkillToAgent 将 skillspool 中的技能安装到指定 Agent（目录复制）
func (a *API) InstallSkillToAgent(skillName, agentID string) (bool, error) {
	_, err := a.install.InstallSkillToAgent(skillName, agentID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UninstallSkillFromAgent 从指定 Agent 目录移除技能
func (a *API) UninstallSkillFromAgent(skillName, agentID string) (bool, error) {
	err := a.install.UninstallSkillFromAgent(skillName, agentID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// --- 新增：项目技能扫描 ---

// ScanProjectSkills 扫描项目目录下所有 Agent 子目录的技能
// projectPath 可以是单个项目目录路径
func (a *API) ScanProjectSkills(projectPath string) []models.ProjectSkill {
	if projectPath == "" {
		return nil
	}
	agentMap := a.agentMgr.List()
	seen := make(map[string]bool)
	var result []models.ProjectSkill

	// 遍历所有已知 Agent 的 skill_location
	for agentID, ag := range agentMap {
		// 扫描项目中该 Agent 的技能目录
		skillDir := filepath.Join(projectPath, ag.SkillLocation)
		entries, err := os.ReadDir(skillDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			skillDirPath := filepath.Join(skillDir, e.Name())
			if _, err := os.Stat(filepath.Join(skillDirPath, "SKILL.md")); err != nil {
				continue
			}
			// 同一技能可能在多个 Agent 目录都出现，只保留一次
			key := e.Name()
			if seen[key] {
				continue
			}
			seen[key] = true

			info, _ := skill.ParseSkillFile(filepath.Join(skillDirPath, "SKILL.md"))
			desc := ""
			ver := ""
			var tags []string
			if info != nil {
				desc = info.Description
				ver = info.Version
				tags = info.Tags
			}

			// 检查是否已存在于 skillspool
			inLibrary := false
			if _, ok := a.reg.Get(key); ok {
				inLibrary = true
			}

			// 计算目录大小
			size := dirSize(skillDirPath)

			result = append(result, models.ProjectSkill{
				Name:        key,
				Description: desc,
				Path:        skillDirPath,
				IsSymlink:   isSymlink(skillDirPath),
				InLibrary:   inLibrary,
				Version:     ver,
				Tags:        tags,
				SizeBytes:   size,
			})
		}
		_ = agentID // 变量可能不使用，避免未使用告警
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].InLibrary != result[j].InLibrary {
			// 不在库中的优先展示（更需要迁移）
			return !result[i].InLibrary && result[j].InLibrary
		}
		return result[i].Name < result[j].Name
	})
	return result
}

// --- 新增：批量同步 ---

// BatchSyncSkills 批量同步技能到指定 Agent
func (a *API) BatchSyncSkills(req models.BatchSyncRequest) (*models.BatchSyncResult, error) {
	result := &models.BatchSyncResult{
		Total:   len(req.SkillNames) * len(req.AgentIDs),
		Results: []models.BatchSyncItemResult{},
	}

	for _, skillName := range req.SkillNames {
		for _, agentID := range req.AgentIDs {
			_, err := a.install.InstallSkillToAgent(skillName, agentID)
			item := models.BatchSyncItemResult{
				SkillName: skillName,
				AgentID:   agentID,
				Success:   err == nil,
			}
			if err != nil {
				item.Error = err.Error()
				result.Failed++
			} else {
				result.Succeeded++
			}
			result.Results = append(result.Results, item)
		}
	}
	return result, nil
}

// --- 辅助函数 ---

func dirSize(path string) int64 {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// ---------- ClawHub 相关 API ----------

// RuntimeStatus 返回 Node.js / clawhub CLI 的运行时状态
func (a *API) RuntimeStatus() models.RuntimeStatus {
	return a.clawhub.RuntimeStatus()
}

// EnsureRuntime 确保 Node.js / clawhub CLI 可用；若缺少 clawhub 则自动安装
func (a *API) EnsureRuntime() (*models.RuntimeStatus, error) {
	return a.clawhub.EnsureRuntime()
}

// SearchClawHub 通过 clawhub CLI 搜索技能
func (a *API) SearchClawHub(keyword string) []models.ClawHubSkill {
	skills, err := a.clawhub.Search(keyword)
	if err != nil {
		return nil
	}
	return skills
}

// InstallFromClawHub 从 ClawHub 市场安装一个技能到 skillspool，并分发到指定 Agent。
// owner: 技能发布者，slug: 技能名，agentIDs: 目标 Agent 列表（空则全部已检测的）。
func (a *API) InstallFromClawHub(owner, slug string, agentIDs []string) (*InstallResult, error) {
	if owner == "" || slug == "" {
		return nil, fmt.Errorf("owner and slug must not be empty")
	}
	if len(agentIDs) == 0 {
		agentIDs = a.agentMgr.ListDetected()
	}
	sourceRef := "clawhub:" + owner + "/" + slug
	internalResults, err := a.install.Install(installer.InstallOptions{
		Source:   sourceRef,
		AgentIDs: agentIDs,
		Scope:    "global",
	})
	if err != nil {
		return nil, err
	}
	if len(internalResults) == 0 {
		return nil, fmt.Errorf("no skill installed from clawhub")
	}
	// 转换第一个结果（通常只有一个）
	ir := internalResults[0]
	agentsMap := make(map[string]InstallLink)
	for agentID, link := range ir.Agents {
		agentsMap[agentID] = InstallLink{
			AgentID: link.AgentID,
			Path:    link.Path,
			Success: link.Success,
			Error:   link.Error,
		}
	}
	return &InstallResult{
		SkillName: ir.SkillName,
		Version:   ir.Version,
		Source:    ir.Source,
		Agents:    agentsMap,
	}, nil
}
