package main

import (
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/app"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending database migrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cfg, err := app.GetConfig(ctx)
			if err != nil {
				return err
			}
			logger := app.GetLogger(cfg)
			return app.RunMigrations(ctx, cfg, logger, agent.EnsureMemorySearchSchema)
		},
	}
}
