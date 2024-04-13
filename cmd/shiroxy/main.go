package main

import (
	"log"
	"shiroxy/cmd/shiroxy/defaults"
	"shiroxy/pkg/cli"
)

func main() {
	configuration, err := cli.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Executing default processes
	defaults.ExecuteDefault(configuration)
}
