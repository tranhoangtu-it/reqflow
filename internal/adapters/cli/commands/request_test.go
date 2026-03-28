package commands_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

// mockHTTPClient implements driven.HTTPClient for CLI testing.
type mockHTTPClient struct {
	doFunc func(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error)
}

func (m *mockHTTPClient) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	return m.doFunc(ctx, req)
}

func newTestApp(mock *mockHTTPClient) *app.App {
	return app.New(mock, nil)
}

func TestGetCommand_ParsesURLFromArgs(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Status: "200 OK", Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.URL != "https://example.com/api" {
		t.Errorf("expected URL https://example.com/api, got %s", capturedReq.URL)
	}
	if capturedReq.Method != domain.MethodGet {
		t.Errorf("expected method GET, got %s", capturedReq.Method)
	}
}

func TestPostCommand_ParsesDataFlag(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 201, Body: []byte("created")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"post", "https://example.com/api", "--data", `{"name":"test"}`})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Method != domain.MethodPost {
		t.Errorf("expected method POST, got %s", capturedReq.Method)
	}
	if string(capturedReq.Body) != `{"name":"test"}` {
		t.Errorf("expected body {\"name\":\"test\"}, got %s", string(capturedReq.Body))
	}
}

func TestHeaderFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "-H", "X-Custom: value1", "-H", "Accept: application/json"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedReq.Headers) < 2 {
		t.Fatalf("expected at least 2 headers, got %d: %v", len(capturedReq.Headers), capturedReq.Headers)
	}

	headerMap := make(map[string]string)
	for _, h := range capturedReq.Headers {
		headerMap[h.Key] = h.Value
	}

	if headerMap["X-Custom"] != "value1" {
		t.Errorf("expected X-Custom header with value 'value1', got %q", headerMap["X-Custom"])
	}
	if headerMap["Accept"] != "application/json" {
		t.Errorf("expected Accept header with value 'application/json', got %q", headerMap["Accept"])
	}
}

func TestQueryFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "-q", "page=1", "-q", "limit=10"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedReq.QueryParams) < 2 {
		t.Fatalf("expected at least 2 query params, got %d: %v", len(capturedReq.QueryParams), capturedReq.QueryParams)
	}

	paramMap := make(map[string]string)
	for _, qp := range capturedReq.QueryParams {
		paramMap[qp.Key] = qp.Value
	}

	if paramMap["page"] != "1" {
		t.Errorf("expected page=1, got %q", paramMap["page"])
	}
	if paramMap["limit"] != "10" {
		t.Errorf("expected limit=10, got %q", paramMap["limit"])
	}
}

func TestAuthBasicFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--auth-basic", "user:pass"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, h := range capturedReq.Headers {
		if h.Key == "Authorization" {
			found = true
			// Basic dXNlcjpwYXNz = base64("user:pass")
			if h.Value != "Basic dXNlcjpwYXNz" {
				t.Errorf("expected Basic auth header value 'Basic dXNlcjpwYXNz', got %q", h.Value)
			}
			break
		}
	}
	if !found {
		t.Error("expected Authorization header not found")
	}
}

func TestAuthBearerFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--auth-bearer", "my-jwt-token"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, h := range capturedReq.Headers {
		if h.Key == "Authorization" && h.Value == "Bearer my-jwt-token" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Authorization: Bearer my-jwt-token, got headers: %v", capturedReq.Headers)
	}
}

func TestAuthDigestFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--auth-digest", "admin:s3cret"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Digest auth does not add an Authorization header on the first pass
	// (it needs a 401 challenge first), but the request should still succeed.
	_ = capturedReq
}

func TestAuthDigestFlag_InvalidFormat(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--auth-digest", "nocolon"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --auth-digest format, got nil")
	}
	if !strings.Contains(err.Error(), "auth-digest") {
		t.Errorf("error should mention --auth-digest: %v", err)
	}
}

func TestAuthAWSFlag_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "https://s3.amazonaws.com/test-bucket",
		"--auth-aws", "AKIDEXAMPLE:wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		"--auth-aws-region", "us-east-1",
		"--auth-aws-service", "s3",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have Authorization header with AWS4-HMAC-SHA256
	found := false
	for _, h := range capturedReq.Headers {
		if h.Key == "Authorization" && strings.HasPrefix(h.Value, "AWS4-HMAC-SHA256") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Authorization: AWS4-HMAC-SHA256 header, got headers: %v", capturedReq.Headers)
	}

	// Should have X-Amz-Date header
	dateFound := false
	for _, h := range capturedReq.Headers {
		if h.Key == "X-Amz-Date" {
			dateFound = true
			break
		}
	}
	if !dateFound {
		t.Errorf("expected X-Amz-Date header, got headers: %v", capturedReq.Headers)
	}
}

