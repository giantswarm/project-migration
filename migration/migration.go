package migration

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"log/slog"

	"giantswarm.io/project-migration/cli"
	"giantswarm.io/project-migration/github"
	"giantswarm.io/project-migration/types"
)

// local type to hold field list responses.
type fieldsResponse struct {
	Fields []types.Field `json:"fields"`
}

const (
	roadmap    = "273"
	roadmapPID = "PVT_kwDOAHNM9M4ABvWx"
)

// Run executes the migration.
func Run(cfg *cli.Config, gh github.Client) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	projects, err := fetchProjects(gh)
	if err != nil {
		return err
	}

	req, err := strconv.Atoi(cfg.Project)
	if err != nil {
		return fmt.Errorf("invalid project number: %s", cfg.Project)
	}
	if !projectExists(projects, req) {
		return fmt.Errorf("project '%s' not found. Exiting", cfg.Project)
	}

	projFields, err := fetchFields(gh, cfg.Project)
	if err != nil {
		return err
	}
	roadFields, err := fetchFields(gh, roadmap)
	if err != nil {
		return err
	}
	if err := validateRequiredFields(projFields.Fields, roadFields.Fields); err != nil {
		return err
	}

	typeOptionID, err := getTypeOptionID(cfg, roadFields.Fields)
	if err != nil {
		return err
	}

	return processItems(cfg, gh, roadFields.Fields, typeOptionID)
}

// validateConfig checks that required config flags are provided and valid.
func validateConfig(cfg *cli.Config) error {
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
	return nil
}

// fetchProjects retrieves and unmarshals projects from GitHub.
func fetchProjects(gh github.Client) ([]types.Project, error) {
	projectsJSON, err := gh.ListProjects()
	if err != nil {
		return nil, err
	}
	var projects []types.Project
	if err := json.Unmarshal([]byte(projectsJSON), &projects); err != nil {
		var projResp struct {
			Projects []types.Project `json:"projects"`
		}
		if err2 := json.Unmarshal([]byte(projectsJSON), &projResp); err2 == nil {
			projects = projResp.Projects
		} else {
			return nil, fmt.Errorf("error parsing projects: %v; error2: %v", err, err2)
		}
	}
	slog.Info("Retrieved projects", "projects", projects)
	return projects, nil
}

// projectExists checks that requested project number exists.
func projectExists(projects []types.Project, req int) bool {
	for _, p := range projects {
		if p.Number == req {
			return true
		}
	}
	return false
}

// fetchFields retrieves fields for a given project.
func fetchFields(gh github.Client, project string) (fieldsResponse, error) {
	fieldsJSON, err := gh.FieldList(project)
	if err != nil {
		return fieldsResponse{}, err
	}
	var fields fieldsResponse
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		// Return an empty fieldsResponse instead of 'fields'
		return fieldsResponse{}, fmt.Errorf("error parsing fields for project %s: %v", project, err)
	}
	return fields, nil
}

// findField returns the pointer to a field with the given name.
func findField(fields []types.Field, name string) *types.Field {
	for i, f := range fields {
		if f.Name == name {
			return &fields[i]
		}
	}
	return nil
}

// validateRequiredFields ensures required fields exist in both project and roadmap.
func validateRequiredFields(projFields, roadFields []types.Field) error {
	var missing []string
	projStatus := findField(projFields, "Status")
	roadStatus := findField(roadFields, "Status")
	projKind := findField(projFields, "Kind")
	roadKind := findField(roadFields, "Kind")
	projWorkstream := findField(projFields, "Workstream")
	roadWorkstream := findField(roadFields, "Workstream")

	if projStatus == nil {
		missing = append(missing, "Status in project")
	}
	if roadStatus == nil {
		missing = append(missing, "Status in roadmap")
	}
	if projKind == nil {
		missing = append(missing, "Kind in project")
	}
	if roadKind == nil {
		missing = append(missing, "Kind in roadmap")
	}
	if projWorkstream == nil {
		missing = append(missing, "Workstream in project")
	}
	if roadWorkstream == nil {
		missing = append(missing, "Workstream in roadmap")
	}
	if len(missing) > 0 {
		slog.Error("Required fields missing", "fields", missing)
		return fmt.Errorf("required fields missing in project or roadmap")
	}
	return nil
}

