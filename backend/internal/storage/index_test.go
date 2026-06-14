package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestNewIndex_CreatesNew(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	idx, err := NewIndex(path)
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	if idx.data.Version != 1 {
		t.Errorf("expected version 1, got %d", idx.data.Version)
	}
	if idx.data.Skills == nil {
		t.Error("expected non-nil Skills map")
	}
	if len(idx.data.Skills) != 0 {
		t.Errorf("expected empty skills, got %d", len(idx.data.Skills))
	}

	// Verify file was created on disk
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("index.json was not created on disk")
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var index models.Index
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("invalid JSON written: %v", err)
	}
}

func TestNewIndex_LoadsExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")

	// Create initial index
	idx1, err := NewIndex(path)
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	// Add a skill
	entry := models.IndexEntry{
		Name:      "test-skill",
		Namespace: "github.com/test",
		Versions:  []string{"1.0.0"},
		Latest:    "1.0.0",
		Source:    "https://github.com/test/test-skill",
		SourceType: "github",
	}
	if err := idx1.Add(entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := idx1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load from existing file
	idx2, err := NewIndex(path)
	if err != nil {
		t.Fatalf("NewIndex reload failed: %v", err)
	}

	entries, _ := idx2.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got %q", entries[0].Name)
	}
}

func TestIndex_Add_NewEntry(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entry := models.IndexEntry{
		Name:        "my-skill",
		Namespace:   "github.com/user",
		Versions:    []string{"1.0.0"},
		Latest:      "1.0.0",
		Source:      "https://github.com/user/my-skill",
		SourceType:  "github",
		Description: "A test skill",
		Tags:        []string{"test", "example"},
	}

	if err := idx.Add(entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	key := "github.com/user/my-skill"
	got, err := idx.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", got.Name)
	}
	if got.Description != "A test skill" {
		t.Errorf("expected Description 'A test skill', got %q", got.Description)
	}
	if len(got.Versions) != 1 || got.Versions[0] != "1.0.0" {
		t.Errorf("expected versions [1.0.0], got %v", got.Versions)
	}
	if idx.data.LastUpdate == "" {
		t.Error("LastUpdate should not be empty after Add")
	}
}

func TestIndex_Add_MergeVersions(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	// Add initial entry
	entry1 := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"1.0.0"},
		Latest:    "1.0.0",
	}
	if err := idx.Add(entry1); err != nil {
		t.Fatalf("Add initial failed: %v", err)
	}

	// Add with new version
	entry2 := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"2.0.0"},
		Latest:    "2.0.0",
		Tags:      []string{"updated"},
	}
	if err := idx.Add(entry2); err != nil {
		t.Fatalf("Add update failed: %v", err)
	}

	key := "github.com/user/my-skill"
	got, _ := idx.Get(key)
	if len(got.Versions) != 2 {
		t.Errorf("expected 2 versions, got %v", got.Versions)
	}
	if got.Latest != "2.0.0" {
		t.Errorf("expected Latest '2.0.0', got %q", got.Latest)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "updated" {
		t.Errorf("expected tags to be overwritten, got %v", got.Tags)
	}
}

func TestIndex_Add_DuplicateVersion(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entry := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"1.0.0", "1.0.0"},
		Latest:    "1.0.0",
	}
	if err := idx.Add(entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Add same version again
	if err := idx.Add(entry); err != nil {
		t.Fatalf("Add duplicate failed: %v", err)
	}

	key := "github.com/user/my-skill"
	got, _ := idx.Get(key)
	if len(got.Versions) != 1 {
		t.Errorf("expected 1 version (deduplicated), got %v", got.Versions)
	}
}

func TestIndex_Remove(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entry := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"1.0.0"},
		Latest:    "1.0.0",
	}
	idx.Add(entry)

	key := "github.com/user/my-skill"
	if err := idx.Remove(key); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := idx.Get(key); err == nil {
		t.Error("expected error after Remove, got nil")
	}

	entries, _ := idx.List()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after Remove, got %d", len(entries))
	}
}

func TestIndex_Remove_NonExistent(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	// Removing non-existent should not error
	if err := idx.Remove("nonexistent/skill"); err != nil {
		t.Errorf("Remove non-existent should not error: %v", err)
	}
}

func TestIndex_Get_NotFound(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	_, err = idx.Get("nonexistent/skill")
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

func TestIndex_List(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entries := []models.IndexEntry{
		{Name: "skill-a", Namespace: "ns1", Versions: []string{"1.0.0"}, Latest: "1.0.0"},
		{Name: "skill-b", Namespace: "ns1", Versions: []string{"1.0.0"}, Latest: "1.0.0"},
		{Name: "skill-c", Namespace: "ns2", Versions: []string{"2.0.0"}, Latest: "2.0.0"},
	}

	for _, e := range entries {
		if err := idx.Add(e); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	list, err := idx.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 entries, got %d", len(list))
	}
}

func TestIndex_List_Empty(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	list, err := idx.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestIndex_UpdateLatest(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entry := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"1.0.0", "2.0.0"},
		Latest:    "1.0.0",
	}
	idx.Add(entry)

	key := "github.com/user/my-skill"
	if err := idx.UpdateLatest(key, "2.0.0"); err != nil {
		t.Fatalf("UpdateLatest failed: %v", err)
	}

	got, _ := idx.Get(key)
	if got.Latest != "2.0.0" {
		t.Errorf("expected Latest '2.0.0', got %q", got.Latest)
	}
}

func TestIndex_UpdateLatest_NotFound(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	err = idx.UpdateLatest("nonexistent/skill", "1.0.0")
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

func TestIndex_SaveAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	idx, err := NewIndex(path)
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	entry := models.IndexEntry{
		Name:      "my-skill",
		Namespace: "github.com/user",
		Versions:  []string{"1.0.0"},
		Latest:    "1.0.0",
	}
	idx.Add(entry)
	idx.Save()

	// Reload from disk
	if err := idx.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	key := "github.com/user/my-skill"
	got, err := idx.Get(key)
	if err != nil {
		t.Fatalf("Get after Reload failed: %v", err)
	}
	if got.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", got.Name)
	}
}

func TestIndex_Reload_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := NewIndex(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestIndex_ThreadSafety(t *testing.T) {
	idx, err := NewIndex(filepath.Join(t.TempDir(), "index.json"))
	if err != nil {
		t.Fatalf("NewIndex failed: %v", err)
	}

	done := make(chan bool)
	n := 20

	// Concurrent writes
	for i := 0; i < n; i++ {
		go func(i int) {
			entry := models.IndexEntry{
				Name:      "skill",
				Namespace: "ns",
				Versions:  []string{"1.0.0"},
				Latest:    "1.0.0",
			}
			idx.Add(entry)
			done <- true
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	// Should have only 1 entry (same key merged), but no crash
	entries, _ := idx.List()
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}