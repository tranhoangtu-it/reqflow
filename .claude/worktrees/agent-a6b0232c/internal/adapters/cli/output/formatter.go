package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// Formatter renders an HTTP response to a writer.
type Formatter interface {
	FormatResponse(w io.Writer, resp domain.HTTPResponse) error
}

// New returns the appropriate Formatter for the given output format.
// Defaults to PrettyFormatter for unrecognized formats.
func New(format domain.OutputFormat, noColor bool) Formatter {
	switch format {
	case domain.OutputJSON:
		return &JSONFormatter{}
	case domain.OutputRaw:
		return &RawFormatter{}
	case domain.OutputMinimal:
		return &MinimalFormatter{}
	default:
		return &PrettyFormatter{noColor: noColor}
	}
}
