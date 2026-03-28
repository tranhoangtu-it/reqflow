package domain

// AuthType represents the type of authentication.
type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBasic  AuthType = "basic"
	AuthBearer AuthType = "bearer"
	AuthAPIKey AuthType = "apikey"
)

// AuthConfig holds authentication configuration for a request.
type AuthConfig struct {
	Type   AuthType
	Basic  *BasicAuthConfig
	Bearer *BearerAuthConfig
	APIKey *APIKeyAuthConfig
}

// BasicAuthConfig holds credentials for HTTP Basic authentication.
type BasicAuthConfig struct {
	Username string
	Password string
}

// BearerAuthConfig holds a token for Bearer authentication.
type BearerAuthConfig struct {
	Token  string
	Prefix string // defaults to "Bearer" if empty
}

// APIKeyLocation specifies where to send the API key.
type APIKeyLocation string

const (
	APIKeyInHeader APIKeyLocation = "header"
	APIKeyInQuery  APIKeyLocation = "query"
)

// APIKeyAuthConfig holds configuration for API key authentication.
type APIKeyAuthConfig struct {
	Key      string
	Value    string
	Location APIKeyLocation
}
