package main

import (
	"fmt"
	"os"

	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/spf13/cobra"
)

func (cli *CLI) registerDoctorCmd() {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostic checks on the skill repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			report := operations.RunDoctor(repoPath)

			var failed, warnings int
			for _, check := range report.Checks {
				switch check.Status {
				case "pass":
					fmt.Printf("[PASS] %s\n", check.Name)
				case "warn":
					fmt.Printf("[WARN] %s - %s\n", check.Name, check.Message)
					warnings++
				case "fail":
					fmt.Printf("[FAIL] %s - %s\n", check.Name, check.Message)
					failed++
				}
			}

			fmt.Fprintln(os.Stdout)

			if failed == 0 && warnings == 0 {
				fmt.Println("All checks passed!")
			} else {
				fmt.Printf("%d check(s) failed, %d warning(s)\n", failed, warnings)
			}

			return nil
		},
	}

	cli.RootCmd.AddCommand(cmd)
}
