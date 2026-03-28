package auth

import (
	"fmt"

	"github.com/ye-kart/reqflow/internal/domain"
)

// Apply routes the request to the correct auth provider based on config.Type.
// Returns the request unchanged if config is nil or Type is "none".
// Returns domain.ErrInvalidAuth for unknown types or missing sub-configs.
func Apply(req domain.HTTPRequest, config *domain.AuthConfig) (domain.HTTPRequest, error) {
	if config == nil {
		return req, nil
	}

	switch config.Type {
	case domain.AuthNone, "":
		return req, nil

	case domain.AuthBasic:
		if config.Basic == nil {
			return req, fmt.Errorf("basic auth config is nil: %w", domain.ErrInvalidAuth)
		}
		return ApplyBasic(req, *config.Basic), nil

	case domain.AuthBearer:
		if config.Bearer == nil {
			return req, fmt.Errorf("bearer auth config is nil: %w", domain.ErrInvalidAuth)
		}
		return ApplyBearer(req, *config.Bearer), nil

	case domain.AuthAPIKey:
		if config.APIKey == nil {
			return req, fmt.Errorf("api key auth config is nil: %w", domain.ErrInvalidAuth)
		}
		return ApplyAPIKey(req, *config.APIKey), nil

	default:
		return req, fmt.Errorf("unknown auth type %q: %w", config.Type, domain.ErrInvalidAuth)
	}
}
