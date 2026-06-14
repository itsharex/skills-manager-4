package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestNewLockFile_CreatesNew(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")
	lf, err := NewLockFile(path)
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	if lf.data.Version != 1 {
		t.Errorf("expected version 1, got %d", lf.data.Version)
	}
	if lf.data.Skills == nil {
		t.Error("expected non-nil Skills map")
	}
	if len(lf.data.Skills) != 0 {
		t.Errorf("expected empty skills, got %d", len(lf.data.Skills))
	}

	// Verify file was created on disk
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("lock.json was not created on disk")
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var lock models.LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatalf("invalid JSON written: %v", err)
	}
}

func TestNewLockFile_LoadsExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")

	lf1, err := NewLockFile(path)
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-01-01T00:00:00Z",
		Source:      "https://github.com/test/my-skill",
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/path/to/skill", Mode: "symlink"},
		},
	}
	if err := lf1.Track(entry); err != nil {
		t.Fatalf("Track failed: %v", err)
	}
	if err := lf1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	lf2, err := NewLockFile(path)
	if err != nil {
		t.Fatalf("NewLockFile reload failed: %v", err)
	}

	entries, _ := lf2.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].SkillID.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", entries[0].SkillID.Name)
	}
}

func TestLockFile_Track_NewEntry(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-01-01T00:00:00Z",
		Source:      "https://github.com/test/my-skill",
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/home/agent1/skills", Mode: "symlink"},
		},
	}

	if err := lf.Track(entry); err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	key := "github.com/test/my-skill@1.0.0"
	got, err := lf.GetBySkill(key)
	if err != nil {
		t.Fatalf("GetBySkill failed: %v", err)
	}
	if got.SkillID.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", got.SkillID.Name)
	}
	if got.SkillID.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", got.SkillID.Version)
	}
	if len(got.Agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(got.Agents))
	}
	if got.Agents[0].AgentID != "agent-1" {
		t.Errorf("expected AgentID 'agent-1', got %q", got.Agents[0].AgentID)
	}
}

func TestLockFile_Track_MergeAgents(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	// Track first agent
	entry1 := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-01-01T00:00:00Z",
		Source:      "https://github.com/test/my-skill",
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/path1", Mode: "symlink"},
		},
	}
	lf.Track(entry1)

	// Track second agent for same skill
	entry2 := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-06-01T00:00:00Z",
		Source:      "https://github.com/test/my-skill",
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-2", Path: "/path2", Mode: "copy"},
		},
	}
	lf.Track(entry2)

	key := "github.com/test/my-skill@1.0.0"
	got, _ := lf.GetBySkill(key)
	if len(got.Agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(got.Agents))
	}
	// InstalledAt and Source should be updated from the latest Track call
	if got.InstalledAt != "2024-06-01T00:00:00Z" {
		t.Errorf("expected InstalledAt to be updated, got %q", got.InstalledAt)
	}
}

func TestLockFile_Track_DuplicateAgent(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/path1", Mode: "symlink"},
		},
	}
	lf.Track(entry)

	// Same agent again
	lf.Track(entry)

	key := "github.com/test/my-skill@1.0.0"
	got, _ := lf.GetBySkill(key)
	if len(got.Agents) != 1 {
		t.Errorf("expected 1 agent (deduplicated), got %d", len(got.Agents))
	}
}

func TestLockFile_Untrack(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/path1", Mode: "symlink"},
			{AgentID: "agent-2", Path: "/path2", Mode: "copy"},
		},
	}
	lf.Track(entry)

	key := "github.com/test/my-skill@1.0.0"

	// Untrack agent-1
	if err := lf.Untrack(key, "agent-1"); err != nil {
		t.Fatalf("Untrack failed: %v", err)
	}

	got, _ := lf.GetBySkill(key)
	if len(got.Agents) != 1 {
		t.Fatalf("expected 1 agent remaining, got %d", len(got.Agents))
	}
	if got.Agents[0].AgentID != "agent-2" {
		t.Errorf("expected remaining agent 'agent-2', got %q", got.Agents[0].AgentID)
	}

	// Untrack last agent should remove the entry entirely
	if err := lf.Untrack(key, "agent-2"); err != nil {
		t.Fatalf("Untrack failed: %v", err)
	}

	_, err = lf.GetBySkill(key)
	if err == nil {
		t.Error("expected error after removing last agent")
	}
}

