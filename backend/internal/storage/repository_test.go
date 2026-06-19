package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestNewRepository(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	if r.Root != root {
		t.Errorf("expected Root %q, got %q", root, r.Root)
	}
	if r.Paths.SkillsDir != root {
		t.Errorf("expected SkillsDir %q, got %q", root, r.Paths.SkillsDir)
	}
	if r.Paths.MetaDir != filepath.Join(root, ".meta") {
		t.Errorf("expected MetaDir %q, got %q", filepath.Join(root, ".meta"), r.Paths.MetaDir)
	}
	if r.Paths.IndexPath != filepath.Join(root, ".meta", "index.json") {
		t.Errorf("expected IndexPath %q, got %q", filepath.Join(root, ".meta", "index.json"), r.Paths.IndexPath)
	}
}

func TestSkillPath(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	path := r.SkillPath("github.com/owner", "my-skill", "1.0.0")
	expected := filepath.Join(root, "my-skill")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestStore_Directory(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// Create a source directory with files
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Test Skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: srcDir,
		Namespace: "github.com/test",
		Name:      "test-skill",
		Version:   "0.1.0",
	}

	dest, err := r.Store(skill, "github.com/test", "0.1.0")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	expectedDest := filepath.Join(root, "test-skill")
	if dest != expectedDest {
		t.Errorf("expected dest %q, got %q", expectedDest, dest)
	}

	// Verify files were copied
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); os.IsNotExist(err) {
		t.Error("SKILL.md not copied")
	}
	if _, err := os.Stat(filepath.Join(dest, "main.go")); os.IsNotExist(err) {
		t.Error("main.go not copied")
	}

	// Verify source remains
	if _, err := os.Stat(filepath.Join(srcDir, "SKILL.md")); os.IsNotExist(err) {
		t.Error("source SKILL.md was moved instead of copied")
	}
}

func TestStore_SingleFile(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// Create a source file
	srcFile := filepath.Join(t.TempDir(), "skill.tar.gz")
	if err := os.WriteFile(srcFile, []byte("compressed skill data"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: srcFile,
		Namespace: "registry",
		Name:      "packaged-skill",
		Version:   "2.0.0",
	}

	dest, err := r.Store(skill, "registry", "2.0.0")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "skill.tar.gz")); os.IsNotExist(err) {
		t.Error("file not copied")
	}
}

func TestRemove(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// Create a skill directory
	skillDir := r.SkillPath("test-ns", "test-skill", "1.0.0")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !r.Exists("test-ns", "test-skill", "1.0.0") {
		t.Error("Exists should return true before Remove")
	}

	if err := r.Remove("test-ns", "test-skill", "1.0.0"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if r.Exists("test-ns", "test-skill", "1.0.0") {
		t.Error("Exists should return false after Remove")
	}

	// Removing non-existent should not error
	if err := r.Remove("test-ns", "nonexistent", "0.0.1"); err != nil {
		t.Errorf("Remove non-existent should not error: %v", err)
	}
}

func TestUpdateLatest(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// Create a skill directory
	versionDir := r.SkillPath("github.com/test", "my-skill", "1.0.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// UpdateLatest is a no-op in flat layout, should not error
	if err := r.UpdateLatest("github.com/test", "my-skill", "1.0.0"); err != nil {
		t.Fatalf("UpdateLatest failed: %v", err)
	}
}

func TestUpdateLatest_VersionNotExist(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// UpdateLatest is a no-op in flat layout, even for non-existent version
	err := r.UpdateLatest("github.com/test", "my-skill", "0.0.0")
	if err != nil {
		t.Errorf("UpdateLatest should be no-op even for non-existent version: %v", err)
	}
}

func TestListSkills_Empty(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	skills, err := r.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills failed: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected empty list, got %v", skills)
	}
}

func TestListSkills_WithSkills(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	// Create some skill directories with SKILL.md (flat layout)
	skillNames := []string{"skill-a", "skill-b", "my-skill"}
	for _, name := range skillNames {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, "SKILL.md"), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a directory without SKILL.md (should be skipped)
	emptyDir := filepath.Join(root, "empty-skill")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .meta directory (should be skipped)
	metaDir := filepath.Join(root, ".meta")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		t.Fatal(err)
	}

	skills, err := r.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills failed: %v", err)
	}

	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(skills), skills)
	}
}

func TestExists(t *testing.T) {
	root := t.TempDir()
	r := NewRepository(root)

	if r.Exists("ns", "skill", "1.0.0") {
		t.Error("Exists should return false before creation")
	}

	skillDir := r.SkillPath("ns", "skill", "1.0.0")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if !r.Exists("ns", "skill", "1.0.0") {
		t.Error("Exists should return true after creation")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal-name", "normal-name"},
		{"bad:name*with?chars", "bad_name_with_chars"},
		{"spaces in name", "spaces-in-name"},
		{"back\\slash", "back_slash"},
	}

	for _, tt := range tests {
		result := sanitizeName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}