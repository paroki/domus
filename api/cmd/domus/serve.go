/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}
		app := config.GetFiber(cfg)
		return app.Listen(fmt.Sprintf(":%d", cfg.Port), fiber.ListenConfig{
			EnablePrefork: cfg.Api.Prefork,
		})
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
