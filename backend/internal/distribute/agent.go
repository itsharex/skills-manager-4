package distribute

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Agent represents a detected or configured AI agent.
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SkillsDir   string `json:"skills_dir"`
	AutoDetected bool  `json:"auto_detected"`
	Enabled     bool   `json:"enabled"`
	DetectCmd   string `json:"-"` // CLI binary name to detect via exec.LookPath (e.g. "codex")
	DetectPath  string `json:"-"` // Config/dir path to detect via os.Stat (e.g. "~/.cursor/")
}

// KnownAgents returns the list of known AI agents and their default skills directories.
// This is used for auto-detection during `skill init` and `skill doctor`.
//
// Detection methods:
//   - CLI agents: exec.LookPath(DetectCmd) — checks if binary is in PATH
//   - IDE/Desktop agents: os.Stat(DetectPath) — checks if config/app directory exists
//   - VSCode extensions: os.Stat(DetectPath) — checks if extension is installed in ~/.vscode/extensions/
func KnownAgents() []Agent {
	home, _ := os.UserHomeDir()

	agents := []Agent{
		// =====================================================
		// CLI agents — detected by exec.LookPath(DetectCmd)
		// =====================================================

		// Claude Code — npm install -g @anthropic-ai/claude-code → command: claude
		{ID: "claude-code", Name: "Claude Code", SkillsDir: filepath.Join(home, ".claude", "skills"), DetectCmd: "claude"},

		// Gemini CLI — npm install -g @google/gemini-cli → command: gemini
		{ID: "gemini-cli", Name: "Gemini CLI", SkillsDir: filepath.Join(home, ".gemini", "skills"), DetectCmd: "gemini"},

		// OpenAI Codex CLI — npm install -g @openai/codex → command: codex
		{ID: "codex-cli", Name: "Codex CLI", SkillsDir: filepath.Join(home, ".codex", "skills"), DetectCmd: "codex"},

		// Antigravity CLI (Google) — command: agy
		// Skills dir: ~/.gemini/antigravity/skills/
		{ID: "antigravity-cli", Name: "Antigravity CLI", SkillsDir: filepath.Join(home, ".gemini", "antigravity", "skills"), DetectCmd: "agy"},

		// Aider — pip install aider-install → command: aider
		{ID: "aider", Name: "Aider", SkillsDir: filepath.Join(home, ".aider", "skills"), DetectCmd: "aider"},

		// OpenCode CLI — npm install -g opencode-ai → command: opencode
		// Desktop app: /Applications/OpenCode.app, config: ~/.config/opencode/
		{ID: "opencode", Name: "OpenCode", SkillsDir: filepath.Join(home, ".config", "opencode", "skills"), DetectCmd: "opencode"},

		// OpenCode Desktop — config dir: ~/.config/opencode/
		{ID: "opencode-desktop", Name: "OpenCode Desktop", SkillsDir: filepath.Join(home, ".config", "opencode", "skills"), DetectPath: filepath.Join(home, ".config", "opencode")},

		// SWE-agent — pip install sweagent → command: sweagent
		{ID: "swe-agent", Name: "SWE-agent", SkillsDir: filepath.Join(home, ".swe-agent", "skills"), DetectCmd: "sweagent"},

		// Goose (Block) — brew install block-goose-cli → command: goose
		{ID: "goose", Name: "Goose", SkillsDir: filepath.Join(home, ".config", "goose", "skills"), DetectCmd: "goose"},

		// Ollama — brew install ollama → command: ollama
		{ID: "ollama", Name: "Ollama", SkillsDir: filepath.Join(home, ".ollama", "skills"), DetectCmd: "ollama"},

		// Amazon Q Developer CLI — command: q
		{ID: "amazon-q-cli", Name: "Amazon Q CLI", SkillsDir: filepath.Join(home, ".aws", "amazonq", "skills"), DetectCmd: "q"},

		// Hermes Agent — command: hermes
		{ID: "hermes", Name: "Hermes", SkillsDir: filepath.Join(home, ".hermes", "skills"), DetectCmd: "hermes"},

		// GitHub Copilot CLI — command: github-copilot-cli
		{ID: "copilot-cli", Name: "Copilot CLI", SkillsDir: filepath.Join(home, ".github-copilot", "skills"), DetectCmd: "github-copilot-cli"},

		// CodeGPT — command: codegpt
		{ID: "codegpt", Name: "CodeGPT", SkillsDir: filepath.Join(home, ".codegpt", "skills"), DetectCmd: "codegpt"},

		// Tabby — command: tabby
		{ID: "tabby", Name: "Tabby", SkillsDir: filepath.Join(home, ".tabby", "skills"), DetectCmd: "tabby"},

		// =====================================================
		// IDE / Desktop apps — detected by os.Stat(DetectPath)
		// =====================================================

		// Trae CN 国内版 (字节跳动) — config dir: ~/.trae-cn/
		{ID: "trae-cn", Name: "Trae CN", SkillsDir: filepath.Join(home, ".trae-cn", "skills"), DetectPath: filepath.Join(home, ".trae-cn")},

		// Trae 国际版 — config dir: ~/.trae/
		{ID: "trae", Name: "Trae", SkillsDir: filepath.Join(home, ".trae", "skills"), DetectPath: filepath.Join(home, ".trae")},

		// Trae AICC / Trae Work CN — config dir: ~/.trae-aicc/
		{ID: "trae-aicc", Name: "Trae AICC", SkillsDir: filepath.Join(home, ".trae-aicc", "skills"), DetectPath: filepath.Join(home, ".trae-aicc")},

		// Antigravity IDE (Google) — config dir: ~/.antigravity-ide/
		{ID: "antigravity-ide", Name: "Antigravity IDE", SkillsDir: filepath.Join(home, ".gemini", "antigravity-ide", "skills"), DetectPath: filepath.Join(home, ".antigravity-ide")},

		// Codex Desktop (OpenAI) — config dir: ~/.codex/
		{ID: "codex-desktop", Name: "Codex Desktop", SkillsDir: filepath.Join(home, ".codex", "skills"), DetectPath: filepath.Join(home, ".codex")},

		// Cursor — config dir: ~/.cursor/
		{ID: "cursor", Name: "Cursor", SkillsDir: filepath.Join(home, ".cursor", "skills"), DetectPath: filepath.Join(home, ".cursor")},

		// Windsurf (Codeium) — config dir: ~/.windsurf/
		{ID: "windsurf", Name: "Windsurf", SkillsDir: filepath.Join(home, ".windsurf", "skills"), DetectPath: filepath.Join(home, ".windsurf")},

		// Kiro (Amazon) — config dir: ~/.kiro/
		{ID: "kiro", Name: "Kiro", SkillsDir: filepath.Join(home, ".kiro", "skills"), DetectPath: filepath.Join(home, ".kiro")},

		// Claude Desktop — config dir: ~/Library/Application Support/Claude/ (macOS)
		{ID: "claude-desktop", Name: "Claude Desktop", SkillsDir: filepath.Join(home, "Library", "Application Support", "Claude", "skills"), DetectPath: filepath.Join(home, "Library", "Application Support", "Claude")},

		// Zed — config dir: ~/.config/zed/
		{ID: "zed", Name: "Zed", SkillsDir: filepath.Join(home, ".config", "zed", "skills"), DetectPath: filepath.Join(home, ".config", "zed")},

		// JetBrains AI / Junie — config dir: ~/.JetBrains/ (detects any JetBrains IDE)
		{ID: "jetbrains-ai", Name: "JetBrains AI", SkillsDir: filepath.Join(home, ".jetbrains-ai", "skills"), DetectPath: filepath.Join(home, ".JetBrains")},

		// Devin — config dir: ~/.devin/
		{ID: "devin", Name: "Devin", SkillsDir: filepath.Join(home, ".devin", "skills"), DetectPath: filepath.Join(home, ".devin")},

		// =====================================================
		// VSCode extensions — detected by extension dir in ~/.vscode/extensions/
		// =====================================================

		// Claude Code VSCode Extension — anthropic.claude-code
		{ID: "claude-code-vscode", Name: "Claude Code (VSCode)", SkillsDir: filepath.Join(home, ".claude", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "anthropic.claude-code")},

		// Continue — VSCode extension: continue.continue
		{ID: "continue", Name: "Continue", SkillsDir: filepath.Join(home, ".continue", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "continue.continue")},

		// Cline — VSCode extension: saoudrizwan.claude-dev
		{ID: "cline", Name: "Cline", SkillsDir: filepath.Join(home, ".cline", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "saoudrizwan.claude-dev")},

		// Roo Code / Kilo Code — VSCode extension: rooveterinaryinc.roo-cline
		{ID: "roo-code", Name: "Roo Code", SkillsDir: filepath.Join(home, ".roocode", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "rooveterinaryinc.roo-cline")},

		// GitHub Copilot — VSCode extension: github.copilot
		{ID: "github-copilot", Name: "GitHub Copilot", SkillsDir: filepath.Join(home, ".github-copilot", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "github.copilot")},

		// Tabnine — config dir: ~/.TabNine/
		{ID: "tabnine", Name: "Tabnine", SkillsDir: filepath.Join(home, ".tabnine", "skills"), DetectPath: filepath.Join(home, ".TabNine")},

		// Supermaven — VSCode extension: supermaven.supermaven
		{ID: "supermaven", Name: "Supermaven", SkillsDir: filepath.Join(home, ".supermaven", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "supermaven.supermaven")},

		// CodeRabbit — VSCode extension: coderabbit.coderabbit
		{ID: "coderabbit", Name: "CodeRabbit", SkillsDir: filepath.Join(home, ".coderabbit", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "coderabbit.coderabbit")},

		// Cody (Sourcegraph) — VSCode extension: sourcegraph.cody-ai
		{ID: "cody", Name: "Cody", SkillsDir: filepath.Join(home, ".cody", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "sourcegraph.cody-ai")},

		// Codeium — VSCode extension: codeium.codeium
		{ID: "codeium", Name: "Codeium", SkillsDir: filepath.Join(home, ".codeium", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "codeium.codeium")},

		// Pieces — VSCode extension: pieces.pieces-vscode
		{ID: "pieces", Name: "Pieces", SkillsDir: filepath.Join(home, ".pieces", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "pieces.pieces-vscode")},

		// OpenClaw — VSCode extension: openclaw.openclaw
		{ID: "openclaw", Name: "OpenClaw", SkillsDir: filepath.Join(home, ".openclaw", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "openclaw.openclaw")},

		// Amazon Q Developer (IDE) — VSCode extension: amazonwebservices.amazon-q
		{ID: "amazon-q", Name: "Amazon Q", SkillsDir: filepath.Join(home, ".amazon-q", "skills"), DetectPath: filepath.Join(home, ".vscode", "extensions", "amazonwebservices.amazon-q")},
	}

	return agents
}

