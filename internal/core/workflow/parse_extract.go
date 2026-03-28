package workflow

import (
	"fmt"
	"strings"
)

// ParseExtractString parses an inline extraction expression.
// Supported formats:
//   - "$.field"          → ("", "$.field", nil) — bare JSONPath, no label
//   - "name=$.field"     → ("name", "$.field", nil) — labeled extraction
//   - ""                 → error
func ParseExtractString(s string) (varName string, jsonPath string, err error) {
	if s == "" {
		return "", "", fmt.Errorf("empty extract expression")
	}

	// If the string starts with "$.", it's a bare JSONPath.
	if strings.HasPrefix(s, "$.") {
		return "", s, nil
	}

	// Otherwise, expect "name=$.path" format.
	idx := strings.Index(s, "=")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid extract expression: %q (expected $.path or name=$.path)", s)
	}

	varName = s[:idx]
	jsonPath = s[idx+1:]

	if varName == "" || jsonPath == "" {
		return "", "", fmt.Errorf("invalid extract expression: %q (name and path must not be empty)", s)
	}

	return varName, jsonPath, nil
}
