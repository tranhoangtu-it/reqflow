package request

import (
	"encoding/json"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// BuildRequest transforms a RequestConfig into a fully resolved HTTPRequest,
// applying variable substitution and auto-detecting content-type.
func BuildRequest(config domain.RequestConfig, vars map[string]string) (domain.HTTPRequest, error) {
	req := domain.HTTPRequest{
		Method: config.Method,
		URL:    substituteVars(config.URL, vars),
		Body:   config.Body,
	}

	// Copy and substitute headers.
	if len(config.Headers) > 0 {
		req.Headers = make([]domain.Header, len(config.Headers))
		for i, h := range config.Headers {
			req.Headers[i] = domain.Header{
				Key:   h.Key,
				Value: substituteVars(h.Value, vars),
			}
		}
	}

	// Copy and substitute query params.
	if len(config.QueryParams) > 0 {
		req.QueryParams = make([]domain.QueryParam, len(config.QueryParams))
		for i, qp := range config.QueryParams {
			req.QueryParams[i] = domain.QueryParam{
				Key:   qp.Key,
				Value: substituteVars(qp.Value, vars),
			}
		}
	}

	// Resolve content-type: explicit takes precedence, otherwise auto-detect.
	req.ContentType = config.ContentType
	if req.ContentType == "" && len(config.Body) > 0 {
		if isJSON(config.Body) {
			req.ContentType = "application/json"
		}
	}

	return req, nil
}

// substituteVars replaces all {{varName}} placeholders in s with values from vars.
// Missing variables are left as-is.
func substituteVars(s string, vars map[string]string) string {
	if vars == nil || !strings.Contains(s, "{{") {
		return s
	}

	result := s
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// isJSON reports whether data looks like valid JSON.
func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
