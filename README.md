# OmniDrop

A lightweight REST API server that bridges external applications with OmniFocus on macOS. Create OmniFocus tasks programmatically through a simple HTTP API with built-in service management and easy installation.

## Features

- 🚀 Simple REST API for task creation
- 🔐 Bearer token authentication for security
- 📝 Support for task notes, projects, and tags
- ⏰ Automatic due date setting (end of current day)
- 🍎 Native OmniFocus integration via AppleScript
- 🔧 Environment-based configuration with `.env` support
- 🛠️ Comprehensive build system with Makefile
- 🔄 macOS LaunchAgent service management
- 📋 Automatic installation and service setup

## Quick Start

Get up and running in 3 commands:

```bash
# 1. Clone and setup
git clone https://github.com/yourusername/omnidrop.git
cd omnidrop

# 2. Configure (edit TOKEN)
cp .env.example .env

# 3. Install and start service
make install && make start
```

Your OmniDrop server is now running at `http://localhost:8787`!

## Prerequisites

- macOS (required for OmniFocus and AppleScript)
- OmniFocus 3 or later installed
- Go 1.25.0 or later
- Python 3 (for JSON parsing in AppleScript)

## Installation

### Production Installation (Recommended)

Full installation with automatic service management:

```bash
# Clone repository
git clone https://github.com/yourusername/omnidrop.git
cd omnidrop

# Configure environment
cp .env.example .env
# Edit .env and set your secure TOKEN

# Install everything (binary, service, AppleScript)
make install

# Start the service
make start

# Check status
make status
```

This installs:
- Binary: `~/bin/omnidrop-server`
- AppleScript: `~/.local/share/omnidrop/omnidrop.applescript`
- LaunchAgent: `~/Library/LaunchAgents/com.oshiire.omnidrop.plist`
- Logs: `~/.local/log/omnidrop/`

### Development Installation

For development and testing:

```bash
# Clone repository
git clone https://github.com/yourusername/omnidrop.git
cd omnidrop

# Install dependencies
make deps

# Configure environment
cp .env.example .env
# Edit .env and set your TOKEN

# Run development server
make dev
```

## Service Management

Control the OmniDrop service with simple commands:

```bash
# Start service
make start

# Stop service
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

### Service Status

Check if OmniDrop is running:
```bash
make status
# Output: Service status with PID if running

launchctl list | grep omnidrop
# Output: Shows service details if loaded
```

## Development

### Available Make Targets

**Development:**
```bash
make dev     # Run with go run (requires TOKEN env var)
make run     # Build and run binary
make build   # Build binary only
make test    # Run tests
```

**Installation:**
```bash
make install    # Install service and components
make uninstall  # Remove everything
make clean      # Clean build artifacts
make deps       # Download dependencies
```

**Service Management:**
```bash
make start       # Start service
make stop        # Stop service
make status      # Check status
make logs        # Show recent logs
make logs-follow # Follow logs real-time
```

### Development Workflow

```bash
# Development cycle
TOKEN="dev-token" make dev    # Direct development
make test                     # Run tests
make build                    # Build binary
make run                      # Test built binary

# When ready for production
make install                  # Install service
make start                    # Start service
make logs                     # Check logs
```

### Running Tests

```bash
make test
```

### Building for Distribution

```bash
# Build for current platform
make build

# Manual builds for specific platforms
GOOS=darwin GOARCH=amd64 go build -o build/bin/omnidrop-server-amd64 ./cmd/omnidrop-server
GOOS=darwin GOARCH=arm64 go build -o build/bin/omnidrop-server-arm64 ./cmd/omnidrop-server
```

## Configuration

Create a `.env` file in the project root with the following variables:

```env
# Required: Authentication token for API access
TOKEN=your-secret-token-here

# Optional: Server port (default: 8787)
PORT=8787
```

The server loads environment variables in this order:
1. Project root `.env` file (recommended)
2. Command directory `.env` file (fallback)
3. System environment variables

## API Reference

### Create Task

**Endpoint:** `POST /tasks`

**Headers:**
- `Authorization: Bearer <your-token>` (required)
- `Content-Type: application/json` (required)

**Request Body:**
```json
{
  "title": "Task title",        // Required: The task name
  "note": "Task description",   // Optional: Additional notes
  "project": "Project Name",    // Optional: Exact project name in OmniFocus
  "tags": ["tag1", "tag2"]      // Optional: Array of tag names
}
```

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

### Health Check

**Endpoint:** `GET /health`

Returns server status and version information.

### Example Requests

#### Basic Task
```bash
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Review pull request"}'
```

#### Task with Project and Tags
```bash
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation",
    "note": "Update API docs and add examples",
    "project": "Work",
    "tags": ["urgent", "documentation"]
  }'
