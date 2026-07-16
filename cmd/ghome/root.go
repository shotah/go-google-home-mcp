package main

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "ghome",
	Short:   "Google Home / Nest Device Access CLI + MCP server",
	Long:    "Control Nest devices via the Smart Device Management API. Auth once with `ghome login`, then use CLI commands or `ghome mcp`.",
	Version: version,
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(devicesCmd)
	rootCmd.AddCommand(structuresCmd)
	rootCmd.AddCommand(roomsCmd)
	rootCmd.AddCommand(thermostatCmd)
}

func printJSON(v any) error {
	enc := jsonEncoder()
	return enc.Encode(v)
}
