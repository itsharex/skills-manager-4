package operations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func TestCollectStats_Empty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stats := CollectStats(nil, nil, root)

	if stats == nil {
		t.Fatal("expected non-nil SkillStats")
	}
	if stats.TotalSkills != 0 {
		t.Errorf("expected TotalSkills 0, got %d", stats.TotalSkills)
	}
	if stats.TotalVersions != 0 {
		t.Errorf("expected TotalVersions 0, got %d", stats.TotalVersions)
	}
	if stats.TotalNamespaces != 0 {
		t.Errorf("expected TotalNamespaces 0, got %d", stats.TotalNamespaces)
	}
	if stats.TotalAgents != 0 {
		t.Errorf("expected TotalAgents 0, got %d", stats.TotalAgents)
	}
	if stats.InstalledSkills != 0 {
		t.Errorf("expected InstalledSkills 0, got %d", stats.InstalledSkills)
	}
	if stats.DiskUsageBytes != 0 {
		t.Errorf("expected DiskUsageBytes 0, got %d", stats.DiskUsageBytes)
	}
}

func TestCollectStats_WithData(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a file in the skills dir to contribute to disk usage
	skillFile := filepath.Join(skillsDir, "myskill@1.0.0", "skill.md")
	if err := os.MkdirAll(filepath.Dir(skillFile), 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("hello world")
	if err := os.WriteFile(skillFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"myskill": {
				Name:      "myskill",
				Namespace: "testns",
				Versions:  []string{"1.0.0", "2.0.0"},
				Latest:    "2.0.0",
			},
			"otherskill": {
				Name:      "otherskill",
				Namespace: "testns",
				Versions:  []string{"1.0.0"},
				Latest:    "1.0.0",
			},
		},
	}

	lock := &models.LockFile{
		Version: 1,
		Skills: map[string]models.LockEntry{
			"myskill": {
				SkillID: models.SkillID{
					Namespace: "testns",
					Name:      "myskill",
					Version:   "2.0.0",
				},
				Agents: []models.LockAgentBinding{
					{AgentID: "claude", Path: "/some/path", Mode: "symlink"},
					{AgentID: "cursor", Path: "/some/path", Mode: "symlink"},
				},
			},
			"otherskill": {
				SkillID: models.SkillID{
					Namespace: "testns",
					Name:      "otherskill",
					Version:   "1.0.0",
				},
				Agents: []models.LockAgentBinding{
					{AgentID: "claude", Path: "/some/path", Mode: "symlink"},
				},
			},
		},
	}

	stats := CollectStats(index, lock, root)

	if stats.TotalSkills != 2 {
		t.Errorf("expected TotalSkills 2, got %d", stats.TotalSkills)
	}
	if stats.TotalVersions != 3 {
		t.Errorf("expected TotalVersions 3, got %d", stats.TotalVersions)
	}
	if stats.TotalNamespaces != 1 {
		t.Errorf("expected TotalNamespaces 1, got %d", stats.TotalNamespaces)
	}
	if stats.TotalAgents != 2 {
		t.Errorf("expected TotalAgents 2, got %d", stats.TotalAgents)
	}
	if stats.InstalledSkills != 2 {
		t.Errorf("expected InstalledSkills 2, got %d", stats.InstalledSkills)
	}
	if stats.DiskUsageBytes <= 0 {
		t.Errorf("expected positive DiskUsageBytes, got %d", stats.DiskUsageBytes)
	}

	// Check SkillsPerAgent
	if len(stats.SkillsPerAgent) != 2 {
		t.Errorf("expected 2 agents in SkillsPerAgent, got %d", len(stats.SkillsPerAgent))
	}
	if stats.SkillsPerAgent["claude"] != 2 {
		t.Errorf("expected claude to have 2 skills, got %d", stats.SkillsPerAgent["claude"])
	}
	if stats.SkillsPerAgent["cursor"] != 1 {
		t.Errorf("expected cursor to have 1 skill, got %d", stats.SkillsPerAgent["cursor"])
	}

	// Check SkillsPerVersion
	if len(stats.SkillsPerVersion) != 2 {
		t.Errorf("expected 2 versions in SkillsPerVersion, got %d", len(stats.SkillsPerVersion))
	}
	if stats.SkillsPerVersion["2.0.0"] != 1 {
		t.Errorf("expected 1 skill at version 2.0.0, got %d", stats.SkillsPerVersion["2.0.0"])
	}
	if stats.SkillsPerVersion["1.0.0"] != 1 {
		t.Errorf("expected 1 skill at version 1.0.0, got %d", stats.SkillsPerVersion["1.0.0"])
	}
}

