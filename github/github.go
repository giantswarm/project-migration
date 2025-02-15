package github

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"log/slog"
)

// Client defines a human‐readable interface for GitHub project commands.
type Client interface {
	ListProjects() (string, error)
	FieldList(project string) (string, error)
	ListItems(project string) (string, error)
	AddItem(targetProject, url string) (string, error)
	EditItem(projectID, itemID, fieldID, optionID string) (string, error)
	EditItemDate(projectID, itemID, fieldID, date string) (string, error)
	ArchiveItem(project, item string) (string, error)
}

// clientImpl implements Client.
type clientImpl struct {
	Verbose bool
}

var appendFlags = []string{"--owner", "giantswarm", "-L", "10000", "--format", "json"}

func (c *clientImpl) runCmd(cmdStr string, args ...string) (string, error) {
	cmdLine := fmt.Sprintf("%s %s", cmdStr, strings.Join(args, " "))
	if c.Verbose {
		slog.Info("Executing command", "cmd", cmdLine)
	}
	cmd := exec.Command(cmdStr, args...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		slog.Error("Error executing command", "cmd", cmdStr, "stderr", errOut.String())
		return "", err
	}
	if c.Verbose {
		slog.Info("Command output", "output", out.String())
	}
	return out.String(), nil
}

func (c *clientImpl) ListProjects() (string, error) {
	args := append([]string{"project", "list"}, appendFlags...)
	return c.runCmd("gh", args...)
}

func (c *clientImpl) FieldList(project string) (string, error) {
	args := append([]string{"project", "field-list", project}, appendFlags...)
	return c.runCmd("gh", args...)
}

func (c *clientImpl) ListItems(project string) (string, error) {
	args := append([]string{"project", "item-list", project}, appendFlags...)
	return c.runCmd("gh", args...)
}

func (c *clientImpl) AddItem(targetProject, url string) (string, error) {
	return c.runCmd("gh", "project", "item-add", targetProject, "--owner", "giantswarm", "--format", "json", "--url", url)
}

func (c *clientImpl) EditItem(projectID, itemID, fieldID, optionID string) (string, error) {
	return c.runCmd("gh", "project", "item-edit", "--project-id", projectID, "--id", itemID, "--field-id", fieldID, "--single-select-option-id", optionID)
}

func (c *clientImpl) EditItemDate(projectID, itemID, fieldID, date string) (string, error) {
	return c.runCmd("gh", "project", "item-edit", "--project-id", projectID, "--id", itemID, "--field-id", fieldID, "--date", date)
}

func (c *clientImpl) ArchiveItem(project, item string) (string, error) {
	return c.runCmd("gh", "project", "item-archive", project, "--id", item, "--owner", "giantswarm")
}

// NewClient returns a new GitHub client.
func NewClient(verbose bool) Client {
	return &clientImpl{Verbose: verbose}
}
