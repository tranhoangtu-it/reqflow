package exporter

import (
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ExportCurl generates a curl command string from a RequestConfig.
// The output uses multi-line format with backslash continuations when there
// are multiple flags for readability.
func ExportCurl(config domain.RequestConfig) string {
	var parts []string

	// Method (omit for GET since it's the default).
	if config.Method != domain.MethodGet {
		parts = append(parts, fmt.Sprintf("-X %s", config.Method))
	}

	// Auth.
	if config.Auth != nil {
		switch config.Auth.Type {
		case domain.AuthBasic:
			if config.Auth.Basic != nil {
				parts = append(parts, fmt.Sprintf("-u '%s:%s'", config.Auth.Basic.Username, config.Auth.Basic.Password))
			}
		case domain.AuthBearer:
			if config.Auth.Bearer != nil {
				prefix := "Bearer"
				if config.Auth.Bearer.Prefix != "" {
					prefix = config.Auth.Bearer.Prefix
				}
				parts = append(parts, fmt.Sprintf("-H 'Authorization: %s %s'", prefix, config.Auth.Bearer.Token))
			}
		}
	}

	// Headers.
	for _, h := range config.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", h.Key, h.Value))
	}

	// Body.
	if len(config.Body) > 0 {
		parts = append(parts, fmt.Sprintf("-d '%s'", string(config.Body)))
	}

	// Build the final command.
	// If there are no extra parts, just "curl URL".
	if len(parts) == 0 {
		return fmt.Sprintf("curl %s", config.URL)
	}

	// If there's only one part and it's a method flag, keep it on one line.
	if len(parts) == 1 && strings.HasPrefix(parts[0], "-X ") {
		return fmt.Sprintf("curl %s %s", parts[0], config.URL)
	}

	// Multi-line format: method on first line, rest with continuations.
	var sb strings.Builder
	sb.WriteString("curl")

	// Put method on the first line if present.
	startIdx := 0
	if len(parts) > 0 && strings.HasPrefix(parts[0], "-X ") {
		sb.WriteString(" ")
		sb.WriteString(parts[0])
		startIdx = 1
	}

	for i := startIdx; i < len(parts); i++ {
		sb.WriteString(" \\\n  ")
		sb.WriteString(parts[i])
	}
	sb.WriteString(" \\\n  ")
	sb.WriteString(config.URL)

	return sb.String()
}

