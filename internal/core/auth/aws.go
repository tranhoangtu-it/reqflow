package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

const (
	awsAlgorithm   = "AWS4-HMAC-SHA256"
	awsTermination = "aws4_request"
)

// SignAWS returns a new HTTPRequest with AWS Signature V4 authentication applied.
// It adds the Authorization, X-Amz-Date, and optionally X-Amz-Security-Token headers.
func SignAWS(req domain.HTTPRequest, config domain.AWSAuthConfig, now time.Time) domain.HTTPRequest {
	// Format timestamps.
	dateStamp := now.UTC().Format("20060102")
	amzDate := now.UTC().Format("20060102T150405Z")

	// Parse URL to extract host, path, and query.
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		// If URL parsing fails, return request unchanged.
		return req
	}

	host := parsedURL.Host
	uri := parsedURL.Path
	if uri == "" {
		uri = "/"
	}
	query := parsedURL.RawQuery

	// Build headers list including Host and X-Amz-Date.
	headers := make([]domain.Header, len(req.Headers))
	copy(headers, req.Headers)

	// Add Host header if not already present.
	hasHost := false
	for _, h := range headers {
		if strings.EqualFold(h.Key, "host") {
			hasHost = true
			break
		}
	}
	if !hasHost {
		headers = append(headers, domain.Header{Key: "Host", Value: host})
	}

	// Add X-Amz-Date header.
	headers = append(headers, domain.Header{Key: "X-Amz-Date", Value: amzDate})

	// Add security token header if present.
	if config.SessionToken != "" {
		headers = append(headers, domain.Header{Key: "X-Amz-Security-Token", Value: config.SessionToken})
	}

	// Build canonical request.
	canonical, signedHeaders := buildCanonicalRequest(string(req.Method), uri, query, headers, req.Body)

	// Create string to sign.
	credentialScope := fmt.Sprintf("%s/%s/%s/%s", dateStamp, config.Region, config.Service, awsTermination)
	stringToSign := buildStringToSign(amzDate, credentialScope, canonical)

	// Derive signing key and compute signature.
	signingKey := deriveSigningKey(config.SecretKey, dateStamp, config.Region, config.Service)
	signature := hmacSHA256Hex(signingKey, stringToSign)

	// Build Authorization header.
	authHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		awsAlgorithm, config.AccessKey, credentialScope, signedHeaders, signature)

	// Build final request with all new headers.
	result := req
	result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+3)
	copy(result.Headers, req.Headers)

	// Add Host if not already present.
	if !hasHost {
		result.Headers = append(result.Headers, domain.Header{Key: "Host", Value: host})
	}

	result.Headers = append(result.Headers, domain.Header{Key: "X-Amz-Date", Value: amzDate})

	if config.SessionToken != "" {
		result.Headers = append(result.Headers, domain.Header{Key: "X-Amz-Security-Token", Value: config.SessionToken})
	}

	result.Headers = append(result.Headers, domain.Header{Key: "Authorization", Value: authHeader})

	return result
}

// buildCanonicalRequest creates the canonical request string for AWS Sig V4.
// Returns the canonical request and the signed headers string.
func buildCanonicalRequest(method, uri, query string, headers []domain.Header, body []byte) (string, string) {
	// URI encode the path.
	canonicalURI := uri
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Sort query parameters.
	canonicalQuery := buildCanonicalQueryString(query)

	// Build canonical headers (lowercase, sorted, trimmed values).
	type headerEntry struct {
		key   string
		value string
	}
	var entries []headerEntry
	for _, h := range headers {
		entries = append(entries, headerEntry{
			key:   strings.ToLower(h.Key),
			value: strings.TrimSpace(h.Value),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	var canonicalHeaders strings.Builder
	var signedHeadersList []string
	for _, e := range entries {
		canonicalHeaders.WriteString(e.key)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(e.value)
		canonicalHeaders.WriteString("\n")
		signedHeadersList = append(signedHeadersList, e.key)
	}
	signedHeaders := strings.Join(signedHeadersList, ";")

	// Hash the payload.
	payloadHash := sha256Hex(body)

	canonical := strings.Join([]string{
		method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	return canonical, signedHeaders
}

// buildCanonicalQueryString sorts query parameters and returns the canonical query string.
func buildCanonicalQueryString(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	params := strings.Split(rawQuery, "&")
	sort.Strings(params)
	return strings.Join(params, "&")
}

// buildStringToSign creates the string to sign for AWS Sig V4.
func buildStringToSign(amzDate, credentialScope, canonicalRequest string) string {
	return strings.Join([]string{
		awsAlgorithm,
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
}

// deriveSigningKey derives the signing key using the HMAC chain:
// kDate = HMAC("AWS4" + secretKey, dateStamp)
// kRegion = HMAC(kDate, region)
// kService = HMAC(kRegion, service)
// kSigning = HMAC(kService, "aws4_request")
func deriveSigningKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, awsTermination)
	return kSigning
}

// sha256Hex computes the SHA-256 hash of data and returns it as a lowercase hex string.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// hmacSHA256 computes HMAC-SHA256 of msg using key.
func hmacSHA256(key []byte, msg string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))
	return mac.Sum(nil)
}

// hmacSHA256Hex computes HMAC-SHA256 of msg using key and returns hex string.
func hmacSHA256Hex(key []byte, msg string) string {
	return fmt.Sprintf("%x", hmacSHA256(key, msg))
}
