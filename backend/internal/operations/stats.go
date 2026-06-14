package operations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// SkillStats holds aggregated statistics about the skill repository.
type SkillStats struct {
	TotalSkills      int            `json:"total_skills"`
	TotalVersions    int            `json:"total_versions"`
	TotalNamespaces  int            `json:"total_namespaces"`
	TotalAgents      int            `json:"total_agents"`
	InstalledSkills  int            `json:"installed_skills"`
	DiskUsageBytes   int64          `json:"disk_usage_bytes"`
	SkillsPerAgent   map[string]int `json:"skills_per_agent,omitempty"`
	SkillsPerVersion map[string]int `json:"skills_per_version,omitempty"`
}

// CollectStats gathers statistics from the index, lock file, and filesystem.
func CollectStats(index *models.Index, lock *models.LockFile, root string) *SkillStats {
	stats := &SkillStats{
		SkillsPerAgent:   make(map[string]int),
		SkillsPerVersion: make(map[string]int),
	}

	if index != nil {
		stats.TotalSkills = len(index.Skills)
		namespaces := make(map[string]bool)
		for _, entry := range index.Skills {
			namespaces[entry.Namespace] = true
			stats.TotalVersions += len(entry.Versions)
			version := entry.Latest
			if version == "" && len(entry.Versions) > 0 {
				version = entry.Versions[0]
			}
			if version != "" {
				stats.SkillsPerVersion[version]++
			}
		}
		stats.TotalNamespaces = len(namespaces)
	}

	if lock != nil {
		stats.InstalledSkills = len(lock.Skills)
		for _, entry := range lock.Skills {
			for _, agent := range entry.Agents {
				stats.SkillsPerAgent[agent.AgentID]++
			}
		}
		stats.TotalAgents = len(stats.SkillsPerAgent)
	}

	// Calculate disk usage
	stats.DiskUsageBytes = calculateDiskUsage(filepath.Join(root, "skills"))

	return stats
}

// calculateDiskUsage recursively calculates directory size in bytes.
func calculateDiskUsage(dir string) int64 {
	var size int64
	info, err := os.Stat(dir)
	if os.IsNotExist(err) || !info.IsDir() {
		return 0
	}

	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				size += fi.Size()
			}
		}
		return nil
	})

	return size
}

// FormatBytes converts bytes to a human-readable string.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CollectAgentStats gathers per-agent statistics.
func CollectAgentStats() map[string]string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	agents := map[string]string{
		"claude":        filepath.Join(home, ".claude", "skills"),
		"cursor":        filepath.Join(home, ".cursor", "skills"),
		"windsurf":      filepath.Join(home, ".windsurf", "skills"),
		"github-copilot": filepath.Join(home, ".github-copilot", "skills"),
	}

	stats := make(map[string]string)
	for name, dir := range agents {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) || !info.IsDir() {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		skillCount := 0
		for _, entry := range entries {
			if strings.Contains(entry.Name(), "@") || strings.HasSuffix(entry.Name(), ".md") {
				skillCount++
			}
		}
		stats[name] = fmt.Sprintf("%d skills", skillCount)
	}

	return stats
}

// SummaryLine returns a one-line summary of the stats.
func (s *SkillStats) SummaryLine() string {
	return fmt.Sprintf(
		"%d skills (%d versions) across %d namespaces | %d installed to %d agents | disk: %s",
		s.TotalSkills, s.TotalVersions, s.TotalNamespaces,
		s.InstalledSkills, s.TotalAgents, FormatBytes(s.DiskUsageBytes),
	)
}