func TestLockFile_Untrack_NotFound(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	err = lf.Untrack("nonexistent/skill@1.0.0", "agent-1")
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

func TestLockFile_GetBySkill_NotFound(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	_, err = lf.GetBySkill("nonexistent/skill@1.0.0")
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

func TestLockFile_GetByAgent(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	// Add multiple skills with different agents
	lf.Track(models.LockEntry{
		SkillID: models.SkillID{Namespace: "ns1", Name: "skill-a", Version: "1.0.0"},
		Agents:  []models.LockAgentBinding{{AgentID: "agent-1", Path: "/p1", Mode: "symlink"}},
	})
	lf.Track(models.LockEntry{
		SkillID: models.SkillID{Namespace: "ns1", Name: "skill-b", Version: "1.0.0"},
		Agents:  []models.LockAgentBinding{{AgentID: "agent-1", Path: "/p2", Mode: "copy"}},
	})
	lf.Track(models.LockEntry{
		SkillID: models.SkillID{Namespace: "ns2", Name: "skill-c", Version: "2.0.0"},
		Agents:  []models.LockAgentBinding{{AgentID: "agent-2", Path: "/p3", Mode: "symlink"}},
	})

	entries, err := lf.GetByAgent("agent-1")
	if err != nil {
		t.Fatalf("GetByAgent failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for agent-1, got %d", len(entries))
	}

	entries, err = lf.GetByAgent("agent-2")
	if err != nil {
		t.Fatalf("GetByAgent failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry for agent-2, got %d", len(entries))
	}

	entries, err = lf.GetByAgent("agent-nonexistent")
	if err != nil {
		t.Fatalf("GetByAgent failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for non-existent agent, got %d", len(entries))
	}
}

func TestLockFile_List(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	lf.Track(models.LockEntry{
		SkillID: models.SkillID{Namespace: "ns1", Name: "skill-a", Version: "1.0.0"},
		Agents:  []models.LockAgentBinding{{AgentID: "a1", Path: "/p1", Mode: "symlink"}},
	})
	lf.Track(models.LockEntry{
		SkillID: models.SkillID{Namespace: "ns1", Name: "skill-b", Version: "1.0.0"},
		Agents:  []models.LockAgentBinding{{AgentID: "a2", Path: "/p2", Mode: "copy"}},
	})

	entries, err := lf.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLockFile_List_Empty(t *testing.T) {
	lf, err := NewLockFile(filepath.Join(t.TempDir(), "lock.json"))
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entries, err := lf.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d", len(entries))
	}
}

func TestLockFile_SaveAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")
	lf, err := NewLockFile(path)
	if err != nil {
		t.Fatalf("NewLockFile failed: %v", err)
	}

	entry := models.LockEntry{
		SkillID: models.SkillID{
			Namespace: "github.com/test",
			Name:      "my-skill",
			Version:   "1.0.0",
		},
		InstalledAt: "2024-01-01T00:00:00Z",
		Source:      "https://github.com/test/my-skill",
		Agents: []models.LockAgentBinding{
			{AgentID: "agent-1", Path: "/home/agent1", Mode: "symlink"},
		},
	}
	lf.Track(entry)
	lf.Save()

	// Reload from disk
	if err := lf.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	key := "github.com/test/my-skill@1.0.0"
	got, err := lf.GetBySkill(key)
	if err != nil {
		t.Fatalf("GetBySkill after Reload failed: %v", err)
	}
	if got.SkillID.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", got.SkillID.Name)
	}
}

func TestLockFile_Reload_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := NewLockFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLockFile_ThreadSafety(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")
	lf, err := NewLockFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			entry := models.LockEntry{
				SkillID: models.SkillID{
					Namespace: "test",
					Name:      fmt.Sprintf("skill-%d", n%3),
					Version:   "1.0.0",
				},
				InstalledAt: "now",
				Source:      "test",
				Agents: []models.LockAgentBinding{
					{AgentID: fmt.Sprintf("agent-%d", n), Path: "/tmp", Mode: "symlink"},
				},
			}
			_ = lf.Track(entry)
		}(i)
	}
	wg.Wait()

	entries, err := lf.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 unique skills, got %d", len(entries))
	}

	// Concurrent read/write
	var wg2 sync.WaitGroup
	for range 5 {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			_, _ = lf.List()
			_, _ = lf.GetByAgent("agent-0")
		}()
	}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		_ = lf.Track(models.LockEntry{
			SkillID: models.SkillID{Namespace: "concurrent", Name: "test", Version: "1.0.0"},
			Agents: []models.LockAgentBinding{
				{AgentID: "concurrent-agent", Path: "/tmp", Mode: "symlink"},
			},
		})
	}()
	wg2.Wait()

	// Verify no data corruption
	entry, err := lf.GetBySkill("concurrent/test@1.0.0")
	if err != nil {
		t.Fatalf("concurrent entry should exist: %v", err)
	}
	if len(entry.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(entry.Agents))
	}
}