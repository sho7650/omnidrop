# OmniDrop CLI Tool Design

## Overview

OmniDropサーバーの管理用コマンドラインツール。OAuth クライアント管理、コマンド管理、サーバー制御を提供。

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   omnidrop-cli                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │
│  │   Client     │  │   Command    │  │   Config    │  │
│  │  Management  │  │  Management  │  │ Management  │  │
│  └──────────────┘  └──────────────┘  └─────────────┘  │
│         │                  │                 │         │
│         └──────────────────┴─────────────────┘         │
│                         │                               │
│                         ▼                               │
│              ┌──────────────────────┐                  │
│              │  Configuration File  │                  │
│              │      I/O Layer       │                  │
│              └──────────────────────┘                  │
│                         │                               │
│                         ▼                               │
│      ┌──────────────────────────────────────┐          │
│      │  ~/.local/share/omnidrop/            │          │
│      │  ├── oauth-clients.yaml              │          │
│      │  └── commands.yaml                   │          │
│      └──────────────────────────────────────┘          │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Command Structure

```
omnidrop-cli
├── client                      # OAuth client management
│   ├── add                     # Register new client
│   ├── list                    # List all clients
│   ├── show <client-id>        # Show client details
│   ├── update <client-id>      # Update client scopes
│   ├── remove <client-id>      # Remove client
│   └── regenerate <client-id>  # Regenerate secret
├── command                     # Command management
│   ├── list                    # List all commands
│   ├── show <command-name>     # Show command details
│   ├── validate                # Validate commands.yaml
│   └── test <command-name>     # Test command execution
├── token                       # Token management
│   ├── generate <client-id>    # Generate new token
│   └── decode <token>          # Decode JWT token
├── config                      # Configuration management
│   ├── path                    # Show config file paths
│   ├── validate                # Validate all config files
│   ├── reload                  # Reload server config
│   └── init                    # Initialize config files
└── server                      # Server management
    ├── status                  # Check server status
    ├── start                   # Start server
    ├── stop                    # Stop server
    └── restart                 # Restart server
```

## Command Specifications

### Client Management

#### `omnidrop-cli client add`

**Purpose**: 新しいOAuthクライアントを登録

**Usage**:
```bash
omnidrop-cli client add \
  --name "n8n Video Processing" \
  --scopes "tasks:write,automation:video-convert"

# Short form
omnidrop-cli client add -n "Defy Notifications" -s "automation:notify"
```

**Flags**:
- `--name, -n`: クライアント名（必須）
- `--scopes, -s`: カンマ区切りのスコープリスト（必須）
- `--output, -o`: 出力形式 [text|json|yaml] (デフォルト: text)

**Output**:
```
✅ Client created successfully

Client ID:     n8n-workflow-1
Client Secret: bQvZ9mK3xP7nR2wL4jT8hD6sG1fY5cV0
Name:          n8n Video Processing
Scopes:        tasks:write, automation:video-convert
Created:       2025-01-15 10:30:45

⚠️  Save the client secret securely. It cannot be retrieved later.
```

**JSON Output** (`--output json`):
```json
{
  "client_id": "n8n-workflow-1",
  "client_secret": "bQvZ9mK3xP7nR2wL4jT8hD6sG1fY5cV0",
  "name": "n8n Video Processing",
  "scopes": ["tasks:write", "automation:video-convert"],
  "created_at": "2025-01-15T10:30:45Z"
}
```

---

#### `omnidrop-cli client list`

**Purpose**: 登録されているクライアント一覧を表示

**Usage**:
```bash
omnidrop-cli client list

# With output format
omnidrop-cli client list --output json
```

**Flags**:
- `--output, -o`: 出力形式 [text|json|yaml] (デフォルト: text)
- `--show-disabled`: 無効化されたクライアントも表示

**Output**:
```
CLIENT ID            NAME                          SCOPES                                    CREATED
n8n-workflow-1       n8n Video Processing          tasks:write, automation:video-convert     2025-01-15
defy-notification    Defy Notification System      automation:notify                         2025-01-15
test-client          Test Client                   tasks:read, tasks:write                   2025-01-14
```

