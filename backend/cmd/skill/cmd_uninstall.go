package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func (cli *CLI) registerUninstallCmd() {
	var namespace, version string

	cmd := &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Parse namespace from name if it contains "/"
			if strings.Contains(name, "/") {
				parts := strings.SplitN(name, "/", 2)
				if namespace == "" {
					namespace = parts[0]
				}
				name = parts[1]
			}
			if namespace == "" {
				namespace = "local"
			}

			// If version not specified, try to get from index
			if version == "" {
				entries, err := cli.Index.List()
				if err == nil {
					for _, e := range entries {
						if e.Name == name && e.Namespace == namespace {
							version = e.Latest
							break
						}
					}
				}
			}
			if version == "" {
				version = "latest"
			}

			if err := cli.getInstaller().Uninstall(namespace, name, version); err != nil {
				return fmt.Errorf("uninstall: %w", err)
			}

			fmt.Printf("✓ %s@%s uninstalled\n", name, version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the skill")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version of the skill")

	cli.RootCmd.AddCommand(cmd)
}