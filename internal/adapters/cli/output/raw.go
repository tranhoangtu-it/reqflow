package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// RawFormatter outputs only the response body with no decoration.
type RawFormatter struct{}

func (f *RawFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	if len(resp.Body) == 0 {
		return nil
	}
	_, err := w.Write(resp.Body)
	return err
}
