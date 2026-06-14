package models

import (
	"encoding/json"
	"testing"
)

func TestSkillID_JSON(t *testing.T) {
	sid := SkillID{
		Namespace: "github.com/example",
		Name:      "my-skill",
		Version:   "1.2.3",
	}

	data, err := json.Marshal(sid)
	if err != nil {
		t.Fatalf("marshal SkillID: %v", err)
	}

	var got SkillID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal SkillID: %v", err)
	}

	if got != sid {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, sid)
	}
}

func TestSkillID_Empty(t *testing.T) {
	sid := SkillID{}
	data, err := json.Marshal(sid)
	if err != nil {
		t.Fatalf("marshal empty SkillID: %v", err)
	}

	var got SkillID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal empty SkillID: %v", err)
	}

	if got.Namespace != "" || got.Name != "" || got.Version != "" {
		t.Errorf("expected all empty fields, got %+v", got)
	}
}

func TestSkillID_Partial(t *testing.T) {
	sid := SkillID{
		Name:    "skill-only",
		Version: "latest",
	}

	data, err := json.Marshal(sid)
	if err != nil {
		t.Fatalf("marshal partial SkillID: %v", err)
	}

	var got SkillID
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal partial SkillID: %v", err)
	}

	if got.Namespace != "" {
		t.Errorf("expected empty namespace, got %q", got.Namespace)
	}
	if got.Name != "skill-only" {
		t.Errorf("expected Name 'skill-only', got %q", got.Name)
	}
	if got.Version != "latest" {
		t.Errorf("expected Version 'latest', got %q", got.Version)
	}
}

func TestConfig_JSON(t *testing.T) {
	cfg := Config{
		RepoPath:     "/tmp/.skill-repo",
		InstallMode:  "symlink",
		AutoFallback: true,
		DefaultAgents: []string{"claude", "cursor"},
		LinkTargets: []LinkTarget{
			{ID: "claude", Path: "/home/user/.claude/skills", Enabled: true},
		},
		Repositories: []RepoSource{
			{Name: "official", URL: "https://skills.example.com", Type: "registry", Enabled: true},
		},
		CacheTTL: 3600,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal Config: %v", err)
	}

	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Config: %v", err)
	}

	if got.RepoPath != cfg.RepoPath {
		t.Errorf("RepoPath: got %q, want %q", got.RepoPath, cfg.RepoPath)
	}
	if got.InstallMode != cfg.InstallMode {
		t.Errorf("InstallMode: got %q, want %q", got.InstallMode, cfg.InstallMode)
	}
	if got.AutoFallback != cfg.AutoFallback {
		t.Errorf("AutoFallback: got %v, want %v", got.AutoFallback, cfg.AutoFallback)
	}
	if got.CacheTTL != cfg.CacheTTL {
		t.Errorf("CacheTTL: got %d, want %d", got.CacheTTL, cfg.CacheTTL)
	}
	if len(got.DefaultAgents) != len(cfg.DefaultAgents) {
		t.Errorf("DefaultAgents length: got %d, want %d", len(got.DefaultAgents), len(cfg.DefaultAgents))
	}
	if len(got.LinkTargets) != len(cfg.LinkTargets) {
		t.Errorf("LinkTargets length: got %d, want %d", len(got.LinkTargets), len(cfg.LinkTargets))
	}
	if len(got.Repositories) != len(cfg.Repositories) {
		t.Errorf("Repositories length: got %d, want %d", len(got.Repositories), len(cfg.Repositories))
	}
}

func TestConfig_Empty(t *testing.T) {
	cfg := Config{}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal empty Config: %v", err)
	}

	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal empty Config: %v", err)
	}

	if got.RepoPath != "" {
		t.Errorf("expected empty RepoPath, got %q", got.RepoPath)
	}
	if got.DefaultAgents != nil {
		t.Errorf("expected nil DefaultAgents, got %v", got.DefaultAgents)
	}
	if got.LinkTargets != nil {
		t.Errorf("expected nil LinkTargets, got %v", got.LinkTargets)
	}
}

func TestConfig_PartialJSON(t *testing.T) {
	raw := `{"repo_path":"/custom/path","install_mode":"copy"}`
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("unmarshal partial JSON: %v", err)
	}
	if cfg.RepoPath != "/custom/path" {
		t.Errorf("RepoPath: got %q, want /custom/path", cfg.RepoPath)
	}
	if cfg.InstallMode != "copy" {
		t.Errorf("InstallMode: got %q, want copy", cfg.InstallMode)
	}
	if cfg.AutoFallback {
		t.Error("expected AutoFallback to be false by default")
	}
	if cfg.CacheTTL != 0 {
		t.Errorf("expected CacheTTL 0, got %d", cfg.CacheTTL)
	}
}

