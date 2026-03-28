package workflow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// operators recognized by the assertion parser, ordered longest-first so
// two-char operators are matched before single-char ones.
var assertionOperators = []string{"==", "!=", "contains", "<", ">"}

// ParseAssertionString parses a human-readable assertion string into a
// domain.Assertion. Supported formats:
//
//	"status == 200"
//	"body.name == John"
//	"body contains 'hello'"
//	"header.Content-Type == application/json"
//	"duration < 500"
func ParseAssertionString(s string) (domain.Assertion, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return domain.Assertion{}, fmt.Errorf("empty assertion string")
	}

	field, operator, rawExpected, err := splitAssertion(s)
	if err != nil {
		return domain.Assertion{}, err
	}

	expected := parseExpectedValue(rawExpected)

	return domain.Assertion{
		Field:    field,
		Operator: operator,
		Expected: expected,
	}, nil
}

// splitAssertion splits "field operator value" into its three components.
func splitAssertion(s string) (field, operator, value string, err error) {
	for _, op := range assertionOperators {
		// Look for " op " (surrounded by spaces) to avoid partial matches.
		needle := " " + op + " "
		idx := strings.Index(s, needle)
		if idx >= 0 {
			field = strings.TrimSpace(s[:idx])
			value = strings.TrimSpace(s[idx+len(needle):])
			return field, op, value, nil
		}
	}

	return "", "", "", fmt.Errorf("invalid assertion format: %q (no recognized operator)", s)
}

// parseExpectedValue converts the raw expected string to an appropriate Go
// type. Quoted strings have their quotes stripped; numeric strings become int.
func parseExpectedValue(raw string) interface{} {
	// Strip surrounding single quotes.
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		return raw[1 : len(raw)-1]
	}

	// Try integer conversion.
	if n, err := strconv.Atoi(raw); err == nil {
		return n
	}

	return raw
}
