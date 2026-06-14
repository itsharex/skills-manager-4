package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// --- Compile-time interface checks ---

func TestResolverInterface(t *testing.T) {
	var _ Resolver = (*GitHubResolver)(nil)
	var _ Resolver = (*HTTPResolver)(nil)
	var _ Resolver = (*LocalResolver)(nil)
}

// --- NewResolver ---

func TestNewResolver_UnsupportedSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty source", ""},
		{"unknown scheme", "unknown://example.com"},
		{"random string", "some-random-source"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewResolver(tt.source)
			if err == nil {
				t.Errorf("NewResolver(%q) expected error, got resolver %T", tt.source, r)
			}
		})
	}
}

func TestNewResolver_FindsCorrectResolver(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   string
	}{
		{"GitHub - github.com URL", "https://github.com/owner/repo", "*source.GitHubResolver"},
		{"GitHub - gh: prefix", "gh:owner/repo", "*source.GitHubResolver"},
		{"GitHub - git@ URL", "git@github.com:owner/repo.git", "*source.GitHubResolver"},
		{"HTTP - https URL", "https://registry.example.com/index.json", "*source.HTTPResolver"},
		{"HTTP - http URL", "http://registry.example.com/index.json", "*source.HTTPResolver"},
		{"Local - absolute path", "/tmp/skills", "*source.LocalResolver"},
		{"Local - relative path", "./skills", "*source.LocalResolver"},
		{"Local - home path", "~/skills", "*source.LocalResolver"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewResolver(tt.source)
			if err != nil {
				t.Fatalf("NewResolver(%q) unexpected error: %v", tt.source, err)
			}
			typeStr := getResolverType(r)
			if typeStr != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, typeStr)
			}
		})
	}
}

func getResolverType(r Resolver) string {
	switch r.(type) {
	case *GitHubResolver:
		return "*source.GitHubResolver"
	case *HTTPResolver:
		return "*source.HTTPResolver"
	case *LocalResolver:
		return "*source.LocalResolver"
	default:
		return "unknown"
	}
}

// --- GitHubResolver.CanHandle ---

func TestGitHubResolver_CanHandle(t *testing.T) {
	r := &GitHubResolver{}

	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"github.com prefix", "github.com/owner/repo", true},
		{"gh: prefix", "gh:owner/repo", true},
		{"https github.com", "https://github.com/owner/repo", true},
		{"http github.com", "http://github.com/owner/repo", true},
		{"git@ github.com", "git@github.com:owner/repo.git", true},
		{"https github.com with .git", "https://github.com/owner/repo.git", true},
		{"github.com trailing slash", "github.com/owner/repo/", true},
		{"generic https", "https://example.com", false},
		{"generic http", "http://example.com", false},
		{"registry prefix", "registry:test", false},
		{"local path", "/tmp/skills", false},
		{"relative path", "./skills", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.CanHandle(tt.source)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestGitHubResolver_Resolve_InvalidSource(t *testing.T) {
	r := &GitHubResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, "invalid-source", ResolveOptions{})
	if err == nil {
		t.Error("expected error for invalid source")
	}
}

// --- HTTPResolver.CanHandle ---

func TestHTTPResolver_CanHandle(t *testing.T) {
	r := &HTTPResolver{}

	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"https non-github", "https://registry.example.com/index.json", true},
		{"http non-github", "http://registry.example.com/index.json", true},
		{"https with path", "https://example.com/skills/v1/index.json", true},
		{"registry: prefix", "registry:my-registry", true},
		{"https github.com", "https://github.com/owner/repo", false},
		{"http github.com", "http://github.com/owner/repo", false},
		{"github.com", "github.com/owner/repo", false},
		{"local path", "/tmp/skills", false},
		{"relative path", "./skills", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.CanHandle(tt.source)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestHTTPResolver_Resolve_RegistryName(t *testing.T) {
	r := &HTTPResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, "registry:my-registry", ResolveOptions{})
	if err == nil {
		t.Error("expected error for registry:name source")
	}
}

