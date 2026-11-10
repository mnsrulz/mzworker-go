package main

import (
	"log/slog"
	"os"

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
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}
