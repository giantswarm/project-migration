package cli

import (
	"flag"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	// Reset global flags and simulate command-line arguments.
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
	os.Args = []string{"cmd", "-p", "301", "-t", "team", "-n", "Rocket", "-a", "KaaS", "-f", "Product Strategy", "-d"}

	cfg := Parse()

	if cfg.Project != "301" {
		t.Errorf("Expected project '301', got '%s'", cfg.Project)
	}
	if cfg.Type != "team" {
		t.Errorf("Expected type 'team', got '%s'", cfg.Type)
	}
	if cfg.Name != "Rocket" {
		t.Errorf("Expected name 'Rocket', got '%s'", cfg.Name)
	}
	if cfg.Area != "KaaS" {
		t.Errorf("Expected area 'KaaS', got '%s'", cfg.Area)
	}
	if cfg.Function != "Product Strategy" {
		t.Errorf("Expected function 'Product Strategy', got '%s'", cfg.Function)
	}
	if !cfg.DryRun {
		t.Errorf("Expected DryRun to be true")
	}
}
