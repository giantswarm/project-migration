package migration

import (
	"strings"
	"testing"

	"giantswarm.io/project-migration/cli"
)

// stubClient implements github.Client with valid responses.
type stubClient struct{}

func (s *stubClient) ListProjects() (string, error) {
	return `[{"id": "proj-123", "number": 301}]`, nil
}

func (s *stubClient) FieldList(project string) (string, error) {
	if project == "301" {
		// Project fields with required fields
		return `{"fields": [
			{"id": "f-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}]},
			{"id": "f-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "f-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]}
		]}`, nil
	} else if project == "273" {
		// Roadmap fields including additional required fields for team.
		return `{"fields": [
			{"id": "rf-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}]},
			{"id": "rf-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "rf-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]},
			{"id": "rf-team", "name": "Team", "options": [{"id": "team-rocket", "name": "Rocket Team"}]},
			{"id": "rf-area", "name": "Area", "options": [{"id": "area-kaas", "name": "KaaS"}]},
			{"id": "rf-function", "name": "Function", "options": [{"id": "func-strat", "name": "Product Strategy"}]},
			{"id": "rf-startdate", "name": "Start Date", "options": []},
			{"id": "rf-targetdate", "name": "Target Date", "options": []}
		]}`, nil
	}
	return "", nil
}

func (s *stubClient) ListItems(project string) (string, error) {
	// Return one non-draft item.
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

func TestRunSuccess(t *testing.T) {
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
	if err := Run(cfg, stub); err != nil {
		t.Errorf("Expected Run to succeed, got error: %v", err)
	}
}

// stubClientMissingField simulates missing a required field (e.g. "Status") in project fields.
type stubClientMissingField struct{}

func (s *stubClientMissingField) ListProjects() (string, error) {
	return `[{"id": "proj-123", "number": 301}]`, nil
}

func (s *stubClientMissingField) FieldList(project string) (string, error) {
	if project == "301" {
		// Missing "Status" field.
		return `{"fields": [
			{"id": "f-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "f-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]}
		]}`, nil
	} else if project == "273" {
		return `{"fields": [
			{"id": "rf-status", "name": "Status", "options": [{"id": "s-backlog", "name": "Backlog"}]},
			{"id": "rf-kind", "name": "Kind", "options": [{"id": "k-feature", "name": "Feature"}]},
			{"id": "rf-workstream", "name": "Workstream", "options": [{"id": "w-eng", "name": "Engineering"}]},
			{"id": "rf-team", "name": "Team", "options": [{"id": "team-rocket", "name": "Rocket Team"}]},
			{"id": "rf-area", "name": "Area", "options": [{"id": "area-kaas", "name": "KaaS"}]},
			{"id": "rf-function", "name": "Function", "options": [{"id": "func-strat", "name": "Product Strategy"}]},
			{"id": "rf-startdate", "name": "Start Date", "options": []},
			{"id": "rf-targetdate", "name": "Target Date", "options": []}
		]}`, nil
	}
	return "", nil
}

func (s *stubClientMissingField) ListItems(project string) (string, error) {
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

func (s *stubClientMissingField) AddItem(targetProject, url string) (string, error) {
	return `{"id": "new-item-1"}`, nil
}

func (s *stubClientMissingField) EditItem(projectID, itemID, fieldID, optionID string) (string, error) {
	return "", nil
}

func (s *stubClientMissingField) EditItemDate(projectID, itemID, fieldID, date string) (string, error) {
	return "", nil
}

func (s *stubClientMissingField) ArchiveItem(project, item string) (string, error) {
	return "", nil
}

func TestRunMissingProjectField(t *testing.T) {
	stub := &stubClientMissingField{}
	cfg := &cli.Config{
		Project: "301",
		Type:    "team",
		Name:    "Rocket",
		Area:    "KaaS",
		// Function provided to meet all roadmap requirements.
		Function: "Product Strategy",
		DryRun:   true,
		Verbose:  false,
	}
	err := Run(cfg, stub)
	if err == nil || !strings.Contains(err.Error(), "required fields missing") {
		t.Errorf("Expected error due to missing project field, got: %v", err)
	}
}
