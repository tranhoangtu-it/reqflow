package domain

import "time"

// OutputFormat determines how responses are displayed.
type OutputFormat string

const (
	OutputPretty  OutputFormat = "pretty"
	OutputJSON    OutputFormat = "json"
	OutputRaw     OutputFormat = "raw"
	OutputMinimal OutputFormat = "minimal"
)

// AppConfig holds global application configuration.
type AppConfig struct {
	Timeout  time.Duration
	DataDir  string
	LogLevel string
	NoColor  bool
	Output   OutputFormat
}
