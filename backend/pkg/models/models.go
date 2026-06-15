package models

// SkillID uniquely identifies a skill: namespace:name@version
type SkillID struct {
	Namespace string `json:"namespace"` // github:owner/repo / local / registry:name
	Name      string `json:"name"`
	Version   string `json:"version"`
}

// Config - ~/.skill-repo/config.json
type Config struct {
	RepoPath      string         `json:"repo_path"`      // default ~/.skill-repo/
	PoolPath      string         `json:"pool_path"`      // local skill pool directory, default ~/.skill-pool/
	InstallMode   string         `json:"install_mode"`   // "symlink" | "copy"
	AutoFallback  bool           `json:"auto_fallback"`  // symlink fail -> copy
	DefaultAgents []string       `json:"default_agents"`
	MarketSources []MarketSource `json:"market_sources"` // configured market/search sources
	LinkTargets   []LinkTarget   `json:"link_targets"`
	Repositories  []RepoSource   `json:"repositories"`
	CacheTTL      int            `json:"cache_ttl"` // seconds
}

type MarketSource struct {
	Name    string `json:"name"`
	URL     string `json:"url"`     // local path, GitHub URL, or registry URL
	Type    string `json:"type"`    // "pool" | "github" | "registry"
	Enabled bool   `json:"enabled"`
	Branch  string `json:"branch,omitempty"` // GitHub branch, default "main"
}

type LinkTarget struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

type RepoSource struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Type    string `json:"type"` // "registry" | "github"
	Enabled bool   `json:"enabled"`
}

// Index - ~/.skill-repo/index.json
type Index struct {
	Version    int                   `json:"version"`
	LastUpdate string                `json:"last_update"`
	Skills     map[string]IndexEntry `json:"skills"`
}

type IndexEntry struct {
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	Versions      []string `json:"versions"`
	Latest        string   `json:"latest"`
	Source        string   `json:"source"`
	SourceType    string   `json:"source_type"`
	InstalledSize string   `json:"installed_size"`
	Tags          []string `json:"tags"`
	Description   string   `json:"description"`
}

// LockFile - ~/.skill-repo/lock.json
type LockFile struct {
	Version int                 `json:"version"`
	Skills  map[string]LockEntry `json:"skills"`
}

type LockEntry struct {
	SkillID     SkillID            `json:"skill_id"`
	InstalledAt string             `json:"installed_at"`
	Source      string             `json:"source"`
	Agents      []LockAgentBinding `json:"agents"`
}

type LockAgentBinding struct {
	AgentID string `json:"agent_id"`
	Path    string `json:"path"`
	Mode    string `json:"mode"` // "symlink" | "copy"
}

// ResolvedSkill - result from source resolvers
type ResolvedSkill struct {
	LocalPath string `json:"localPath"` // local temp path
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Cleanup   func() `json:"-"` // cleanup temp files
}

// ClawHubSkill represents a skill in the ClawHub registry.
type ClawHubSkill struct {
	Owner       string   `json:"owner"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
	Author      string   `json:"author"`
}

// RuntimeStatus reports the ClawHub runtime environment.
type RuntimeStatus struct {
	NodeInstalled    bool   `json:"node_installed"`
	NodePath         string `json:"node_path"`
	NodeVersion      string `json:"node_version"`
	HasNpm           bool   `json:"has_npm"`
	ClawHubInstalled bool   `json:"clawhub_installed"`
	ClawHubPath      string `json:"clawhub_path"`
	ClawHubVersion   string `json:"clawhub_version"`
	RegistryReachable bool  `json:"registry_reachable"`
	RegistryName     string `json:"registry_name"`
	Message          string `json:"message"`
}

// SkillInfo is a lightweight skill metadata struct.
type SkillInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MarketSearchResult is the aggregated search result from market search.
type MarketSearchResult struct {
	SourceName string              `json:"sourceName"`
	SourceType string              `json:"sourceType"` // "pool" | "clawhub" | "skillssh" | "github" | "registry"
	Skills     []MarketSearchSkill `json:"skills"`
	Error      string              `json:"error,omitempty"`
}

// MarketSearchSkill is a single skill in a market search result.
type MarketSearchSkill struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Source      string `json:"source"` // owner/repo for GitHub, owner/slug for ClawHub
	Installs    int    `json:"installs,omitempty"`
}

// RepoPaths helper
type RepoPaths struct {
	Root       string
	SkillsDir  string // Root/skills/
	IndexPath  string // Root/index.json
	LockPath   string // Root/lock.json
	ConfigPath string // Root/config.json
}