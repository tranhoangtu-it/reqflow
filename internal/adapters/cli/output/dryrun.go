package output

import (
	"fmt"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// FormatDryRun writes the request that would be sent without actually sending it.
// Only request lines (with > prefix) are shown, prefixed by a DRY RUN indicator.
func FormatDryRun(w io.Writer, req domain.HTTPRequest, noColor bool) error {
	indicator := "DRY RUN"
	if !noColor {
		indicator = bold(yellow(indicator))
	}
	fmt.Fprintf(w, "[ %s ]\n", indicator)

	return writeRequestLines(w, req, noColor)
}
