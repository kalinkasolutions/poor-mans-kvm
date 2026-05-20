package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Monitor struct {
	ConnectInputCode    string
	DisconnectInputCode string
	DisplayBusNumber    string
	VcpInputSourceCode  string
}

type Config struct {
	DeviceID  string
	Monitors  []Monitor
}

func loadConfig(configLocation string, cfg *Config) error {
	logMessage("loading config from %s", configLocation)

	data, err := os.ReadFile(configLocation)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("error unmarshalling config file: %w", err)
	}

	logMessage("config loaded: %+v", *cfg)
	return nil
}
