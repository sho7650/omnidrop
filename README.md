# OmniDrop

A lightweight REST API server that bridges external applications with both OmniFocus and local file systems on macOS. Create OmniFocus tasks and manage files programmatically through a simple HTTP API with built-in observability, service management, and comprehensive testing infrastructure.

## Features

### Core Functionality
- 🚀 **Dual-purpose REST API** - Task creation + File operations
- 🔐 Bearer token authentication for security
- 📝 OmniFocus task creation with notes, projects, and tags
- 📂 **Secure file operations** with path traversal protection
- 🏗️ **Hierarchical project support** with path-based project references
- 🏷️ **Automatic tag creation** - creates new tags in OmniFocus if they don't exist
- ⏰ Automatic due date setting (end of current day)

### Integration & Operations
- 🍎 Native OmniFocus 4 integration via enhanced AppleScript
- 📊 **Prometheus metrics** for HTTP, OmniFocus, and file operations
- 📝 **Structured logging** with request IDs and JSON output
- 🔧 **Environment-based configuration** with complete isolation (production/development/test)
- 🛠️ **Comprehensive build system** with Makefile and testing infrastructure

### Service Management
- 🔄 **Reliable LaunchAgent lifecycle management** with graceful shutdown
- 📋 **Smart plist updates** - preserves custom settings by default
- 💾 **Automatic backup** with timestamped plist backups (FORCE_PLIST=1)
- 🧪 **Complete testing suite** with isolated environments and production protection
- 📊 **Multi-environment support** with port validation and script isolation

## Quick Start

Get up and running in 3 commands:

```bash
# 1. Clone and setup
git clone https://github.com/sho7650/omnidrop.git
cd omnidrop

# 2. Configure (edit TOKEN)
cp .env.example .env

# 3. Install and start service
make install
```

Your OmniDrop server is now running at `http://localhost:8787`!

## Prerequisites

- macOS (required for OmniFocus and AppleScript)
- OmniFocus 4 or later installed and running
- Go 1.25.0 or later
- Python 3 (for JSON parsing in AppleScript)

## Installation

### Production Installation (Recommended)

Full installation with automatic service management and proper lifecycle handling:

```bash
# Clone repository
git clone https://github.com/sho7650/omnidrop.git
cd omnidrop

# Configure environment
cp .env.example .env
# Edit .env and set your secure TOKEN

# Install everything with proper service lifecycle
make install

# Update installation (preserves existing plist)
make install

# Force plist update (creates automatic backup)
make install FORCE_PLIST=1
```

The install process:
1. **🛑 Stops** existing service gracefully
2. **📤 Unloads** service completely to prevent conflicts
3. **📁 Installs** files (binary, script, plist)
4. **💾 Protects** existing plist configuration (unless FORCE_PLIST=1)
5. **🚀 Loads** with persistence (`-w` flag) for reliable startup
6. **✅ Verifies** service startup with user feedback

This installs:
- Binary: `~/bin/omnidrop-server` (with graceful shutdown support)
- AppleScript: `~/.local/share/omnidrop/omnidrop.applescript`
- LaunchAgent: `~/Library/LaunchAgents/com.oshiire.omnidrop.plist`
- Logs: `~/.local/log/omnidrop/`
- Files: `~/.local/share/omnidrop/files/` (configurable via OMNIDROP_FILES_DIR)

**Plist Update Behavior:**
- **Default (`make install`)**: Preserves existing plist to protect custom settings
- **Force (`FORCE_PLIST=1`)**: Updates plist with automatic timestamped backup
- **New installation**: Always creates plist from template

### Development Installation

For development and testing with complete environment isolation:

```bash
# Clone repository
git clone https://github.com/sho7650/omnidrop.git
cd omnidrop

# Install dependencies
make deps

# Configure environment
cp .env.example .env
# Edit .env and set your TOKEN

# Run isolated development server (port 8788)
make dev-isolated

# Or run with direct go run
TOKEN=dev-token make dev
```

## Environment Management

OmniDrop supports complete environment separation for safe development and testing:

### Environment Variables

**Required:**
- `TOKEN`: Bearer token for API authentication