```

#### Using HTTPie
```bash
http POST localhost:8787/tasks \
  Authorization:"Bearer your-secret-token" \
  title="Call client" \
  note="Discuss project timeline" \
  project="Sales"
```

## Project Structure

```
omnidrop/
├── cmd/
│   └── omnidrop-server/
│       └── main.go             # HTTP server implementation
├── init/
│   └── launchd/
│       └── com.oshiire.omnidrop.plist  # LaunchAgent template
├── build/                      # Build artifacts (created by make)
├── test/
│   └── integration/            # Integration tests
├── omnidrop.applescript        # OmniFocus integration script
├── Makefile                    # Build and service management
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── .env.example                # Environment configuration template
├── .env                        # Local environment configuration (git-ignored)
├── CLAUDE.md                   # AI assistant guidance
└── README.md                   # This file
```

## How It Works

1. **HTTP Server** (`cmd/omnidrop-server/main.go`):
   - Receives POST requests with task data
   - Validates Bearer token authentication
   - Passes task data to AppleScript bridge

2. **AppleScript Bridge** (`omnidrop.applescript`):
   - Parses JSON data using Python
   - Creates tasks in OmniFocus using AppleScript automation
   - Sets due date to 23:59:59 of current day
   - Handles project assignment and tag management

3. **Service Management** (`Makefile` + `init/launchd/`):
   - Installs binary and AppleScript to standard locations
   - Creates LaunchAgent for automatic startup
   - Provides service control commands
   - Manages logs and configuration

## Important Notes

- **Project Names**: Must exactly match existing projects in OmniFocus
- **Tags**: Non-existent tags are silently ignored
- **Due Dates**: All tasks are automatically set to due at end of current day (23:59:59)
- **Inbox**: Tasks without a project are created in the OmniFocus inbox
- **Security**: Always use HTTPS in production and keep your token secure
- **Service**: Automatically starts on login when installed via `make install`

## Troubleshooting

### Service Issues

**Service won't start:**
```bash
make status          # Check if service is loaded
make logs           # Check for error messages
launchctl list | grep omnidrop  # Verify LaunchAgent status
```

**Service keeps crashing:**
```bash
make logs           # Check error logs
make uninstall      # Clean removal
make install        # Fresh installation
```

### API Issues

**"AppleScript error" Response:**
- Ensure OmniFocus is running
- Check that project/tag names match exactly with OmniFocus
- Verify Python 3 is installed: `python3 --version`
- Check AppleScript permissions in System Preferences

**"Unauthorized" Response:**
- Verify the token in your `.env` file matches the request
- Ensure the Authorization header format is: `Bearer <token>`
- Check if `.env` file is being loaded: look for log message on startup

**"Server Won't Start":**
- Check if port is already in use: `lsof -i :8787`
- Verify `.env` file exists and TOKEN is set
- Check file permissions on AppleScript file

### Development Issues

**Build failures:**
```bash
make clean          # Clean build artifacts
make deps           # Reinstall dependencies
go mod tidy         # Clean module dependencies
```

**Environment loading issues:**
```bash
make dev            # Use TOKEN= env var directly
ls -la .env         # Verify .env file exists
cat .env            # Check .env contents (remove TOKEN first!)
```

### Log Analysis

```bash
# View recent logs
make logs

# Follow logs in real-time
make logs-follow

# View specific log files
tail -f ~/.local/log/omnidrop/stdout.log
tail -f ~/.local/log/omnidrop/stderr.log
```

## Security Considerations

- **Token Storage**: Never commit `.env` files with real tokens
- **HTTPS**: Use a reverse proxy (nginx/caddy) for HTTPS in production
- **Token Rotation**: Regularly rotate your authentication tokens
- **Network Access**: Bind to localhost only unless external access is required
- **File Permissions**: Ensure proper permissions on `.env` and log files
- **Service Security**: LaunchAgent runs as user, not root (safer)

## Uninstalling

Complete removal of OmniDrop:

```bash
make uninstall
```

This removes:
- Binary from `~/bin/`
- AppleScript from `~/.local/share/omnidrop/`
- LaunchAgent from `~/Library/LaunchAgents/`
- Log directory from `~/.local/log/omnidrop/`
- Configuration from `~/.config/omnidrop/`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

Built with:
- [Go](https://golang.org/) - The programming language
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [OmniFocus](https://www.omnigroup.com/omnifocus) - Task management application