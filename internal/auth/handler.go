package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"omnidrop/internal/observability"
)

// TokenHandler handles OAuth 2.0 token requests
type TokenHandler struct {
	repository  *Repository
	jwtManager  *JWTManager
	tokenExpiry time.Duration
	logger      *slog.Logger
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(repository *Repository, jwtManager *JWTManager, tokenExpiry time.Duration, logger *slog.Logger) *TokenHandler {
	return &TokenHandler{
		repository:  repository,
		jwtManager:  jwtManager,
		tokenExpiry: tokenExpiry,
		logger:      logger,
	}
}

// HandleToken handles POST /oauth/token requests
func (h *TokenHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req TokenRequest

	// Support both JSON and form-urlencoded
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, http.StatusBadRequest, ErrorInvalidRequest, "Invalid request body")
			return
		}
	} else {
		// Form-urlencoded
		if err := r.ParseForm(); err != nil {
			h.respondError(w, http.StatusBadRequest, ErrorInvalidRequest, "Invalid request body")
			return
		}
		req.GrantType = r.FormValue("grant_type")
		req.ClientID = r.FormValue("client_id")
		req.ClientSecret = r.FormValue("client_secret")
	}

	// Validate grant type
	if req.GrantType != "client_credentials" {
		h.respondError(w, http.StatusBadRequest, ErrorUnsupportedGrantType, "Only client_credentials grant type is supported")
		return
	}

	// Validate client credentials
	if req.ClientID == "" || req.ClientSecret == "" {
		h.respondError(w, http.StatusBadRequest, ErrorInvalidRequest, "Missing client_id or client_secret")
		return
	}

	// Authenticate client
	client, err := h.repository.Authenticate(req.ClientID, req.ClientSecret)
	if err != nil {
		h.logger.Warn("Client authentication failed",
			slog.String("client_id", req.ClientID),
			slog.String("error", err.Error()))

		// Record metrics
		observability.TokenValidationTotal.WithLabelValues("invalid").Inc()

		h.respondError(w, http.StatusUnauthorized, ErrorInvalidClient, "Client authentication failed")
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(client, h.tokenExpiry)
	if err != nil {
		h.logger.Error("Failed to generate token",
			slog.String("client_id", client.ClientID),
			slog.String("error", err.Error()))

		h.respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate token")
		return
	}

	// Record metrics
	observability.TokenIssuedTotal.WithLabelValues(client.ClientID).Inc()

	// Respond with token
	response := TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.tokenExpiry.Seconds()),
		Scope:       strings.Join(client.Scopes, " "),
	}

	h.logger.Info("Token issued",
		slog.String("client_id", client.ClientID),
		slog.Int("scopes_count", len(client.Scopes)))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// respondError sends an OAuth error response
func (h *TokenHandler) respondError(w http.ResponseWriter, status int, errorCode, description string) {
	response := ErrorResponse{
		Error:            errorCode,
		ErrorDescription: description,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
