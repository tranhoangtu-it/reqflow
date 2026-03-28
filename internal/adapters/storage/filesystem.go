package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
	"gopkg.in/yaml.v3"
)

// envFile is the YAML-serializable representation of an environment file.
type envFile struct {
	Name      string            `yaml:"name"`
	Variables map[string]string `yaml:"variables"`
}

// Filesystem implements driven.Storage using the local filesystem.
type Filesystem struct{}

// NewFilesystem creates a new Filesystem storage adapter.
func NewFilesystem() *Filesystem {
	return &Filesystem{}
}

// ReadEnvironment parses a YAML environment file and returns a domain.Environment.
func (f *Filesystem) ReadEnvironment(path string) (domain.Environment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Environment{}, fmt.Errorf("reading environment file: %w", err)
	}

	var ef envFile
	if err := yaml.Unmarshal(data, &ef); err != nil {
		return domain.Environment{}, fmt.Errorf("parsing environment YAML: %w", err)
	}

	env := domain.Environment{
		Name: ef.Name,
	}

	// Sort keys for deterministic ordering.
	keys := make([]string, 0, len(ef.Variables))
	for k := range ef.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		env.Variables = append(env.Variables, domain.Variable{
			Key:   k,
			Value: ef.Variables[k],
			Scope: domain.ScopeEnvironment,
		})
	}

	return env, nil
}

// WriteEnvironment writes a domain.Environment to a YAML file.
func (f *Filesystem) WriteEnvironment(path string, env domain.Environment) error {
	ef := envFile{
		Name:      env.Name,
		Variables: make(map[string]string, len(env.Variables)),
	}

	for _, v := range env.Variables {
		ef.Variables[v.Key] = v.Value
	}

	data, err := yaml.Marshal(&ef)
	if err != nil {
		return fmt.Errorf("marshaling environment YAML: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing environment file: %w", err)
	}

	return nil
}

// ListEnvironments returns the names (without .yaml extension) of all environment
// files in the given directory.
func (f *Filesystem) ListEnvironments(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			names = append(names, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
		}
	}

	return names, nil
}

// ReadFile reads raw bytes from a file.
func (f *Filesystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes raw bytes to a file.
func (f *Filesystem) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
