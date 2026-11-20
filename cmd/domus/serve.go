package main

import (
	"github.com/paroki/domus/internal/app"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cfg, err := app.GetConfig(ctx)
			if err != nil {
				return err
			}
			logger := app.GetLogger(cfg)
			flush, err := app.InitSentry(cfg)
			if err != nil {
				return err
			}
			defer flush()
			// ponytail: always auto-migrate on serve; add APP_AUTO_MIGRATE to Config when separate migration job needed.
			if err := app.RunMigrations(ctx, cfg, logger); err != nil {
				return err
			}

			db, err := app.OpenDB(ctx, cfg)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			enforcer, err := authz.NewEnforcer(db)
			if err != nil {
				return err
			}

			userCfg, err := user.GetConfig(ctx)
			if err != nil {
				return err
			}
			userFeature := user.New(db, enforcer, userCfg, logger)
			if err := authz.RegisterFeatures(ctx, enforcer, userFeature); err != nil {
				return err
			}
			if err := userFeature.Bootstrap(ctx); err != nil {
				return err
			}

			fiberApp := app.GetFiber(cfg, logger, userFeature)
			return app.New(cfg, fiberApp, logger).Serve(ctx)
		},
	}
}
