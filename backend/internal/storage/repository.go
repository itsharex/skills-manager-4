package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Repository manages the local skill pool.
type Repository struct {
	Root  string
	Paths models.PoolPaths
}

// NewRepository creates a Repository instance backed by the given pool path.
func NewRepository(poolPath string) *Repository {
	return &Repository{
		Root: poolPath,
		Paths: models.PoolPaths{
			PoolPath:   poolPath,
			Root:       poolPath,
			SkillsDir:  poolPath,
			MetaDir:    filepath.Join(poolPath, ".meta"),
			IndexPath:  filepath.Join(poolPath, ".meta", "index.json"),
			LockPath:   filepath.Join(poolPath, ".meta", "lock.json"),
			ConfigPath: filepath.Join(poolPath, ".meta", "config.json"),
		},
	}
}

// SkillPath returns Root/{name}/ (flat structure, ignores namespace and version).
func (r *Repository) SkillPath(namespace, name, version string) string {
	return filepath.Join(r.Root, name)
}

// Store copies a resolved skill into the pool.
func (r *Repository) Store(skill models.ResolvedSkill, namespace, version string) (string, error) {
	dest := r.SkillPath("", skill.Name, "")

	// Create destination directory
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", fmt.Errorf("create skill directory: %w", err)
	}

	// Copy from source
	srcInfo, err := os.Stat(skill.LocalPath)
	if err != nil {
		return "", fmt.Errorf("stat source: %w", err)
	}

	if srcInfo.IsDir() {
		if err := copyDir(skill.LocalPath, dest); err != nil {
			return "", fmt.Errorf("copy skill directory: %w", err)
		}
	} else {
		if err := copyFile(skill.LocalPath, filepath.Join(dest, filepath.Base(skill.LocalPath))); err != nil {
			return "", fmt.Errorf("copy skill file: %w", err)
		}
	}

	return dest, nil
}

// Remove deletes a skill from the pool.
func (r *Repository) Remove(namespace, name, version string) error {
	path := r.SkillPath("", name, "")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // already gone
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove skill %s: %w", name, err)
	}
	return nil
}

// UpdateLatest is a no-op in the flat pool layout (no latest symlink needed).
func (r *Repository) UpdateLatest(namespace, name, version string) error {
	log.Printf("[WARNING] UpdateLatest is a no-op in flat pool layout; namespace=%s name=%s version=%s", namespace, name, version)
	return nil
}

// ListSkills returns all skills stored in the pool (flat layout).
func (r *Repository) ListSkills() ([]string, error) {
	skillsDir := r.Paths.SkillsDir
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var skills []string
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip .meta directory
		if entry.Name() == ".meta" {
			continue
		}
		// Check for SKILL.md
		skillMDPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMDPath); err != nil {
			continue
		}
		skills = append(skills, entry.Name())
	}

	return skills, nil
}

// Exists checks if a specific skill exists in the pool.
func (r *Repository) Exists(namespace, name, version string) bool {
	path := r.SkillPath("", name, "")
	_, err := os.Stat(path)
	return err == nil
}

// --- helpers ---

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

// sanitizeName sanitizes a path component for safe filesystem use.
func sanitizeName(s string) string {
	replacer := strings.NewReplacer(
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "-",
	)
	return replacer.Replace(s)
}