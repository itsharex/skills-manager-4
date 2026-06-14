package operations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestDefaultRepoPath(t *testing.T) {
	path := DefaultRepoPath()
	if path == "" {
		t.Fatal("DefaultRepoPath returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got %q", path)
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent", "config.json"))
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.InstallMode != "symlink" {
		t.Errorf("expected default install mode 'symlink', got %q", cfg.InstallMode)
	}
	if !cfg.AutoFallback {
		t.Error("expected AutoFallback to default to true")
	}
	if cfg.CacheTTL != 3600 {
		t.Errorf("expected default CacheTTL 3600, got %d", cfg.CacheTTL)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	input := models.Config{
		RepoPath:     "/tmp/test-repo",
		InstallMode:  "copy",
		AutoFallback: false,
		CacheTTL:     7200,
		DefaultAgents: []string{"trae", "claude"},
	}
	data, _ := json.MarshalIndent(input, "", "  ")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.RepoPath != "/tmp/test-repo" {
		t.Errorf("expected RepoPath /tmp/test-repo, got %q", cfg.RepoPath)
	}
	if cfg.InstallMode != "copy" {
		t.Errorf("expected InstallMode 'copy', got %q", cfg.InstallMode)
	}
	if cfg.AutoFallback {
		t.Error("expected AutoFallback false")
	}
	if cfg.CacheTTL != 7200 {
		t.Errorf("expected CacheTTL 7200, got %d", cfg.CacheTTL)
	}
	if len(cfg.DefaultAgents) != 2 || cfg.DefaultAgents[0] != "trae" {
		t.Errorf("unexpected DefaultAgents: %v", cfg.DefaultAgents)
	}
}

func TestLoadConfig_PartialDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Only set RepoPath, everything else should get defaults
	partial := map[string]any{
		"repo_path": "/custom/path",
	}
	data, _ := json.Marshal(partial)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.RepoPath != "/custom/path" {
		t.Errorf("expected RepoPath /custom/path, got %q", cfg.RepoPath)
	}
	if cfg.InstallMode != "symlink" {
		t.Errorf("expected default InstallMode 'symlink', got %q", cfg.InstallMode)
	}
	if !cfg.AutoFallback {
		t.Error("expected AutoFallback to default to true")
	}
	if cfg.CacheTTL != 3600 {
		t.Errorf("expected default CacheTTL 3600, got %d", cfg.CacheTTL)
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	cfg := &models.Config{
		RepoPath:     "/my/repo",
		InstallMode:  "symlink",
		AutoFallback: true,
		CacheTTL:     1800,
	}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists and content is correct
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var loaded models.Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if loaded.RepoPath != "/my/repo" {
		t.Errorf("expected RepoPath /my/repo, got %q", loaded.RepoPath)
	}
	if loaded.CacheTTL != 1800 {
		t.Errorf("expected CacheTTL 1800, got %d", loaded.CacheTTL)
	}
}

func TestGetRepoPaths(t *testing.T) {
	paths := GetRepoPaths("/tmp/.skill-repo")

	if paths.Root != "/tmp/.skill-repo" {
		t.Errorf("expected Root /tmp/.skill-repo, got %q", paths.Root)
	}
	if paths.SkillsDir != "/tmp/.skill-repo/skills" {
		t.Errorf("expected SkillsDir /tmp/.skill-repo/skills, got %q", paths.SkillsDir)
	}
	if paths.IndexPath != "/tmp/.skill-repo/index.json" {
		t.Errorf("expected IndexPath /tmp/.skill-repo/index.json, got %q", paths.IndexPath)
	}
	if paths.LockPath != "/tmp/.skill-repo/lock.json" {
		t.Errorf("expected LockPath /tmp/.skill-repo/lock.json, got %q", paths.LockPath)
	}
	if paths.ConfigPath != "/tmp/.skill-repo/config.json" {
		t.Errorf("expected ConfigPath /tmp/.skill-repo/config.json, got %q", paths.ConfigPath)
	}
}

func TestEnsureRepoDir(t *testing.T) {
	root := t.TempDir()

	if err := EnsureRepoDir(root); err != nil {
		t.Fatalf("EnsureRepoDir failed: %v", err)
	}

	// Verify directories exist
	paths := GetRepoPaths(root)
	if _, err := os.Stat(paths.Root); os.IsNotExist(err) {
		t.Error("root directory not created")
	}
	if _, err := os.Stat(paths.SkillsDir); os.IsNotExist(err) {
		t.Error("skills directory not created")
	}
}

func TestInitRepo(t *testing.T) {
	root := t.TempDir()

	if err := InitRepo(root); err != nil {
		t.Fatalf("InitRepo failed: %v", err)
	}

	paths := GetRepoPaths(root)

	// Verify config.json exists and has defaults
	cfg, err := LoadConfig(paths.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig after InitRepo failed: %v", err)
	}
	if cfg.RepoPath != root {
		t.Errorf("expected RepoPath %q, got %q", root, cfg.RepoPath)
	}

	// Verify index.json exists
	indexData, err := os.ReadFile(paths.IndexPath)
	if err != nil {
		t.Fatalf("index.json not found: %v", err)
	}
	var index models.Index
	if err = json.Unmarshal(indexData, &index); err != nil {
		t.Fatalf("unmarshal index: %v", err)
	}
	if index.Version != 1 {
		t.Errorf("expected index version 1, got %d", index.Version)
	}

	// Verify lock.json exists
	lockData, err := os.ReadFile(paths.LockPath)
	if err != nil {
		t.Fatalf("lock.json not found: %v", err)
	}
	var lock models.LockFile
	if err := json.Unmarshal(lockData, &lock); err != nil {
		t.Fatalf("unmarshal lock: %v", err)
	}
	if lock.Version != 1 {
		t.Errorf("expected lock version 1, got %d", lock.Version)
	}

	// Verify directories exist
	if _, err := os.Stat(paths.SkillsDir); os.IsNotExist(err) {
		t.Error("skills directory not created")
	}
}

// --- Phase 1: Config model extension tests ---

func TestDefaultPoolPath(t *testing.T) {
	path := DefaultPoolPath()
	if path == "" {
		t.Fatal("DefaultPoolPath returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got %q", path)
	}
}

func TestDefaultConfig_HasPoolPath(t *testing.T) {
	cfg := defaultConfig()
	if cfg.PoolPath == "" {
		t.Fatal("defaultConfig should set PoolPath")
	}
	if cfg.PoolPath != DefaultPoolPath() {
		t.Errorf("expected PoolPath %q, got %q", DefaultPoolPath(), cfg.PoolPath)
	}
}

func TestLoadConfig_OldFormatBackwardCompat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write old-style config WITHOUT pool_path and market_sources
	oldData := map[string]any{
		"repo_path":     "/old/repo",
		"install_mode":  "copy",
		"auto_fallback": false,
		"cache_ttl":     1800,
	}
	data, _ := json.Marshal(oldData)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// PoolPath should get default value
	if cfg.PoolPath == "" {
		t.Error("PoolPath should have default value for old config")
	}
	if cfg.PoolPath != DefaultPoolPath() {
		t.Errorf("expected PoolPath %q, got %q", DefaultPoolPath(), cfg.PoolPath)
	}

	// MarketSources should be nil (not cause error) for old config
	if cfg.MarketSources != nil {
		t.Error("expected MarketSources nil for old config without market_sources field")
	}

	// Other fields preserved
	if cfg.RepoPath != "/old/repo" {
		t.Errorf("expected RepoPath /old/repo, got %q", cfg.RepoPath)
	}
	if cfg.InstallMode != "copy" {
		t.Errorf("expected InstallMode copy, got %q", cfg.InstallMode)
	}
}

func TestSaveConfig_WithMarketSources(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &models.Config{
		RepoPath:     "/test/repo",
		PoolPath:     "/test/pool",
		InstallMode:  "symlink",
		AutoFallback: true,
		CacheTTL:     3600,
		MarketSources: []models.MarketSource{
			{Name: "My Pool", URL: "/local/path", Type: "pool", Enabled: true},
			{Name: "GitHub Skills", URL: "https://github.com/owner/repo", Type: "github", Enabled: true, Branch: "main"},
			{Name: "Registry", URL: "https://registry.example.com", Type: "registry", Enabled: false},
		},
	}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Reload and verify
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig after SaveConfig failed: %v", err)
	}

	if loaded.PoolPath != "/test/pool" {
		t.Errorf("expected PoolPath /test/pool, got %q", loaded.PoolPath)
	}

	if len(loaded.MarketSources) != 3 {
		t.Fatalf("expected 3 MarketSources, got %d", len(loaded.MarketSources))
	}

	if loaded.MarketSources[0].Name != "My Pool" {
		t.Errorf("expected source[0].Name 'My Pool', got %q", loaded.MarketSources[0].Name)
	}
	if loaded.MarketSources[0].Type != "pool" {
		t.Errorf("expected source[0].Type 'pool', got %q", loaded.MarketSources[0].Type)
	}
	if !loaded.MarketSources[0].Enabled {
		t.Error("expected source[0].Enabled true")
	}

	if loaded.MarketSources[1].Name != "GitHub Skills" {
		t.Errorf("expected source[1].Name 'GitHub Skills', got %q", loaded.MarketSources[1].Name)
	}
	if loaded.MarketSources[1].Type != "github" {
		t.Errorf("expected source[1].Type 'github', got %q", loaded.MarketSources[1].Type)
	}
	if loaded.MarketSources[1].Branch != "main" {
		t.Errorf("expected source[1].Branch 'main', got %q", loaded.MarketSources[1].Branch)
	}

	if loaded.MarketSources[2].Enabled {
		t.Error("expected source[2].Enabled false")
	}
}

func TestMarketSource_ZeroValue(t *testing.T) {
	var ms models.MarketSource
	if ms.Name != "" {
		t.Errorf("expected empty Name, got %q", ms.Name)
	}
	if ms.Type != "" {
		t.Errorf("expected empty Type, got %q", ms.Type)
	}
	if ms.Enabled {
		t.Error("expected Enabled false")
	}
	if ms.Branch != "" {
		t.Errorf("expected empty Branch, got %q", ms.Branch)
	}
}