**Environment Control:**
- `OMNIDROP_ENV`: Environment mode (`production`, `development`, `test`)
- `PORT`: Server port (8787=production, 8788-8799=test range)
- `OMNIDROP_SCRIPT`: Explicit path to AppleScript file (overrides auto-detection)
- `OMNIDROP_FILES_DIR`: Base directory for file operations (default: `~/.local/share/omnidrop/files`)

### Environment-Specific Configuration

The server automatically selects appropriate configurations:

| Environment | Port Range | Script Location | Use Case |
|-------------|------------|-----------------|----------|
| `production` | 8787 (protected) | `~/.local/share/omnidrop/omnidrop.applescript` | Production deployment |
| `development` | Any | `./omnidrop.applescript` | Local development |
| `test` | 8788-8799 | Isolated test location | Automated testing |

### Multi-Environment Commands

```bash
# Development (port 8788)
make dev-isolated

# Staging/Testing (port 8790)
TOKEN=test-token make staging

# Production (port 8787, with confirmation)
TOKEN=prod-token make production-run
```

## Service Management

### Basic Service Control

```bash
# Start service
make start

# Stop service (now works reliably with graceful shutdown)
make stop

# Check service status
make status

# View logs (last 20 lines)
make logs

# Follow logs in real-time
make logs-follow

# Restart service
make stop && make start
```

### Advanced Service Management

```bash
# Check detailed service status
launchctl list | grep com.oshiire.omnidrop

# Manual service control (if needed)
launchctl stop com.oshiire.omnidrop
launchctl start com.oshiire.omnidrop

# Reinstall service with proper lifecycle
make install  # Now idempotent and safe to run multiple times
```

## Testing Infrastructure

### Comprehensive Testing Suite

```bash
# Run pre-flight validation (checks environment safety)
make test-preflight

# Run complete isolated test suite (7 comprehensive tests)
make test-isolated

# Run standard Go tests
make test
```

### Test Environments

**Isolated Testing:**
- Complete environment separation from production
- Temporary test directories and AppleScript files
- Port range validation (8788-8799)
- Automatic cleanup after testing

**Production Protection:**
- Port 8787 reserved for production only
- Production script path protection
- Environment variable validation
- Pre-flight safety checks

### Test Coverage

The test suite covers:
1. ✅ Simple task creation (no tags)
2. ✅ Task with existing tags
3. ✅ Task with automatic tag creation
4. ✅ Task with mixed existing and new tags
5. ✅ Task with hierarchical project assignment
6. ✅ Complex task with all features
7. ✅ Invalid request handling (authorization)

## Development

### Available Make Targets

**Development:**
```bash
make dev              # Run with go run (requires TOKEN env var)
make dev-isolated     # Run isolated development server (port 8788)
make run              # Build and run binary
make build            # Build binary only
make test             # Run standard Go tests
make test-isolated    # Run comprehensive isolated test suite
make test-preflight   # Run environment validation checks
```

**Environment Management:**
```bash
make staging          # Run staging environment (port 8790)
make production-run   # Run production server (port 8787, protected)
```

**Installation & Service:**
```bash
make install          # Install with proper LaunchAgent lifecycle
make uninstall        # Complete removal
make clean            # Clean build artifacts
make deps             # Download dependencies
```

**Service Management:**
```bash
make start            # Start service (now reliable)
make stop             # Stop service (graceful shutdown)
make status           # Check status
make logs             # Show recent logs
make logs-follow      # Follow logs real-time
make help             # Show all available targets
```

### Development Workflow

```bash
# Safe development cycle
TOKEN="dev-token" make dev-isolated    # Isolated development (port 8788)
make test-preflight                    # Validate environment safety
make test-isolated                     # Run comprehensive tests
make build                             # Build binary

# When ready for production testing
TOKEN="staging-token" make staging     # Staging test (port 8790)

# Production deployment
make install                           # Install with proper lifecycle
make status                            # Verify service
make logs                              # Check logs
```

### Testing Best Practices

**Always use test environments:**
```bash
# ✅ Good: Use test ports
make dev-isolated    # port 8788
make staging         # port 8790

# ❌ Bad: Never use production port for testing
# PORT=8787 # NEVER for testing
```

**Run validation before changes:**
```bash
make test-preflight  # Check environment safety
make test-isolated   # Run full test suite
```