func TestAuthAWSFlag_InvalidFormat(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "https://s3.amazonaws.com/test-bucket",
		"--auth-aws", "nocolon",
		"--auth-aws-region", "us-east-1",
		"--auth-aws-service", "s3",
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --auth-aws format, got nil")
	}
	if !strings.Contains(err.Error(), "auth-aws") {
		t.Errorf("error should mention --auth-aws: %v", err)
	}
}

func TestAuthOAuth2Flags_ParsedCorrectly(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			// Return a token response for the OAuth2 token request,
			// and a normal response for the actual request.
			if strings.Contains(req.URL, "token") {
				return domain.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"access_token":"test-token-123","token_type":"Bearer"}`),
				}, nil
			}
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "https://api.example.com/resource",
		"--auth-oauth2-url", "https://auth.example.com/token",
		"--auth-oauth2-client", "my-client:my-secret",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The final request should have Authorization: Bearer header
	found := false
	for _, h := range capturedReq.Headers {
		if h.Key == "Authorization" && h.Value == "Bearer test-token-123" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Authorization: Bearer test-token-123, got headers: %v", capturedReq.Headers)
	}
}

func TestVerboseFlag_ShowsRequestAndResponseDetails(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Status:     "200 OK",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
				},
				Body: []byte(`{"ok":true}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://api.example.com/users", "--verbose", "--no-color"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should have request lines
	if !strings.Contains(output, "> GET") {
		t.Errorf("expected verbose request line with '> GET', got:\n%s", output)
	}

	// Should have response lines
	if !strings.Contains(output, "< HTTP/1.1 200 OK") {
		t.Errorf("expected verbose response line '< HTTP/1.1 200 OK', got:\n%s", output)
	}
}

func TestDryRunFlag_DoesNotMakeHTTPCall(t *testing.T) {
	callCount := 0
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			callCount++
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--dry-run", "--no-color"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 0 {
		t.Errorf("expected no HTTP calls with --dry-run, got %d", callCount)
	}
}

func TestDryRunFlag_ShowsDryRunIndicator(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--dry-run", "--no-color"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected 'DRY RUN' indicator, got:\n%s", output)
	}
}

func TestDryRunFlag_ShowsRequestDetails(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--dry-run", "--no-color", "-H", "Accept: application/json"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "> GET /api HTTP/1.1") {
		t.Errorf("expected request line, got:\n%s", output)
	}
	if !strings.Contains(output, "> Host: example.com") {
		t.Errorf("expected host header, got:\n%s", output)
	}
	if !strings.Contains(output, "> Accept: application/json") {
		t.Errorf("expected Accept header, got:\n%s", output)
	}
}

func TestMissingURL_ReturnsError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
}

func TestGetCommand_404_ReturnsExitError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 404,
				Status:     "404 Not Found",
				Body:       []byte("not found"),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/missing"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}

	var exitErr *domain.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *domain.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != domain.ExitHTTPError {
		t.Errorf("Code = %d, want %d", exitErr.Code, domain.ExitHTTPError)
	}
}

func TestGetCommand_200_ReturnsNoError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Status:     "200 OK",
				Body:       []byte("ok"),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/ok"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error for 200 response, got: %v", err)
	}
}

func TestGetCommand_NoFailOnError_404_ReturnsNil(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 404,
				Status:     "404 Not Found",
				Body:       []byte("not found"),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/missing", "--no-fail-on-error"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error with --no-fail-on-error, got: %v", err)
	}
}

func TestGetCommand_500_ReturnsExitError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 500,
				Status:     "500 Internal Server Error",
				Body:       []byte("server error"),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/error"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}

	var exitErr *domain.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *domain.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != domain.ExitHTTPError {
		t.Errorf("Code = %d, want %d", exitErr.Code, domain.ExitHTTPError)
	}
}

func TestGetCommand_WaitForFlag_TriggersPolling(t *testing.T) {
	callCount := 0
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			callCount++
			if callCount < 3 {
				return domain.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"status":"pending"}`),
				}, nil
			}
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"status":"completed"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "https://example.com/jobs/1",
		"--wait-for", "$.status == 'completed'",
		"--poll", "10ms",
		"--poll-timeout", "5s",
		"--no-fail-on-error",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount < 3 {
		t.Errorf("expected at least 3 poll attempts, got %d", callCount)
	}

	output := buf.String()
	if !strings.Contains(output, "completed") {
		t.Errorf("expected output to contain final 'completed' response, got:\n%s", output)
	}
}

func TestGetCommand_WaitForFlag_PollTimeout(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"status":"pending"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"get", "https://example.com/jobs/1",
		"--wait-for", "$.status == 'completed'",
		"--poll", "10ms",
		"--poll-timeout", "50ms",
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got: %v", err)
	}
}
