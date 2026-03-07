package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "creates manager with valid secret",
			secret: testSecret,
		},
		{
			name:   "creates manager with minimum length secret",
			secret: "12345678901234567890123456789012", // 32 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJWTManager(tt.secret)

			assert.NotNil(t, jm, "JWTManager should not be nil")
			assert.Equal(t, []byte(tt.secret), jm.secret, "secret should be set correctly")
			assert.Equal(t, DefaultIssuer, jm.issuer, "issuer should be set to default")
		})
	}
}

func TestJWTManager_GenerateToken(t *testing.T) {
	jm := newTestJWTManager()
	client := newTestOAuthClient("test-client", []string{"tasks:write", "files:read"})

	t.Run("generates valid token", func(t *testing.T) {
		token, err := jm.GenerateToken(client, 1*time.Hour)

		require.NoError(t, err, "should not return error")
		assert.NotEmpty(t, token, "token should not be empty")

		// Verify token structure (3 parts separated by dots)
		parts := strings.Split(token, ".")
		assert.Len(t, parts, 3, "JWT should have 3 parts")
	})

	t.Run("includes required claims", func(t *testing.T) {
		tokenString, err := jm.GenerateToken(client, 1*time.Hour)
		require.NoError(t, err)

		// Parse token to verify claims
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})
		require.NoError(t, err)

		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok, "claims should be MapClaims")

		// Verify required claims
		assert.Equal(t, DefaultIssuer, claims["iss"], "issuer should match")
		assert.Equal(t, client.ClientID, claims["sub"], "subject should be client_id")
		assert.Equal(t, client.ClientID, claims["client_id"], "client_id claim should exist")
		assert.NotNil(t, claims["scopes"], "scopes claim should exist")
		assert.NotNil(t, claims["iat"], "iat claim should exist")
		assert.NotNil(t, claims["exp"], "exp claim should exist")
		assert.NotNil(t, claims["jti"], "jti claim should exist")
	})

	t.Run("sets correct expiry", func(t *testing.T) {
		expiry := 2 * time.Hour
		tokenString, err := jm.GenerateToken(client, expiry)
		require.NoError(t, err)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})
		require.NoError(t, err)

		claims := token.Claims.(jwt.MapClaims)
		iat := int64(claims["iat"].(float64))
		exp := int64(claims["exp"].(float64))

		// Expiry should be approximately iat + expiry duration
		expectedExp := iat + int64(expiry.Seconds())
		assert.InDelta(t, expectedExp, exp, 2, "expiry should be correct")
	})

	t.Run("encodes scopes correctly", func(t *testing.T) {
		expectedScopes := []string{"tasks:write", "files:read", "automation:*"}
		clientWithScopes := newTestOAuthClient("scope-client", expectedScopes)

		tokenString, err := jm.GenerateToken(clientWithScopes, 1*time.Hour)
		require.NoError(t, err)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})
		require.NoError(t, err)

		claims := token.Claims.(jwt.MapClaims)
		scopesInterface := claims["scopes"].([]interface{})

		scopes := make([]string, len(scopesInterface))
		for i, s := range scopesInterface {
			scopes[i] = s.(string)
		}

		assert.Equal(t, expectedScopes, scopes, "scopes should match")
	})

	t.Run("generates unique JTI for each token", func(t *testing.T) {
		token1, err := jm.GenerateToken(client, 1*time.Hour)
		require.NoError(t, err)

		token2, err := jm.GenerateToken(client, 1*time.Hour)
		require.NoError(t, err)

		// Parse both tokens
		parsedToken1, _ := jwt.Parse(token1, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})
		parsedToken2, _ := jwt.Parse(token2, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})

		jti1 := parsedToken1.Claims.(jwt.MapClaims)["jti"].(string)
		jti2 := parsedToken2.Claims.(jwt.MapClaims)["jti"].(string)

		assert.NotEqual(t, jti1, jti2, "each token should have unique JTI")
	})
}

