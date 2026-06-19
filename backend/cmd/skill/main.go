package main

import (
	"fmt"
	"os"

	"github.com/skillsmanager/skillsmanager/backend/internal/distribute"
	"github.com/skillsmanager/skillsmanager/backend/internal/operations"
	"github.com/skillsmanager/skillsmanager/backend/internal/source"
	"github.com/skillsmanager/skillsmanager/backend/internal/storage"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
	"github.com/spf13/cobra"
)

// CLI state shared across all commands.
type CLI struct {
	RootCmd     *cobra.Command
	Repo        *storage.Repository
	Index       *storage.Index
	Lock        *storage.LockFile
	InstallMode string
}

// NewCLI creates the CLI with all subcommands wired.
func NewCLI() *CLI {
	cli := &CLI{}

	rootCmd := &cobra.Command{
		Use:   "skill",
		Short: "Skills Manager - Install and manage AI agent skills",
		Long: `A cross-platform CLI tool for discovering, installing, and managing
AI agent skills from multiple sources (GitHub, HTTP registries, local files).`,
		PersistentPreRunE: cli.initRepo,
		SilenceUsage:      true,
	}

	rootCmd.PersistentFlags().String("pool", operations.DefaultPoolPath(), "Pool root path")
	rootCmd.PersistentFlags().Bool("copy", false, "Force copy mode instead of symlinks")
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	cli.RootCmd = rootCmd

	// Register all subcommands
	cli.registerInitCmd()
	cli.registerConfigCmd()
	cli.registerSearchCmd()
	cli.registerListCmd()
	cli.registerInfoCmd()
	cli.registerInstallCmd()
	cli.registerUninstallCmd()
	cli.registerUpdateCmd()
	cli.registerSyncCmd()
	cli.registerEditCmd()
	cli.registerDoctorCmd()
	cli.registerExportCmd()
	cli.registerImportCmd()
	cli.registerStatsCmd()

	return cli
}

// initRepo initializes or loads the pool state before each command.
func (cli *CLI) initRepo(cmd *cobra.Command, args []string) error {
	poolPath, _ := cmd.Flags().GetString("pool")
	forceCopy, _ := cmd.Flags().GetBool("copy")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Fprintf(os.Stderr, "Using pool: %s\n", poolPath)
	}

	// Ensure pool exists
	if err := operations.EnsurePoolDir(poolPath); err != nil {
		return fmt.Errorf("ensure pool: %w", err)
	}

	paths := operations.GetPoolPaths(poolPath)

	// Load config
	cfg, err := operations.LoadConfig(paths.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize storage
	repo := storage.NewRepository(poolPath)
	idx, err := storage.NewIndex(paths.IndexPath)
	if err != nil {
		return fmt.Errorf("init index: %w", err)
	}
	lock, err := storage.NewLockFile(paths.LockPath)
	if err != nil {
		return fmt.Errorf("init lock: %w", err)
	}

	if forceCopy {
		cli.InstallMode = "copy"
	} else {
		cli.InstallMode = cfg.InstallMode
	}

	cli.Repo = repo
	cli.Index = idx
	cli.Lock = lock
	return nil
}

// getInstaller creates an Installer from the current CLI state.
func (cli *CLI) getInstaller() *distribute.Installer {
	return distribute.NewInstaller(cli.Repo, cli.Index, cli.Lock)
}

// resolveSource is a bridge that uses source.NewResolver.
func resolveSource(sourceStr string) ([]models.ResolvedSkill, error) {
	resolver, err := source.NewResolver(sourceStr)
	if err != nil {
		return nil, err
	}
	return resolver.Resolve(nil, sourceStr, source.ResolveOptions{})
}

// printError prints an error to stderr and marks the command as failed.
func printError(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
}

func main() {
	cli := NewCLI()
	if err := cli.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}