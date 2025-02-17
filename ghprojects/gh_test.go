package ghprojects

import (
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
