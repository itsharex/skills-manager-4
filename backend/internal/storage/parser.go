package storage

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsedSkill represents the result of parsing a SKILL.md file.
type ParsedSkill struct {
	Frontmatter map[string]any // parsed YAML frontmatter fields
	Body        string         // markdown content after frontmatter
	Name        string         // from frontmatter "name"
	Description string         // from frontmatter "description"
	Version     string         // from frontmatter "version"
	Tags        []string       // from frontmatter "tags"
	Author      string         // from frontmatter "author"
}

// ParseSkillFile parses a SKILL.md file and returns structured metadata.
func ParseSkillFile(path string) (*ParsedSkill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}
	return ParseSkillContent(string(data))
}

// ParseSkillContent parses SKILL.md content from a string.
func ParseSkillContent(content string) (*ParsedSkill, error) {
	content = strings.TrimSpace(content)

	// Must start with ---
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter found")
	}

	// Remove the opening ---
	rest := content[3:]
	rest = strings.TrimLeft(rest, "\n\r")

	// Find the closing ---
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, fmt.Errorf("unclosed frontmatter: no closing --- found")
	}

	yamlStr := strings.TrimSpace(rest[:endIdx])
	var frontmatter map[string]any
	if yamlStr != "" {
		if err := yaml.Unmarshal([]byte(yamlStr), &frontmatter); err != nil {
			return nil, fmt.Errorf("parse frontmatter: %w", err)
		}
	}

	// Body is everything after the closing ---
	body := ""
	if endIdx+4 < len(rest) {
		body = strings.TrimSpace(rest[endIdx+4:])
	}

	skill := &ParsedSkill{
		Frontmatter: frontmatter,
		Body:        body,
	}

	// Extract known fields
	if name, ok := frontmatter["name"]; ok {
		skill.Name = fmt.Sprintf("%v", name)
	}
	if desc, ok := frontmatter["description"]; ok {
		skill.Description = fmt.Sprintf("%v", desc)
	}
	if version, ok := frontmatter["version"]; ok {
		skill.Version = fmt.Sprintf("%v", version)
	}
	if author, ok := frontmatter["author"]; ok {
		skill.Author = fmt.Sprintf("%v", author)
	}
	if tagsRaw, ok := frontmatter["tags"]; ok {
		skill.Tags = parseTags(tagsRaw)
	}

	return skill, nil
}

// parseTags handles both []interface{} and comma-separated string tag formats.
func parseTags(tagsRaw any) []string {
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
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
		return tags
	default:
		return nil
	}
}

// ValidateSkill checks that a ParsedSkill has required fields (name, description, version).
func ValidateSkill(skill *ParsedSkill) error {
	if skill == nil {
		return fmt.Errorf("skill is nil")
	}
	if skill.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if skill.Description == "" {
		return fmt.Errorf("skill description is required")
	}
	if skill.Version == "" {
		return fmt.Errorf("skill version is required")
	}
	return nil
}