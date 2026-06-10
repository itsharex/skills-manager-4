package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"gopkg.in/yaml.v3"
)

// Config 配置管理
type Config struct {
	path    string
	Data    *models.Config
}

// Load 从指定路径加载配置，不存在则使用默认配置
func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigPath()
	}
	path = expandPath(path)

	cfg := &Config{
		path: path,
		Data: defaultConfig(),
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg.Data); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	// 自动检测 Agent
	cfg.detectAgents()

	// 保存（可能有新增的 detected 状态）
	_ = cfg.Save()

	return cfg, nil
}

// Save 保存配置
func (c *Config) Save() error {
	data, err := yaml.Marshal(c.Data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

// Path 返回配置文件路径
func (c *Config) Path() string {
	return c.path
}

// AddAgent 手动添加 Agent 配置
func (c *Config) AddAgent(id string, agent models.Agent) {
	if c.Data.Agents == nil {
		c.Data.Agents = make(map[string]models.Agent)
	}
	c.Data.Agents[id] = agent
}

// --- 内部 ---

func (c *Config) detectAgents() {
	home, _ := os.UserHomeDir()
	if home == "" {
		return
	}

	defaults := defaultAgents()
	for id, agent := range defaults {
		existing, ok := c.Data.Agents[id]
		if !ok {
			existing = agent
		}
		// 多策略检测 Agent 是否实际存在
		detected, installed, hasSkills := c.detectAgentMultiStrategy(&existing)
		existing.Detected = detected
		existing.Installed = installed
		_ = hasSkills
		if c.Data.Agents == nil {
			c.Data.Agents = make(map[string]models.Agent)
		}
		c.Data.Agents[id] = existing
	}
}

// detectAgentMultiStrategy 多策略检测 Agent 是否存在
// 返回: (detected, installed, hasSkills)
//   - detected: 通过任意一种策略检测到 Agent 的存在
//   - installed: 在 Agent 中实际安装了至少一个技能
//   - hasSkills: 保留字段，便于未来扩展
func (c *Config) detectAgentMultiStrategy(agent *models.Agent) (bool, bool, bool) {
	home, _ := os.UserHomeDir()
	detected := false
	installed := false

	// 策略 1：检测 GlobalLocation 目录（含 skills 子目录）是否存在
	if agent.GlobalLocation != "" {
		globalPath := expandPath(agent.GlobalLocation)
		if _, err := os.Stat(globalPath); err == nil {
			detected = true
			// 判断是否有技能（子目录有 SKILL.md 或 SKILL.yaml）
			if hasAnySkill(globalPath) {
				installed = true
			}
		}
	}

	// 策略 2：检测 GlobalDirectoryKey 对应的父目录是否存在
	// 例如 ~/.trae-cn/ 存在但 ~/.trae-cn/skills/ 不存在
	if !detected && agent.GlobalDirectoryKey != "" {
		keyPath := expandPath(filepath.Join(home, agent.GlobalDirectoryKey))
		parent := filepath.Dir(keyPath)
		if _, err := os.Stat(parent); err == nil {
			detected = true
		}
		// 同时检测 home 下与 GlobalDirectoryKey 共享前缀的目录
		// 例如 ~/.trae-cn/ 存在则视作 trae-cn 已检测
		// 这里 parent 已经是 ~/.trae-cn
	}

	// 策略 3：检测 GlobalLocation 路径的各级父目录（用于目录结构嵌套较深的情况）
	if !detected && agent.GlobalLocation != "" {
		// 向上递归 3 级父目录
		p := expandPath(agent.GlobalLocation)
		for i := 0; i < 3; i++ {
			p = filepath.Dir(p)
			if p == "" || p == "." || p == "/" {
				break
			}
			if _, err := os.Stat(p); err == nil {
				detected = true
				break
			}
		}
	}

	// 策略 4：检测 Agent 可执行文件是否在 PATH 中
	if !detected {
		execName := agentExecutableName(agent)
		if execName != "" {
			if _, err := lookupExecutable(execName); err == nil {
				detected = true
			}
		}
	}

	// 策略 5：检测 user-level 配置目录（独立于技能目录）
	// 例如 trae-cn 在 Windows 上一定会创建 ~/.trae-cn 目录
	if !detected {
		configDirs := agentConfigDirs(agent)
		for _, d := range configDirs {
			expanded := expandPath(d)
			if _, err := os.Stat(expanded); err == nil {
				detected = true
				break
			}
		}
	}

	return detected, installed, false
}

// hasAnySkill 判断目录中是否包含至少一个技能（存在 SKILL.md 或 SKILL.yaml）
func hasAnySkill(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(dir, e.Name())
		// 检查常见的技能元数据文件
		for _, meta := range []string{"SKILL.md", "SKILL.yaml", "SKILL.yml", "skill.md"} {
			if _, err := os.Stat(filepath.Join(sub, meta)); err == nil {
				return true
			}
		}
	}
	return false
}

// agentExecutableName 根据 Agent 名称返回可执行文件名（用于 PATH 检测）
func agentExecutableName(agent *models.Agent) string {
	name := strings.ToLower(agent.Name)
	switch {
	case strings.Contains(name, "trae"):
		return "trae"
	case strings.Contains(name, "claude"):
		return "claude"
	case strings.Contains(name, "cursor"):
		return "cursor"
	case strings.Contains(name, "codex"):
		return "codex"
	case strings.Contains(name, "copilot"):
		return "copilot"
	case strings.Contains(name, "windsurf"):
		return "windsurf"
	case strings.Contains(name, "aider"):
		return "aider"
	case strings.Contains(name, "continue"):
		return "continue"
	case strings.Contains(name, "gemini"):
		return "gemini"
	case strings.Contains(name, "antigravity") || strings.Contains(name, "anti-gravity"):
		return "antigravity"
	case strings.Contains(name, "cline"):
		return "cline"
	case strings.Contains(name, "roo"):
		return "roo"
	case strings.Contains(name, "kiro"):
		return "kiro"
	case strings.Contains(name, "bolt"):
		return "bolt"
	case strings.Contains(name, "same"):
		return "same"
	}
	return ""
}

// agentConfigDirs 返回 Agent 强相关的 user-level 配置目录
// 即使技能目录不存在，这些目录被创建也意味着 Agent 已经被用户使用过
func agentConfigDirs(agent *models.Agent) []string {
	name := strings.ToLower(agent.Name)
	switch name {
	case "trae cn", "trae-cn":
		return []string{"~/.trae-cn"}
	case "trae":
		return []string{"~/.trae"}
	case "claude code", "claude":
		return []string{"~/.claude"}
	case "cursor":
		return []string{"~/.cursor"}
	case "antigravity":
		// Antigravity 是基于 Gemini 的，检测多个候选目录
		return []string{
			"~/.gemini/antigravity",
			"~/.antigravity",
			"~/.config/antigravity",
		}
	case "windsurf":
		return []string{"~/.codeium/windsurf", "~/.windsurf"}
	case "codex":
		return []string{"~/.codex"}
	case "aider":
		return []string{"~/.aider"}
	case "continue":
		return []string{"~/.continue"}
	case "gemini cli", "gemini":
		return []string{"~/.gemini"}
	case "cline":
		return []string{"~/.cline"}
	case "roo code", "roo":
		return []string{"~/.roo-code", "~/.roo"}
	case "kiro":
		return []string{"~/.kiro"}
	case "bolt":
		return []string{"~/.bolt"}
	}
	// 兜底：从 GlobalDirectoryKey 推断
	if agent.GlobalDirectoryKey != "" {
		key := agent.GlobalDirectoryKey
		// 取第一个路径段作为候选目录
		if idx := strings.Index(key, "/"); idx > 0 {
			return []string{"~/" + key[:idx]}
		}
	}
	return nil
}

// lookupExecutable 在 PATH 中查找可执行文件（Windows 下会带 .exe 后缀）
func lookupExecutable(name string) (string, error) {
	paths := os.Getenv("PATH")
	if paths == "" {
		return "", fmt.Errorf("PATH is empty")
	}
	// Windows 平台下常见可执行文件后缀
	suffixes := []string{""}
	if runtime.GOOS == "windows" {
		suffixes = []string{".exe", ".cmd", ".bat", ".COM", ""}
	}
	for _, p := range filepath.SplitList(paths) {
		for _, suf := range suffixes {
			full := filepath.Join(p, name+suf)
			if _, err := os.Stat(full); err == nil {
				return full, nil
			}
		}
	}
	return "", fmt.Errorf("not found in PATH: %s", name)
}

func defaultConfig() *models.Config {
	cfg := &models.Config{
		Agents: defaultAgents(),
	}
	cfg.Skillspool.Root = defaultSkillspoolRoot()
	cfg.SkillMarket = models.SkillMarketConfig{
		CacheEnabled:     true,
		CacheExpiryHours: 24,
	}
	return cfg
}

func defaultAgents() map[string]models.Agent {
	home, _ := os.UserHomeDir()

	// 说明：
	//   GlobalLocation: 全局技能目录（用户 home 下）
	//   SkillLocation:  项目级技能目录（项目根目录下）
	//   SupportsProject: 是否支持项目级安装
	//   GlobalDirectoryKey / ProjectDirectoryKey:
	//     用于检测多 Agent 是否共享同一目录
	//     key 相同说明多个 Agent 写入同个目录，有数据混乱风险

	return map[string]models.Agent{
		// ======= 标准 Agent =======

		"cursor": {
			Name:                "Cursor",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".cursor", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".cursor/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"claude-code": {
			Name:                "Claude Code",
			SkillLocation:       ".claude/skills",
			GlobalLocation:      filepath.Join(home, ".claude", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".claude/skills",
			ProjectDirectoryKey: ".claude/skills",
		},
		"codex": {
			Name:                "Codex",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".codex", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".codex/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"opencode": {
			Name:                "OpenCode",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".config", "opencode", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".config/opencode/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"antigravity": {
			Name:                "Antigravity",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".gemini", "antigravity", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".gemini/antigravity/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"amp": {
			Name:                "Amp",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".config", "agents", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".config/agents/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"kimi-cli": {
			Name:                "Kimi Code CLI",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".config", "agents", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".config/agents/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"augment": {
			Name:                "Augment",
			SkillLocation:       ".augment/skills",
			GlobalLocation:      filepath.Join(home, ".augment", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".augment/skills",
			ProjectDirectoryKey: ".augment/skills",
		},
		"openclaw": {
			Name:                "OpenClaw",
			SkillLocation:       "skills",
			GlobalLocation:      filepath.Join(home, ".openclaw", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".openclaw/skills",
			ProjectDirectoryKey: "skills",
		},
		"copaw": {
			Name:                "Copaw",
			SkillLocation:       ".copaw/skill_pool",
			GlobalLocation:      filepath.Join(home, ".copaw", "skill_pool"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".copaw/skill_pool",
			ProjectDirectoryKey: ".copaw/skill_pool",
		},
		"cline": {
			Name:                "Cline",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".agents", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".agents/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"codebuddy": {
			Name:                "CodeBuddy",
			SkillLocation:       ".codebuddy/skills",
			GlobalLocation:      filepath.Join(home, ".codebuddy", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".codebuddy/skills",
			ProjectDirectoryKey: ".codebuddy/skills",
		},
		"commandcode": {
			Name:                "Command Code",
			SkillLocation:       ".commandcode/skills",
			GlobalLocation:      filepath.Join(home, ".commandcode", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".commandcode/skills",
			ProjectDirectoryKey: ".commandcode/skills",
		},
		"continue": {
			Name:                "Continue",
			SkillLocation:       ".continue/skills",
			GlobalLocation:      filepath.Join(home, ".continue", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".continue/skills",
			ProjectDirectoryKey: ".continue/skills",
		},
		"crush": {
			Name:                "Crush",
			SkillLocation:       ".crush/skills",
			GlobalLocation:      filepath.Join(home, ".config", "crush", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".config/crush/skills",
			ProjectDirectoryKey: ".crush/skills",
		},
		"junie": {
			Name:                "Junie",
			SkillLocation:       ".junie/skills",
			GlobalLocation:      filepath.Join(home, ".junie", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".junie/skills",
			ProjectDirectoryKey: ".junie/skills",
		},
		"iflow-cli": {
			Name:                "iFlow CLI",
			SkillLocation:       ".iflow/skills",
			GlobalLocation:      filepath.Join(home, ".iflow", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".iflow/skills",
			ProjectDirectoryKey: ".iflow/skills",
		},
		"kiro-cli": {
			Name:                "Kiro CLI",
			SkillLocation:       ".kiro/skills",
			GlobalLocation:      filepath.Join(home, ".kiro", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".kiro/skills",
			ProjectDirectoryKey: ".kiro/skills",
		},
		"kode": {
			Name:                "Kode",
			SkillLocation:       ".kode/skills",
			GlobalLocation:      filepath.Join(home, ".kode", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".kode/skills",
			ProjectDirectoryKey: ".kode/skills",
		},
		"mcpjam": {
			Name:                "MCPJam",
			SkillLocation:       ".mcpjam/skills",
			GlobalLocation:      filepath.Join(home, ".mcpjam", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".mcpjam/skills",
			ProjectDirectoryKey: ".mcpjam/skills",
		},
		"mistral-vibe": {
			Name:                "Mistral Vibe",
			SkillLocation:       ".vibe/skills",
			GlobalLocation:      filepath.Join(home, ".vibe", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".vibe/skills",
			ProjectDirectoryKey: ".vibe/skills",
		},
		"mux": {
			Name:                "Mux",
			SkillLocation:       ".mux/skills",
			GlobalLocation:      filepath.Join(home, ".mux", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".mux/skills",
			ProjectDirectoryKey: ".mux/skills",
		},
		"openclaude": {
			Name:                "OpenClaude IDE",
			SkillLocation:       ".openclaude/skills",
			GlobalLocation:      filepath.Join(home, ".openclaude", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".openclaude/skills",
			ProjectDirectoryKey: ".openclaude/skills",
		},
		"openhands": {
			Name:                "OpenHands",
			SkillLocation:       ".openhands/skills",
			GlobalLocation:      filepath.Join(home, ".openhands", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".openhands/skills",
			ProjectDirectoryKey: ".openhands/skills",
		},
		"pi": {
			Name:                "Pi",
			SkillLocation:       ".pi/skills",
			GlobalLocation:      filepath.Join(home, ".pi", "agent", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".pi/agent/skills",
			ProjectDirectoryKey: ".pi/skills",
		},
		"qoder": {
			Name:                "Qoder",
			SkillLocation:       ".qoder/skills",
			GlobalLocation:      filepath.Join(home, ".qoder", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".qoder/skills",
			ProjectDirectoryKey: ".qoder/skills",
		},
		"qoder-work": {
			Name:                "QoderWork",
			SkillLocation:       ".qoderwork/skills",
			GlobalLocation:      filepath.Join(home, ".qoderwork", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".qoderwork/skills",
			ProjectDirectoryKey: ".qoderwork/skills",
		},
		"qwen-code": {
			Name:                "Qwen Code",
			SkillLocation:       ".qwen/skills",
			GlobalLocation:      filepath.Join(home, ".qwen", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".qwen/skills",
			ProjectDirectoryKey: ".qwen/skills",
		},
		"trae": {
			Name:                "Trae",
			SkillLocation:       ".trae/skills",
			GlobalLocation:      filepath.Join(home, ".trae", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".trae/skills",
			ProjectDirectoryKey: ".trae/skills",
		},
		"trae-cn": {
			Name:                "Trae CN",
			SkillLocation:       ".trae/skills",
			GlobalLocation:      filepath.Join(home, ".trae-cn", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".trae-cn/skills",
			ProjectDirectoryKey: ".trae/skills",
		},
		"zencoder": {
			Name:                "Zencoder",
			SkillLocation:       ".zencoder/skills",
			GlobalLocation:      filepath.Join(home, ".zencoder", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".zencoder/skills",
			ProjectDirectoryKey: ".zencoder/skills",
		},
		"neovate": {
			Name:                "Neovate",
			SkillLocation:       ".neovate/skills",
			GlobalLocation:      filepath.Join(home, ".neovate", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".neovate/skills",
			ProjectDirectoryKey: ".neovate/skills",
		},
		"pochi": {
			Name:                "Pochi",
			SkillLocation:       ".pochi/skills",
			GlobalLocation:      filepath.Join(home, ".pochi", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".pochi/skills",
			ProjectDirectoryKey: ".pochi/skills",
		},
		"adal": {
			Name:                "AdaL",
			SkillLocation:       ".adal/skills",
			GlobalLocation:      filepath.Join(home, ".adal", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".adal/skills",
			ProjectDirectoryKey: ".adal/skills",
		},
		"kilocode": {
			Name:                "Kilo Code",
			SkillLocation:       ".kilocode/skills",
			GlobalLocation:      filepath.Join(home, ".kilocode", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".kilocode/skills",
			ProjectDirectoryKey: ".kilocode/skills",
		},
		"roo-code": {
			Name:                "Roo Code",
			SkillLocation:       ".roo/skills",
			GlobalLocation:      filepath.Join(home, ".roo", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".roo/skills",
			ProjectDirectoryKey: ".roo/skills",
		},
		"goose": {
			Name:                "Goose",
			SkillLocation:       ".goose/skills",
			GlobalLocation:      filepath.Join(home, ".config", "goose", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".config/goose/skills",
			ProjectDirectoryKey: ".goose/skills",
		},
		"gemini-cli": {
			Name:                "Gemini CLI",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".gemini", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".gemini/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"github-copilot": {
			Name:                "GitHub Copilot",
			SkillLocation:       ".agents/skills",
			GlobalLocation:      filepath.Join(home, ".copilot", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".copilot/skills",
			ProjectDirectoryKey: ".agents/skills",
		},
		"clawdbot": {
			Name:                "Clawdbot",
			SkillLocation:       ".clawdbot/skills",
			GlobalLocation:      filepath.Join(home, ".clawdbot", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".clawdbot/skills",
			ProjectDirectoryKey: ".clawdbot/skills",
		},
		"droid": {
			Name:                "Droid",
			SkillLocation:       ".factory/skills",
			GlobalLocation:      filepath.Join(home, ".factory", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".factory/skills",
			ProjectDirectoryKey: ".factory/skills",
		},
		"windsurf": {
			Name:                "Windsurf",
			SkillLocation:       ".windsurf/skills",
			GlobalLocation:      filepath.Join(home, ".codeium", "windsurf", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".codeium/windsurf/skills",
			ProjectDirectoryKey: ".windsurf/skills",
		},
		"moltbot": {
			Name:                "MoltBot",
			SkillLocation:       ".moltbot/skills",
			GlobalLocation:      filepath.Join(home, ".moltbot", "skills"),
			SupportsProject:     true,
			GlobalDirectoryKey:  ".moltbot/skills",
			ProjectDirectoryKey: ".moltbot/skills",
		},

		// ======= 仅支持全局安装的 Agent =======

		"hermes-agent": {
			Name:                "Hermes Agent",
			SkillLocation:       "",
			GlobalLocation:      filepath.Join(home, ".hermes", "skills"),
			SupportsProject:     false,
			GlobalDirectoryKey:  ".hermes/skills",
			ProjectDirectoryKey: "",
		},
	}
}

func defaultSkillspoolRoot() string {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "SkillsManager", "skillspool")
		}
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "SkillsManager", "skillspool")
		}
	default:
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", "skillsmanager", "skillspool")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skillsmanager", "skillspool")
}

func defaultConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "SkillsManager", "agents.yaml")
		}
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "SkillsManager", "agents.yaml")
		}
	default:
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "skillsmanager", "agents.yaml")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skillsmanager", "agents.yaml")
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