func TestJWTManager_ValidateToken(t *testing.T) {
	jm := newTestJWTManager()
	client := newTestOAuthClient("test-client", []string{"tasks:write"})

	t.Run("validates valid token successfully", func(t *testing.T) {
		tokenString := generateValidToken(t, jm, client)

		claims, err := jm.ValidateToken(tokenString)

		require.NoError(t, err, "should not return error for valid token")
		assert.NotNil(t, claims, "claims should not be nil")
		assert.Equal(t, client.ClientID, claims.ClientID, "client_id should match")
		assert.Equal(t, client.Scopes, claims.Scopes, "scopes should match")
	})

	t.Run("rejects expired token", func(t *testing.T) {
		tokenString := generateExpiredToken(t, jm, client)

		claims, err := jm.ValidateToken(tokenString)

		assert.ErrorIs(t, err, ErrTokenExpired, "should return ErrTokenExpired")
		assert.Nil(t, claims, "claims should be nil for expired token")
	})

	t.Run("rejects token with wrong signature", func(t *testing.T) {
		tokenString := generateTokenWithWrongSecret(t, client)

		claims, err := jm.ValidateToken(tokenString)

		assert.Error(t, err, "should return error for wrong signature")
		assert.ErrorIs(t, err, ErrInvalidToken, "should be ErrInvalidToken")
		assert.Nil(t, claims, "claims should be nil")
	})

	t.Run("rejects token with wrong issuer", func(t *testing.T) {
		tokenString := generateTokenWithDifferentIssuer(t, jm, client)

		claims, err := jm.ValidateToken(tokenString)

		assert.ErrorIs(t, err, ErrInvalidIssuer, "should return ErrInvalidIssuer")
		assert.Nil(t, claims, "claims should be nil")
	})

	t.Run("rejects empty token", func(t *testing.T) {
		claims, err := jm.ValidateToken("")

		assert.Error(t, err, "should return error for empty token")
		assert.ErrorIs(t, err, ErrInvalidToken, "should be ErrInvalidToken")
		assert.Nil(t, claims, "claims should be nil")
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"not-a-jwt",
			"part1.part2",
			"part1.part2.part3.part4",
			"invalid.base64.here!",
		}

		for _, token := range malformedTokens {
			claims, err := jm.ValidateToken(token)

			assert.Error(t, err, "should return error for malformed token: %s", token)
			assert.Nil(t, claims, "claims should be nil for malformed token")
		}
	})

	t.Run("rejects token missing client_id claim", func(t *testing.T) {
		now := time.Now()
		claims := jwt.MapClaims{
			"iss":    DefaultIssuer,
			"sub":    "test-subject",
			"scopes": []string{"tasks:write"},
			"iat":    now.Unix(),
			"exp":    now.Add(1 * time.Hour).Unix(),
			"jti":    "test-jti",
			// Missing client_id
		}
		tokenString := generateTokenWithCustomClaims(t, jm, claims)

		result, err := jm.ValidateToken(tokenString)

		assert.Error(t, err, "should return error for missing client_id")
		assert.Nil(t, result, "claims should be nil")
	})
}

// Security-focused tests

