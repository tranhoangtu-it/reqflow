package output

import (
	"fmt"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

// FormatVerbose writes a verbose representation of the HTTP request and response
// to w. Request lines are prefixed with "> " and response lines with "< ".
// When noColor is false, request lines are colored cyan and response lines yellow.
func FormatVerbose(w io.Writer, req domain.HTTPRequest, resp domain.HTTPResponse, noColor bool) error {
	if err := writeRequestLines(w, req, noColor); err != nil {
		return err
	}

	// Response status line
	statusLine := fmt.Sprintf("< HTTP/1.1 %s", resp.Status)
	if !noColor {
		statusLine = yellow(statusLine)
	}
	fmt.Fprintln(w, statusLine)

	// Response headers
	for _, h := range resp.Headers {
		line := fmt.Sprintf("< %s: %s", h.Key, h.Value)
		if !noColor {
			line = yellow(line)
		}
		fmt.Fprintln(w, line)
	}

	// Blank response separator
	fmt.Fprintln(w, "<")

	// Body
	if len(resp.Body) > 0 {
		body := prettyPrintBody(resp.Body)
		fmt.Fprintln(w, body)
	}

	return nil
}
