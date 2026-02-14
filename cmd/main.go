package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yyewolf/kaloupile/pkg/config"
	"github.com/yyewolf/kaloupile/pkg/dependencies"
	"github.com/yyewolf/kaloupile/pkg/kind"
	"github.com/yyewolf/kaloupile/pkg/routes"
	"github.com/yyewolf/kaloupile/pkg/sync"
)

const defaultConfigPath = "config.yml"

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "kaloupile",
		Short: "Kaloupile cluster management",
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", defaultConfigPath, "Path to config.yml")

	cmd.AddCommand(newSetupCommand())
	cmd.AddCommand(newDependenciesCommand(&configPath))
	cmd.AddCommand(newRoutesCommand(&configPath))
	cmd.AddCommand(newSyncCommand(&configPath))
	cmd.AddCommand(newCleanupCommand())

	return cmd
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Install prerequisites and create the kind cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runStep("ensure kind cluster", kind.EnsureCluster); err != nil {
				return err
			}

			return runStep("install prerequisites", kind.InstallPrerequisites)
		},
	}
}

func newDependenciesCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "dependencies",
		Short: "Install dependencies according to config",
		RunE: func(cmd *cobra.Command, args []string) error {
			logStep("load config")
			cfg, err := config.LoadFromFile(*configPath)
			if err != nil {
				return err
			}
			if err := config.ValidateDependencies(cfg); err != nil {
				return err
			}
			logDone("load config")

			if err := runStep("install cert-manager dependencies", func() error {
				return dependencies.InstallCertManagerDependencies(cfg)
			}); err != nil {
				return err
			}

			if err := runStep("install postgresql", func() error {
				return dependencies.InstallPostgreSQL(cfg)
			}); err != nil {
				return err
			}
			if err := runStep("install fake smtp", dependencies.InstallFakeSMTP); err != nil {
				return err
			}

			return nil
		},
	}
}

func newRoutesCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "Install routes for apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			logStep("load config")
			cfg, err := config.LoadFromFile(*configPath)
			if err != nil {
				return err
			}
			logDone("load config")

			return runStep("install routes", func() error {
				return routes.InstallRoutes(cfg)
			})
		},
	}
}

func newSyncCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Run sync tasks (e.g. PostgreSQL users/databases)",
		RunE: func(cmd *cobra.Command, args []string) error {
			logStep("load config")
			cfg, err := config.LoadFromFile(*configPath)
			if err != nil {
				return err
			}
			logDone("load config")

			return runStep("sync postgresql", func() error {
				return sync.SyncPostgreSQL(cfg)
			})
		},
	}
}

func newCleanupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Delete the kind cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStep("delete kind cluster", kind.DeleteCluster)
		},
	}
}

func runStep(name string, fn func() error) error {
	logStep(name)
	start := time.Now()
	err := fn()
	if err != nil {
		logFail(name, err)
		return err
	}
	logDoneWithDuration(name, time.Since(start))
	return nil
}

func logStep(name string) {
	fmt.Printf("==> %s\n", name)
}

func logDone(name string) {
	fmt.Printf("<== %s done\n", name)
}

func logDoneWithDuration(name string, elapsed time.Duration) {
	fmt.Printf("<== %s done (%s)\n", name, elapsed.Round(time.Millisecond))
}

func logFail(name string, err error) {
	fmt.Printf("<== %s failed: %v\n", name, err)
}
