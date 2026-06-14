package source

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
)

// ─── github.go: scanSkillFiles edge cases ──────────────────────────────────

// TestScanSkillFiles_RootSkillParseFallback covers the fallback path when
// a root SKILL.md exists but storage.ParseSkillFile fails (e.g., no frontmatter).
// Expected: a ResolvedSkill with repo name and version "latest".
func TestScanSkillFiles_RootSkillParseFallback(t *testing.T) {
	tmpDir := t.TempDir()
	// Content without frontmatter (no "---" prefix) — ParseSkillFile will fail
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte("No frontmatter here"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("expected fallback success, got error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Version != "latest" {
		t.Errorf("expected version 'latest' for parse fallback, got %q", skills[0].Version)
	}
	if skills[0].Name != filepath.Base(tmpDir) {
		t.Errorf("expected name %q, got %q", filepath.Base(tmpDir), skills[0].Name)
	}
}

// TestScanSkillFiles_MultiSubdirWithParseFailure tests the multi-skill
// path where a subdirectory has a valid SKILL.md and another has one that
// fails to parse—the failing one should still be included with version "latest".
func TestScanSkillFiles_MultiSubdirWithParseFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Subdir with valid SKILL.md
	goodDir := filepath.Join(tmpDir, "good-skill")
	os.MkdirAll(goodDir, 0755)
	os.WriteFile(filepath.Join(goodDir, "SKILL.md"), []byte(`---
name: good-skill
description: A good skill
version: 2.0.0
---

Content`), 0644)

	// Subdir with unparseable SKILL.md (no frontmatter)
	badDir := filepath.Join(tmpDir, "bad-skill")
	os.MkdirAll(badDir, 0755)
	os.WriteFile(filepath.Join(badDir, "SKILL.md"), []byte("No frontmatter"), 0644)

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	skillMap := make(map[string]string)
	for _, s := range skills {
		skillMap[s.Name] = s.Version
	}
	if v, ok := skillMap["good-skill"]; !ok {
		t.Error("missing good-skill")
	} else if v != "2.0.0" {
		t.Errorf("good-skill version: want 2.0.0, got %s", v)
	}
	if v, ok := skillMap["bad-skill"]; !ok {
		t.Error("missing bad-skill (should be included despite parse failure)")
	} else if v != "latest" {
		t.Errorf("bad-skill version: want 'latest', got %s", v)
	}
}

// TestScanSkillFiles_NoSkillsFound covers the error path when no SKILL.md
// exists at root or in any subdirectory.
func TestScanSkillFiles_NoSkillsFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Create an empty subdirectory with no SKILL.md
	os.MkdirAll(filepath.Join(tmpDir, "empty-subdir"), 0755)

	r := &LocalResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for no SKILL.md files")
	}
	if !strings.Contains(err.Error(), "no SKILL.md files found") {
		t.Errorf("expected 'no SKILL.md files found' error, got: %v", err)
	}
}

// TestScanSkillFiles_SkipsDotGit covers the .git directory skip logic.
func TestScanSkillFiles_SkipsDotGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .git directory (no SKILL.md inside)
	os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)

	// Create a real skill subdirectory
	skillDir := filepath.Join(tmpDir, "real-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: real-skill
description: Real skill
version: 1.0.0
---

Content`), 0644)

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, tmpDir, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "real-skill" {
		t.Errorf("expected 'real-skill', got %q", skills[0].Name)
	}
}

// ─── http.go: error paths and edge cases ───────────────────────────────────

// TestHTTPResolver_Resolve_InvalidJSON tests the JSON unmarshal error path.
func TestHTTPResolver_Resolve_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "parse registry index") {
		t.Errorf("expected 'parse registry index' error, got: %v", err)
	}
}

// TestHTTPResolver_Resolve_EntryNoVersion tests that an entry without
// a version field defaults to "latest".
func TestHTTPResolver_Resolve_EntryNoVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "no-version-skill", "description": "No version field"}]`))
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Version != "latest" {
		t.Errorf("expected version 'latest', got %q", skills[0].Version)
	}
	if skills[0].Name != "no-version-skill" {
		t.Errorf("expected name 'no-version-skill', got %q", skills[0].Name)
	}
}

