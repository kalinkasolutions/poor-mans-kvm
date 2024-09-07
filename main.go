package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	DeviceID             string
	ConnectInputCode     string
	DisconnectInputCode  string
	DisplayBusNumbers    []string
	UsbPollingIntervalMs int32
	VcpInputSourceCode   string
}

var config Config
var ddcUtilLocation string
var xsetLocation string
var lsusbLocation string

func main() {

	logMessage("Starting poor man's kvm\n")

	var configLocation string
	flag.StringVar(&configLocation, "configLocation", "default", "config location")
	flag.Parse()

	loadConfig(configLocation)

	ddcUtilLocation = executeCommand("which", "ddcutil")
	xsetLocation = executeCommand("which", "xset")
	lsusbLocation = executeCommand("which", "lsusb")

	var lastState string

	for {
		currentState, err := getUSBDevices()
		if err != nil {
			logMessage("Error getting USB devices: %v", err)
			continue
		}

		if lastState != "" {
			if strings.Contains(currentState, config.DeviceID) && !strings.Contains(lastState, config.DeviceID) {
				changeMonitorInput(config.ConnectInputCode)
			} else if !strings.Contains(currentState, config.DeviceID) && strings.Contains(lastState, config.DeviceID) {
				changeMonitorInput(config.DisconnectInputCode)
			}
		}

		lastState = currentState
		time.Sleep(time.Duration(config.UsbPollingIntervalMs) * time.Millisecond)
	}
}

func getUSBDevices() (string, error) {
	cmd := exec.Command(lsusbLocation)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func changeMonitorInput(inputCode string) {

	if inputCode == config.ConnectInputCode {
		executeCommand(xsetLocation, "dpms", "force", "on")
	}

	for _, display := range config.DisplayBusNumbers {
		executeCommand(ddcUtilLocation, "--bus", display, "setvcp", config.VcpInputSourceCode, inputCode)
		logMessage("Switching monitor: %s to input: %s\n", display, inputCode)
	}
}

func executeCommand(name string, arg ...string) string {
	logMessage("executing: %s %s\n", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logMessage("Failed to execute: %v\n", err.Error())
		logMessage("Error details: %s\n", string(output)) // Log stderr and stdout
		return ""
	}
	return strings.TrimSpace(string(output))

}

func loadConfig(configLocation string) {

	logMessage("loading config from %s", configLocation)

	data, err := os.ReadFile(configLocation)
	if err != nil {
		logMessage("error reading config file: %v", err)
		os.Exit(1)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		logMessage("error unmarshalling config file: %v", err)
		os.Exit(1)
	}
	logMessage("config loaded: %+v", config)
}

func logMessage(log string, args ...any) {
	fmt.Printf("%s - %s\n", getTimeStamp(), fmt.Sprintf(log, args...))
}

func getTimeStamp() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}
