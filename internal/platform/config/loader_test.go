package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/platform/config"
)

func TestLoad_ReturnsDefaultsWhenNoConfigFilesExist(t *testing.T) {
	// Use a temp dir with no config files.
	tmpDir := t.TempDir()

	cfg, err := config.Load(
		config.WithGlobalDir(filepath.Join(tmpDir, "global")),
		config.WithProjectDir(filepath.Join(tmpDir, "project")),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.Output != domain.OutputPretty {
		t.Errorf("expected default output 'pretty', got %q", cfg.Output)
	}
	if cfg.NoColor != false {
		t.Errorf("expected default no_color false, got %v", cfg.NoColor)
	}
	if len(cfg.DefaultHeaders) != 0 {
		t.Errorf("expected no default headers, got %v", cfg.DefaultHeaders)
	}
}

func TestLoadFile_ParsesValidYAMLConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	content := `default_output: json
default_timeout: 10s
no_color: true
default_headers:
  User-Agent: "reqflow/0.1"
  Accept: "application/json"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.LoadFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
	if cfg.Output != domain.OutputJSON {
		t.Errorf("expected output 'json', got %q", cfg.Output)
	}
	if cfg.NoColor != true {
		t.Errorf("expected no_color true, got %v", cfg.NoColor)
	}

	headerMap := make(map[string]string)
	for _, h := range cfg.DefaultHeaders {
		headerMap[h.Key] = h.Value
	}
	if headerMap["User-Agent"] != "reqflow/0.1" {
		t.Errorf("expected User-Agent header 'reqflow/0.1', got %q", headerMap["User-Agent"])
	}
	if headerMap["Accept"] != "application/json" {
		t.Errorf("expected Accept header 'application/json', got %q", headerMap["Accept"])
	}
}

func TestLoadFile_ReturnsErrorForInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	content := `invalid: yaml: [broken`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := config.LoadFile(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_ProjectLocalOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_output: pretty
default_timeout: 30s
no_color: false
`
	projectContent := `default_output: json
default_timeout: 10s
`

	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".reqflow.yaml"), []byte(projectContent), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := config.Load(
		config.WithGlobalDir(globalDir),
		config.WithProjectDir(projectDir),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Output != domain.OutputJSON {
		t.Errorf("expected project-local output 'json' to override global, got %q", cfg.Output)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected project-local timeout 10s to override global, got %v", cfg.Timeout)
	}
}

func TestLoad_EmptyFieldsInProjectLocalDontOverrideGlobal(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_output: json
default_timeout: 30s
no_color: true
`
	// Project-local only sets output, leaving timeout and no_color unset.
	projectContent := `default_output: raw
`

	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".reqflow.yaml"), []byte(projectContent), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := config.Load(
		config.WithGlobalDir(globalDir),
		config.WithProjectDir(projectDir),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Output != domain.OutputRaw {
		t.Errorf("expected project output 'raw', got %q", cfg.Output)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected global timeout 30s preserved, got %v", cfg.Timeout)
	}
	if cfg.NoColor != true {
		t.Errorf("expected global no_color true preserved, got %v", cfg.NoColor)
	}
}

func TestLoad_DefaultHeadersMerged(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_headers:
  User-Agent: "reqflow/0.1"
  Accept: "application/json"
`
	projectContent := `default_headers:
  X-Project: "my-api"
  Accept: "text/html"
`

	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".reqflow.yaml"), []byte(projectContent), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := config.Load(
		config.WithGlobalDir(globalDir),
		config.WithProjectDir(projectDir),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headerMap := make(map[string]string)
	for _, h := range cfg.DefaultHeaders {
		headerMap[h.Key] = h.Value
	}

	// Global header preserved.
	if headerMap["User-Agent"] != "reqflow/0.1" {
		t.Errorf("expected User-Agent 'reqflow/0.1', got %q", headerMap["User-Agent"])
	}
	// Project-local header added.
	if headerMap["X-Project"] != "my-api" {
		t.Errorf("expected X-Project 'my-api', got %q", headerMap["X-Project"])
	}
	// Project-local overrides global for same key.
	if headerMap["Accept"] != "text/html" {
		t.Errorf("expected Accept 'text/html' (project override), got %q", headerMap["Accept"])
	}
}

func TestSaveFile_WritesValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &domain.AppConfig{
		Timeout: 15 * time.Second,
		Output:  domain.OutputJSON,
		NoColor: true,
		DefaultHeaders: []domain.Header{
			{Key: "User-Agent", Value: "reqflow/0.1"},
		},
	}

	if err := config.SaveFile(cfgPath, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Re-load and verify round-trip.
	loaded, err := config.LoadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload saved config: %v", err)
	}

	if loaded.Timeout != 15*time.Second {
		t.Errorf("expected timeout 15s, got %v", loaded.Timeout)
	}
	if loaded.Output != domain.OutputJSON {
		t.Errorf("expected output 'json', got %q", loaded.Output)
	}
	if loaded.NoColor != true {
		t.Errorf("expected no_color true, got %v", loaded.NoColor)
	}
	if len(loaded.DefaultHeaders) != 1 || loaded.DefaultHeaders[0].Key != "User-Agent" {
		t.Errorf("expected User-Agent header, got %v", loaded.DefaultHeaders)
	}
}
