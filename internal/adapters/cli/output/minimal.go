package output

import (
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// MinimalFormatter outputs status code + body only.
type MinimalFormatter struct{}

func (f *MinimalFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	return nil
}
