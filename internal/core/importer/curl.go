package importer

import (
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ParseCurl parses a cURL command string into a domain.RequestConfig.
// It handles common curl flags and returns an error for invalid input.
func ParseCurl(curlCmd string) (domain.RequestConfig, error) {
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.TrimSpace(curlCmd)

	if curlCmd == "" {
		return domain.RequestConfig{}, fmt.Errorf("empty curl command")
	}

	tokens := tokenize(curlCmd)
	if len(tokens) > 0 && tokens[0] == "curl" {
		tokens = tokens[1:]
	}

	p := &curlParser{tokens: tokens}
	return p.parse()
}

// curlParser holds state while iterating over tokenized curl arguments.
type curlParser struct {
	tokens    []string
	pos       int
	config    domain.RequestConfig
	methodSet bool
	hasData   bool
	url       string
}

func (p *curlParser) parse() (domain.RequestConfig, error) {
	p.config.Method = domain.MethodGet

	for p.pos = 0; p.pos < len(p.tokens); p.pos++ {
		tok := p.tokens[p.pos]

		var err error
		switch tok {
		case "-X", "--request":
			err = p.handleMethod(tok)
		case "-H", "--header":
			err = p.handleHeader(tok)
		case "-d", "--data", "--data-raw":
			err = p.handleData(tok)
		case "-u", "--user":
			err = p.handleUser(tok)
		case "-A", "--user-agent":
			err = p.handleUserAgent(tok)
		case "-b", "--cookie":
			err = p.handleCookie(tok)
		case "--compressed":
			p.config.Headers = append(p.config.Headers, domain.Header{
				Key: "Accept-Encoding", Value: "gzip, deflate, br",
			})
		case "-L", "--location", "-k", "--insecure":
			// Recognized but not critical; skip.
		default:
			if isURL(tok) {
				p.url = tok
			}
		}
		if err != nil {
			return domain.RequestConfig{}, err
		}
	}

	if p.url == "" {
		return domain.RequestConfig{}, fmt.Errorf("no URL found in curl command")
	}
	p.config.URL = p.url

	if p.hasData && !p.methodSet {
		p.config.Method = domain.MethodPost
	}

	return p.config, nil
}

// nextValue advances pos and returns the next token or an error.
func (p *curlParser) nextValue(flag string) (string, error) {
	p.pos++
	if p.pos >= len(p.tokens) {
		return "", fmt.Errorf("missing value for %s", flag)
	}
	return p.tokens[p.pos], nil
}

func (p *curlParser) handleMethod(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	p.config.Method = domain.HTTPMethod(strings.ToUpper(val))
	p.methodSet = true
	return nil
}

func (p *curlParser) handleHeader(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	key, value, ok := parseHeaderValue(val)
	if ok {
		p.config.Headers = append(p.config.Headers, domain.Header{Key: key, Value: value})
	}
	return nil
}

func (p *curlParser) handleData(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	p.config.Body = []byte(val)
	p.hasData = true
	return nil
}

func (p *curlParser) handleUser(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	parts := strings.SplitN(val, ":", 2)
	if len(parts) == 2 {
		p.config.Auth = &domain.AuthConfig{
			Type: domain.AuthBasic,
			Basic: &domain.BasicAuthConfig{
				Username: parts[0],
				Password: parts[1],
			},
		}
	}
	return nil
}

func (p *curlParser) handleUserAgent(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	p.config.Headers = append(p.config.Headers, domain.Header{Key: "User-Agent", Value: val})
	return nil
}

func (p *curlParser) handleCookie(flag string) error {
	val, err := p.nextValue(flag)
	if err != nil {
		return err
	}
	p.config.Headers = append(p.config.Headers, domain.Header{Key: "Cookie", Value: val})
	return nil
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
