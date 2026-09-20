package main

import (
	"log"

	"github.com/mnsrulz/mzworker-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
