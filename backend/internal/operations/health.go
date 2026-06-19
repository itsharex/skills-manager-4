package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// HealthCheckResult represents the outcome of a single diagnostic check.
type HealthCheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message,omitempty"`
}

// HealthReport aggregates all diagnostic checks.
type HealthReport struct {
	PoolPath string             `json:"pool_path"`
	Checks   []HealthCheckResult `json:"checks"`
	AllPass  bool               `json:"all_pass"`
}

// RunDoctor runs all diagnostic checks and returns a health report.
func RunDoctor(poolPath string) *HealthReport {
	report := &HealthReport{
		PoolPath: poolPath,
		Checks:   make([]HealthCheckResult, 0),
		AllPass:  true,
	}

	paths := GetPoolPaths(poolPath)

	report.Checks = append(report.Checks, checkPoolRoot(poolPath))
	report.Checks = append(report.Checks, checkSkillsDir(paths.SkillsDir))
	report.Checks = append(report.Checks, checkMetaDir(paths.MetaDir))
	report.Checks = append(report.Checks, checkFileExists("index.json", paths.IndexPath))
	report.Checks = append(report.Checks, checkFileExists("lock.json", paths.LockPath))
	report.Checks = append(report.Checks, checkFileExists("config.json", paths.ConfigPath))
	report.Checks = append(report.Checks, checkConfigValid(paths.ConfigPath))
	report.Checks = append(report.Checks, checkBrokenSymlinks(paths.SkillsDir))
	report.Checks = append(report.Checks, checkDiskSpace(poolPath))

	for _, c := range report.Checks {
		if c.Status == "fail" {
			report.AllPass = false
			break
		}
	}

	return report
}

// checkPoolRoot validates that the pool root directory exists.
func checkPoolRoot(root string) HealthCheckResult {
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return HealthCheckResult{
			Name: "pool_root", Status: "fail",
			Message: fmt.Sprintf("Pool root does not exist: %s", root),
		}
	}
	if err != nil {
		return HealthCheckResult{
			Name: "pool_root", Status: "fail",
			Message: fmt.Sprintf("Cannot access pool root: %v", err),
		}
	}
	if !info.IsDir() {
		return HealthCheckResult{
			Name: "pool_root", Status: "fail",
			Message: fmt.Sprintf("Pool root is not a directory: %s", root),
		}
	}
	return HealthCheckResult{Name: "pool_root", Status: "pass", Message: "OK"}
}

// checkSkillsDir validates the skills storage directory.
func checkSkillsDir(skillsDir string) HealthCheckResult {
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) {
		return HealthCheckResult{
			Name: "skills_dir", Status: "warn",
			Message: fmt.Sprintf("Skills directory does not exist (no skills installed yet): %s", skillsDir),
		}
	}
	if err != nil {
		return HealthCheckResult{
			Name: "skills_dir", Status: "fail",
			Message: fmt.Sprintf("Cannot access skills directory: %v", err),
		}
	}
	if !info.IsDir() {
		return HealthCheckResult{
			Name: "skills_dir", Status: "fail",
			Message: fmt.Sprintf("Skills path is not a directory: %s", skillsDir),
		}
	}
	return HealthCheckResult{Name: "skills_dir", Status: "pass"}
}

// checkMetaDir validates that the .meta directory exists.
func checkMetaDir(metaDir string) HealthCheckResult {
	info, err := os.Stat(metaDir)
	if os.IsNotExist(err) {
		return HealthCheckResult{
			Name: "meta_dir", Status: "warn",
			Message: fmt.Sprintf("Meta directory does not exist (will be created on first write): %s", metaDir),
		}
	}
	if err != nil {
		return HealthCheckResult{
			Name: "meta_dir", Status: "fail",
			Message: fmt.Sprintf("Cannot access meta directory: %v", err),
		}
	}
	if !info.IsDir() {
		return HealthCheckResult{
			Name: "meta_dir", Status: "fail",
			Message: fmt.Sprintf("Meta path is not a directory: %s", metaDir),
		}
	}
	return HealthCheckResult{Name: "meta_dir", Status: "pass"}
}

