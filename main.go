package main

import (
	"os"

	"project-migration/cli"
	"project-migration/logger"
	"project-migration/migration" // <-- new migration package import
)

// ----- Constants -----
const (
	roadmap          = "273"
	roadmapProjectID = "PVT_kwDOAHNM9M4ABvWx"
)

func main() {
	// ...existing flag parsing...
	opts := cli.ParseFlags()

	// Validate required parameters.
	if opts.Project == "" || opts.Type == "" || opts.Name == "" {
		logger.Logger.Error("Missing required flags. Exiting")
		os.Exit(1)
	}
	if opts.Type != "team" && opts.Type != "sig" && opts.Type != "wg" {
		logger.Logger.Error("Type must be either team, sig or wg. Exiting")
		os.Exit(1)
	}

	// Retrieve project existence using migration package.
	exists, err := migration.ProjectExists(opts)
	if err != nil || !exists {
		logger.Logger.Error("Project does not exist", "err", err)
		os.Exit(1)
	}

	// Retrieve field responses for the source project and the roadmap.
	projectFields, roadmapFields, err := migration.RetrieveFieldResponses(opts.Project, roadmap)
	if err != nil {
		logger.Logger.Error("Error retrieving field responses", "err", err)
		os.Exit(1)
	}

	// Validate fields using migration package function.
	if err := migration.ValidateFields(projectFields, roadmapFields, opts); err != nil {
		logger.Logger.Error("Field validation failed", "err", err)
		os.Exit(1)
	}

	// Migrate items using migration package function.
	if err := migration.MigrateItems(opts, roadmap, roadmapProjectID, roadmapFields); err != nil {
		logger.Logger.Error("Migration failed", "err", err)
		os.Exit(1)
	}
}
