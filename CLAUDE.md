# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OmniDrop is a lightweight REST API server that provides Omnichannel Drop functionality - bridging external applications with both OmniFocus on macOS and local file systems. It receives HTTP POST requests for task creation in OmniFocus and file operations on the local file system.

## Architecture

The system consists of three main components:

1. **HTTP Server** (`main.go`): Go-based REST API that:
   - Listens on a configurable port (default: 8787)
   - Requires Bearer token authentication
   - Accepts POST requests to `/tasks` and `/files` endpoints
   - Validates JSON payloads and handles routing
   - **Environment-aware script resolution** for development/test isolation

2. **AppleScript Bridge** (`omnidrop.applescript`): Handles **OmniFocus 4** integration:
   - Creates tasks in OmniFocus with specified properties
   - Sets due dates to end of current day (23:59:59)
   - Supports hierarchical project assignment
   - **Multi-strategy tag management** with automatic tag creation

3. **Files Service**: Handles **Local File Operations**:
   - Creates files in specified directories within a secure base path
   - **Path traversal protection** prevents unauthorized file access
   - **Automatic directory creation** for nested folder structures
   - **Environment-based configuration** for file storage location

## Development Commands

```bash
# Build the server
go build -o omnidrop-server .

# Install to system (LaunchAgent)
make install

# Install with plist update (creates timestamped backup)
make install FORCE_PLIST=1

# Run the server (requires TOKEN environment variable)
TOKEN="your-secret-token" ./omnidrop-server

# IMPORTANT: For development/testing, NEVER use port 8787 (production port)
# Use alternative ports for testing:
PORT=8788 TOKEN="test-token" ./omnidrop-server
PORT=8789 TOKEN="test-token" ./omnidrop-server

# Test the API endpoint with simple project name (use test port!)
curl -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task","note":"Description","project":"Work","tags":["urgent"]}'

# Test with hierarchical project path (use test port!)
curl -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task","note":"Description","project":"Getting Things Done/3. Projects/お仕事/Enablement","tags":["urgent"]}'

# Test file creation endpoint (use test port!)
curl -X POST http://localhost:8788/files \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"filename":"report.txt","content":"Monthly report content"}'

# Test file creation with directory (use test port!)
curl -X POST http://localhost:8788/files \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"filename":"data.json","content":"{}","directory":"reports/2025"}'
```

## Environment Configuration

### Environment Variables

The server supports environment-based configuration for complete isolation:

**Required Variables:**
- `TOKEN`: Bearer token for API authentication

**Environment Control:**
- `OMNIDROP_ENV`: Environment mode (`production`, `development`, `test`)
- `PORT`: Server port (8787=production, 8788-8799=test range)
- `OMNIDROP_SCRIPT`: Explicit path to AppleScript file (overrides auto-detection)
- `OMNIDROP_FILES_DIR`: Base directory for file operations (default: `~/.local/share/omnidrop/files`)

### Environment-Specific Script Resolution

The server automatically selects the appropriate AppleScript file based on environment:

```
production:   ~/.local/share/omnidrop/omnidrop.applescript
development:  ./omnidrop.applescript (current directory)
test:         Uses OMNIDROP_SCRIPT environment variable
legacy:       Fallback to installed or current directory
```

### Production Protection

**CRITICAL**: Port 8787 is reserved for production only. Test environments MUST use ports 8788-8799 to prevent production interference.

## API Contract

### POST /tasks

Request body:
```json
{
  "title": "string (required)",
  "note": "string (optional)",
  "project": "string (optional) - supports hierarchical paths",
  "tags": ["string array (optional)"]
}
```

Response:
```json
{
  "status": "ok|error",
  "created": true|false,
  "reason": "string (only on error)"
}
```

### POST /files

Request body:
```json
{
  "filename": "string (required) - name of the file to create",
  "content": "string (required) - content to write to the file",
  "directory": "string (optional) - subdirectory path within base directory"
}
```

Response:
```json
{
  "status": "ok|error",
  "created": true|false,
  "path": "string (relative path of created file, on success)",
  "reason": "string (error message, only on error)"
}
```

**Security Features:**
- Path traversal protection prevents access outside base directory
- Automatic directory creation for nested structures
- File overwrite protection (returns error if file exists)

## Test Execution Environment

### Makefile Targets

