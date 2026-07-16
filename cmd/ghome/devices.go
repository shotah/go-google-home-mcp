package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/shotah/go-google-home-mcp/home"
)

var (
	flagStructure string
	flagRoom      string
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

var structureGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get one structure by id",
	RunE:  runStructureGet,
}

var roomsCmd = &cobra.Command{
	Use:   "rooms",
	Short: "List or get rooms in a structure",
	Args:  cobra.NoArgs,
	RunE:  runRooms,
}

func init() {
	structureGetCmd.Flags().StringVar(&flagStructure, "structure", "", "Structure id or full resource name")
	_ = structureGetCmd.MarkFlagRequired("structure")
	structuresCmd.AddCommand(structureGetCmd)

	roomsCmd.Flags().StringVar(&flagStructure, "structure", "", "Structure id or full resource name")
	_ = roomsCmd.MarkFlagRequired("structure")
	roomsCmd.Flags().StringVar(&flagRoom, "room", "", "Room id (omit to list all rooms)")
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

func runStructureGet(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	structure, err := client.GetStructure(ctx, flagStructure)
	if err != nil {
		return err
	}
	return printJSON(structure)
}

func runRooms(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if flagRoom != "" {
		room, err := client.GetRoom(ctx, flagStructure, flagRoom)
		if err != nil {
			return err
		}
		return printJSON(room)
	}

	rooms, err := client.ListRooms(ctx, flagStructure)
	if err != nil {
		return err
	}
	return printJSON(rooms)
}
