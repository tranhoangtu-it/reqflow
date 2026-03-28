package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/adapters/cli/output"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/core/variable"
	"github.com/ye-kart/reqflow/internal/core/workflow"
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

// addCurlFlag adds the --curl flag for exporting instead of executing.
func addCurlFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("curl", false, "print the equivalent cURL command instead of executing")
}

// addPollFlags adds polling flags to a command.
func addPollFlags(cmd *cobra.Command) {
	cmd.Flags().String("wait-for", "", `poll until JSONPath condition is met (e.g. "$.status == 'completed'")`)
	cmd.Flags().Duration("poll", 2*time.Second, "interval between poll attempts")
	cmd.Flags().Duration("poll-timeout", 60*time.Second, "maximum time to wait for condition")
}

// addFailOnErrorFlag adds the --fail-on-error and --no-fail-on-error flags.
func addFailOnErrorFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("fail-on-error", true, "exit with non-zero code on HTTP 4xx/5xx")
	cmd.Flags().Bool("no-fail-on-error", false, "do not exit with non-zero code on HTTP 4xx/5xx")
}

func newGetCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <url>",
		Short: "Send a GET request",
		Args:  cobra.ExactArgs(1),
		RunE:  makeRunE(a, domain.MethodGet, false),
	}
	addAuthFlags(cmd)
	addCurlFlag(cmd)
	addFailOnErrorFlag(cmd)
	addPollFlags(cmd)
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
	addCurlFlag(cmd)
	addFailOnErrorFlag(cmd)
	addPollFlags(cmd)
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
	addCurlFlag(cmd)
	addFailOnErrorFlag(cmd)
	addPollFlags(cmd)
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
	addCurlFlag(cmd)
	addFailOnErrorFlag(cmd)
	addPollFlags(cmd)
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
	addCurlFlag(cmd)
	addFailOnErrorFlag(cmd)
	addPollFlags(cmd)
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

		// Build headers: config defaults first, then CLI flags override by key.
		cfg := configFromContext(cmd.Context())
		var configHeaders []domain.Header
		if cfg != nil {
			configHeaders = cfg.DefaultHeaders
		}
		headers, _ := cmd.Flags().GetStringSlice("header")
		var cliHeaders []domain.Header
		for _, h := range headers {
			key, value, ok := parseHeader(h)
			if ok {
				cliHeaders = append(cliHeaders, domain.Header{Key: key, Value: value})
			}
		}
		config.Headers = mergeHeaders(configHeaders, cliHeaders)

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

		// If --curl flag is set, print the curl equivalent and return.
		curlExport, _ := cmd.Flags().GetBool("curl")
		if curlExport {
			return printCurlExport(cmd, config)
		}

		// Load environment variables if -e flag is set.
		var vars map[string]string
		envName, _ := cmd.Flags().GetString("env")
		if envName != "" && a.Storage != nil {
			envDir, _ := cmd.Flags().GetString("env-dir")
			envPath := filepath.Join(envDir, envName+".yaml")
			env, err := a.Storage.ReadEnvironment(envPath)
			if err != nil {
				return fmt.Errorf("loading environment %q: %w", envName, err)
			}
			vars = variable.Resolve(env.Variables)
		}

		// Parse display flags.
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		trace, _ := cmd.Flags().GetBool("trace")
		noColor, _ := cmd.Flags().GetBool("no-color")
		w := cmd.OutOrStdout()

		// Dry-run mode: build request and display it without sending.
		if dryRun {
			req, err := a.HTTPExecutor.BuildRequest(config, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}
			return output.FormatDryRun(w, req, noColor)
		}

		// Enable trace timing if requested.
		if trace {
			a.EnableTrace()
		}

		// Check for polling flags.
		waitFor, _ := cmd.Flags().GetString("wait-for")
		if waitFor != "" {
			return executePollRequest(cmd, a, config, vars, waitFor, verbose, trace, noColor, timeout)
		}

		// Execute the request.
		ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(timeout))
		defer cancel()

		result, err := a.HTTPExecutor.Execute(ctx, config, vars)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		// Verbose mode: show request and response details.
		if verbose {
			if err := output.FormatVerbose(w, result.Request, result.Response, noColor); err != nil {
				return err
			}
		} else {
			// Normal mode: format and write response.
			outputFmt, _ := cmd.Flags().GetString("output")
			formatter := output.New(domain.OutputFormat(outputFmt), noColor)
			if err := formatter.FormatResponse(w, result.Response); err != nil {
				return err
			}
		}

		// Show trace timing if requested.
		if trace {
			fmt.Fprintln(w)
			if err := output.FormatTrace(w, result.Response.Timing, noColor); err != nil {
				return err
			}
		}

		// Check if response indicates an error and --fail-on-error is active.
		if result.Response.StatusCode >= 400 {
			if shouldFailOnError(cmd) {
				return domain.NewHTTPError(result.Response.StatusCode, nil)
			}
		}

		return nil
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

