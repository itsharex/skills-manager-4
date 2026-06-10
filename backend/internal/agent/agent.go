package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// InstallToAgent 为指定 Agent 安装技能（目录复制，不使用软链接以避免临时数据污染）
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

	// 目录复制（不使用软链接，避免临时数据污染 skillspool）
	destPath := filepath.Join(installDir, sanitize(skillName))
	if err := copyDir(target, destPath); err != nil {
		return "", fmt.Errorf("copy skill to agent %q: %w", agentID, err)
	}
	return destPath, nil
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

	targetPath := filepath.Join(installDir, sanitize(skillName))
	if _, err := os.Lstat(targetPath); err == nil {
		return os.RemoveAll(targetPath)
	}
	return nil
}

// ScanAgentSkills 扫描指定 Agent 的全局目录，返回已安装的技能名列表
func (m *Manager) ScanAgentSkills(agentID string) []string {
	agent, ok := m.agents[agentID]
	if !ok {
		return nil
	}
	skillDir := expandPath(agent.GlobalLocation)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil
	}
	var result []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// 检查是否包含 SKILL.md 以确认为技能目录
		if _, err := os.Stat(filepath.Join(skillDir, e.Name(), "SKILL.md")); err == nil {
			result = append(result, e.Name())
		}
	}
	sort.Strings(result)
	return result
}

// IsSkillInstalled 检查指定技能是否已安装到 Agent（全局目录）
func (m *Manager) IsSkillInstalled(agentID, skillName string) bool {
	agent, ok := m.agents[agentID]
	if !ok {
		return false
	}
	skillDir := expandPath(agent.GlobalLocation)
	target := filepath.Join(skillDir, sanitize(skillName))
	info, err := os.Stat(target)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// --- 内部: 目录复制（不使用软链接）---

// copyDir 递归复制目录，不使用软链接，确保项目中不会污染 skillspool
func copyDir(src, dst string) error {
	// 清理旧目录
	if _, err := os.Lstat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("remove existing target: %w", err)
		}
	}
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