func TestCollectStats_WithData_FallbackVersion(t *testing.T) {
	t.Parallel()

	// Entry with versions but no latest; should fall back to first version
	root := t.TempDir()
	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"fallback": {
				Name:      "fallback",
				Namespace: "ns",
				Versions:  []string{"3.0.0"},
				Latest:    "",
			},
		},
	}

	stats := CollectStats(index, nil, root)
	if stats.TotalVersions != 1 {
		t.Errorf("expected TotalVersions 1, got %d", stats.TotalVersions)
	}
	if stats.SkillsPerVersion["3.0.0"] != 1 {
		t.Errorf("expected 1 skill at version 3.0.0 (fallback), got %d", stats.SkillsPerVersion["3.0.0"])
	}
}

func TestCollectStats_WithData_EmptyVersions(t *testing.T) {
	t.Parallel()

	// Entry with no versions and no latest
	root := t.TempDir()
	index := &models.Index{
		Version: 1,
		Skills: map[string]models.IndexEntry{
			"empty": {
				Name:      "empty",
				Namespace: "ns",
				Versions:  nil,
				Latest:    "",
			},
		},
	}

	stats := CollectStats(index, nil, root)
	if stats.TotalVersions != 0 {
		t.Errorf("expected TotalVersions 0, got %d", stats.TotalVersions)
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatBytes_Negative(t *testing.T) {
	t.Parallel()

	result := FormatBytes(-100)
	if !strings.Contains(result, "-") {
		t.Errorf("expected negative representation, got %q", result)
	}
}

func TestCollectAgentStats(t *testing.T) {
	t.Parallel()

	stats := CollectAgentStats()
	// Function should not panic; return value may be empty if no agent dirs exist
	_ = stats
}

func TestCollectAgentStats_WithCustomDir(t *testing.T) {
	t.Parallel()

	// Create a temp agent-style directory structure
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}

	// Just verify the function doesn't panic or error
	stats := CollectAgentStats()
	if stats == nil {
		// This is fine on systems without any agent dirs
		return
	}
	for agentName, summary := range stats {
		if agentName == "" {
			t.Error("expected non-empty agent name")
		}
		if summary == "" {
			t.Errorf("expected non-empty summary for agent %q", agentName)
		}
	}
	_ = home
}

func TestSkillStats_SummaryLine(t *testing.T) {
	t.Parallel()

	stats := &SkillStats{
		TotalSkills:     3,
		TotalVersions:   5,
		TotalNamespaces: 2,
		InstalledSkills: 2,
		TotalAgents:     1,
		DiskUsageBytes:  2048,
	}

	line := stats.SummaryLine()
	if line == "" {
		t.Fatal("expected non-empty summary line")
	}

	// Check that key values appear in the output
	expectedParts := []string{"3 skills", "5 versions", "2 namespaces", "2 installed", "1 agents", "2.0 KB"}
	for _, part := range expectedParts {
		if !strings.Contains(line, part) {
			t.Errorf("expected summary line to contain %q, got: %s", part, line)
		}
	}
}

func TestSkillStats_SummaryLine_ZeroValues(t *testing.T) {
	t.Parallel()

	stats := &SkillStats{}
	line := stats.SummaryLine()
	if line == "" {
		t.Fatal("expected non-empty summary line for zero values")
	}
	if !strings.Contains(line, "0 skills") {
		t.Errorf("expected summary line to contain zero values, got: %s", line)
	}
}

func TestCalculateDiskUsage_MissingDir(t *testing.T) {
	t.Parallel()

	size := calculateDiskUsage("/nonexistent/path")
	if size != 0 {
		t.Errorf("expected 0 for missing dir, got %d", size)
	}
}

func TestCalculateDiskUsage_WithFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create some files
	f1 := filepath.Join(dir, "a.txt")
	os.WriteFile(f1, []byte("hello"), 0o644) // 5 bytes

	subDir := filepath.Join(dir, "sub")
	os.MkdirAll(subDir, 0o755)
	f2 := filepath.Join(subDir, "b.txt")
	os.WriteFile(f2, []byte("world"), 0o644) // 5 bytes

	// Total should be >= 10 bytes (may include dir entry overhead on some systems)
	size := calculateDiskUsage(dir)
	if size < 10 {
		t.Errorf("expected at least 10 bytes, got %d", size)
	}
}