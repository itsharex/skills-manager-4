package distribute

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Installer orchestrates skill installation and uninstallation.
type Installer struct {
	Repo      *storage.Repository
	Index     *storage.Index
	Lock      *storage.LockFile
	ForceCopy bool
}

// InstallOptions configures a skill installation.
type InstallOptions struct {
	Namespace string   // override namespace
	Version   string   // specific version to install
	Agents    []string // agents to install to
	ForceCopy bool     // force copy mode
	NoSync    bool     // skip agent sync (store only)
}

// InstallResult records the outcome of an installation.
type InstallResult struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	StorePath string `json:"store_path"`
	Synced    bool   `json:"synced"`
	SyncMode  string `json:"sync_mode,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewInstaller creates a new Installer instance.
func NewInstaller(repo *storage.Repository, idx *storage.Index, lock *storage.LockFile) *Installer {
	return &Installer{
		Repo:  repo,
		Index: idx,
		Lock:  lock,
	}
}

// buildIndexKey builds the index key: "namespace/name"
func buildIndexKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// buildLockKey builds the lock key: "namespace/name@version"
func buildLockKey(namespace, name, version string) string {
	return fmt.Sprintf("%s/%s@%s", namespace, name, version)
}

// Install installs a resolved skill into the repository and optionally syncs to agents.
func (inst *Installer) Install(skill models.ResolvedSkill, opts InstallOptions) (*InstallResult, error) {
	// Ensure temp resources from resolution are cleaned up after storing
	defer func() {
		if skill.Cleanup != nil {
			skill.Cleanup()
		}
	}()

	namespace := opts.Namespace
	if namespace == "" {
		namespace = skill.Namespace
	}
	version := opts.Version
	if version == "" {
		version = skill.Version
	}
	if version == "" {
		version = "latest"
	}

	forceCopy := opts.ForceCopy || inst.ForceCopy

	// Step 1: Store skill in repository
	storePath, err := inst.Repo.Store(skill, namespace, version)
	if err != nil {
		return nil, fmt.Errorf("store skill: %w", err)
	}

	// Step 2: Update index
	indexKey := buildIndexKey(namespace, skill.Name)
	existingEntry, _ := inst.Index.Get(indexKey)
	versions := []string{version}
	if existingEntry != nil {
		versions = append(existingEntry.Versions, version)
	}

	entry := models.IndexEntry{
		Name:       skill.Name,
		Namespace:  namespace,
		Versions:   versions,
		Latest:     version,
		Source:     skill.LocalPath,
		SourceType: namespace,
		Tags:       nil,
		Description: "",
	}

	if err := inst.Index.Add(entry); err != nil {
		return nil, fmt.Errorf("add to index: %w", err)
	}
	if err := inst.Index.Save(); err != nil {
		return nil, fmt.Errorf("save index: %w", err)
	}

	// Step 3: Update latest symlink in repo
	if err := inst.Repo.UpdateLatest(namespace, skill.Name, version); err != nil {
		fmt.Fprintf(os.Stderr, "warning: update latest symlink: %v\n", err)
	}

	result := &InstallResult{
		Name:      skill.Name,
		Namespace: namespace,
		Version:   version,
		StorePath: storePath,
	}

	// Step 4: Sync to agents (if requested)
	if !opts.NoSync {
		agentIDs := opts.Agents
		if len(agentIDs) == 0 {
			detected, _ := DetectAgents()
			if len(detected) > 0 {
				agentIDs = make([]string, len(detected))
				for i, a := range detected {
					agentIDs[i] = a.ID
				}
			} else {
				agentIDs = []string{"default"}
			}
		}

		var syncErrors []string
		for _, agentID := range agentIDs {
			syncResult, err := SyncSkillToAgent(storePath, agentID, forceCopy)
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("%s: %v", agentID, err))
				continue
			}
			result.SyncMode = syncResult.Mode

			// Track in lock file
			lockEntry := models.LockEntry{
				SkillID: models.SkillID{
					Namespace: namespace,
					Name:      skill.Name,
					Version:   version,
				},
				InstalledAt: time.Now().UTC().Format(time.RFC3339),
				Source:      storePath,
				Agents: []models.LockAgentBinding{
					{AgentID: agentID, Path: storePath, Mode: syncResult.Mode},
				},
			}
			inst.Lock.Track(lockEntry)
		}

		if err := inst.Lock.Save(); err != nil {
			return nil, fmt.Errorf("save lock file: %w", err)
		}

		result.Synced = true
		if len(syncErrors) > 0 {
			result.Error = fmt.Sprintf("partial sync failures: %s", syncErrors[0])
		}
	}

	return result, nil
}

// Uninstall removes a skill from the repository and all agents.
func (inst *Installer) Uninstall(namespace, name, version string) error {
	lockKey := buildLockKey(namespace, name, version)

	// Step 1: Get lock entry to find which agents have this skill installed
	lockEntry, err := inst.Lock.GetBySkill(lockKey)
	if err == nil {
		// Step 2: Remove from agents
		for _, agent := range lockEntry.Agents {
			skillDirName := fmt.Sprintf("%s@%s", name, version)
			if err := UnsyncSkillFromAgent(skillDirName, agent.AgentID); err != nil {
				fmt.Fprintf(os.Stderr, "warning: remove from agent %s: %v\n", agent.AgentID, err)
			}
		}

		// Step 3: Remove from lock file (remove all agent bindings)
		for _, agent := range lockEntry.Agents {
			inst.Lock.Untrack(lockKey, agent.AgentID)
		}
		if err := inst.Lock.Save(); err != nil {
			return fmt.Errorf("save lock file: %w", err)
		}
	}

	// Step 4: Remove version from index (not entire entry, in case other versions exist)
	indexKey := buildIndexKey(namespace, name)
	if existingEntry, err := inst.Index.Get(indexKey); err == nil {
		remainingVersions := make([]string, 0, len(existingEntry.Versions))
		for _, v := range existingEntry.Versions {
			if v != version {
				remainingVersions = append(remainingVersions, v)
			}
		}
		if len(remainingVersions) == 0 {
			inst.Index.Remove(indexKey)
		} else {
			latest := remainingVersions[0]
			if existingEntry.Latest == version {
				latest = remainingVersions[0]
			}
			existingEntry.Versions = remainingVersions
			existingEntry.Latest = latest
			inst.Index.Add(*existingEntry)
		}
	} else {
		inst.Index.Remove(indexKey)
	}
	if err := inst.Index.Save(); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	// Step 5: Remove from repository
	if err := inst.Repo.Remove(namespace, name, version); err != nil {
		return fmt.Errorf("remove from repository: %w", err)
	}

	return nil
}

// UpdateSkill updates a skill to a new version.
func (inst *Installer) UpdateSkill(namespace, name, oldVersion, newVersion string, agents []string, forceCopy bool) error {
	oldLockKey := buildLockKey(namespace, name, oldVersion)

	// Get lock entry to find current agents
	lockEntry, err := inst.Lock.GetBySkill(oldLockKey)
	if err != nil {
		return fmt.Errorf("skill %s not found in lock: %w", oldLockKey, err)
	}

	// Use existing agents if none specified
	if len(agents) == 0 {
		for _, a := range lockEntry.Agents {
			agents = append(agents, a.AgentID)
		}
	}

	// Step 1: Store new version first (before removing old, to ensure atomicity)
	oldPath := inst.Repo.SkillPath(namespace, name, oldVersion)
	oldModel := models.ResolvedSkill{
		LocalPath: oldPath,
		Namespace: namespace,
		Name:      name,
		Version:   newVersion,
	}

	newPath, err := inst.Repo.Store(oldModel, namespace, newVersion)
	if err != nil {
		return fmt.Errorf("store new version: %w", err)
	}

	// Step 2: Update index — merge versions instead of replacing
	indexKey := buildIndexKey(namespace, name)
	existingEntry, _ := inst.Index.Get(indexKey)
	versions := []string{newVersion}
	if existingEntry != nil {
		// Keep existing versions that are not the old version being replaced
		for _, v := range existingEntry.Versions {
			if v != oldVersion {
				versions = append(versions, v)
			}
		}
	}
	newEntry := models.IndexEntry{
		Name:       name,
		Namespace:  namespace,
		Versions:   versions,
		Latest:     newVersion,
		Source:     oldPath,
		SourceType: namespace,
	}
	inst.Index.Add(newEntry)
	if err := inst.Index.Save(); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	// Step 3: Update latest symlink
	if err := inst.Repo.UpdateLatest(namespace, name, newVersion); err != nil {
		fmt.Fprintf(os.Stderr, "warning: update latest symlink: %v\n", err)
	}

	// Step 4: Remove old version from repo
	if err := inst.Repo.Remove(namespace, name, oldVersion); err != nil {
		fmt.Fprintf(os.Stderr, "warning: remove old version: %v\n", err)
	}

	// Step 5: Remove old version from agents
	oldDirName := fmt.Sprintf("%s@%s", name, oldVersion)
	for _, agent := range lockEntry.Agents {
		if err := UnsyncSkillFromAgent(oldDirName, agent.AgentID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: remove old version from %s: %v\n", agent.AgentID, err)
		}
	}

	// Step 6: Remove old lock entries
	for _, agent := range lockEntry.Agents {
		inst.Lock.Untrack(oldLockKey, agent.AgentID)
	}

	// Step 7: Sync new version to agents
	for _, agentID := range agents {
		result, err := SyncSkillToAgent(newPath, agentID, forceCopy || inst.ForceCopy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: sync to %s: %v\n", agentID, err)
			continue
		}
		lockEntry := models.LockEntry{
			SkillID: models.SkillID{
				Namespace: namespace,
				Name:      name,
				Version:   newVersion,
			},
			InstalledAt: time.Now().UTC().Format(time.RFC3339),
			Source:      newPath,
			Agents: []models.LockAgentBinding{
				{AgentID: agentID, Path: newPath, Mode: result.Mode},
			},
		}
		inst.Lock.Track(lockEntry)
	}

	if err := inst.Lock.Save(); err != nil {
		return fmt.Errorf("save lock file: %w", err)
	}

	return nil
}

// CleanupStaleLinks removes broken symlinks in agent directories.
func (inst *Installer) CleanupStaleLinks() (int, error) {
	count := 0
	entries, err := inst.Lock.List()
	if err != nil {
		return 0, fmt.Errorf("list lock entries: %w", err)
	}

	for _, entry := range entries {
		for _, agent := range entry.Agents {
			agentSkillsDir, err := GetAgentSkillsDir(agent.AgentID)
			if err != nil {
				continue
			}

			linkPath := filepath.Join(agentSkillsDir, fmt.Sprintf("%s@%s", entry.SkillID.Name, entry.SkillID.Version))
			broken, err := IsSymlinkBroken(linkPath)
			if err != nil || !broken {
				continue
			}

			if err := os.Remove(linkPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: remove stale link %s: %v\n", linkPath, err)
				continue
			}
			count++
		}
	}

	return count, nil
}

// InstallFromSource resolves a source string and installs the skill(s).
func (inst *Installer) InstallFromSource(source string, opts InstallOptions) ([]*InstallResult, error) {
	resolver, err := NewResolverFromSource(source)
	if err != nil {
		return nil, fmt.Errorf("resolve source: %w", err)
	}

	resolveOpts := sourceResolveOptions{
		Version: opts.Version,
	}

	skills, err := resolver.Resolve(nil, source, resolveOpts)
	if err != nil {
		return nil, fmt.Errorf("resolve skills: %w", err)
	}

	var results []*InstallResult
	for _, skill := range skills {
		result, err := inst.Install(skill, opts)
		if err != nil {
			return results, fmt.Errorf("install %s: %w", skill.Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// sourceResolver is an interface matching source.Resolver to avoid direct dependency.
type sourceResolver interface {
	Resolve(ctx interface{}, source string, opts interface{}) ([]models.ResolvedSkill, error)
	CanHandle(source string) bool
}

// sourceResolveOptions mimics source.ResolveOptions.
type sourceResolveOptions struct {
	SubPath string
	Version string
	Ref     string
}

// NewResolverFromSource creates a resolver from a source string.
// This is a bridge that mimics source.NewResolver to avoid import cycles.
// In production, the CLI layer will wire the actual source.NewResolver.
var NewResolverFromSource = func(source string) (sourceResolver, error) {
	return nil, fmt.Errorf("source resolver not wired; use CLI commands instead")
}

// InstallFromConfig installs a skill from a configured repository source.
func (inst *Installer) InstallFromRegistry(skillName string, opts InstallOptions) (*InstallResult, error) {
	sourceURL := fmt.Sprintf("registry:%s", skillName)
	results, err := inst.InstallFromSource(sourceURL, opts)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no skills found for %s", skillName)
	}
	return results[0], nil
}