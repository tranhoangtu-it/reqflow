package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ANSI color helpers.
func cyan(s string) string   { return "\033[36m" + s + "\033[0m" }
func bold(s string) string   { return "\033[1m" + s + "\033[0m" }

// writeRequestLines writes the HTTP request in verbose format with > prefixes.
// It includes the request line, Host header, and all request headers.
func writeRequestLines(w io.Writer, req domain.HTTPRequest, noColor bool) error {
	u, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("parsing request URL: %w", err)
	}

	path := u.RequestURI()

	reqLine := fmt.Sprintf("> %s %s HTTP/1.1", req.Method, path)
	hostLine := fmt.Sprintf("> Host: %s", u.Host)

	if !noColor {
		reqLine = cyan(reqLine)
		hostLine = cyan(hostLine)
	}
	fmt.Fprintln(w, reqLine)
	fmt.Fprintln(w, hostLine)

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

// prettyPrintBody returns the body as pretty-printed JSON if possible,
// otherwise as a plain string.
func prettyPrintBody(body []byte) string {
	var prettyJSON bytes.Buffer
	if json.Indent(&prettyJSON, body, "", "  ") == nil {
		return prettyJSON.String()
	}
	return string(body)
}
