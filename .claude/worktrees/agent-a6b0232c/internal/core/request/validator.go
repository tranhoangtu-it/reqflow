package request

import (
	"fmt"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ValidateRequest validates a fully resolved HTTPRequest.
func ValidateRequest(req domain.HTTPRequest) error {
	return validateMethodAndURL(req.Method, req.URL)
}

// ValidateConfig validates a RequestConfig before building.
func ValidateConfig(config domain.RequestConfig) error {
	return validateMethodAndURL(config.Method, config.URL)
}

// validateMethodAndURL is the shared validation logic for method and URL checks.
func validateMethodAndURL(method domain.HTTPMethod, rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("%w", domain.ErrEmptyURL)
	}

	if !method.IsValid() {
		return fmt.Errorf("%w: %s", domain.ErrInvalidMethod, method)
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("%w: %s", domain.ErrInvalidURL, rawURL)
	}

	return nil
}
