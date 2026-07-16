package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/shotah/go-google-home-mcp/home"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for LLM / ZeroClaw integration",
	Long:  "Expose Nest Device Access tools over stdio MCP (list devices, thermostat control).",
	RunE:  runMCP,
}

func runMCP(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	s := server.NewMCPServer(
		"google-home",
		version,
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool(
		"list_devices",
		mcp.WithDescription("List Nest / Google Home devices linked to Device Access (thermostats, cameras, doorbells, displays)."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		devices, err := client.ListDevices(ctx)
		if err != nil {
			return mcpError(err), nil
		}
		out := make([]home.DeviceSummary, 0, len(devices))
		for _, d := range devices {
			out = append(out, home.SummarizeDevice(d))
		}
		return mcpJSON(out)
	})

	s.AddTool(mcp.NewTool(
		"get_device",
		mcp.WithDescription("Get full SDM traits for one device."),
		mcp.WithString("device", mcp.Required(), mcp.Description("Device id or full enterprises/.../devices/... name")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		device, err := req.RequireString("device")
		if err != nil {
			return mcpError(err), nil
		}
		dev, err := client.GetDevice(ctx, device)
		if err != nil {
			return mcpError(err), nil
		}
		return mcpJSON(dev)
	})

	s.AddTool(mcp.NewTool(
		"list_structures",
		mcp.WithDescription("List Nest structures (homes) authorized for this project."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		structures, err := client.ListStructures(ctx)
		if err != nil {
			return mcpError(err), nil
		}
		return mcpJSON(structures)
	})

	s.AddTool(mcp.NewTool(
		"set_thermostat_mode",
		mcp.WithDescription("Set thermostat mode: HEAT, COOL, HEATCOOL, or OFF."),
		mcp.WithString("device", mcp.Required(), mcp.Description("Device id or resource name")),
		mcp.WithString("mode", mcp.Required(), mcp.Description("HEAT|COOL|HEATCOOL|OFF")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		device, err := req.RequireString("device")
		if err != nil {
			return mcpError(err), nil
		}
		mode, err := req.RequireString("mode")
		if err != nil {
			return mcpError(err), nil
		}
		if err := client.SetThermostatMode(ctx, device, mode); err != nil {
			return mcpError(err), nil
		}
		return mcpJSON(map[string]any{"ok": true, "mode": mode})
	})

	s.AddTool(mcp.NewTool(
		"set_thermostat_temperature",
		mcp.WithDescription("Set thermostat heat and/or cool setpoints in Celsius."),
		mcp.WithString("device", mcp.Required(), mcp.Description("Device id or resource name")),
		mcp.WithNumber("heat_celsius", mcp.Description("Heat setpoint °C")),
		mcp.WithNumber("cool_celsius", mcp.Description("Cool setpoint °C")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		device, err := req.RequireString("device")
		if err != nil {
			return mcpError(err), nil
		}
		args := req.GetArguments()
		heat, hasHeat := numberArg(args, "heat_celsius")
		cool, hasCool := numberArg(args, "cool_celsius")
		switch {
		case hasHeat && hasCool:
			err = client.SetThermostatHeatCool(ctx, device, heat, cool)
		case hasHeat:
			err = client.SetThermostatHeat(ctx, device, heat)
		case hasCool:
			err = client.SetThermostatCool(ctx, device, cool)
		default:
			return mcpError(fmt.Errorf("provide heat_celsius and/or cool_celsius")), nil
		}
		if err != nil {
			return mcpError(err), nil
		}
		return mcpJSON(map[string]any{"ok": true, "heat_celsius": heat, "cool_celsius": cool})
	})

	s.AddTool(mcp.NewTool(
		"set_thermostat_eco",
		mcp.WithDescription("Enable or disable thermostat eco mode."),
		mcp.WithString("device", mcp.Required(), mcp.Description("Device id or resource name")),
		mcp.WithString("mode", mcp.Required(), mcp.Description("MANUAL_ECO or OFF")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		device, err := req.RequireString("device")
		if err != nil {
			return mcpError(err), nil
		}
		mode, err := req.RequireString("mode")
		if err != nil {
			return mcpError(err), nil
		}
		if err := client.SetThermostatEco(ctx, device, mode); err != nil {
			return mcpError(err), nil
		}
		return mcpJSON(map[string]any{"ok": true, "mode": mode})
	})

	return server.ServeStdio(s)
}

func numberArg(args map[string]any, key string) (float64, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func mcpJSON(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcpError(err), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func mcpError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}
