package main

import (
	"log"
	"shiroxy/internal/cli"
)

func main() {
	_, err := cli.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