func TestHTTPResolver_Resolve_InvalidSource(t *testing.T) {
	r := &HTTPResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, "not-a-url", ResolveOptions{})
	if err == nil {
		t.Error("expected error for invalid source")
	}
}

func TestHTTPResolver_Resolve_Success(t *testing.T) {
	// Create a test HTTP server that returns a registry index
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"name": "pdf-tool", "description": "PDF utility", "version": "1.0.0", "url": "https://example.com/pdf.zip", "tags": ["pdf"]},
			{"name": "web-search", "description": "Web search tool", "version": "2.1.0", "url": "https://example.com/web.zip", "tags": ["web"]}
		]`))
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	// Check first skill
	if skills[0].Name != "pdf-tool" {
		t.Errorf("expected name 'pdf-tool', got %q", skills[0].Name)
	}
	if skills[0].Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", skills[0].Version)
	}

	// Check second skill
	if skills[1].Name != "web-search" {
		t.Errorf("expected name 'web-search', got %q", skills[1].Name)
	}
	if skills[1].Version != "2.1.0" {
		t.Errorf("expected version '2.1.0', got %q", skills[1].Version)
	}

	// Verify namespace contains the registry URL
	if skills[0].Namespace != "registry:"+server.URL {
		t.Errorf("expected namespace 'registry:%s', got %q", server.URL, skills[0].Namespace)
	}
}

func TestHTTPResolver_Resolve_EmptyRegistry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err == nil {
		t.Error("expected error for empty registry")
	}
}

func TestHTTPResolver_Resolve_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err == nil {
		t.Error("expected error for server error")
	}
}

// --- LocalResolver.CanHandle ---

func TestLocalResolver_CanHandle(t *testing.T) {
	r := &LocalResolver{}

	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{"absolute path", "/home/user/skills", true},
		{"absolute path with spaces", "/home/user/my skills", true},
		{"home path", "~/skills", true},
		{"home path subdir", "~/skills/my-skill", true},
		{"relative path ./", "./skills", true},
		{"relative path ../", "../skills", true},
		{"relative deep", "./path/to/skills", true},
		{"ZIP absolute", "/path/to/skills.zip", true},
		{"ZIP relative", "./skills.zip", true},
		{"https URL", "https://example.com", false},
		{"github URL", "github.com/owner/repo", false},
		{"gh: prefix", "gh:owner/repo", false},
		{"registry prefix", "registry:test", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.CanHandle(tt.source)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestLocalResolver_Resolve_NonExistentPath(t *testing.T) {
	r := &LocalResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, "/nonexistent/path/that/does/not/exist", ResolveOptions{})
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestLocalResolver_Resolve_SingleSkillDir(t *testing.T) {
	// Create a temp directory with a SKILL.md at root
	tmpDir := t.TempDir()
	skillContent := `---
name: my-test-skill
description: A test skill
version: 1.2.3
tags: [test]
---

# My Test Skill
`
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	// Use dir name as skill name for single-skill repos
	dirName := filepath.Base(tmpDir)
	if skills[0].Name != dirName {
		t.Errorf("expected name %q, got %q", dirName, skills[0].Name)
	}
	if skills[0].Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", skills[0].Version)
	}
	if skills[0].Namespace != "local" {
		t.Errorf("expected namespace 'local', got %q", skills[0].Namespace)
	}
	if skills[0].LocalPath != tmpDir {
		t.Errorf("expected LocalPath %q, got %q", tmpDir, skills[0].LocalPath)
	}
}

func TestLocalResolver_Resolve_MultiSkillDir(t *testing.T) {
	// Create a temp directory with multiple skill subdirectories
	tmpDir := t.TempDir()

	// First skill
	skill1Dir := filepath.Join(tmpDir, "skill-one")
	os.MkdirAll(skill1Dir, 0755)
	skill1Content := `---
