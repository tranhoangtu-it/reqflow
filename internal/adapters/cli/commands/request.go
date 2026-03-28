package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/adapters/cli/output"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

// addBodyFlags adds --data and --content-type flags to commands that accept a body.
func addBodyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("data", "d", "", "request body (JSON string)")
	cmd.Flags().String("content-type", "", "content type")
}

// addAuthFlags adds authentication flags to a command.
func addAuthFlags(cmd *cobra.Command) {
	cmd.Flags().String("auth-basic", "", `basic auth credentials (format "username:password")`)
	cmd.Flags().String("auth-bearer", "", "bearer token")
	cmd.Flags().String("auth-apikey-header", "", `API key in header (format "HeaderName:Value")`)
	cmd.Flags().String("auth-apikey-query", "", `API key in query (format "paramName=Value")`)
}

func newGetCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <url>",
		Short: "Send a GET request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodGet, false),
	}
	addAuthFlags(cmd)
	return cmd
}

func newPostCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post <url>",
		Short: "Send a POST request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodPost, true),
	}
	addBodyFlags(cmd)
	addAuthFlags(cmd)
	return cmd
}

func newPutCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put <url>",
		Short: "Send a PUT request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodPut, true),
	}
	addBodyFlags(cmd)
	addAuthFlags(cmd)
	return cmd
}

func newPatchCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch <url>",
		Short: "Send a PATCH request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodPatch, true),
	}
	addBodyFlags(cmd)
	addAuthFlags(cmd)
	return cmd
}

func newDeleteCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <url>",
		Short: "Send a DELETE request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodDelete, false),
	}
	addAuthFlags(cmd)
	return cmd
}

// makeRunE creates a RunE function for an HTTP method command.
func makeRunE(a *app.App, method domain.HTTPMethod, hasBody bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		url := args[0]

		config := domain.RequestConfig{
			Method: method,
			URL:    url,
		}

		// Parse timeout from persistent flags.
		timeout, _ := cmd.Flags().GetDuration("timeout")
		if timeout > 0 {
			config.Timeout = timeout
		}

		// Apply default headers from config (these can be overridden by CLI flags).
		cfg := configFromContext(cmd.Context())
		if cfg != nil && len(cfg.DefaultHeaders) > 0 {
			for _, h := range cfg.DefaultHeaders {
				config.Headers = append(config.Headers, domain.Header{Key: h.Key, Value: h.Value})
			}
		}

		// Parse headers from persistent flags (override config defaults).
		headers, _ := cmd.Flags().GetStringSlice("header")
		cliHeaderKeys := make(map[string]bool)
		for _, h := range headers {
			key, value, ok := parseHeader(h)
			if ok {
				cliHeaderKeys[key] = true
				config.Headers = append(config.Headers, domain.Header{Key: key, Value: value})
			}
		}

		// Remove config default headers that were overridden by CLI flags.
		if len(cliHeaderKeys) > 0 {
			filtered := make([]domain.Header, 0, len(config.Headers))
			seen := make(map[string]bool)
			// Walk in reverse so CLI flags (later) take precedence over config (earlier).
			for i := len(config.Headers) - 1; i >= 0; i-- {
				h := config.Headers[i]
				if !seen[h.Key] {
					seen[h.Key] = true
					filtered = append(filtered, h)
				}
			}
			// Reverse to restore order.
			for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
			config.Headers = filtered
		}

		// Parse query params from persistent flags.
		queries, _ := cmd.Flags().GetStringSlice("query")
		for _, q := range queries {
			key, value, ok := parseQueryParam(q)
			if ok {
				config.QueryParams = append(config.QueryParams, domain.QueryParam{Key: key, Value: value})
			}
		}

		// Parse body flags if applicable.
		if hasBody {
			data, _ := cmd.Flags().GetString("data")
			if data != "" {
				config.Body = []byte(data)
			}
			contentType, _ := cmd.Flags().GetString("content-type")
			if contentType != "" {
				config.ContentType = contentType
			}
		}

		// Parse auth flags.
		authConfig, err := parseAuthFlags(cmd)
		if err != nil {
			return err
		}
		if authConfig != nil {
			config.Auth = authConfig
		}

		// Execute the request.
		ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(timeout))
		defer cancel()

		result, err := a.HTTPExecutor.Execute(ctx, config, nil)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		// Format and write response.
		outputFmt, _ := cmd.Flags().GetString("output")
		noColor, _ := cmd.Flags().GetBool("no-color")
		formatter := output.New(domain.OutputFormat(outputFmt), noColor)
		return formatter.FormatResponse(os.Stdout, result.Response)
	}
}

// parseHeader splits a header string on the first ": " separator.
func parseHeader(s string) (key, value string, ok bool) {
	idx := strings.Index(s, ": ")
	if idx < 0 {
		return "", "", false
	}
	return s[:idx], s[idx+2:], true
}

// parseQueryParam splits a query param string on the first "=" separator.
func parseQueryParam(s string) (key, value string, ok bool) {
	idx := strings.Index(s, "=")
	if idx < 0 {
		return "", "", false
	}
	return s[:idx], s[idx+1:], true
}

// parseAuthFlags reads auth-related flags and returns an AuthConfig if any are set.
func parseAuthFlags(cmd *cobra.Command) (*domain.AuthConfig, error) {
	basic, _ := cmd.Flags().GetString("auth-basic")
	if basic != "" {
		parts := strings.SplitN(basic, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --auth-basic format, expected 'username:password'")
		}
		return &domain.AuthConfig{
			Type: domain.AuthBasic,
			Basic: &domain.BasicAuthConfig{
				Username: parts[0],
				Password: parts[1],
			},
		}, nil
	}

	bearer, _ := cmd.Flags().GetString("auth-bearer")
	if bearer != "" {
		return &domain.AuthConfig{
			Type: domain.AuthBearer,
			Bearer: &domain.BearerAuthConfig{
				Token: bearer,
			},
		}, nil
	}

	apikeyHeader, _ := cmd.Flags().GetString("auth-apikey-header")
	if apikeyHeader != "" {
		parts := strings.SplitN(apikeyHeader, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --auth-apikey-header format, expected 'HeaderName:Value'")
		}
		return &domain.AuthConfig{
			Type: domain.AuthAPIKey,
			APIKey: &domain.APIKeyAuthConfig{
				Key:      parts[0],
				Value:    parts[1],
				Location: domain.APIKeyInHeader,
			},
		}, nil
	}

	apikeyQuery, _ := cmd.Flags().GetString("auth-apikey-query")
	if apikeyQuery != "" {
		parts := strings.SplitN(apikeyQuery, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --auth-apikey-query format, expected 'paramName=Value'")
		}
		return &domain.AuthConfig{
			Type: domain.AuthAPIKey,
			APIKey: &domain.APIKeyAuthConfig{
				Key:      parts[0],
				Value:    parts[1],
				Location: domain.APIKeyInQuery,
			},
		}, nil
	}

	return nil, nil
}

// resolveTimeout returns a sensible timeout value.
func resolveTimeout(d time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return 30 * time.Second
}
