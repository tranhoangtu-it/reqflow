package workflow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// EvaluateCondition evaluates a JSONPath condition against a JSON body.
// Conditions have the form: "$.field == 'value'", "$.field > 5", etc.
// Supported operators: ==, !=, <, >, <=, >=
// String values are in single quotes, numbers without quotes.
// Returns (true, nil) when condition is met, (false, nil) when not met,
// and (false, error) on parse error.
func EvaluateCondition(body []byte, condition string) (bool, error) {
	path, operator, expected, err := parseCondition(condition)
	if err != nil {
		return false, err
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return false, fmt.Errorf("parsing JSON body: %w", err)
	}

	actual, err := extractPath(data, path)
	if err != nil {
		// Missing field returns false, not error
		return false, nil
	}

	return compareCondition(actual, expected, operator), nil
}

// parseCondition parses a condition string like "$.status == 'completed'" into
// its components: JSONPath, operator, and expected value.
func parseCondition(condition string) (path string, operator string, expected interface{}, err error) {
	condition = strings.TrimSpace(condition)

	operators := []string{"<=", ">=", "!=", "==", "<", ">"}
	var opIdx int
	var foundOp string

	for _, op := range operators {
		idx := strings.Index(condition, " "+op+" ")
		if idx >= 0 {
			opIdx = idx
			foundOp = op
			break
		}
	}

	if foundOp == "" {
		return "", "", nil, fmt.Errorf("invalid condition syntax: no operator found in %q", condition)
	}

	path = strings.TrimSpace(condition[:opIdx])
	valueStr := strings.TrimSpace(condition[opIdx+len(foundOp)+2:])

	if path == "" || valueStr == "" {
		return "", "", nil, fmt.Errorf("invalid condition syntax: empty path or value in %q", condition)
	}

	// Parse the expected value
	expected, err = parseConditionValue(valueStr)
	if err != nil {
		return "", "", nil, fmt.Errorf("invalid condition value %q: %w", valueStr, err)
	}

	return path, foundOp, expected, nil
}

// parseConditionValue parses a value from a condition expression.
// Single-quoted strings return as string, unquoted values try numeric first.
func parseConditionValue(s string) (interface{}, error) {
	// String value in single quotes
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") && len(s) >= 2 {
		return s[1 : len(s)-1], nil
	}

	// Try numeric
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("unrecognized value format: %s", s)
}

// compareCondition compares an actual JSON value against an expected value.
func compareCondition(actual, expected interface{}, operator string) bool {
	switch operator {
	case "==":
		return conditionValuesEqual(actual, expected)
	case "!=":
		return !conditionValuesEqual(actual, expected)
	case "<":
		a, b, ok := toNumericPair(actual, expected)
		return ok && a < b
	case ">":
		a, b, ok := toNumericPair(actual, expected)
		return ok && a > b
	case "<=":
		a, b, ok := toNumericPair(actual, expected)
		return ok && a <= b
	case ">=":
		a, b, ok := toNumericPair(actual, expected)
		return ok && a >= b
	default:
		return false
	}
}

// conditionValuesEqual checks equality between an actual JSON value and expected.
func conditionValuesEqual(actual, expected interface{}) bool {
	// If both can be numeric, compare as numbers
	a, b, ok := toNumericPair(actual, expected)
	if ok {
		return a == b
	}

	// Fall back to string comparison
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// toNumericPair attempts to convert both values to float64.
func toNumericPair(a, b interface{}) (float64, float64, bool) {
	af, aOk := toConditionFloat(a)
	bf, bOk := toConditionFloat(b)
	if aOk && bOk {
		return af, bf, true
	}
	return 0, 0, false
}

// toConditionFloat converts a value to float64 if possible.
func toConditionFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}
