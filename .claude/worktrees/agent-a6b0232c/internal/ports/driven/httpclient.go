package driven

import (
	"context"

	"github.com/ye-kart/reqflow/internal/domain"
)

// HTTPClient is the driven port for making HTTP requests.
type HTTPClient interface {
	Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error)
}
