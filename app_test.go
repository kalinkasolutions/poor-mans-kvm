package main

import (
	"fmt"
	"os"
	"testing"
)

// writeTempScript creates an executable shell script that prints output and returns its path.
func writeTempScript(t *testing.T, output string) string {
	t.Helper()
	f, err := os.CreateTemp("", "mock-lsusb-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	fmt.Fprintf(f, "#!/bin/sh\nprintf '%%s' %q\n", output)
	f.Close()
	if err := os.Chmod(f.Name(), 0755); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestRunCommand_Succeeds(t *testing.T) {
	runCommand("echo", "hello")
}

func TestRunCommand_Failure(t *testing.T) {
	runCommand("false")
}

func TestGetUSBDevices(t *testing.T) {
	app := &App{lsusb: "echo"}
	out, err := app.getUSBDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "\n" {
		t.Errorf("got %q, want newline", out)
	}
}

func TestGetUSBDevices_Error(t *testing.T) {
	app := &App{lsusb: "false"}
	_, err := app.getUSBDevices()
	if err == nil {
		t.Error("expected error from failing command, got nil")
	}
}

func TestDeviceConnected_Present(t *testing.T) {
	app := &App{config: Config{DeviceID: "1234:5678"}}
	output := "Bus 001 Device 001: ID 1d6b:0002 Linux Foundation\nBus 001 Device 002: ID 1234:5678 Test Device\n"
	if !app.deviceConnected(output) {
		t.Error("expected device to be detected as connected")
	}
}

func TestDeviceConnected_Absent(t *testing.T) {
	app := &App{config: Config{DeviceID: "1234:5678"}}
	output := "Bus 001 Device 001: ID 1d6b:0002 Linux Foundation\n"
	if app.deviceConnected(output) {
		t.Error("expected device to be detected as absent")
	}
}

func TestDeviceConnected_Empty(t *testing.T) {
	app := &App{config: Config{DeviceID: "1234:5678"}}
	if app.deviceConnected("") {
		t.Error("expected empty output to report device absent")
	}
}

func TestHandleEvent_Connect(t *testing.T) {
	lsusb := writeTempScript(t, "Bus 001 Device 002: ID 1234:5678 Test Device\n")
	app := &App{
		lsusb:   lsusb,
		ddcutil: "echo",
		config: Config{
			DeviceID: "1234:5678",
			Monitors: []Monitor{{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"}},
		},
	}
	newState := app.handleEvent(false)
	if !newState {
		t.Error("expected connected state after connect event")
	}
}

func TestHandleEvent_Disconnect(t *testing.T) {
	lsusb := writeTempScript(t, "Bus 001 Device 001: ID 1d6b:0002 Linux Foundation\n")
	app := &App{
		lsusb:   lsusb,
		ddcutil: "echo",
		config: Config{
			DeviceID: "1234:5678",
			Monitors: []Monitor{{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"}},
		},
	}
	newState := app.handleEvent(true)
	if newState {
		t.Error("expected disconnected state after remove event")
	}
}

func TestHandleEvent_NoChange(t *testing.T) {
	lsusb := writeTempScript(t, "Bus 001 Device 002: ID 1234:5678 Test Device\n")
	app := &App{
		lsusb:   lsusb,
		ddcutil: "echo",
		config:  Config{DeviceID: "1234:5678"},
	}
	newState := app.handleEvent(true)
	if !newState {
		t.Error("expected state to remain connected when no change")
	}
}

func TestHandleEvent_LsusbError(t *testing.T) {
	app := &App{
		lsusb:  "false",
		config: Config{DeviceID: "1234:5678"},
	}
	newState := app.handleEvent(true)
	if !newState {
		t.Error("expected last known state to be preserved on lsusb error")
	}
}

func TestChangeMonitorInput_Connect(t *testing.T) {
	app := &App{
		ddcutil: "echo",
		config: Config{
			Monitors: []Monitor{
				{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
				{ConnectInputCode: "0x21", DisconnectInputCode: "0x22", DisplayBusNumber: "5", VcpInputSourceCode: "0x60"},
			},
		},
	}
	app.changeMonitorInput(true)
}

func TestChangeMonitorInput_Disconnect(t *testing.T) {
	app := &App{
		ddcutil: "echo",
		config: Config{
			Monitors: []Monitor{
				{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
			},
		},
	}
	app.changeMonitorInput(false)
}

func TestNewApp(t *testing.T) {
	app, err := newApp(writeTempConfig(t, Config{
		DeviceID: "1234:5678",
		Monitors: []Monitor{
			{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
		},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ddcutil == "" {
		t.Error("ddcutil path should not be empty")
	}
	if app.lsusb == "" {
		t.Error("lsusb path should not be empty")
	}
	if app.udevadm == "" {
		t.Error("udevadm path should not be empty")
	}
	if app.config.DeviceID != "1234:5678" {
		t.Errorf("DeviceID: got %q, want %q", app.config.DeviceID, "1234:5678")
	}
}

func TestNewApp_MissingConfig(t *testing.T) {
	_, err := newApp("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for missing config, got nil")
	}
}