func TestIndex_Population(t *testing.T) {
	idx := Index{
		Version:    1,
		LastUpdate: "2025-01-15T10:00:00Z",
		Skills: map[string]IndexEntry{
			"github.com/example/my-skill": {
				Name:          "my-skill",
				Namespace:     "github.com/example",
				Versions:      []string{"1.0.0", "1.1.0", "2.0.0"},
				Latest:        "2.0.0",
				Source:        "https://github.com/example/my-skill",
				SourceType:    "github",
				InstalledSize: "1.2 MB",
				Tags:          []string{"productivity", "utility"},
				Description:   "A sample skill",
			},
			"github.com/example/other-skill": {
				Name:      "other-skill",
				Namespace: "github.com/example",
				Versions:  []string{"0.1.0"},
				Latest:    "0.1.0",
				SourceType: "registry",
				Tags:      []string{"demo"},
			},
		},
	}

	data, err := json.Marshal(idx)
	if err != nil {
		t.Fatalf("marshal Index: %v", err)
	}

	var got Index
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Index: %v", err)
	}

	if got.Version != idx.Version {
		t.Errorf("Version: got %d, want %d", got.Version, idx.Version)
	}
	if got.LastUpdate != idx.LastUpdate {
		t.Errorf("LastUpdate: got %q, want %q", got.LastUpdate, idx.LastUpdate)
	}
	if len(got.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(got.Skills))
	}

	entry, ok := got.Skills["github.com/example/my-skill"]
	if !ok {
		t.Fatal("expected key 'github.com/example/my-skill' in skills map")
	}
	if entry.Latest != "2.0.0" {
		t.Errorf("Latest: got %q, want 2.0.0", entry.Latest)
	}
	if len(entry.Versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(entry.Versions))
	}
	if len(entry.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(entry.Tags))
	}
}

func TestIndex_NilSkills(t *testing.T) {
	idx := Index{
		Version: 1,
	}

	data, err := json.Marshal(idx)
	if err != nil {
		t.Fatalf("marshal Index with nil skills: %v", err)
	}

	var got Index
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Index with nil skills: %v", err)
	}

	if got.Skills != nil {
		t.Logf("note: Skills map is %v (not nil) after unmarshal", got.Skills)
	}
}

func TestIndex_EmptySkillsMap(t *testing.T) {
	idx := Index{
		Version:    1,
		LastUpdate: "",
		Skills:     make(map[string]IndexEntry),
	}

	data, err := json.Marshal(idx)
	if err != nil {
		t.Fatalf("marshal Index with empty skills: %v", err)
	}

	var got Index
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Index with empty skills: %v", err)
	}

	if got.Skills == nil {
		t.Fatal("expected non-nil Skills map")
	}
	if len(got.Skills) != 0 {
		t.Errorf("expected empty Skills map, got %d entries", len(got.Skills))
	}
}

func TestLockFile_Tracking(t *testing.T) {
	lf := LockFile{
		Version: 1,
		Skills: map[string]LockEntry{
			"github.com/example/my-skill@1.0.0": {
				SkillID: SkillID{
					Namespace: "github.com/example",
					Name:      "my-skill",
					Version:   "1.0.0",
				},
				InstalledAt: "2025-01-15T10:00:00Z",
				Source:      "https://github.com/example/my-skill",
				Agents: []LockAgentBinding{
					{AgentID: "claude", Path: "/home/user/.claude/skills/my-skill", Mode: "symlink"},
					{AgentID: "cursor", Path: "/home/user/.cursor/skills/my-skill", Mode: "copy"},
				},
			},
			"github.com/example/other@0.1.0": {
				SkillID: SkillID{
					Namespace: "github.com/example",
					Name:      "other",
					Version:   "0.1.0",
				},
				InstalledAt: "2025-01-16T08:30:00Z",
				Source:      "registry",
				Agents: []LockAgentBinding{
					{AgentID: "claude", Path: "/home/user/.claude/skills/other", Mode: "symlink"},
				},
			},
		},
	}

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("marshal LockFile: %v", err)
	}

	var got LockFile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal LockFile: %v", err)
	}

	if got.Version != lf.Version {
		t.Errorf("Version: got %d, want %d", got.Version, lf.Version)
	}
	if len(got.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(got.Skills))
	}

	entry, ok := got.Skills["github.com/example/my-skill@1.0.0"]
	if !ok {
		t.Fatal("expected key 'github.com/example/my-skill@1.0.0'")
	}
	if entry.SkillID.Name != "my-skill" {
		t.Errorf("SkillID.Name: got %q, want my-skill", entry.SkillID.Name)
	}
	if entry.SkillID.Version != "1.0.0" {
		t.Errorf("SkillID.Version: got %q, want 1.0.0", entry.SkillID.Version)
	}
	if len(entry.Agents) != 2 {
		t.Fatalf("expected 2 agent bindings, got %d", len(entry.Agents))
	}
	if entry.Agents[0].Mode != "symlink" {
		t.Errorf("Agent[0].Mode: got %q, want symlink", entry.Agents[0].Mode)
	}
}

