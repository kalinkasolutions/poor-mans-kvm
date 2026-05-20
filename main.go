package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

func main() {
	if err := run(); err != nil {
		logMessage("%v", err)
		os.Exit(1)
	}
}

func run() error {
	logMessage("Starting poor man's kvm %s", runtime.GOOS)
	if u, err := user.Current(); err == nil {
		logMessage("Running as user: %s", u.Username)
	}

	var configLocation string
	flag.StringVar(&configLocation, "configLocation", "./config.json", "config location")
	flag.Parse()

	app, err := newApp(configLocation)
	if err != nil {
		return err
	}

	return app.watch()
}

func newApp(configLocation string) (*App, error) {
	app := &App{}

	if err := loadConfig(configLocation, &app.config); err != nil {
		return nil, err
	}

	var err error
	app.ddcutil, err = exec.LookPath("ddcutil")
	if err != nil {
		return nil, fmt.Errorf("ddcutil must be installed")
	}

	app.lsusb, err = exec.LookPath("lsusb")
	if err != nil {
		return nil, fmt.Errorf("lsusb must be installed")
	}

	app.udevadm, err = exec.LookPath("udevadm")
	if err != nil {
		return nil, fmt.Errorf("udevadm must be installed")
	}

	return app, nil
}
