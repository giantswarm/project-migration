package main_test

import (
	"os"
	"os/exec"
	"testing"
)

func TestMainSuccess(t *testing.T) {
	// Run main.go with valid flags and TEST_MAIN set so that it exits early.
	cmd := exec.Command("go", "run", "main.go", "-p", "301", "-t", "team", "-n", "TestTeam")
	cmd.Env = append(os.Environ(), "TEST_MAIN=1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success exit, but got error: %v", err)
	}
}

func TestMainMissingFlag(t *testing.T) {
	// Run main.go missing the -n flag.
	cmd := exec.Command("go", "run", "main.go", "-p", "301", "-t", "team")
	if err := cmd.Run(); err == nil {
		t.Fatalf("Expected failure due to missing flag(s), but command succeeded")
	}
}

func TestMainMissingProject(t *testing.T) {
	// Omit the -p flag.
	cmd := exec.Command("go", "run", "main.go", "-t", "team", "-n", "TestTeam")
	if err := cmd.Run(); err == nil {
		t.Fatalf("Expected failure due to missing -p flag, but command succeeded")
	}
}

func TestMainMissingType(t *testing.T) {
	// Omit the -t flag.
	cmd := exec.Command("go", "run", "main.go", "-p", "301", "-n", "TestTeam")
	if err := cmd.Run(); err == nil {
		t.Fatalf("Expected failure due to missing -t flag, but command succeeded")
	}
}

func TestMainInvalidType(t *testing.T) {
	// Provide an invalid value for -t flag.
	cmd := exec.Command("go", "run", "main.go", "-p", "301", "-t", "invalid", "-n", "TestTeam")
	if err := cmd.Run(); err == nil {
		t.Fatalf("Expected failure due to invalid -t flag value, but command succeeded")
	}
}
