package auth

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/ye-kart/reqflow/internal/domain"
)

// DigestChallenge holds the parsed fields from a WWW-Authenticate: Digest header.
type DigestChallenge struct {
	Realm     string
	Nonce     string
	QOP       string
	Opaque    string
	Algorithm string
}

// ParseWWWAuthenticate parses a WWW-Authenticate header value into a DigestChallenge.
// It expects the header to start with "Digest " followed by comma-separated key=value pairs.
func ParseWWWAuthenticate(header string) DigestChallenge {
	var c DigestChallenge

	// Strip "Digest " prefix (case-insensitive).
	trimmed := header
	if len(trimmed) > 7 && strings.EqualFold(trimmed[:7], "digest ") {
		trimmed = trimmed[7:]
	}

	// Split on commas and parse each key=value pair.
	parts := strings.Split(trimmed, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		value := strings.TrimSpace(part[idx+1:])

		// Strip surrounding quotes if present.
		value = stripQuotes(value)

		switch strings.ToLower(key) {
		case "realm":
			c.Realm = value
		case "nonce":
			c.Nonce = value
		case "qop":
			c.QOP = value
		case "opaque":
			c.Opaque = value
		case "algorithm":
			c.Algorithm = value
		}
	}

	return c
}

// ComputeDigestResponse computes the Digest Authorization header value.
// It follows RFC 2617 with MD5 and qop=auth:
//   - HA1 = MD5(username:realm:password)
//   - HA2 = MD5(method:uri)
//   - response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
//
// cnonce is the client nonce string, nc is the nonce count (1-based).
// Returns the full Digest Authorization header value.
func ComputeDigestResponse(challenge DigestChallenge, method, uri, username, password, cnonce string, nc int) string {
	ha1 := md5Hex(fmt.Sprintf("%s:%s:%s", username, challenge.Realm, password))
	ha2 := md5Hex(fmt.Sprintf("%s:%s", method, uri))

	ncStr := fmt.Sprintf("%08x", nc)

	var response string
	if challenge.QOP == "auth" || challenge.QOP == "auth-int" {
		response = md5Hex(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, challenge.Nonce, ncStr, cnonce, challenge.QOP, ha2))
	} else {
		// RFC 2069 compatibility (no qop).
		response = md5Hex(fmt.Sprintf("%s:%s:%s", ha1, challenge.Nonce, ha2))
	}

	// Build the Authorization header value.
	parts := []string{
		fmt.Sprintf(`username="%s"`, username),
		fmt.Sprintf(`realm="%s"`, challenge.Realm),
		fmt.Sprintf(`nonce="%s"`, challenge.Nonce),
		fmt.Sprintf(`uri="%s"`, uri),
		fmt.Sprintf(`qop=%s`, challenge.QOP),
		fmt.Sprintf(`nc=%s`, ncStr),
		fmt.Sprintf(`cnonce="%s"`, cnonce),
		fmt.Sprintf(`response="%s"`, response),
	}

	if challenge.Opaque != "" {
		parts = append(parts, fmt.Sprintf(`opaque="%s"`, challenge.Opaque))
	}

	return "Digest " + strings.Join(parts, ", ")
}

// ApplyDigest returns a new HTTPRequest with Digest authentication applied.
// It generates a cnonce and computes the digest response for nc=1.
func ApplyDigest(req domain.HTTPRequest, challenge DigestChallenge, username, password string) domain.HTTPRequest {
	// Extract URI path from the full URL.
	uri := extractURIPath(req.URL)

	// Generate a deterministic-ish cnonce from username+nonce for reproducibility,
	// but in production this should be random.
	cnonce := md5Hex(fmt.Sprintf("%s:%s", username, challenge.Nonce))[:16]

	authValue := ComputeDigestResponse(challenge, string(req.Method), uri, username, password, cnonce, 1)

	result := req
	result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+1)
	copy(result.Headers, req.Headers)
	result.Headers = append(result.Headers, domain.Header{
		Key:   "Authorization",
		Value: authValue,
	})

	return result
}

// stripQuotes removes surrounding double quotes from a string.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// md5Hex computes the MD5 hash of s and returns it as a lowercase hex string.
func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}

// extractURIPath extracts the path component from a URL string.
func extractURIPath(rawURL string) string {
	// Find the start of the path after the scheme and host.
	// e.g., "https://example.com/path" -> "/path"
	idx := strings.Index(rawURL, "://")
	if idx >= 0 {
		rest := rawURL[idx+3:]
		pathIdx := strings.Index(rest, "/")
		if pathIdx >= 0 {
			return rest[pathIdx:]
		}
		return "/"
	}
	// If no scheme, assume it's already a path.
	return rawURL
}
