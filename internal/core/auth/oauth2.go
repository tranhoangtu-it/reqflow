package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// BuildTokenRequest constructs an HTTPRequest for fetching an OAuth2 token.
// This is the pure, testable part of the OAuth2 flow.
func BuildTokenRequest(config domain.OAuth2Config) domain.HTTPRequest {
	form := url.Values{}
	form.Set("grant_type", config.GrantType)
	form.Set("client_id", config.ClientID)
	form.Set("client_secret", config.ClientSecret)

	if config.Scope != "" {
		form.Set("scope", config.Scope)
	}

	// For password grant, include username and password.
	if config.GrantType == "password" {
		form.Set("username", config.Username)
		form.Set("password", config.Password)
	}

	return domain.HTTPRequest{
		Method: domain.MethodPost,
		URL:    config.TokenURL,
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/x-www-form-urlencoded"},
		},
		Body: []byte(form.Encode()),
	}
}

// ApplyOAuth2 returns a new HTTPRequest with the OAuth2 Bearer token applied.
func ApplyOAuth2(req domain.HTTPRequest, token string) domain.HTTPRequest {
	result := req
	result.Headers = make([]domain.Header, len(req.Headers), len(req.Headers)+1)
	copy(result.Headers, req.Headers)
	result.Headers = append(result.Headers, domain.Header{
		Key:   "Authorization",
		Value: "Bearer " + token,
	})
	return result
}

// oauth2TokenResponse represents the JSON response from an OAuth2 token endpoint.
type oauth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// FetchOAuth2Token makes an HTTP call to the token endpoint and returns the access token.
// This is the impure part of the OAuth2 flow that requires network access.
func FetchOAuth2Token(ctx context.Context, config domain.OAuth2Config, httpClient driven.HTTPClient) (string, error) {
	tokenReq := BuildTokenRequest(config)

	resp, err := httpClient.Do(ctx, tokenReq)
	if err != nil {
		return "", fmt.Errorf("oauth2 token request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth2 token endpoint returned status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var tokenResp oauth2TokenResponse
	if err := json.Unmarshal(resp.Body, &tokenResp); err != nil {
		return "", fmt.Errorf("oauth2 token response parse error: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("oauth2 token error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("oauth2 token response missing access_token")
	}

	return tokenResp.AccessToken, nil
}
