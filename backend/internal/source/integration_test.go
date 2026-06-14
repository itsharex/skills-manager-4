package source

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// ─── Test 1: LocalSource basic resolve ─────────────────────────────────────

func TestIntegration_LocalSource_Resolve(t *testing.T) {
	tmpDir := t.TempDir()
	skillContent := `---
name: my-integration-skill
description: Integration test skill
version: 1.0.0
author: tester
tags: [integration, test]
---

# Integration Test Skill

This is a test skill for integration testing.
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

	// For a single-skill repo, the name is the directory base name
	dirName := filepath.Base(tmpDir)
	if skills[0].Name != dirName {
		t.Errorf("expected Name=%q, got %q", dirName, skills[0].Name)
	}
	if skills[0].Namespace != "local" {
		t.Errorf("expected Namespace='local', got %q", skills[0].Namespace)
	}
	if skills[0].Version != "1.0.0" {
		t.Errorf("expected Version='1.0.0', got %q", skills[0].Version)
	}
}

// ─── Test 2: LocalSource ZIP resolve ───────────────────────────────────────

func TestIntegration_LocalSource_ZIP(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("myskill/SKILL.md")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	_, err = w.Write([]byte(`---
name: zip-skill
description: Skill extracted from ZIP
version: 2.0.0
author: zipper
tags: [zip, compressed]
---

# ZIP Skill

Content from ZIP archive.
`))
	if err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	zw.Close()
	f.Close()

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, zipPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	if skills[0].Name != "myskill" {
		t.Errorf("expected Name='myskill', got %q", skills[0].Name)
	}
	if skills[0].Version != "2.0.0" {
		t.Errorf("expected Version='2.0.0', got %q", skills[0].Version)
	}
	if skills[0].Namespace != "local" {
		t.Errorf("expected Namespace='local', got %q", skills[0].Namespace)
	}

	// Verify Cleanup is set
	if skills[0].Cleanup == nil {
		t.Fatal("expected Cleanup function to be set for ZIP source")
	}

	// Verify temp directory exists before cleanup
	if _, err := os.Stat(skills[0].LocalPath); os.IsNotExist(err) {
		t.Error("expected temp directory to exist before cleanup")
	}

	// Call cleanup
	skills[0].Cleanup()

	// Verify temp directory is removed after cleanup
	if _, err := os.Stat(skills[0].LocalPath); !os.IsNotExist(err) {
		t.Error("expected temp directory to be removed after cleanup")
	}
}

// ─── Test 3: HTTPResolver with mock server ─────────────────────────────────

func TestIntegration_HTTPResolver_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"name": "http-skill-one", "description": "First HTTP skill", "version": "1.0.0", "url": "https://example.com/one.zip", "tags": ["http"]},
			{"name": "http-skill-two", "description": "Second HTTP skill", "version": "2.0.0", "url": "https://example.com/two.zip", "tags": ["web", "api"]}
		]`))
	}))
	defer server.Close()

	r := &HTTPResolver{}

	// CanHandle should return true for the server URL
	if !r.CanHandle(server.URL) {
		t.Errorf("CanHandle(%q) should return true", server.URL)
	}

	ctx := context.Background()
	skills, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	// Verify first skill metadata
	if skills[0].Name != "http-skill-one" {
		t.Errorf("expected Name='http-skill-one', got %q", skills[0].Name)
	}
	if skills[0].Version != "1.0.0" {
		t.Errorf("expected Version='1.0.0', got %q", skills[0].Version)
	}

	// Verify second skill metadata
	if skills[1].Name != "http-skill-two" {
		t.Errorf("expected Name='http-skill-two', got %q", skills[1].Name)
	}
	if skills[1].Version != "2.0.0" {
		t.Errorf("expected Version='2.0.0', got %q", skills[1].Version)
	}

	// Verify namespace contains the registry URL
	expectedNamespace := "registry:" + server.URL
	for i, skill := range skills {
		if skill.Namespace != expectedNamespace {
			t.Errorf("skills[%d].Namespace expected %q, got %q", i, expectedNamespace, skill.Namespace)
		}
	}
}

