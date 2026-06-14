package main

import (
	"fmt"

	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/spf13/cobra"
)

// registerInitCmd registers the `skill init` subcommand.
func (cli *CLI) registerInitCmd() {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new skill repository",
		Long:  "Creates a new empty skill repository with default configuration at the given path.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			if err := operations.InitRepo(repoPath); err != nil {
				return fmt.Errorf("init repo: %w", err)
			}

			fmt.Printf("Initialized empty skill repository at %s\n", repoPath)
			return nil
		},
	}

	cli.RootCmd.AddCommand(cmd)
}