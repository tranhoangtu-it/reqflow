package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
)

func TestRunCommand_RequiresFileArgument(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)
	root.SetArgs([]string{"run"})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing file argument, got nil")
	}
}

func TestRunCommand_FileNotFound(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)
	root.SetArgs([]string{"run", "/nonexistent/workflow.yaml"})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestRunCommand_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "bad.yaml")
	os.WriteFile(yamlPath, []byte(`{{{invalid`), 0644)

	a := &app.App{}
	root := commands.NewRootCommand(a)
	root.SetArgs([]string{"run", yamlPath})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestRunCommand_DryRunValidates(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "workflow.yaml")
	yamlContent := `
name: test-workflow
steps:
  - name: get users
    method: GET
    url: https://api.example.com/users
`
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	a := &app.App{}
	root := commands.NewRootCommand(a)
	root.SetArgs([]string{"run", yamlPath, "--dry-run"})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected dry-run output, got empty")
	}
}

func TestRunCommand_ExistsAsSubcommand(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'run' subcommand to be registered")
	}
}
