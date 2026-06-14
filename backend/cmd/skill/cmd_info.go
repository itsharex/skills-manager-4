package main

import (
	"fmt"
	"os"
	"strings"

	sourcepkg "github.com/skillsmanager/skillsmanager/backend/internal/source"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerInfoCmd() {
	// --- skill info <name> ---
	infoCmd := &cobra.Command{
		Use:     "info <name>",
		Aliases: []string{"show"},
		Short:   "Show detailed information about an installed skill",
		Long:    "Display all metadata for an installed skill, including versions, source, description, and tags.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			namespace, _ := cmd.Flags().GetString("namespace")

			var entry *models.IndexEntry
			var err error

			if namespace != "" {
				key := fmt.Sprintf("%s/%s", namespace, name)
				entry, err = cli.Index.Get(key)
				if err != nil {
					return fmt.Errorf("skill not found: %w", err)
				}
			} else {
				// Search across all namespaces
				entries, listErr := cli.Index.List()
				if listErr != nil {
					return fmt.Errorf("list index: %w", listErr)
				}
				var matches []models.IndexEntry
				for _, e := range entries {
					if e.Name == name {
						matches = append(matches, e)
					}
				}
				if len(matches) == 0 {
					return fmt.Errorf("skill %q not found in any namespace", name)
				}
				if len(matches) > 1 {
					return fmt.Errorf("skill %q found in multiple namespaces (%s); use --namespace to disambiguate",
						name, formatNamespaces(matches))
				}
				entry = &matches[0]
			}

			printSkillInfo(entry)
			return nil
		},
	}

	infoCmd.Flags().String("namespace", "", "Skill namespace")

	// --- skill validate <source> ---
	validateCmd := &cobra.Command{
		Use:   "validate <source>",
		Short: "Validate a SKILL.md file",
		Long:  "Parse and validate a SKILL.md file, reporting any errors or warnings.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourcePath := args[0]

			data, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("read skill file: %w", err)
			}

			parsed, err := storage.ParseSkillContent(string(data))
			if err != nil {
				return fmt.Errorf("parse skill content: %w", err)
			}

			result := sourcepkg.ValidateParsedSkill(parsed)

			if result.Valid && len(result.Warnings) == 0 {
				fmt.Println("✓ Skill file is valid")
				return nil
			}

			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", e.Field, e.Message)
			}
			for _, w := range result.Warnings {
				fmt.Fprintf(os.Stderr, "Warning [%s]: %s\n", w.Field, w.Message)
			}

			if !result.Valid {
				return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
			}
			return nil
		},
	}

	cli.RootCmd.AddCommand(infoCmd)
	cli.RootCmd.AddCommand(validateCmd)
}

// printSkillInfo prints all metadata for an IndexEntry to stdout.
func printSkillInfo(entry *models.IndexEntry) {
	fmt.Printf("Name:        %s\n", entry.Name)
	fmt.Printf("Namespace:   %s\n", entry.Namespace)
	fmt.Printf("Versions:    %s\n", strings.Join(entry.Versions, ", "))
	fmt.Printf("Latest:      %s\n", entry.Latest)
	fmt.Printf("Source:      %s\n", entry.Source)
	if entry.Description != "" {
		fmt.Printf("Description: %s\n", entry.Description)
	}
	if len(entry.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(entry.Tags, ", "))
	}
}

// formatNamespaces returns a comma-separated list of namespaces from entries.
func formatNamespaces(entries []models.IndexEntry) string {
	ns := make([]string, len(entries))
	for i, e := range entries {
		ns[i] = e.Namespace
	}
	return strings.Join(ns, ", ")
}