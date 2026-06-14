package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

// Repository manages the local skill repository.
type Repository struct {
	Root  string
	Paths models.RepoPaths
}

// NewRepository creates a Repository instance backed by the given root.
func NewRepository(root string) *Repository {
	return &Repository{
		Root:  root,
		Paths: models.RepoPaths{
			Root:       root,
			SkillsDir:  filepath.Join(root, "skills"),
			IndexPath:  filepath.Join(root, "index.json"),
			LockPath:   filepath.Join(root, "lock.json"),
			ConfigPath: filepath.Join(root, "config.json"),
		},
	}
}

// SkillPath returns Root/skills/{namespace}/{name}@{version}/
func (r *Repository) SkillPath(namespace, name, version string) string {
	return filepath.Join(r.Paths.SkillsDir, namespace, fmt.Sprintf("%s@%s", name, version))
}

// Store copies a resolved skill into the repository.
func (r *Repository) Store(skill models.ResolvedSkill, namespace, version string) (string, error) {
	dest := r.SkillPath(namespace, skill.Name, version)

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

// Remove deletes a skill version from the repository.
func (r *Repository) Remove(namespace, name, version string) error {
	path := r.SkillPath(namespace, name, version)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // already gone
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove skill %s/%s@%s: %w", namespace, name, version, err)
	}
	return nil
}

// UpdateLatest creates/updates a "latest" symlink for the skill version.
// The symlink is placed at {skillsDir}/{namespace}/{name}@latest pointing to {name}@{version}.
func (r *Repository) UpdateLatest(namespace, name, version string) error {
	skillDir := r.Paths.SkillsDir
	versionDir := r.SkillPath(namespace, name, version)
	latestPath := filepath.Join(skillDir, namespace, fmt.Sprintf("%s@latest", name))

	// Ensure version dir exists
	if _, err := os.Stat(versionDir); err != nil {
		return fmt.Errorf("version directory not found: %w", err)
	}

	// Ensure the parent namespace dir exists
	nsDir := filepath.Join(skillDir, namespace)
	if err := os.MkdirAll(nsDir, 0o755); err != nil {
		return fmt.Errorf("create namespace directory: %w", err)
	}

	// Remove existing latest if present
	if _, err := os.Lstat(latestPath); err == nil {
		if err := os.Remove(latestPath); err != nil {
			return fmt.Errorf("remove existing latest symlink: %w", err)
		}
	}

	// Create symlink (relative to namespace dir)
	relTarget := fmt.Sprintf("%s@%s", name, version)
	if err := os.Symlink(relTarget, latestPath); err != nil {
		return fmt.Errorf("create latest symlink: %w", err)
	}

	return nil
}

// ListSkills returns all skills stored in the repository.
func (r *Repository) ListSkills() ([]string, error) {
	skillsDir := r.Paths.SkillsDir
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var skills []string
	err := filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("access error at %s: %w", path, err)
		}
		if !d.IsDir() {
			return nil
		}
		// A directory containing "@" is a skill version entry (name@version)
		if strings.Contains(d.Name(), "@") {
			rel, _ := filepath.Rel(skillsDir, path)
			skills = append(skills, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk skills directory: %w", err)
	}

	return skills, nil
}

// Exists checks if a specific skill version exists in the repository.
func (r *Repository) Exists(namespace, name, version string) bool {
	path := r.SkillPath(namespace, name, version)
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