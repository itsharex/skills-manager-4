package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
		// 检测 Agent 是否实际存在
		globalPath := expandPath(existing.GlobalLocation)
		if _, err := os.Stat(globalPath); err == nil {
			existing.Detected = true
			existing.Installed = true
		} else {
			existing.Detected = false
		}
		if c.Data.Agents == nil {
			c.Data.Agents = make(map[string]models.Agent)
		}
		c.Data.Agents[id] = existing
	}
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
	return map[string]models.Agent{
		// 主流 AI Coding Agent
		"claude-code": {
			Name:           "Claude Code",
			SkillLocation:  ".claude/skills",
			GlobalLocation: filepath.Join(home, ".claude", "skills"),
		},
		"github-copilot": {
			Name:           "GitHub Copilot",
			SkillLocation:  ".github/copilot/skills",
			GlobalLocation: filepath.Join(home, ".github/copilot", "skills"),
		},
		"cursor": {
			Name:           "Cursor",
			SkillLocation:  ".cursor/skills",
			GlobalLocation: filepath.Join(home, ".cursor", "skills"),
		},
		"windsurf": {
			Name:           "Windsurf",
			SkillLocation:  ".windsurf/skills",
			GlobalLocation: filepath.Join(home, ".windsurf", "skills"),
		},
		"trae": {
			Name:           "Trae",
			SkillLocation:  ".trae/skills",
			GlobalLocation: filepath.Join(home, ".trae-cn", "skills"),
		},
		"gemini-cli": {
			Name:           "Gemini CLI",
			SkillLocation:  ".gemini/skills",
			GlobalLocation: filepath.Join(home, ".gemini", "skills"),
		},
		"openai-codex": {
			Name:           "OpenAI Codex CLI",
			SkillLocation:  ".codex/skills",
			GlobalLocation: filepath.Join(home, ".codex", "skills"),
		},
		"openclaw": {
			Name:           "OpenClaw",
			SkillLocation:  ".openclaw/skills",
			GlobalLocation: filepath.Join(home, ".openclaw", "skills"),
		},
		"tabnine": {
			Name:           "Tabnine",
			SkillLocation:  ".tabnine/skills",
			GlobalLocation: filepath.Join(home, ".tabnine", "skills"),
		},
		"supermaven": {
			Name:           "Supermaven",
			SkillLocation:  ".supermaven/skills",
			GlobalLocation: filepath.Join(home, ".supermaven", "skills"),
		},
		"aider": {
			Name:           "Aider",
			SkillLocation:  ".aider/skills",
			GlobalLocation: filepath.Join(home, ".aider", "skills"),
		},
		"coderabbit": {
			Name:           "CodeRabbit",
			SkillLocation:  ".coderabbit/skills",
			GlobalLocation: filepath.Join(home, ".coderabbit", "skills"),
		},
		"devin": {
			Name:           "Devin",
			SkillLocation:  ".devin/skills",
			GlobalLocation: filepath.Join(home, ".devin", "skills"),
		},
		"junie": {
			Name:           "Junie",
			SkillLocation:  ".junie/skills",
			GlobalLocation: filepath.Join(home, ".junie", "skills"),
		},
		"llion": {
			Name:           "Llion",
			SkillLocation:  ".llion/skills",
			GlobalLocation: filepath.Join(home, ".llion", "skills"),
		},
		"rogue": {
			Name:           "Rogue",
			SkillLocation:  ".rogue/skills",
			GlobalLocation: filepath.Join(home, ".rogue", "skills"),
		},
		"hermes": {
			Name:           "Hermes",
			SkillLocation:  ".hermes/skills",
			GlobalLocation: filepath.Join(home, ".hermes", "skills"),
		},
		"antigravity": {
			Name:           "Antigravity",
			SkillLocation:  ".antigravity/skills",
			GlobalLocation: filepath.Join(home, ".antigravity", "skills"),
		},
		"swe-agent": {
			Name:           "SWE-agent",
			SkillLocation:  ".swe-agent/skills",
			GlobalLocation: filepath.Join(home, ".swe-agent", "skills"),
		},
		"opencode": {
			Name:           "Opencode",
			SkillLocation:  ".opencode/skills",
			GlobalLocation: filepath.Join(home, ".opencode", "skills"),
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
