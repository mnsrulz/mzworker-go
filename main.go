package main

import (
	"log"

	"github.com/mnsrulz/mzworker-go/cmd"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitSHA    = "unknown"
)

func main() {
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.GitSHA = GitSHA
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
