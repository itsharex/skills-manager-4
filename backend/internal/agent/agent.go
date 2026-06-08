package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Manager 管理 Agent 的技能分发
type Manager struct {
	agents map[string]models.Agent
}

// New 创建 Manager
func New(agents map[string]models.Agent) *Manager {
	return &Manager{agents: agents}
}

// Get 返回指定 ID 的 Agent
func (m *Manager) Get(id string) (models.Agent, bool) {
	a, ok := m.agents[id]
	return a, ok
}

// List 返回所有 Agent
func (m *Manager) List() map[string]models.Agent {
	return m.agents
}

// ListDetected 返回已检测到的 Agent ID 列表
func (m *Manager) ListDetected() []string {
	var ids []string
	for id, a := range m.agents {
		if a.Detected {
			ids = append(ids, id)
		}
	}
	return ids
}

// InstallToAgent 为指定 Agent 安装技能（创建软连接）
// target 是 skillspool 中技能版本目录的绝对路径
// agentID 是 Agent ID
// scope 是安装范围: "global" 或 "project"
// projectPath 是项目目录路径（当 scope == "project" 时使用）
func (m *Manager) InstallToAgent(skillName string, target string, agentID string, scope string, projectPath string) (string, error) {
	agent, ok := m.agents[agentID]
	if !ok {
		return "", fmt.Errorf("agent %q not found", agentID)
	}

	// 确定安装位置
	var installDir string
	switch scope {
	case "project":
		if projectPath == "" {
			return "", fmt.Errorf("project path required for project scope")
		}
		installDir = filepath.Join(projectPath, agent.SkillLocation)
	default: // global
		installDir = expandPath(agent.GlobalLocation)
	}

	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create agent skill dir: %w", err)
	}

	// 创建软连接（或降级）
	linkPath := filepath.Join(installDir, sanitize(skillName))
	err := createLink(target, linkPath)
	if err != nil {
		return "", fmt.Errorf("create link for agent %q: %w", agentID, err)
	}
	return linkPath, nil
}

// RemoveFromAgent 从指定 Agent 移除技能
func (m *Manager) RemoveFromAgent(skillName string, agentID string, scope string, projectPath string) error {
	agent, ok := m.agents[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	var installDir string
	switch scope {
	case "project":
		if projectPath == "" {
			return fmt.Errorf("project path required for project scope")
		}
		installDir = filepath.Join(projectPath, agent.SkillLocation)
	default:
		installDir = expandPath(agent.GlobalLocation)
	}

	linkPath := filepath.Join(installDir, sanitize(skillName))
	if _, err := os.Lstat(linkPath); err == nil {
		return os.RemoveAll(linkPath)
	}
	return nil
}

// --- 内部: 软连接创建（含降级策略）---

// createLink 尝试以多种方式将 target 链接到 linkPath
// 优先级: symlink → junction (Windows) → 目录复制
func createLink(target, linkPath string) error {
	// 清理旧链接
	if info, err := os.Lstat(linkPath); err == nil {
		// 检查是否已经是正确的链接
		if info.Mode()&os.ModeSymlink != 0 {
			if existingTarget, err := os.Readlink(linkPath); err == nil {
				if existingTarget == target {
					return nil
				}
			}
		}
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("remove existing: %w", err)
		}
	}

	// 尝试 1: symlink
	if err := os.Symlink(target, linkPath); err == nil {
		return nil
	}

	// Windows 降级策略
	if runtime.GOOS == "windows" {
		// 尝试 2: junction (通过 cmd /c mklink /J)
		if tryJunction(target, linkPath) {
			return nil
		}
	}

	// 最后降级: 目录复制
	return copyDir(target, linkPath)
}

func tryJunction(target, linkPath string) bool {
	// 简化: 直接尝试 os.Symlink 在 Windows 上有时需要管理员权限
	// 通过 cmd /c mklink /D 尝试目录软链接
	cmd := exec.Command("cmd", "/c", "mklink", "/D", linkPath, target)
	return cmd.Run() == nil
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

// --- 工具函数 ---

func sanitize(s string) string {
	// 取最后一段作为链接名（例如 @scope/name → name）
	parts := strings.Split(filepath.ToSlash(s), "/")
	name := parts[len(parts)-1]
	replacer := strings.NewReplacer(
		"\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", " ", "-",
	)
	return replacer.Replace(name)
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
