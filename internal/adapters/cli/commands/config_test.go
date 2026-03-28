package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
)

func TestConfigList_ShowsResolvedValues(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_output: pretty
default_timeout: 30s
no_color: false
default_headers:
  User-Agent: "reqflow/0.1"
`
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	a := &app.App{}
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "default_output: pretty") {
		t.Errorf("expected output to contain 'default_output: pretty', got:\n%s", output)
	}
	if !strings.Contains(output, "default_timeout: 30s") {
		t.Errorf("expected output to contain 'default_timeout: 30s', got:\n%s", output)
	}
	if !strings.Contains(output, "User-Agent") {
		t.Errorf("expected output to contain 'User-Agent' header, got:\n%s", output)
	}
}

func TestConfigGet_ReturnsCorrectValue(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_timeout: 45s
`
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	a := &app.App{}
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "get", "default_timeout"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "45s" {
		t.Errorf("expected '45s', got %q", output)
	}
}

func TestConfigSet_UpdatesGlobalConfig(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	a := &app.App{}
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "set", "default_output", "json"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the global config file was written.
	data, err := os.ReadFile(filepath.Join(globalDir, "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read global config: %v", err)
	}
	if !strings.Contains(string(data), "json") {
		t.Errorf("expected global config to contain 'json', got:\n%s", string(data))
	}
}

func TestConfigSet_WithLocalFlag_UpdatesLocalConfig(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	a := &app.App{}
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "set", "default_output", "json", "--local"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the project-local config file was written.
	data, err := os.ReadFile(filepath.Join(projectDir, ".reqflow.yaml"))
	if err != nil {
		t.Fatalf("failed to read project config: %v", err)
	}
	if !strings.Contains(string(data), "json") {
		t.Errorf("expected project config to contain 'json', got:\n%s", string(data))
	}
}
