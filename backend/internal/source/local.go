package source

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func (r *LocalResolver) CanHandle(source string) bool {
	// Handle absolute paths
	if strings.HasPrefix(source, "/") {
		return true
	}
	// Handle home directory paths
	if strings.HasPrefix(source, "~") {
		return true
	}
	// Handle relative paths
	if strings.HasPrefix(source, ".") {
		return true
	}
	return false
}

func (r *LocalResolver) Resolve(ctx context.Context, source string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	// Expand home directory
	if strings.HasPrefix(source, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("expand home dir: %w", err)
		}
		source = filepath.Join(home, source[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(source)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path: %w", err)
	}

	// Check if it's a ZIP file
	if isZipFile(absPath) {
		return resolveZipSource(absPath, opts)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("access path %s: %w", absPath, err)
	}

	if info.IsDir() {
		return resolveDirectorySource(absPath, opts)
	}

	// Single file — must be SKILL.md
	if filepath.Base(absPath) == "SKILL.md" {
		return resolveSingleSkillFile(absPath, opts)
	}

	return nil, fmt.Errorf("unsupported local source: %s", source)
}

// isZipFile checks if the given path is a ZIP file based on extension.
func isZipFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip"
}

// resolveZipSource extracts a ZIP file to a temp directory and scans for skills.
func resolveZipSource(zipPath string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "skillsmanager-zip-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	// Extract all files
	for _, f := range reader.File {
		// Prevent zip slip vulnerability
		fpath := filepath.Join(tmpDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(tmpDir)+string(os.PathSeparator)) {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("create dir: %w", err)
		}

		// Write file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("create file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("open zip entry: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("extract file: %w", err)
		}
	}

	// Scan extracted directory for skills
	skills, err := scanSkillFiles(tmpDir, "local", filepath.Base(tmpDir))
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// Set cleanup and apply version
	for i := range skills {
		if opts.Version != "" {
			skills[i].Version = opts.Version
		}
		skills[i].Cleanup = func() {
			os.RemoveAll(tmpDir)
		}
	}

	return skills, nil
}

// resolveDirectorySource scans a directory for SKILL.md files.
func resolveDirectorySource(dirPath string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	dirName := filepath.Base(dirPath)

	skills, err := scanSkillFiles(dirPath, "local", dirName)
	if err != nil {
		return nil, err
	}

	// Apply version from options if specified
	for i := range skills {
		if opts.Version != "" {
			skills[i].Version = opts.Version
		}
		// Local directory skills don't need cleanup
		skills[i].Cleanup = func() {}
	}

	return skills, nil
}

// resolveSingleSkillFile parses a single SKILL.md file.
func resolveSingleSkillFile(filePath string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	parsed, err := storage.ParseSkillFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse skill file: %w", err)
	}

	version := parsed.Version
	if version == "" {
		version = "latest"
	}
	if opts.Version != "" {
		version = opts.Version
	}

	skillDir := filepath.Dir(filePath)
	dirName := filepath.Base(skillDir)

	return []models.ResolvedSkill{
		{
			LocalPath: skillDir,
			Namespace: "local",
			Name:      dirName,
			Version:   version,
			Cleanup:   func() {},
		},
	}, nil
}