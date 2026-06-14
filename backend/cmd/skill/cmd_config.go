package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/spf13/cobra"
)

// registerConfigCmd registers the `skill config` and `skill repo` subcommands.
func (cli *CLI) registerConfigCmd() {
	cli.registerConfigSubcommands()
	cli.registerRepoSubcommands()
}

// registerConfigSubcommands sets up `skill config get` and `skill config set`.
func (cli *CLI) registerConfigSubcommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage skill repository configuration",
		Long:  "View and modify skill repository configuration settings.",
	}

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Display the value of a configuration key.

Available keys:
  repo_path       Repository root path
  install_mode    Installation mode ("symlink" or "copy")
  auto_fallback   Auto-fallback from symlink to copy (true/false)
  default_agents  Default agent list
  cache_ttl       Cache TTL in seconds
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			repoPath, _ := cmd.Flags().GetString("repo")
			paths := operations.GetRepoPaths(repoPath)

			cfg, err := operations.LoadConfig(paths.ConfigPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			val, err := configFieldByJSONTag(cfg, key)
			if err != nil {
				printError("unknown config key: %s", key)
				return fmt.Errorf("unknown config key: %s", key)
			}

			fmt.Println(val)
			return nil
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set the value of a configuration key.

Available keys:
  repo_path       Repository root path (string)
  install_mode    Installation mode: "symlink" or "copy"
  auto_fallback   Auto-fallback from symlink to copy: true or false
  cache_ttl       Cache TTL in seconds (integer)
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			repoPath, _ := cmd.Flags().GetString("repo")
			paths := operations.GetRepoPaths(repoPath)

			cfg, err := operations.LoadConfig(paths.ConfigPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if err := setConfigFieldByJSONTag(cfg, key, value); err != nil {
				printError("set config: %v", err)
				return err
			}

			if err := operations.SaveConfig(paths.ConfigPath, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}

	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	cli.RootCmd.AddCommand(configCmd)
}

// registerRepoSubcommands sets up `skill repo add`, `skill repo remove`, and `skill repo list`.
func (cli *CLI) registerRepoSubcommands() {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repository sources",
		Long:  "Add, remove, and list skill repository sources.",
	}

	addCmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a repository source",
		Long: `Add a new skill repository source.

Example:
  skill repo add my-skills https://skills.example.com --type registry
  skill repo add my-org github:my-org/skills --type github
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]
			sourceType, _ := cmd.Flags().GetString("type")

			if sourceType != "registry" && sourceType != "github" {
				printError("invalid type %q; must be \"registry\" or \"github\"", sourceType)
				return fmt.Errorf("invalid type %q; must be \"registry\" or \"github\"", sourceType)
			}

			repoPath, _ := cmd.Flags().GetString("repo")
			paths := operations.GetRepoPaths(repoPath)

			cfg, err := operations.LoadConfig(paths.ConfigPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Check for duplicate name
			for _, r := range cfg.Repositories {
				if r.Name == name {
					printError("repository %q already exists", name)
					return fmt.Errorf("repository %q already exists", name)
				}
			}

			cfg.Repositories = append(cfg.Repositories, models.RepoSource{
				Name:    name,
				URL:     url,
				Type:    sourceType,
				Enabled: true,
			})

			if err := operations.SaveConfig(paths.ConfigPath, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Added repository %s (%s)\n", name, url)
			return nil
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository source",
		Long:  "Remove a repository source by name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			repoPath, _ := cmd.Flags().GetString("repo")
			paths := operations.GetRepoPaths(repoPath)

			cfg, err := operations.LoadConfig(paths.ConfigPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			found := false
			filtered := make([]models.RepoSource, 0, len(cfg.Repositories))
			for _, r := range cfg.Repositories {
				if r.Name == name {
					found = true
					continue
				}
				filtered = append(filtered, r)
			}

			if !found {
				printError("repository %q not found", name)
				return fmt.Errorf("repository %q not found", name)
			}

			cfg.Repositories = filtered
			if err := operations.SaveConfig(paths.ConfigPath, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("Removed repository %s\n", name)
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all repository sources",
		Long:  "Display all configured repository sources.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")
			paths := operations.GetRepoPaths(repoPath)

			cfg, err := operations.LoadConfig(paths.ConfigPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if len(cfg.Repositories) == 0 {
				fmt.Println("No repositories configured")
				return nil
			}

			fmt.Printf("%-20s %-40s %-12s %s\n", "Name", "URL", "Type", "Enabled")
			fmt.Printf("%-20s %-40s %-12s %s\n", "----", "---", "----", "-------")
			for _, r := range cfg.Repositories {
				enabled := "yes"
				if !r.Enabled {
					enabled = "no"
				}
				fmt.Printf("%-20s %-40s %-12s %s\n", r.Name, r.URL, r.Type, enabled)
			}
			return nil
		},
	}

	addCmd.Flags().String("type", "registry", "Source type (registry or github)")

	repoCmd.AddCommand(addCmd)
	repoCmd.AddCommand(removeCmd)
	repoCmd.AddCommand(listCmd)
	cli.RootCmd.AddCommand(repoCmd)
}

// configFieldByJSONTag looks up a Config field by its JSON tag name and returns its value as a string.
func configFieldByJSONTag(cfg *models.Config, key string) (string, error) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		// Handle json:"name,omitempty" etc.
		tagName := strings.Split(tag, ",")[0]
		if tagName != key {
			continue
		}

		fv := v.Field(i)
		return formatFieldValue(fv), nil
	}

	return "", fmt.Errorf("unknown config key: %s", key)
}

// setConfigFieldByJSONTag looks up a Config field by its JSON tag and sets it from a string value.
func setConfigFieldByJSONTag(cfg *models.Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		tagName := strings.Split(tag, ",")[0]
		if tagName != key {
			continue
		}

		fv := v.Field(i)
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(value)
			return nil
		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid bool value %q", value)
			}
			fv.SetBool(b)
			return nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value %q", value)
			}
			fv.SetInt(n)
			return nil
		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid float value %q", value)
			}
			fv.SetFloat(n)
			return nil
		default:
			return fmt.Errorf("cannot set field %q of type %s", key, fv.Type())
		}
	}

	return fmt.Errorf("unknown config key: %s", key)
}

// formatFieldValue formats a reflect.Value as a string for display.
func formatFieldValue(fv reflect.Value) string {
	switch fv.Kind() {
	case reflect.String:
		return fv.String()
	case reflect.Bool:
		return strconv.FormatBool(fv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fv.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fv.Float(), 'g', -1, 64)
	case reflect.Slice:
		elems := make([]string, 0, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			elems = append(elems, fmt.Sprintf("%v", fv.Index(i).Interface()))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	default:
		return fmt.Sprintf("%v", fv.Interface())
	}
}