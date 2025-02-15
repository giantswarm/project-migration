package main

import (
	"fmt"
	"strings"
	"testing"

	"giantswarm.io/project-migration/cli"
)

// save the original runCmd so we can restore it.
var origRunCmd = runCmd

// testRunCmd is a stub that returns canned responses.
func testRunCmd(cmdStr string, args ...string) string {
	key := fmt.Sprintf("%s %s", cmdStr, strings.Join(args, " "))
	switch {
	case strings.Contains(key, "gh project list"):
		return `[{"id": "proj-123", "number": 301}]`
	case strings.Contains(key, "gh project field-list") && strings.Contains(key, "301"):
		return `{"fields": [
			{"id": "f-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}, {"id": "s-later", "name": "Later 🌃"}]},
			{"id": "f-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "f-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]}
		]}`
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
	case strings.Contains(key, "gh project item-add"):
		return `{"id": "new-item-1"}`
	case strings.Contains(key, "gh project item-edit"):
		return ""
	case strings.Contains(key, "gh project item-archive"):
		return ""
	default:
		return ""
	}
}

func TestRunMigrationSuccess(t *testing.T) {
	runCmd = testRunCmd
	defer func() { runCmd = origRunCmd }()

	cfg := &cli.Config{
		Project:  "301",
		Type:     "team",
		Name:     "Rocket",
		Area:     "KaaS",
		Function: "Product Strategy",
		DryRun:   true,
	}

	if err := runMigration(cfg); err != nil {
		t.Errorf("Expected migration to succeed, got error: %v", err)
	}
}

func TestRunMigrationMissingProject(t *testing.T) {
	runCmd = testRunCmd
	defer func() { runCmd = origRunCmd }()

	cfg := &cli.Config{
		Project: "",
		Type:    "team",
		Name:    "Rocket",
	}

	if err := runMigration(cfg); err == nil {
		t.Errorf("Expected error for missing project number")
	}
}
