package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestCollectionList_ShowsCollections(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	// Create collection files.
	for _, name := range []string{"api-tests", "integration"} {
		col := domain.Collection{Name: name}
		if err := store.WriteCollection(filepath.Join(dir, name+".yaml"), col); err != nil {
			t.Fatal(err)
		}
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"collection", "list", "--collection-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "api-tests") {
		t.Errorf("expected output to contain 'api-tests', got: %s", output)
	}
	if !strings.Contains(output, "integration") {
		t.Errorf("expected output to contain 'integration', got: %s", output)
	}
}

func TestCollectionShow_DisplaysContent(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	col := domain.Collection{
		Name:        "my-api",
		Description: "My API collection",
		Requests: []domain.SavedRequest{
			{
				Name: "Health Check",
				Config: domain.RequestConfig{
					Method: domain.MethodGet,
					URL:    "https://api.example.com/health",
				},
			},
		},
		Folders: []domain.Folder{
			{
				Name: "Users",
				Requests: []domain.SavedRequest{
					{
						Name: "List Users",
						Config: domain.RequestConfig{
							Method: domain.MethodGet,
							URL:    "https://api.example.com/users",
						},
					},
				},
			},
		},
	}
	if err := store.WriteCollection(filepath.Join(dir, "my-api.yaml"), col); err != nil {
		t.Fatal(err)
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"collection", "show", "my-api", "--collection-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "my-api") {
		t.Errorf("expected output to contain 'my-api', got: %s", output)
	}
	if !strings.Contains(output, "Health Check") {
		t.Errorf("expected output to contain 'Health Check', got: %s", output)
	}
	if !strings.Contains(output, "Users") {
		t.Errorf("expected output to contain 'Users', got: %s", output)
	}
	if !strings.Contains(output, "List Users") {
		t.Errorf("expected output to contain 'List Users', got: %s", output)
	}
}

func TestCollectionCreate_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"collection", "create", "new-collection", "--collection-dir", dir})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created.
	path := filepath.Join(dir, "new-collection.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected collection file to exist")
	}

	// Verify it can be read back.
	loaded, err := store.ReadCollection(path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if loaded.Name != "new-collection" {
		t.Errorf("expected name 'new-collection', got %q", loaded.Name)
	}
}

func TestCollectionAdd_AddsRequest(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewFilesystem()

	// Create an initial collection.
	col := domain.Collection{
		Name: "my-api",
		Requests: []domain.SavedRequest{
			{
				Name: "Health Check",
				Config: domain.RequestConfig{
					Method: domain.MethodGet,
					URL:    "https://api.example.com/health",
				},
			},
		},
	}
	if err := store.WriteCollection(filepath.Join(dir, "my-api.yaml"), col); err != nil {
		t.Fatal(err)
	}

	mock := &mockHTTPClient{doFunc: noopDoFunc}
	a := newTestAppWithStorage(mock, store)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"collection", "add", "my-api", "Create User",
		"--method", "POST",
		"--url", "https://api.example.com/users",
		"--collection-dir", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the request was added.
	loaded, err := store.ReadCollection(filepath.Join(dir, "my-api.yaml"))
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	if len(loaded.Requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(loaded.Requests))
	}

	// Original request should still exist.
	if loaded.Requests[0].Name != "Health Check" {
		t.Errorf("expected first request 'Health Check', got %q", loaded.Requests[0].Name)
	}

	// New request should be added.
	if loaded.Requests[1].Name != "Create User" {
		t.Errorf("expected second request 'Create User', got %q", loaded.Requests[1].Name)
	}
	if loaded.Requests[1].Config.Method != domain.MethodPost {
		t.Errorf("expected method POST, got %q", loaded.Requests[1].Config.Method)
	}
	if loaded.Requests[1].Config.URL != "https://api.example.com/users" {
		t.Errorf("expected URL 'https://api.example.com/users', got %q", loaded.Requests[1].Config.URL)
	}
}
