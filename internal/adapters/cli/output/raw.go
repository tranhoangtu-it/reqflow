package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// RawFormatter outputs only the response body.
type RawFormatter struct{}

func (f *RawFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	return nil
}