func TestLockFile_Empty(t *testing.T) {
	lf := LockFile{
		Version: 1,
		Skills:  make(map[string]LockEntry),
	}

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("marshal empty LockFile: %v", err)
	}

	var got LockFile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal empty LockFile: %v", err)
	}

	if got.Version != 1 {
		t.Errorf("Version: got %d, want 1", got.Version)
	}
	if len(got.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(got.Skills))
	}
}

func TestLockFile_NilSkills(t *testing.T) {
	lf := LockFile{
		Version: 1,
	}

	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatalf("marshal LockFile with nil skills: %v", err)
	}

	var got LockFile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal LockFile with nil skills: %v", err)
	}

	if got.Skills == nil {
		t.Log("note: Skills remains nil after unmarshal of nil map")
	}
}

func TestLinkTarget(t *testing.T) {
	lt := LinkTarget{
		ID:      "claude",
		Path:    "/home/user/.claude/skills",
		Enabled: true,
	}

	data, err := json.Marshal(lt)
	if err != nil {
		t.Fatalf("marshal LinkTarget: %v", err)
	}

	var got LinkTarget
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal LinkTarget: %v", err)
	}

	if got.ID != lt.ID {
		t.Errorf("ID: got %q, want %q", got.ID, lt.ID)
	}
	if got.Path != lt.Path {
		t.Errorf("Path: got %q, want %q", got.Path, lt.Path)
	}
	if got.Enabled != lt.Enabled {
		t.Errorf("Enabled: got %v, want %v", got.Enabled, lt.Enabled)
	}
}

func TestLinkTarget_Disabled(t *testing.T) {
	lt := LinkTarget{
		ID:      "disabled-agent",
		Path:    "/some/path",
		Enabled: false,
	}
	data, _ := json.Marshal(lt)
	var got LinkTarget
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal disabled LinkTarget: %v", err)
	}
	if got.Enabled {
		t.Error("expected Enabled to be false")
	}
}

func TestRepoSource(t *testing.T) {
	rs := RepoSource{
		Name:    "official",
		URL:     "https://skills.example.com/api",
		Type:    "registry",
		Enabled: true,
	}

	data, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("marshal RepoSource: %v", err)
	}

	var got RepoSource
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal RepoSource: %v", err)
	}

	if got.Name != rs.Name {
		t.Errorf("Name: got %q, want %q", got.Name, rs.Name)
	}
	if got.URL != rs.URL {
		t.Errorf("URL: got %q, want %q", got.URL, rs.URL)
	}
	if got.Type != rs.Type {
		t.Errorf("Type: got %q, want %q", got.Type, rs.Type)
	}
	if got.Enabled != rs.Enabled {
		t.Errorf("Enabled: got %v, want %v", got.Enabled, rs.Enabled)
	}
}

func TestRepoSource_GitHubType(t *testing.T) {
	rs := RepoSource{
		Name:    "community-skills",
		URL:     "https://github.com/community/skills",
		Type:    "github",
		Enabled: false,
	}

	data, _ := json.Marshal(rs)
	var got RepoSource
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal RepoSource: %v", err)
	}
	if got.Type != "github" {
		t.Errorf("Type: got %q, want github", got.Type)
	}
	if got.Enabled {
		t.Error("expected Enabled to be false")
	}
}

