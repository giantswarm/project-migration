package main

import (
	"os"

	"log/slog"

	"giantswarm.io/project-migration/cli"
	"giantswarm.io/project-migration/github"
	"giantswarm.io/project-migration/log"       // new import for log package
	"giantswarm.io/project-migration/migration" // new import for migration package
	// ...existing import...
)

var verbose bool

func main() {
	// Set our custom handler from the new log package.
	slog.SetDefault(slog.New(log.NewHandler()))
	cfg := cli.Parse()
	verbose = cfg.Verbose

	// Create a GitHub client from the new package.
	ghClient := github.NewClient(verbose)
	if err := migration.Run(cfg, ghClient); err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
}
