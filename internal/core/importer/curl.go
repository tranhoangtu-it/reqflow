package importer

import (
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ParseCurl parses a cURL command string into a domain.RequestConfig.
// It handles common curl flags and returns an error for invalid input.
func ParseCurl(curlCmd string) (domain.RequestConfig, error) {
	// Normalize: remove backslash line continuations.
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.TrimSpace(curlCmd)

	if curlCmd == "" {
		return domain.RequestConfig{}, fmt.Errorf("empty curl command")
	}

	// Strip the "curl" prefix if present.
	tokens := tokenize(curlCmd)
	if len(tokens) > 0 && tokens[0] == "curl" {
		tokens = tokens[1:]
	}

	var (
		config     domain.RequestConfig
		methodSet  bool
		hasData    bool
		url        string
	)

	config.Method = domain.MethodGet

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		switch tok {
		case "-X", "--request":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			config.Method = domain.HTTPMethod(strings.ToUpper(tokens[i]))
			methodSet = true

		case "-H", "--header":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			key, value, ok := parseHeaderValue(tokens[i])
			if ok {
				config.Headers = append(config.Headers, domain.Header{Key: key, Value: value})
			}

		case "-d", "--data", "--data-raw":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			config.Body = []byte(tokens[i])
			hasData = true

		case "-u", "--user":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			parts := strings.SplitN(tokens[i], ":", 2)
			if len(parts) == 2 {
				config.Auth = &domain.AuthConfig{
					Type: domain.AuthBasic,
					Basic: &domain.BasicAuthConfig{
						Username: parts[0],
						Password: parts[1],
					},
				}
			}

		case "-A", "--user-agent":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			config.Headers = append(config.Headers, domain.Header{Key: "User-Agent", Value: tokens[i]})

		case "-b", "--cookie":
			i++
			if i >= len(tokens) {
				return domain.RequestConfig{}, fmt.Errorf("missing value for %s", tok)
			}
			config.Headers = append(config.Headers, domain.Header{Key: "Cookie", Value: tokens[i]})

		case "--compressed":
			config.Headers = append(config.Headers, domain.Header{Key: "Accept-Encoding", Value: "gzip, deflate, br"})

		case "-L", "--location", "-k", "--insecure":
			// Flags stored but not critical for now; skip.

		default:
			// Not a flag — treat as URL if it looks like one.
			if isURL(tok) {
				url = tok
			}
		}
	}

	if url == "" {
		return domain.RequestConfig{}, fmt.Errorf("no URL found in curl command")
	}

	config.URL = url

	// If -d was used but no explicit method, default to POST.
	if hasData && !methodSet {
		config.Method = domain.MethodPost
	}

	return config, nil
}

// tokenize splits a curl command string into tokens, respecting single and
// double quoted strings.
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle

		case ch == '"' && !inSingle:
			inDouble = !inDouble

		case ch == '\\' && inDouble && i+1 < len(runes):
			// Handle escaped characters inside double quotes.
			next := runes[i+1]
			switch next {
			case '"', '\\':
				current.WriteRune(next)
				i++
			default:
				current.WriteRune(ch)
				current.WriteRune(next)
				i++
			}

		case (ch == ' ' || ch == '\t') && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// parseHeaderValue splits a "Key: Value" string into key and value.
func parseHeaderValue(s string) (key, value string, ok bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:]), true
}

// isURL checks if the token looks like a URL.
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
