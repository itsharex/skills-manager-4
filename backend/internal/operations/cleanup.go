package operations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanupStats records the result of a cleanup operation.
type CleanupStats struct {
	BrokenSymlinksRemoved int    `json:"broken_symlinks_removed"`
	StaleTempDirsRemoved  int    `json:"stale_temp_dirs_removed"`
	OrphanedEntriesFixed  int    `json:"orphaned_entries_fixed"`
	Errors                []string `json:"errors,omitempty"`
}

// CleanupBrokenSymlinks removes broken symlinks in the skills directory.
// These can occur when a skill version is deleted but the "latest" symlink
// still points to it, or when external storage is unmounted.
func CleanupBrokenSymlinks(skillsDir string) (int, error) {
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return 0, nil
	}

	count := 0
	err = filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if d.Type()&os.ModeSymlink != 0 {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if removeErr := os.Remove(path); removeErr != nil {
					return fmt.Errorf("remove broken symlink %s: %w", path, removeErr)
				}
				count++
			}
		}
		return nil
	})

	return count, err
}

// CleanupStaleTempDirs removes temporary directories left by failed operations.
// These are typically named with a "skillsmanager-*" prefix.
func CleanupStaleTempDirs() (int, error) {
	tmpDir := os.TempDir()
	count := 0

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return 0, fmt.Errorf("read temp dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), "skillsmanager-") {
			fullPath := filepath.Join(tmpDir, entry.Name())
			if err := os.RemoveAll(fullPath); err != nil {
				return count, fmt.Errorf("remove stale temp dir %s: %w", fullPath, err)
			}
			count++
		}
	}

	return count, nil
}

// CleanupOrphanedSkills removes skill directories that are not tracked
// in the index and are not "latest" symlinks.
func CleanupOrphanedSkills(skillsDir string, trackedVersions map[string]bool) (int, error) {
	info, err := os.Stat(skillsDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return 0, nil
	}

	count := 0
	err = filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Only process directories matching name@version pattern
		name := d.Name()
		if !strings.Contains(name, "@") {
			return nil
		}
		// Skip "latest" symlinks
		if strings.HasSuffix(name, "@latest") {
			return nil
		}
		// Skip if tracked
		if trackedVersions != nil {
			rel, _ := filepath.Rel(skillsDir, path)
			if trackedVersions[rel] {
				return nil
			}
		}
		// Remove untracked skill directory
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove orphaned skill %s: %w", path, err)
		}
		count++
		return nil
	})

	return count, err
}

// RunFullCleanup executes all cleanup operations and returns a summary.
func RunFullCleanup(skillsDir string, trackedVersions map[string]bool) *CleanupStats {
	stats := &CleanupStats{}

	brokenCount, err := CleanupBrokenSymlinks(skillsDir)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("broken symlinks: %v", err))
	}
	stats.BrokenSymlinksRemoved = brokenCount

	staleCount, err := CleanupStaleTempDirs()
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("stale temp dirs: %v", err))
	}
	stats.StaleTempDirsRemoved = staleCount

	orphanedCount, err := CleanupOrphanedSkills(skillsDir, trackedVersions)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("orphaned skills: %v", err))
	}
	stats.OrphanedEntriesFixed = orphanedCount

	return stats
}