## API Reference

### Create Task

**Endpoint:** `POST /tasks`

**Headers:**
- `Authorization: Bearer <your-token>` (required)
- `Content-Type: application/json` (required)

**Request Body:**
```json
{
  "title": "Task title",                    // Required: The task name
  "note": "Task description",               // Optional: Additional notes
  "project": "Project Name",                // Optional: Project name or hierarchical path
  "tags": ["tag1", "tag2"]                  // Optional: Array of tag names (auto-created if missing)
}
```

### Enhanced Project Support

**Hierarchical Projects:**
```json
{
  "title": "Review documentation",
  "project": "Getting Things Done/3. Projects/Work/Documentation"
}
```

**Simple Projects:**
```json
{
  "title": "Quick task",
  "project": "Work"
}
```

### Automatic Tag Management

**Automatic Tag Creation:**
```json
{
  "title": "New feature development",
  "tags": ["urgent", "new-feature", "auto-created-tag"]
}
```
- Existing tags are applied normally
- Non-existent tags are automatically created in OmniFocus
- Multi-strategy tag assignment with fallback mechanisms

**Response:**

Success (200 OK):
```json
{
  "status": "ok",
  "created": true
}
```

Error (4xx/5xx):
```json
{
  "status": "error",
  "created": false,
  "reason": "Error description"
}
```

### Create File

**Endpoint:** `POST /files`

**Headers:**
- `Authorization: Bearer <your-token>` (required)
- `Content-Type: application/json` (required)

**Request Body:**
```json
{
  "filename": "report.txt",                 // Required: Name of the file to create
  "content": "File content here",           // Required: Content to write to the file
  "directory": "reports/2025"               // Optional: Subdirectory path within base directory
}
```

**Security Features:**
- Path traversal protection prevents access outside base directory (`~/.local/share/omnidrop/files/` by default)
- Automatic directory creation for nested structures
- File overwrite protection (returns error if file exists)
- Configurable base directory via `OMNIDROP_FILES_DIR` environment variable

**Response:**

Success (200 OK):
```json
{
  "status": "ok",
  "created": true,
  "path": "reports/2025/report.txt"
}
```

Error (4xx/5xx):
```json
{
  "status": "error",
  "created": false,
  "reason": "Error description"
}
```

### Health Check

**Endpoint:** `GET /health`

Returns server status and version information.

### Metrics

**Endpoint:** `GET /metrics`

Exposes Prometheus-compatible metrics (no authentication required):
- HTTP request metrics (rate, duration, size)
- OmniFocus operation metrics (success/failure, duration)
- File operation metrics (success/failure, duration)

### Example Requests

#### Basic Task
```bash
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Review pull request"}'
```

#### Task with Hierarchical Project and Auto-Created Tags
```bash
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation",
    "note": "Update API docs and add examples",
    "project": "Getting Things Done/3. Projects/Work/Documentation",
    "tags": ["urgent", "documentation", "new-auto-tag"]
  }'
```

#### Development Testing (use test port)
```bash
curl -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test task creation",
    "project": "Test Project",
    "tags": ["test", "development"]
  }'
```

#### Create File
```bash
curl -X POST http://localhost:8787/files \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"filename":"report.txt","content":"Monthly report content"}'
```

#### Create File with Directory
```bash
curl -X POST http://localhost:8787/files \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "data.json",
    "content": "{\"status\":\"complete\"}",
    "directory": "reports/2025"
  }'
```

#### View Metrics
```bash
# View all metrics
curl http://localhost:8787/metrics

# Filter HTTP metrics
curl -s http://localhost:8787/metrics | grep "omnidrop_http"

# Filter OmniFocus metrics
curl -s http://localhost:8787/metrics | grep "omnidrop_omnifocus"

# Filter file operation metrics
curl -s http://localhost:8787/metrics | grep "omnidrop_files"
```

## Project Structure

