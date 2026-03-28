package output

import (
	"fmt"
	"io"

	"github.com/ye-kart/reqflow/internal/domain"
)

func green(s string) string  { return "\033[32m" + s + "\033[0m" }
func yellow(s string) string { return "\033[33m" + s + "\033[0m" }
func red(s string) string    { return "\033[31m" + s + "\033[0m" }

// PrettyFormatter outputs colored, human-readable HTTP responses.
type PrettyFormatter struct {
	noColor bool
}

func (f *PrettyFormatter) FormatResponse(w io.Writer, resp domain.HTTPResponse) error {
	// Status line
	statusLine := fmt.Sprintf("HTTP/1.1 %s", resp.Status)
	if !f.noColor {
		statusLine = f.colorizeStatus(statusLine, resp.StatusCode)
	}
	fmt.Fprintln(w, statusLine)

	// Headers
	for _, h := range resp.Headers {
		fmt.Fprintf(w, "%s: %s\n", h.Key, h.Value)
	}

	// Body
	if len(resp.Body) > 0 {
		fmt.Fprintln(w) // blank separator line
		body := prettyPrintBody(resp.Body)
		fmt.Fprintln(w, body)
	}

	// Duration
	durationMs := resp.Duration.Milliseconds()
	fmt.Fprintf(w, "\n(took %dms)\n", durationMs)

	return nil
}

func (f *PrettyFormatter) colorizeStatus(s string, code int) string {
	switch {
	case code >= 200 && code < 300:
		return green(s)
	case code >= 300 && code < 400:
		return yellow(s)
	case code >= 400:
		return red(s)
	default:
		return s
	}
}
