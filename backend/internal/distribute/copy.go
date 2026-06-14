package distribute

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopySkill copies skill files from sourceDir to destDir.
// This is used as a fallback when symlinks are not available.
func CopySkill(sourceDir, destDir string) error {
	srcInfo, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDir(sourceDir, destDir)
	}

	// Single file
	destFile := filepath.Join(destDir, filepath.Base(sourceDir))
	return copyFile(sourceDir, destFile)
}

// RemoveCopiedSkill removes skill files that were installed via copy mode.
func RemoveCopiedSkill(destDir string) error {
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return nil // already gone
	}
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("remove copied skill: %w", err)
	}
	return nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}