package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// LockFile manages installation locks (lock.json).
type LockFile struct {
	data *models.LockFile
	path string
	mu   sync.RWMutex
}

// NewLockFile loads or creates a lock file from the given path.
func NewLockFile(path string) (*LockFile, error) {
	lf := &LockFile{
		data: &models.LockFile{
			Version: 1,
			Skills:  make(map[string]models.LockEntry),
		},
		path: path,
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create lock directory: %w", err)
		}
		if err := lf.Save(); err != nil {
			return nil, fmt.Errorf("save initial lock: %w", err)
		}
		return lf, nil
	}

	if err := lf.Reload(); err != nil {
		return nil, fmt.Errorf("load lock: %w", err)
	}

	return lf, nil
}

// Track records a skill installation for an agent.
func (lf *LockFile) Track(entry models.LockEntry) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	key := fmt.Sprintf("%s/%s@%s", entry.SkillID.Namespace, entry.SkillID.Name, entry.SkillID.Version)

	if existing, ok := lf.data.Skills[key]; ok {
		// Merge agents - avoid duplicates by agentID
		agentMap := make(map[string]bool, len(existing.Agents))
		for _, a := range existing.Agents {
			agentMap[a.AgentID] = true
		}
		for _, a := range entry.Agents {
			if !agentMap[a.AgentID] {
				existing.Agents = append(existing.Agents, a)
				agentMap[a.AgentID] = true
			}
		}
		// Update other fields
		existing.InstalledAt = entry.InstalledAt
		existing.Source = entry.Source
		lf.data.Skills[key] = existing
	} else {
		lf.data.Skills[key] = entry
	}

	return nil
}

// Untrack removes an agent binding from a skill's lock entry.
// If no agents remain, the entire entry is removed.
func (lf *LockFile) Untrack(skillKey, agentID string) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	entry, ok := lf.data.Skills[skillKey]
	if !ok {
		return fmt.Errorf("skill %q not found in lock file", skillKey)
	}

	filtered := make([]models.LockAgentBinding, 0, len(entry.Agents))
	for _, a := range entry.Agents {
		if a.AgentID != agentID {
			filtered = append(filtered, a)
		}
	}

	if len(filtered) == 0 {
		delete(lf.data.Skills, skillKey)
	} else {
		entry.Agents = filtered
		lf.data.Skills[skillKey] = entry
	}

	return nil
}

// GetBySkill returns the lock entry for a specific skill.
func (lf *LockFile) GetBySkill(skillKey string) (*models.LockEntry, error) {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	entry, ok := lf.data.Skills[skillKey]
	if !ok {
		return nil, fmt.Errorf("skill %q not found in lock file", skillKey)
	}
	return &entry, nil
}

// GetByAgent returns all skills locked for a specific agent.
func (lf *LockFile) GetByAgent(agentID string) ([]models.LockEntry, error) {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	var entries []models.LockEntry
	for _, entry := range lf.data.Skills {
		for _, a := range entry.Agents {
			if a.AgentID == agentID {
				entries = append(entries, entry)
				break
			}
		}
	}
	return entries, nil
}

// List returns all lock entries.
func (lf *LockFile) List() ([]models.LockEntry, error) {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	entries := make([]models.LockEntry, 0, len(lf.data.Skills))
	for _, entry := range lf.data.Skills {
		entries = append(entries, entry)
	}
	return entries, nil
}

// Save persists the lock file to disk.
func (lf *LockFile) Save() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	data, err := json.MarshalIndent(lf.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}

	if err := os.WriteFile(lf.path, data, 0o644); err != nil {
		return fmt.Errorf("write lock: %w", err)
	}
	return nil
}

// Reload reloads the lock file from disk.
func (lf *LockFile) Reload() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	data, err := os.ReadFile(lf.path)
	if err != nil {
		return fmt.Errorf("read lock: %w", err)
	}

	var lock models.LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("parse lock: %w", err)
	}

	if lock.Skills == nil {
		lock.Skills = make(map[string]models.LockEntry)
	}

	lf.data = &lock
	return nil
}