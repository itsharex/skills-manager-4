package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerListCmd() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		Long:  "List all skills installed in the repository with their namespace, latest version, and version count.",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, _ := cmd.Flags().GetString("namespace")
			tagsFilter, _ := cmd.Flags().GetStringSlice("tags")

			entries, err := cli.Index.List()
			if err != nil {
				return fmt.Errorf("list index: %w", err)
			}

			// Filter by namespace
			if namespace != "" {
				var filtered []models.IndexEntry
				for _, e := range entries {
					if e.Namespace == namespace {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}

			// Filter by tags (any match)
			if len(tagsFilter) > 0 {
				tagSet := make(map[string]bool, len(tagsFilter))
				for _, t := range tagsFilter {
					tagSet[strings.ToLower(t)] = true
				}
				var filtered []models.IndexEntry
				for _, e := range entries {
					for _, t := range e.Tags {
						if tagSet[strings.ToLower(t)] {
							filtered = append(filtered, e)
							break
						}
					}
				}
				entries = filtered
			}

			if len(entries) == 0 {
				fmt.Println("No skills installed")
				return nil
			}

			// Sort by namespace, then name
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Namespace != entries[j].Namespace {
					return entries[i].Namespace < entries[j].Namespace
				}
				return entries[i].Name < entries[j].Name
			})

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "Name\tNamespace\tLatest Version\tVersions Count")
			fmt.Fprintln(w, "----\t---------\t--------------\t---------------")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", e.Name, e.Namespace, e.Latest, len(e.Versions))
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().String("namespace", "", "Filter by namespace")
	cmd.Flags().StringSlice("tags", nil, "Filter by tags (comma-separated)")

	cli.RootCmd.AddCommand(cmd)
}