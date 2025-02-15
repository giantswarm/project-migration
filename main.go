package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"project-migration/cli"
	"project-migration/types"
)

// ----- Constants -----
const (
	roadmap          = "273"
	roadmapProjectID = "PVT_kwDOAHNM9M4ABvWx"
	ghOwnerFlags     = "--owner giantswarm -L 10000 --format json"
)

// runGh runs a GitHub CLI command with provided arguments and returns its output.
func runGh(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running gh %v: %w", args, err)
	}
	return out.String(), nil
}

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
		fmt.Println("Project number is missing. Exiting")
		os.Exit(1)
	}
	if opts.Type == "" {
		fmt.Println("Type is missing. Exiting")
		os.Exit(1)
	}
	if opts.Name == "" {
		fmt.Println("Name is missing. Exiting")
		os.Exit(1)
	}
	if opts.Type != "team" && opts.Type != "sig" && opts.Type != "wg" {
		fmt.Println("Type must be either team, sig or wg. Exiting")
		os.Exit(1)
	}

	// ----- Retrieve project details -----
	// Call "gh project list" and find the specific project by number.
	out, err := runGh(append([]string{"project", "list"}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var projList types.ProjectList
	if err := json.Unmarshal([]byte(out), &projList); err != nil {
		fmt.Printf("Error parsing project list: %v\n", err)
		os.Exit(1)
	}
	var sourceProject *types.Project
	for _, p := range projList.Projects {
		// Convert project number to string comparison if necessary.
		if fmt.Sprintf("%d", p.Number) == opts.Project {
			sourceProject = &p
			break
		}
	}
	if sourceProject == nil {
		fmt.Printf("Project '%s' not found. Exiting\n", opts.Project)
		os.Exit(1)
	}

	// ----- Retrieve field details for both source project and roadmap -----
	projectFieldsOut, err := runGh(append([]string{"project", "field-list", opts.Project}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var projectFields types.FieldResponse
	if err := json.Unmarshal([]byte(projectFieldsOut), &projectFields); err != nil {
		fmt.Printf("Error parsing project fields: %v\n", err)
		os.Exit(1)
	}

	roadmapFieldsOut, err := runGh(append([]string{"project", "field-list", roadmap}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var roadmapFields types.FieldResponse
	if err := json.Unmarshal([]byte(roadmapFieldsOut), &roadmapFields); err != nil {
		fmt.Printf("Error parsing roadmap fields: %v\n", err)
		os.Exit(1)
	}

	// ----- Validate that required field options exist -----
	required := []string{"Status", "Kind", "Workstream"}
	fieldAbort := false
	for _, fieldName := range required {
		projField := findField(projectFields.Fields, fieldName)
		roadField := findField(roadmapFields.Fields, fieldName)
		if projField == nil {
			fmt.Printf("%s field missing in project\n", fieldName)
			fieldAbort = true
			continue
		}
		if roadField == nil {
			fmt.Printf("%s field missing in roadmap\n", fieldName)
			fieldAbort = true
			continue
		}
		// For each option in the project field, check if it exists in the roadmap field.
		for _, projOpt := range projField.Options {
			if findOptionByName(roadField, projOpt.Name) == nil {
				fmt.Printf("Project's %s %s doesn't exist in roadmap\n", fieldName, projOpt.Name)
				fieldAbort = true
			}
		}
	}
	// Validate type (team, sig, wg) option existence in roadmap.
	switch opts.Type {
	case "team":
		teamField := findField(roadmapFields.Fields, "Team")
		if teamField == nil || findOptionByPrefix(teamField, opts.Name) == nil {
			fmt.Printf("Team '%s' not found in roadmap\n", opts.Name)
			fieldAbort = true
		}
	case "sig":
		sigField := findField(roadmapFields.Fields, "SIG")
		if sigField == nil || findOptionByPrefix(sigField, opts.Name) == nil {
			fmt.Printf("SIG '%s' not found in roadmap\n", opts.Name)
			fieldAbort = true
		}
	case "wg":
		wgField := findField(roadmapFields.Fields, "Working Group")
		if wgField == nil || findOptionByPrefix(wgField, opts.Name) == nil {
			fmt.Printf("WG '%s' not found in roadmap\n", opts.Name)
			fieldAbort = true
		}
	}
	// Validate optional fields.
	if opts.Area != "" {
		areaField := findField(roadmapFields.Fields, "Area")
		if areaField == nil || findOptionByPrefix(areaField, opts.Area) == nil {
			fmt.Printf("Area '%s' not found in roadmap\n", opts.Area)
			fieldAbort = true
		}
	}
	if opts.Function != "" {
		funcField := findField(roadmapFields.Fields, "Function")
		if funcField == nil || findOptionByPrefix(funcField, opts.Function) == nil {
			fmt.Printf("Function '%s' not found in roadmap\n", opts.Function)
			fieldAbort = true
		}
	}
	if fieldAbort {
		fmt.Println("There are fields in the project board that are not in the roadmap board. Exiting")
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
	itemsOut, err := runGh(append([]string{"project", "item-list", opts.Project}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var itemList types.ItemList
	if err := json.Unmarshal([]byte(itemsOut), &itemList); err != nil {
		fmt.Printf("Error parsing item list: %v\n", err)
		os.Exit(1)
	}

	// ----- Process each item (migrate item) -----
	for _, item := range itemList.Items {
		// Skip items of type DraftIssue.
		if item.Content.Type == "DraftIssue" {
			fmt.Printf("Skipping draft: %s\n", item.Content.Title)
			continue
		}
		fmt.Printf("Adding issue '%s' to roadmap board\n", item.Title)
		// Add the item to the roadmap.
		addOut, err := runGh("project", "item-add", roadmap, "--owner", "giantswarm", "--format", "json", "--url", item.Content.URL)
		if err != nil {
			fmt.Printf("Error adding item: %v\n", err)
			continue
		}
		var added struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(addOut), &added); err != nil {
			fmt.Printf("Error parsing new item: %v\n", err)
			continue
		}

		// ----- Edit type field based on provided type -----
		switch opts.Type {
		case "team":
			teamField := findField(roadmapFields.Fields, "Team")
			opt := findOptionByPrefix(teamField, opts.Name)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadTeamID, "--single-select-option-id", opt.ID)
			}
		case "sig":
			sigField := findField(roadmapFields.Fields, "SIG")
			opt := findOptionByPrefix(sigField, opts.Name)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadSigID, "--single-select-option-id", opt.ID)
			}
		case "wg":
			wgField := findField(roadmapFields.Fields, "Working Group")
			opt := findOptionByPrefix(wgField, opts.Name)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadWGID, "--single-select-option-id", opt.ID)
			}
		}
		if err != nil {
			fmt.Printf("Error editing type field for item %s: %v\n", added.ID, err)
		}

		// ----- Optional fields: Area and Function -----
		if opts.Area != "" {
			areaField := findField(roadmapFields.Fields, "Area")
			if opt := findOptionByPrefix(areaField, opts.Area); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadAreaID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing area field for item %s: %v\n", added.ID, err)
				}
			}
		}
		if opts.Function != "" {
			funcField := findField(roadmapFields.Fields, "Function")
			if opt := findOptionByPrefix(funcField, opts.Function); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadFunctionID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing function field for item %s: %v\n", added.ID, err)
				}
			}
		}

		// ----- Update Status, Kind, and Workstream fields -----
		if item.Status != "" {
			statusField := findField(roadmapFields.Fields, "Status")
			if opt := findOptionByName(statusField, item.Status); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadStatusID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing status for item %s: %v\n", added.ID, err)
				}
			} else {
				fmt.Printf("Status '%s' not found in roadmap.\n", item.Status)
			}
		}
		if item.Kind != "" {
			kindField := findField(roadmapFields.Fields, "Kind")
			if opt := findOptionByName(kindField, item.Kind); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadKindID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing kind for item %s: %v\n", added.ID, err)
				}
			} else {
				fmt.Printf("Kind '%s' not found in roadmap.\n", item.Kind)
			}
		}
		if item.Workstream != "" {
			worksField := findField(roadmapFields.Fields, "Workstream")
			if opt := findOptionByName(worksField, item.Workstream); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadWorkstreamID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing workstream for item %s: %v\n", added.ID, err)
				}
			} else {
				fmt.Printf("Workstream '%s' not found in roadmap.\n", item.Workstream)
			}
		}

		// ----- Update date fields (Start Date and Target Date) -----
		if item.StartDate != "" && item.StartDate != "null" {
			_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadStartDateID, "--date", item.StartDate)
			if err != nil {
				fmt.Printf("Error editing start date for item %s: %v\n", added.ID, err)
			}
		}
		if item.TargetDate != "" && item.TargetDate != "null" {
			_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadTargetDateID, "--date", item.TargetDate)
			if err != nil {
				fmt.Printf("Error editing target date for item %s: %v\n", added.ID, err)
			}
		}

		// ----- Archive the original item if not in dry-run mode -----
		if !opts.DryRun {
			_, err = runGh("project", "item-archive", opts.Project, "--id", item.ID, "--owner", "giantswarm")
			if err != nil {
				fmt.Printf("Error archiving item %s: %v\n", item.ID, err)
			}
		}
	}
}
