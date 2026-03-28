package workflow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ExtractValues extracts values from a JSON body using simple dot-notation
// JSONPath expressions. Each expression maps a variable name to a path like
// $.field.nested or $.items[0].id. All extracted values are returned as strings.
func ExtractValues(body []byte, expressions map[string]string) (map[string]string, error) {
	if len(expressions) == 0 {
		return map[string]string{}, nil
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing JSON body: %w", err)
	}

	result := make(map[string]string, len(expressions))
	for varName, expr := range expressions {
		val, err := extractPath(data, expr)
		if err != nil {
			return nil, fmt.Errorf("extracting %q (%s): %w", varName, expr, err)
		}
		result[varName] = toString(val)
	}

	return result, nil
}

// extractPath navigates the parsed JSON using a dot-notation path.
// Supports: $.field, $.field.nested, $.items[0], $.items[0].field
func extractPath(data interface{}, path string) (interface{}, error) {
	// Strip leading "$." prefix
	path = strings.TrimPrefix(path, "$.")
	if path == "$" || path == "" {
		return data, nil
	}

	segments := splitPath(path)
	current := data

	for _, seg := range segments {
		fieldName, index, hasIndex := parseSegment(seg)

		if fieldName != "" {
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at %q, got %T", fieldName, current)
			}
			val, exists := obj[fieldName]
			if !exists {
				return nil, fmt.Errorf("field %q not found", fieldName)
			}
			current = val
		}

		if hasIndex {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array for index [%d], got %T", index, current)
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("index %d out of bounds (length %d)", index, len(arr))
			}
			current = arr[index]
		}
	}

	return current, nil
}

// splitPath splits a dot-notation path into segments, respecting array brackets.
// E.g., "items[0].id" -> ["items[0]", "id"]
func splitPath(path string) []string {
	var segments []string
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		ch := path[i]
		if ch == '.' {
			if current.Len() > 0 {
				segments = append(segments, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// parseSegment parses a path segment like "field", "field[0]", or "[0]".
// Returns the field name (may be empty for bare index), the index (if any),
// and whether an index was present.
func parseSegment(seg string) (string, int, bool) {
	bracketStart := strings.Index(seg, "[")
	if bracketStart < 0 {
		return seg, 0, false
	}

	bracketEnd := strings.Index(seg, "]")
	if bracketEnd < 0 || bracketEnd <= bracketStart+1 {
		return seg, 0, false
	}

	fieldName := seg[:bracketStart]
	indexStr := seg[bracketStart+1 : bracketEnd]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return seg, 0, false
	}

	return fieldName, index, true
}

// toString converts a JSON value to its string representation.
func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return "null"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
