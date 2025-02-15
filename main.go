package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ----- Data Types for JSON responses -----
type Project struct {
	Number int    `json:"number"`
	ID     string `json:"id"`
	// ... other fields not required for migration ...
}

type ProjectList struct {
	Projects []Project `json:"projects"`
}

type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Field struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Options []Option `json:"options"`
	// ... type, etc.
}

type FieldResponse struct {
	Fields []Field `json:"fields"`
}

type Content struct {
	URL   string `json:"url"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Item struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Content    Content `json:"content"`
	Status     string  `json:"status"`
	Kind       string  `json:"kind"`
	Workstream string  `json:"workstream"`
	StartDate  string  `json:"start Date"`
	TargetDate string  `json:"target Date"`
	// ... other fields ...
}

type ItemList struct {
	Items []Item `json:"items"`
}

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
func findField(fields []Field, name string) *Field {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

// findOptionByName searches for an option with the provided name in a field.
func findOptionByName(field *Field, name string) *Option {
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
func findOptionByPrefix(field *Field, prefix string) *Option {
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
	projectFlag := flag.String("p", "", "Project Number (eg 301)")
	dryRun := flag.Bool("d", false, "Dry run mode")
	typeFlag := flag.String("t", "", "Type (team, sig, wg)")
	nameFlag := flag.String("n", "", "Name of Team, SIG or WG")
	areaFlag := flag.String("a", "", "Area (eg KaaS)")
	functionFlag := flag.String("f", "", "Function (eg 'Product Strategy')")
	flag.Parse()

	// Validate required parameters.
	if *projectFlag == "" {
		fmt.Println("Project number is missing. Exiting")
		os.Exit(1)
	}
	if *typeFlag == "" {
		fmt.Println("Type is missing. Exiting")
		os.Exit(1)
	}
	if *nameFlag == "" {
		fmt.Println("Name is missing. Exiting")
		os.Exit(1)
	}
	if *typeFlag != "team" && *typeFlag != "sig" && *typeFlag != "wg" {
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
	var projList ProjectList
	if err := json.Unmarshal([]byte(out), &projList); err != nil {
		fmt.Printf("Error parsing project list: %v\n", err)
		os.Exit(1)
	}
	var sourceProject *Project
	for _, p := range projList.Projects {
		// Convert project number to string comparison if necessary.
		if fmt.Sprintf("%d", p.Number) == *projectFlag {
			sourceProject = &p
			break
		}
	}
	if sourceProject == nil {
		fmt.Printf("Project '%s' not found. Exiting\n", *projectFlag)
		os.Exit(1)
	}

	// ----- Retrieve field details for both source project and roadmap -----
	projectFieldsOut, err := runGh(append([]string{"project", "field-list", *projectFlag}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var projectFields FieldResponse
	if err := json.Unmarshal([]byte(projectFieldsOut), &projectFields); err != nil {
		fmt.Printf("Error parsing project fields: %v\n", err)
		os.Exit(1)
	}

	roadmapFieldsOut, err := runGh(append([]string{"project", "field-list", roadmap}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var roadmapFields FieldResponse
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
	switch *typeFlag {
	case "team":
		teamField := findField(roadmapFields.Fields, "Team")
		if teamField == nil || findOptionByPrefix(teamField, *nameFlag) == nil {
			fmt.Printf("Team '%s' not found in roadmap\n", *nameFlag)
			fieldAbort = true
		}
	case "sig":
		sigField := findField(roadmapFields.Fields, "SIG")
		if sigField == nil || findOptionByPrefix(sigField, *nameFlag) == nil {
			fmt.Printf("SIG '%s' not found in roadmap\n", *nameFlag)
			fieldAbort = true
		}
	case "wg":
		wgField := findField(roadmapFields.Fields, "Working Group")
		if wgField == nil || findOptionByPrefix(wgField, *nameFlag) == nil {
			fmt.Printf("WG '%s' not found in roadmap\n", *nameFlag)
			fieldAbort = true
		}
	}
	// Validate optional fields.
	if *areaFlag != "" {
		areaField := findField(roadmapFields.Fields, "Area")
		if areaField == nil || findOptionByPrefix(areaField, *areaFlag) == nil {
			fmt.Printf("Area '%s' not found in roadmap\n", *areaFlag)
			fieldAbort = true
		}
	}
	if *functionFlag != "" {
		funcField := findField(roadmapFields.Fields, "Function")
		if funcField == nil || findOptionByPrefix(funcField, *functionFlag) == nil {
			fmt.Printf("Function '%s' not found in roadmap\n", *functionFlag)
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
	itemsOut, err := runGh(append([]string{"project", "item-list", *projectFlag}, strings.Split(ghOwnerFlags, " ")...)...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var itemList ItemList
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
		switch *typeFlag {
		case "team":
			teamField := findField(roadmapFields.Fields, "Team")
			opt := findOptionByPrefix(teamField, *nameFlag)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadTeamID, "--single-select-option-id", opt.ID)
			}
		case "sig":
			sigField := findField(roadmapFields.Fields, "SIG")
			opt := findOptionByPrefix(sigField, *nameFlag)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadSigID, "--single-select-option-id", opt.ID)
			}
		case "wg":
			wgField := findField(roadmapFields.Fields, "Working Group")
			opt := findOptionByPrefix(wgField, *nameFlag)
			if opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadWGID, "--single-select-option-id", opt.ID)
			}
		}
		if err != nil {
			fmt.Printf("Error editing type field for item %s: %v\n", added.ID, err)
		}

		// ----- Optional fields: Area and Function -----
		if *areaFlag != "" {
			areaField := findField(roadmapFields.Fields, "Area")
			if opt := findOptionByPrefix(areaField, *areaFlag); opt != nil {
				_, err = runGh("project", "item-edit", "--project-id", roadmapProjectID, "--id", added.ID, "--field-id", roadAreaID, "--single-select-option-id", opt.ID)
				if err != nil {
					fmt.Printf("Error editing area field for item %s: %v\n", added.ID, err)
				}
			}
		}
		if *functionFlag != "" {
			funcField := findField(roadmapFields.Fields, "Function")
			if opt := findOptionByPrefix(funcField, *functionFlag); opt != nil {
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
		if !*dryRun {
			_, err = runGh("project", "item-archive", *projectFlag, "--id", item.ID, "--owner", "giantswarm")
			if err != nil {
				fmt.Printf("Error archiving item %s: %v\n", item.ID, err)
			}
		}
	}
}
