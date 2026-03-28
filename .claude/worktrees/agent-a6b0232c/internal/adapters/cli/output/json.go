package output

import (
	"encoding/json"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// JSONFormatter outputs structured JSON representation of HTTP responses.
type JSONFormatter struct{}

type jsonOutput struct {
	StatusCode int                    `json:"status_code"`
	Status     string                 `json:"status"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	DurationMs int64                  `json:"duration_ms"`
	Size       int64                  `json:"size"`
}

func (f *JSONFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	out := jsonOutput{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    flattenHeaders(resp.Headers),
		Body:       parseBody(resp.Body),
		DurationMs: resp.Duration.Milliseconds(),
		Size:       resp.Size,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func flattenHeaders(headers []domain.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for _, h := range headers {
		if _, exists := result[h.Key]; !exists {
			result[h.Key] = h.Value
		}
	}
	return result
}

func parseBody(body []byte) interface{} {
	if len(body) == 0 {
		return nil
	}

	var parsed interface{}
	if json.Unmarshal(body, &parsed) == nil {
		return parsed
	}
	return string(body)
}