// TestHTTPResolver_Resolve_ExtraFields tests that extra JSON fields
// in registry entries are ignored (no parse error).
func TestHTTPResolver_Resolve_ExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "extra", "description": "Has extras", "version": "1.0.0", "url": "https://x.com", "tags": ["a"], "unknownField": "ignored", "extraObj": {"nested": true}}]`))
	}))
	defer server.Close()

	r := &HTTPResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, server.URL, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "extra" {
		t.Errorf("expected name 'extra', got %q", skills[0].Name)
	}
}

// ─── local.go: edge cases ──────────────────────────────────────────────────

// TestLocalResolver_Resolve_UnsupportedFileType tests the error path when
// a file is not a directory, not SKILL.md, and not a ZIP.
func TestLocalResolver_Resolve_UnsupportedFileType(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, filePath, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for unsupported file type")
	}
	if !strings.Contains(err.Error(), "unsupported local source") {
		t.Errorf("expected 'unsupported local source' error, got: %v", err)
	}
}

// TestLocalResolver_Resolve_SingleFileParseError tests that a single SKILL.md
// file with invalid content returns an error from resolveSingleSkillFile.
func TestLocalResolver_Resolve_SingleFileParseError(t *testing.T) {
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")
	// Write content without frontmatter → ParseSkillFile error
	if err := os.WriteFile(skillPath, []byte("No frontmatter here"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, skillPath, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for invalid SKILL.md content")
	}
	if !strings.Contains(err.Error(), "parse skill file") {
		t.Errorf("expected 'parse skill file' error, got: %v", err)
	}
}

// TestLocalResolver_Resolve_HomeDirExpansion tests the ~/ prefix expansion path.
func TestLocalResolver_Resolve_HomeDirExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	r := &LocalResolver{}
	ctx := context.Background()
	// Use ~ with a guaranteed non-existent path
	_, err = r.Resolve(ctx, "~/nonexistent-test-path-abc-123", ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for non-existent ~ path")
	}
	// The error message should contain the expanded path (starting with home dir)
	if !strings.Contains(err.Error(), home) {
		t.Errorf("expected error to contain home dir %q, got: %v", home, err)
	}
}

// ─── local.go: ZIP file handling ──────────────────────────────────────────

// TestLocalResolver_Resolve_CorruptedZip tests that opening an invalid ZIP
// file returns a proper error.
func TestLocalResolver_Resolve_CorruptedZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")
	// Write garbage data with .zip extension
	if err := os.WriteFile(zipPath, []byte("this is not a valid zip file"), 0644); err != nil {
		t.Fatalf("write fake zip: %v", err)
	}

	r := &LocalResolver{}
	ctx := context.Background()
	_, err := r.Resolve(ctx, zipPath, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for corrupted zip")
	}
	if !strings.Contains(err.Error(), "open zip") {
		t.Errorf("expected 'open zip' error, got: %v", err)
	}
}

// TestLocalResolver_Resolve_ValidZipNoSkill tests a valid ZIP with no SKILL.md.
func TestLocalResolver_Resolve_ValidZipNoSkill(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	// Add a non-SKILL.md file
	w, _ := zw.Create("readme.txt")
	w.Write([]byte("hello"))
	zw.Close()
	f.Close()

	r := &LocalResolver{}
	ctx := context.Background()
	_, err = r.Resolve(ctx, zipPath, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for zip with no SKILL.md")
	}
	if !strings.Contains(err.Error(), "no SKILL.md files found") {
		t.Errorf("expected 'no SKILL.md files found' error, got: %v", err)
	}
}

// TestLocalResolver_Resolve_ZipWithNestedDir tests a ZIP with a subdirectory
// containing SKILL.md.
func TestLocalResolver_Resolve_ZipWithNestedDir(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)

	// Add a directory entry and a SKILL.md inside it
	zw.Create("myskill/")
	w, _ := zw.Create("myskill/SKILL.md")
	w.Write([]byte(`---
name: nested-skill
description: A nested skill
version: 3.0.0
---

