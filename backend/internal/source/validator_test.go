package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSkillContent_ValidSkill(t *testing.T) {
	content := `---
name: my-skill
description: A useful test skill
version: 1.0.0
author: test-author
tags: [test, utility]
---

# My Skill

This is the skill body content.
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors=%v warnings=%v", result.Errors, result.Warnings)
	}
}

func TestValidateSkillContent_MissingName(t *testing.T) {
	content := `---
description: No name here
version: 1.0.0
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for missing name")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least 1 error for missing name")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on field 'name'")
	}
}

func TestValidateSkillContent_MissingDescription(t *testing.T) {
	content := `---
name: test-skill
version: 1.0.0
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for missing description")
	}
}

func TestValidateSkillContent_MissingVersion(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for missing version")
	}
}

func TestValidateSkillContent_InvalidNameFormat(t *testing.T) {
	content := `---
name: invalid name with spaces!
description: A test skill
version: 1.0.0
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for bad name format")
	}
}

func TestValidateSkillContent_InvalidSemver(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: not-a-version
author: me
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid (semver warning is not an error)")
	}
	if len(result.Warnings) == 0 {
		t.Error("expected at least 1 warning for bad semver")
	}
}

func TestValidateSkillContent_LatestVersion(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: latest
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid for 'latest' version, got errors=%v", result.Errors)
	}
}

func TestValidateSkillContent_InvalidTags(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
tags: [valid-tag, invalid tag with spaces, another-valid]
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid (tag warnings are not errors)")
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warnings for invalid tags")
	}
}

func TestValidateSkillContent_DuplicateTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			"duplicate exact",
			`---
name: test-skill
description: A test skill
version: 1.0.0
tags: [duplicate, duplicate]
---
`,
		},
		{
			"duplicate case-insensitive",
			`---
name: test-skill
description: A test skill
version: 1.0.0
tags: [Test, test]
---
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSkillContent(tt.content)
			if err != nil {
				t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
			}
			found := false
			for _, w := range result.Warnings {
				if w.Field == "tags" {
					found = true
					break
				}
			}
			if !found {
				t.Error("expected warning for duplicate tags")
			}
		})
	}
}

func TestValidateSkillContent_EmptyBody(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
---
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid (empty body is just a warning)")
	}
	found := false
	for _, w := range result.Warnings {
		if w.Field == "body" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for empty body")
	}
}

func TestValidateSkillContent_MissingAuthor(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
tags: [test]
---

Body content here.
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid (missing author is just a warning)")
	}
	found := false
	for _, w := range result.Warnings {
		if w.Field == "author" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for missing author")
	}
}

func TestValidateSkillContent_NoFrontmatter(t *testing.T) {
	content := `# Just a heading

Some content without frontmatter.
`
	_, err := ValidateSkillContent(content)
	if err == nil {
		t.Error("expected error for content without frontmatter")
	}
}

func TestValidateSkillContent_CommaSeparatedTags(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
tags: test, utility, web
---

Body
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid for comma-separated tags, got errors=%v", result.Errors)
	}
}

func TestValidateSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "SKILL.md")
	content := `---
name: file-skill
description: Skill loaded from file
version: 2.0.0
author: tester
tags: [file]
---

# File Skill
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := ValidateSkillFile(path)
	if err != nil {
		t.Fatalf("ValidateSkillFile() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors=%v warnings=%v", result.Errors, result.Warnings)
	}
}

func TestValidateSkillFile_Nonexistent(t *testing.T) {
	_, err := ValidateSkillFile("/nonexistent/path/SKILL.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidateParsedSkill_Nil(t *testing.T) {
	result := ValidateParsedSkill(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Valid {
		t.Error("expected invalid for nil skill")
	}
}

func TestValidateSkillContent_MultipleErrors(t *testing.T) {
	content := `---
invalid-field: true
---

`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for missing multiple required fields")
	}
	if len(result.Errors) < 1 {
		t.Errorf("expected at least 1 error, got %d", len(result.Errors))
	}
	// storage.ValidateSkill returns on first missing field, so we get exactly 1 error
}

func TestValidateSkillContent_EdgeCaseEmptyFrontmatter(t *testing.T) {
	content := `---
---

# Just body
`
	_, err := ValidateSkillContent(content)
	if err == nil {
		t.Error("expected error for empty frontmatter")
	}
}

func TestValidateSkillContent_TagsAsCommaString(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
tags: valid-tag, spaces-ok-too, another-one
---

Body
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid for string tags, got errors=%v", result.Errors)
	}
}

func TestValidateSkillContent_PerfectValidSemver(t *testing.T) {
	content := `---
name: perfect-skill
description: Perfectly valid skill
version: 1.22.333
author: author-name
tags: [alpha, beta, gamma]
---

# Perfect Skill

Full documentation here.
`
	result, err := ValidateSkillContent(content)
	if err != nil {
		t.Fatalf("ValidateSkillContent() unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors=%v warnings=%v", result.Errors, result.Warnings)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(result.Errors))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(result.Warnings))
	}
}