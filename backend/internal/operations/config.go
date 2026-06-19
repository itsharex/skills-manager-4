package operations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// expandPath expands ~ to the user's home directory and cleans the path.
func expandPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(home, path[2:])
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = home
	}
	// Clean trailing slashes and resolve . or .. elements
	return filepath.Clean(path)
}

// DefaultPoolPath returns the default local skill pool path (~/.skill-pool/)
func DefaultPoolPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".skill-pool")
	}
	return filepath.Join(home, ".skill-pool")
}

// Deprecated: Use DefaultPoolPath instead.
func DefaultRepoPath() string {
	return DefaultPoolPath()
}

// defaultConfig returns a Config with defaults filled in.
func defaultConfig() *models.Config {
	poolPath := DefaultPoolPath()
	return &models.Config{
		RepoPath:     poolPath,
		PoolPath:     poolPath,
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

	// Apply defaults for zero-value fields and expand ~ in paths
	cfg.PoolPath = expandPath(cfg.PoolPath)
	cfg.RepoPath = expandPath(cfg.RepoPath)
	if cfg.PoolPath == "" {
		// Backward compat: if old config has repo_path but no pool_path
		if cfg.RepoPath != "" {
			cfg.PoolPath = cfg.RepoPath
		} else {
			cfg.PoolPath = DefaultPoolPath()
		}
	}
	if cfg.RepoPath == "" {
		cfg.RepoPath = cfg.PoolPath
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
	// Clean paths before saving to avoid ~ or trailing slashes
	cfg.PoolPath = expandPath(cfg.PoolPath)
	cfg.RepoPath = expandPath(cfg.RepoPath)

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

// GetPoolPaths returns all paths derived from pool root.
func GetPoolPaths(poolPath string) models.PoolPaths {
	return models.PoolPaths{
		PoolPath:   poolPath,
		Root:       poolPath,
		SkillsDir:  poolPath, // skills directly under pool root
		MetaDir:    filepath.Join(poolPath, ".meta"),
		IndexPath:  filepath.Join(poolPath, ".meta", "index.json"),
		LockPath:   filepath.Join(poolPath, ".meta", "lock.json"),
		ConfigPath: filepath.Join(poolPath, ".meta", "config.json"),
	}
}

// Deprecated: Use GetPoolPaths instead.
func GetRepoPaths(root string) models.PoolPaths {
	return GetPoolPaths(root)
}

// EnsurePoolDir creates the pool directory structure.
func EnsurePoolDir(poolPath string) error {
	paths := GetPoolPaths(poolPath)
	dirs := []string{
		paths.PoolPath,
		paths.SkillsDir,
		paths.MetaDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}
	return nil
}

// Deprecated: Use EnsurePoolDir instead.
func EnsureRepoDir(root string) error {
	return EnsurePoolDir(root)
}

// InitPool creates a fresh pool with default config.
func InitPool(poolPath string) error {
	if err := EnsurePoolDir(poolPath); err != nil {
		return err
	}

	cfg := defaultConfig()
	cfg.PoolPath = poolPath
	cfg.RepoPath = poolPath

	paths := GetPoolPaths(poolPath)
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

// Deprecated: Use InitPool instead.
func InitRepo(root string) error {
	return InitPool(root)
}