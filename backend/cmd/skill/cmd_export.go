package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func (cli *CLI) registerExportCmd() {
	var namespace, version, output string

	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export a skill to a local directory or print its path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			skillPath := cli.Repo.SkillPath(namespace, name, version)

			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				return fmt.Errorf("skill not found at %s", skillPath)
			}

			if output != "" {
				if err := copyDir(skillPath, output); err != nil {
					return fmt.Errorf("export skill: %w", err)
				}
				fmt.Printf("Exported skill %q to %s\n", name, output)
			} else {
				fmt.Println(skillPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the skill")
	cmd.Flags().StringVarP(&version, "version", "v", "latest", "Version of the skill")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory path")

	cli.RootCmd.AddCommand(cmd)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
