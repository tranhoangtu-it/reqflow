package commands_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
	featurehttp "github.com/ye-kart/reqflow/internal/features/http"
)

func TestEnvFlag_LoadsVariablesIntoRequest(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	// Create an environment with a base_url variable.
	env := domain.Environment{
		Name: "dev",
		Variables: []domain.Variable{
			{Key: "base_url", Value: "https://dev.api.example.com", Scope: domain.ScopeEnvironment},
		},
	}
	if err := store.WriteEnvironment(filepath.Join(dir, "dev.yaml"), env); err != nil {
		t.Fatal(err)
	}

	// Capture what variables are passed to the executor.
	var capturedURL string
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedURL = req.URL
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := &app.App{
		HTTPExecutor: featurehttp.NewExecutor(mock),
		Storage:      store,
	}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "{{base_url}}/users",
		"-e", "dev",
		"--env-dir", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://dev.api.example.com/users"
	if capturedURL != expected {
		t.Errorf("expected URL %q, got %q", expected, capturedURL)
	}
}

func TestEnvFlag_NotSet_PassesNilVars(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := &app.App{
		HTTPExecutor: featurehttp.NewExecutor(mock),
		Storage:      storage.NewFilesystem(),
	}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