Content`))
	zw.Close()
	f.Close()

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, zipPath, ResolveOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "myskill" {
		t.Errorf("expected name 'myskill', got %q", skills[0].Name)
	}
	if skills[0].Version != "3.0.0" {
		t.Errorf("expected version '3.0.0', got %q", skills[0].Version)
	}
}

// TestLocalResolver_Resolve_ZipWithVersionOverride tests opts.Version
// override on a ZIP source.
func TestLocalResolver_Resolve_ZipWithVersionOverride(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("SKILL.md")
	w.Write([]byte(`---
name: override-test
description: Version override test
version: 1.0.0
---

Content`))
	zw.Close()
	f.Close()

	r := &LocalResolver{}
	ctx := context.Background()
	skills, err := r.Resolve(ctx, zipPath, ResolveOptions{Version: "4.5.6"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Version != "4.5.6" {
		t.Errorf("expected version '4.5.6' from override, got %q", skills[0].Version)
	}
}

// TestLocalResolver_Resolve_ZipSlipPrevention tests the zip slip vulnerability
// protection path (filepath traversal via "../" in ZIP entries).
func TestLocalResolver_Resolve_ZipSlipPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "skills.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	// Create an entry with path traversal
	w, _ := zw.Create("../../evil.txt")
	w.Write([]byte("malicious content"))
	zw.Close()
	f.Close()

	r := &LocalResolver{}
	ctx := context.Background()
	_, err = r.Resolve(ctx, zipPath, ResolveOptions{})
	if err == nil {
		t.Fatal("expected error for zip slip attempt")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("expected 'illegal file path' error, got: %v", err)
	}
}

// ─── validator.go: edge cases ──────────────────────────────────────────────

// TestParseTagsFromFrontmatter_EdgeCases directly tests the parseTagsFromFrontmatter
// function for unsupported types (nil, non-slice/non-string).
func TestParseTagsFromFrontmatter_EdgeCases(t *testing.T) {
	// nil input
	result := parseTagsFromFrontmatter(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	// integer input (unsupported type)
	result = parseTagsFromFrontmatter(42)
	if result != nil {
		t.Errorf("expected nil for int input, got %v", result)
	}

	// []string input (the type assertion branch)
	result = parseTagsFromFrontmatter([]string{"a", "b"})
	if len(result) != 2 || result[0] != "a" || result[1] != "b" {
		t.Errorf("expected [a b], got %v", result)
	}
}

// TestValidateParsedSkill_EmptyTagsList tests the path where tags is an
// empty list, which should trigger the early return in validateTags.
func TestValidateParsedSkill_EmptyTagsList(t *testing.T) {
	skill := &storage.ParsedSkill{
		Frontmatter: map[string]any{
			"tags": []any{},
		},
		Name:        "test",
		Description: "test skill",
		Version:     "1.0.0",
		Body:        "Content",
	}
	result := ValidateParsedSkill(skill)
	for _, w := range result.Warnings {
		if w.Field == "tags" {
			t.Errorf("unexpected tag warning for empty tags list: %s", w.Message)
		}
	}
}

// TestValidateParsedSkill_BodyOnlyWhitespace tests the warning for
// body that is only whitespace (TrimSpace check).
func TestValidateParsedSkill_BodyOnlyWhitespace(t *testing.T) {
	skill := &storage.ParsedSkill{
		Name:        "test",
		Description: "test skill",
		Version:     "1.0.0",
		Body:        "   \n  \t  ",
	}
	result := ValidateParsedSkill(skill)
	found := false
	for _, w := range result.Warnings {
		if w.Field == "body" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for whitespace-only body")
	}
}

// TestValidateParsedSkill_NilFrontmatterTags tests that if the
// frontmatter has no "tags" key, no tag validation is attempted.
func TestValidateParsedSkill_NilFrontmatterTags(t *testing.T) {
	skill := &storage.ParsedSkill{
		Name:        "test",
		Description: "test skill",
		Version:     "1.0.0",
		Body:        "Content",
	}
	result := ValidateParsedSkill(skill)
	if !result.Valid {
		t.Errorf("expected valid, got errors=%v", result.Errors)
	}
}

// ─── source.go: NewResolver edge cases ────────────────────────────────────
// (Already covered by TestNewResolver_UnsupportedSource and
//  TestNewResolver_FindsCorrectResolver in source_test.go)