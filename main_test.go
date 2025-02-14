package main

import (
	"fmt"
	"strings"
	"testing"
)

// save the original runCmd so we can restore it.
var origRunCmd = runCmd

// testRunCmd is a stub that returns canned responses.
func testRunCmd(cmdStr string, args ...string) string {
	// join command and args to simulate a key.
	key := fmt.Sprintf("%s %s", cmdStr, strings.Join(args, " "))
	switch {
	// simulate project list returns one project with number matching flag.
	case strings.Contains(key, "gh project list"):
		return `[{"id": "proj-123", "number": 301}]`
	// simulate project field-list for source project.
	case strings.Contains(key, "gh project field-list") && strings.Contains(key, "301"):
		// return minimal fields for Status, Kind, and Workstream
		return `{"fields": [
			{"id": "f-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}, {"id": "s-later", "name": "Later 🌃"}]},
			{"id": "f-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "f-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]}
		]}`
	// simulate project field-list for roadmap.
	case strings.Contains(key, "gh project field-list") && strings.Contains(key, roadmap):
		return `{"fields": [
			{"id": "rf-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}, {"id": "s-later", "name": "Later 🌃"}]},
			{"id": "rf-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "rf-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]},
			{"id": "rf-team", "name": "Team", "options": [{"id": "team-rocket", "name": "Rocket Team"}]},
			{"id": "rf-sig", "name": "SIG", "options": [{"id": "sig-rocket", "name": "Rocket SIG"}]},
			{"id": "rf-working group", "name": "Working Group", "options": [{"id": "wg-rocket", "name": "Rocket WG"}]},
			{"id": "rf-area", "name": "Area", "options": [{"id": "area-kaas", "name": "KaaS"}]},
			{"id": "rf-function", "name": "Function", "options": [{"id": "func-strat", "name": "Product Strategy"}]},
			{"id": "rf-startdate", "name": "Start Date", "options": []},
			{"id": "rf-targetdate", "name": "Target Date", "options": []}
		]}`
	// simulate item list returning one non-draft issue with date fields.
	case strings.Contains(key, "gh project item-list"):
		return `{"items": [{
			"id": "item-1",
			"title": "Test Issue",
			"status": "Backlog",
			"kind": "Feature",
			"workstream": "Engineering",
			"start Date": "2023-10-01",
			"target Date": "2023-11-01",
			"content": {"type": "Issue", "title": "Test Issue", "url": "http://example.com/issue/1"}
		}]}`
	// simulate item add returns a new item ID.
	case strings.Contains(key, "gh project item-add"):
		return `{"id": "new-item-1"}`
	// for item-edit and item-archive, return empty string.
	case strings.Contains(key, "gh project item-edit"):
		return ""
	case strings.Contains(key, "gh project item-archive"):
		return ""
	default:
		return ""
	}
}

func TestRunMigrationSuccess(t *testing.T) {
	// Override runCmd with our test stub.
	runCmd = testRunCmd
	defer func() { runCmd = origRunCmd }()

	// Setup flags for a successful migration.
	*project = "301"
	*typ = "team"
	*name = "Rocket"
	*area = "KaaS"
	*functionF = "Product Strategy"
	*dryRun = true

	// Call runMigration.
	if err := runMigration(); err != nil {
		t.Errorf("Expected migration to succeed, got error: %v", err)
	}
}

// Additional tests can be added to simulate errors based on flags or missing fields.
func TestRunMigrationMissingProject(t *testing.T) {
	// Restore runCmd in case it is used.
	runCmd = testRunCmd
	defer func() { runCmd = origRunCmd }()

	*project = ""
	*typ = "team"
	*name = "Rocket"
	if err := runMigration(); err == nil {
		t.Errorf("Expected error for missing project number")
	}
}
