package main

import (
	"strings"
	"testing"

	"giantswarm.io/project-migration/cli"
	"giantswarm.io/project-migration/migration" // new import
)

const roadmap = "273" // added constant for roadmap

// stubClient implements github.Client for testing.
type stubClient struct{}

func (s *stubClient) ListProjects() (string, error) {
	return `[{"id": "proj-123", "number": 301}]`, nil
}

func (s *stubClient) FieldList(project string) (string, error) {
	if strings.Contains(project, "301") {
		return `{"fields": [
			{"id": "f-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}, {"id": "s-later", "name": "Later 🌃"}]},
			{"id": "f-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "f-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]}
		]}`, nil
	} else if project == roadmap {
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
		]}`, nil
	}
	return "", nil
}

func (s *stubClient) ListItems(project string) (string, error) {
	return `{"items": [{
		"id": "item-1",
		"title": "Test Issue",
		"status": "Backlog",
		"kind": "Feature",
		"workstream": "Engineering",
		"start Date": "2023-10-01",
		"target Date": "2023-11-01",
		"content": {"type": "Issue", "title": "Test Issue", "url": "http://example.com/issue/1"}
	}]}`, nil
}

func (s *stubClient) AddItem(targetProject, url string) (string, error) {
	return `{"id": "new-item-1"}`, nil
}

func (s *stubClient) EditItem(projectID, itemID, fieldID, optionID string) (string, error) {
	return "", nil
}

func (s *stubClient) EditItemDate(projectID, itemID, fieldID, date string) (string, error) {
	return "", nil
}

func (s *stubClient) ArchiveItem(project, item string) (string, error) {
	return "", nil
}

func TestRunMigrationSuccess(t *testing.T) {
	stub := &stubClient{}
	cfg := &cli.Config{
		Project:  "301",
		Type:     "team",
		Name:     "Rocket",
		Area:     "KaaS",
		Function: "Product Strategy",
		DryRun:   true,
		Verbose:  false,
	}
	// Updated to use migration.Run instead of runMigration.
	if err := migration.Run(cfg, stub); err != nil {
		t.Errorf("Expected migration to succeed, got error: %v", err)
	}
}

func TestRunMigrationMissingProject(t *testing.T) {
	stub := &stubClient{}
	cfg := &cli.Config{
		Project: "",
		Type:    "team",
		Name:    "Rocket",
	}
	// Updated to use migration.Run.
	if err := migration.Run(cfg, stub); err == nil {
		t.Errorf("Expected error for missing project number")
	}
}