```
omnidrop/
├── cmd/
│   └── omnidrop-server/
│       └── main.go                     # HTTP server with graceful shutdown
├── scripts/
│   ├── run-isolated-test.sh            # Comprehensive test execution
│   ├── test-preflight.sh               # Environment validation
│   └── test-tag-functionality.sh       # Tag handling tests
├── init/
│   └── launchd/
│       └── com.oshiire.omnidrop.plist  # LaunchAgent with environment config
├── build/                              # Build artifacts (created by make)
├── omnidrop.applescript                # Enhanced OmniFocus 4 integration
├── Makefile                            # Comprehensive build and service management
├── go.mod                              # Go module definition
├── go.sum                              # Dependency checksums
├── .env.example                        # Environment configuration template
├── .env                                # Local environment configuration (git-ignored)
├── CLAUDE.md                           # AI assistant guidance with implementation details
└── README.md                           # This file
```

## How It Works

### 1. HTTP Server (`cmd/omnidrop-server/main.go`)
- Receives POST requests with task data
- Validates Bearer token authentication
- **Environment-aware AppleScript path resolution**
- **Graceful shutdown with SIGTERM handling**
- **Production environment protection and validation**
- Passes task data to AppleScript bridge

### 2. Enhanced AppleScript Bridge (`omnidrop.applescript`)
- **OmniFocus 4 compatibility** with proper API usage
- **Hierarchical project resolution** with folder navigation
- **Multi-strategy tag assignment** with automatic tag creation
- **Comprehensive error handling** with detailed logging
- Parses JSON data using Python integration
- Sets due date to 23:59:59 of current day

### 3. Service Management (`Makefile` + `init/launchd/`)
- **Proper LaunchAgent lifecycle management** (stop → unload → install → load -w)
- **Idempotent installation** - safe to run multiple times
- **Environment-specific configuration** with production protection
- **Comprehensive testing infrastructure** with isolation
- Provides service control commands with validation
- Manages logs and configuration

## Important Notes

### Project and Tag Handling
- **Hierarchical Projects**: Use path format like `"Folder/Subfolder/Project"`
- **Simple Projects**: Use exact project name like `"Work"`
- **Automatic Tag Creation**: New tags are automatically created in OmniFocus
- **Tag Assignment**: Multi-strategy approach with fallback mechanisms
- **Due Dates**: All tasks are automatically set to due at end of current day (23:59:59)
- **Inbox**: Tasks without a project are created in the OmniFocus inbox

### Environment Safety
- **Production Protection**: Port 8787 is protected and requires explicit confirmation
- **Environment Isolation**: Complete separation between development, test, and production
- **Script Isolation**: Each environment uses its own AppleScript file
- **Test Port Range**: 8788-8799 reserved for testing to prevent production conflicts

### Service Management
- **Graceful Shutdown**: Server properly handles SIGTERM for reliable `launchctl stop`
- **Reliable Startup**: LaunchAgent loads with `-w` flag for persistence
- **Idempotent Install**: `make install` can be run multiple times safely
- **Service Validation**: Install process verifies successful startup

## Troubleshooting

### Service Issues

**Service won't start:**
```bash
make status                    # Check if service is loaded
make logs                      # Check for error messages
make test-preflight           # Validate environment
launchctl list | grep omnidrop  # Verify LaunchAgent status
```

**Service keeps crashing or won't stop:**
```bash
make logs                      # Check error logs
make uninstall                 # Clean removal with proper lifecycle
make install                   # Fresh installation with validation
```

**Installation conflicts:**
```bash
# The new install process handles this automatically:
make install                   # Stops → unloads → installs → loads → verifies
```

### Environment Issues

**Wrong environment or port conflicts:**
```bash
make test-preflight           # Comprehensive environment validation
echo $OMNIDROP_ENV           # Check current environment
echo $PORT                   # Check current port
```

**Production environment protection:**
```bash
# If accidentally using production resources:
unset OMNIDROP_ENV PORT OMNIDROP_SCRIPT  # Clear environment
make dev-isolated             # Use safe development environment
```

### API Issues

**"AppleScript error" Response:**
- Ensure OmniFocus 4 is running
- Check that project paths use correct folder hierarchy
- Verify Python 3 is installed: `python3 --version`
- Check AppleScript permissions in System Preferences
- Review AppleScript logs for detailed error information

**Tag creation issues:**
- New tags are automatically created - no manual setup needed
- Check OmniFocus tags section to verify new tags were created
- Review AppleScript logs for tag assignment details

