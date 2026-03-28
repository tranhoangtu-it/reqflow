package driven

import "github.com/ye-kart/reqflow/internal/domain"

// Storage is the driven port for reading and writing collections, environments,
// and other persistent data.
type Storage interface {
	ReadEnvironment(path string) (domain.Environment, error)
	WriteEnvironment(path string, env domain.Environment) error
	ListEnvironments(dir string) ([]string, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
}
