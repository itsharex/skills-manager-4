package installer

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/agent"
	"github.com/skillsmanager/skillsmanager/backend/internal/registry"
	"github.com/skillsmanager/skillsmanager/backend/internal/skill"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/internal/version"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Installer 封装完整的安装流程
type Installer struct {
	store    *storage.Storage
	reg      *registry.Registry
	agentMgr *agent.Manager
}

// New 创建 Installer
func New(store *storage.Storage, reg *registry.Registry, agentMgr *agent.Manager) *Installer {
	return &Installer{store: store, reg: reg, agentMgr: agentMgr}
}

// InstallOptions 安装选项
type InstallOptions struct {
	Source     string   // 来源: GitHub URL / npx 命令 / 本地路径
	SubPath    string   // 仓库中子目录（可选）
	Version    string   // 指定版本（可选）
	Ref        string   // git ref（分支/标签/commit）
	AgentIDs   []string // 目标 Agent
	Scope      string   // "global" 或 "project"
	ProjectDir string   // 项目目录（scope == "project" 时）
}

// InstallResult 安装结果
type InstallResult struct {
	SkillName string                 `json:"skill_name"`
	Version   string                 `json:"version"`
	Source    models.Source          `json:"source"`
	Agents    map[string]InstallLink `json:"agents"`
}

