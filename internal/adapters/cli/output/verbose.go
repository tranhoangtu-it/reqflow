package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
)

func cyan(s string) string { return "\033[36m" + s + "\033[0m" }

// FormatVerbose writes a verbose representation of the HTTP request and response
// to w. Request lines are prefixed with "> " and response lines with "< ".
// When noColor is false, request lines are colored cyan and response lines yellow.
func FormatVerbose(w io.Writer, req domain.HTTPRequest, resp domain.HTTPResponse, noColor bool) error {
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

	// Blank request separator
	fmt.Fprintln(w, ">")

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
		body := formatBody(resp.Body)
		fmt.Fprintln(w, body)
	}

	return nil
}

func formatBody(body []byte) string {
	var prettyJSON bytes.Buffer
	if json.Indent(&prettyJSON, body, "", "  ") == nil {
		return prettyJSON.String()
	}
	return string(body)
}