**"Unauthorized" Response:**
- Verify the token in your `.env` file matches the request
- Ensure the Authorization header format is: `Bearer <token>`
- Check if `.env` file is being loaded: look for log message on startup

**"Server Won't Start":**
- Check if port is already in use: `lsof -i :8787`
- Run environment validation: `make test-preflight`
- Verify `.env` file exists and TOKEN is set
- Check file permissions on AppleScript file

### Development Issues

**Build failures:**
```bash
make clean                    # Clean build artifacts
make deps                     # Reinstall dependencies
go mod tidy                   # Clean module dependencies
```

**Test failures:**
```bash
make test-preflight          # Check environment setup
make test-isolated           # Run isolated test suite
# Check logs for specific test failure details
```

**Environment loading issues:**
```bash
TOKEN=test-token make dev-isolated  # Use explicit token
ls -la .env                  # Verify .env file exists
# Check OMNIDROP_ENV, PORT, OMNIDROP_SCRIPT variables
```

### Log Analysis

```bash
# View recent logs with timestamps
make logs

# Follow logs in real-time for debugging
make logs-follow

# View specific log files
tail -f ~/.local/log/omnidrop/stdout.log
tail -f ~/.local/log/omnidrop/stderr.log

# Check AppleScript execution logs for task creation details
grep -i "tag\|project\|error" ~/.local/log/omnidrop/stdout.log
```

## Security Considerations

- **Token Storage**: Never commit `.env` files with real tokens
- **Environment Separation**: Use different tokens for development/staging/production
- **HTTPS**: Use a reverse proxy (nginx/caddy) for HTTPS in production
- **Token Rotation**: Regularly rotate your authentication tokens
- **Network Access**: Bind to localhost only unless external access is required
- **File Permissions**: Ensure proper permissions on `.env` and log files
- **Service Security**: LaunchAgent runs as user, not root (safer)
- **Port Protection**: Production port 8787 is protected from accidental test usage
- **File Operations**: Path traversal protection prevents access outside configured base directory
- **Metrics Endpoint**: `/metrics` endpoint is public - use firewall rules if external access is restricted

## Uninstalling

Complete removal of OmniDrop with proper service lifecycle:

```bash
make uninstall
```

This safely removes:
- Stops and unloads the LaunchAgent service
- Binary from `~/bin/`
- AppleScript from `~/.local/share/omnidrop/`
- LaunchAgent from `~/Library/LaunchAgents/`
- Log directory from `~/.local/log/omnidrop/`
- Configuration from `~/.config/omnidrop/`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines
- Use `make test-preflight` before submitting changes
- Run `make test-isolated` to verify all functionality
- Follow environment separation best practices
- Test with multiple environments (development/staging)

## License

MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

Built with:
- [Go](https://golang.org/) - The programming language
- [go-chi](https://github.com/go-chi/chi) - Lightweight HTTP router
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [Prometheus client](https://github.com/prometheus/client_golang) - Metrics collection
- [slog-http](https://github.com/samber/slog-http) - Structured logging middleware
- [OmniFocus](https://www.omnigroup.com/omnifocus) - Task management application

## Recent Improvements

**v2.1 - Observability & File Operations (2025-01):**
- 📂 **File operations endpoint** with secure path handling and automatic directory creation
- 📊 **Prometheus metrics** for HTTP, OmniFocus, and file operations monitoring
- 📝 **Structured logging** with request IDs and JSON output for better observability
- 💾 **Smart plist updates** - preserves custom settings with FORCE_PLIST option
- 🔒 **Enhanced security** with file path traversal protection

**v2.0 - Environment Management & Reliability (2024-12):**
- ✨ **Hierarchical project support** with folder path navigation
- 🏷️ **Automatic tag creation** with multi-strategy assignment
- 🔧 **Complete environment separation** (production/development/test)
- 🛠️ **Proper LaunchAgent lifecycle management** with graceful shutdown
- 📋 **Idempotent installation** process with configuration protection
- 🧪 **Comprehensive testing infrastructure** with production protection
- 🔄 **Reliable service management** with startup verification

These improvements provide enterprise-grade reliability with complete environment isolation, robust service management, and comprehensive observability for production deployments.