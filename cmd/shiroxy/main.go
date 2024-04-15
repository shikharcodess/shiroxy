// Shiroxy main entry point
// Author: @ShikharY10

package main

import (
	"log"
	"shiroxy/cmd/shiroxy/defaults"
	"shiroxy/pkg/cli"
	"shiroxy/pkg/logger"
)

func main() {
	// Reading Configuration
	configuration, err := cli.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Starting Logger
	logHandler, err := logger.StartLogger(configuration.Logging)
	if err != nil {
		log.Fatal(err)
	}

	// Executing Default Processes
	err = defaults.ExecuteDefault(configuration, logHandler)
	if err != nil {
		log.Fatal(err)
	}
}
