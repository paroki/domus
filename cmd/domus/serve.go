package main

import (
	"strings"

	"github.com/paroki/domus/internal/app"
	"github.com/paroki/domus/internal/event"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/ref"
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
			userCfg, err := user.GetConfig(ctx)
			if err != nil {
				return err
			}
			if err := app.ValidateStartupEnv(cfg, app.StartupOAuthConfig{
				GoogleID:     userCfg.GoogleID,
				GoogleSecret: userCfg.GoogleSecret,
				GitHubID:     userCfg.GitHubID,
				GitHubSecret: userCfg.GitHubSecret,
			}, logger); err != nil {
				return err
			}
			// ponytail: cipher constructed at startup to fail fast; wire into features when encrypted columns land.
			if _, err = crypt.New(cfg.Secret); err != nil {
				return err
			}

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

			var userOpts []user.Option
			if strings.TrimSpace(cfg.SMTPHost) != "" {
				mailer, err := mailx.New(mailx.Config{
					Host: cfg.SMTPHost,
					Port: cfg.SMTPPort,
					User: cfg.SMTPUser,
					Pass: cfg.SMTPPass,
					From: cfg.SMTPFrom,
				})
				if err != nil {
					return err
				}
				userOpts = append(userOpts, user.WithOTPSender(user.NewMailOTPSender(mailer, userCfg.OTPTTL)))
			}
			userFeature := user.New(db, enforcer, userCfg, logger, userOpts...)
			refFeature := ref.New(db, enforcer, userFeature)
			orgCfg, err := org.GetConfig(ctx)
			if err != nil {
				return err
			}
			orgFeature := org.New(db, enforcer, userFeature, refFeature, orgCfg)
			userFeature.SetInviteJoiner(orgFeature)
			eventCfg, err := event.GetConfig(ctx)
			if err != nil {
				return err
			}
			eventFeature := event.New(db, enforcer, userFeature, orgFeature, eventCfg)
			if err := authz.RegisterFeatures(ctx, enforcer, userFeature, refFeature, orgFeature, eventFeature); err != nil {
				return err
			}
			settingStore := setting.NewStore(db)
			setting.RegisterFeatures(settingStore, userFeature, refFeature, orgFeature, eventFeature)
			if err := settingStore.Load(ctx); err != nil {
				return err
			}
			settingFeature := setting.NewFeature(
				settingStore,
				userFeature.AuthenticateMiddleware(),
				userFeature.RequireAuth(),
				userFeature.Require(user.PermUsers, authz.ActionRead),
				userFeature.Require(user.PermUsers, authz.ActionWrite),
			)
			if err := userFeature.Bootstrap(ctx); err != nil {
				return err
			}

			fiberApp := app.GetFiber(cfg, logger, userFeature, refFeature, orgFeature, eventFeature, settingFeature)
			return app.New(cfg, fiberApp, logger).Serve(ctx)
		},
	}
}