// DetectAgents scans the system for installed agents.
// For CLI agents: checks if the binary is in PATH via exec.LookPath.
// For IDE/Desktop agents: checks if the config/application directory exists.
// Returns all detected agents (non-installed agents are excluded).
func DetectAgents() ([]Agent, error) {
	var detected []Agent

	for _, agent := range KnownAgents() {
		if agent.detected() {
			agent.AutoDetected = true
			detected = append(detected, agent)
		}
	}

	return detected, nil
}

// DetectedAgents returns all known agents with detection status populated.
// Unlike DetectAgents, this returns ALL agents (both detected and not),
// with AutoDetected set correctly for each.
func DetectedAgents() []Agent {
	all := KnownAgents()
	for i := range all {
		all[i].AutoDetected = all[i].detected()
	}
	return all
}

// detected checks whether this agent is installed on the system.
func (a *Agent) detected() bool {
	if a.DetectCmd != "" {
		_, err := exec.LookPath(a.DetectCmd)
		return err == nil
	}
	if a.DetectPath != "" {
		// For VSCode extensions, the directory name includes version suffix
		// e.g. ~/.vscode/extensions/github.copilot-1.234.0/
		// So we check if any directory starting with DetectPath exists
		if strings.Contains(a.DetectPath, filepath.Join(".vscode", "extensions")) {
			parentDir := filepath.Dir(a.DetectPath)
			prefix := filepath.Base(a.DetectPath)
			entries, err := os.ReadDir(parentDir)
			if err != nil {
				return false
			}
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
					return true
				}
			}
			return false
		}
		info, err := os.Stat(a.DetectPath)
		return err == nil && info.IsDir()
	}
	return false
}

