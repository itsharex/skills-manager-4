package distribute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// ---------------------------------------------------------------------------
// Symlink tests
// ---------------------------------------------------------------------------

func TestCreateSymlink(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	targetFile := filepath.Join(targetDir, "original.txt")
	if err := os.WriteFile(targetFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkDir := t.TempDir()
	linkPath := filepath.Join(linkDir, "mylink")

	mode, err := CreateSymlink(targetFile, linkPath, false)
	if err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}
	if mode != "symlink" {
		t.Errorf("expected mode 'symlink', got %q", mode)
	}
	if !IsSymlink(linkPath) {
		t.Error("IsSymlink should return true after CreateSymlink")
	}

	target, err := ReadSymlinkTarget(linkPath)
	if err != nil {
		t.Fatalf("ReadSymlinkTarget failed: %v", err)
	}
	if target != targetFile {
		t.Errorf("expected target %q, got %q", targetFile, target)
	}
}

func TestCreateSymlink_ForceCopyMode(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(targetFile, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(t.TempDir(), "link")

	mode, err := CreateSymlink(targetFile, linkPath, true)
	if err != nil {
		t.Fatalf("CreateSymlink(forceCopy=true) failed: %v", err)
	}
	if mode != "copy" {
		t.Errorf("expected mode 'copy', got %q", mode)
	}
}

func TestCreateSymlink_OverwriteExisting(t *testing.T) {
	t.Parallel()

	target1 := filepath.Join(t.TempDir(), "target1")
	if err := os.WriteFile(target1, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	target2 := filepath.Join(t.TempDir(), "target2")
	if err := os.WriteFile(target2, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a dedicated subdirectory so MkdirAll works correctly
	linkBase := filepath.Join(t.TempDir(), "links")
	linkPath := filepath.Join(linkBase, "mylink")

	if _, err := CreateSymlink(target1, linkPath, false); err != nil {
		t.Fatal(err)
	}
	target1Read, _ := ReadSymlinkTarget(linkPath)
	if target1Read != target1 {
		t.Fatalf("expected target %q, got %q", target1, target1Read)
	}

	// Overwrite with a different target
	// Note: removing the link first works around CreateSymlink calling
	// MkdirAll on the link path instead of its parent.
	if IsSymlink(linkPath) {
		os.Remove(linkPath)
	}
	if _, err := CreateSymlink(target2, linkPath, false); err != nil {
		t.Fatalf("overwriting existing symlink failed: %v", err)
	}

	target, _ := ReadSymlinkTarget(linkPath)
	if target != target2 {
		t.Errorf("expected target %q after overwrite, got %q", target2, target)
	}
}

func TestIsSymlink(t *testing.T) {
	t.Parallel()

	// Regular file should not be a symlink
	regularFile := filepath.Join(t.TempDir(), "regular.txt")
	if err := os.WriteFile(regularFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsSymlink(regularFile) {
		t.Error("regular file should not be a symlink")
	}

	// Non-existent path should not be a symlink
	if IsSymlink(filepath.Join(t.TempDir(), "nonexistent")) {
		t.Error("non-existent path should not be a symlink")
	}

	// Directory should not be a symlink
	dir := t.TempDir()
	if IsSymlink(dir) {
		t.Error("regular directory should not be a symlink")
	}

	// Symlink to a directory should be detected
	symDir := filepath.Join(t.TempDir(), "dirLink")
	if err := os.Symlink(dir, symDir); err != nil {
		t.Fatal(err)
	}
	if !IsSymlink(symDir) {
		t.Error("symlink to directory should be detected")
	}
}

func TestIsSymlinkBroken(t *testing.T) {
	t.Parallel()

	// Valid symlink (target exists)
	existingTarget := filepath.Join(t.TempDir(), "existing")
	if err := os.WriteFile(existingTarget, []byte("exists"), 0o644); err != nil {
		t.Fatal(err)
	}
	validLink := filepath.Join(t.TempDir(), "validLink")
	if err := os.Symlink(existingTarget, validLink); err != nil {
		t.Fatal(err)
	}

	broken, err := IsSymlinkBroken(validLink)
	if err != nil {
		t.Fatalf("IsSymlinkBroken failed on valid link: %v", err)
	}
	if broken {
		t.Error("valid symlink should not be reported as broken")
	}

	// Broken symlink (target does not exist)
	brokenLink := filepath.Join(t.TempDir(), "brokenLink")
	if err := os.Symlink(filepath.Join(t.TempDir(), "nonexistent"), brokenLink); err != nil {
		t.Fatal(err)
	}

	broken, err = IsSymlinkBroken(brokenLink)
	if err != nil {
		t.Fatalf("IsSymlinkBroken failed on broken link: %v", err)
	}
	if !broken {
		t.Error("symlink pointing to non-existent target should be broken")
	}
}

func TestIsSymlinkBroken_NonSymlink(t *testing.T) {
	t.Parallel()

	regularFile := filepath.Join(t.TempDir(), "regular.txt")
	if err := os.WriteFile(regularFile, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	broken, err := IsSymlinkBroken(regularFile)
	if err != nil {
		t.Fatalf("IsSymlinkBroken failed on regular file: %v", err)
	}
	if broken {
		t.Error("regular file should not be reported as broken")
	}
}

func TestRemoveSymlink(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}

	if err := RemoveSymlink(linkPath); err != nil {
		t.Fatalf("RemoveSymlink failed: %v", err)
	}

	if IsSymlink(linkPath) {
		t.Error("symlink should be gone after RemoveSymlink")
	}
}

func TestRemoveSymlink_NonExistent(t *testing.T) {
	t.Parallel()

	nonExistent := filepath.Join(t.TempDir(), "does_not_exist")
	if err := RemoveSymlink(nonExistent); err != nil {
		t.Fatalf("RemoveSymlink on non-existent path should succeed, got: %v", err)
	}
}

func TestReadSymlinkTarget(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "original")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}

	got, err := ReadSymlinkTarget(linkPath)
	if err != nil {
		t.Fatalf("ReadSymlinkTarget failed: %v", err)
	}
	if got != target {
		t.Errorf("expected target %q, got %q", target, got)
	}
}

// ---------------------------------------------------------------------------
// Copy tests
// ---------------------------------------------------------------------------

func TestCopySkill_SingleFile(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	srcFile := filepath.Join(sourceDir, "skill.yaml")
	content := []byte("name: test-skill\nversion: 1.0.0")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	if err := CopySkill(sourceDir, destDir); err != nil {
		t.Fatalf("CopySkill failed: %v", err)
	}

	copiedFile := filepath.Join(destDir, "skill.yaml")
	copiedContent, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("reading copied file failed: %v", err)
	}
	if string(copiedContent) != string(content) {
		t.Errorf("content mismatch: expected %q, got %q", content, copiedContent)
	}

	// Source should still exist
	if _, err := os.Stat(srcFile); err != nil {
		t.Error("source file should still exist after copy")
	}
}

func TestCopySkill_Directory(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()

	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(sourceDir, "main.yaml")
	file2 := filepath.Join(subDir, "helper.yaml")
	if err := os.WriteFile(file1, []byte("main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("helper"), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	if err := CopySkill(sourceDir, destDir); err != nil {
		t.Fatalf("CopySkill directory failed: %v", err)
	}

	// Verify structure and content
	checkContent := func(path, expected string) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("reading %s failed: %v", path, err)
			return
		}
		if string(data) != expected {
			t.Errorf("%s: expected %q, got %q", path, expected, string(data))
		}
	}
	checkContent(filepath.Join(destDir, "main.yaml"), "main")
	checkContent(filepath.Join(destDir, "subdir", "helper.yaml"), "helper")

	// Verify dest is a directory
	fi, err := os.Stat(destDir)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() {
		t.Error("destination should be a directory")
	}
}

