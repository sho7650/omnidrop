package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidIssuer is returned when token issuer is invalid
	ErrInvalidIssuer = errors.New("invalid token issuer")
)

const (
	// DefaultIssuer is the default JWT issuer
	DefaultIssuer = "omnidrop"
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secret []byte
	issuer string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: DefaultIssuer,
	}
}

// GenerateToken generates a new JWT token for the given client
func (jm *JWTManager) GenerateToken(client *OAuthClient, expiry time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	// Generate unique token ID
	jti, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("failed to generate JTI: %w", err)
	}

	// Create claims
	claims := jwt.MapClaims{
		"iss":       jm.issuer,
		"sub":       client.ClientID,
		"client_id": client.ClientID,
		"scopes":    client.Scopes,
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
		"jti":       jti,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(jm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer
	issuer, ok := mapClaims["iss"].(string)
	if !ok || issuer != jm.issuer {
		return nil, ErrInvalidIssuer
	}

	// Convert to our Claims struct
	claims, err := mapClaimsToClaims(mapClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims, nil
}

// mapClaimsToClaims converts jwt.MapClaims to our Claims struct
func mapClaimsToClaims(mc jwt.MapClaims) (*Claims, error) {
	claims := &Claims{}

	// Extract required fields
	if clientID, ok := mc["client_id"].(string); ok {
		claims.ClientID = clientID
	} else {
		return nil, errors.New("missing client_id claim")
	}

	if sub, ok := mc["sub"].(string); ok {
		claims.Subject = sub
	}

	if iss, ok := mc["iss"].(string); ok {
		claims.Issuer = iss
	}

	// Extract scopes
	if scopesInterface, ok := mc["scopes"].([]interface{}); ok {
		scopes := make([]string, len(scopesInterface))
		for i, s := range scopesInterface {
			if str, ok := s.(string); ok {
				scopes[i] = str
			}
		}
		claims.Scopes = scopes
	}

	// Extract timestamps
	if iat, ok := mc["iat"].(float64); ok {
		claims.IssuedAt = int64(iat)
	}

	if exp, ok := mc["exp"].(float64); ok {
		claims.ExpiresAt = int64(exp)
	}

	if jti, ok := mc["jti"].(string); ok {
		claims.JWTID = jti
	}

	return claims, nil
}

// generateJTI generates a unique JWT ID
func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