// ─── Test 4: HTTPResolver with malformed responses ─────────────────────────

func TestIntegration_HTTPResolver_MockMalformed(t *testing.T) {
	// Sub-test: invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`this is not valid json`))
		}))
		defer server.Close()

		r := &HTTPResolver{}
		ctx := context.Background()
		_, err := r.Resolve(ctx, server.URL, ResolveOptions{})
		if err == nil {
			t.Fatal("expected error for invalid JSON response")
		}
	})

	// Sub-test: 500 Internal Server Error
	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		r := &HTTPResolver{}
		ctx := context.Background()
		_, err := r.Resolve(ctx, server.URL, ResolveOptions{})
		if err == nil {
			t.Fatal("expected error for 500 Internal Server Error")
		}
	})
}

// ─── Test 5: Full install flow (resolve → store → index → lock → validate) ─

func TestIntegration_FullInstallFlow(t *testing.T) {
	// Create a valid SKILL.md in a source temp directory
	sourceDir := t.TempDir()
	skillContent := `---
name: full-flow-skill
description: Full integration flow test skill
version: 3.0.0
author: integrator
tags: [integration, full-flow]
---

# Full Flow Skill

This skill tests the complete install flow.
`
	if err := os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// Initialize a storage repository in another temp directory
	repoDir := t.TempDir()
	repo := storage.NewRepository(repoDir)

	// Create a ResolvedSkill
	skill := models.ResolvedSkill{
		LocalPath: sourceDir,
		Namespace: "local",
		Name:      filepath.Base(sourceDir),
		Version:   "3.0.0",
		Cleanup:   func() {},
	}

	// Store the skill via Repository.Store()
	destPath, err := repo.Store(skill, skill.Namespace, skill.Version)
	if err != nil {
		t.Fatalf("Store() unexpected error: %v", err)
	}
	if destPath == "" {
		t.Fatal("expected non-empty destination path from Store()")
	}

	// Verify the skill was stored by reading back from disk
	storedSkillPath := filepath.Join(destPath, "SKILL.md")
	if _, err := os.Stat(storedSkillPath); os.IsNotExist(err) {
		t.Fatalf("stored SKILL.md not found at %s", storedSkillPath)
	}
	parsed, err := storage.ParseSkillFile(storedSkillPath)
	if err != nil {
		t.Fatalf("ParseSkillFile() of stored skill: %v", err)
	}
	if parsed.Name != "full-flow-skill" {
		t.Errorf("expected parsed Name='full-flow-skill', got %q", parsed.Name)
	}
	if parsed.Version != "3.0.0" {
		t.Errorf("expected parsed Version='3.0.0', got %q", parsed.Version)
	}

	// Add to index via storage.Index.Add()
	idx, err := storage.NewIndex(repo.Paths.IndexPath)
	if err != nil {
		t.Fatalf("NewIndex() unexpected error: %v", err)
	}
	indexEntry := models.IndexEntry{
		Name:        skill.Name,
		Namespace:   skill.Namespace,
		Versions:    []string{skill.Version},
		Latest:      skill.Version,
		Source:      sourceDir,
		SourceType:  "local",
		Description: "Full integration flow test skill",
		Tags:        []string{"integration", "full-flow"},
	}
	if err := idx.Add(indexEntry); err != nil {
		t.Fatalf("Index.Add() unexpected error: %v", err)
	}
	if err := idx.Save(); err != nil {
		t.Fatalf("Index.Save() unexpected error: %v", err)
	}

	// Verify index by reading back from disk
	idx2, err := storage.NewIndex(repo.Paths.IndexPath)
	if err != nil {
		t.Fatalf("NewIndex() reload unexpected error: %v", err)
	}
	entries, err := idx2.List()
	if err != nil {
		t.Fatalf("Index.List() unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 index entry, got %d", len(entries))
	}
	if entries[0].Name != skill.Name {
		t.Errorf("expected index entry Name=%q, got %q", skill.Name, entries[0].Name)
	}

	// Create a lock entry via storage.LockFile.Track()
	lf, err := storage.NewLockFile(repo.Paths.LockPath)
	if err != nil {
		t.Fatalf("NewLockFile() unexpected error: %v", err)
	}
	lockEntry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: skill.Namespace,
			Name:      skill.Name,
			Version:   skill.Version,
		},
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		Source:      sourceDir,
		Agents: []models.LockAgentBinding{
			{AgentID: "test-agent", Path: destPath, Mode: "copy"},
		},
	}
	if err := lf.Track(lockEntry); err != nil {
		t.Fatalf("LockFile.Track() unexpected error: %v", err)
	}
	if err := lf.Save(); err != nil {
		t.Fatalf("LockFile.Save() unexpected error: %v", err)
	}

	// Verify lock entry by reading back from disk
	lf2, err := storage.NewLockFile(repo.Paths.LockPath)
	if err != nil {
		t.Fatalf("NewLockFile() reload unexpected error: %v", err)
	}
	lockEntries, err := lf2.List()
	if err != nil {
		t.Fatalf("LockFile.List() unexpected error: %v", err)
	}
	if len(lockEntries) != 1 {
		t.Fatalf("expected 1 lock entry, got %d", len(lockEntries))
	}
	if lockEntries[0].SkillID.Name != skill.Name {
		t.Errorf("expected lock SkillID.Name=%q, got %q", skill.Name, lockEntries[0].SkillID.Name)
	}

	// Use Validator to validate the SKILL.md
	result, err := ValidateSkillFile(filepath.Join(sourceDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ValidateSkillFile() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid SKILL.md, got errors=%v", result.Errors)
	}
}

