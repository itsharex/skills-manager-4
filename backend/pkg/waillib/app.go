// Package waillib provides a public bridge layer for Wails to access internal packages.
package waillib

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/clawhub"
	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/skillsmanager/skillsmanager/backend/internal/source"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application struct with bridge methods for the frontend.
type App struct {
	ctx          context.Context
	repo         *storage.Repository
	index        *storage.Index
	lock         *storage.LockFile
	installer    *distribute.Installer
	clawhubMgr   *clawhub.Manager // persistent ClawHub manager with cache
}

// ListedSkill is a flattened skill entry for the frontend skill list.
// Skills with the same name from different agents are merged into one entry.
type ListedSkill struct {
	Name        string   `json:"name"`
	AgentIDs    []string `json:"agentIds"`
	AgentNames  []string `json:"agentNames"`
	Paths       []string `json:"paths"`
	StorePath   string   `json:"storePath,omitempty"` // pool storage path (e.g. ~/.skill-pool/<name>/)
	Latest      string   `json:"latest"`
	Versions    []string `json:"versions"`
	Description string   `json:"description"`
	InPool      bool     `json:"inPool"`
}

// AgentInfo describes a detected or known AI agent.
type AgentInfo struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Path                string `json:"path"`
	SkillsDir           string `json:"skillsDir"`
	ProjectSkillsSubdir string `json:"projectSkillsSubdir"`
	Detected            bool   `json:"detected"`
}

// InstallUIOptions configures a skill installation from the frontend.
type InstallUIOptions struct {
	Namespace string   `json:"namespace"`
	Version   string   `json:"version"`
	Agents    []string `json:"agents"`
	NoSync    bool     `json:"noSync"`
}