---

#### `omnidrop-cli client show`

**Purpose**: クライアントの詳細情報を表示

**Usage**:
```bash
omnidrop-cli client show n8n-workflow-1
```

**Output**:
```
Client Details
──────────────────────────────────────────
Client ID:     n8n-workflow-1
Name:          n8n Video Processing
Status:        Active
Scopes:        
  - tasks:write
  - automation:video-convert
Created:       2025-01-15 10:30:45
Last Updated:  2025-01-15 12:15:30
```

---

#### `omnidrop-cli client update`

**Purpose**: クライアントのスコープを更新

**Usage**:
```bash
# Add scopes
omnidrop-cli client update n8n-workflow-1 --add-scope "files:write"

# Remove scopes
omnidrop-cli client update n8n-workflow-1 --remove-scope "tasks:write"

# Replace all scopes
omnidrop-cli client update n8n-workflow-1 --scopes "automation:*"

# Update name
omnidrop-cli client update n8n-workflow-1 --name "New Name"
```

**Flags**:
- `--name, -n`: 新しい名前
- `--add-scope`: 追加するスコープ（複数指定可）
- `--remove-scope`: 削除するスコープ（複数指定可）
- `--scopes, -s`: スコープを完全置換

---

#### `omnidrop-cli client remove`

**Purpose**: クライアントを削除

**Usage**:
```bash
omnidrop-cli client remove n8n-workflow-1

# Force delete without confirmation
omnidrop-cli client remove n8n-workflow-1 --force
```

**Flags**:
- `--force, -f`: 確認なしで削除

**Output**:
```
⚠️  Are you sure you want to remove client 'n8n-workflow-1'? [y/N]: y
✅ Client removed successfully
```

---

#### `omnidrop-cli client regenerate`

**Purpose**: クライアントシークレットを再生成

**Usage**:
```bash
omnidrop-cli client regenerate n8n-workflow-1
```

**Output**:
```
⚠️  This will invalidate the current client secret. Continue? [y/N]: y

✅ Client secret regenerated successfully

New Client Secret: vP2wL9xK8nR7mT6jD4hG3sY5cF1bQ0zA

⚠️  Save the new secret securely. The old secret is now invalid.
```

---

### Command Management

#### `omnidrop-cli command list`

**Purpose**: 登録されているコマンド一覧を表示

**Usage**:
```bash
omnidrop-cli command list

# Filter by type
omnidrop-cli command list --type applescript

# Show full details
omnidrop-cli command list --verbose
```

**Flags**:
- `--type, -t`: コマンドタイプでフィルタ [applescript|shortcuts|shell|builtin]
- `--verbose, -v`: 詳細情報を表示
- `--output, -o`: 出力形式 [text|json|yaml]

**Output**:
```
NAME              TYPE          ENDPOINT                          SCOPES
open-url          applescript   /automation/open-url              automation:browser
run-shortcut      shortcuts     /automation/shortcuts/{name}      automation:shortcuts
video-convert     shell         /automation/video-convert         automation:video-convert
notify            applescript   /automation/notify                automation:notify
set-clipboard     applescript   /automation/clipboard/set         automation:clipboard
create-file       builtin       /files                            files:write
create-task       builtin       /tasks                            tasks:write
```

---

#### `omnidrop-cli command show`

**Purpose**: コマンドの詳細情報を表示

**Usage**:
```bash
omnidrop-cli command show video-convert
```

**Output**:
```
Command: video-convert
──────────────────────────────────────────
Name:        video-convert
Description: Convert video file using ffmpeg
Type:        shell
Endpoint:    /automation/video-convert
Script:      /usr/local/bin/convert-video.sh
Timeout:     600s (10m0s)
Working Dir: /tmp/omnidrop/video
Scopes:      automation:video-convert

Parameters:
  input_path (string, required)
    Description: Path to input video file
    Validation:  ^/Users/.*/Videos/.*\.(mp4|mov|avi|mkv)$
  
  output_format (string, optional, default: mp4)
    Description: Output video format
    Enum:        [mp4, webm, avi, mkv]
  
  quality (string, optional, default: medium)
    Description: Output quality preset
    Enum:        [low, medium, high, ultra]

Environment:
  FFMPEG_PATH=/usr/local/bin/ffmpeg
  FFMPEG_THREADS=4
```

