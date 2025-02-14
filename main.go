package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Project struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

type Field struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Options []FieldOption `json:"options"`
}

type FieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Status     interface{} `json:"status"`
	Kind       interface{} `json:"kind"`
	Workstream interface{} `json:"workstream"`
	StartDate  string      `json:"start Date"`
	TargetDate string      `json:"target Date"`
	Content    struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"content"`
}

var (
	project   = flag.String("p", "", "Project Number (eg 301)")
	dryRun    = flag.Bool("d", false, "Dry run")
	typ       = flag.String("t", "", "Type (eg 'team, sig, wg')")
	name      = flag.String("n", "", "Name of Team, SIG or WG (eg Rocket)")
	area      = flag.String("a", "", "Area (eg KaaS)")
	functionF = flag.String("f", "", "Function (eg 'Product Strategy')")
)

const (
	roadmap     = "273"
	roadmapPID  = "PVT_kwDOAHNM9M4ABvWx"
	appendFlags = "--owner giantswarm -L 10000 --format json"
)

func usage() {
	fmt.Println("Usage:")
	flag.PrintDefaults()
	os.Exit(0)
}

var runCmd func(cmdStr string, args ...string) string

func init() {
	runCmd = func(cmdStr string, args ...string) string {
		cmd := exec.Command(cmdStr, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			log.Fatalf("Error executing %s: %v", cmdStr, err)
		}
		return out.String()
	}
}

// Replace main() with a new version that calls runMigration().
func main() {
	flag.Usage = usage
	flag.Parse()
	if err := runMigration(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

// runMigration contains the migration logic and returns an error on failure.
func runMigration() error {
	// Validate required flags.
	if *project == "" {
		return fmt.Errorf("project number is missing. Exiting")
	}
	if *typ == "" {
		return fmt.Errorf("type is missing. Exiting")
	}
	if *name == "" {
		return fmt.Errorf("name is missing. Exiting")
	}
	if *typ != "team" && *typ != "sig" && *typ != "wg" {
		return fmt.Errorf("type must be either team, sig or wg. Exiting")
	}

	projectsJSON := runCmd("gh", "project", "list", appendFlags)
	var projects []Project
	if err := json.Unmarshal([]byte(projectsJSON), &projects); err != nil {
		return fmt.Errorf("error parsing projects: %v", err)
	}

	found := false
	for _, p := range projects {
		if fmt.Sprintf("%d", p.Number) == *project {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("project '%s' not found. Exiting", *project)
	}

	projectFieldsJSON := runCmd("gh", "project", "field-list", *project, appendFlags)
	var projectFields struct {
		Fields []Field `json:"fields"`
	}
	if err := json.Unmarshal([]byte(projectFieldsJSON), &projectFields); err != nil {
		return fmt.Errorf("error parsing project fields: %v", err)
	}

	roadmapFieldsJSON := runCmd("gh", "project", "field-list", roadmap, appendFlags)
	var roadmapFields struct {
		Fields []Field `json:"fields"`
	}
	if err := json.Unmarshal([]byte(roadmapFieldsJSON), &roadmapFields); err != nil {
		return fmt.Errorf("error parsing roadmap fields: %v", err)
	}

	findField := func(fields []Field, name string) *Field {
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

	if projStatus == nil || roadStatus == nil || projKind == nil || roadKind == nil || projWorkstream == nil || roadWorkstream == nil {
		return fmt.Errorf("required fields missing in project or roadmap")
	}

	roadTeam := findField(roadmapFields.Fields, "Team")
	roadSIG := findField(roadmapFields.Fields, "SIG")
	roadWG := findField(roadmapFields.Fields, "Working Group")
	roadArea := findField(roadmapFields.Fields, "Area")
	roadFunction := findField(roadmapFields.Fields, "Function")
	roadStartDate := findField(roadmapFields.Fields, "Start Date")
	roadTargetDate := findField(roadmapFields.Fields, "Target Date")

	validateOptions := func(projectField, roadmapField *Field) {
		for _, opt := range projectField.Options {
			foundOpt := false
			for _, ropt := range roadmapField.Options {
				if ropt.Name == opt.Name {
					foundOpt = true
					break
				}
			}
			if !foundOpt {
				log.Printf("'%s' not found in roadmap %s", opt.Name, roadmapField.Name)
			}
		}
	}
	validateOptions(projStatus, roadStatus)
	validateOptions(projKind, roadKind)
	validateOptions(projWorkstream, roadWorkstream)

	var typeOptionID string
	if *typ == "team" && roadTeam != nil {
		for _, o := range roadTeam.Options {
			if strings.HasPrefix(o.Name, *name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("team '%s' not found in roadmap", *name)
		}
	} else if *typ == "sig" && roadSIG != nil {
		for _, o := range roadSIG.Options {
			if strings.HasPrefix(o.Name, *name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("SIG '%s' not found in roadmap", *name)
		}
	} else if *typ == "wg" && roadWG != nil {
		for _, o := range roadWG.Options {
			if strings.HasPrefix(o.Name, *name) {
				typeOptionID = o.ID
				break
			}
		}
		if typeOptionID == "" {
			return fmt.Errorf("WG '%s' not found in roadmap", *name)
		}
	}

	var areaOptionID, functionOptionID string
	if *area != "" && roadArea != nil {
		for _, o := range roadArea.Options {
			if strings.HasPrefix(o.Name, *area) {
				areaOptionID = o.ID
				break
			}
		}
		if areaOptionID == "" {
			return fmt.Errorf("area '%s' not found in roadmap", *area)
		}
	}

	if *functionF != "" && roadFunction != nil {
		for _, o := range roadFunction.Options {
			if strings.HasPrefix(o.Name, *functionF) {
				functionOptionID = o.ID
				break
			}
		}
		if functionOptionID == "" {
			return fmt.Errorf("function '%s' not found in roadmap", *functionF)
		}
	}

	itemsJSON := runCmd("gh", "project", "item-list", *project, appendFlags)
	var items struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
		return fmt.Errorf("error parsing items: %v", err)
	}

	for _, item := range items.Items {
		if item.Content.Type == "DraftIssue" {
			log.Printf("Skipping draft: %s", item.Content.Title)
			continue
		}

		log.Printf("Adding issue '%s' to the roadmap board", item.Title)
		addOut := runCmd("gh", "project", "item-add", roadmap, "--owner", "giantswarm", "--format", "json", "--url", item.Content.URL)
		var newItem struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(addOut), &newItem); err != nil {
			log.Printf("Error adding item for %s: %v", item.Title, err)
			continue
		}

		switch *typ {
		case "team":
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadTeam.ID, "--single-select-option-id", typeOptionID)
		case "sig":
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadSIG.ID, "--single-select-option-id", typeOptionID)
		case "wg":
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadWG.ID, "--single-select-option-id", typeOptionID)
		}

		if areaOptionID != "" {
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadArea.ID, "--single-select-option-id", areaOptionID)
		}
		if functionOptionID != "" {
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadFunction.ID, "--single-select-option-id", functionOptionID)
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
				runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
					"--field-id", roadStatus.ID, "--single-select-option-id", statusOptID)
			} else {
				log.Printf("Status '%s' not found in roadmap.", s)
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
				runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
					"--field-id", roadKind.ID, "--single-select-option-id", kindOptID)
			} else {
				log.Printf("Kind '%s' not found in roadmap.", k)
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
				runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
					"--field-id", roadWorkstream.ID, "--single-select-option-id", wsOptID)
			} else {
				log.Printf("Workstream '%s' not found in roadmap.", ws)
			}
		}

		if item.StartDate != "" && item.StartDate != "null" && roadStartDate != nil {
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadStartDate.ID, "--date", item.StartDate)
		}
		if item.TargetDate != "" && item.TargetDate != "null" && roadTargetDate != nil {
			runCmd("gh", "project", "item-edit", "--project-id", roadmapPID, "--id", newItem.ID,
				"--field-id", roadTargetDate.ID, "--date", item.TargetDate)
		}

		if !*dryRun {
			runCmd("gh", "project", "item-archive", *project, "--id", item.ID, "--owner", "giantswarm")
		}
	}
	return nil
}
