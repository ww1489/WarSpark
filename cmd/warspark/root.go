package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

type rootOptions struct {
	configPath string
}

func execute() error {
	opts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:   appconfig.AppName,
		Short: appconfig.AppDisplayName + " backend",
		Long:  appconfig.AppDisplayName + " backend powered by Gin, MySQL, and Redis.",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(
		&opts.configPath,
		"config",
		"configs/config.yaml",
		"config file path",
	)
	rootCmd.AddCommand(newServerCommand(opts))
	rootCmd.AddCommand(newMigrateCommand(opts))
	rootCmd.AddCommand(newCocapiCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
