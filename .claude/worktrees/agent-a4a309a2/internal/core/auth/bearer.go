package auth

import (
	"github.com/ye-kart/reqflow/internal/domain"
)

// ApplyBearer returns a new HTTPRequest with Bearer authentication applied.
// If Prefix is empty, it defaults to "Bearer".
func ApplyBearer(req domain.HTTPRequest, config domain.BearerAuthConfig) domain.HTTPRequest {
	prefix := config.Prefix
	if prefix == "" {
		prefix = "Bearer"
	}

	result := req
	result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+1)
	copy(result.Headers, req.Headers)
	result.Headers = append(result.Headers, domain.Header{
		Key:   "Authorization",
		Value: prefix + " " + config.Token,
	})

	return result
}
