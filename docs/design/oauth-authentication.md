# OAuth 2.0 Authentication System Design

## Overview

OmniDrop v2.0のOAuth 2.0 Client Credentials Flow実装の詳細設計。

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     OAuth 2.0 Layer                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐      ┌──────────────────┐            │
│  │  Token Endpoint │      │  JWT Validator   │            │
│  │  /oauth/token   │      │  Middleware      │            │
│  └─────────────────┘      └──────────────────┘            │
│          │                         │                       │
│          ▼                         ▼                       │
│  ┌─────────────────┐      ┌──────────────────┐            │
│  │  Client Manager │      │  Scope Validator │            │
│  └─────────────────┘      └──────────────────┘            │
│          │                         │                       │
│          ▼                         ▼                       │
│  ┌─────────────────────────────────────┐                  │
│  │     OAuth Clients Repository        │                  │
│  │  (YAML + Optional SQLite Cache)     │                  │
│  └─────────────────────────────────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Token Endpoint (`/oauth/token`)

**Responsibility**: OAuth 2.0トークン発行

**Request**:
```json
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
&client_id=n8n-workflow-1
&client_secret=xxx
```

**Response (Success)**:
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "scope": "tasks:write files:write automation:video-convert"
}
```

**Response (Error)**:
```json
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "invalid_client",
  "error_description": "Client authentication failed"
}
```

**Error Codes**:
- `invalid_request`: リクエストパラメータ不正
- `invalid_client`: クライアント認証失敗
- `unauthorized_client`: 権限なし
- `unsupported_grant_type`: サポートされていないgrant_type

### 2. JWT Structure

**Header**:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload**:
```json
{
  "iss": "omnidrop",
  "sub": "n8n-workflow-1",
  "client_id": "n8n-workflow-1",
  "scopes": ["tasks:write", "files:write", "automation:video-convert"],
  "iat": 1705320000,
  "exp": 1705406400,
  "jti": "unique-token-id"
}
```

**Claims**:
- `iss` (Issuer): 常に "omnidrop"
- `sub` (Subject): client_id
- `client_id`: クライアント識別子
- `scopes`: 許可されたスコープ配列
- `iat` (Issued At): 発行時刻（Unix timestamp）
- `exp` (Expiration): 有効期限（Unix timestamp）
- `jti` (JWT ID): トークンID（オプション、リボケーション用）

### 3. Scope Definition

**Format**: `resource:action`

**標準スコープ**:
```yaml
# リソース別スコープ
tasks:read         # タスク読み取り（将来用）
tasks:write        # タスク作成
files:read         # ファイル読み取り（将来用）
files:write        # ファイル作成
automation:*       # すべての自動化コマンド
automation:{name}  # 特定コマンドのみ（例: automation:video-convert）

# システムスコープ
admin              # 管理者権限（CLI使用）
```

**Scope Matching Algorithm**:
```
Required: automation:video-convert
Client has: automation:*
→ MATCH (ワイルドカード)

Required: automation:video-convert
Client has: automation:video-convert
→ MATCH (完全一致)

Required: automation:video-convert
Client has: tasks:write
→ NO MATCH
```

### 4. Client Management

**Data Structure**:
```go
type OAuthClient struct {
    ClientID         string    `yaml:"client_id"`
    ClientSecretHash string    `yaml:"client_secret_hash"`
    Name             string    `yaml:"name"`
    Scopes           []string  `yaml:"scopes"`
    CreatedAt        time.Time `yaml:"created_at"`
    UpdatedAt        time.Time `yaml:"updated_at,omitempty"`
    Disabled         bool      `yaml:"disabled,omitempty"`
}
```

**Storage**: `~/.local/share/omnidrop/oauth-clients.yaml`

```yaml
clients:
  - client_id: n8n-workflow-1
    client_secret_hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
    name: "n8n Video Processing Workflow"
    scopes:
      - tasks:write
      - automation:video-convert
    created_at: 2025-01-15T10:00:00Z
    disabled: false
```

**Secret Hashing**:
```go
import "golang.org/x/crypto/bcrypt"

// 生成時
secret := generateRandomString(32)
hash, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)

// 検証時
err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(providedSecret))
if err != nil {
    return ErrInvalidCredentials
}
```

### 5. Authentication Middleware

**Flow**:
```
Request
  │
  ├─ Extract Bearer token from Authorization header
  │
  ├─ Parse and validate JWT
  │  ├─ Signature verification
  │  ├─ Expiration check
  │  └─ Issuer verification
  │
  ├─ Extract scopes from claims
  │
  ├─ Store in request context
  │
  └─ Continue to handler
