package operations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// DefaultRepoPath returns the default repo path (~/.skill-repo/)
func DefaultRepoPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".skill-repo")
	}
	return filepath.Join(home, ".skill-repo")
}

// DefaultPoolPath returns the default local skill pool path (~/.skill-pool/)
func DefaultPoolPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".skill-pool")
	}
	return filepath.Join(home, ".skill-pool")
}

// defaultConfig returns a Config with defaults filled in.
func defaultConfig() *models.Config {
	return &models.Config{
		RepoPath:     DefaultRepoPath(),
		PoolPath:     DefaultPoolPath(),
		InstallMode:  "symlink",
		AutoFallback: true,
		CacheTTL:     3600,
	}
}

// LoadConfig reads config from file, applying defaults for missing fields.
func LoadConfig(path string) (*models.Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Apply defaults for zero-value fields
	if cfg.RepoPath == "" {
		cfg.RepoPath = DefaultRepoPath()
	}
	if cfg.PoolPath == "" {
		cfg.PoolPath = DefaultPoolPath()
	}
	if cfg.InstallMode == "" {
		cfg.InstallMode = "symlink"
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 3600
	}

	return cfg, nil
}

// SaveConfig writes config to file.
func SaveConfig(path string, cfg *models.Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// GetRepoPaths returns all paths derived from repo root.
func GetRepoPaths(root string) models.RepoPaths {
	return models.RepoPaths{
		Root:       root,
		SkillsDir:  filepath.Join(root, "skills"),
		IndexPath:  filepath.Join(root, "index.json"),
		LockPath:   filepath.Join(root, "lock.json"),
		ConfigPath: filepath.Join(root, "config.json"),
	}
}

// EnsureRepoDir creates the repo directory structure.
func EnsureRepoDir(root string) error {
	paths := GetRepoPaths(root)
	dirs := []string{
		paths.Root,
		paths.SkillsDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}
	return nil
}

// InitRepo creates a fresh repo with default config.
func InitRepo(root string) error {
	if err := EnsureRepoDir(root); err != nil {
		return err
	}

	cfg := defaultConfig()
	cfg.RepoPath = root

	paths := GetRepoPaths(root)
	if err := SaveConfig(paths.ConfigPath, cfg); err != nil {
		return fmt.Errorf("save default config: %w", err)
	}

	// Write empty index
	index := models.Index{
		Version:    1,
		LastUpdate: "",
		Skills:     make(map[string]models.IndexEntry),
	}
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err = os.WriteFile(paths.IndexPath, indexData, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	// Write empty lock
	lock := models.LockFile{
		Version: 1,
		Skills:  make(map[string]models.LockEntry),
	}
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	if err = os.WriteFile(paths.LockPath, lockData, 0o644); err != nil {
		return fmt.Errorf("write lock: %w", err)
	}

	return nil
}