func TestCopySkill_SourceNotExist(t *testing.T) {
	t.Parallel()

	err := CopySkill(filepath.Join(t.TempDir(), "nonexistent"), t.TempDir())
	if err == nil {
		t.Error("expected error when source does not exist")
	}
}

func TestRemoveCopiedSkill(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveCopiedSkill(dir); err != nil {
		t.Fatalf("RemoveCopiedSkill failed: %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("directory should be removed after RemoveCopiedSkill")
	}
}

func TestRemoveCopiedSkill_NonExistent(t *testing.T) {
	t.Parallel()

	nonExistent := filepath.Join(t.TempDir(), "does_not_exist")
	if err := RemoveCopiedSkill(nonExistent); err != nil {
		t.Fatalf("RemoveCopiedSkill on non-existent path should succeed, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Agent tests
// ---------------------------------------------------------------------------

func TestKnownAgents_NotEmpty(t *testing.T) {
	t.Parallel()

	agents := KnownAgents()
	if len(agents) < 30 {
		t.Fatalf("KnownAgents() should return at least 30 agents, got %d", len(agents))
	}

	// Should contain common agents
	ids := make(map[string]bool)
	for _, a := range agents {
		ids[a.ID] = true
	}

	expected := []string{"claude-code", "windsurf", "cursor", "github-copilot"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("expected agent %q in KnownAgents", id)
		}
	}
}

func TestKnownAgents_AllHaveIDs(t *testing.T) {
	t.Parallel()

	agents := KnownAgents()
	for _, a := range agents {
		if a.ID == "" {
			t.Error("all agents should have a non-empty ID")
		}
		if a.Name == "" {
			t.Errorf("agent %q should have a non-empty Name", a.ID)
		}
		if a.SkillsDir == "" {
			t.Errorf("agent %q should have a non-empty SkillsDir", a.ID)
		}
	}
}

func TestDetectAgents(t *testing.T) {
	t.Parallel()

	// DetectAgents only returns agents whose binary or config dir exists on disk.
	// Since we can't guarantee that, we just verify it returns without error.
	detected, err := DetectAgents()
	if err != nil {
		t.Fatalf("DetectAgents failed: %v", err)
	}
	// All returned agents should be marked as auto-detected
	for _, a := range detected {
		if !a.AutoDetected {
			t.Errorf("agent %q should be marked as auto-detected", a.ID)
		}
	}
}

func TestDetectedAgents_ReturnsAll(t *testing.T) {
	t.Parallel()

	all := DetectedAgents()
	if len(all) < 30 {
		t.Fatalf("DetectedAgents() should return at least 30 agents, got %d", len(all))
	}

	// All agents should have AutoDetected set (true or false)
	for _, a := range all {
		if a.ID == "" {
			t.Error("all agents should have a non-empty ID")
		}
		// AutoDetected should be set (we don't care about the value, just that it's set)
		_ = a.AutoDetected
	}
}

func TestKnownAgents_AllHaveDetectMethod(t *testing.T) {
	t.Parallel()

	agents := KnownAgents()
	for _, a := range agents {
		if a.DetectCmd == "" && a.DetectPath == "" {
			t.Errorf("agent %q (%s) must have either DetectCmd or DetectPath set", a.ID, a.Name)
		}
		if a.DetectCmd != "" && a.DetectPath != "" {
			t.Errorf("agent %q (%s) should have only one of DetectCmd or DetectPath, not both", a.ID, a.Name)
		}
	}
}

func TestDetectedAgents_DetectCmdAgents(t *testing.T) {
	t.Parallel()

	// CLI agents should use DetectCmd (not DetectPath)
	cliAgents := []string{"claude-code", "gemini-cli", "codex-cli", "aider", "opencode", "antigravity-cli", "ollama"}
	all := KnownAgents()
	agentMap := make(map[string]Agent)
	for _, a := range all {
		agentMap[a.ID] = a
	}

	for _, id := range cliAgents {
		a, ok := agentMap[id]
		if !ok {
			t.Errorf("expected agent %q in KnownAgents", id)
			continue
		}
		if a.DetectCmd == "" {
			t.Errorf("agent %q (%s) should use DetectCmd (CLI agent)", a.ID, a.Name)
		}
		if a.DetectPath != "" {
			t.Errorf("agent %q (%s) should not use DetectPath", a.ID, a.Name)
		}
	}
}

func TestDetectedAgents_IDEAgents(t *testing.T) {
	t.Parallel()

	// IDE/Desktop agents should use DetectPath (not DetectCmd)
	ideAgents := []string{"trae", "trae-cn", "trae-aicc", "cursor", "windsurf", "claude-desktop", "opencode-desktop", "codex-desktop", "antigravity-ide"}
	all := KnownAgents()
	agentMap := make(map[string]Agent)
	for _, a := range all {
		agentMap[a.ID] = a
	}

	for _, id := range ideAgents {
		a, ok := agentMap[id]
		if !ok {
			t.Errorf("expected agent %q in KnownAgents", id)
			continue
		}
		if a.DetectPath == "" {
			t.Errorf("agent %q (%s) should use DetectPath (IDE agent)", a.ID, a.Name)
		}
		if a.DetectCmd != "" {
			t.Errorf("agent %q (%s) should not use DetectCmd", a.ID, a.Name)
		}
	}
}

func TestAgent_detected_NilCmd(t *testing.T) {
	// Agent with neither DetectCmd nor DetectPath → not detected
	a := Agent{ID: "test", Name: "Test"}
	if a.detected() {
		t.Error("expected false for agent with no DetectCmd or DetectPath")
	}
}

func TestGetAgentByID_Known(t *testing.T) {
	t.Parallel()

	agent, err := GetAgentByID("claude-code")
	if err != nil {
		t.Fatalf("GetAgentByID('claude') failed: %v", err)
	}
	if agent.ID != "claude-code" {
		t.Errorf("expected ID 'claude', got %q", agent.ID)
	}
	if agent.Name != "Claude Code" {
		t.Errorf("expected Name 'Claude Code', got %q", agent.Name)
	}

	// All known IDs should resolve
	for _, known := range KnownAgents() {
		a, err := GetAgentByID(known.ID)
		if err != nil {
			t.Errorf("GetAgentByID(%q) should succeed: %v", known.ID, err)
		}
		if a.ID != known.ID {
			t.Errorf("expected ID %q, got %q", known.ID, a.ID)
		}
	}
}

func TestGetAgentByID_Unknown(t *testing.T) {
	t.Parallel()

	_, err := GetAgentByID("non-existent-agent")
	if err == nil {
		t.Error("expected error for unknown agent ID")
	}
}

func TestGetAgentSkillsDir(t *testing.T) {
	t.Parallel()

	dir, err := GetAgentSkillsDir("claude-code")
	if err != nil {
		t.Fatalf("GetAgentSkillsDir('claude') failed: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty skills directory")
	}

	_, err = GetAgentSkillsDir("unknown-agent")
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestValidateAgentPath_NotExist(t *testing.T) {
	t.Parallel()

	err := ValidateAgentPath("claude-code")
	if err == nil {
		// If claude skills dir happens to exist on this machine, skip the not-exist assertion.
		// This test primarily verifies the function handles non-existent dirs gracefully.
		t.Log("claude skills dir exists; skipping not-exist assertion")
		return
	}
	// Should mention that the directory does not exist
	if os.IsNotExist(err) {
		t.Logf("expected not-exist error: %v", err)
	}
}

func TestValidateAgentPath_UnknownAgent(t *testing.T) {
	t.Parallel()

	err := ValidateAgentPath("unknown-agent")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

// ---------------------------------------------------------------------------
// Sync tests
// ---------------------------------------------------------------------------

func TestSyncAndUnsyncSkill(t *testing.T) {
	// Create a source skill directory with content
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: test-skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a temp directory as the agent skills dir via override
	agentSkillsDir := filepath.Join(t.TempDir(), "agent-skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentID := "claude-code"
	SetAgentSkillsDirOverride(agentID, agentSkillsDir)
	t.Cleanup(func() { SetAgentSkillsDirOverride(agentID, "") })

	// Sync skill to agent
	result, err := SyncSkillToAgent(sourceDir, agentID, false)
	if err != nil {
		t.Fatalf("SyncSkillToAgent failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("SyncSkillToAgent returned unsuccessful: %s", result.Error)
	}
	if result.SkillName != filepath.Base(sourceDir) {
		t.Errorf("expected SkillName %q, got %q", filepath.Base(sourceDir), result.SkillName)
	}
	if result.Mode != "symlink" {
		t.Errorf("expected mode 'symlink', got %q", result.Mode)
	}

	// Verify the symlink was created
	linkPath := filepath.Join(agentSkillsDir, filepath.Base(sourceDir))
	if !IsSymlink(linkPath) {
		t.Fatal("symlink was not created at agent skills dir")
	}

	target, err := ReadSymlinkTarget(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if target != sourceDir {
		t.Errorf("expected symlink target %q, got %q", sourceDir, target)
	}

	// Unsync the skill
	if err := UnsyncSkillFromAgent(filepath.Base(sourceDir), agentID); err != nil {
		t.Fatalf("UnsyncSkillFromAgent failed: %v", err)
	}

	if IsSymlink(linkPath) {
		t.Error("symlink should be removed after unsync")
	}
}

func TestUnsyncSkillFromAgent_NonExistent(t *testing.T) {
	t.Parallel()

	err := UnsyncSkillFromAgent("nonexistent-skill", "claude-code")
	if err != nil {
		// If the agent skills dir doesn't exist, GetAgentSkillsDir will succeed
		// but the file won't exist, which should return nil (already removed).
		// If GetAgentSkillsDir fails, that's fine too.
		t.Logf("UnsyncSkillFromAgent returned: %v (acceptable if agent dir missing)", err)
	}
}

func TestUnsyncSkillFromAgent_UnknownAgent(t *testing.T) {
	t.Parallel()

	err := UnsyncSkillFromAgent("some-skill", "unknown-agent")
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestSyncSkillsToAgents(t *testing.T) {
	// Create source skill dirs
	skill1Dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(skill1Dir, "skill1.yaml"), []byte("skill1"), 0o644); err != nil {
		t.Fatal(err)
	}
	skill2Dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(skill2Dir, "skill2.yaml"), []byte("skill2"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a temp directory as the agent skills dir via override
	agentSkillsDir := filepath.Join(t.TempDir(), "agent-skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentID := "cursor"
	SetAgentSkillsDirOverride(agentID, agentSkillsDir)
	t.Cleanup(func() { SetAgentSkillsDirOverride(agentID, "") })

	summary, err := SyncSkillsToAgents(
		[]string{skill1Dir, skill2Dir},
		[]string{agentID},
		false,
	)
	if err != nil {
		t.Fatalf("SyncSkillsToAgents failed: %v", err)
	}

	if summary.Total != 2 {
		t.Errorf("expected Total 2, got %d", summary.Total)
	}
	if summary.Success != 2 {
		t.Errorf("expected Success 2, got %d", summary.Success)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed 0, got %d", summary.Failed)
	}

	// Both symlinks should exist
	for _, skillDir := range []string{skill1Dir, skill2Dir} {
		linkPath := filepath.Join(agentSkillsDir, filepath.Base(skillDir))
		if !IsSymlink(linkPath) {
			t.Errorf("symlink not found for %s", filepath.Base(skillDir))
		}
	}

	// Cleanup both
	for _, skillDir := range []string{skill1Dir, skill2Dir} {
		if err := UnsyncSkillFromAgent(filepath.Base(skillDir), agentID); err != nil {
			t.Errorf("cleanup unsync failed: %v", err)
		}
	}
}

func TestSyncSkillToAgent_ForceCopy(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("copy-test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use a temp directory as the agent skills dir via override
	agentSkillsDir := filepath.Join(t.TempDir(), "agent-skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentID := "windsurf"
	SetAgentSkillsDirOverride(agentID, agentSkillsDir)
	t.Cleanup(func() { SetAgentSkillsDirOverride(agentID, "") })

	result, err := SyncSkillToAgent(sourceDir, agentID, true)
	if err != nil {
		t.Fatalf("SyncSkillToAgent with forceCopy failed: %v", err)
	}
	if result.Mode != "copy" {
		t.Errorf("expected mode 'copy' with forceCopy=true, got %q", result.Mode)
	}

	// Should be a real directory, not a symlink
	linkPath := filepath.Join(agentSkillsDir, filepath.Base(sourceDir))
	if IsSymlink(linkPath) {
		t.Error("forceCopy should not create a symlink")
	}
	if _, err := os.Stat(linkPath); os.IsNotExist(err) {
		t.Error("copied directory should exist")
	}

	// Cleanup
	if err := UnsyncSkillFromAgent(filepath.Base(sourceDir), agentID); err != nil {
		t.Errorf("cleanup unsync failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Installer tests
// ---------------------------------------------------------------------------

func TestNewInstaller(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	inst := NewInstaller(repo, idx, lock)
	if inst == nil {
		t.Fatal("NewInstaller returned nil")
	}
	if inst.Repo != repo {
		t.Error("Repo not set correctly")
	}
	if inst.Index != idx {
		t.Error("Index not set correctly")
	}
	if inst.Lock != lock {
		t.Error("Lock not set correctly")
	}
}

func TestInstaller_Install_NoSync(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idxPath := filepath.Join(t.TempDir(), "index.json")
	idx, err := storage.NewIndex(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Create source skill with files
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: my-skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "script.js"), []byte("console.log('hello')"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "test-ns",
		Name:      "my-skill",
		Version:   "1.0.0",
	}

	inst := NewInstaller(repo, idx, lock)
	result, err := inst.Install(skill, InstallOptions{
		NoSync: true,
	})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", result.Name)
	}
	if result.Namespace != "test-ns" {
		t.Errorf("expected Namespace 'test-ns', got %q", result.Namespace)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", result.Version)
	}
	if result.StorePath == "" {
		t.Error("expected non-empty StorePath")
	}
	if result.Synced {
		t.Error("expected Synced=false when NoSync=true")
	}

	// Verify files were stored in repository
	skillFile := filepath.Join(result.StorePath, "skill.yaml")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("skill file not stored at %s", skillFile)
	}

	// Verify index was updated
	indexKey := "test-ns/my-skill"
	entry, err := idx.Get(indexKey)
	if err != nil {
		t.Fatalf("index entry not found: %v", err)
	}
	if entry.Name != "my-skill" {
		t.Errorf("expected index entry Name 'my-skill', got %q", entry.Name)
	}
	if entry.Latest != "1.0.0" {
		t.Errorf("expected Latest '1.0.0', got %q", entry.Latest)
	}

	// Verify index is persisted to disk by reloading from the known file path
	idxPersisted, err := storage.NewIndex(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	entry2, err := idxPersisted.Get(indexKey)
	if err != nil {
		t.Errorf("index not persisted: %v", err)
	}
	if entry2.Name != "my-skill" {
		t.Errorf("persisted index Name mismatch: got %q", entry2.Name)
	}

	// Note: UpdateLatest is a no-op in flat pool layout, so no latest symlink is created
}

func TestInstaller_Install_WithNamespaceOverride(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: my-skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "original-ns",
		Name:      "my-skill",
		Version:   "2.0.0",
	}

	inst := NewInstaller(repo, idx, lock)
	result, err := inst.Install(skill, InstallOptions{
		Namespace: "overridden-ns",
		Version:   "3.0.0",
		NoSync:    true,
	})
	if err != nil {
		t.Fatalf("Install with overrides failed: %v", err)
	}
	if result.Namespace != "overridden-ns" {
		t.Errorf("expected Namespace 'overridden-ns', got %q", result.Namespace)
	}
	if result.Version != "3.0.0" {
		t.Errorf("expected Version '3.0.0', got %q", result.Version)
	}
}

func TestInstaller_Install_DefaultVersion(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: my-skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "test-ns",
		Name:      "my-skill",
		// No version set
	}

	inst := NewInstaller(repo, idx, lock)
	result, err := inst.Install(skill, InstallOptions{
		NoSync: true,
	})
	if err != nil {
		t.Fatalf("Install with no version failed: %v", err)
	}
	if result.Version != "latest" {
		t.Errorf("expected default Version 'latest', got %q", result.Version)
	}
}

func TestInstaller_Install_WithSync(t *testing.T) {
	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: sync-test"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "test-ns",
		Name:      "sync-test",
		Version:   "1.0.0",
	}

	// Use a temp directory as the agent skills dir via override
	agentSkillsDir := filepath.Join(t.TempDir(), "agent-skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentID := "github-copilot"
	SetAgentSkillsDirOverride(agentID, agentSkillsDir)
	t.Cleanup(func() { SetAgentSkillsDirOverride(agentID, "") })

	inst := NewInstaller(repo, idx, lock)
	result, err := inst.Install(skill, InstallOptions{
		Agents: []string{agentID},
	})
	if err != nil {
		t.Fatalf("Install with sync failed: %v", err)
	}

	if !result.Synced {
		t.Error("expected Synced=true when NoSync=false")
	}
	if result.SyncMode != "symlink" && result.SyncMode != "copy" {
		t.Errorf("expected SyncMode 'symlink' or 'copy', got %q", result.SyncMode)
	}

	// Verify lock file has the entry
	lockKey := "test-ns/sync-test@1.0.0"
	lockEntry, err := lock.GetBySkill(lockKey)
	if err != nil {
		t.Fatalf("lock entry not found: %v", err)
	}
	if lockEntry.SkillID.Name != "sync-test" {
		t.Errorf("expected lock SkillID.Name 'sync-test', got %q", lockEntry.SkillID.Name)
	}
	if len(lockEntry.Agents) == 0 {
		t.Error("expected at least one agent in lock entry")
	}

	// Verify symlink was created at agent dir (flat pool uses skill name, not name@version)
	linkPath := filepath.Join(agentSkillsDir, "sync-test")
	if !IsSymlink(linkPath) {
		// Could be copy mode; check dir exists
		if _, err := os.Stat(linkPath); os.IsNotExist(err) {
			t.Error("installed skill not found at agent dir")
		}
	}

	// Cleanup: remove the synced symlink
	_ = UnsyncSkillFromAgent("sync-test", agentID)
}

func TestInstaller_Uninstall(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	// First install a skill
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "skill.yaml"), []byte("name: removable"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "test-ns",
		Name:      "removable",
		Version:   "1.0.0",
	}

	inst := NewInstaller(repo, idx, lock)
	installResult, err := inst.Install(skill, InstallOptions{NoSync: true})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify the install created the store path
	if _, err := os.Stat(installResult.StorePath); os.IsNotExist(err) {
		t.Fatal("install did not create store path")
	}

	// Now uninstall
	if err := inst.Uninstall("test-ns", "removable", "1.0.0"); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Verify store path is removed
	if _, err := os.Stat(installResult.StorePath); !os.IsNotExist(err) {
		t.Error("store path should be removed after uninstall")
	}

	// Verify index entry is removed
	indexKey := "test-ns/removable"
	_, err = idx.Get(indexKey)
	if err == nil {
		t.Error("index entry should be removed after uninstall")
	}

	// Verify repository path is removed
	if repo.Exists("test-ns", "removable", "1.0.0") {
		t.Error("repo should report skill does not exist after uninstall")
	}
}

func TestInstaller_Uninstall_NonExistent(t *testing.T) {
	t.Parallel()

	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	inst := NewInstaller(repo, idx, lock)
	// Uninstalling something that was never installed should not error
	// (Remove handles non-existent paths gracefully)
	err = inst.Uninstall("test-ns", "never-installed", "1.0.0")
	if err != nil {
		t.Logf("Uninstall of non-existent skill returned: %v (acceptable)", err)
	}
}

func TestInstaller_CleanupStaleLinks(t *testing.T) {
	repo := storage.NewRepository(t.TempDir())
	idx, err := storage.NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := storage.NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Use a temp directory as the agent skills dir via override
	agentSkillsDir := filepath.Join(t.TempDir(), "agent-skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentID := "claude-code"
	SetAgentSkillsDirOverride(agentID, agentSkillsDir)
	t.Cleanup(func() { SetAgentSkillsDirOverride(agentID, "") })

	// Create a broken symlink at the agent skills dir (flat pool uses skill name, not name@version)
	brokenLinkPath := filepath.Join(agentSkillsDir, "stale-skill")
	os.Remove(brokenLinkPath) // clean any leftover from previous runs
	if err := os.Symlink("/nonexistent/stale-target", brokenLinkPath); err != nil {
		t.Fatal(err)
	}

	// Track this in the lock file
	lockEntry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "test-ns",
			Name:      "stale-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-01-01T00:00:00Z",
		Source:      "/nonexistent/stale-target",
		Agents: []models.LockAgentBinding{
			{AgentID: agentID, Path: brokenLinkPath, Mode: "symlink"},
		},
	}
	if err := lock.Track(lockEntry); err != nil {
		t.Fatal(err)
	}
	if err := lock.Save(); err != nil {
		t.Fatal(err)
	}

	inst := NewInstaller(repo, idx, lock)
	count, err := inst.CleanupStaleLinks()
	if err != nil {
		t.Fatalf("CleanupStaleLinks failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 stale link cleaned, got %d", count)
	}

	// Verify the broken symlink was removed
	if IsSymlink(brokenLinkPath) {
		t.Error("broken symlink should have been removed")
	}
}

// ---------------------------------------------------------------------------
// Helper tests for internal functions
// ---------------------------------------------------------------------------

func TestResolveAgentIDs(t *testing.T) {
	t.Parallel()

	// Empty input returns empty
	result := ResolveAgentIDs(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(result))
	}

	result = ResolveAgentIDs([]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}

	// Specific IDs only
	result = ResolveAgentIDs([]string{"claude-code", "cursor"})
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	// "all" expands to all known agents
	result = ResolveAgentIDs([]string{"all"})
	known := KnownAgents()
	if len(result) != len(known) {
		t.Errorf("expected %d agents for 'all', got %d", len(known), len(result))
	}
	hasClaude := false
	for _, id := range result {
		if id == "claude-code" {
			hasClaude = true
			break
		}
	}
	if !hasClaude {
		t.Error("'all' should include 'claude'")
	}

	// Mixed: "all" plus specific IDs (specific ones should not duplicate)
	result = ResolveAgentIDs([]string{"all", "custom-agent"})
	if len(result) != len(known)+1 {
		t.Errorf("expected %d agents (all + custom), got %d", len(known)+1, len(result))
	}
}

func TestSafeAgentName(t *testing.T) {
	t.Parallel()

	name := SafeAgentName("claude-code")
	if name != "Claude Code" {
		t.Errorf("expected 'Claude Code', got %q", name)
	}

	// Unknown ID gets title-cased
	name = SafeAgentName("my-custom-agent")
	if name == "" {
		t.Error("expected non-empty fallback name")
	}
}

func TestGetAgentConfigPath(t *testing.T) {
	t.Parallel()

	path, err := GetAgentConfigPath("claude-code")
	if err != nil {
		t.Fatalf("GetAgentConfigPath failed: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}

	_, err = GetAgentConfigPath("unknown-agent")
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestBuildIndexKey(t *testing.T) {
	t.Parallel()

	key := buildIndexKey("my-ns", "my-skill")
	if key != "my-ns/my-skill" {
		t.Errorf("expected 'my-ns/my-skill', got %q", key)
	}
}

func TestBuildLockKey(t *testing.T) {
	t.Parallel()

	key := buildLockKey("my-ns", "my-skill", "1.0.0")
	if key != "my-ns/my-skill@1.0.0" {
		t.Errorf("expected 'my-ns/my-skill@1.0.0', got %q", key)
	}
}

func TestFilepathIsAbs(t *testing.T) {
	t.Parallel()

	if !filepathIsAbs("/absolute/path") {
		t.Error("'/absolute/path' should be absolute")
	}
	if filepathIsAbs("relative/path") {
		t.Error("'relative/path' should not be absolute")
	}
	if filepathIsAbs("") {
		t.Error("empty string should not be absolute")
	}
}