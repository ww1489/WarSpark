package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ww1489/WarSpark/internal/app"
)

func newServerCommand(rootOpts *rootOptions) *cobra.Command {
	opts := &app.StartOptions{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ConfigPath = rootOpts.configPath
			return app.Start(context.Background(), *opts)
		},
	}

	cmd.Flags().IntVar(&opts.PortOverride, "port", 0, "override server port; 0 uses config file")
	return cmd
}
