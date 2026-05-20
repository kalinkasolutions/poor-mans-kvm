package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	want := Config{
		DeviceID: "1234:5678",
		Monitors: []Monitor{
			{
				ConnectInputCode:    "0x11",
				DisconnectInputCode: "0x12",
				DisplayBusNumber:    "3",
				VcpInputSourceCode:  "0x60",
			},
		},
	}

	var got Config
	if err := loadConfig(writeTempConfig(t, want), &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.DeviceID != want.DeviceID {
		t.Errorf("DeviceID: got %q, want %q", got.DeviceID, want.DeviceID)
	}
	if len(got.Monitors) != 1 {
		t.Fatalf("Monitors: got %d, want 1", len(got.Monitors))
	}
	m := got.Monitors[0]
	if m.ConnectInputCode != "0x11" || m.DisconnectInputCode != "0x12" {
		t.Errorf("Monitor input codes: got connect=%q disconnect=%q", m.ConnectInputCode, m.DisconnectInputCode)
	}
	if m.DisplayBusNumber != "3" {
		t.Errorf("DisplayBusNumber: got %q, want %q", m.DisplayBusNumber, "3")
	}
	if m.VcpInputSourceCode != "0x60" {
		t.Errorf("VcpInputSourceCode: got %q, want %q", m.VcpInputSourceCode, "0x60")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	var cfg Config
	if err := loadConfig("/nonexistent/config.json", &cfg); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	fmt.Fprint(f, "not json")
	f.Close()

	var cfg Config
	if err := loadConfig(f.Name(), &cfg); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func writeTempConfig(t *testing.T, cfg Config) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
