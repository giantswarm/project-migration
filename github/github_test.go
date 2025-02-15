package github

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// fakeExecCommand uses the helper process technique.
// It returns a command that runs the test binary in helper mode.
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	// Pass the expected output via an env var.
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "FAKE_OUTPUT=" + os.Getenv("FAKE_OUTPUT")}
	return cmd
}

// TestHelperProcess is not a real test. It is invoked as a helper process.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Write the fake output from environment variable.
	fakeOut := os.Getenv("FAKE_OUTPUT")
	// Write to Stdout.
	os.Stdout.Write([]byte(fakeOut))
	os.Exit(0)
}

func TestListProjects(t *testing.T) {
	// Set the expected output.
	expected := `[{"id": "proj-123", "number": 301}]`
	os.Setenv("FAKE_OUTPUT", expected)
	// Override execCommand.
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects returned an error: %v", err)
	}
	if strings.TrimSpace(out) != expected {
		t.Errorf("Expected output %q, got %q", expected, out)
	}
}

func TestFieldList(t *testing.T) {
	expected := `{"fields": [{"id": "f-status", "name": "Status"}]}`
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.FieldList("301")
	if err != nil {
		t.Fatalf("FieldList returned an error: %v", err)
	}
	if strings.TrimSpace(out) != expected {
		t.Errorf("Expected output %q, got %q", expected, out)
	}
}

func TestListItems(t *testing.T) {
	expected := `{"items": [{"id": "item-1", "title": "Test Issue"}]}`
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.ListItems("301")
	if err != nil {
		t.Fatalf("ListItems returned an error: %v", err)
	}
	if strings.TrimSpace(out) != expected {
		t.Errorf("Expected output %q, got %q", expected, out)
	}
}

func TestAddItem(t *testing.T) {
	expected := `{"id": "new-item-1"}`
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.AddItem("roadmap", "http://example.com/issue/1")
	if err != nil {
		t.Fatalf("AddItem returned an error: %v", err)
	}
	if strings.TrimSpace(out) != expected {
		t.Errorf("Expected output %q, got %q", expected, out)
	}
}

func TestEditItem(t *testing.T) {
	expected := ``
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.EditItem("projID", "itemID", "fieldID", "optionID")
	if err != nil {
		t.Fatalf("EditItem returned an error: %v", err)
	}
	if out != expected {
		t.Errorf("Expected empty output, got %q", out)
	}
}

func TestEditItemDate(t *testing.T) {
	expected := ``
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.EditItemDate("projID", "itemID", "fieldID", "2023-10-10")
	if err != nil {
		t.Fatalf("EditItemDate returned an error: %v", err)
	}
	if out != expected {
		t.Errorf("Expected empty output, got %q", out)
	}
}

func TestArchiveItem(t *testing.T) {
	expected := ``
	os.Setenv("FAKE_OUTPUT", expected)
	orig := execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = orig }()

	client := NewClient(false)
	out, err := client.ArchiveItem("projID", "itemID")
	if err != nil {
		t.Fatalf("ArchiveItem returned an error: %v", err)
	}
	if out != expected {
		t.Errorf("Expected empty output, got %q", out)
	}
}
