package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/shotah/go-google-home-mcp/home"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List Nest devices",
	Args:  cobra.NoArgs,
	RunE:  runDevices,
}

var structuresCmd = &cobra.Command{
	Use:   "structures",
	Short: "List Nest structures (homes)",
	Args:  cobra.NoArgs,
	RunE:  runStructures,
}

func runDevices(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	devices, err := client.ListDevices(ctx)
	if err != nil {
		return err
	}
	out := make([]home.DeviceSummary, 0, len(devices))
	for _, d := range devices {
		out = append(out, home.SummarizeDevice(d))
	}
	return printJSON(out)
}

func runStructures(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	structures, err := client.ListStructures(ctx)
	if err != nil {
		return err
	}
	return printJSON(structures)
}
