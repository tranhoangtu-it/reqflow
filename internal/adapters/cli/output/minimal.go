package output

import (
	"fmt"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// MinimalFormatter outputs status code + body only.
type MinimalFormatter struct{}

func (f *MinimalFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	fmt.Fprintln(w, resp.Status)
	if len(resp.Body) > 0 {
		fmt.Fprintln(w, string(resp.Body))
	}
	return nil
}
