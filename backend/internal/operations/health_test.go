package operations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestRunDoctor_ValidRepo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := InitPool(root); err != nil {
		t.Fatalf("InitPool failed: %v", err)
	}

	report := RunDoctor(root)

	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.PoolPath != root {
		t.Errorf("expected PoolPath %q, got %q", root, report.PoolPath)
	}
	if !report.AllPass {
		t.Error("expected AllPass true for a valid repo")
	}
	if len(report.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}

	// Verify each check is either pass or warn (no fail for a valid repo)
	for _, c := range report.Checks {
		if c.Status == "fail" {
			t.Errorf("unexpected fail for check %q: %s", c.Name, c.Message)
		}
	}
}

func TestRunDoctor_NoRepo(t *testing.T) {
	t.Parallel()

	nonExistent := filepath.Join(t.TempDir(), "nonexistent-pool")
	report := RunDoctor(nonExistent)

	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.AllPass {
		t.Error("expected AllPass false for missing pool")
	}

	// Verify pool_root check fails
	foundPoolRootFail := false
	for _, c := range report.Checks {
		if c.Name == "pool_root" && c.Status == "fail" {
			foundPoolRootFail = true
			break
		}
	}
	if !foundPoolRootFail {
		t.Error("expected pool_root check to fail")
	}
}

func TestRunDoctor_BrokenSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := InitPool(root); err != nil {
		t.Fatalf("InitPool failed: %v", err)
	}

	paths := GetPoolPaths(root)

	// Create a broken symlink in the skills directory
	badLink := filepath.Join(paths.SkillsDir, "bogus-skill@latest")
	if err := os.Symlink("/nonexistent-target", badLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	report := RunDoctor(root)

	foundBrokenWarn := false
	for _, c := range report.Checks {
		if c.Name == "broken_symlinks" && c.Status == "warn" {
			foundBrokenWarn = true
			break
		}
	}
	if !foundBrokenWarn {
		t.Error("expected broken_symlinks check to return warn")
	}
}

func TestCheckAgentAccess_Valid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	agentDir := filepath.Join(dir, "skills")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result := CheckAgentAccess("testagent", agentDir)
	if result.Status != "pass" {
		t.Errorf("expected pass, got %q: %s", result.Status, result.Message)
	}
	if result.Name != "agent_testagent" {
		t.Errorf("expected name agent_testagent, got %q", result.Name)
	}
}

func TestCheckAgentAccess_NotExist(t *testing.T) {
	t.Parallel()

	result := CheckAgentAccess("missing", "/nonexistent/path/for/agent")
	if result.Status != "warn" {
		t.Errorf("expected warn, got %q: %s", result.Status, result.Message)
	}
}

func TestCheckAgentAccess_NotDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := CheckAgentAccess("notadir", filePath)
	if result.Status != "fail" {
		t.Errorf("expected fail, got %q: %s", result.Status, result.Message)
	}
}

func TestValidateSkillIndex_Valid(t *testing.T) {
	t.Parallel()

	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"test-skill": {
				Name:      "test-skill",
				Namespace: "test",
				Versions:  []string{"1.0.0"},
				Latest:    "1.0.0",
			},
		},
	}

	results := ValidateSkillIndex(index)
	if len(results) != 0 {
		t.Errorf("expected no issues, got %d", len(results))
	}
}

func TestValidateSkillIndex_Nil(t *testing.T) {
	t.Parallel()

	results := ValidateSkillIndex(nil)
	if len(results) == 0 {
		t.Fatal("expected at least one result for nil index")
	}
	if results[0].Status != "fail" {
		t.Errorf("expected fail status, got %q", results[0].Status)
	}
	if results[0].Name != "index_integrity" {
		t.Errorf("expected index_integrity check, got %q", results[0].Name)
	}
}

func TestValidateSkillIndex_EmptyName(t *testing.T) {
	t.Parallel()

	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"empty-name": {
				Name:      "",
				Namespace: "test",
				Versions:  []string{"1.0.0"},
				Latest:    "1.0.0",
			},
		},
	}

	results := ValidateSkillIndex(index)

	foundEmptyNameWarn := false
	for _, r := range results {
		if r.Status == "warn" && r.Name == "index_entry_empty-name" {
			foundEmptyNameWarn = true
			break
		}
	}
	if !foundEmptyNameWarn {
		t.Error("expected warn for entry with empty name")
	}
}

func TestValidateSkillIndex_NoLatest(t *testing.T) {
	t.Parallel()

	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"no-latest": {
				Name:      "no-latest",
				Namespace: "test",
				Versions:  []string{"1.0.0", "2.0.0"},
				Latest:    "", // no latest set
			},
		},
	}

	results := ValidateSkillIndex(index)

	foundNoLatestWarn := false
	for _, r := range results {
		if r.Status == "warn" && r.Name == "index_entry_no-latest" {
			foundNoLatestWarn = true
			break
		}
	}
	if !foundNoLatestWarn {
		t.Error("expected warn for entry with versions but no latest")
	}
}