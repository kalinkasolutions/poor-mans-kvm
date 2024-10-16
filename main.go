package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Montor struct {
	ConnectInputCode    string
	DisconnectInputCode string
	DisplayBusNumber    string
	VcpInputSourceCode  string
}

type Config struct {
	DeviceID             string
	UsbPollingIntervalMs int32
	Monitors             []Montor
}

var config Config
var ddcUtilLocation string
var lsusbLocation string

func main() {

	logMessage("Starting poor man's kvm%s\n", runtime.GOOS)
	logMessage("Running as user: %s", executeCommand("whoami"))

	var configLocation string
	flag.StringVar(&configLocation, "configLocation", "./config.json", "config location")
	flag.Parse()

	loadConfig(configLocation)

	ddcUtilLocation = executeCommand("which", "ddcutil")
	if ddcUtilLocation == "" {
		logMessage("ddcutil must be installed")
		os.Exit(1)
	}

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
				changeMonitorInput(true)
			} else if !strings.Contains(currentState, config.DeviceID) && strings.Contains(lastState, config.DeviceID) {
				changeMonitorInput(false)
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

func changeMonitorInput(connect bool) {
	for _, monitor := range config.Monitors {
		inputCode := monitor.ConnectInputCode
		if !connect {
			inputCode = monitor.DisconnectInputCode
		}
		executeCommand(ddcUtilLocation, "--bus", monitor.DisplayBusNumber, "setvcp", monitor.VcpInputSourceCode, inputCode)
		logMessage("Switching monitor: %s to input: %s\n", monitor.DisplayBusNumber, inputCode)
	}
}

func executeCommand(name string, arg ...string) string {
	logMessage("executing: %s %s\n", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logMessage("Failed to execute: %v\n", err.Error())
		logMessage("Error details: %s\n", string(output))
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
