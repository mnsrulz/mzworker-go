package main

import (
	"log"

	"github.com/mnsrulz/mzworker-go/cmd"
)

var version = "dev"

func main() {
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
