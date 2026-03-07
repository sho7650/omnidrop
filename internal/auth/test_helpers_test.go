package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	// testSecret is a 32-character secret for testing JWT operations
	testSecret = "test-secret-key-for-jwt-testing!"
	// testLegacyToken is a sample legacy token for testing
	testLegacyToken = "test-legacy-token-12345"
)

// newTestJWTManager creates a JWTManager for testing purposes
func newTestJWTManager() *JWTManager {
	return NewJWTManager(testSecret)
}

// newTestJWTManagerWithSecret creates a JWTManager with a custom secret
func newTestJWTManagerWithSecret(secret string) *JWTManager {
	return NewJWTManager(secret)
}

// newTestOAuthClient creates an OAuthClient for testing purposes
func newTestOAuthClient(clientID string, scopes []string) *OAuthClient {
	return &OAuthClient{
		ClientID:         clientID,
		ClientSecretHash: hashPassword(nil, "test-secret"),
		Name:             "Test Client",
		Scopes:           scopes,
		CreatedAt:        time.Now(),
		Disabled:         false,
	}
}

// generateValidToken generates a valid JWT token for testing
func generateValidToken(t *testing.T, jm *JWTManager, client *OAuthClient) string {
	t.Helper()
	token, err := jm.GenerateToken(client, 1*time.Hour)
	require.NoError(t, err, "failed to generate valid token")
	return token
}

// generateExpiredToken generates an expired JWT token for testing
func generateExpiredToken(t *testing.T, jm *JWTManager, client *OAuthClient) string {
	t.Helper()
	// Generate token with negative expiry to create an already-expired token
	now := time.Now()
	expiresAt := now.Add(-1 * time.Hour) // Expired 1 hour ago

	claims := jwt.MapClaims{
		"iss":       DefaultIssuer,
		"sub":       client.ClientID,
		"client_id": client.ClientID,
		"scopes":    client.Scopes,
		"iat":       now.Add(-2 * time.Hour).Unix(), // Issued 2 hours ago
		"exp":       expiresAt.Unix(),
		"jti":       "test-jti-expired",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jm.secret)
	require.NoError(t, err, "failed to generate expired token")
	return tokenString
}

// generateTokenWithCustomClaims generates a token with custom claims for testing
func generateTokenWithCustomClaims(t *testing.T, jm *JWTManager, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jm.secret)
	require.NoError(t, err, "failed to generate token with custom claims")
	return tokenString
}

// generateTokenWithWrongSecret generates a token signed with a different secret
func generateTokenWithWrongSecret(t *testing.T, client *OAuthClient) string {
	t.Helper()
	wrongJM := newTestJWTManagerWithSecret("wrong-secret-key-for-testing!!")
	return generateValidToken(t, wrongJM, client)
}

// generateTokenWithDifferentIssuer generates a token with a different issuer
func generateTokenWithDifferentIssuer(t *testing.T, jm *JWTManager, client *OAuthClient) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       "wrong-issuer",
		"sub":       client.ClientID,
		"client_id": client.ClientID,
		"scopes":    client.Scopes,
		"iat":       now.Unix(),
		"exp":       now.Add(1 * time.Hour).Unix(),
		"jti":       "test-jti-wrong-issuer",
	}
	return generateTokenWithCustomClaims(t, jm, claims)
}

// generateTokenNearExpiry generates a token that expires in the specified duration
func generateTokenNearExpiry(t *testing.T, jm *JWTManager, client *OAuthClient, expiresIn time.Duration) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       DefaultIssuer,
		"sub":       client.ClientID,
		"client_id": client.ClientID,
		"scopes":    client.Scopes,
		"iat":       now.Unix(),
		"exp":       now.Add(expiresIn).Unix(),
		"jti":       "test-jti-near-expiry",
	}
	return generateTokenWithCustomClaims(t, jm, claims)
}

// hashPassword generates a bcrypt hash of the given password
// If t is nil, it panics on error (for use in struct initialization)
func hashPassword(t *testing.T, password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if t != nil {
			t.Helper()
			require.NoError(t, err, "failed to hash password")
		} else {
			panic("failed to hash password: " + err.Error())
		}
	}
	return string(hash)
}

