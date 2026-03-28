package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestReadEnvironment_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dev.yaml")

	content := `name: dev
variables:
  base_url: "https://dev.api.example.com"
  api_key: "dev-key-123"
  timeout: "5000"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fs := storage.NewFilesystem()
	env, err := fs.ReadEnvironment(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Name != "dev" {
		t.Errorf("expected name 'dev', got %q", env.Name)
	}

	vars := variableMap(env.Variables)
	if vars["base_url"] != "https://dev.api.example.com" {
		t.Errorf("expected base_url 'https://dev.api.example.com', got %q", vars["base_url"])
	}
	if vars["api_key"] != "dev-key-123" {
		t.Errorf("expected api_key 'dev-key-123', got %q", vars["api_key"])
	}
	if vars["timeout"] != "5000" {
		t.Errorf("expected timeout '5000', got %q", vars["timeout"])
	}
}

func TestReadEnvironment_NonexistentFile(t *testing.T) {
	fs := storage.NewFilesystem()
	_, err := fs.ReadEnvironment("/nonexistent/path/env.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestWriteEnvironment_CreatesValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "staging.yaml")

	env := domain.Environment{
		Name: "staging",
		Variables: []domain.Variable{
			{Key: "host", Value: "staging.example.com", Scope: domain.ScopeEnvironment},
			{Key: "port", Value: "8080", Scope: domain.ScopeEnvironment},
		},
	}

	fs := storage.NewFilesystem()
	err := fs.WriteEnvironment(path, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist")
	}
}

func TestListEnvironments_ReturnsYAMLFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some .yaml files and a non-yaml file.
	for _, name := range []string{"dev.yaml", "staging.yaml", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	fs := storage.NewFilesystem()
	names, err := fs.ListEnvironments(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 environments, got %d: %v", len(names), names)
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["dev"] {
		t.Error("expected 'dev' in list")
	}
	if !nameSet["staging"] {
		t.Error("expected 'staging' in list")
	}
}

func TestListEnvironments_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	fs := storage.NewFilesystem()
	names, err := fs.ListEnvironments(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("expected 0 environments, got %d", len(names))
	}
}

func TestRoundTrip_WriteReadEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prod.yaml")

	original := domain.Environment{
		Name: "prod",
		Variables: []domain.Variable{
			{Key: "base_url", Value: "https://api.example.com", Scope: domain.ScopeEnvironment},
			{Key: "api_key", Value: "prod-key-456", Scope: domain.ScopeEnvironment},
		},
	}

	fs := storage.NewFilesystem()
	if err := fs.WriteEnvironment(path, original); err != nil {
		t.Fatalf("write error: %v", err)
	}

	loaded, err := fs.ReadEnvironment(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, loaded.Name)
	}

	originalVars := variableMap(original.Variables)
	loadedVars := variableMap(loaded.Variables)

	for key, want := range originalVars {
		if got := loadedVars[key]; got != want {
			t.Errorf("variable %q: expected %q, got %q", key, want, got)
		}
	}
}

// variableMap converts a slice of Variables to a key-value map for easier assertions.
func variableMap(vars []domain.Variable) map[string]string {
	m := make(map[string]string)
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m
}
