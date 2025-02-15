package migration

import (
	"encoding/json"
	"errors"
	"fmt"

	"project-migration/cli"
	"project-migration/gh"
	"project-migration/logger"
	"project-migration/types"
)

// --- Helper functions ---
func findField(fields []types.Field, name string) *types.Field {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

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

func findOptionByPrefix(field *types.Field, prefix string) *types.Option {
	if field == nil {
		return nil
	}
	for i := range field.Options {
		if len(field.Options[i].Name) >= len(prefix) && field.Options[i].Name[:len(prefix)] == prefix {
			return &field.Options[i]
		}
	}
	return nil
}

// --- New Retrieval Functions ---

// ProjectExists checks if the project exists using the command-line options.
func ProjectExists(opts cli.Options) (bool, error) {
	out, err := gh.ListProjects()
	if err != nil {
		return false, fmt.Errorf("error retrieving project list: %w", err)
	}
	var projList types.ProjectList
	if err := json.Unmarshal([]byte(out), &projList); err != nil {
		return false, fmt.Errorf("error parsing project list: %w", err)
	}
	for _, p := range projList.Projects {
		if fmt.Sprintf("%d", p.Number) == opts.Project {
			return true, nil
		}
	}
	return false, fmt.Errorf("project not found: %s", opts.Project)
}

// RetrieveFieldResponses retrieves field responses for both the source project and the roadmap board.
func RetrieveFieldResponses(project, roadmap string) (*types.FieldResponse, *types.FieldResponse, error) {
	projFieldsOut, err := gh.GetFieldList(project)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving project fields: %w", err)
	}
	var projectFields types.FieldResponse
	if err := json.Unmarshal([]byte(projFieldsOut), &projectFields); err != nil {
		return nil, nil, fmt.Errorf("error parsing project fields: %w", err)
	}

	roadmapFieldsOut, err := gh.GetFieldList(roadmap)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving roadmap fields: %w", err)
	}
	var roadmapFields types.FieldResponse
	if err := json.Unmarshal([]byte(roadmapFieldsOut), &roadmapFields); err != nil {
		return nil, nil, fmt.Errorf("error parsing roadmap fields: %w", err)
	}

	return &projectFields, &roadmapFields, nil
}