// ─── Test 6: Validator round-trip (valid → invalid → validate both) ────────

func TestIntegration_Validator_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	// ── Step 1: Create a valid SKILL.md ──
	validContent := `---
name: roundtrip-skill
description: Roundtrip validation test skill
version: 1.0.0
author: validator
tags: [test, roundtrip]
---

# Roundtrip Skill

This skill validates the validate-invalidate-revalidate cycle.
`
	if err := os.WriteFile(skillPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("write valid SKILL.md: %v", err)
	}

	// Validate - should pass
	result, err := ValidateSkillFile(skillPath)
	if err != nil {
		t.Fatalf("ValidateSkillFile() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid result, got errors=%v warnings=%v", result.Errors, result.Warnings)
	}

	// ── Step 2: Modify to remove a required field (name) ──
	invalidContent := `---
description: Missing name field intentionally
version: 1.0.0
author: validator
---

Content without a name field.
`
	if err := os.WriteFile(skillPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("write invalid SKILL.md: %v", err)
	}

	// Validate - should fail
	result, err = ValidateSkillFile(skillPath)
	if err != nil {
		t.Fatalf("ValidateSkillFile() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid result for SKILL.md missing required 'name' field")
	}
	foundNameError := false
	for _, e := range result.Errors {
		if e.Field == "name" {
			foundNameError = true
			break
		}
	}
	if !foundNameError {
		t.Errorf("expected error on field 'name', got errors=%v", result.Errors)
	}

	// ── Step 3: Modify to remove another required field (description) ──
	invalidContent2 := `---
name: no-desc-skill
version: 1.0.0
author: validator
---

Content without a description.
`
	if err := os.WriteFile(skillPath, []byte(invalidContent2), 0644); err != nil {
		t.Fatalf("write second invalid SKILL.md: %v", err)
	}

	result, err = ValidateSkillFile(skillPath)
	if err != nil {
		t.Fatalf("ValidateSkillFile() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid result for SKILL.md missing required 'description' field")
	}
	foundDescError := false
	for _, e := range result.Errors {
		if e.Field == "description" {
			foundDescError = true
			break
		}
	}
	if !foundDescError {
		t.Errorf("expected error on field 'description', got errors=%v", result.Errors)
	}
}