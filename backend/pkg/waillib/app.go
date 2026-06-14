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

	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/skillsmanager/skillsmanager/backend/internal/source"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application struct with bridge methods for the frontend.
type App struct {
	ctx       context.Context
	repo      *storage.Repository
	index     *storage.Index
	lock      *storage.LockFile
	installer *distribute.Installer
}

// ListedSkill is a flattened skill entry for the frontend skill list.
// Skills with the same name from different agents are merged into one entry.
type ListedSkill struct {
	Name        string   `json:"name"`
	AgentIDs    []string `json:"agentIds"`
	AgentNames  []string `json:"agentNames"`
	Paths       []string `json:"paths"`
	Latest      string   `json:"latest"`
	Versions    []string `json:"versions"`
	Description string   `json:"description"`
	InPool      bool     `json:"inPool"`
}

// AgentInfo describes a detected or known AI agent.
type AgentInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	SkillsDir string `json:"skillsDir"`
	Detected  bool   `json:"detected"`
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
	repoPath := operations.DefaultRepoPath()
	paths := operations.GetRepoPaths(repoPath)
	repo := storage.NewRepository(repoPath)

	index, err := storage.NewIndex(paths.IndexPath)
	if err != nil {
		println("Warning: failed to load index:", err.Error())
	}
	lock, err := storage.NewLockFile(paths.LockPath)
	if err != nil {
		println("Warning: failed to load lock file:", err.Error())
	}

	return &App{
		repo:      repo,
		index:     index,
		lock:      lock,
		installer: distribute.NewInstaller(repo, index, lock),
	}
}

// Startup stores the context.
func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

// Shutdown is called when the app shuts down.
func (a *App) Shutdown(ctx context.Context) {}

// GetConfig loads and returns the current configuration.
func (a *App) GetConfig() models.Config {
	paths := operations.GetRepoPaths(operations.DefaultRepoPath())
	cfg, err := operations.LoadConfig(paths.ConfigPath)
	if err != nil {
		return models.Config{}
	}
	return *cfg
}

// SaveConfig persists the configuration to disk.
func (a *App) SaveConfig(cfg models.Config) error {
	paths := operations.GetRepoPaths(operations.DefaultRepoPath())
	return operations.SaveConfig(paths.ConfigPath, &cfg)
}