// ValidateFields checks for existence of required fields and options.
// Returns an error with all messages concatenated if any validations fail.
func ValidateFields(projectFields, roadmapFields *types.FieldResponse, opts cli.Options) error {
	var errMsg string
	required := []string{"Status", "Kind", "Workstream"}
	for _, fieldName := range required {
		projField := findField(projectFields.Fields, fieldName)
		roadField := findField(roadmapFields.Fields, fieldName)
		if projField == nil {
			errMsg += fmt.Sprintf("%s field missing in project\n", fieldName)
		}
		if roadField == nil {
			errMsg += fmt.Sprintf("%s field missing in roadmap\n", fieldName)
		}
		if projField != nil && roadField != nil {
			for _, projOpt := range projField.Options {
				if findOptionByName(roadField, projOpt.Name) == nil {
					errMsg += fmt.Sprintf("Project's %s %s doesn't exist in roadmap\n", fieldName, projOpt.Name)
				}
			}
		}
	}
	switch opts.Type {
	case "team":
		teamField := findField(roadmapFields.Fields, "Team")
		if teamField == nil || findOptionByPrefix(teamField, opts.Name) == nil {
			errMsg += fmt.Sprintf("Team '%s' not found in roadmap\n", opts.Name)
		}
	case "sig":
		sigField := findField(roadmapFields.Fields, "SIG")
		if sigField == nil || findOptionByPrefix(sigField, opts.Name) == nil {
			errMsg += fmt.Sprintf("SIG '%s' not found in roadmap\n", opts.Name)
		}
	case "wg":
		wgField := findField(roadmapFields.Fields, "Working Group")
		if wgField == nil || findOptionByPrefix(wgField, opts.Name) == nil {
			errMsg += fmt.Sprintf("WG '%s' not found in roadmap\n", opts.Name)
		}
	}
	if opts.Area != "" {
		areaField := findField(roadmapFields.Fields, "Area")
		if areaField == nil || findOptionByPrefix(areaField, opts.Area) == nil {
			errMsg += fmt.Sprintf("Area '%s' not found in roadmap\n", opts.Area)
		}
	}
	if opts.Function != "" {
		funcField := findField(roadmapFields.Fields, "Function")
		if funcField == nil || findOptionByPrefix(funcField, opts.Function) == nil {
			errMsg += fmt.Sprintf("Function '%s' not found in roadmap\n", opts.Function)
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

// MigrateItems retrieves and processes items from the source project and applies the migration.
// It uses the roadmap board number, roadmap project ID, and the roadmap fields.
func MigrateItems(opts cli.Options, roadmap, roadmapProjectID string, roadmapFields *types.FieldResponse) error {
	// Retrieve items from source project.
	itemsOut, err := gh.GetItemList(opts.Project)
	if err != nil {
		return fmt.Errorf("error retrieving item list: %w", err)
	}
	var itemList types.ItemList
	if err := json.Unmarshal([]byte(itemsOut), &itemList); err != nil {
		return fmt.Errorf("error parsing item list: %w", err)
	}

	// Process each item.
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

		// Update type-specific field.
		switch opts.Type {
		case "team":
			teamField := findField(roadmapFields.Fields, "Team")
			if opt := findOptionByPrefix(teamField, opts.Name); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, teamField.ID, opt.ID)
			}
		case "sig":
			sigField := findField(roadmapFields.Fields, "SIG")
			if opt := findOptionByPrefix(sigField, opts.Name); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, sigField.ID, opt.ID)
			}
		case "wg":
			wgField := findField(roadmapFields.Fields, "Working Group")
			if opt := findOptionByPrefix(wgField, opts.Name); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, wgField.ID, opt.ID)
			}
		}
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Error editing type field for item %s: %v", added.ID, err))
		}

		// Update optional fields.
		if opts.Area != "" {
			areaField := findField(roadmapFields.Fields, "Area")
			if opt := findOptionByPrefix(areaField, opts.Area); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, areaField.ID, opt.ID)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing area field for item %s: %v", added.ID, err))
				}
			}
		}
		if opts.Function != "" {
			funcField := findField(roadmapFields.Fields, "Function")
			if opt := findOptionByPrefix(funcField, opts.Function); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, funcField.ID, opt.ID)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing function field for item %s: %v", added.ID, err))
				}
			}
		}

		// Update remaining fields.
		if item.Status != "" {
			statusField := findField(roadmapFields.Fields, "Status")
			if opt := findOptionByName(statusField, item.Status); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, statusField.ID, opt.ID)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing status for item %s: %v", added.ID, err))
				}
			} else {
				logger.Logger.Error(fmt.Sprintf("Status '%s' not found in roadmap", item.Status))
			}
		}
		if item.Kind != "" {
			kindField := findField(roadmapFields.Fields, "Kind")
			if opt := findOptionByName(kindField, item.Kind); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, kindField.ID, opt.ID)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing kind for item %s: %v", added.ID, err))
				}
			} else {
				logger.Logger.Error(fmt.Sprintf("Kind '%s' not found in roadmap", item.Kind))
			}
		}
		if item.Workstream != "" {
			worksField := findField(roadmapFields.Fields, "Workstream")
			if opt := findOptionByName(worksField, item.Workstream); opt != nil {
				_, err = gh.EditItemSingle(roadmapProjectID, added.ID, worksField.ID, opt.ID)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing workstream for item %s: %v", added.ID, err))
				}
			} else {
				logger.Logger.Error(fmt.Sprintf("Workstream '%s' not found in roadmap", item.Workstream))
			}
		}

		if item.StartDate != "" && item.StartDate != "null" {
			startDateField := findField(roadmapFields.Fields, "Start Date")
			if startDateField != nil {
				_, err = gh.EditItemDate(roadmapProjectID, added.ID, startDateField.ID, item.StartDate)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing start date for item %s: %v", added.ID, err))
				}
			}
		}
		if item.TargetDate != "" && item.TargetDate != "null" {
			targetDateField := findField(roadmapFields.Fields, "Target Date")
			if targetDateField != nil {
				_, err = gh.EditItemDate(roadmapProjectID, added.ID, targetDateField.ID, item.TargetDate)
				if err != nil {
					logger.Logger.Error(fmt.Sprintf("Error editing target date for item %s: %v", added.ID, err))
				}
			}
		}

		if !opts.DryRun {
			_, err = gh.ArchiveItem(opts.Project, item.ID)
			if err != nil {
				logger.Logger.Error("Error archiving item", "id", item.ID, "err", err)
			}
		}
	}
	return nil
}
