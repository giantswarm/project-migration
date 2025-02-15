package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"project-migration/cli"
	"project-migration/gh"
	"project-migration/logger" // <-- added logger package import
	"project-migration/types"
)

// ----- Constants -----
const (
	roadmap          = "273"
	roadmapProjectID = "PVT_kwDOAHNM9M4ABvWx"
)

// findField searches for a field with a given name in a slice of fields.
func findField(fields []types.Field, name string) *types.Field {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

// findOptionByName searches for an option with the provided name in a field.
func findOptionByName(field *types.Field, name string) *types.Option {
	if field == nil {
		return nil
	}
	for i := range field.Options {
		if field.Options[i].Name == name {
			return &field.Options[i]
		}
	}
	return nil
}

// findOptionByPrefix searches for an option whose name starts with the provided prefix.
func findOptionByPrefix(field *types.Field, prefix string) *types.Option {
	if field == nil {
		return nil
	}
	for i := range field.Options {
		if strings.HasPrefix(field.Options[i].Name, prefix) {
			return &field.Options[i]
		}
	}
	return nil
}

func main() {
	// ----- Parse command-line flags -----
	opts := cli.ParseFlags()

	// Validate required parameters.
	if opts.Project == "" {
		logger.Logger.Error("Project number is missing. Exiting")
		os.Exit(1)
	}
	if opts.Type == "" {
		logger.Logger.Error("Type is missing. Exiting")
		os.Exit(1)
	}
	if opts.Name == "" {
		logger.Logger.Error("Name is missing. Exiting")
		os.Exit(1)
	}
	if opts.Type != "team" && opts.Type != "sig" && opts.Type != "wg" {
		logger.Logger.Error("Type must be either team, sig or wg. Exiting")
		os.Exit(1)
	}

	// ----- Retrieve project details -----
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

	// ----- Retrieve field details for both source project and roadmap -----
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

	// ----- Validate that required field options exist -----
	required := []string{"Status", "Kind", "Workstream"}
	fieldAbort := false
	for _, fieldName := range required {
		projField := findField(projectFields.Fields, fieldName)
		roadField := findField(roadmapFields.Fields, fieldName)
		if projField == nil {
			logger.Logger.Error(fmt.Sprintf("%s field missing in project", fieldName))
			fieldAbort = true
			continue
		}
		if roadField == nil {
			logger.Logger.Error(fmt.Sprintf("%s field missing in roadmap", fieldName))
			fieldAbort = true
			continue
		}
		for _, projOpt := range projField.Options {
			if findOptionByName(roadField, projOpt.Name) == nil {
				logger.Logger.Error(fmt.Sprintf("Project's %s %s doesn't exist in roadmap", fieldName, projOpt.Name))
				fieldAbort = true
			}
		}
	}
	switch opts.Type {
	case "team":
		teamField := findField(roadmapFields.Fields, "Team")
		if teamField == nil || findOptionByPrefix(teamField, opts.Name) == nil {
			logger.Logger.Error(fmt.Sprintf("Team '%s' not found in roadmap", opts.Name))
			fieldAbort = true
		}
	case "sig":
		sigField := findField(roadmapFields.Fields, "SIG")
		if sigField == nil || findOptionByPrefix(sigField, opts.Name) == nil {
			logger.Logger.Error(fmt.Sprintf("SIG '%s' not found in roadmap", opts.Name))
			fieldAbort = true
		}
	case "wg":
		wgField := findField(roadmapFields.Fields, "Working Group")
		if wgField == nil || findOptionByPrefix(wgField, opts.Name) == nil {
			logger.Logger.Error(fmt.Sprintf("WG '%s' not found in roadmap", opts.Name))
			fieldAbort = true
		}
	}
	if opts.Area != "" {
		areaField := findField(roadmapFields.Fields, "Area")
		if areaField == nil || findOptionByPrefix(areaField, opts.Area) == nil {
			logger.Logger.Error("Area '%s' not found in roadmap", opts.Area)
			fieldAbort = true
		}
	}
	if opts.Function != "" {
		funcField := findField(roadmapFields.Fields, "Function")
		if funcField == nil || findOptionByPrefix(funcField, opts.Function) == nil {
			logger.Logger.Error("Function '%s' not found in roadmap", opts.Function)
			fieldAbort = true
		}
	}
	if fieldAbort {
		logger.Logger.Error("There are fields in the project board that are not in the roadmap board. Exiting")
		os.Exit(1)
	}

	// ----- Retrieve IDs for fields from the roadmap for later edits -----
	getFieldID := func(fieldName string) string {
		if f := findField(roadmapFields.Fields, fieldName); f != nil {
			return f.ID
		}
		return ""
	}
	roadStatusID := getFieldID("Status")
	roadKindID := getFieldID("Kind")
	roadWorkstreamID := getFieldID("Workstream")
	roadTeamID := getFieldID("Team")
	roadSigID := getFieldID("SIG")
	roadWGID := getFieldID("Working Group")
	roadAreaID := getFieldID("Area")
	roadFunctionID := getFieldID("Function")
	roadStartDateID := getFieldID("Start Date")
	roadTargetDateID := getFieldID("Target Date")

	// ----- Retrieve items (issues) from the source project -----
	itemsOut, err := gh.GetItemList(opts.Project)
	if err != nil {
		logger.Logger.Error("Error retrieving item list", "err", err)
		os.Exit(1)
	}
	var itemList types.ItemList
	if err := json.Unmarshal([]byte(itemsOut), &itemList); err != nil {
		logger.Logger.Error("Error parsing item list", "err", err)
		os.Exit(1)
	}

	// ----- Process each item (migrate item) -----
	for _, item := range itemList.Items {
		if item.Content.Type == "DraftIssue" {
			logger.Logger.Info("Skipping draft", "title", item.Content.Title)
			continue
		}
		logger.Logger.Info("Adding issue to roadmap board", "title", item.Title)
		addOut, err := gh.AddItem(roadmap, item.Content.URL)
		if err != nil {
			logger.Logger.Error("Error adding item", "err", err)
			continue
		}
		var added struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(addOut), &added); err != nil {
			logger.Logger.Error("Error parsing new item", "err", err)
			continue
		}

		switch opts.Type {
		case "team":
			teamField := findField(roadmapFields.Fields, "Team")
			opt := findOptionByPrefix(teamField, opts.Name)
			if opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadTeamID, opt.ID)
			}
		case "sig":
			sigField := findField(roadmapFields.Fields, "SIG")
			opt := findOptionByPrefix(sigField, opts.Name)
			if opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadSigID, opt.ID)
			}
		case "wg":
			wgField := findField(roadmapFields.Fields, "Working Group")
			opt := findOptionByPrefix(wgField, opts.Name)
			if opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadWGID, opt.ID)
			}
		}
		if err != nil {
			logger.Logger.Error("Error editing type field for item %s: %v\n", added.ID, err)
		}

		if opts.Area != "" {
			areaField := findField(roadmapFields.Fields, "Area")
			if opt := findOptionByPrefix(areaField, opts.Area); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadAreaID, opt.ID)
				if err != nil {
					logger.Logger.Error("Error editing area field for item %s: %v\n", added.ID, err)
				}
			}
		}
		if opts.Function != "" {
			funcField := findField(roadmapFields.Fields, "Function")
			if opt := findOptionByPrefix(funcField, opts.Function); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadFunctionID, opt.ID)
				if err != nil {
					logger.Logger.Error("Error editing function field for item %s: %v\n", added.ID, err)
				}
			}
		}

		if item.Status != "" {
			statusField := findField(roadmapFields.Fields, "Status")
			if opt := findOptionByName(statusField, item.Status); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadStatusID, opt.ID)
				if err != nil {
					logger.Logger.Error("Error editing status for item %s: %v\n", added.ID, err)
				}
			} else {
				logger.Logger.Error("Status '%s' not found in roadmap.\n", item.Status)
			}
		}
		if item.Kind != "" {
			kindField := findField(roadmapFields.Fields, "Kind")
			if opt := findOptionByName(kindField, item.Kind); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadKindID, opt.ID)
				if err != nil {
					logger.Logger.Error("Error editing kind for item %s: %v\n", added.ID, err)
				}
			} else {
				logger.Logger.Error("Kind '%s' not found in roadmap.\n", item.Kind)
			}
		}
		if item.Workstream != "" {
			worksField := findField(roadmapFields.Fields, "Workstream")
			if opt := findOptionByName(worksField, item.Workstream); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, roadWorkstreamID, opt.ID)
				if err != nil {
					logger.Logger.Error("Error editing workstream for item %s: %v\n", added.ID, err)
				}
			} else {
				logger.Logger.Error("Workstream '%s' not found in roadmap.\n", item.Workstream)
			}
		}

		if item.StartDate != "" && item.StartDate != "null" {
			_, err = gh.EditItemDate(roadmapProjectID, added.ID, roadStartDateID, item.StartDate)
			if err != nil {
				logger.Logger.Error("Error editing start date for item %s: %v\n", added.ID, err)
			}
		}
		if item.TargetDate != "" && item.TargetDate != "null" {
			_, err = gh.EditItemDate(roadmapProjectID, added.ID, roadTargetDateID, item.TargetDate)
			if err != nil {
				logger.Logger.Error("Error editing target date for item %s: %v\n", added.ID, err)
			}
		}

		if !opts.DryRun {
			_, err = gh.ArchiveItem(opts.Project, item.ID)
			if err != nil {
				logger.Logger.Error("Error archiving item", "id", item.ID, "err", err)
			}
		}
	}
}