// ListSkills returns all skills found on the machine, merged by name.
// Skills with the same name from different agents are merged into one entry
// with multiple paths and agent IDs.
func (a *App) ListSkills() []ListedSkill {
	inPool := a.poolSkillDirSet()
	merged := make(map[string]*ListedSkill) // key = skill name

	// 1. Skills from the index (installed via skill install)
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
			}
		}
	}

	// 2. Skills found in detected agents' global skill directories
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
			if existing, ok := merged[name]; ok {
				// Merge: add this agent's path
				existing.AgentIDs = append(existing.AgentIDs, ag.ID)
				existing.AgentNames = append(existing.AgentNames, ag.Name)
				existing.Paths = append(existing.Paths, skillPath)
			} else {
				merged[name] = &ListedSkill{
					Name:        name,
					AgentIDs:    []string{ag.ID},
					AgentNames:  []string{ag.Name},
					Paths:       []string{skillPath},
					Latest:      "",
					Versions:    nil,
					Description: "",
					InPool:      inPool[name],
				}
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
		info := AgentInfo{ID: ag.ID, Name: ag.Name, Path: ag.SkillsDir, SkillsDir: ag.SkillsDir, Detected: ag.AutoDetected}
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
	r := operations.RunDoctor(operations.DefaultRepoPath())
	if r == nil {
		return operations.HealthReport{}
	}
	return *r
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

// ListPool reads the configured local pool directory and returns all skills found there.
func (a *App) ListPool() []DiscoveredSkill {
	cfg := a.GetConfig()
	results := make([]DiscoveredSkill, 0)
	entries, err := os.ReadDir(cfg.PoolPath)
	if err != nil {
		return results
	}
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

// scanProjectPool recursively scans the given directory for skill definitions.
// It walks all subdirectories looking for SKILL.md files and records the
// directory relationship for later archiving to pool.
func (a *App) scanProjectPool(projectPath string) []DiscoveredSkill {
	inPool := a.poolSkillDirSet()
	results := make([]DiscoveredSkill, 0)
	seenPaths := make(map[string]bool)

	// Recursively walk the directory tree looking for SKILL.md
	filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}

		// WalkDir uses Lstat — symlinks to dirs are NOT IsDir. Resolve them manually.
		isDir := d.IsDir()
		if !isDir && d.Type()&os.ModeSymlink != 0 {
			if target, statErr := os.Stat(path); statErr == nil && target.IsDir() {
				isDir = true
			}
		}
		if !isDir {
			return nil
		}
		// Skip well-known non-skill directories only — do NOT skip all hidden dirs,
		// because agent skill directories like .opencode/skills/ are hidden.
		name := d.Name()
		if name == ".git" || name == "node_modules" || name == "__pycache__" ||
			name == "target" || name == ".idea" || name == ".vscode" ||
			name == ".vs" || name == "dist" || name == "build" ||
			name == ".next" || name == ".nuxt" || name == ".output" ||
			name == "vendor" || name == ".tox" || name == "venv" ||
			name == ".venv" || name == "env" || name == ".env" {
			return filepath.SkipDir
		}
		// Check if this directory contains SKILL.md
		skillMDPath := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillMDPath); err != nil {
			return nil // no SKILL.md here, keep walking
		}
		if seenPaths[path] {
			return nil
		}
		seenPaths[path] = true
		skillName := filepath.Base(path)
		results = append(results, DiscoveredSkill{
			Name:          skillName,
			Namespace:     "project:" + filepath.Base(projectPath),
			Version:       "",
			Path:          path,
			AlreadyInPool: inPool[skillName],
		})
		return nil
	})
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

// ArchiveToPool moves a skill from an agent directory into the pool (copy + delete source).
// This is different from ImportToPool which only copies.
func (a *App) ArchiveToPool(sourcePath string) error {
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
		// Skill may not have SKILL.md, still allow archiving the directory
	}

	skillName := filepath.Base(sourcePath)
	destDir := filepath.Join(cfg.PoolPath, skillName)

	// If already in pool, just delete the source
	if _, err := os.Stat(destDir); err == nil {
		return os.RemoveAll(sourcePath)
	}

	// Create pool directory if needed
	if err := os.MkdirAll(cfg.PoolPath, 0755); err != nil {
		return fmt.Errorf("create pool directory: %w", err)
	}

	// Copy to pool first
	if err := copyDir(sourcePath, destDir); err != nil {
		return fmt.Errorf("copy skill to pool: %w", err)
	}

	// Then remove the source
	if err := os.RemoveAll(sourcePath); err != nil {
		// Non-fatal: the copy succeeded, just log the error
		println("Warning: failed to remove source after archiving:", err.Error())
	}

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
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:               title,
		CanCreateDirectories: true,
	})
}

// OpenFileDialog shows a native file picker dialog.
func (a *App) OpenFileDialog(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
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

// DeleteSkillFromPool removes a skill directory from the pool path.
func (a *App) DeleteSkillFromPool(skillPath string) error {
	cfg := a.GetConfig()
	if cfg.PoolPath == "" {
		return fmt.Errorf("pool path is not configured")
	}
	// Only allow deleting from pool path
	if !strings.HasPrefix(skillPath, cfg.PoolPath) {
		return fmt.Errorf("can only delete skills from pool directory")
	}
	info, err := os.Stat(skillPath)
	if err != nil {
		return fmt.Errorf("skill path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill path is not a directory")
	}
	return os.RemoveAll(skillPath)
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
