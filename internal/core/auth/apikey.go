package auth

import (
	"github.com/ye-kart/reqflow/internal/domain"
)

// ApplyAPIKey returns a new HTTPRequest with API key authentication applied.
// When Location is "header" (or empty, which defaults to "header"), the key-value
// is added as a request header. When Location is "query", it is added as a query parameter.
func ApplyAPIKey(req domain.HTTPRequest, config domain.APIKeyAuthConfig) domain.HTTPRequest {
	location := config.Location
	if location == "" {
		location = domain.APIKeyInHeader
	}

	result := req

	if location == domain.APIKeyInQuery {
		result.QueryParams = make([]domain.QueryParam, len(req.QueryParams), len(req.QueryParams)+1)
		copy(result.QueryParams, req.QueryParams)
		result.QueryParams = append(result.QueryParams, domain.QueryParam{
			Key:   config.Key,
			Value: config.Value,
		})
		// Still copy headers to avoid shared slice
		result.Headers = make([]domain.Header, len(req.Headers))
		copy(result.Headers, req.Headers)
	} else {
		result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+1)
		copy(result.Headers, req.Headers)
		result.Headers = append(result.Headers, domain.Header{
			Key:   config.Key,
			Value: config.Value,
		})
		// Still copy query params to avoid shared slice
		result.QueryParams = make([]domain.QueryParam, len(req.QueryParams))
		copy(result.QueryParams, req.QueryParams)
	}

	return result
}
