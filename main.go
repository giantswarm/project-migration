package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"log/slog"

	"giantswarm.io/project-migration/cli"
	"giantswarm.io/project-migration/github"
	"giantswarm.io/project-migration/log"   // new import for log package
	"giantswarm.io/project-migration/types" // ...existing import...
)

// Add missing constants.
const (
	roadmap    = "273"
	roadmapPID = "PVT_kwDOAHNM9M4ABvWx"
)

// Global verbose flag.
var verbose bool

func main() {
	// Set our custom handler from the new log package.
	slog.SetDefault(slog.New(log.NewHandler()))
	cfg := cli.Parse()
	verbose = cfg.Verbose

	// Create a GitHub client from the new package.
	ghClient := github.NewClient(verbose)
	if err := runMigration(cfg, ghClient); err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
}

// runMigration now accepts a github client.
func runMigration(cfg *cli.Config, gh github.Client) error {
	// Validate required flags.
	if cfg.Project == "" {
		return fmt.Errorf("project number is missing. Exiting")
	}
	if cfg.Type == "" {
		return fmt.Errorf("type is missing. Exiting")
	}
	if cfg.Name == "" {
		return fmt.Errorf("name is missing. Exiting")
	}
	if cfg.Type != "team" && cfg.Type != "sig" && cfg.Type != "wg" {
		return fmt.Errorf("type must be either team, sig or wg. Exiting")
	}

	// Updated projects JSON parsing logic.
	projectsJSON, err := gh.ListProjects()
	if err != nil {
		return err
	}
	var projects []types.Project // updated to use types.Project
	// Try to unmarshal directly as a slice.
	if err := json.Unmarshal([]byte(projectsJSON), &projects); err != nil {
		// Try to unmarshal into a wrapper with a "projects" field.
		var projResp struct {
			Projects []types.Project `json:"projects"`
		}
		if err2 := json.Unmarshal([]byte(projectsJSON), &projResp); err2 == nil {
			projects = projResp.Projects
		} else {
			return fmt.Errorf("error parsing projects: %v; error2: %v", err, err2)
		}
	}
	// Only log if verbose is enabled.
	if verbose {
		slog.Info("Retrieved projects", "projects", projects)
	}
	requested, err := strconv.Atoi(cfg.Project)
	if err != nil {
		return fmt.Errorf("invalid project number: %s", cfg.Project)
	}
	if verbose {
		slog.Info("Requested project number", "projectNumber", requested)
	}

	found := false
	for _, p := range projects {
		if p.Number == requested {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("project '%s' not found. Exiting", cfg.Project)
	}

	projectFieldsJSON, err := gh.FieldList(cfg.Project)
	if err != nil {
		return err
	}
	var projectFields struct {
		Fields []types.Field `json:"fields"` // updated to use types.Field
	}
	if err := json.Unmarshal([]byte(projectFieldsJSON), &projectFields); err != nil {
		return fmt.Errorf("error parsing project fields: %v", err)
	}

	roadmapFieldsJSON, err := gh.FieldList(roadmap)
	if err != nil {
		return err
	}
	var roadmapFields struct {
		Fields []types.Field `json:"fields"` // updated to use types.Field
	}
	if err := json.Unmarshal([]byte(roadmapFieldsJSON), &roadmapFields); err != nil {
		return fmt.Errorf("error parsing roadmap fields: %v", err)
	}

	findField := func(fields []types.Field, name string) *types.Field {
		for i, f := range fields {
			if f.Name == name {
				return &fields[i]
			}
		}
		return nil
	}

	projStatus := findField(projectFields.Fields, "Status")
	roadStatus := findField(roadmapFields.Fields, "Status")
	projKind := findField(projectFields.Fields, "Kind")
	roadKind := findField(roadmapFields.Fields, "Kind")
	projWorkstream := findField(projectFields.Fields, "Workstream")
	roadWorkstream := findField(roadmapFields.Fields, "Workstream")

	// Log missing required fields if any.
	var missingFields []string
	if projStatus == nil {
		missingFields = append(missingFields, "Status in project")
	}
	if roadStatus == nil {
		missingFields = append(missingFields, "Status in roadmap")
	}
	if projKind == nil {
		missingFields = append(missingFields, "Kind in project")
	}
	if roadKind == nil {
		missingFields = append(missingFields, "Kind in roadmap")
	}
	if projWorkstream == nil {
		missingFields = append(missingFields, "Workstream in project")
	}
	if roadWorkstream == nil {
		missingFields = append(missingFields, "Workstream in roadmap")
	}
	if len(missingFields) > 0 {
		slog.Error("Required fields missing", "fields", missingFields)
		return fmt.Errorf("required fields missing in project or roadmap")
	}

	roadTeam := findField(roadmapFields.Fields, "Team")
	roadSIG := findField(roadmapFields.Fields, "SIG")
	roadWG := findField(roadmapFields.Fields, "Working Group")
	roadArea := findField(roadmapFields.Fields, "Area")
	roadFunction := findField(roadmapFields.Fields, "Function")
	roadStartDate := findField(roadmapFields.Fields, "Start Date")
	roadTargetDate := findField(roadmapFields.Fields, "Target Date")

	validateOptions := func(projectField, roadmapField *types.Field) {
		for _, opt := range projectField.Options {
			foundOpt := false
			for _, ropt := range roadmapField.Options {
				if ropt.Name == opt.Name {
					foundOpt = true
					break
				}
			}
			if !foundOpt {
				slog.Info("Option not found in roadmap", "option", opt.Name, "field", roadmapField.Name)
			}
		}
	}
	validateOptions(projStatus, roadStatus)
	validateOptions(projKind, roadKind)
	validateOptions(projWorkstream, roadWorkstream)

	var typeOptionID string
	if cfg.Type == "team" && roadTeam != nil {
		for _, o := range roadTeam.Options {
			if strings.HasPrefix(o.Name, cfg.Name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("team '%s' not found in roadmap", cfg.Name)
		}
	} else if cfg.Type == "sig" && roadSIG != nil {
		for _, o := range roadSIG.Options {
			if strings.HasPrefix(o.Name, cfg.Name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("SIG '%s' not found in roadmap", cfg.Name)
		}
	} else if cfg.Type == "wg" && roadWG != nil {
		for _, o := range roadWG.Options {
			if strings.HasPrefix(o.Name, cfg.Name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("WG '%s' not found in roadmap", cfg.Name)
		}
	}

	var areaOptionID, functionOptionID string
	if cfg.Area != "" && roadArea != nil {
		for _, o := range roadArea.Options {
			if strings.HasPrefix(o.Name, cfg.Area) {
				areaOptionID = o.ID
				break
			}
		}
		if areaOptionID == "" {
			return fmt.Errorf("area '%s' not found in roadmap", cfg.Area)
		}
	}

	if cfg.Function != "" && roadFunction != nil {
		for _, o := range roadFunction.Options {
			if strings.HasPrefix(o.Name, cfg.Function) {
				functionOptionID = o.ID
				break
			}
		}
		if functionOptionID == "" {
			return fmt.Errorf("function '%s' not found in roadmap", cfg.Function)
		}
	}

	itemsJSON, err := gh.ListItems(cfg.Project)
	if err != nil {
		return err
	}
	var items struct {
		Items []types.Item `json:"items"` // updated to use types.Item
	}
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
		return fmt.Errorf("error parsing items: %v", err)
	}

	for _, item := range items.Items {
		if item.Content.Type == "DraftIssue" {
			slog.Info("Skipping draft", "title", item.Content.Title)
			continue
		}

		slog.Info("Adding issue to roadmap", "title", item.Title)
		addOut, err := gh.AddItem(roadmap, item.Content.URL)
		if err != nil {
			slog.Error("Error adding item", "title", item.Title, "error", err)
			continue
		}
		var newItem struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(addOut), &newItem); err != nil {
			slog.Error("Error parsing new item", "title", item.Title, "error", err)
			continue
		}

		switch cfg.Type {
		case "team":
			_, err = gh.EditItem(roadmapPID, newItem.ID, roadTeam.ID, typeOptionID)
		case "sig":
			_, err = gh.EditItem(roadmapPID, newItem.ID, roadSIG.ID, typeOptionID)
		case "wg":
			_, err = gh.EditItem(roadmapPID, newItem.ID, roadWG.ID, typeOptionID)
		}
		if err != nil {
			slog.Error("Error editing item", "item", newItem.ID, "error", err)
		}

		if areaOptionID != "" {
			_, err = gh.EditItem(roadmapPID, newItem.ID, roadArea.ID, areaOptionID)
			if err != nil {
				slog.Error("Error editing area", "item", newItem.ID, "error", err)
			}
		}
		if functionOptionID != "" {
			_, err = gh.EditItem(roadmapPID, newItem.ID, roadFunction.ID, functionOptionID)
			if err != nil {
				slog.Error("Error editing function", "item", newItem.ID, "error", err)
			}
		}

		if s, ok := item.Status.(string); ok && s != "" {
			var statusOptID string
			for _, o := range roadStatus.Options {
				if o.Name == s {
					statusOptID = o.ID
					break
				}
			}
			if statusOptID != "" {
				_, err = gh.EditItem(roadmapPID, newItem.ID, roadStatus.ID, statusOptID)
				if err != nil {
					slog.Error("Error editing status", "item", newItem.ID, "error", err)
				}
			} else {
				slog.Info("Status not found in roadmap", "status", s)
			}
		}

		if k, ok := item.Kind.(string); ok && k != "" {
			var kindOptID string
			for _, o := range roadKind.Options {
				if o.Name == k {
					kindOptID = o.ID
					break
				}
			}
			if kindOptID != "" {
				_, err = gh.EditItem(roadmapPID, newItem.ID, roadKind.ID, kindOptID)
				if err != nil {
					slog.Error("Error editing kind", "item", newItem.ID, "error", err)
				}
			} else {
				slog.Info("Kind not found in roadmap", "kind", k)
			}
		}

		if ws, ok := item.Workstream.(string); ok && ws != "" {
			var wsOptID string
			for _, o := range roadWorkstream.Options {
				if o.Name == ws {
					wsOptID = o.ID
					break
				}
			}
			if wsOptID != "" {
				_, err = gh.EditItem(roadmapPID, newItem.ID, roadWorkstream.ID, wsOptID)
				if err != nil {
					slog.Error("Error editing workstream", "item", newItem.ID, "error", err)
				}
			} else {
				slog.Info("Workstream not found in roadmap", "workstream", ws)
			}
		}

		if item.StartDate != "" && item.StartDate != "null" && roadStartDate != nil {
			_, err = gh.EditItemDate(roadmapPID, newItem.ID, roadStartDate.ID, item.StartDate)
			if err != nil {
				slog.Error("Error editing start date", "item", newItem.ID, "error", err)
			}
		}
		if item.TargetDate != "" && item.TargetDate != "null" && roadTargetDate != nil {
			_, err = gh.EditItemDate(roadmapPID, newItem.ID, roadTargetDate.ID, item.TargetDate)
			if err != nil {
				slog.Error("Error editing target date", "item", newItem.ID, "error", err)
			}
		}

		if !cfg.DryRun {
			_, err = gh.ArchiveItem(cfg.Project, item.ID)
			if err != nil {
				slog.Error("Error archiving item", "item", item.ID, "error", err)
			}
		}
	}
	return nil
}
