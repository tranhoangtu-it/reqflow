package commands_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
	featurehttp "github.com/ye-kart/reqflow/internal/features/http"
)

func newTestAppWithStorage(mock *mockHTTPClient, store *storage.Filesystem) *app.App {
	return &app.App{
		HTTPExecutor: featurehttp.NewExecutor(mock),
		Storage:      store,
	}
}

func TestEnvList_ShowsEnvironments(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	// Create environment files.
	for _, name := range []string{"dev.yaml", "staging.yaml"} {
		env := domain.Environment{Name: strings.TrimSuffix(name, ".yaml")}
		if err := store.WriteEnvironment(filepath.Join(dir, name), env); err != nil {
			t.Fatal(err)
		}
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"env", "list", "--env-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "dev") {
		t.Errorf("expected output to contain 'dev', got: %s", output)
	}
	if !strings.Contains(output, "staging") {
		t.Errorf("expected output to contain 'staging', got: %s", output)
	}
}

func TestEnvShow_DisplaysVariables(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	env := domain.Environment{
		Name: "dev",
		Variables: []domain.Variable{
			{Key: "base_url", Value: "https://dev.api.example.com", Scope: domain.ScopeEnvironment},
			{Key: "api_key", Value: "dev-key-123", Scope: domain.ScopeEnvironment},
		},
	}
	if err := store.WriteEnvironment(filepath.Join(dir, "dev.yaml"), env); err != nil {
		t.Fatal(err)
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"env", "show", "dev", "--env-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "base_url") {
		t.Errorf("expected output to contain 'base_url', got: %s", output)
	}
	if !strings.Contains(output, "https://dev.api.example.com") {
		t.Errorf("expected output to contain URL, got: %s", output)
	}
}

func TestEnvCreate_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"env", "create", "production", "--env-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created.
	path := filepath.Join(dir, "production.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected environment file to exist")
	}

	// Verify it can be read back.
	loaded, err := store.ReadEnvironment(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if loaded.Name != "production" {
		t.Errorf("expected name 'production', got %q", loaded.Name)
	}
}

func TestEnvSet_UpdatesVariable(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	// Create an initial environment.
	env := domain.Environment{
		Name: "dev",
		Variables: []domain.Variable{
			{Key: "base_url", Value: "https://old.example.com", Scope: domain.ScopeEnvironment},
		},
	}
	if err := store.WriteEnvironment(filepath.Join(dir, "dev.yaml"), env); err != nil {
		t.Fatal(err)
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"env", "set", "dev", "api_key", "new-key-456", "--env-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the variable was set.
	loaded, err := store.ReadEnvironment(filepath.Join(dir, "dev.yaml"))
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	vars := make(map[string]string)
	for _, v := range loaded.Variables {
		vars[v.Key] = v.Value
	}

	if vars["api_key"] != "new-key-456" {
		t.Errorf("expected api_key 'new-key-456', got %q", vars["api_key"])
	}
	// Original variable should still exist.
	if vars["base_url"] != "https://old.example.com" {
		t.Errorf("expected base_url preserved, got %q", vars["base_url"])
	}
}

func noopDoFunc(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
	return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
}