---

#### `omnidrop-cli command validate`

**Purpose**: commands.yaml の検証

**Usage**:
```bash
omnidrop-cli command validate

# Specify custom config file
omnidrop-cli command validate --config /path/to/commands.yaml
```

**Flags**:
- `--config, -c`: 検証する設定ファイルパス

**Output (Success)**:
```
✅ Configuration is valid

Commands:     7
  AppleScript: 3
  Shortcuts:   1
  Shell:       1
  Builtin:     2

Total Endpoints: 7
```

**Output (Error)**:
```
❌ Configuration validation failed

Error in command 'video-convert':
  - Script file not found: /usr/local/bin/convert-video.sh

Error in command 'notify':
  - Parameter 'sound': default value 'invalid' not in enum
```

---

#### `omnidrop-cli command test`

**Purpose**: コマンドをテスト実行

**Usage**:
```bash
# Dry run (show what would be executed)
omnidrop-cli command test open-url \
  --param url=https://example.com \
  --dry-run

# Actual execution
omnidrop-cli command test notify \
  --param title="Test" \
  --param message="This is a test"
```

**Flags**:
- `--param, -p`: パラメータ (key=value形式、複数指定可)
- `--dry-run`: 実行せずにコマンド内容のみ表示
- `--timeout`: タイムアウト時間（設定ファイルの値を上書き）

**Output**:
```
🧪 Testing command: notify

Parameters:
  title:   "Test"
  message: "This is a test"
  sound:   "" (default)

Executing AppleScript...
──────────────────────────────────────────
on run argv
  set notifTitle to item 1 of argv
  set notifMessage to item 2 of argv
  set notifSound to item 3 of argv
  
  if notifSound is not "" then
    display notification notifMessage with title notifTitle sound name notifSound
  else
    display notification notifMessage with title notifTitle
  end if
end run
──────────────────────────────────────────

✅ Command executed successfully

Duration: 0.234s
Output:   (no output)
```

---

### Token Management

#### `omnidrop-cli token generate`

**Purpose**: アクセストークンを生成

**Usage**:
```bash
# Generate token with default expiry (24h)
omnidrop-cli token generate n8n-workflow-1

# Custom expiry
omnidrop-cli token generate n8n-workflow-1 --expires-in 7d

# Output as environment variable
omnidrop-cli token generate n8n-workflow-1 --export
```

**Flags**:
- `--expires-in, -e`: 有効期限 (例: 1h, 24h, 7d, 30d)
- `--export`: シェルのexport形式で出力
- `--output, -o`: 出力形式 [text|json|export]

**Output (default)**:
```
✅ Token generated successfully

Access Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJvbW5pZHJvcCIsInN1YiI6Im44bi13b3JrZmxvdy0xIiwiY2xpZW50X2lkIjoibjhuLXdvcmtmbG93LTEiLCJzY29wZXMiOlsidGFza3M6d3JpdGUiLCJhdXRvbWF0aW9uOnZpZGVvLWNvbnZlcnQiXSwiaWF0IjoxNzM3MDEyMDAwLCJleHAiOjE3MzcwOTg0MDB9.xyz123

Client ID:    n8n-workflow-1
Scopes:       tasks:write, automation:video-convert
Expires:      2025-01-16 10:30:45 (in 24 hours)
```

**Output (--export)**:
```bash
export OMNIDROP_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export OMNIDROP_CLIENT_ID="n8n-workflow-1"
export OMNIDROP_EXPIRES_AT="2025-01-16T10:30:45Z"
```

Usage:
```bash
# Set token in current shell
eval $(omnidrop-cli token generate n8n-workflow-1 --export)

# Use token
curl -H "Authorization: Bearer $OMNIDROP_TOKEN" http://localhost:8787/tasks
```

---

#### `omnidrop-cli token decode`

