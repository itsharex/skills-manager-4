package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// registerSearchCmd registers the `skill search <source>` subcommand.
func (cli *CLI) registerSearchCmd() {
	cmd := &cobra.Command{
		Use:   "search <source>",
		Short: "Search for skills from a source",
		Long: `Search for available skills from a source string.

Supported sources:
  - GitHub:   gh:owner/repo or github:owner/repo
  - HTTP:     https://...
  - Local:    /path/to/dir or ./path
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			// resolveSource is a package-level function from main.go.
			skills, err := resolveSource(source)
			if err != nil {
				return fmt.Errorf("resolve source: %w", err)
			}

			if len(skills) == 0 {
				fmt.Println("No skills found")
				return nil
			}

			// Table header
			fmt.Printf("%-30s %-25s %-15s\n", "Name", "Namespace", "Version")
			fmt.Printf("%-30s %-25s %-15s\n", "------", "---------", "-------")

			for _, s := range skills {
				name := s.Name
				if name == "" {
					name = "-"
				}
				ns := s.Namespace
				if ns == "" {
					ns = "-"
				}
				ver := s.Version
				if ver == "" {
					ver = "-"
				}
				fmt.Printf("%-30s %-25s %-15s\n", name, ns, ver)
			}

			return nil
		},
	}

	cmd.Flags().String("subpath", "", "Subpath within the source")
	cmd.Flags().String("version", "", "Specific version to search for")

	cli.RootCmd.AddCommand(cmd)
}