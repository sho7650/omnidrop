package auth

import "time"

// OAuthClient represents an OAuth 2.0 client
type OAuthClient struct {
	ClientID         string    `yaml:"client_id"`
	ClientSecretHash string    `yaml:"client_secret_hash"`
	Name             string    `yaml:"name"`
	Scopes           []string  `yaml:"scopes"`
	CreatedAt        time.Time `yaml:"created_at"`
	UpdatedAt        time.Time `yaml:"updated_at,omitempty"`
	Disabled         bool      `yaml:"disabled,omitempty"`
}

// OAuthConfig represents the OAuth clients configuration file structure
type OAuthConfig struct {
	Clients []OAuthClient `yaml:"clients"`
}

// TokenRequest represents an OAuth 2.0 token request
type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
}

// TokenResponse represents an OAuth 2.0 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

// ErrorResponse represents an OAuth 2.0 error response
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Claims represents JWT token claims
type Claims struct {
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
	// Standard JWT claims
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	JWTID     string `json:"jti,omitempty"`
}

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// ContextKeyClaims is the context key for storing OAuth claims
	ContextKeyClaims ContextKey = "oauth_claims"
)

// OAuth error codes as defined in RFC 6749
const (
	ErrorInvalidRequest          = "invalid_request"
	ErrorInvalidClient           = "invalid_client"
	ErrorInvalidGrant            = "invalid_grant"
	ErrorUnauthorizedClient      = "unauthorized_client"
	ErrorUnsupportedGrantType    = "unsupported_grant_type"
	ErrorInvalidScope            = "invalid_scope"
	ErrorInsufficientPermissions = "insufficient_permissions"
)
