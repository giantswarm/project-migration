package ghprojects

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const ownerFlags = "--owner giantswarm -L 10000 --format json"

// run is an internal helper to execute the gh command.
func run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running gh %v: %w", args, err)
	}
	return out.String(), nil
}

// ListProjects executes "gh project list" with default flags if none are provided.
func ListProjects(flags ...string) (string, error) {
	if len(flags) == 0 {
		flags = strings.Split(ownerFlags, " ")
	}
	args := append([]string{"project", "list"}, flags...)
	return run(args...)
}

// GetFieldList executes "gh project field-list <project>" with default flags if none are provided.
func GetFieldList(project string, flags ...string) (string, error) {
	if len(flags) == 0 {
		flags = strings.Split(ownerFlags, " ")
	}
	args := append([]string{"project", "field-list", project}, flags...)
	return run(args...)
}

// GetItemList executes "gh project item-list <project>" with default flags if none are provided.
func GetItemList(project string, flags ...string) (string, error) {
	if len(flags) == 0 {
		flags = strings.Split(ownerFlags, " ")
	}
	args := append([]string{"project", "item-list", project}, flags...)
	return run(args...)
}

// AddItem executes "gh project item-add" for the given project and URL.
func AddItem(project, url string) (string, error) {
	args := []string{"project", "item-add", project, "--owner", "giantswarm", "--format", "json", "--url", url}
	return run(args...)
}

// EditItemSingle wraps the single-select edit command.
func EditItemSingle(projectID, id, fieldID, optionID string) (string, error) {
	args := []string{"project", "item-edit", "--project-id", projectID, "--id", id, "--field-id", fieldID, "--single-select-option-id", optionID}
	return run(args...)
}

// EditItemDate wraps the date edit command.
func EditItemDate(projectID, id, fieldID, date string) (string, error) {
	args := []string{"project", "item-edit", "--project-id", projectID, "--id", id, "--field-id", fieldID, "--date", date}
	return run(args...)
}

// ArchiveItem executes "gh project item-archive" for the given project and item ID.
func ArchiveItem(project, id string) (string, error) {
	args := []string{"project", "item-archive", project, "--id", id, "--owner", "giantswarm"}
	return run(args...)
}
