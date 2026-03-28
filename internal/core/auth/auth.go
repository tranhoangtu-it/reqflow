package auth

import (
	"fmt"
	"time"

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

	case domain.AuthDigest:
		if config.Digest == nil {
			return req, fmt.Errorf("digest auth config is nil: %w", domain.ErrInvalidAuth)
		}
		// Digest auth requires a challenge-response flow. The dispatcher
		// returns the request unchanged; the caller must handle the 401
		// challenge and call ApplyDigest with the parsed challenge.
		return req, nil

	case domain.AuthAWS:
		if config.AWS == nil {
			return req, fmt.Errorf("aws auth config is nil: %w", domain.ErrInvalidAuth)
		}
		return SignAWS(req, *config.AWS, time.Now()), nil

	case domain.AuthOAuth2:
		// OAuth2 requires HTTP calls to fetch a token, so it cannot go
		// through this pure dispatcher. Handle it at the feature level.
		return req, fmt.Errorf("oauth2 auth must be handled separately (requires HTTP calls): %w", domain.ErrInvalidAuth)

	default:
		return req, fmt.Errorf("unknown auth type %q: %w", config.Type, domain.ErrInvalidAuth)
	}
}
