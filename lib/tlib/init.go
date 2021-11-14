package tlib

import (
	"os"
)

// Init has to be run after settings kp flags
func Init() {
	_, err := KP.Parse(os.Args[1:])
	if err != nil {
		Error("msg", "Error parsing command line arguments", "error", err)
		KP.Usage(os.Args[1:])
		os.Exit(2)
	}

	if config.env == "prod" {
		Config.Prod = true
	}

	applyLogLevel(config.logLevel)

	Info("msg", "initialisation done", "loglevel", config.logLevel, "prod", Config.Prod)

}
