package main

import (
	"fmt"

	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerInstallCmd() {
	cmd := &cobra.Command{
		Use:   "install <source>",
		Short: "Install a skill from a source",
		Long: `Install a skill from a source string such as a local path, GitHub URL, or registry reference.

Examples:
  skill install ./skills/my-skill
  skill install github.com/owner/repo
  skill install my-skill --namespace custom --version 1.0.0`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			namespace, _ := cmd.Flags().GetString("namespace")
			version, _ := cmd.Flags().GetString("version")
			agents, _ := cmd.Flags().GetStringSlice("agents")
			noSync, _ := cmd.Flags().GetBool("no-sync")
			forceCopy, _ := cmd.Root().Flags().GetBool("copy")

			// Resolve source
			skills, err := resolveSource(source)
			if err != nil {
				return fmt.Errorf("resolve source: %w", err)
			}

			installer := cli.getInstaller()

			for _, skill := range skills {
				opts := distribute.InstallOptions{
					Namespace: namespace,
					Version:   version,
					Agents:    agents,
					ForceCopy: forceCopy,
					NoSync:    noSync,
				}

				result, err := installer.Install(skill, opts)
				if err != nil {
					printError("install %s: %v", skill.Name, err)
					continue
				}

				agentsCount := len(agents)
				if agentsCount == 0 {
					agentsCount = 1 // default agent
				}
				fmt.Printf("✓ %s@%s installed to %d agent(s)\n", result.Name, result.Version, agentsCount)

				if result.Error != "" {
					fmt.Fprintf(cmd.ErrOrStderr(), "  warning: %s\n", result.Error)
				}
			}

			return nil
		},
	}

	cmd.Flags().String("namespace", "", "Override namespace for the skill")
	cmd.Flags().String("version", "", "Specific version to install")
	cmd.Flags().StringSlice("agents", nil, "Comma-separated agent IDs to install to")
	cmd.Flags().Bool("no-sync", false, "Skip agent sync, store only")

	cli.RootCmd.AddCommand(cmd)
}