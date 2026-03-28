package domain

// AuthType represents the type of authentication.
type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBasic  AuthType = "basic"
	AuthBearer AuthType = "bearer"
	AuthAPIKey AuthType = "apikey"
	AuthOAuth2 AuthType = "oauth2"
	AuthDigest AuthType = "digest"
	AuthAWS    AuthType = "aws"
)

// AuthConfig holds authentication configuration for a request.
type AuthConfig struct {
	Type   AuthType
	Basic  *BasicAuthConfig
	Bearer *BearerAuthConfig
	APIKey *APIKeyAuthConfig
	OAuth2 *OAuth2Config
	Digest *DigestAuthConfig
	AWS    *AWSAuthConfig
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

// OAuth2Config holds configuration for OAuth 2.0 authentication.
type OAuth2Config struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	GrantType    string // "client_credentials", "password"
	Username     string // for password grant
	Password     string // for password grant
}

// DigestAuthConfig holds credentials for HTTP Digest authentication.
type DigestAuthConfig struct {
	Username string
	Password string
}

// AWSAuthConfig holds configuration for AWS Signature V4 authentication.
type AWSAuthConfig struct {
	AccessKey    string
	SecretKey    string
	Region       string
	Service      string
	SessionToken string // optional
}
