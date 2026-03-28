package runner

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

// WebhookResult holds the captured data from a webhook callback.
type WebhookResult struct {
	Body    []byte
	Headers map[string]string
	Error   error
}

// WebhookListener starts an HTTP server that waits for a single callback request.
type WebhookListener struct {
	port     int
	path     string
	timeout  time.Duration
	capture  string
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
}

// NewWebhookListener creates a new WebhookListener from the given config.
func NewWebhookListener(config domain.ListenConfig) *WebhookListener {
	return &WebhookListener{
		port:    config.Port,
		path:    config.Path,
		timeout: config.Timeout,
		capture: config.Capture,
	}
}

// Port returns the actual port the listener is bound to.
// This is useful when port 0 is used to get an ephemeral port.
func (w *WebhookListener) Port() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.listener != nil {
		return w.listener.Addr().(*net.TCPAddr).Port
	}
	return w.port
}

// Start begins listening for a webhook callback. It returns a channel that
// will receive exactly one WebhookResult when either a matching request
// arrives, the timeout elapses, or the context is cancelled.
func (w *WebhookListener) Start(ctx context.Context) (chan WebhookResult, error) {
	resultCh := make(chan WebhookResult, 1)

	// Create a listener on the configured port (0 = ephemeral)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", w.port))
	if err != nil {
		return nil, fmt.Errorf("listen on port %d: %w", w.port, err)
	}

	w.mu.Lock()
	w.listener = ln
	w.mu.Unlock()

	// Create a context with timeout
	listenCtx, cancel := context.WithTimeout(ctx, w.timeout)

	mux := http.NewServeMux()
	mux.HandleFunc(w.path, func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		headers := make(map[string]string)
		for key := range r.Header {
			headers[key] = r.Header.Get(key)
		}

		rw.WriteHeader(http.StatusOK)

		select {
		case resultCh <- WebhookResult{Body: body, Headers: headers}:
		default:
			// Result already sent (shouldn't happen with buffered channel of 1,
			// but guard against duplicate requests)
		}
		cancel()
	})

	// Handle non-matching paths with 404
	wrapper := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != w.path {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		mux.ServeHTTP(rw, r)
	})

	w.mu.Lock()
	w.server = &http.Server{Handler: wrapper}
	w.mu.Unlock()

	// Serve in the background
	go func() {
		if err := w.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			select {
			case resultCh <- WebhookResult{Error: fmt.Errorf("server error: %w", err)}:
			default:
			}
		}
	}()

	// Watch for timeout or context cancellation
	go func() {
		<-listenCtx.Done()

		// Shut down the server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		w.server.Shutdown(shutdownCtx)

		// If no result was sent yet, send an error
		err := listenCtx.Err()
		if err != nil {
			select {
			case resultCh <- WebhookResult{Error: fmt.Errorf("webhook listener: %w", err)}:
			default:
				// Result already sent by handler, which is fine
			}
		}
	}()

	return resultCh, nil
}

// Stop shuts down the webhook listener server.
func (w *WebhookListener) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return w.server.Shutdown(ctx)
	}
	return nil
}
