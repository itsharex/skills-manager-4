package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Index manages the global skill index (index.json).
type Index struct {
	data *models.Index // in-memory state
	path string        // path to index.json
	mu   sync.RWMutex  // thread safety
}

// NewIndex loads or creates an index from the given path.
func NewIndex(path string) (*Index, error) {
	idx := &Index{
		data: &models.Index{
			Version:    1,
			LastUpdate: "",
			Skills:     make(map[string]models.IndexEntry),
		},
		path: path,
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create index directory: %w", err)
		}
		if err := idx.Save(); err != nil {
			return nil, fmt.Errorf("save initial index: %w", err)
		}
		return idx, nil
	}

	if err := idx.Reload(); err != nil {
		return nil, fmt.Errorf("load index: %w", err)
	}

	return idx, nil
}

// Add adds or updates a skill entry in the index.
func (idx *Index) Add(entry models.IndexEntry) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	key := fmt.Sprintf("%s/%s", entry.Namespace, entry.Name)

	// Deduplicate the incoming entry's versions list
	deduped := make([]string, 0, len(entry.Versions))
	seen := make(map[string]bool, len(entry.Versions))
	for _, v := range entry.Versions {
		if !seen[v] {
			deduped = append(deduped, v)
			seen[v] = true
		}
	}
	entry.Versions = deduped

	if existing, ok := idx.data.Skills[key]; ok {
		// Merge versions list (append new version if not already present)
		versionSet := make(map[string]bool, len(existing.Versions))
		for _, v := range existing.Versions {
			versionSet[v] = true
		}
		for _, v := range entry.Versions {
			if !versionSet[v] {
				existing.Versions = append(existing.Versions, v)
				versionSet[v] = true
			}
		}
		// Update other fields from the new entry
		existing.Latest = entry.Latest
		existing.Source = entry.Source
		existing.SourceType = entry.SourceType
		existing.InstalledSize = entry.InstalledSize
		existing.Tags = entry.Tags
		existing.Description = entry.Description
		idx.data.Skills[key] = existing
	} else {
		idx.data.Skills[key] = entry
	}

	idx.data.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// Remove removes a skill entry by name (key format: namespace/name).
func (idx *Index) Remove(key string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.data.Skills, key)
	idx.data.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// Get returns a single skill entry by key.
func (idx *Index) Get(key string) (*models.IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entry, ok := idx.data.Skills[key]
	if !ok {
		return nil, fmt.Errorf("skill %q not found in index", key)
	}
	return &entry, nil
}

// List returns all skill entries.
func (idx *Index) List() ([]models.IndexEntry, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entries := make([]models.IndexEntry, 0, len(idx.data.Skills))
	for _, entry := range idx.data.Skills {
		entries = append(entries, entry)
	}
	return entries, nil
}

// UpdateLatest updates the latest version for a skill.
func (idx *Index) UpdateLatest(key, version string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	entry, ok := idx.data.Skills[key]
	if !ok {
		return fmt.Errorf("skill %q not found in index", key)
	}
	entry.Latest = version
	idx.data.Skills[key] = entry
	idx.data.LastUpdate = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// Save persists the index to disk.
func (idx *Index) Save() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	data, err := json.MarshalIndent(idx.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := os.WriteFile(idx.path, data, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// Reload reloads the index from disk.
func (idx *Index) Reload() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	data, err := os.ReadFile(idx.path)
	if err != nil {
		return fmt.Errorf("read index: %w", err)
	}

	var index models.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("parse index: %w", err)
	}

	if index.Skills == nil {
		index.Skills = make(map[string]models.IndexEntry)
	}

	idx.data = &index
	return nil
}