type InstallLink struct {
	AgentID string `json:"agent_id"`
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Install 执行完整的安装流程
// 返回安装的技能列表
func (i *Installer) Install(opts InstallOptions) ([]InstallResult, error) {
	// 1. 解析来源，获取本地临时目录
	srcType, localPath, cleanup, err := i.resolveSource(opts)
	if err != nil {
		return nil, fmt.Errorf("resolve source: %w", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// 2. 扫描 SKILL.md
	skills, err := skill.ScanDirectory(localPath)
	if err != nil {
		return nil, fmt.Errorf("scan skills: %w", err)
	}
	if len(skills) == 0 {
		return nil, fmt.Errorf("no SKILL.md found in %s", localPath)
	}

	// 3. 安装每个技能
	var results []InstallResult
	for skillPath, info := range skills {
		result, err := i.installOne(skillPath, info, opts, srcType, localPath)
		if err != nil {
			return results, fmt.Errorf("install %q: %w", info.Name, err)
		}
		results = append(results, result)
	}

	// 4. 保存注册表
	if err := i.reg.Save(); err != nil {
		return results, fmt.Errorf("save registry: %w", err)
	}

	return results, nil
}

// --- 内部 ---

func (i *Installer) installOne(skillPath string, info *models.SkillInfo, opts InstallOptions, srcType models.SourceType, repoPath string) (InstallResult, error) {
	// 确定技能名（可能包含子目录层次）
	skillName := info.Name
	if opts.SubPath != "" && srcType == models.SourceGitHub {
		// 如果在子目录，把路径信息加入名称
		sub := strings.TrimPrefix(skillPath, repoPath)
		sub = strings.Trim(sub, string(os.PathSeparator))
		if sub != "" {
			// 使用子目录作为命名空间
			parts := strings.Split(filepath.ToSlash(sub), "/")
			if len(parts) > 1 {
				skillName = strings.Join(parts[:len(parts)-1], "/") + "/" + info.Name
			}
		}
	}

	// 确定版本号
	versionStr := opts.Version
	if versionStr == "" {
		versionStr = info.Version
	}
	if versionStr == "" {
		// 使用日期作为版本
		versionStr = time.Now().Format("2006.01.02")
	}

	// 确保版本目录
	targetDir, err := i.store.EnsureVersionDir(skillName, versionStr)
	if err != nil {
		return InstallResult{}, fmt.Errorf("ensure version dir: %w", err)
	}

	// 复制文件到 skillspool
	if err := copyDirContents(skillPath, targetDir); err != nil {
		return InstallResult{}, fmt.Errorf("copy files: %w", err)
	}

	// 更新 latest
	versions, _ := i.store.ListVersions(skillName)
	latest := version.Latest(versions)
	if latest != "" {
		if err := i.store.UpdateLatest(skillName, latest); err != nil {
			// 不致命错误
			fmt.Fprintf(os.Stderr, "warn: update latest: %v\n", err)
		}
	}

	// 确定 source 信息
	source := models.Source{Type: srcType}
	switch srcType {
	case models.SourceGitHub:
		source.URL = opts.Source
		source.Ref = opts.Ref
		source.SubPath = opts.SubPath
	case models.SourceNpx:
		source.Command = opts.Source
	case models.SourceLocal:
		source.Path = opts.Source
	}

	// 计算相对路径（相对 skillspool 根）
	relPath := strings.TrimPrefix(targetDir, i.store.Root())
	relPath = strings.Trim(relPath, string(os.PathSeparator))

	// 更新注册表
	i.reg.AddSkill(skillName, info, source, versionStr, relPath, opts.AgentIDs)

	// 分发到 Agent
	result := InstallResult{
		SkillName: skillName,
		Version:   versionStr,
		Source:    source,
		Agents:    make(map[string]InstallLink),
	}

	for _, agentID := range opts.AgentIDs {
		linkPath, err := i.agentMgr.InstallToAgent(skillName, targetDir, agentID, opts.Scope, opts.ProjectDir)
		link := InstallLink{AgentID: agentID, Path: linkPath, Success: err == nil}
		if err != nil {
			link.Error = err.Error()
		}
		result.Agents[agentID] = link
	}

	return result, nil
}

// resolveSource 解析来源，返回本地目录路径
func (i *Installer) resolveSource(opts InstallOptions) (models.SourceType, string, func(), error) {
	// 尝试判断来源类型
	switch {
	case looksLikeGitHub(opts.Source):
		// 克隆到临时目录
		tmp, err := os.MkdirTemp("", "skills-")
		if err != nil {
			return "", "", nil, err
		}
		cleanup := func() { os.RemoveAll(tmp) }

		cloneDir := filepath.Join(tmp, "repo")
		args := []string{"clone", "--depth", "1"}
		if opts.Ref != "" {
			args = append(args, "--branch", opts.Ref)
		}
		args = append(args, opts.Source, cloneDir)

		cmd := exec.Command("git", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cleanup()
			return "", "", nil, fmt.Errorf("git clone: %w", err)
		}

		// 如果指定了子目录
		localPath := cloneDir
		if opts.SubPath != "" {
			localPath = filepath.Join(cloneDir, filepath.FromSlash(opts.SubPath))
		}

		return models.SourceGitHub, localPath, cleanup, nil

	case looksLikeNpx(opts.Source):
		// npx 指令: 目前支持从 npx 目录解析
		// 简化处理: 从本地 npx 缓存查找，或提示不支持
		return "", "", nil, fmt.Errorf("npx source not fully implemented (use GitHub or local path for now)")

	default:
		// 假设是本地路径
		p := expandPath(opts.Source)
		if _, err := os.Stat(p); err != nil {
			return "", "", nil, fmt.Errorf("local path not found: %s", opts.Source)
		}
		return models.SourceLocal, p, nil, nil
	}
}

// --- 工具函数 ---

func looksLikeGitHub(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasPrefix(s, "https://github.com") || strings.HasPrefix(s, "git@github.com") {
		return true
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		// 其他 git 仓库
		u, err := url.Parse(s)
		if err == nil && strings.Contains(u.Host, "git") {
			return true
		}
	}
	// user/repo 格式
	if strings.Count(s, "/") == 1 && !strings.Contains(s, " ") && !strings.Contains(s, string(os.PathSeparator)) {
		return true
	}
	return false
}

func looksLikeNpx(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "npx ") || strings.HasPrefix(s, "npm ")
}

func expandPath(p string) string {
	p = os.ExpandEnv(p)
	if len(p) > 0 && p[0] == '~' {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[1:])
		}
	}
	return p
}

// copyDirContents 复制目录内容（不复制自身）
func copyDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, _ := e.Info()
			mode := os.FileMode(0o644)
			if info != nil {
				mode = info.Mode()
			}
			if err := os.WriteFile(dstPath, data, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
