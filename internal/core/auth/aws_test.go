package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestSignAWS_AddsAuthorizationHeader(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/x-www-form-urlencoded; charset=utf-8"},
		},
	}

	config := domain.AWSAuthConfig{
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		Region:    "us-east-1",
		Service:   "iam",
	}

	now := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)
	got := SignAWS(req, config, now)

	// Should have Authorization header.
	var authValue string
	for _, h := range got.Headers {
		if h.Key == "Authorization" {
			authValue = h.Value
			break
		}
	}
	if authValue == "" {
		t.Fatal("expected Authorization header, got none")
	}

	// Should start with AWS4-HMAC-SHA256.
	if !strings.HasPrefix(authValue, "AWS4-HMAC-SHA256") {
		t.Errorf("expected Authorization to start with AWS4-HMAC-SHA256, got: %s", authValue)
	}

	// Should contain Credential, SignedHeaders, Signature components.
	for _, want := range []string{"Credential=", "SignedHeaders=", "Signature="} {
		if !strings.Contains(authValue, want) {
			t.Errorf("Authorization missing %q, got: %s", want, authValue)
		}
	}
}

func TestSignAWS_AddsDateHeaders(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://s3.amazonaws.com/test-bucket",
	}

	config := domain.AWSAuthConfig{
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		Region:    "us-east-1",
		Service:   "s3",
	}

	now := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	got := SignAWS(req, config, now)

	// Should have X-Amz-Date header.
	var dateValue string
	for _, h := range got.Headers {
		if h.Key == "X-Amz-Date" {
			dateValue = h.Value
			break
		}
	}
	if dateValue == "" {
		t.Fatal("expected X-Amz-Date header, got none")
	}
	if dateValue != "20230115T103000Z" {
		t.Errorf("X-Amz-Date = %q, want %q", dateValue, "20230115T103000Z")
	}
}

func TestSignAWS_IncludesSessionToken(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://s3.amazonaws.com/test-bucket",
	}

	config := domain.AWSAuthConfig{
		AccessKey:    "AKIDEXAMPLE",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		Region:       "us-east-1",
		Service:      "s3",
		SessionToken: "FwoGZXIvYXdzEA0aDHTest",
	}

	now := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	got := SignAWS(req, config, now)

	// Should have X-Amz-Security-Token header.
	var tokenValue string
	for _, h := range got.Headers {
		if h.Key == "X-Amz-Security-Token" {
			tokenValue = h.Value
			break
		}
	}
	if tokenValue == "" {
		t.Fatal("expected X-Amz-Security-Token header, got none")
	}
	if tokenValue != config.SessionToken {
		t.Errorf("X-Amz-Security-Token = %q, want %q", tokenValue, config.SessionToken)
	}
}

func TestSignAWS_PreservesExistingHeaders(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodPost,
		URL:    "https://dynamodb.us-east-1.amazonaws.com/",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "X-Custom", Value: "value"},
		},
	}

	config := domain.AWSAuthConfig{
		AccessKey: "AKIDEXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		Region:    "us-east-1",
		Service:   "dynamodb",
	}

	now := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	got := SignAWS(req, config, now)

	// Verify existing headers are preserved.
	for _, orig := range req.Headers {
		found := false
		for _, h := range got.Headers {
			if h.Key == orig.Key && h.Value == orig.Value {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("existing header %q: %q was not preserved", orig.Key, orig.Value)
		}
	}

	// Verify original not modified.
	for _, h := range req.Headers {
		if h.Key == "Authorization" || h.Key == "X-Amz-Date" {
			t.Errorf("original request was modified: found %q header", h.Key)
		}
	}
}

func TestCanonicalRequest(t *testing.T) {
	// Test that the canonical request is properly formed.
	method := "GET"
	uri := "/"
	query := "Action=ListUsers&Version=2010-05-08"
	headers := []domain.Header{
		{Key: "Content-Type", Value: "application/x-www-form-urlencoded; charset=utf-8"},
		{Key: "Host", Value: "iam.amazonaws.com"},
		{Key: "X-Amz-Date", Value: "20150830T123600Z"},
	}

	canonical, signedHeaders := buildCanonicalRequest(method, uri, query, headers, []byte(""))

	// Canonical request must contain the method, uri, query, headers.
	if !strings.HasPrefix(canonical, "GET\n") {
		t.Errorf("canonical request should start with method, got: %s", canonical[:20])
	}

	// Signed headers should be lowercase and sorted.
	if signedHeaders != "content-type;host;x-amz-date" {
		t.Errorf("signedHeaders = %q, want %q", signedHeaders, "content-type;host;x-amz-date")
	}
}

func TestSigningKeyDerivation(t *testing.T) {
	key := deriveSigningKey("wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "20150830", "us-east-1", "iam")

	// The signing key should be non-nil and 32 bytes (SHA-256 output).
	if len(key) != 32 {
		t.Errorf("signing key length = %d, want 32", len(key))
	}
}
