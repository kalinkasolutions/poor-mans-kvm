package main

import (
	"bufio"
	"os/exec"
	"strings"
)

type App struct {
	config  Config
	ddcutil string
	lsusb   string
	udevadm string
}

func (a *App) watch() error {
	output, _ := a.getUSBDevices()
	lastConnected := a.deviceConnected(output)

	cmd := exec.Command(a.udevadm, "monitor", "--udev", "--subsystem-match=usb")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "UDEV") {
			lastConnected = a.handleEvent(lastConnected)
		}
	}

	return cmd.Wait()
}

func (a *App) handleEvent(lastConnected bool) bool {
	output, err := a.getUSBDevices()
	if err != nil {
		logMessage("Error getting USB devices: %v", err)
		return lastConnected
	}
	connected := a.deviceConnected(output)
	if connected != lastConnected {
		a.changeMonitorInput(connected)
	}
	return connected
}

func (a *App) deviceConnected(usbOutput string) bool {
	for _, line := range strings.Split(usbOutput, "\n") {
		if strings.Contains(line, a.config.DeviceID) {
			return true
		}
	}
	return false
}

func (a *App) getUSBDevices() (string, error) {
	cmd := exec.Command(a.lsusb)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (a *App) changeMonitorInput(connect bool) {
	for _, monitor := range a.config.Monitors {
		inputCode := monitor.ConnectInputCode
		if !connect {
			inputCode = monitor.DisconnectInputCode
		}
		go runCommand(a.ddcutil, "--bus", monitor.DisplayBusNumber, "setvcp", monitor.VcpInputSourceCode, inputCode)
		logMessage("Switching monitor: %s to input: %s", monitor.DisplayBusNumber, inputCode)
	}
}

func runCommand(name string, arg ...string) {
	logMessage("executing: %s %s", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)

	if output, err := cmd.CombinedOutput(); err != nil {
		logMessage("Failed to execute: %v", err.Error())
		logMessage("Error details: %s", string(output))
	}
}
