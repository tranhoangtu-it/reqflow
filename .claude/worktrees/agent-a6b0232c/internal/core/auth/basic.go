package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/ye-kart/reqflow/internal/domain"
)

// ApplyBasic returns a new HTTPRequest with Basic authentication applied.
// It encodes username:password as Base64 and adds an Authorization header.
func ApplyBasic(req domain.HTTPRequest, config domain.BasicAuthConfig) domain.HTTPRequest {
	encoded := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", config.Username, config.Password)),
	)

	result := req
	result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+1)
	copy(result.Headers, req.Headers)
	result.Headers = append(result.Headers, domain.Header{
		Key:   "Authorization",
		Value: "Basic " + encoded,
	})

	return result
}
