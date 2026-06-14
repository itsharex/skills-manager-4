package operations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupBrokenSymlinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a broken symlink
	badLink := filepath.Join(skillsDir, "bogus-skill@latest")
	if err := os.Symlink("/nonexistent-target", badLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	// Create a valid symlink that should NOT be removed
	realTarget := filepath.Join(root, "real-file")
	if err := os.WriteFile(realTarget, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	goodLink := filepath.Join(skillsDir, "real-skill@latest")
	if err := os.Symlink(realTarget, goodLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	count, err := CleanupBrokenSymlinks(skillsDir)
	if err != nil {
		t.Fatalf("CleanupBrokenSymlinks failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 broken symlink removed, got %d", count)
	}

	// Verify broken symlink is gone
	if _, err := os.Stat(badLink); !os.IsNotExist(err) {
		t.Error("expected broken symlink to be removed")
	}

	// Verify good symlink still exists
	if _, err := os.Stat(goodLink); os.IsNotExist(err) {
		t.Error("expected valid symlink to remain")
	}
}

func TestCleanupBrokenSymlinks_NoSkillsDir(t *testing.T) {
	t.Parallel()

	count, err := CleanupBrokenSymlinks("/nonexistent/skills")
	if err != nil {
		t.Fatalf("CleanupBrokenSymlinks on missing dir should not error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 broken symlinks for missing dir, got %d", count)
	}
}

func TestCleanupStaleTempDirs(t *testing.T) {
	// Note: cannot use t.Parallel() here because this modifies global temp state

	// Create a temp dir with the skillsmanager- prefix
	staleDir, err := os.MkdirTemp("", "skillsmanager-*")
	if err != nil {
		t.Fatalf("failed to create stale temp dir: %v", err)
	}
	defer os.RemoveAll(staleDir) // cleanup in case test fails

	// Verify the dir exists before cleanup
	if _, err = os.Stat(staleDir); os.IsNotExist(err) {
		t.Fatal("stale temp dir should exist before cleanup")
	}

	count, err := CleanupStaleTempDirs()
	if err != nil {
		t.Fatalf("CleanupStaleTempDirs failed: %v", err)
	}
	if count < 1 {
		t.Errorf("expected at least 1 stale temp dir removed, got %d", count)
	}

	// Verify the stale dir is gone
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("expected stale temp dir to be removed")
	}
}

func TestCleanupOrphanedSkills(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create tracked skill dirs (should NOT be removed)
	tracked1 := filepath.Join(skillsDir, "tracked-skill@1.0.0")
	tracked2 := filepath.Join(skillsDir, "tracked-skill@2.0.0")
	if err := os.MkdirAll(tracked1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tracked2, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create untracked skill dir (should be removed)
	untracked := filepath.Join(skillsDir, "untracked-skill@1.0.0")
	if err := os.MkdirAll(untracked, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a non-versioned dir (no @ symbol - should be skipped)
	plainDir := filepath.Join(skillsDir, "plain-directory")
	if err := os.MkdirAll(plainDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a "latest" symlink dir (should be skipped)
	latestLink := filepath.Join(skillsDir, "tracked-skill@latest")
	if err := os.Symlink(tracked1, latestLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	trackedVersions := map[string]bool{
		"tracked-skill@1.0.0": true,
		"tracked-skill@2.0.0": true,
	}

	count, err := CleanupOrphanedSkills(skillsDir, trackedVersions)
	if err != nil {
		t.Fatalf("CleanupOrphanedSkills failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 orphaned skill removed, got %d", count)
	}

	// Verify tracked dirs still exist
	if _, err := os.Stat(tracked1); os.IsNotExist(err) {
		t.Error("expected tracked skill dir to remain")
	}
	if _, err := os.Stat(tracked2); os.IsNotExist(err) {
		t.Error("expected tracked skill dir to remain")
	}

	// Verify untracked dir is removed
	if _, err := os.Stat(untracked); !os.IsNotExist(err) {
		t.Error("expected untracked skill dir to be removed")
	}

	// Verify plain dir (no @) still exists
	if _, err := os.Stat(plainDir); os.IsNotExist(err) {
		t.Error("expected non-versioned dir to remain")
	}

	// Verify latest symlink still exists
	if _, err := os.Stat(latestLink); os.IsNotExist(err) {
		t.Error("expected @latest symlink to remain")
	}
}

func TestCleanupOrphanedSkills_NilTracked(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a versioned dir
	versioned := filepath.Join(skillsDir, "some-skill@1.0.0")
	if err := os.MkdirAll(versioned, 0o755); err != nil {
		t.Fatal(err)
	}

	// When trackedVersions is nil, all versioned dirs are considered untracked
	count, err := CleanupOrphanedSkills(skillsDir, nil)
	if err != nil {
		t.Fatalf("CleanupOrphanedSkills failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 orphaned skill removed with nil tracked, got %d", count)
	}
}

func TestRunFullCleanup(t *testing.T) {
	// Cannot use t.Parallel due to CleanupStaleTempDirs modifying global state

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a broken symlink
	badLink := filepath.Join(skillsDir, "broken@latest")
	if err := os.Symlink("/nonexistent", badLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	// Create a stale temp dir
	staleDir, err := os.MkdirTemp("", "skillsmanager-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(staleDir)

	// Create an orphaned (untracked) skill dir
	untracked := filepath.Join(skillsDir, "orphan@1.0.0")
	if err := os.MkdirAll(untracked, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a tracked skill dir that should remain
	tracked := filepath.Join(skillsDir, "kept@1.0.0")
	if err := os.MkdirAll(tracked, 0o755); err != nil {
		t.Fatal(err)
	}

	stats := RunFullCleanup(skillsDir, map[string]bool{"kept@1.0.0": true})

	if stats == nil {
		t.Fatal("expected non-nil CleanupStats")
	}
	if stats.BrokenSymlinksRemoved != 1 {
		t.Errorf("expected 1 broken symlink removed, got %d", stats.BrokenSymlinksRemoved)
	}
	if stats.StaleTempDirsRemoved < 1 {
		t.Errorf("expected at least 1 stale temp dir removed, got %d", stats.StaleTempDirsRemoved)
	}
	if stats.OrphanedEntriesFixed != 1 {
		t.Errorf("expected 1 orphaned entry fixed, got %d", stats.OrphanedEntriesFixed)
	}

	// Verify orphan removed and tracked kept
	if _, err := os.Stat(untracked); !os.IsNotExist(err) {
		t.Error("expected orphaned skill dir to be removed")
	}
	if _, err := os.Stat(tracked); os.IsNotExist(err) {
		t.Error("expected tracked skill dir to remain")
	}

	// Verify stale dir is gone
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("expected stale temp dir to be removed")
	}
}

func TestRunFullCleanup_MissingSkillsDir(t *testing.T) {
	t.Parallel()

	stats := RunFullCleanup("/nonexistent/skills", nil)
	if stats == nil {
		t.Fatal("expected non-nil CleanupStats")
	}
	if stats.BrokenSymlinksRemoved != 0 {
		t.Errorf("expected 0 broken symlinks, got %d", stats.BrokenSymlinksRemoved)
	}
	if stats.OrphanedEntriesFixed != 0 {
		t.Errorf("expected 0 orphaned entries, got %d", stats.OrphanedEntriesFixed)
	}
}

func TestCleanupBrokenSymlinks_NonDirSkillsDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	count, err := CleanupBrokenSymlinks(filePath)
	if err != nil {
		t.Fatalf("CleanupBrokenSymlinks should not error for non-dir: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestCleanupOrphanedSkills_SkipsLatestSymlink(t *testing.T) {
	t.Parallel()

	// Verify that dirs named @latest are excluded even when not tracked
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a real target for the symlink first
	targetDir := filepath.Join(skillsDir, "realskill@1.0.0")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a @latest symlink
	latestLink := filepath.Join(skillsDir, "realskill@latest")
	if err := os.Symlink(targetDir, latestLink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	// The @latest should not be removed, even though it's not in trackedVersions
	// Track realskill@1.0.0 so it remains; only verify @latest behavior
	count, err := CleanupOrphanedSkills(skillsDir, map[string]bool{"realskill@1.0.0": true})
	if err != nil {
		t.Fatalf("CleanupOrphanedSkills failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 removals (latest symlink should be skipped), got %d", count)
	}

	// The real version dir should still exist
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Error("expected tracked version dir to remain")
	}
	// The @latest symlink should still exist
	if _, err := os.Stat(latestLink); os.IsNotExist(err) {
		t.Error("expected @latest symlink to remain")
	}
}

func TestCleanupStaleTempDirs_IgnoresOtherDirs(t *testing.T) {
	// Cannot use t.Parallel due to global state

	// Create a temp dir that does NOT have the skillsmanager- prefix
	otherDir, err := os.MkdirTemp("", "other-prefix-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(otherDir)

	count, err := CleanupStaleTempDirs()
	if err != nil {
		t.Fatalf("CleanupStaleTempDirs failed: %v", err)
	}

	// The other-prefix dir should not be removed
	if _, err := os.Stat(otherDir); os.IsNotExist(err) {
		t.Error("expected non-skillsmanager temp dir to remain")
	}

	// count could be anything, but at minimum CleanupStaleTempDirs should not error
	_ = count
}

func TestCleanupOrphanedSkills_OnlyVersionedDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a regular file (not a dir, not versioned)
	regularFile := filepath.Join(skillsDir, "readme.md")
	if err := os.WriteFile(regularFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a dir without @ (not versioned)
	plainDir := filepath.Join(skillsDir, "some-directory")
	if err := os.MkdirAll(plainDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a versioned dir (should be removed)
	versioned := filepath.Join(skillsDir, "orphan@1.0.0")
	if err := os.MkdirAll(versioned, 0o755); err != nil {
		t.Fatal(err)
	}

	// WalkDir iterates in lexical order; ensure orphan@1.0.0 is cleaned
	count, err := CleanupOrphanedSkills(skillsDir, map[string]bool{})
	if err != nil {
		t.Fatalf("CleanupOrphanedSkills failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 removal, got %d", count)
	}

	// Verify regular file and plain dir are untouched
	if _, err := os.Stat(regularFile); os.IsNotExist(err) {
		t.Error("expected regular file to remain")
	}
	if _, err := os.Stat(plainDir); os.IsNotExist(err) {
		t.Error("expected non-versioned dir to remain")
	}
}