// checkFileExists checks that a specific file exists.
func checkFileExists(name, path string) HealthCheckResult {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return HealthCheckResult{
			Name: name, Status: "fail",
			Message: fmt.Sprintf("File not found: %s", path),
		}
	}
	if err != nil {
		return HealthCheckResult{
			Name: name, Status: "fail",
			Message: fmt.Sprintf("Cannot access %s: %v", path, err),
		}
	}
	return HealthCheckResult{Name: name, Status: "pass"}
}

// checkConfigValid validates the config file can be loaded.
func checkConfigValid(configPath string) HealthCheckResult {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return HealthCheckResult{
			Name: "config_valid", Status: "fail",
			Message: fmt.Sprintf("Config load failed: %v", err),
		}
	}
	if cfg.PoolPath == "" {
		return HealthCheckResult{
			Name: "config_valid", Status: "warn",
			Message: "Pool path is empty in config",
		}
	}
	return HealthCheckResult{Name: "config_valid", Status: "pass"}
}

// checkBrokenSymlinks scans the skills directory for broken symlinks.
func checkBrokenSymlinks(skillsDir string) HealthCheckResult {
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return HealthCheckResult{Name: "broken_symlinks", Status: "pass"}
	}

	count := 0
	err = filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if d.Type()&os.ModeSymlink != 0 {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				count++
			}
		}
		return nil
	})
	if err != nil {
		return HealthCheckResult{
			Name: "broken_symlinks", Status: "warn",
			Message: fmt.Sprintf("Error scanning symlinks: %v", err),
		}
	}
	if count > 0 {
		return HealthCheckResult{
			Name: "broken_symlinks", Status: "warn",
			Message: fmt.Sprintf("Found %d broken symlink(s)", count),
		}
	}
	return HealthCheckResult{Name: "broken_symlinks", Status: "pass"}
}

// checkDiskSpace checks available disk space on the repo partition.
func checkDiskSpace(root string) HealthCheckResult {
	// Simple check: can we write a temp file?
	tmpFile := filepath.Join(root, ".health-check-tmp")
	if err := os.WriteFile(tmpFile, []byte("ok"), 0o644); err != nil {
		return HealthCheckResult{
			Name: "disk_space", Status: "fail",
			Message: fmt.Sprintf("Cannot write to pool directory: %v", err),
		}
	}
	os.Remove(tmpFile)
	return HealthCheckResult{Name: "disk_space", Status: "pass"}
}

// CheckAgentAccess verifies that an agent's skills directory is accessible.
func CheckAgentAccess(agentID, skillsDir string) HealthCheckResult {
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) {
		return HealthCheckResult{
			Name: fmt.Sprintf("agent_%s", agentID), Status: "warn",
			Message: fmt.Sprintf("Agent %q skills directory does not exist: %s", agentID, skillsDir),
		}
	}
	if err != nil {
		return HealthCheckResult{
			Name: fmt.Sprintf("agent_%s", agentID), Status: "fail",
			Message: fmt.Sprintf("Cannot access agent %q: %v", agentID, err),
		}
	}
	if !info.IsDir() {
		return HealthCheckResult{
			Name: fmt.Sprintf("agent_%s", agentID), Status: "fail",
			Message: fmt.Sprintf("Agent %q path is not a directory", agentID),
		}
	}
	return HealthCheckResult{
		Name: fmt.Sprintf("agent_%s", agentID), Status: "pass",
	}
}

// ValidateSkillIndex checks the integrity of the skill index.
func ValidateSkillIndex(index *models.Index) []HealthCheckResult {
	var results []HealthCheckResult

	if index == nil {
		results = append(results, HealthCheckResult{
			Name: "index_integrity", Status: "fail",
			Message: "Index is nil",
		})
		return results
	}

	if index.Version <= 0 {
		results = append(results, HealthCheckResult{
			Name: "index_version", Status: "warn",
			Message: fmt.Sprintf("Invalid index version: %d", index.Version),
		})
	}

	for key, entry := range index.Skills {
		if entry.Name == "" {
			results = append(results, HealthCheckResult{
				Name: fmt.Sprintf("index_entry_%s", key), Status: "warn",
				Message: fmt.Sprintf("Entry %q has empty name", key),
			})
		}
		if entry.Latest == "" && len(entry.Versions) > 0 {
			results = append(results, HealthCheckResult{
				Name: fmt.Sprintf("index_entry_%s", key), Status: "warn",
				Message: fmt.Sprintf("Entry %q has versions but no latest", key),
			})
		}
	}

	return results
}