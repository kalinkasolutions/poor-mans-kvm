package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
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

// writeArgCaptureScript creates a script that appends its arguments (space-joined) to a temp
// file on each invocation, so tests can assert which args were passed.
func writeArgCaptureScript(t *testing.T) (scriptPath, capturePath string) {
	t.Helper()
	cap, err := os.CreateTemp("", "ddcutil-capture-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	cap.Close()
	t.Cleanup(func() { os.Remove(cap.Name()) })
	capturePath = cap.Name()

	script, err := os.CreateTemp("", "ddcutil-mock-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(script, "#!/bin/sh\necho \"$@\" >> \"%s\"\n", capturePath)
	script.Close()
	if err := os.Chmod(script.Name(), 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(script.Name()) })
	scriptPath = script.Name()
	return
}

func readCapturedLines(t *testing.T, capturePath string) []string {
	t.Helper()
	data, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

func TestChangeMonitorInput_Connect(t *testing.T) {
	ddcutil, capture := writeArgCaptureScript(t)
	app := &App{
		ddcutil: ddcutil,
		config: Config{
			Monitors: []Monitor{
				{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
				{ConnectInputCode: "0x21", DisconnectInputCode: "0x22", DisplayBusNumber: "5", VcpInputSourceCode: "0x60"},
			},
		},
	}
	app.changeMonitorInput(true)

	lines := readCapturedLines(t, capture)
	if len(lines) != 2 {
		t.Fatalf("expected 2 ddcutil invocations, got %d: %v", len(lines), lines)
	}
	if lines[0] != "--bus 3 setvcp 0x60 0x11" {
		t.Errorf("monitor 1: got %q, want \"--bus 3 setvcp 0x60 0x11\"", lines[0])
	}
	if lines[1] != "--bus 5 setvcp 0x60 0x21" {
		t.Errorf("monitor 2: got %q, want \"--bus 5 setvcp 0x60 0x21\"", lines[1])
	}
}

func TestChangeMonitorInput_Disconnect(t *testing.T) {
	ddcutil, capture := writeArgCaptureScript(t)
	app := &App{
		ddcutil: ddcutil,
		config: Config{
			Monitors: []Monitor{
				{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
			},
		},
	}
	app.changeMonitorInput(false)

	lines := readCapturedLines(t, capture)
	if len(lines) != 1 {
		t.Fatalf("expected 1 ddcutil invocation, got %d: %v", len(lines), lines)
	}
	if lines[0] != "--bus 3 setvcp 0x60 0x12" {
		t.Errorf("got %q, want \"--bus 3 setvcp 0x60 0x12\"", lines[0])
	}
}

func TestChangeMonitorInput_Parallel(t *testing.T) {
	ddcutil, capture := writeArgCaptureScript(t)
	app := &App{
		ddcutil: ddcutil,
		config: Config{
			ParallelMonitorSwitch: true,
			Monitors: []Monitor{
				{ConnectInputCode: "0x11", DisconnectInputCode: "0x12", DisplayBusNumber: "3", VcpInputSourceCode: "0x60"},
				{ConnectInputCode: "0x21", DisconnectInputCode: "0x22", DisplayBusNumber: "5", VcpInputSourceCode: "0x60"},
			},
		},
	}
	app.changeMonitorInput(true)

	// goroutines are detached; give them time to complete
	time.Sleep(200 * time.Millisecond)

	lines := readCapturedLines(t, capture)
	if len(lines) != 2 {
		t.Fatalf("expected 2 ddcutil invocations, got %d: %v", len(lines), lines)
	}
	got := map[string]bool{lines[0]: true, lines[1]: true}
	if !got["--bus 3 setvcp 0x60 0x11"] {
		t.Error("missing invocation for monitor 1: --bus 3 setvcp 0x60 0x11")
	}
	if !got["--bus 5 setvcp 0x60 0x21"] {
		t.Error("missing invocation for monitor 2: --bus 5 setvcp 0x60 0x21")
	}
}

func TestNewApp(t *testing.T) {
	if _, err := exec.LookPath("ddcutil"); err != nil {
		t.Skip("ddcutil not installed")
	}
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