**Purpose**: JWTトークンをデコード

**Usage**:
```bash
omnidrop-cli token decode "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# From environment variable
omnidrop-cli token decode "$OMNIDROP_TOKEN"
```

**Output**:
```
Token Details
──────────────────────────────────────────
Header:
  Algorithm: HS256
  Type:      JWT

Payload:
  Issuer:    omnidrop
  Subject:   n8n-workflow-1
  Client ID: n8n-workflow-1
  Scopes:    
    - tasks:write
    - automation:video-convert
  Issued At:  2025-01-15 10:30:45
  Expires At: 2025-01-16 10:30:45
  
Status:      ✅ Valid (expires in 23h 45m)
```

---

### Config Management

#### `omnidrop-cli config path`

**Purpose**: 設定ファイルのパスを表示

**Usage**:
```bash
omnidrop-cli config path
```

**Output**:
```
Configuration Paths
──────────────────────────────────────────
OAuth Clients: /Users/sho/.local/share/omnidrop/oauth-clients.yaml
Commands:      /Users/sho/.local/share/omnidrop/commands.yaml

Environment Variables:
  OMNIDROP_OAUTH_CLIENTS_FILE: (not set)
  OMNIDROP_COMMANDS_FILE:      (not set)
  OMNIDROP_JWT_SECRET:         ✅ set
```

---

#### `omnidrop-cli config validate`

**Purpose**: すべての設定ファイルを検証

**Usage**:
```bash
omnidrop-cli config validate
```

**Output**:
```
Validating configuration files...

✅ OAuth Clients (/Users/sho/.local/share/omnidrop/oauth-clients.yaml)
   Clients: 3
   All client IDs are unique
   All client secrets are properly hashed

✅ Commands (/Users/sho/.local/share/omnidrop/commands.yaml)
   Commands: 7
   All command names are unique
   All endpoints are unique
   All scripts exist and are executable

✅ Configuration is valid
```

---

#### `omnidrop-cli config init`

**Purpose**: 設定ファイルを初期化

**Usage**:
```bash
omnidrop-cli config init

# Force overwrite existing files
omnidrop-cli config init --force
```

**Flags**:
- `--force, -f`: 既存ファイルを上書き

**Output**:
```
Initializing OmniDrop configuration...

📁 Creating directory: /Users/sho/.local/share/omnidrop
✅ Created oauth-clients.yaml (empty)
✅ Created commands.yaml (with examples)

Generated JWT secret: vX3mK9nR2wL7jT5hD8sG4cY1fQ6bP0zA

Add this to your .env file:
──────────────────────────────────────────
OMNIDROP_JWT_SECRET=vX3mK9nR2wL7jT5hD8sG4cY1fQ6bP0zA
──────────────────────────────────────────

⚠️  Save the JWT secret securely and add it to your server's environment.

Next steps:
1. Add the JWT secret to your .env file
2. Create your first OAuth client: omnidrop-cli client add
3. Start the server: omnidrop-cli server start
```

---

#### `omnidrop-cli config reload`

**Purpose**: サーバーの設定をリロード

**Usage**:
```bash
omnidrop-cli config reload

# With custom server URL
omnidrop-cli config reload --server http://localhost:8788
```