// GetAgentByID looks up an agent by its ID from KnownAgents.
func GetAgentByID(id string) (*Agent, error) {
	for _, agent := range KnownAgents() {
		if agent.ID == id {
			return &agent, nil
		}
	}
	return nil, fmt.Errorf("unknown agent: %s", id)
}

// GetAgentConfigPath returns the path to an agent's skill configuration file.
func GetAgentConfigPath(agentID string) (string, error) {
	agent, err := GetAgentByID(agentID)
	if err != nil {
		return "", err
	}
	return filepath.Join(agent.SkillsDir, ".skills.json"), nil
}

// ValidateAgentPath checks if an agent's skills directory exists and is writable.
func ValidateAgentPath(agentID string) error {
	agent, err := GetAgentByID(agentID)
	if err != nil {
		return err
	}

	info, err := os.Stat(agent.SkillsDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("agent %q skills directory does not exist: %s", agentID, agent.SkillsDir)
	}
	if err != nil {
		return fmt.Errorf("access agent %q skills directory: %w", agentID, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("agent %q path is not a directory: %s", agentID, agent.SkillsDir)
	}

	// Check writability by attempting to create a temp file
	tmpFile := filepath.Join(agent.SkillsDir, ".skills-write-test")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("agent %q skills directory is not writable: %w", agentID, err)
	}
	f.Close()
	os.Remove(tmpFile)

	return nil
}

// GetAgentSkillsDir returns the skills directory for a given agent.
// This is the directory where skill symlinks or copies are placed.
func GetAgentSkillsDir(agentID string) (string, error) {
	agent, err := GetAgentByID(agentID)
	if err != nil {
		return "", err
	}
	return agent.SkillsDir, nil
}

// ResolveAgentIDs resolves agent IDs, handling the special "all" keyword.
// If "all" is included, returns all known agent IDs.
func ResolveAgentIDs(ids []string) []string {
	var result []string
	hasAll := false
	for _, id := range ids {
		if id == "all" || id == "*" {
			hasAll = true
		} else {
			result = append(result, id)
		}
	}
	if hasAll {
		for _, agent := range KnownAgents() {
			result = append(result, agent.ID)
		}
	}
	return result
}

// SafeAgentName sanitizes an agent name for display.
func SafeAgentName(id string) string {
	for _, agent := range KnownAgents() {
		if agent.ID == id {
			return agent.Name
		}
	}
	// Fallback: title-case the ID
	return strings.ReplaceAll(strings.Title(strings.ReplaceAll(id, "-", " ")), " ", "-")
}