package main

import (
	"fmt"

	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerImportCmd() {
	var namespace, version string
	var agents []string

	cmd := &cobra.Command{
		Use:   "import <source>",
		Short: "Import a skill from a source (GitHub, HTTP, or local path)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceStr := args[0]

			skills, err := resolveSource(sourceStr)
			if err != nil {
				return fmt.Errorf("resolve source: %w", err)
			}

			if len(skills) == 0 {
				return fmt.Errorf("no skills resolved from %q", sourceStr)
			}

			installer := cli.getInstaller()
			if cli.InstallMode == "copy" {
				installer.ForceCopy = true
			}

			for _, skill := range skills {
				fmt.Printf("Importing %s/%s...\n", skill.Namespace, skill.Name)

				opts := distribute.InstallOptions{
					Namespace: namespace,
					Version:   version,
					Agents:    agents,
					ForceCopy: installer.ForceCopy,
					NoSync:    len(agents) == 0,
				}

				result, err := installer.Install(skill, opts)
				if err != nil {
					printError("install %s/%s: %v", skill.Namespace, skill.Name, err)
					continue
				}

				fmt.Printf("Imported %s/%s@%s to %s\n",
					result.Namespace, result.Name, result.Version, result.StorePath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Override namespace")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Specific version to import")
	cmd.Flags().StringSliceVarP(&agents, "agents", "a", nil, "Target agents (comma-separated)")

	cli.RootCmd.AddCommand(cmd)
}