func TestJWTManager_ValidateToken_AlgorithmConfusion(t *testing.T) {
	jm := newTestJWTManager()

	t.Run("rejects none algorithm", func(t *testing.T) {
		// Create a token with "none" algorithm (unsigned)
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"omnidrop","client_id":"test","scopes":["tasks:write"],"exp":9999999999}`))
		tokenString := header + "." + payload + "."

		claims, err := jm.ValidateToken(tokenString)

		assert.Error(t, err, "should reject none algorithm")
		assert.Nil(t, claims, "claims should be nil")
	})

	t.Run("only accepts HS256 algorithm", func(t *testing.T) {
		client := newTestOAuthClient("test-client", []string{"tasks:write"})

		// Generate a valid token and verify it uses HS256
		tokenString := generateValidToken(t, jm, client)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jm.secret, nil
		})
		require.NoError(t, err)

		assert.Equal(t, "HS256", token.Method.Alg(), "should use HS256 algorithm")
	})
}

func TestJWTManager_ValidateToken_TamperedPayload(t *testing.T) {
	jm := newTestJWTManager()
	client := newTestOAuthClient("test-client", []string{"tasks:write"})

	t.Run("detects payload tampering", func(t *testing.T) {
		tokenString := generateValidToken(t, jm, client)

		// Tamper with the payload
		parts := strings.Split(tokenString, ".")
		require.Len(t, parts, 3)

		// Decode, modify, and re-encode payload
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		require.NoError(t, err)

		// Change the payload (e.g., modify scopes)
		tamperedPayload := strings.Replace(string(payloadBytes), "tasks:write", "admin:*", 1)
		parts[1] = base64.RawURLEncoding.EncodeToString([]byte(tamperedPayload))

		tamperedToken := strings.Join(parts, ".")

		claims, err := jm.ValidateToken(tamperedToken)

		assert.Error(t, err, "should detect tampered payload")
		assert.Nil(t, claims, "claims should be nil for tampered token")
	})
}

func TestJWTManager_ValidateToken_BoundaryConditions(t *testing.T) {
	jm := newTestJWTManager()
	client := newTestOAuthClient("test-client", []string{"tasks:write"})

	t.Run("accepts token just before expiry", func(t *testing.T) {
		// Generate token that expires in 5 seconds
		tokenString := generateTokenNearExpiry(t, jm, client, 5*time.Second)

		claims, err := jm.ValidateToken(tokenString)

		assert.NoError(t, err, "should accept token before expiry")
		assert.NotNil(t, claims, "claims should not be nil")
	})

	t.Run("rejects token just after expiry", func(t *testing.T) {
		// Generate token that expired 1 second ago
		tokenString := generateTokenNearExpiry(t, jm, client, -1*time.Second)

		claims, err := jm.ValidateToken(tokenString)

		assert.ErrorIs(t, err, ErrTokenExpired, "should reject expired token")
		assert.Nil(t, claims, "claims should be nil")
	})
}

func TestMapClaimsToClaims(t *testing.T) {
	t.Run("converts all fields correctly", func(t *testing.T) {
		now := time.Now()
		mc := jwt.MapClaims{
			"client_id": "test-client",
			"sub":       "test-subject",
			"iss":       "test-issuer",
			"scopes":    []interface{}{"tasks:write", "files:read"},
			"iat":       float64(now.Unix()),
			"exp":       float64(now.Add(1 * time.Hour).Unix()),
			"jti":       "test-jti-123",
		}

		claims, err := mapClaimsToClaims(mc)

		require.NoError(t, err)
		assert.Equal(t, "test-client", claims.ClientID)
		assert.Equal(t, "test-subject", claims.Subject)
		assert.Equal(t, "test-issuer", claims.Issuer)
		assert.Equal(t, []string{"tasks:write", "files:read"}, claims.Scopes)
		assert.Equal(t, now.Unix(), claims.IssuedAt)
		assert.Equal(t, "test-jti-123", claims.JWTID)
	})

	t.Run("returns error when client_id is missing", func(t *testing.T) {
		mc := jwt.MapClaims{
			"sub":    "test-subject",
			"iss":    "test-issuer",
			"scopes": []interface{}{"tasks:write"},
		}

		claims, err := mapClaimsToClaims(mc)

		assert.Error(t, err, "should return error for missing client_id")
		assert.Nil(t, claims, "claims should be nil")
		assert.Contains(t, err.Error(), "client_id", "error should mention client_id")
	})

	t.Run("handles missing optional fields", func(t *testing.T) {
		mc := jwt.MapClaims{
			"client_id": "test-client",
			// Missing: sub, iss, scopes, iat, exp, jti
		}

		claims, err := mapClaimsToClaims(mc)

		require.NoError(t, err, "should not error for missing optional fields")
		assert.Equal(t, "test-client", claims.ClientID)
		assert.Empty(t, claims.Subject)
		assert.Empty(t, claims.Issuer)
		assert.Nil(t, claims.Scopes)
		assert.Zero(t, claims.IssuedAt)
		assert.Zero(t, claims.ExpiresAt)
		assert.Empty(t, claims.JWTID)
	})

	t.Run("handles empty scopes array", func(t *testing.T) {
		mc := jwt.MapClaims{
			"client_id": "test-client",
			"scopes":    []interface{}{},
		}

		claims, err := mapClaimsToClaims(mc)

		require.NoError(t, err)
		assert.Empty(t, claims.Scopes)
	})

	t.Run("handles non-string values in scopes array", func(t *testing.T) {
		mc := jwt.MapClaims{
			"client_id": "test-client",
			"scopes":    []interface{}{"valid-scope", 123, nil},
		}

		claims, err := mapClaimsToClaims(mc)

		require.NoError(t, err)
		// Non-string values should be filtered out
		assert.Len(t, claims.Scopes, 1)
		assert.Equal(t, "valid-scope", claims.Scopes[0])
	})
}

func TestGenerateJTI(t *testing.T) {
	t.Run("generates unique values", func(t *testing.T) {
		seen := make(map[string]bool)

		for i := 0; i < 100; i++ {
			jti, err := generateJTI()
			require.NoError(t, err)

			assert.NotEmpty(t, jti, "JTI should not be empty")
			assert.False(t, seen[jti], "JTI should be unique")
			seen[jti] = true
		}
	})

	t.Run("generates valid hex string", func(t *testing.T) {
		jti, err := generateJTI()
		require.NoError(t, err)

		// Should be 32 hex characters (16 bytes * 2)
		assert.Len(t, jti, 32, "JTI should be 32 characters")

		// Should only contain hex characters
		for _, c := range jti {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"JTI should only contain hex characters")
		}
	})
}
