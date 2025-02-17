package ghprojects

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRunInvalidCommand(t *testing.T) {
	// Calling run with an invalid command should return an error.
	_, err := run("nonexistentcommand")
	if err == nil {
		t.Errorf("Expected error for invalid command")
	}
}

func TestListProjectsHelp(t *testing.T) {
	// Use a flag that triggers help; if the gh CLI is not installed, skip the test.
	out, err := ListProjects("--help")
	if err != nil {
		t.Skip("gh tool not available, skipping test")
	}
	// Check for "USAGE" (case sensitive) in output.
	if !strings.Contains(out, "USAGE") {
		t.Errorf("Expected output to contain 'USAGE', got: %v", out)
	}
}

func TestGetFieldListInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Pass an unlikely project identifier to provoke an error.
	_, err := GetFieldList("nonexistent_project")
	if err == nil {
		t.Errorf("Expected error retrieving fields for nonexistent project")
	}
}

func TestGetItemListInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Use a nonexistent project to force a failure.
	_, err := GetItemList("nonexistent_project")
	if err == nil {
		t.Errorf("Expected error retrieving items for nonexistent project")
	}
}

func TestAddItemInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Attempt to add an item using an invalid URL.
	_, err := AddItem("nonexistent_project", "invalid_url")
	if err == nil {
		t.Errorf("Expected error adding item with invalid URL")
	}
}

func TestEditItemSingleInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Provide dummy IDs; expect error.
	_, err := EditItemSingle("dummy_project", "dummy_id", "dummy_field", "dummy_option")
	if err == nil {
		t.Errorf("Expected error editing item with invalid parameters")
	}
}

func TestEditItemDateInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Provide dummy IDs and an invalid date format.
	_, err := EditItemDate("dummy_project", "dummy_id", "dummy_field", "invalid-date")
	if err == nil {
		t.Errorf("Expected error editing date with invalid parameters")
	}
}

func TestArchiveItemInvalid(t *testing.T) {
	// Skip test if gh CLI is not available.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available, skipping test")
	}
	// Attempt to archive an item for a nonexistent project.
	_, err := ArchiveItem("nonexistent_project", "dummy_item")
	if err == nil {
		t.Errorf("Expected error archiving item with invalid parameters")
	}
}
