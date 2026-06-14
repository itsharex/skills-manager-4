package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func (cli *CLI) registerEditCmd() {
	var namespace, version string

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Open a skill's SKILL.md in the system editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			skillPath := cli.Repo.SkillPath(namespace, name, version)
			skillMDPath := skillPath + "/SKILL.md"

			if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
				return fmt.Errorf("SKILL.md not found at %s", skillMDPath)
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				if runtime.GOOS == "windows" {
					editor = "notepad"
				} else {
					editor = "vim"
				}
			}

			fmt.Printf("Editing: %s\n", skillMDPath)

			editCmd := exec.Command(editor, skillMDPath)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			if err := editCmd.Run(); err != nil {
				return fmt.Errorf("editor exited with error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the skill")
	cmd.Flags().StringVarP(&version, "version", "v", "latest", "Version of the skill")

	cli.RootCmd.AddCommand(cmd)
}
