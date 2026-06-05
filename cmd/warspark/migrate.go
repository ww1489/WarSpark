package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ww1489/WarSpark/internal/app"
)

type migrateOptions struct {
	path  string
	steps int
	force int
}

func newMigrateCommand(rootOpts *rootOptions) *cobra.Command {
	opts := &migrateOptions{path: "migrations"}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
	}
	cmd.PersistentFlags().StringVar(&opts.path, "path", "migrations", "migration files path")

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.MigrateUp(app.MigrationOptions{
				ConfigPath:     rootOpts.configPath,
				MigrationsPath: opts.path,
			})
		},
	})

	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Roll back migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.MigrateDown(app.MigrationOptions{
				ConfigPath:     rootOpts.configPath,
				MigrationsPath: opts.path,
			}, opts.steps)
		},
	}
	downCmd.Flags().IntVar(&opts.steps, "steps", 1, "number of migrations to roll back")
	cmd.AddCommand(downCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := app.CurrentMigrationVersion(app.MigrationOptions{
				ConfigPath:     rootOpts.configPath,
				MigrationsPath: opts.path,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%t\n", version.Version, version.Dirty)
			return nil
		},
	})

	forceCmd := &cobra.Command{
		Use:   "force",
		Short: "Force migration version after manual repair",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.MigrateForce(app.MigrationOptions{
				ConfigPath:     rootOpts.configPath,
				MigrationsPath: opts.path,
			}, opts.force)
		},
	}
	forceCmd.Flags().IntVar(&opts.force, "version", 0, "migration version to force")
	_ = forceCmd.MarkFlagRequired("version")
	cmd.AddCommand(forceCmd)

	return cmd
}