// getTypeOptionID extracts the matching option ID for the given type.
func getTypeOptionID(cfg *cli.Config, roadFields []types.Field) (string, error) {
	var typeOptionID string
	var field *types.Field
	switch cfg.Type {
	case "team":
		field = findField(roadFields, "Team")
		if field == nil {
			return "", fmt.Errorf("team field not found in roadmap")
		}
	case "sig":
		field = findField(roadFields, "SIG")
		if field == nil {
			return "", fmt.Errorf("SIG field not found in roadmap")
		}
	case "wg":
		field = findField(roadFields, "Working Group")
		if field == nil {
			return "", fmt.Errorf("working group field not found in roadmap")
		}
	}
	for _, o := range field.Options {
		if strings.HasPrefix(o.Name, cfg.Name) {
			typeOptionID = o.ID
			break
		}
	}
	if typeOptionID == "" {
		return "", fmt.Errorf("%s '%s' not found in roadmap", strings.ToUpper(cfg.Type), cfg.Name)
	}
	return typeOptionID, nil
}

// processItems handles the migration of items.
func processItems(cfg *cli.Config, gh github.Client, roadFields []types.Field, typeOptionID string) error {
	itemsJSON, err := gh.ListItems(cfg.Project)
	if err != nil {
		return err
	}
	var items struct {
		Items []types.Item `json:"items"`
	}
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
		return fmt.Errorf("error parsing items: %v", err)
	}

	// Extract common roadmap fields needed for editing.
	roadStatus := findField(roadFields, "Status")
	roadKind := findField(roadFields, "Kind")
	roadWorkstream := findField(roadFields, "Workstream")
	roadStartDate := findField(roadFields, "Start Date")
	roadTargetDate := findField(roadFields, "Target Date")

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

		// Edit item based on type.
		switch cfg.Type {
		case "team":
			_, err = gh.EditItem(roadmapPID, newItem.ID, findField(roadFields, "Team").ID, typeOptionID)
		case "sig":
			_, err = gh.EditItem(roadmapPID, newItem.ID, findField(roadFields, "SIG").ID, typeOptionID)
		case "wg":
			_, err = gh.EditItem(roadmapPID, newItem.ID, findField(roadFields, "Working Group").ID, typeOptionID)
		}
		if err != nil {
			slog.Error("Error editing type field", "item", newItem.ID, "error", err)
		}
		// Optional fields.
		if cfg.Area != "" {
			var areaOptionID string
			for _, o := range findField(roadFields, "Area").Options {
				if strings.HasPrefix(o.Name, cfg.Area) {
					areaOptionID = o.ID
					break
				}
			}
			if areaOptionID == "" {
				slog.Error("Area not found", "area", cfg.Area)
			} else {
				_, err = gh.EditItem(roadmapPID, newItem.ID, findField(roadFields, "Area").ID, areaOptionID)
				if err != nil {
					slog.Error("Error editing area", "item", newItem.ID, "error", err)
				}
			}
		}
		if cfg.Function != "" {
			var funcOptionID string
			for _, o := range findField(roadFields, "Function").Options {
				if strings.HasPrefix(o.Name, cfg.Function) {
					funcOptionID = o.ID
					break
				}
			}
			if funcOptionID == "" {
				slog.Error("Function not found", "function", cfg.Function)
			} else {
				_, err = gh.EditItem(roadmapPID, newItem.ID, findField(roadFields, "Function").ID, funcOptionID)
				if err != nil {
					slog.Error("Error editing function", "item", newItem.ID, "error", err)
				}
			}
		}
		// Edit status, kind and workstream if available.
		if s, ok := item.Status.(string); ok && s != "" && roadStatus != nil {
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
		if k, ok := item.Kind.(string); ok && k != "" && roadKind != nil {
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
		if ws, ok := item.Workstream.(string); ok && ws != "" && roadWorkstream != nil {
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
		// Edit dates if provided.
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
