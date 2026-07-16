package home

import "context"

// SetThermostatMode sets HEAT / COOL / HEATCOOL / OFF.
func (c *Client) SetThermostatMode(ctx context.Context, deviceName, mode string) error {
	return c.ExecuteCommand(ctx, deviceName, "sdm.devices.commands.ThermostatMode.SetMode", map[string]any{
		"mode": mode,
	})
}

// SetThermostatHeat sets the heat setpoint (°C).
func (c *Client) SetThermostatHeat(ctx context.Context, deviceName string, celsius float64) error {
	return c.ExecuteCommand(ctx, deviceName, "sdm.devices.commands.ThermostatTemperatureSetpoint.SetHeat", map[string]any{
		"heatCelsius": celsius,
	})
}

// SetThermostatCool sets the cool setpoint (°C).
func (c *Client) SetThermostatCool(ctx context.Context, deviceName string, celsius float64) error {
	return c.ExecuteCommand(ctx, deviceName, "sdm.devices.commands.ThermostatTemperatureSetpoint.SetCool", map[string]any{
		"coolCelsius": celsius,
	})
}

// SetThermostatHeatCool sets both setpoints (°C).
func (c *Client) SetThermostatHeatCool(ctx context.Context, deviceName string, heatC, coolC float64) error {
	return c.ExecuteCommand(ctx, deviceName, "sdm.devices.commands.ThermostatTemperatureSetpoint.SetRange", map[string]any{
		"heatCelsius": heatC,
		"coolCelsius": coolC,
	})
}

// SetThermostatEco enables or disables MANUAL_ECO.
func (c *Client) SetThermostatEco(ctx context.Context, deviceName, mode string) error {
	return c.ExecuteCommand(ctx, deviceName, "sdm.devices.commands.ThermostatEco.SetMode", map[string]any{
		"mode": mode,
	})
}
