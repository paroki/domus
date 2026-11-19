package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// @title       Domus API
// @version     0.1.0
// @description Parish Administration Digitalized API.
// @BasePath    /api
func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("domus exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	context.AfterFunc(ctx, stop)

	rootCmd := &cobra.Command{
		Use:           "domus",
		Short:         "Domus backend service",
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newServeCmd())

	return rootCmd.ExecuteContext(ctx)
}
