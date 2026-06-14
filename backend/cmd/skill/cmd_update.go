package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func (cli *CLI) registerUpdateCmd() {
	var version string
	var agents []string

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a skill to a new version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			namespace := ""

			// Parse namespace from name if it contains "/"
			if strings.Contains(name, "/") {
				parts := strings.SplitN(name, "/", 2)
				namespace = parts[0]
				name = parts[1]
			}
			if namespace == "" {
				namespace = "local"
			}

			// Find current version from lock file
			entries, err := cli.Lock.List()
			if err != nil {
				return fmt.Errorf("list lock: %w", err)
			}

			var oldVersion string
			for _, e := range entries {
				if e.SkillID.Name == name && e.SkillID.Namespace == namespace {
					oldVersion = e.SkillID.Version
					break
				}
			}
			if oldVersion == "" {
				return fmt.Errorf("skill %s/%s not found in lock file", namespace, name)
			}

			if version == "" {
				return fmt.Errorf("--version is required")
			}
			newVersion := version

			forceCopy := cli.InstallMode == "copy"

			if err := cli.getInstaller().UpdateSkill(namespace, name, oldVersion, newVersion, agents, forceCopy); err != nil {
				return fmt.Errorf("update: %w", err)
			}

			fmt.Printf("✓ %s updated from %s to %s\n", name, oldVersion, newVersion)
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "New version for the skill")
	cmd.Flags().StringSliceVar(&agents, "agents", nil, "Agent IDs to sync to (comma-separated)")

	cli.RootCmd.AddCommand(cmd)
}