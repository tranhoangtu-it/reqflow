package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// Compile-time check that Client satisfies the driven.HTTPClient interface.
var _ driven.HTTPClient = (*Client)(nil)

const defaultTimeout = 30 * time.Second

// Client is an HTTP client adapter that wraps Go's net/http.Client
// and satisfies the driven.HTTPClient port interface.
type Client struct {
	httpClient *http.Client
	trace      bool
	cookieJar  driven.CookieJar
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithTransport sets a custom http.RoundTripper (useful for testing, proxies).
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.Transport = t
	}
}

// WithTrace enables detailed timing instrumentation via httptrace.
func WithTrace(enabled bool) Option {
	return func(c *Client) {
		c.trace = enabled
	}
}

// WithCookieJar sets a cookie jar for automatic cookie management.
func WithCookieJar(jar driven.CookieJar) Option {
	return func(c *Client) {
		c.cookieJar = jar
	}
}

// New creates a new Client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Do executes an HTTP request and returns the response.
func (c *Client) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	reqURL, err := buildURL(req.URL, req.QueryParams)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, string(req.Method), reqURL, bodyReader)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	for _, h := range req.Headers {
		httpReq.Header.Add(h.Key, h.Value)
	}

	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Add cookies from jar if available and not disabled for this request.
	if c.cookieJar != nil && !req.NoCookies {
		cookies, jarErr := c.cookieJar.GetCookies(reqURL)
		if jarErr == nil {
			for _, dc := range cookies {
				httpReq.AddCookie(&http.Cookie{Name: dc.Name, Value: dc.Value})
			}
		}
	}

	var timing domain.TimingInfo
	if c.trace {
		httpReq = c.withTrace(httpReq, &timing)
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return domain.HTTPResponse{}, err
	}
	defer httpResp.Body.Close()

	if c.trace {
		timing.Total = duration
	}

	// Store Set-Cookie response headers in jar if available and not disabled.
	if c.cookieJar != nil && !req.NoCookies {
		if setCookies := httpResp.Cookies(); len(setCookies) > 0 {
			var domainCookies []domain.Cookie
			for _, hc := range setCookies {
				dc := domain.Cookie{
					Name:     hc.Name,
					Value:    hc.Value,
					Domain:   hc.Domain,
					Path:     hc.Path,
					Expires:  hc.Expires,
					Secure:   hc.Secure,
					HTTPOnly: hc.HttpOnly,
				}
				// If domain not set in cookie, infer from request URL.
				if dc.Domain == "" {
					if u, parseErr := url.Parse(reqURL); parseErr == nil {
						dc.Domain = u.Hostname()
					}
				}
				if dc.Path == "" {
					dc.Path = "/"
				}
				domainCookies = append(domainCookies, dc)
			}
			_ = c.cookieJar.SetCookies(reqURL, domainCookies)
		}
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	var headers []domain.Header
	for key, values := range httpResp.Header {
		for _, v := range values {
			headers = append(headers, domain.Header{Key: key, Value: v})
		}
	}

	return domain.HTTPResponse{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    headers,
		Body:       body,
		Duration:   duration,
		Size:       int64(len(body)),
		Timing:     timing,
	}, nil
}

// withTrace attaches an httptrace.ClientTrace to the request to capture timing.
func (c *Client) withTrace(req *http.Request, timing *domain.TimingInfo) *http.Request {
	var (
		dnsStart     time.Time
		connectStart time.Time
		tlsStart     time.Time
		reqStart     = time.Now()
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				timing.DNSLookup = time.Since(dnsStart)
			}
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil && !connectStart.IsZero() {
				timing.TCPConnect = time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tlsStart.IsZero() {
				timing.TLSHandshake = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() {
			timing.FirstByte = time.Since(reqStart)
		},
	}

	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
}

func buildURL(baseURL string, params []domain.QueryParam) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for _, p := range params {
		q.Add(p.Key, p.Value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
