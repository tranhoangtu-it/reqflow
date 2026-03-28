package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// JSONFormatter outputs structured JSON representation of HTTP responses.
type JSONFormatter struct{}

func (f *JSONFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	return nil
}
