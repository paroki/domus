/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Fiber API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}
		app := config.GetFiber(cfg)
		gofiber.SetupRouter(app)

		// Graceful shutdown
		idleConnsClosed := make(chan struct{})
		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
			<-sigint

			fmt.Println("\nShutting down server...")
			if err := app.Shutdown(); err != nil {
				fmt.Printf("Fiber shutdown error: %v\n", err)
			}
			close(idleConnsClosed)
		}()

		if err := app.Listen(fmt.Sprintf(":%d", cfg.Port), fiber.ListenConfig{
			EnablePrefork: cfg.Api.Prefork,
		}); err != nil {
			return err
		}

		<-idleConnsClosed
		fmt.Println("Server stopped.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