// mergeHeaders combines config default headers with CLI flag headers.
// CLI headers override config headers with the same key.
func mergeHeaders(configHeaders, cliHeaders []domain.Header) []domain.Header {
	if len(configHeaders) == 0 {
		return cliHeaders
	}

	// Build set of CLI header keys for fast lookup.
	cliKeys := make(map[string]bool, len(cliHeaders))
	for _, h := range cliHeaders {
		cliKeys[h.Key] = true
	}

	// Start with config headers that are not overridden.
	merged := make([]domain.Header, 0, len(configHeaders)+len(cliHeaders))
	for _, h := range configHeaders {
		if !cliKeys[h.Key] {
			merged = append(merged, h)
		}
	}
	// Append all CLI headers.
	merged = append(merged, cliHeaders...)
	return merged
}

// shouldFailOnError returns true if the command should return an error for HTTP 4xx/5xx.
// --no-fail-on-error takes precedence over --fail-on-error.
func shouldFailOnError(cmd *cobra.Command) bool {
	noFail, _ := cmd.Flags().GetBool("no-fail-on-error")
	if noFail {
		return false
	}
	failOnError, _ := cmd.Flags().GetBool("fail-on-error")
	return failOnError
}

// resolveTimeout returns a sensible timeout value.
func resolveTimeout(d time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return 30 * time.Second
}

// executePollRequest handles the --wait-for polling execution path.
func executePollRequest(cmd *cobra.Command, a *app.App, config domain.RequestConfig, vars map[string]string, waitFor string, verbose, trace, noColor bool, timeout time.Duration) error {
	pollInterval, _ := cmd.Flags().GetDuration("poll")
	pollTimeout, _ := cmd.Flags().GetDuration("poll-timeout")

	pollCtx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()

	w := cmd.OutOrStdout()
	attempt := 0

	for {
		attempt++

		// Use per-request timeout if set, otherwise use a reasonable default
		reqCtx, reqCancel := context.WithTimeout(pollCtx, resolveTimeout(timeout))
		result, err := a.HTTPExecutor.Execute(reqCtx, config, vars)
		reqCancel()

		if err != nil {
			if pollCtx.Err() != nil {
				return fmt.Errorf("polling timed out after %d attempts", attempt)
			}
			return fmt.Errorf("poll request failed: %w", err)
		}

		// Check condition
		conditionMet, condErr := workflow.EvaluateCondition(result.Response.Body, waitFor)
		if condErr != nil {
			return fmt.Errorf("evaluating poll condition: %w", condErr)
		}

		if conditionMet {
			// Output the final successful response
			if verbose {
				if err := output.FormatVerbose(w, result.Request, result.Response, noColor); err != nil {
					return err
				}
			} else {
				outputFmt, _ := cmd.Flags().GetString("output")
				formatter := output.New(domain.OutputFormat(outputFmt), noColor)
				if err := formatter.FormatResponse(w, result.Response); err != nil {
					return err
				}
			}

			if trace {
				fmt.Fprintln(w)
				if err := output.FormatTrace(w, result.Response.Timing, noColor); err != nil {
					return err
				}
			}

			if result.Response.StatusCode >= 400 {
				if shouldFailOnError(cmd) {
					return domain.NewHTTPError(result.Response.StatusCode, nil)
				}
			}

			return nil
		}

		// Wait for the next poll interval or context cancellation
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("polling timed out after %d attempts", attempt)
		case <-time.After(pollInterval):
			// Continue to next attempt
		}
	}
}
