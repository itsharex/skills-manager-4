package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillContent_Full(t *testing.T) {
	content := `---
name: my-skill
description: Does something useful
version: 1.0.0
tags: [ai, claude, pdf]
author: Someone
---

# My Skill

Content here with some **markdown**.
`

	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Name != "my-skill" {
		t.Errorf("expected Name 'my-skill', got %q", skill.Name)
	}
	if skill.Description != "Does something useful" {
		t.Errorf("expected Description 'Does something useful', got %q", skill.Description)
	}
	if skill.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", skill.Version)
	}
	if skill.Author != "Someone" {
		t.Errorf("expected Author 'Someone', got %q", skill.Author)
	}
	if len(skill.Tags) != 3 || skill.Tags[0] != "ai" || skill.Tags[1] != "claude" || skill.Tags[2] != "pdf" {
		t.Errorf("expected Tags [ai claude pdf], got %v", skill.Tags)
	}

	expectedBody := "# My Skill\n\nContent here with some **markdown**."
	if skill.Body != expectedBody {
		t.Errorf("expected Body %q, got %q", expectedBody, skill.Body)
	}

	if skill.Frontmatter == nil {
		t.Error("expected non-nil Frontmatter")
	}
}

func TestParseSkillContent_NoBody(t *testing.T) {
	content := `---
name: minimal
description: Just metadata
version: 0.1.0
---
`
	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Name != "minimal" {
		t.Errorf("expected Name 'minimal', got %q", skill.Name)
	}
	if skill.Body != "" {
		t.Errorf("expected empty Body, got %q", skill.Body)
	}
}

func TestParseSkillContent_NoFrontmatter(t *testing.T) {
	content := `# Just markdown, no frontmatter`
	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParseSkillContent_UnclosedFrontmatter(t *testing.T) {
	content := `---
name: broken
version: 1.0.0
`
	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
}

func TestParseSkillContent_EmptyFrontmatter(t *testing.T) {
	content := `---
---
`
	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for empty frontmatter")
	}
}

func TestParseSkillContent_TagsAsString(t *testing.T) {
	content := `---
name: tagged-skill
description: A skill with tags
version: 2.0.0
tags: ai, claude, pdf
---
`
	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if len(skill.Tags) != 3 || skill.Tags[0] != "ai" || skill.Tags[1] != "claude" || skill.Tags[2] != "pdf" {
		t.Errorf("expected Tags [ai claude pdf], got %v", skill.Tags)
	}
}

func TestParseSkillContent_TagsQuotedString(t *testing.T) {
	content := `---
name: quoted-tags
description: With quoted string tags
version: 1.0.0
tags: "single-tag"
---
`
	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if len(skill.Tags) != 1 || skill.Tags[0] != "single-tag" {
		t.Errorf("expected Tags [single-tag], got %v", skill.Tags)
	}
}

func TestParseSkillContent_NoTags(t *testing.T) {
	content := `---
name: no-tags
description: No tags here
version: 1.0.0
---
`
	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Tags != nil {
		t.Errorf("expected nil Tags, got %v", skill.Tags)
	}
}

func TestParseSkillContent_ExtraWhitespace(t *testing.T) {
	content := `   
---
name: whitespace
description: Has surrounding whitespace
version: 1.0.0
---

Some body
   `
	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Name != "whitespace" {
		t.Errorf("expected Name 'whitespace', got %q", skill.Name)
	}
	if skill.Body != "Some body" {
		t.Errorf("expected Body 'Some body', got %q", skill.Body)
	}
}

func TestParseSkillContent_InvalidYAML(t *testing.T) {
	content := `---
name: test
description: bad yaml
version: [1.0.0
---
`
	_, err := ParseSkillContent(content)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseSkillFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := `---
name: file-skill
description: From a file
version: 3.0.0
author: Tester
---

# File-based skill
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skill, err := ParseSkillFile(path)
	if err != nil {
		t.Fatalf("ParseSkillFile failed: %v", err)
	}

	if skill.Name != "file-skill" {
		t.Errorf("expected Name 'file-skill', got %q", skill.Name)
	}
	if skill.Description != "From a file" {
		t.Errorf("expected Description 'From a file', got %q", skill.Description)
	}
	if skill.Version != "3.0.0" {
		t.Errorf("expected Version '3.0.0', got %q", skill.Version)
	}
	if skill.Author != "Tester" {
		t.Errorf("expected Author 'Tester', got %q", skill.Author)
	}
}

func TestParseSkillFile_NotFound(t *testing.T) {
	_, err := ParseSkillFile("/nonexistent/path/SKILL.md")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestValidateSkill_Valid(t *testing.T) {
	skill := &ParsedSkill{
		Name:        "valid-skill",
		Description: "A valid skill",
		Version:     "1.0.0",
	}
	if err := ValidateSkill(skill); err != nil {
		t.Fatalf("ValidateSkill failed: %v", err)
	}
}

func TestValidateSkill_Nil(t *testing.T) {
	if err := ValidateSkill(nil); err == nil {
		t.Fatal("expected error for nil skill")
	}
}

func TestValidateSkill_MissingName(t *testing.T) {
	skill := &ParsedSkill{
		Description: "Missing name",
		Version:     "1.0.0",
	}
	if err := ValidateSkill(skill); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateSkill_MissingDescription(t *testing.T) {
	skill := &ParsedSkill{
		Name:    "no-desc",
		Version: "1.0.0",
	}
	if err := ValidateSkill(skill); err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestValidateSkill_MissingVersion(t *testing.T) {
	skill := &ParsedSkill{
		Name:        "no-version",
		Description: "Missing version",
	}
	if err := ValidateSkill(skill); err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestValidateSkill_EmptyFields(t *testing.T) {
	tests := []struct {
		name    string
		skill   *ParsedSkill
		wantErr bool
	}{
		{"all empty", &ParsedSkill{}, true},
		{"empty name", &ParsedSkill{Description: "d", Version: "1"}, true},
		{"empty desc", &ParsedSkill{Name: "n", Version: "1"}, true},
		{"empty version", &ParsedSkill{Name: "n", Description: "d"}, true},
		{"all valid", &ParsedSkill{Name: "n", Description: "d", Version: "1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkill(tt.skill)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSkill() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}