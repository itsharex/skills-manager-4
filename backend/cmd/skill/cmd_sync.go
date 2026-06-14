package main

import (
	"fmt"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerSyncCmd() {
	var name string
	var agents []string
	var all bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync installed skills to agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			forceCopy := cli.InstallMode == "copy"

			if all {
				return cli.syncAll(forceCopy, agents)
			}

			if name != "" {
				return cli.syncByName(name, forceCopy, agents)
			}

			return fmt.Errorf("use --name <name> or --all to specify what to sync")
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Skill name to sync")
	cmd.Flags().StringSliceVar(&agents, "agents", nil, "Agent IDs to sync to (comma-separated)")
	cmd.Flags().BoolVar(&all, "all", false, "Sync all installed skills")

	cli.RootCmd.AddCommand(cmd)
}

func (cli *CLI) syncAll(forceCopy bool, agents []string) error {
	entries, err := cli.Lock.List()
	if err != nil {
		return fmt.Errorf("list lock: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No skills installed")
		return nil
	}

	var totalSummary distribute.SyncSummary
	agentSet := make(map[string]bool)

	for _, entry := range entries {
		skillPath := cli.Repo.SkillPath(entry.SkillID.Namespace, entry.SkillID.Name, entry.SkillID.Version)

		targetAgents := agents
		if len(targetAgents) == 0 {
			for _, a := range entry.Agents {
				targetAgents = append(targetAgents, a.AgentID)
			}
		}
		for _, a := range targetAgents {
			agentSet[a] = true
		}

		summary, err := distribute.SyncSkillsToAgents([]string{skillPath}, targetAgents, forceCopy)
		if err != nil {
			return fmt.Errorf("sync skills: %w", err)
		}
		totalSummary.Total += summary.Total
		totalSummary.Success += summary.Success
		totalSummary.Failed += summary.Failed
		totalSummary.Results = append(totalSummary.Results, summary.Results...)
	}

	fmt.Printf("Synced %d skills to %d agents (%d ok, %d failed)\n",
		len(entries), len(agentSet), totalSummary.Success, totalSummary.Failed)
	return nil
}

func (cli *CLI) syncByName(name string, forceCopy bool, agents []string) error {
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

	entries, err := cli.Lock.List()
	if err != nil {
		return fmt.Errorf("list lock: %w", err)
	}

	for _, entry := range entries {
		if entry.SkillID.Name == name && entry.SkillID.Namespace == namespace {
			skillPath := cli.Repo.SkillPath(entry.SkillID.Namespace, entry.SkillID.Name, entry.SkillID.Version)

			targetAgents := agents
			if len(targetAgents) == 0 {
				for _, a := range entry.Agents {
					targetAgents = append(targetAgents, a.AgentID)
				}
			}

			summary, err := distribute.SyncSkillsToAgents([]string{skillPath}, targetAgents, forceCopy)
			if err != nil {
				return fmt.Errorf("sync skill: %w", err)
			}

			fmt.Printf("Synced %d skills to %d agents (%d ok, %d failed)\n",
				1, len(targetAgents), summary.Success, summary.Failed)
			return nil
		}
	}

	return fmt.Errorf("skill %s/%s not found in lock file", namespace, name)
}