**Flags**:
- `--server, -s`: サーバーURL (デフォルト: http://localhost:8787)
- `--token, -t`: 管理者トークン (環境変数 OMNIDROP_ADMIN_TOKEN から取得)

**Output**:
```
Reloading server configuration...

Connecting to http://localhost:8787...
✅ Configuration reloaded successfully

Loaded:
  OAuth Clients: 3
  Commands:      7
```

---

### Server Management

#### `omnidrop-cli server status`

**Purpose**: サーバーのステータスを確認

**Usage**:
```bash
omnidrop-cli server status

# With custom server URL
omnidrop-cli server status --server http://localhost:8788
```

**Output (Running)**:
```
Server Status
──────────────────────────────────────────
Status:    ✅ Running
URL:       http://localhost:8787
PID:       12345
Uptime:    2h 15m 30s
Version:   2.0.0

Health Check:
  /health:   ✅ OK
  /metrics:  ✅ OK (12 metrics available)
```

**Output (Not Running)**:
```
Server Status
──────────────────────────────────────────
Status:    ❌ Not Running

To start the server:
  omnidrop-cli server start
```

---

#### `omnidrop-cli server start`

**Purpose**: サーバーを起動

**Usage**:
```bash
omnidrop-cli server start

# Start in foreground (for debugging)
omnidrop-cli server start --foreground

# Custom port
omnidrop-cli server start --port 8788
```

**Flags**:
- `--foreground, -f`: フォアグラウンドで起動
- `--port, -p`: ポート番号
- `--env-file`: .envファイルパス

**Output**:
```
Starting OmniDrop server...

Loading configuration...
  OAuth Clients: /Users/sho/.local/share/omnidrop/oauth-clients.yaml (3 clients)
  Commands:      /Users/sho/.local/share/omnidrop/commands.yaml (7 commands)

Starting server on port 8787...
✅ Server started successfully (PID: 12345)

Logs: tail -f /Users/sho/.local/share/omnidrop/logs/server.log
```

---

#### `omnidrop-cli server stop`

**Purpose**: サーバーを停止

**Usage**:
```bash
omnidrop-cli server stop

# Force kill
omnidrop-cli server stop --force
```

**Output**:
```
Stopping OmniDrop server...

Sending shutdown signal to PID 12345...
Waiting for graceful shutdown...
✅ Server stopped successfully
```

---

#### `omnidrop-cli server restart`

**Purpose**: サーバーを再起動

**Usage**:
```bash
omnidrop-cli server restart
```

**Output**:
```
Restarting OmniDrop server...

Stopping current instance (PID: 12345)...
✅ Server stopped

Starting new instance...
✅ Server started successfully (PID: 12456)
```

---

## Implementation Structure

```
cmd/omnidrop-cli/
├── main.go                 # Entry point
├── commands/
│   ├── client/
│   │   ├── add.go
│   │   ├── list.go
│   │   ├── show.go
│   │   ├── update.go
│   │   ├── remove.go
│   │   └── regenerate.go
│   ├── command/
│   │   ├── list.go
│   │   ├── show.go
│   │   ├── validate.go
│   │   └── test.go
│   ├── token/
│   │   ├── generate.go
│   │   └── decode.go
│   ├── config/
│   │   ├── path.go
│   │   ├── validate.go
│   │   ├── init.go
│   │   └── reload.go
│   └── server/
│       ├── status.go
│       ├── start.go
│       ├── stop.go
│       └── restart.go
└── pkg/
    ├── client/             # OAuth client operations
    ├── command/            # Command operations
    ├── token/              # Token operations
    ├── config/             # Config file I/O
    ├── server/             # Server control
    └── ui/                 # Terminal UI helpers
        ├── table.go
        ├── colors.go
        └── prompt.go
```

## Dependencies

```go
// CLI framework
github.com/spf13/cobra        // Command-line interface

// Configuration
gopkg.in/yaml.v3              // YAML parsing

// JWT
github.com/golang-jwt/jwt/v5  // JWT generation/validation

// Security
golang.org/x/crypto/bcrypt    // Password hashing

// Terminal UI
github.com/fatih/color        // Colored output
github.com/olekukonko/tablewriter  // Table formatting
github.com/AlecAivazis/survey/v2   // Interactive prompts

// HTTP client
net/http                      // Server API communication
```

## Configuration Files

CLI自体の設定ファイル（オプション）:

**~/.config/omnidrop/cli-config.yaml**
```yaml
# Default server URL
server_url: http://localhost:8787

# Default output format
output_format: text

# Color output
color_output: true

# Admin token (for config reload)
admin_token_file: ~/.config/omnidrop/admin-token

# Custom config paths
oauth_clients_file: ~/.local/share/omnidrop/oauth-clients.yaml
commands_file: ~/.local/share/omnidrop/commands.yaml
```

## Security Considerations

### Client Secret Display
- 生成時のみ平文で表示
- 以降は表示不可（bcryptハッシュのみ保存）
- 紛失時は再生成が必要

### Admin Token
- `config reload`コマンドには admin スコープが必要
- トークンは環境変数または設定ファイルで管理
- CLIは admin トークン生成もサポート

### File Permissions
```bash
# OAuth clients file (sensitive)
chmod 600 ~/.local/share/omnidrop/oauth-clients.yaml

# Commands file (less sensitive)
chmod 644 ~/.local/share/omnidrop/commands.yaml
```

## User Experience

### Interactive Prompts
削除や再生成など破壊的操作には確認プロンプト:
```bash
$ omnidrop-cli client remove n8n-workflow-1
⚠️  Are you sure you want to remove client 'n8n-workflow-1'? [y/N]: y
```

### Colored Output
```go
// Success (green)
✅ Client created successfully

// Error (red)
❌ Configuration validation failed

// Warning (yellow)
⚠️  Save the client secret securely.

// Info (blue)
ℹ️  Token expires in 23h 45m
```

### Progress Indicators
長時間かかる操作にはプログレスバー:
```bash
Validating configuration files... ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100%
```

## Testing Strategy

### Unit Tests
```go
func TestClientAdd(t *testing.T) {
    // Test client creation
}

func TestScopeValidation(t *testing.T) {
    // Test scope format validation
}
```

### Integration Tests
```bash
# Test full workflow
$ ./test-cli.sh
Creating test client...
Generating token...
Testing server communication...
Cleaning up...
All tests passed!
```

## Installation

### Binary Distribution
```bash
# Download binary
curl -L https://github.com/example/omnidrop/releases/download/v2.0.0/omnidrop-cli-darwin-amd64 -o omnidrop-cli

# Make executable
chmod +x omnidrop-cli

# Move to PATH
sudo mv omnidrop-cli /usr/local/bin/

# Verify installation
omnidrop-cli --version
```

### From Source
```bash
# Clone repository
git clone https://github.com/example/omnidrop.git
cd omnidrop

# Build CLI
make build-cli

# Install
make install-cli
```

## Auto-completion

### Bash
```bash
# Generate completion script
omnidrop-cli completion bash > /usr/local/etc/bash_completion.d/omnidrop-cli

# Or add to ~/.bashrc
source <(omnidrop-cli completion bash)
```

### Zsh
```bash
# Generate completion script
omnidrop-cli completion zsh > ~/.zsh/completion/_omnidrop-cli

# Add to ~/.zshrc
fpath=(~/.zsh/completion $fpath)
autoload -Uz compinit && compinit
```

### Fish
```bash
omnidrop-cli completion fish > ~/.config/fish/completions/omnidrop-cli.fish
```

## Examples

### Quick Start Workflow
```bash
# 1. Initialize configuration
omnidrop-cli config init

# 2. Create first client
omnidrop-cli client add -n "My Workflow" -s "tasks:write,automation:*"

# 3. Generate token
omnidrop-cli token generate my-workflow-1

# 4. Test server
omnidrop-cli server status

# 5. Start server
omnidrop-cli server start
```

### Daily Operations
```bash
# Check server status
omnidrop-cli server status

# List clients
omnidrop-cli client list

# Add new scope to existing client
omnidrop-cli client update n8n-workflow-1 --add-scope "files:write"

# Test a command
omnidrop-cli command test notify --param title="Test" --param message="Hello"

# Reload config after changes
omnidrop-cli config reload
```

## Future Enhancements

### Interactive Mode
```bash
omnidrop-cli interactive

Welcome to OmniDrop CLI!
> client add
Name: My Workflow
Scopes (comma-separated): tasks:write, automation:notify
✅ Client created: my-workflow-1
> exit
```

### Logs Viewer
```bash
omnidrop-cli logs --follow
omnidrop-cli logs --tail 100
omnidrop-cli logs --filter error
```

### Metrics Viewer
```bash
omnidrop-cli metrics
omnidrop-cli metrics --watch
```

### Backup/Restore
```bash
omnidrop-cli backup export backup.tar.gz
omnidrop-cli backup restore backup.tar.gz
```
