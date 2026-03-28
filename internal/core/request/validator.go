package request

import (
	"fmt"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ValidateRequest validates a fully resolved HTTPRequest.
func ValidateRequest(req domain.HTTPRequest) error {
	if req.URL == "" {
		return fmt.Errorf("%w", domain.ErrEmptyURL)
	}

	if !req.Method.IsValid() {
		return fmt.Errorf("%w: %s", domain.ErrInvalidMethod, req.Method)
	}

	u, err := url.ParseRequestURI(req.URL)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("%w: %s", domain.ErrInvalidURL, req.URL)
	}

	return nil
}

// ValidateConfig validates a RequestConfig before building.
func ValidateConfig(config domain.RequestConfig) error {
	if config.URL == "" {
		return fmt.Errorf("%w", domain.ErrEmptyURL)
	}

	if !config.Method.IsValid() {
		return fmt.Errorf("%w: %s", domain.ErrInvalidMethod, config.Method)
	}

	u, err := url.ParseRequestURI(config.URL)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("%w: %s", domain.ErrInvalidURL, config.URL)
	}

	return nil
}