The project includes comprehensive testing infrastructure with complete environment isolation:

```bash
# Pre-flight validation (checks environment safety)
make test-preflight

# Isolated test suite with temporary environment
make test-isolated

# Development server with explicit environment control (port 8788)
make dev-isolated

# Staging environment (port 8790)
make staging

# Protected production environment (port 8787)
make production-run
```

### Manual Testing Scripts

**Pre-flight Validation:**
```bash
# Validates environment safety before testing
./scripts/test-preflight.sh
```

**Isolated Test Execution:**
```bash
# Complete isolated test with cleanup
./scripts/run-isolated-test.sh
```

### Test Environment Features

- **Complete Isolation**: Creates temporary test directory with isolated AppleScript
- **Production Protection**: Validates that port 8787 is never used for testing
- **Automatic Cleanup**: Removes all temporary files and processes on completion
- **Environment Validation**: Pre-flight checks prevent dangerous test execution
- **Comprehensive Testing**: Tests all tag scenarios including auto-creation

## Key Implementation Details

### **OmniFocus 4 Compatibility**

- Compatible with OmniFocus 4 AppleScript API
- The server only accepts POST requests and returns 405 for other methods
- Authentication is mandatory via `Authorization: Bearer <token>` header
- AppleScript execution errors are captured and returned in API responses
- Tasks are automatically assigned today's date at 23:59:59 as due date

### **Hierarchical Project Support**

Projects can be referenced using paths (e.g., "Getting Things Done/3. Projects/お仕事/Enablement")
- Simple project names are supported for backward compatibility
- Project names must exactly match existing items in OmniFocus

### **Enhanced Tag Management**

- **Multi-strategy tag assignment** with fallback mechanisms
- **Automatic tag creation**: Non-existent tags are created in OmniFocus
- **Individual tag processing**: Prevents type conversion errors (-2700)
- **Graceful error handling**: Task creation continues even if some tags fail
- **Strategy A**: Individual tag assignment with `add tagRef to tags of newTask`
- **Strategy B**: Direct property assignment with `set tags to tags & {tagRef}`
- **Strategy C**: Primary tag assignment as fallback

## Environment Safety Guidelines

### Development Workflow

1. **Always start with pre-flight validation:**
   ```bash
   make test-preflight
   ```

2. **Use isolated environments for testing:**
   ```bash
   # For development work
   make dev-isolated

   # For comprehensive testing
   make test-isolated
   ```

3. **Never use production port (8787) for development/testing**

4. **Validate environment variables before starting:**
   - `OMNIDROP_ENV` should be `development` or `test`
   - `PORT` should be 8788-8799 range
   - `OMNIDROP_SCRIPT` should point to development AppleScript

### Installation and Updates

**Initial Installation:**
```bash
# First-time installation
make install
```

**Updating Without Plist Changes:**
```bash
# Updates binary and AppleScript, preserves existing plist configuration
make install
```

**Updating With Plist Changes:**
```bash
# Force plist update (creates timestamped backup automatically)
make install FORCE_PLIST=1
```

**Plist Update Behavior:**
- **Default (`make install`)**: Skips plist if it already exists, preserving custom settings
- **Force mode (`FORCE_PLIST=1`)**: Updates plist with automatic backup to `~/Library/LaunchAgents/com.oshiire.omnidrop.plist.backup.YYYYMMDD_HHMMSS`
- **New installation**: Always creates plist from template

**Manual Plist Backup:**
```bash
cp ~/Library/LaunchAgents/com.oshiire.omnidrop.plist ~/Library/LaunchAgents/com.oshiire.omnidrop.plist.backup
```

### Troubleshooting

**If tags are not being assigned:**
1. Verify correct AppleScript file is being used: check `OMNIDROP_SCRIPT` path
2. Check AppleScript logs for tag assignment strategy results
3. Ensure OmniFocus 4 is running during task creation

**If tests fail:**
1. Run `make test-preflight` to check environment safety
2. Verify no production processes are running on port 8787
3. Check that OmniFocus is accessible for AppleScript execution

**If `make install` reports plist is skipped:**
1. This is normal behavior to protect your custom settings
2. To update plist with new TOKEN or settings: `make install FORCE_PLIST=1`
3. Check backup files: `ls -la ~/Library/LaunchAgents/com.oshiire.omnidrop.plist.backup.*`