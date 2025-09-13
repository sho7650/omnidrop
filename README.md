# OmniDrop

A lightweight REST API server that bridges external applications with OmniFocus on macOS. Create OmniFocus tasks programmatically through a simple HTTP API.

## Features

- 🚀 Simple REST API for task creation
- 🔐 Bearer token authentication for security
- 📝 Support for task notes, projects, and tags
- ⏰ Automatic due date setting (end of current day)
- 🍎 Native OmniFocus integration via AppleScript
- 🔧 Environment-based configuration with `.env` support

## Prerequisites

- macOS (required for OmniFocus and AppleScript)
- OmniFocus 3 or later installed
- Go 1.25.0 or later
- Python 3 (for JSON parsing in AppleScript)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/omnidrop.git
cd omnidrop
```

2. Install dependencies:
```bash
go get github.com/joho/godotenv
```

3. Build the server:
```bash
go build -o omnidrop-server .
```

4. Configure environment variables:
```bash
cp .env.example .env
# Edit .env and set your secure token
```

## Configuration

Create a `.env` file in the project root with the following variables:

```env
# Required: Authentication token for API access
TOKEN=your-secret-token-here

# Optional: Server port (default: 8787)
PORT=8787

# Optional: Path to AppleScript (default: ./omnidrop.applescript)
SCRIPT=./omnidrop.applescript
```

## Usage

### Starting the Server

```bash
# Using .env file (recommended)
./omnidrop-server

# Or with environment variables
TOKEN="your-secret-token" ./omnidrop-server

# Custom port
PORT=8080 TOKEN="your-secret-token" ./omnidrop-server
```

The server will start on `http://localhost:8787` (or your configured port).

### API Reference

#### Create Task

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

## How It Works

1. **HTTP Server** (`main.go`): 
   - Receives POST requests with task data
   - Validates authentication token
   - Passes task data to AppleScript

2. **AppleScript Bridge** (`omnidrop.applescript`):
   - Parses JSON data using Python
   - Creates tasks in OmniFocus using AppleScript automation
   - Sets due date to 23:59:59 of current day
   - Handles project assignment and tag management

## Important Notes

- **Project Names**: Must exactly match existing projects in OmniFocus
- **Tags**: Non-existent tags are silently ignored
- **Due Dates**: All tasks are automatically set to due at end of current day (23:59:59)
- **Inbox**: Tasks without a project are created in the OmniFocus inbox
- **Security**: Always use HTTPS in production and keep your token secure

## Error Handling

The API provides detailed error messages for common issues:

- `401 Unauthorized`: Invalid or missing authentication token
- `400 Bad Request`: Invalid JSON or missing required fields
- `405 Method Not Allowed`: Only POST requests are accepted
- `500 Internal Server Error`: AppleScript execution failed (check OmniFocus is running)

## Development

### Running Tests
```bash
go test ./...
```

### Building for Distribution
```bash
# Build for current platform
go build -o omnidrop-server .

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o omnidrop-server-amd64 .

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o omnidrop-server-arm64 .
```

### Project Structure
```
omnidrop/
├── main.go                 # HTTP server implementation
├── omnidrop.applescript    # OmniFocus integration script
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── .env.example            # Environment configuration template
├── .env                    # Local environment configuration (git-ignored)
├── CLAUDE.md               # AI assistant guidance
└── README.md               # This file
```

## Troubleshooting

### "AppleScript error" Response
- Ensure OmniFocus is running
- Check that project/tag names match exactly with OmniFocus
- Verify Python 3 is installed: `python3 --version`

### "Unauthorized" Response
- Verify the token in your `.env` file matches the request
- Ensure the Authorization header format is: `Bearer <token>`

### Server Won't Start
- Check if port is already in use: `lsof -i :8787`
- Verify `.env` file exists and TOKEN is set
- Check file permissions on AppleScript file

## Security Considerations

- **Token Storage**: Never commit `.env` files with real tokens
- **HTTPS**: Use a reverse proxy (nginx/caddy) for HTTPS in production
- **Token Rotation**: Regularly rotate your authentication tokens
- **Network Access**: Bind to localhost only unless external access is required

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Specify your license here]

## Acknowledgments

Built with:
- [Go](https://golang.org/) - The programming language
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [OmniFocus](https://www.omnigroup.com/omnifocus) - Task management application