package main

import (
	"log"

	"github.com/mnsrulz/mzworker-go/cmd"
)

var Version = "dev"

func main() {
	cmd.Version = Version
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