// InstallUILog records the outcome of a single skill installation.
type InstallUILog struct {
	SkillName string `json:"skillName"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Error     string `json:"error,omitempty"`
}

// NewApp creates a new App instance and initializes backend services.
func NewApp() *App {
	poolPath := operations.DefaultPoolPath()
	paths := operations.GetPoolPaths(poolPath)
	repo := storage.NewRepository(poolPath)

	index, err := storage.NewIndex(paths.IndexPath)
	if err != nil {
		println("Warning: failed to load index:", err.Error())
	}
	lock, err := storage.NewLockFile(paths.LockPath)
	if err != nil {
		println("Warning: failed to load lock file:", err.Error())
	}

	// Initialize operation logger
	if err := operations.InitOpLogger(poolPath); err != nil {
		println("Warning: failed to init op logger:", err.Error())
	}

	return &App{
		repo:      repo,
		index:     index,
		lock:      lock,
		installer: distribute.NewInstaller(repo, index, lock),
		clawhubMgr: clawhub.New(poolPath),
	}
}

// Startup stores the context.
func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

// Shutdown is called when the app shuts down.
func (a *App) Shutdown(ctx context.Context) {}

// GetConfig loads and returns the current configuration.
func (a *App) GetConfig() models.Config {
	paths := operations.GetPoolPaths(operations.DefaultPoolPath())
	cfg, err := operations.LoadConfig(paths.ConfigPath)
	if err != nil {
		return models.Config{}
	}
	// Sync GitHub token to environment on first load
	if cfg.GitHubToken != "" {
		os.Setenv("GITHUB_TOKEN", cfg.GitHubToken)
	}
	return *cfg
}

// SaveConfig persists the configuration to disk.
func (a *App) SaveConfig(cfg models.Config) error {
	paths := operations.GetPoolPaths(operations.DefaultPoolPath())
	err := operations.SaveConfig(paths.ConfigPath, &cfg)
	if err != nil {
		return err
	}
	// Sync GitHub token to environment variable for API access
	if cfg.GitHubToken != "" {
		os.Setenv("GITHUB_TOKEN", cfg.GitHubToken)
	} else {
		os.Unsetenv("GITHUB_TOKEN")
	}
	return nil
}

// ListSkills returns all skills found on the machine, merged by name.
// Skills with the same name from different agents are merged into one entry
// with multiple paths and agent IDs.
func (a *App) ListSkills() []ListedSkill {
	inPool := a.poolSkillDirSet()
	merged := make(map[string]*ListedSkill) // key = skill name

	// 1. Skills from the index (installed via skill install / market install)
	if a.index != nil {
		entries, err := a.index.List()
		if err == nil {
			for _, e := range entries {
				if _, ok := merged[e.Name]; !ok {
					merged[e.Name] = &ListedSkill{
						Name:        e.Name,
						AgentIDs:    []string{},
						AgentNames:  []string{},
						Paths:       []string{},
						Latest:      e.Latest,
						Versions:    e.Versions,
						Description: e.Description,
						InPool:      inPool[e.Name],
					}
				}
				// Set StorePath from the pool if available
				skill := merged[e.Name]
				if skill.StorePath == "" {
					cfg := a.GetConfig()
					poolSkillPath := filepath.Join(cfg.PoolPath, e.Name)
					if _, err := os.Stat(poolSkillPath); err == nil {
						skill.StorePath = poolSkillPath
					}
				}
			}
		}
	}

	// 2. Skills found in detected agents' global skill directories
	cfg := a.GetConfig()
	for _, ag := range distribute.DetectedAgents() {
		entries, err := os.ReadDir(ag.SkillsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(ag.SkillsDir, entry.Name())
			name := entry.Name()
			// Resolve symlink to determine StorePath (pool location)
			storePath := ""
			if linkDest, err := os.Readlink(skillPath); err == nil {
				if !filepath.IsAbs(linkDest) {
					linkDest, _ = filepath.Abs(filepath.Join(filepath.Dir(skillPath), linkDest))
				}
				storePath = linkDest
			} else {
				// Not a symlink — check if this skill exists in Pool
				poolSkillPath := filepath.Join(cfg.PoolPath, name)
				if _, err := os.Stat(poolSkillPath); err == nil {
					storePath = poolSkillPath
				}
			}
			if existing, ok := merged[name]; ok {
				// Merge: add this agent's path
				existing.AgentIDs = append(existing.AgentIDs, ag.ID)
				existing.AgentNames = append(existing.AgentNames, ag.Name)
				existing.Paths = append(existing.Paths, skillPath)
				if existing.StorePath == "" && storePath != "" {
					existing.StorePath = storePath
				}
			} else {
				merged[name] = &ListedSkill{
					Name:        name,
					AgentIDs:    []string{ag.ID},
					AgentNames:  []string{ag.Name},
					Paths:       []string{skillPath},
					StorePath:   storePath,
					Latest:      "",
					Versions:    nil,
					Description: "",
					InPool:      inPool[name],
				}
			}
		}
	}

	// 3. Skills in Pool that are not in any agent or index
	// (e.g. market-installed skills that haven't been symlinked to agents yet)
	poolEntries, err := os.ReadDir(cfg.PoolPath)
	if err == nil {
		for _, entry := range poolEntries {
			if !entry.IsDir() || entry.Name() == ".meta" {
				continue
			}
			name := entry.Name()
			skillMDPath := filepath.Join(cfg.PoolPath, name, "SKILL.md")
			if _, err := os.Stat(skillMDPath); err != nil {
				continue
			}
			if _, ok := merged[name]; ok {
				// Already found via index or agent — just ensure InPool is set
				merged[name].InPool = true
				continue
			}
			// Skill exists in Pool but not in any agent — still show it
			merged[name] = &ListedSkill{
				Name:        name,
				AgentIDs:    []string{},
				AgentNames:  []string{},
				Paths:       []string{},
				StorePath:   filepath.Join(cfg.PoolPath, name),
				Latest:      "",
				Versions:    nil,
				Description: "",
				InPool:      true,
			}
		}
	}

	// Convert map to sorted slice
	result := make([]ListedSkill, 0, len(merged))
	for _, s := range merged {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListAgents detects agents on the filesystem and returns their info, sorted by detection status.
func (a *App) ListAgents() []AgentInfo {
	all := distribute.DetectedAgents()
	var detectedList, notDetectedList []AgentInfo
	for _, ag := range all {
		info := AgentInfo{ID: ag.ID, Name: ag.Name, Path: ag.SkillsDir, SkillsDir: ag.SkillsDir, ProjectSkillsSubdir: ag.ProjectSkillsSubdir, Detected: ag.AutoDetected}
		if ag.AutoDetected {
			detectedList = append(detectedList, info)
		} else {
			notDetectedList = append(notDetectedList, info)
		}
	}
	return append(detectedList, notDetectedList...)
}

// Install resolves a source string and installs skills.
func (a *App) Install(sourceStr string, opts InstallUIOptions) ([]InstallUILog, error) {
	resolver, err := source.NewResolver(sourceStr)
	if err != nil {
		return nil, fmt.Errorf("create resolver: %w", err)
	}
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}
	skills, err := resolver.Resolve(ctx, sourceStr, source.ResolveOptions{Version: opts.Version})
	if err != nil {
		return nil, fmt.Errorf("resolve source: %w", err)
	}
	var logs []InstallUILog
	for _, sk := range skills {
		result, err := a.installer.Install(sk, distribute.InstallOptions{
			Namespace: opts.Namespace, Version: opts.Version,
			Agents: opts.Agents, NoSync: opts.NoSync,
		})
		log := InstallUILog{SkillName: sk.Name, Version: sk.Version}
		if result != nil {
			log.Version = result.Version
			log.Path = result.StorePath
		}
		if err != nil {
			log.Error = err.Error()
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// Uninstall removes a skill from the repository and all agents.
func (a *App) Uninstall(name, version string) error {
	return a.installer.Uninstall("local", name, version)
}

// Search resolves a source string and returns available skills.
func (a *App) Search(sourceStr string) ([]models.ResolvedSkill, error) {
	resolver, err := source.NewResolver(sourceStr)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}
	return resolver.Resolve(ctx, sourceStr, source.ResolveOptions{})
}

// SearchMarket performs an aggregated search across all market sources.
// Built-in sources (ClawHub, skills.sh, local pool) are always searched.
// Configurable market sources (GitHub repos, registries) are searched if enabled.
// Results are grouped by source type. A 30-second timeout prevents UI freeze.
func (a *App) SearchMarket(keyword string) []models.MarketSearchResult {
	cfg := a.GetConfig()
	poolPath := cfg.PoolPath
	if poolPath == "" {
		poolPath = operations.DefaultPoolPath()
	}

	// Use a 30-second timeout to prevent UI freeze from slow network calls.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searcher := source.NewMarketSearcherWithClawHub(poolPath, cfg.MarketSources, a.clawhubMgr)
	return searcher.SearchAll(ctx, keyword)
}

// InstallMarketSkill installs a skill from market search results.
// It handles different source types:
//   - pool: skill is already local, symlink to agent dirs
//   - clawhub/skillssh/github/registry: resolve from remote source and install
func (a *App) InstallMarketSkill(skill models.MarketSearchSkill, agentIDs []string) ([]InstallUILog, error) {
	if len(agentIDs) == 0 {
		return nil, fmt.Errorf("请先选择目标智能体")
	}

	agentsStr := strings.Join(agentIDs, ",")
	operations.LogOp("install", skill.Name,
		fmt.Sprintf("市场安装: %s (来源: %s/%s)", skill.Name, skill.Namespace, skill.Source),
		skill.Source, "", agentsStr, false, "")

	// Pool skills: already local, just symlink to agent dirs
	if skill.LocalPath != "" {
		var logs []InstallUILog
		for _, agentID := range agentIDs {
			skillsDir, err := distribute.GetAgentSkillsDir(agentID)
			if err != nil {
				operations.LogOp("install", skill.Name, fmt.Sprintf("安装到 %s 失败", agentID), skill.LocalPath, "", agentID, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skill.Name, Version: skill.Version, Error: err.Error()})
				continue
			}
			destPath := filepath.Join(skillsDir, filepath.Base(skill.LocalPath))
			if err := a.InstallToAgent(skill.LocalPath, skillsDir, true); err != nil {
				operations.LogOp("install", skill.Name, fmt.Sprintf("symlink %s -> %s 失败", skill.LocalPath, destPath), skill.LocalPath, destPath, agentID, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skill.Name, Version: skill.Version, Error: err.Error()})
				continue
			}
			operations.LogOp("install", skill.Name, fmt.Sprintf("symlink %s -> %s", skill.LocalPath, destPath), skill.LocalPath, destPath, agentID, true, "")
			logs = append(logs, InstallUILog{SkillName: skill.Name, Version: skill.Version, Path: destPath})
		}
		return logs, nil
	}

	// Remote skills: resolve source URL, copy to Pool, then symlink to agents
	sourceURL := buildMarketSourceURL(skill)
	if sourceURL == "" {
		operations.LogOp("install", skill.Name, "无法确定技能来源", skill.Source, "", agentsStr, false, "no source URL")
		return nil, fmt.Errorf("无法确定技能来源: %s", skill.Name)
	}

	operations.LogOp("install", skill.Name, fmt.Sprintf("解析远程源: %s", sourceURL), sourceURL, "", agentsStr, false, "")

	resolver, err := source.NewResolver(sourceURL)
	if err != nil {
		operations.LogOp("install", skill.Name, fmt.Sprintf("创建 resolver 失败: %s", sourceURL), sourceURL, "", agentsStr, false, err.Error())
		return nil, fmt.Errorf("create resolver: %w", err)
	}
	ctx := context.Background()
	if a.ctx != nil {
		ctx = a.ctx
	}
	resolved, err := resolver.Resolve(ctx, sourceURL, source.ResolveOptions{Version: skill.Version})
	if err != nil {
		operations.LogOp("install", skill.Name, fmt.Sprintf("解析源失败: %s", sourceURL), sourceURL, "", agentsStr, false, err.Error())
		return nil, fmt.Errorf("resolve source: %w", err)
	}

	// Filter resolved skills: only install the one matching the user's selection
	var matched []models.ResolvedSkill
	for _, sk := range resolved {
		if strings.EqualFold(sk.Name, skill.Name) {
			matched = append(matched, sk)
		}
	}
	// If no exact match found, install all resolved skills (fallback for single-skill repos)
	if len(matched) == 0 {
		matched = resolved
	}

	cfg := a.GetConfig()
	poolPath := cfg.PoolPath
	if poolPath == "" {
		poolPath = operations.DefaultPoolPath()
	}

	var logs []InstallUILog
	for _, sk := range matched {
		skillName := sk.Name
		poolDest := filepath.Join(poolPath, skillName)

		// If already in pool, skip copy but still symlink to agents
		if _, err := os.Stat(poolDest); err != nil {
			// Copy resolved skill to Pool (~/.skill-pool/<name>/)
			if err := os.MkdirAll(poolPath, 0755); err != nil {
				operations.LogOp("install", skillName, fmt.Sprintf("创建 Pool 目录失败: %s", poolPath), sourceURL, poolDest, agentsStr, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skillName, Version: sk.Version, Error: err.Error()})
				continue
			}
			if err := copyDir(sk.LocalPath, poolDest); err != nil {
				operations.LogOp("install", skillName, fmt.Sprintf("复制到 Pool 失败: %s -> %s", sk.LocalPath, poolDest), sourceURL, poolDest, agentsStr, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skillName, Version: sk.Version, Error: err.Error()})
				continue
			}
			operations.LogOp("install", skillName, fmt.Sprintf("复制到 Pool: %s -> %s", sk.LocalPath, poolDest), sourceURL, poolDest, "", true, "")
		} else {
			operations.LogOp("install", skillName, fmt.Sprintf("已在 Pool 中: %s", poolDest), sourceURL, poolDest, "", true, "")
		}

		// Symlink from Pool to each agent's skills directory
		for _, agentID := range agentIDs {
			skillsDir, err := distribute.GetAgentSkillsDir(agentID)
			if err != nil {
				operations.LogOp("install", skillName, fmt.Sprintf("获取智能体目录失败: %s", agentID), sourceURL, poolDest, agentID, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skillName, Version: sk.Version, Error: err.Error()})
				continue
			}
			destPath := filepath.Join(skillsDir, skillName)
			if err := a.InstallToAgent(poolDest, skillsDir, true); err != nil {
				operations.LogOp("install", skillName, fmt.Sprintf("symlink %s -> %s 失败", poolDest, destPath), sourceURL, destPath, agentID, false, err.Error())
				logs = append(logs, InstallUILog{SkillName: skillName, Version: sk.Version, Error: err.Error()})
				continue
			}
			operations.LogOp("install", skillName, fmt.Sprintf("symlink %s -> %s", poolDest, destPath), sourceURL, destPath, agentID, true, "")
			logs = append(logs, InstallUILog{SkillName: skillName, Version: sk.Version, Path: destPath})
		}
	}
	// Clean up temp dir after all installs (all resolved skills share the same tmpDir)
	for _, sk := range resolved {
		if sk.Cleanup != nil {
			sk.Cleanup()
			break // only need to call once since they all remove the same tmpDir
		}
	}
	return logs, nil
}

// buildMarketSourceURL constructs a resolvable source URL from a MarketSearchSkill.
func buildMarketSourceURL(skill models.MarketSearchSkill) string {
	// For skills.sh: source is "owner/repo/slug", extract "owner/repo"
	if strings.HasPrefix(skill.Namespace, "skillssh:") {
		repo := strings.TrimPrefix(skill.Namespace, "skillssh:")
		return "https://github.com/" + repo
	}
	// For ClawHub: source is "owner/slug"
	if strings.HasPrefix(skill.Namespace, "clawhub:") {
		owner := strings.TrimPrefix(skill.Namespace, "clawhub:")
		if skill.Source != "" {
			return "https://github.com/" + skill.Source
		}
		return "https://github.com/" + owner
	}
	// For GitHub: source is "owner/repo"
	if strings.HasPrefix(skill.Namespace, "github:") {
		repo := strings.TrimPrefix(skill.Namespace, "github:")
		return "https://github.com/" + repo
	}
	// Fallback: use source field directly
	if skill.Source != "" {
		if strings.HasPrefix(skill.Source, "http") {
			return skill.Source
		}
		return "https://github.com/" + skill.Source
	}
	return ""
}

// GetStats collects aggregated statistics based on actual scanned data.
func (a *App) GetStats() operations.SkillStats {
	skills := a.ListSkills()
	agents := a.ListAgents()
	detected := 0
	for _, ag := range agents {
		if ag.Detected {
			detected++
		}
	}
	inPool := 0
	for _, s := range skills {
		if s.InPool {
			inPool++
		}
	}
	return operations.SkillStats{
		TotalSkills:     len(skills),
		TotalVersions:   0,
		TotalNamespaces: 0,
		TotalAgents:     detected,
		InstalledSkills: inPool,
		DiskUsageBytes:  0,
	}
}

// RunDoctor runs all diagnostic checks.
func (a *App) RunDoctor() operations.HealthReport {
	r := operations.RunDoctor(operations.DefaultPoolPath())
	if r == nil {
		return operations.HealthReport{}
	}
	return *r
}

// GetOpLogs returns the last N operation log entries.
func (a *App) GetOpLogs(n int) []models.OpLog {
	return operations.GetOpLogs(n)
}

// DiscoveredSkill represents a skill found during scanning.
type DiscoveredSkill struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Version       string `json:"version"`
	Path          string `json:"path"`
	AgentID       string `json:"agentId,omitempty"`
	AgentName     string `json:"agentName,omitempty"`
	AlreadyInPool bool   `json:"alreadyInPool"`
}

// ListPool reads the configured local pool directory,
// returning all skills found. Market-installed skills are now stored directly
// in the pool, so only the pool directory needs to be scanned.
func (a *App) ListPool() []DiscoveredSkill {
	cfg := a.GetConfig()
	results := make([]DiscoveredSkill, 0)

	// Scan pool directory (~/.skill-pool/)
	entries, err := os.ReadDir(cfg.PoolPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(cfg.PoolPath, entry.Name())
			skillMDPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMDPath); err != nil {
				continue
			}
			results = append(results, DiscoveredSkill{
				Name:          entry.Name(),
				Namespace:     "pool",
				Version:       "",
				Path:          skillPath,
				AlreadyInPool: true,
			})
		}
	}

	return results
}

// poolSkillDirSet returns a set of skill names whose directories exist in the configured pool path.
// Used by ScanLocal to determine alreadyInPool status.
func (a *App) poolSkillDirSet() map[string]bool {
	cfg := a.GetConfig()
	set := make(map[string]bool)
	entries, err := os.ReadDir(cfg.PoolPath)
	if err != nil {
		return set
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMDPath := filepath.Join(cfg.PoolPath, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMDPath); err == nil {
			set[entry.Name()] = true
		}
	}
	return set
}

// ScanLocal scans the local machine for skills. If projectPath is empty, scans all agent
// global skill directories. If projectPath is set, scans known agent skill names under
// the project path and combines with global scan results.
// Results are cross-referenced against the configured pool directory — AlreadyInPool=true
// means the skill exists in the pool.
func (a *App) ScanLocal(projectPath string) []DiscoveredSkill {
	if projectPath == "" {
		return a.scanGlobalPool()
	}
	return a.scanProjectPool(projectPath)
}

// poolNameSet builds a set of skill names currently registered in the index.
func (a *App) poolNameSet() map[string]bool {
	set := make(map[string]bool)
	if a.index == nil {
		return set
	}
	entries, err := a.index.List()
	if err != nil {
		return set
	}
	for _, e := range entries {
		set[e.Name] = true
	}
	return set
}

// scanGlobalPool scans all KnownAgents' SkillsDir directories. Returns all discovered
// skills with AlreadyInPool set based on pool directory contents. Deduplicates by physical path.
func (a *App) scanGlobalPool() []DiscoveredSkill {
	results := make([]DiscoveredSkill, 0)
	seenPaths := make(map[string]bool)
	inPool := a.poolSkillDirSet()

	for _, ag := range distribute.KnownAgents() {
		info, err := os.Stat(ag.SkillsDir)
		if err != nil || !info.IsDir() {
			continue
		}
		entries, err := os.ReadDir(ag.SkillsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(ag.SkillsDir, entry.Name())
			if seenPaths[skillPath] {
				continue
			}
			seenPaths[skillPath] = true
			skillMDPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMDPath); err != nil {
				continue
			}
			_, alreadyIn := inPool[entry.Name()]
			results = append(results, DiscoveredSkill{
				Name:          entry.Name(),
				Namespace:     "agent:" + ag.ID,
				Version:       "",
				Path:          skillPath,
				AgentID:       ag.ID,
				AgentName:     ag.Name,
				AlreadyInPool: alreadyIn,
			})
		}
	}
	return results
}

// scanProjectPool scans known agent project-level skill directories under the given project path.
// It only looks in standard agent subdirectories (e.g. .claude/skills/, .cursor/skills/),
// NOT by walking the entire directory tree. This avoids false positives from non-skill
// directories that happen to contain SKILL.md files.
// alreadyInPool is determined by resolving symlinks (points to pool) or name matching.
func (a *App) scanProjectPool(projectPath string) []DiscoveredSkill {
	cfg := a.GetConfig()
	poolPath := cfg.PoolPath
	inPool := a.poolSkillDirSet()
	results := make([]DiscoveredSkill, 0)
	seenKeys := make(map[string]bool) // deduplicate by "agentID:skillName"

	for _, ag := range distribute.KnownAgents() {
		if ag.ProjectSkillsSubdir == "" {
			continue
		}
		agentSkillsDir := filepath.Join(projectPath, ag.ProjectSkillsSubdir)
		entries, err := os.ReadDir(agentSkillsDir)
		if err != nil {
			continue // directory doesn't exist, skip
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(agentSkillsDir, entry.Name())
			skillMDPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMDPath); err != nil {
				continue // no SKILL.md, not a valid skill
			}

			skillName := entry.Name()
			key := ag.ID + ":" + skillName
			if seenKeys[key] {
				continue
			}
			seenKeys[key] = true

			// Determine alreadyInPool: resolve symlink to check if it points to pool
			alreadyInPool := false
			if linkDest, err := os.Readlink(skillPath); err == nil {
				if !filepath.IsAbs(linkDest) {
					linkDest, _ = filepath.Abs(filepath.Join(filepath.Dir(skillPath), linkDest))
				}
				if poolPath != "" && strings.HasPrefix(linkDest, poolPath) {
					alreadyInPool = true
				}
			}
			// Fallback: name-based check if not already determined
			if !alreadyInPool && inPool[skillName] {
				alreadyInPool = true
			}

			results = append(results, DiscoveredSkill{
				Name:          skillName,
				Namespace:     "project:" + filepath.Base(projectPath),
				Version:       "",
				Path:          skillPath,
				AgentID:       ag.ID,
				AgentName:     ag.Name,
				AlreadyInPool: alreadyInPool,
			})
		}
	}
	return results
}

// isSkillInPool checks if a skill name is already registered in the index.
func (a *App) isSkillInPool(name string) bool {
	return a.poolNameSet()[name]
}

// ImportToPool copies a skill directory into the configured pool path.
// The skill directory must contain a SKILL.md file.
func (a *App) ImportToPool(sourcePath string) error {
	cfg := a.GetConfig()
	if cfg.PoolPath == "" {
		return fmt.Errorf("pool path is not configured")
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	skillMDPath := filepath.Join(sourcePath, "SKILL.md")
	if _, err := os.Stat(skillMDPath); err != nil {
		return fmt.Errorf("source path does not contain SKILL.md: %w", err)
	}

	skillName := filepath.Base(sourcePath)
	destDir := filepath.Join(cfg.PoolPath, skillName)

	// Check if already in pool
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("skill %q already exists in pool", skillName)
	}

	// Create pool directory if needed
	if err := os.MkdirAll(cfg.PoolPath, 0755); err != nil {
		return fmt.Errorf("create pool directory: %w", err)
	}

	// Copy the skill directory recursively using a helper
	if err := copyDir(sourcePath, destDir); err != nil {
		return fmt.Errorf("copy skill to pool: %w", err)
	}

	return nil
}

// DeleteSkill removes a skill from a specific agent's skills directory.
// Handles both regular directories and symlinks (does not follow symlinks).
func (a *App) DeleteSkill(skillPath string) error {
	info, err := os.Lstat(skillPath)
	if err != nil {
		return fmt.Errorf("skill path does not exist: %w", err)
	}
	// Accept both directories and symlinks (which may point to directories)
	if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("skill path is not a directory or symlink")
	}
	return os.RemoveAll(skillPath)
}

// ArchiveToPool copies a skill into the pool directory.
// If the source is a symlink pointing to the pool, it's already archived — skip.
// For agent directory skills (non-symlink), copies to pool and removes the source.
func (a *App) ArchiveToPool(sourcePath string) error {
	cfg := a.GetConfig()
	if cfg.PoolPath == "" {
		return fmt.Errorf("pool path is not configured")
	}

	// If sourcePath is a symlink, resolve it to check if it points to pool
	linkDest, err := os.Readlink(sourcePath)
	if err == nil {
		// It's a symlink — resolve to absolute path
		if !filepath.IsAbs(linkDest) {
			linkDest, _ = filepath.Abs(filepath.Join(filepath.Dir(sourcePath), linkDest))
		}
		// If symlink already points to pool, nothing to do
		if strings.HasPrefix(linkDest, cfg.PoolPath) {
			operations.LogOp("archive", filepath.Base(sourcePath), fmt.Sprintf("已在池中 (symlink -> %s)", linkDest), sourcePath, linkDest, "", true, "")
			return nil
		}
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	skillMDPath := filepath.Join(sourcePath, "SKILL.md")
	if _, err := os.Stat(skillMDPath); err != nil {
		// Skill may not have SKILL.md, still allow archiving the directory
	}

	// Determine skill name: strip @version suffix for repo skills
	dirName := filepath.Base(sourcePath)
	skillName := dirName
	if idx := strings.LastIndex(dirName, "@"); idx > 0 {
		skillName = dirName[:idx]
	}

	destDir := filepath.Join(cfg.PoolPath, skillName)

	// If already in pool, skip copy
	if _, err := os.Stat(destDir); err == nil {
		operations.LogOp("archive", skillName, fmt.Sprintf("已在池中: %s", destDir), sourcePath, destDir, "", true, "")
		return nil
	}

	// Create pool directory if needed
	if err := os.MkdirAll(cfg.PoolPath, 0755); err != nil {
		return fmt.Errorf("create pool directory: %w", err)
	}

	// Copy to pool
	if err := copyDir(sourcePath, destDir); err != nil {
		operations.LogOp("archive", skillName, fmt.Sprintf("归档失败: %s -> %s", sourcePath, destDir), sourcePath, destDir, "", false, err.Error())
		return fmt.Errorf("copy skill to pool: %w", err)
	}

	// Remove source (it's from an agent directory, not repo)
	if err := os.RemoveAll(sourcePath); err != nil {
		println("Warning: failed to remove source after archiving:", err.Error())
	}

	operations.LogOp("archive", skillName, fmt.Sprintf("归档入池: %s -> %s", sourcePath, destDir), sourcePath, destDir, "", true, "")
	return nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// OpenDirectoryDialog shows a native directory picker dialog.
// ShowHiddenFiles is enabled so users can select hidden directories like ~/.skill-pool/.
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:               title,
		CanCreateDirectories: true,
		ShowHiddenFiles:     true,
	})
}

// OpenFileDialog shows a native file picker dialog.
func (a *App) OpenFileDialog(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:           title,
		ShowHiddenFiles: true,
	})
}

// OpenDirectory opens a directory in the system's file manager.
// Cross-platform: macOS uses 'open', Windows uses 'explorer', Linux uses 'xdg-open'.
func (a *App) OpenDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dirPath)
	}
	cmd := exec.CommandContext(a.ctx, "open", dirPath)
	return cmd.Start()
}

// OpenURL opens a URL in the system default browser.
func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

// DeleteSkillFromPool removes a skill from the pool and ALL agent directories that reference it.
// This is a machine-level delete: removes pool files + all agent global symlinks/copies + all project-level copies.
// skillName is the skill directory name under the pool (e.g. "gh-cli").
func (a *App) DeleteSkillFromPool(skillName string) error {
	cfg := a.GetConfig()
	if cfg.PoolPath == "" {
		return fmt.Errorf("pool path is not configured")
	}

	poolSkillPath := filepath.Join(cfg.PoolPath, skillName)

	// Verify the skill exists in pool
	if _, err := os.Stat(poolSkillPath); err != nil {
		return fmt.Errorf("skill %q not found in pool: %w", skillName, err)
	}

	// 1. Remove from all agent global directories (symlinks or copies)
	for _, ag := range distribute.DetectedAgents() {
		agentSkillPath := filepath.Join(ag.SkillsDir, skillName)
		if _, err := os.Lstat(agentSkillPath); err != nil {
			continue
		}
		if err := os.RemoveAll(agentSkillPath); err != nil {
			operations.LogOp("delete", skillName, fmt.Sprintf("从智能体 %s 全局删除失败: %s", ag.Name, agentSkillPath), "", agentSkillPath, ag.ID, false, err.Error())
		} else {
			operations.LogOp("delete", skillName, fmt.Sprintf("从智能体 %s 全局删除: %s", ag.Name, agentSkillPath), "", agentSkillPath, ag.ID, true, "")
		}
	}

	// 2. Remove from pool
	if err := os.RemoveAll(poolSkillPath); err != nil {
		operations.LogOp("delete", skillName, fmt.Sprintf("从 Pool 删除失败: %s", poolSkillPath), "", poolSkillPath, "", false, err.Error())
		return fmt.Errorf("remove skill from pool: %w", err)
	}

	operations.LogOp("delete", skillName, fmt.Sprintf("从 Pool 删除（机器级）: %s", poolSkillPath), "", poolSkillPath, "", true, "")
	return nil
}

// DeleteSkillFromAgent removes a skill from a specific agent's global skills directory only.
// It does NOT remove the skill from the pool. skillPath is the full path in the agent's dir.
func (a *App) DeleteSkillFromAgent(skillPath string) error {
	if _, err := os.Lstat(skillPath); err != nil {
		return fmt.Errorf("skill path does not exist: %w", err)
	}
	skillName := filepath.Base(skillPath)
	if err := os.RemoveAll(skillPath); err != nil {
		operations.LogOp("delete", skillName, fmt.Sprintf("从智能体全局删除失败: %s", skillPath), "", skillPath, "", false, err.Error())
		return fmt.Errorf("remove skill from agent: %w", err)
	}
	operations.LogOp("delete", skillName, fmt.Sprintf("从智能体全局删除: %s", skillPath), "", skillPath, "", true, "")
	return nil
}

// DeleteSkillFromProject removes a skill from a project-level directory only.
// projectSkillPath is the full path to the skill in the project directory.
func (a *App) DeleteSkillFromProject(projectSkillPath string) error {
	if _, err := os.Lstat(projectSkillPath); err != nil {
		return fmt.Errorf("skill path does not exist: %w", err)
	}
	skillName := filepath.Base(projectSkillPath)
	if err := os.RemoveAll(projectSkillPath); err != nil {
		operations.LogOp("delete", skillName, fmt.Sprintf("从项目删除失败: %s", projectSkillPath), "", projectSkillPath, "", false, err.Error())
		return fmt.Errorf("remove skill from project: %w", err)
	}
	operations.LogOp("delete", skillName, fmt.Sprintf("从项目删除: %s", projectSkillPath), "", projectSkillPath, "", true, "")
	return nil
}

// InstallToAgent creates a symlink from the pool skill to the agent's skills directory.
// On Unix (macOS/Linux) os.Symlink is used directly. On Windows, symlink creation
// may require admin privileges; if it fails, falls back to copying the directory.
// If overwrite is true and the destination already exists, it is removed first.
func (a *App) InstallToAgent(skillPath, agentSkillsDir string, overwrite bool) error {
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return fmt.Errorf("skill does not exist: %s", skillPath)
	}
	info, err := os.Stat(skillPath)
	if err != nil {
		return fmt.Errorf("cannot access skill: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill path is not a directory")
	}

	if _, err := os.Stat(agentSkillsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(agentSkillsDir, 0755); err != nil {
			return fmt.Errorf("create agent skills directory: %w", err)
		}
	}

	skillName := filepath.Base(skillPath)
	destPath := filepath.Join(agentSkillsDir, skillName)

	if _, err := os.Lstat(destPath); err == nil {
		if !overwrite {
			return fmt.Errorf("skill already installed: %s", destPath)
		}
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("remove existing skill: %w", err)
		}
	}

	// Try symlink first; fall back to copy on failure (e.g. Windows without privs)
	if err := os.Symlink(skillPath, destPath); err != nil {
		// Fallback: copy directory
		return copyDir(skillPath, destPath)
	}
	return nil
}

// GetAgentSkillsDir returns the skills directory for a given agent.
func (a *App) GetAgentSkillsDir(agentID string) (string, error) {
	return distribute.GetAgentSkillsDir(agentID)
}

// InstallToProject installs a skill (via symlink with copy fallback) into a project's .opencode/skills/ directory.
// The projectSkillsDir is <projectPath>/.opencode/skills/.
// If overwrite is true and the destination already exists, it is removed first.
func (a *App) InstallToProject(skillPath, projectPath string, overwrite bool) error {
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return fmt.Errorf("skill does not exist: %s", skillPath)
	}
	info, err := os.Stat(skillPath)
	if err != nil {
		return fmt.Errorf("cannot access skill: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill path is not a directory")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}

	projectSkillsDir := filepath.Join(projectPath, ".opencode", "skills")
	if err := os.MkdirAll(projectSkillsDir, 0755); err != nil {
		return fmt.Errorf("create project skills directory: %w", err)
	}

	skillName := filepath.Base(skillPath)
	destPath := filepath.Join(projectSkillsDir, skillName)

	if _, err := os.Lstat(destPath); err == nil {
		if !overwrite {
			return fmt.Errorf("skill already installed in project: %s", destPath)
		}
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("remove existing skill in project: %w", err)
		}
	}

	// Try symlink first; fall back to copy on failure
	if err := os.Symlink(skillPath, destPath); err != nil {
		return copyDir(skillPath, destPath)
	}
	return nil
}

// InstallToProjectForAgent installs a skill to a project directory using the
// agent-specific project-level skills subdirectory (e.g. ".claude/skills" for Claude Code).
func (a *App) InstallToProjectForAgent(skillPath, projectPath, agentID string, overwrite bool) error {
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return fmt.Errorf("skill does not exist: %s", skillPath)
	}
	info, err := os.Stat(skillPath)
	if err != nil {
		return fmt.Errorf("cannot access skill: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill path is not a directory")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}

	subdir, err := distribute.GetProjectSkillsDir(agentID)
	if err != nil {
		return fmt.Errorf("get project skills dir for agent %s: %w", agentID, err)
	}

	projectSkillsDir := filepath.Join(projectPath, subdir)
	if err := os.MkdirAll(projectSkillsDir, 0755); err != nil {
		return fmt.Errorf("create project skills directory: %w", err)
	}

	skillName := filepath.Base(skillPath)
	destPath := filepath.Join(projectSkillsDir, skillName)

	if _, err := os.Lstat(destPath); err == nil {
		if !overwrite {
			return fmt.Errorf("skill already installed in project: %s", destPath)
		}
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("remove existing skill in project: %w", err)
		}
	}

	// Try symlink first; fall back to copy on failure
	if err := os.Symlink(skillPath, destPath); err != nil {
		return copyDir(skillPath, destPath)
	}
	return nil
}
