package main

import (
	"encoding/json"
	"fmt"
	"os"

	"project-migration/cli"
	"project-migration/gh"
	"project-migration/logger"
	"project-migration/migration" // <-- new migration package import
	"project-migration/types"
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

	// Retrieve project details.
	out, err := gh.ListProjects()
	if err != nil {
		logger.Logger.Error("Error retrieving project list", "err", err)
		os.Exit(1)
	}
	var projList types.ProjectList
	if err := json.Unmarshal([]byte(out), &projList); err != nil {
		logger.Logger.Error("Error parsing project list", "err", err)
		os.Exit(1)
	}
	var sourceProject *types.Project
	for _, p := range projList.Projects {
		if fmt.Sprintf("%d", p.Number) == opts.Project {
			sourceProject = &p
			break
		}
	}
	if sourceProject == nil {
		logger.Logger.Error("Project not found", "project", opts.Project)
		os.Exit(1)
	}

	// Retrieve field details for the source project.
	projectFieldsOut, err := gh.GetFieldList(opts.Project)
	if err != nil {
		logger.Logger.Error("Error retrieving project fields", "err", err)
		os.Exit(1)
	}
	var projectFields types.FieldResponse
	if err := json.Unmarshal([]byte(projectFieldsOut), &projectFields); err != nil {
		logger.Logger.Error("Error parsing project fields", "err", err)
		os.Exit(1)
	}

	// Retrieve roadmap board field details.
	roadmapFieldsOut, err := gh.GetFieldList(roadmap)
	if err != nil {
		logger.Logger.Error("Error retrieving roadmap fields", "err", err)
		os.Exit(1)
	}
	var roadmapFields types.FieldResponse
	if err := json.Unmarshal([]byte(roadmapFieldsOut), &roadmapFields); err != nil {
		logger.Logger.Error("Error parsing roadmap fields", "err", err)
		os.Exit(1)
	}

	// Validate fields using migration package function.
	if err := migration.ValidateFields(&projectFields, &roadmapFields, opts); err != nil {
		logger.Logger.Error("Field validation failed", "err", err)
		os.Exit(1)
	}

	// Migrate items using migration package function.
	if err := migration.MigrateItems(opts, roadmap, roadmapProjectID, &roadmapFields); err != nil {
		logger.Logger.Error("Migration failed", "err", err)
		os.Exit(1)
	}
}
