# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OmniDrop is a lightweight REST API server that bridges external applications with OmniFocus on macOS. It receives HTTP POST requests with task data and creates corresponding tasks in OmniFocus using AppleScript automation.

## Architecture

The system consists of two main components:

1. **HTTP Server** (`main.go`): Go-based REST API that:
   - Listens on a configurable port (default: 8787)
   - Requires Bearer token authentication
   - Accepts POST requests to `/tasks` endpoint
   - Validates JSON payloads and invokes AppleScript

2. **AppleScript Bridge** (`omnidrop.applescript`): Handles OmniFocus integration:
   - Parses JSON task data using Python
   - Creates tasks in OmniFocus with specified properties
   - Sets due dates to end of current day (23:59:59)
   - Supports project assignment and tag management

## Development Commands

```bash
# Build the server
go build -o omnidrop-server .

# Run the server (requires TOKEN environment variable)
TOKEN="your-secret-token" ./omnidrop-server

# Run with custom port
PORT=8080 TOKEN="your-secret-token" ./omnidrop-server

# Test the API endpoint
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task","note":"Description","project":"Work","tags":["urgent"]}'
```

## Configuration

The server requires environment variables set in `.env` file:
- `TOKEN` (required): Bearer token for API authentication
- `PORT` (optional): Server port, defaults to 8787

## API Contract

**POST /tasks**

Request body:
```json
{
  "title": "string (required)",
  "note": "string (optional)",
  "project": "string (optional)",
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

## Key Implementation Details

- The server only accepts POST requests and returns 405 for other methods
- Authentication is mandatory via `Authorization: Bearer <token>` header
- AppleScript execution errors are captured and returned in API responses
- Tasks are automatically assigned today's date at 23:59:59 as due date
- Project and tag names must exactly match existing items in OmniFocus
- Tags that don't exist in OmniFocus are silently ignored