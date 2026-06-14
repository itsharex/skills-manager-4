package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerStatsCmd() {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show repository statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			// Build Index from List()
			index := &models.Index{
				Skills: make(map[string]models.IndexEntry),
			}
			entries, err := cli.Index.List()
			if err != nil {
				return fmt.Errorf("list index: %w", err)
			}
			for _, entry := range entries {
				key := fmt.Sprintf("%s/%s", entry.Namespace, entry.Name)
				index.Skills[key] = entry
			}

			// Build LockFile from List()
			lock := &models.LockFile{
				Skills: make(map[string]models.LockEntry),
			}
			lockEntries, err := cli.Lock.List()
			if err != nil {
				return fmt.Errorf("list lock entries: %w", err)
			}
			for _, entry := range lockEntries {
				key := fmt.Sprintf("%s/%s@%s", entry.SkillID.Namespace, entry.SkillID.Name, entry.SkillID.Version)
				lock.Skills[key] = entry
			}

			stats := operations.CollectStats(index, lock, repoPath)

			fmt.Printf("Skills: %d | Versions: %d | Namespaces: %d | Installed: %d | Agents: %d | Disk: %s\n",
				stats.TotalSkills, stats.TotalVersions, stats.TotalNamespaces,
				stats.InstalledSkills, stats.TotalAgents, operations.FormatBytes(stats.DiskUsageBytes))

			if len(stats.SkillsPerVersion) > 0 {
				fmt.Fprintln(os.Stdout)
				fmt.Println("Per-version breakdown:")

				var versions []string
				for v := range stats.SkillsPerVersion {
					versions = append(versions, v)
				}
				sort.Strings(versions)

				for _, v := range versions {
					fmt.Printf("  %s: %d skill(s)\n", v, stats.SkillsPerVersion[v])
				}
			}

			return nil
		},
	}

	cli.RootCmd.AddCommand(cmd)
}
