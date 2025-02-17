package cli

import (
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Override os.Args to set flags.
	os.Args = []string{
		"cmd",
		"-p", "301",
		"-t", "team",
		"-n", "TestTeam",
		"-a", "TestArea",
		"-f", "TestFunction",
	}
	opts := ParseFlags()

	if opts.Project != "301" {
		t.Errorf("Expected project '301', got %s", opts.Project)
	}
	if opts.Type != "team" {
		t.Errorf("Expected type 'team', got %s", opts.Type)
	}
	if opts.Name != "TestTeam" {
		t.Errorf("Expected name 'TestTeam', got %s", opts.Name)
	}
	if opts.Area != "TestArea" {
		t.Errorf("Expected area 'TestArea', got %s", opts.Area)
	}
	if opts.Function != "TestFunction" {
		t.Errorf("Expected function 'TestFunction', got %s", opts.Function)
	}
}
