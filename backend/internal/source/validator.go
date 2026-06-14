package source

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
)

var (
	// semverRegex validates semantic versioning format: major.minor.patch (with optional pre-release)
	semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	// nameRegex validates skill names: alphanumeric, hyphens, underscores, 1-100 chars
	nameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,99}$`)
	// tagRegex validates individual tags: alphanumeric, hyphens, underscores, 1-50 chars
	tagRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,49}$`)
)

// ValidationResult holds the outcome of a SKILL.md validation.
type ValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Warnings []ValidationError
}

// ValidationError represents a single validation issue.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Severity string `json:"severity"` // "error" or "warning"
}

// ValidateSkillFile validates a SKILL.md file at the given path.
func ValidateSkillFile(path string) (*ValidationResult, error) {
	parsed, err := storage.ParseSkillFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse skill file: %w", err)
	}
	return ValidateParsedSkill(parsed), nil
}

// ValidateSkillContent validates raw SKILL.md content.
func ValidateSkillContent(content string) (*ValidationResult, error) {
	parsed, err := storage.ParseSkillContent(content)
	if err != nil {
		return nil, fmt.Errorf("parse skill content: %w", err)
	}
	return ValidateParsedSkill(parsed), nil
}

// ValidateParsedSkill validates an already-parsed skill and returns structured results.
func ValidateParsedSkill(skill *storage.ParsedSkill) *ValidationResult {
	result := &ValidationResult{
		Valid:     true,
		Errors:    nil,
		Warnings:  nil,
	}

	if skill == nil {
		result.addError("_root", "Skill is nil")
		result.Valid = false
		return result
	}

	// Check required fields (using existing storage.ValidateSkill)
	if err := storage.ValidateSkill(skill); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "name is required"):
			result.addError("name", "Skill name is required")
		case strings.Contains(msg, "description is required"):
			result.addError("description", "Skill description is required")
		case strings.Contains(msg, "version is required"):
			result.addError("version", "Skill version is required")
		}
	}

	// Validate name format
	if skill.Name != "" && !nameRegex.MatchString(skill.Name) {
		result.addError("name", "Skill name must start with alphanumeric, contain only alphanumeric/hyphens/underscores, and be 1-100 characters")
	}

	// Validate version format (must be valid semver or "latest")
	if skill.Version != "" && skill.Version != "latest" && !semverRegex.MatchString(skill.Version) {
		result.addWarning("version", "Version %q is not valid semver; expected format: major.minor.patch (e.g., 1.0.0)", skill.Version)
	}

	// Validate tags
	if tagsRaw, ok := skill.Frontmatter["tags"]; ok {
		validateTags(skill, tagsRaw, result)
	}

	// Check for recommended fields
	if _, ok := skill.Frontmatter["author"]; !ok {
		result.addWarning("author", "Recommended field 'author' is missing")
	}

	// Validate body is not empty
	if strings.TrimSpace(skill.Body) == "" {
		result.addWarning("body", "Skill body (markdown content) is empty; consider adding documentation")
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// validateTags checks tags format and uniqueness.
func validateTags(skill *storage.ParsedSkill, tagsRaw any, result *ValidationResult) {
	tags := parseTagsFromFrontmatter(tagsRaw)
	if len(tags) == 0 {
		return
	}

	seen := make(map[string]bool)
	for _, tag := range tags {
		if !tagRegex.MatchString(tag) {
			result.addWarning("tags", "Tag %q contains invalid characters; use alphanumeric, hyphens, or underscores", tag)
		}
		if seen[strings.ToLower(tag)] {
			result.addWarning("tags", "Duplicate tag %q (case-insensitive)", tag)
		}
		seen[strings.ToLower(tag)] = true
	}
}

// parseTagsFromFrontmatter extracts tag strings from frontmatter tags field.
func parseTagsFromFrontmatter(tagsRaw any) []string {
	switch v := tagsRaw.(type) {
	case []any:
		tags := make([]string, 0, len(v))
		for _, t := range v {
			tags = append(tags, fmt.Sprintf("%v", t))
		}
		return tags
	case []string:
		return v
	case string:
		parts := strings.Split(v, ",")
		tags := make([]string, 0, len(parts))
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				tags = append(tags, t)
			}
		}
		return tags
	default:
		return nil
	}
}

func (r *ValidationResult) addError(field, format string, args ...any) {
	r.Errors = append(r.Errors, ValidationError{
		Field:    field,
		Message:  fmt.Sprintf(format, args...),
		Severity: "error",
	})
}

func (r *ValidationResult) addWarning(field, format string, args ...any) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:    field,
		Message:  fmt.Sprintf(format, args...),
		Severity: "warning",
	})
}