```

**Implementation Pattern**:
```go
func OAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip OAuth endpoint itself
        if r.URL.Path == "/oauth/token" {
            next.ServeHTTP(w, r)
            return
        }
        
        // Extract token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Validate JWT
        claims, err := validateJWT(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Store claims in context
        ctx := context.WithValue(r.Context(), "oauth_claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 6. Scope Validation

**Handler-level Validation**:
```go
func RequireScopes(scopes ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("oauth_claims").(*Claims)
            
            if !hasRequiredScopes(claims.Scopes, scopes) {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
r.With(RequireScopes("tasks:write")).Post("/tasks", handlers.CreateTask)
r.With(RequireScopes("automation:video-convert")).Post("/automation/video-convert", handlers.VideoConvert)
```

**Wildcard Matching**:
```go
func hasRequiredScopes(clientScopes []string, requiredScopes []string) bool {
    for _, required := range requiredScopes {
        matched := false
        for _, clientScope := range clientScopes {
            if matchScope(clientScope, required) {
                matched = true
                break
            }
        }
        if !matched {
            return false
        }
    }
    return true
}

func matchScope(clientScope, required string) bool {
    // Exact match
    if clientScope == required {
        return true
    }
    
    // Wildcard match: automation:* matches automation:video-convert
    if strings.HasSuffix(clientScope, ":*") {
        prefix := strings.TrimSuffix(clientScope, "*")
        return strings.HasPrefix(required, prefix)
    }
    
    return false
}
```

## Configuration

**Environment Variables**:
```bash
# JWT署名鍵（必須）
OMNIDROP_JWT_SECRET=your-random-secret-key-minimum-32-chars

# トークン有効期限（デフォルト: 24h）
OMNIDROP_TOKEN_EXPIRY=24h

# クライアント設定ファイルパス
OMNIDROP_OAUTH_CLIENTS_FILE=~/.local/share/omnidrop/oauth-clients.yaml

# 後方互換性: 従来のBearer token認証を有効化
OMNIDROP_LEGACY_AUTH_ENABLED=true
```

## Migration Strategy

### Phase 1: OAuth実装（既存認証と並行）

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Try OAuth JWT first
        if claims, err := validateJWT(tokenString); err == nil {
            ctx := context.WithValue(r.Context(), "oauth_claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        
        // Fallback to legacy token
        if os.Getenv("OMNIDROP_LEGACY_AUTH_ENABLED") == "true" {
            expectedToken := os.Getenv("TOKEN")
            if tokenString == expectedToken {
                next.ServeHTTP(w, r)
                return
            }
        }
        
        http.Error(w, "Invalid token", http.StatusUnauthorized)
    })
}
```

### Phase 2: 完全移行

1. すべてのクライアントをOAuthに移行
2. `OMNIDROP_LEGACY_AUTH_ENABLED=false` に設定
3. 従来のTOKEN環境変数を削除

## Security Considerations

### JWT Secret Management
- 最小32文字のランダム文字列
- 環境変数で管理（`.env`ファイル、gitignore）
- 定期的なローテーション推奨

### Client Secret Storage
- bcrypt (cost=10) でハッシュ化
- 平文は生成時のみ表示、再表示不可
- 紛失時は再生成が必要

### Token Lifetime
- デフォルト24時間は妥当
- 短縮する場合はクライアント側でリフレッシュ実装が必要
- 長期トークン（30日など）は避ける

### Scope Management
- 最小権限の原則
- ワークフロー単位で必要最小限のスコープを付与
- `admin` スコープは管理用CLIのみ

## Testing Strategy

### Unit Tests
```go
func TestTokenGeneration(t *testing.T) {
    client := &OAuthClient{
        ClientID: "test-client",
        Scopes:   []string{"tasks:write"},
    }
    
    token, err := GenerateToken(client, 24*time.Hour)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    
    claims, err := ValidateJWT(token)
    assert.NoError(t, err)
    assert.Equal(t, "test-client", claims.ClientID)
    assert.Contains(t, claims.Scopes, "tasks:write")
}

func TestScopeMatching(t *testing.T) {
    tests := []struct {
        clientScopes   []string
        requiredScopes []string
        expected       bool
    }{
        {[]string{"automation:*"}, []string{"automation:video-convert"}, true},
        {[]string{"automation:video-convert"}, []string{"automation:video-convert"}, true},
        {[]string{"tasks:write"}, []string{"automation:video-convert"}, false},
    }
    
    for _, tt := range tests {
        result := hasRequiredScopes(tt.clientScopes, tt.requiredScopes)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Integration Tests
```go
func TestOAuthFlow(t *testing.T) {
    // Setup test server
    server := setupTestServer()
    defer server.Close()
    
    // Request token
    resp := requestToken(server.URL, "test-client", "test-secret")
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var tokenResp TokenResponse
    json.NewDecoder(resp.Body).Decode(&tokenResp)
    assert.NotEmpty(t, tokenResp.AccessToken)
    
    // Use token to access protected endpoint
    req := createRequest("/tasks", tokenResp.AccessToken)
    resp = server.Client().Do(req)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Metrics

```go
// OAuth-specific metrics
var (
    TokenIssuedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "omnidrop_oauth_tokens_issued_total",
            Help: "Total number of OAuth tokens issued",
        },
        []string{"client_id"},
    )
    
    TokenValidationTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "omnidrop_oauth_token_validations_total",
            Help: "Total number of token validation attempts",
        },
        []string{"result"}, // success, expired, invalid
    )
    
    ScopeValidationFailures = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "omnidrop_oauth_scope_validation_failures_total",
            Help: "Total number of scope validation failures",
        },
        []string{"client_id", "required_scope"},
    )
)
```

## References

- [RFC 6749: OAuth 2.0 Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 7519: JSON Web Token (JWT)](https://datatracker.ietf.org/doc/html/rfc7519)
- [RFC 6750: Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750)