func TestResolvedSkill(t *testing.T) {
	called := false
	rs := ResolvedSkill{
		LocalPath: "/tmp/skill-abc123",
		Namespace: "github.com/example",
		Name:      "my-skill",
		Version:   "1.0.0",
		Cleanup: func() {
			called = true
		},
	}

	if rs.LocalPath != "/tmp/skill-abc123" {
		t.Errorf("LocalPath: got %q, want /tmp/skill-abc123", rs.LocalPath)
	}
	if rs.Namespace != "github.com/example" {
		t.Errorf("Namespace: got %q, want github.com/example", rs.Namespace)
	}
	if rs.Name != "my-skill" {
		t.Errorf("Name: got %q, want my-skill", rs.Name)
	}
	if rs.Version != "1.0.0" {
		t.Errorf("Version: got %q, want 1.0.0", rs.Version)
	}

	rs.Cleanup()
	if !called {
		t.Error("expected Cleanup func to be called")
	}
}

func TestResolvedSkill_Empty(t *testing.T) {
	rs := ResolvedSkill{}
	if rs.LocalPath != "" {
		t.Errorf("expected empty LocalPath, got %q", rs.LocalPath)
	}
	if rs.Cleanup != nil {
		t.Error("expected nil Cleanup")
	}
}

func TestRepoPaths(t *testing.T) {
	rp := RepoPaths{
		Root:       "/home/user/.skill-repo",
		SkillsDir:  "/home/user/.skill-repo/skills",
		IndexPath:  "/home/user/.skill-repo/index.json",
		LockPath:   "/home/user/.skill-repo/lock.json",
		ConfigPath: "/home/user/.skill-repo/config.json",
	}

	if rp.Root != "/home/user/.skill-repo" {
		t.Errorf("Root: got %q", rp.Root)
	}
	if rp.SkillsDir != "/home/user/.skill-repo/skills" {
		t.Errorf("SkillsDir: got %q", rp.SkillsDir)
	}
	if rp.IndexPath != "/home/user/.skill-repo/index.json" {
		t.Errorf("IndexPath: got %q", rp.IndexPath)
	}
	if rp.LockPath != "/home/user/.skill-repo/lock.json" {
		t.Errorf("LockPath: got %q", rp.LockPath)
	}
	if rp.ConfigPath != "/home/user/.skill-repo/config.json" {
		t.Errorf("ConfigPath: got %q", rp.ConfigPath)
	}
}

func TestIndexEntry(t *testing.T) {
	entry := IndexEntry{
		Name:          "test-skill",
		Namespace:     "local",
		Versions:      []string{"0.1.0", "0.2.0"},
		Latest:        "0.2.0",
		Source:        "file:///tmp/skills",
		SourceType:    "local",
		InstalledSize: "512 KB",
		Tags:          []string{"test"},
		Description:   "A test skill entry",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal IndexEntry: %v", err)
	}

	var got IndexEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal IndexEntry: %v", err)
	}

	if got.Name != entry.Name {
		t.Errorf("Name: got %q, want %q", got.Name, entry.Name)
	}
	if got.Namespace != entry.Namespace {
		t.Errorf("Namespace: got %q, want %q", got.Namespace, entry.Namespace)
	}
	if len(got.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(got.Versions))
	}
	if got.Latest != "0.2.0" {
		t.Errorf("Latest: got %q, want 0.2.0", got.Latest)
	}
	if len(got.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(got.Tags))
	}
}

func TestLockAgentBinding(t *testing.T) {
	binding := LockAgentBinding{
		AgentID: "windsurf",
		Path:    "/home/user/.windsurf/skills/my-skill",
		Mode:    "symlink",
	}

	data, err := json.Marshal(binding)
	if err != nil {
		t.Fatalf("marshal LockAgentBinding: %v", err)
	}

	var got LockAgentBinding
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal LockAgentBinding: %v", err)
	}

	if got.AgentID != "windsurf" {
		t.Errorf("AgentID: got %q, want windsurf", got.AgentID)
	}
	if got.Path != binding.Path {
		t.Errorf("Path: got %q, want %q", got.Path, binding.Path)
	}
	if got.Mode != "symlink" {
		t.Errorf("Mode: got %q, want symlink", got.Mode)
	}
}

func TestJSON_UnknownFieldsIgnored(t *testing.T) {
	// Extra fields in JSON should be ignored (forward compatibility)
	raw := `{"namespace":"ns","name":"n","version":"1","unknown_field":"should_be_ignored"}`
	var sid SkillID
	if err := json.Unmarshal([]byte(raw), &sid); err != nil {
		t.Fatalf("unmarshal with extra field: %v", err)
	}
	if sid.Namespace != "ns" || sid.Name != "n" || sid.Version != "1" {
		t.Errorf("unexpected values: %+v", sid)
	}
}