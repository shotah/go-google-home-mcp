package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagThermostatDevice string
	flagThermostatMode   string
	flagHeatC            float64
	flagCoolC            float64
	flagEcoMode          string
	flagFanMode          string
	flagFanDuration      time.Duration
)

var thermostatCmd = &cobra.Command{
	Use:   "thermostat",
	Short: "Control Nest thermostats",
}

var thermostatModeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Set thermostat mode (HEAT|COOL|HEATCOOL|OFF)",
	RunE:  runThermostatMode,
}

var thermostatTempCmd = &cobra.Command{
	Use:   "temp",
	Short: "Set heat and/or cool setpoints (°C)",
	RunE:  runThermostatTemp,
}

var thermostatEcoCmd = &cobra.Command{
	Use:   "eco",
	Short: "Set eco mode (MANUAL_ECO|OFF)",
	RunE:  runThermostatEco,
}

var thermostatFanCmd = &cobra.Command{
	Use:   "fan",
	Short: "Set fan timer (ON|OFF)",
	RunE:  runThermostatFan,
}

func init() {
	thermostatCmd.PersistentFlags().StringVar(&flagThermostatDevice, "device", "", "Device id or full resource name (required)")
	_ = thermostatCmd.MarkPersistentFlagRequired("device")

	thermostatModeCmd.Flags().StringVar(&flagThermostatMode, "mode", "", "HEAT|COOL|HEATCOOL|OFF")
	_ = thermostatModeCmd.MarkFlagRequired("mode")

	thermostatTempCmd.Flags().Float64Var(&flagHeatC, "heat", 0, "Heat setpoint °C")
	thermostatTempCmd.Flags().Float64Var(&flagCoolC, "cool", 0, "Cool setpoint °C")

	thermostatEcoCmd.Flags().StringVar(&flagEcoMode, "mode", "", "MANUAL_ECO|OFF")
	_ = thermostatEcoCmd.MarkFlagRequired("mode")

	thermostatFanCmd.Flags().StringVar(&flagFanMode, "mode", "", "ON|OFF")
	_ = thermostatFanCmd.MarkFlagRequired("mode")
	thermostatFanCmd.Flags().DurationVar(&flagFanDuration, "duration", 0, "How long to run when ON (e.g. 1h, 30m); SDM default is 15m")

	thermostatCmd.AddCommand(thermostatModeCmd, thermostatTempCmd, thermostatEcoCmd, thermostatFanCmd)
}

func runThermostatMode(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.SetThermostatMode(ctx, flagThermostatDevice, flagThermostatMode); err != nil {
		return err
	}
	fmt.Println(`{"ok":true,"command":"SetMode","mode":"` + flagThermostatMode + `"}`)
	return nil
}

func runThermostatTemp(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch {
	case flagHeatC != 0 && flagCoolC != 0:
		err = client.SetThermostatHeatCool(ctx, flagThermostatDevice, flagHeatC, flagCoolC)
	case flagHeatC != 0:
		err = client.SetThermostatHeat(ctx, flagThermostatDevice, flagHeatC)
	case flagCoolC != 0:
		err = client.SetThermostatCool(ctx, flagThermostatDevice, flagCoolC)
	default:
		return fmt.Errorf("provide --heat and/or --cool")
	}
	if err != nil {
		return err
	}
	fmt.Println(`{"ok":true,"command":"SetTemperature"}`)
	return nil
}

func runThermostatEco(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.SetThermostatEco(ctx, flagThermostatDevice, flagEcoMode); err != nil {
		return err
	}
	fmt.Println(`{"ok":true,"command":"SetEco","mode":"` + flagEcoMode + `"}`)
	return nil
}

func runThermostatFan(_ *cobra.Command, _ []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.SetFanTimer(ctx, flagThermostatDevice, flagFanMode, flagFanDuration); err != nil {
		return err
	}
	fmt.Println(`{"ok":true,"command":"SetFanTimer","mode":"` + flagFanMode + `"}`)
	return nil
}
