package workflow

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// EvaluateAssertions evaluates a list of assertions against an HTTP response.
// Each assertion is evaluated independently and results are returned for all.
func EvaluateAssertions(assertions []domain.Assertion, resp domain.HTTPResponse) []domain.AssertionResult {
	results := make([]domain.AssertionResult, 0, len(assertions))

	for _, a := range assertions {
		result := evaluateOne(a, resp)
		results = append(results, result)
	}

	return results
}

func evaluateOne(a domain.Assertion, resp domain.HTTPResponse) domain.AssertionResult {
	switch {
	case a.Field == "status":
		return evaluateStatus(a, resp)
	case a.Field == "body":
		return evaluateBodyString(a, resp)
	case strings.HasPrefix(a.Field, "body."):
		return evaluateBodyField(a, resp)
	case strings.HasPrefix(a.Field, "header."):
		return evaluateHeader(a, resp)
	default:
		return domain.AssertionResult{
			Assertion: a,
			Passed:    false,
			Message:   fmt.Sprintf("unknown field type: %s", a.Field),
		}
	}
}

func evaluateStatus(a domain.Assertion, resp domain.HTTPResponse) domain.AssertionResult {
	actual := resp.StatusCode
	expected := toFloat64(a.Expected)

	passed := compareNumeric(float64(actual), expected, a.Operator)

	msg := ""
	if !passed {
		msg = fmt.Sprintf("status: expected %v %s %v, got %d", a.Field, a.Operator, a.Expected, actual)
	}

	return domain.AssertionResult{
		Assertion: a,
		Passed:    passed,
		Actual:    actual,
		Message:   msg,
	}
}

func evaluateBodyString(a domain.Assertion, resp domain.HTTPResponse) domain.AssertionResult {
	bodyStr := string(resp.Body)

	switch a.Operator {
	case "contains":
		expected := fmt.Sprintf("%v", a.Expected)
		passed := strings.Contains(bodyStr, expected)
		msg := ""
		if !passed {
			msg = fmt.Sprintf("body does not contain %q", expected)
		}
		return domain.AssertionResult{
			Assertion: a,
			Passed:    passed,
			Actual:    bodyStr,
			Message:   msg,
		}
	default:
		return domain.AssertionResult{
			Assertion: a,
			Passed:    false,
			Message:   fmt.Sprintf("unsupported operator %q for body", a.Operator),
		}
	}
}

func evaluateBodyField(a domain.Assertion, resp domain.HTTPResponse) domain.AssertionResult {
	// Convert "body.field.path" to "$.field.path" for extraction
	jsonPath := "$." + strings.TrimPrefix(a.Field, "body.")

	if a.Operator == "exists" {
		return evaluateBodyExists(a, resp, jsonPath)
	}

	val, err := extractPath(parseJSON(resp.Body), jsonPath)
	if err != nil {
		return domain.AssertionResult{
			Assertion: a,
			Passed:    false,
			Message:   fmt.Sprintf("failed to extract %s: %v", a.Field, err),
		}
	}

	passed := compareValues(val, a.Expected, a.Operator)
	msg := ""
	if !passed {
		msg = fmt.Sprintf("%s: expected %v %s %v, got %v", a.Field, a.Expected, a.Operator, a.Expected, val)
	}

	return domain.AssertionResult{
		Assertion: a,
		Passed:    passed,
		Actual:    val,
		Message:   msg,
	}
}

func evaluateBodyExists(a domain.Assertion, resp domain.HTTPResponse, jsonPath string) domain.AssertionResult {
	_, err := extractPath(parseJSON(resp.Body), jsonPath)
	passed := err == nil

	msg := ""
	if !passed {
		msg = fmt.Sprintf("%s does not exist", a.Field)
	}

	return domain.AssertionResult{
		Assertion: a,
		Passed:    passed,
		Message:   msg,
	}
}

func evaluateHeader(a domain.Assertion, resp domain.HTTPResponse) domain.AssertionResult {
	headerName := strings.TrimPrefix(a.Field, "header.")
	headerVal := ""
	found := false
	for _, h := range resp.Headers {
		if strings.EqualFold(h.Key, headerName) {
			headerVal = h.Value
			found = true
			break
		}
	}

	if a.Operator == "exists" {
		msg := ""
		if !found {
			msg = fmt.Sprintf("header %s does not exist", headerName)
		}
		return domain.AssertionResult{
			Assertion: a,
			Passed:    found,
			Actual:    headerVal,
			Message:   msg,
		}
	}

	if !found {
		return domain.AssertionResult{
			Assertion: a,
			Passed:    false,
			Actual:    nil,
			Message:   fmt.Sprintf("header %s not found", headerName),
		}
	}

	passed := compareValues(headerVal, a.Expected, a.Operator)
	msg := ""
	if !passed {
		msg = fmt.Sprintf("header %s: expected %v %s %v, got %q", headerName, a.Expected, a.Operator, a.Expected, headerVal)
	}

	return domain.AssertionResult{
		Assertion: a,
		Passed:    passed,
		Actual:    headerVal,
		Message:   msg,
	}
}

func parseJSON(body []byte) interface{} {
	var data interface{}
	json.Unmarshal(body, &data)
	return data
}

func compareValues(actual, expected interface{}, operator string) bool {
	switch operator {
	case "==":
		return valuesEqual(actual, expected)
	case "!=":
		return !valuesEqual(actual, expected)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "<":
		a, b := toFloat64(actual), toFloat64(expected)
		return a < b
	case ">":
		a, b := toFloat64(actual), toFloat64(expected)
		return a > b
	default:
		return false
	}
}

func compareNumeric(actual, expected float64, operator string) bool {
	switch operator {
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	case "<":
		return actual < expected
	case ">":
		return actual > expected
	default:
		return false
	}
}

func valuesEqual(a, b interface{}) bool {
	// Normalize numeric types: JSON numbers are float64, YAML integers are int
	af := toFloat64OrNil(a)
	bf := toFloat64OrNil(b)
	if af != nil && bf != nil {
		return *af == *bf
	}

	// Compare booleans directly
	ab, aIsBool := a.(bool)
	bb, bIsBool := b.(bool)
	if aIsBool && bIsBool {
		return ab == bb
	}

	// Fall back to string comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func toFloat64OrNil(v interface{}) *float64 {
	switch val := v.(type) {
	case float64:
		return &val
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	default:
		return nil
	}
}
