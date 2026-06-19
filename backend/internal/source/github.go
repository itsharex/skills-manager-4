package source

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

func (r *GitHubResolver) CanHandle(source string) bool {
	if strings.HasPrefix(source, "gh:") {
		return true
	}
	if strings.HasPrefix(source, "https://github.com/") || strings.HasPrefix(source, "http://github.com/") {
		return true
	}
	if strings.HasPrefix(source, "git@github.com:") {
		return true
	}
	if strings.HasPrefix(source, "github.com/") {
		return true
	}
	return false
}

func (r *GitHubResolver) Resolve(ctx context.Context, source string, opts ResolveOptions) ([]models.ResolvedSkill, error) {
	ownerRepo := parseGitHubOwnerRepo(source)
	if ownerRepo == "" {
		return nil, fmt.Errorf("invalid GitHub source: %s", source)
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "skillsmanager-github-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Clone the repository with depth 1
	repoURL := fmt.Sprintf("https://github.com/%s.git", ownerRepo)
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	// Extract repo name (part after owner/)
	parts := strings.SplitN(ownerRepo, "/", 2)
	repoName := parts[1]
	namespace := "github:" + ownerRepo

	// Scan for SKILL.md files
	skills, err := scanSkillFiles(tmpDir, namespace, repoName)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("scan skills: %w", err)
	}

	// Apply version from options if specified
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

// parseGitHubOwnerRepo extracts owner/repo from various GitHub URL formats.
// Returns the owner/repo part and any remaining subpath.
func parseGitHubOwnerRepo(source string) string {
	// Strip protocol and domain prefix to get to the path portion
	var path string
	switch {
	case strings.HasPrefix(source, "git@github.com:"):
		path = strings.TrimPrefix(source, "git@github.com:")
	case strings.HasPrefix(source, "gh:"):
		path = strings.TrimPrefix(source, "gh:")
	case strings.HasPrefix(source, "https://github.com/"):
		path = strings.TrimPrefix(source, "https://github.com/")
	case strings.HasPrefix(source, "http://github.com/"):
		path = strings.TrimPrefix(source, "http://github.com/")
	case strings.HasPrefix(source, "github.com/"):
		path = strings.TrimPrefix(source, "github.com/")
	default:
		return ""
	}

	// Strip trailing .git and /
	path = strings.TrimSuffix(path, ".git")
	path = strings.TrimSuffix(path, "/")

	// Split by / to get path segments
	segments := strings.SplitN(path, "/", 3)
	if len(segments) < 2 {
		return ""
	}

	// Only return owner/repo — ignore any extra path segments (tree, blob, subdirs, etc.)
	return segments[0] + "/" + segments[1]
}

// scanSkillFiles scans a directory for SKILL.md files and returns ResolvedSkill entries.
// If SKILL.md exists at root, it's a single-skill repo.
// Otherwise, it looks for SKILL.md in subdirectories (multi-skill repo).
func scanSkillFiles(rootDir, namespace, repoName string) ([]models.ResolvedSkill, error) {
	rootSkillPath := filepath.Join(rootDir, "SKILL.md")
	if info, err := os.Stat(rootSkillPath); err == nil && !info.IsDir() {
		// Single-skill repo
		parsed, err := storage.ParseSkillFile(rootSkillPath)
		if err != nil {
			// If parsing fails, still return a skill with repo name
			return []models.ResolvedSkill{
				{
					LocalPath: rootDir,
					Namespace: namespace,
					Name:      repoName,
					Version:   "latest",
				},
			}, nil
		}
		version := parsed.Version
		if version == "" {
			version = "latest"
		}
		return []models.ResolvedSkill{
			{
				LocalPath: rootDir,
				Namespace: namespace,
				Name:      repoName,
				Version:   version,
			},
		}, nil
	}

	// Multi-skill repo - look for SKILL.md in subdirectories
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var skills []models.ResolvedSkill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}
		skillPath := filepath.Join(rootDir, entry.Name(), "SKILL.md")
		if info, err := os.Stat(skillPath); err == nil && !info.IsDir() {
			parsed, err := storage.ParseSkillFile(skillPath)
			version := "latest"
			if err == nil && parsed.Version != "" {
				version = parsed.Version
			}
			skills = append(skills, models.ResolvedSkill{
				LocalPath: filepath.Join(rootDir, entry.Name()),
				Namespace: namespace,
				Name:      entry.Name(),
				Version:   version,
			})
		} else {
			// Check one level deeper: e.g. skills/<name>/SKILL.md
			// This handles skills.sh repos that nest skills under a skills/ directory
			subDir := filepath.Join(rootDir, entry.Name())
			subEntries, subErr := os.ReadDir(subDir)
			if subErr != nil {
				continue
			}
			for _, sub := range subEntries {
				if !sub.IsDir() {
					continue
				}
				// Skip .git directory
				if sub.Name() == ".git" {
					continue
				}
				subSkillPath := filepath.Join(subDir, sub.Name(), "SKILL.md")
				if info, err := os.Stat(subSkillPath); err == nil && !info.IsDir() {
					parsed, err := storage.ParseSkillFile(subSkillPath)
					version := "latest"
					if err == nil && parsed.Version != "" {
						version = parsed.Version
					}
					skills = append(skills, models.ResolvedSkill{
						LocalPath: filepath.Join(subDir, sub.Name()),
						Namespace: namespace,
						Name:      sub.Name(),
						Version:   version,
					})
				}
			}
		}
	}

	if len(skills) == 0 {
		return nil, fmt.Errorf("no SKILL.md files found in %s", rootDir)
	}

	return skills, nil
}