name: skill-one
description: First skill
version: 1.0.0
---
`
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// Second skill
	skill2Dir := filepath.Join(tmpDir, "skill-two")
	os.MkdirAll(skill2Dir, 0755)
	skill2Content := `---
name: skill-two
description: Second skill
version: 2.0.0
---
`
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// Third skill (no SKILL.md — should be skipped)
	skill3Dir := filepath.Join(tmpDir, "skill-three")
	os.MkdirAll(skill3Dir, 0755)

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	// Verify skills (order may vary)
	skillMap := make(map[string]string)
	for _, s := range skills {
		skillMap[s.Name] = s.Version
	}
	if v, ok := skillMap["skill-one"]; !ok {
		t.Error("missing skill-one")
	} else if v != "1.0.0" {
		t.Errorf("skill-one version: expected 1.0.0, got %s", v)
	}
	if v, ok := skillMap["skill-two"]; !ok {
		t.Error("missing skill-two")
	} else if v != "2.0.0" {
		t.Errorf("skill-two version: expected 2.0.0, got %s", v)
	}

	if _, ok := skillMap["skill-three"]; ok {
		t.Error("skill-three should not be in results (no SKILL.md)")
	}
}

func TestLocalResolver_Resolve_SingleSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillContent := `---
name: direct-skill
description: A directly referenced skill
version: 3.0.0
---

# Content
`
	skillPath := filepath.Join(tmpDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, skillPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != filepath.Base(tmpDir) {
		t.Errorf("expected name %q, got %q", filepath.Base(tmpDir), skills[0].Name)
	}
	if skills[0].Version != "3.0.0" {
		t.Errorf("expected version '3.0.0', got %q", skills[0].Version)
	}
}

func TestLocalResolver_Resolve_SingleSkillDirNoFrontmatter(t *testing.T) {
	// Test that a SKILL.md without frontmatter still produces a skill with default values
	tmpDir := t.TempDir()
	skillContent := `# Just a heading`

	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Version != "latest" {
		t.Errorf("expected version 'latest' for missing frontmatter, got %q", skills[0].Version)
	}
}

// --- parseGitHubOwnerRepo ---

func TestParseGitHubOwnerRepo(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{"github.com/owner/repo", "owner/repo"},
		{"gh:owner/repo", "owner/repo"},
		{"https://github.com/owner/repo", "owner/repo"},
		{"http://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"git@github.com:owner/repo.git", "owner/repo"},
		{"github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo/", "owner/repo"},
		{"github.com/owner/my-cool-skill", "owner/my-cool-skill"},
		{"invalid", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			got := parseGitHubOwnerRepo(tt.source)
			if got != tt.want {
				t.Errorf("parseGitHubOwnerRepo(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}

// --- Integration: Local resolver with Version override ---

func TestLocalResolver_Resolve_WithVersionOverride(t *testing.T) {
	tmpDir := t.TempDir()
	skillContent := `---
name: versioned-skill
description: A skill
version: 1.0.0
---
`
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{Version: "2.0.0"})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Version != "2.0.0" {
		t.Errorf("expected version '2.0.0' from override, got %q", skills[0].Version)
	}
}

// --- NewResolver routing order ---

func TestNewResolver_GitHubBeforeHTTP(t *testing.T) {
	// GitHub URLs should be handled by GitHubResolver, not HTTPResolver
	r, err := NewResolver("https://github.com/owner/repo")
	if err != nil {
		t.Fatalf("NewResolver() unexpected error: %v", err)
	}
	if _, ok := r.(*GitHubResolver); !ok {
		t.Errorf("expected GitHubResolver for github URL, got %T", r)
	}
}

func TestNewResolver_LocalBeforeHTTP(t *testing.T) {
	// Local paths should be handled by LocalResolver
	r, err := NewResolver("/some/local/path")
	if err != nil {
		t.Fatalf("NewResolver() unexpected error: %v", err)
	}
	if _, ok := r.(*LocalResolver); !ok {
		t.Errorf("expected LocalResolver for local path, got %T", r)
	}
}