package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// PrettyFormatter outputs colored, human-readable HTTP responses.
type PrettyFormatter struct {
	noColor bool
}

func (f *PrettyFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	return nil
}
