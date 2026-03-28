package output

import (
	"fmt"
	"io"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
)

func bold(s string) string { return "\033[1m" + s + "\033[0m" }

// FormatDryRun writes the request that would be sent without actually sending it.
// Only request lines (with > prefix) are shown, prefixed by a DRY RUN indicator.
func FormatDryRun(w io.Writer, req domain.HTTPRequest, noColor bool) error {
	indicator := "DRY RUN"
	if !noColor {
		indicator = bold(yellow(indicator))
	}
	fmt.Fprintf(w, "[ %s ]\n", indicator)

	u, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("parsing request URL: %w", err)
	}

	path := u.RequestURI()

	// Request line
	reqLine := fmt.Sprintf("> %s %s HTTP/1.1", req.Method, path)
	hostLine := fmt.Sprintf("> Host: %s", u.Host)

	if !noColor {
		reqLine = cyan(reqLine)
		hostLine = cyan(hostLine)
	}
	fmt.Fprintln(w, reqLine)
	fmt.Fprintln(w, hostLine)

	// Request headers
	for _, h := range req.Headers {
		line := fmt.Sprintf("> %s: %s", h.Key, h.Value)
		if !noColor {
			line = cyan(line)
		}
		fmt.Fprintln(w, line)
	}

	fmt.Fprintln(w, ">")

	return nil
}
