# Architecture Documentation

## System Overview

OmniDrop follows a simple yet robust architecture pattern that bridges HTTP REST APIs with native macOS automation through AppleScript.

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐      ┌────────────┐
│   Client    │─────▶│  HTTP Server │─────▶│ AppleScript │─────▶│ OmniFocus  │
│ Application │ HTTP │   (Go)       │ exec │   Bridge    │ OSA  │    App     │
└─────────────┘      └──────────────┘      └─────────────┘      └────────────┘
```

## Component Architecture

### 1. HTTP Server Layer (`cmd/omnidrop-server/main.go`)

**Responsibilities:**
- HTTP request handling and routing
- Bearer token authentication
- Request validation and JSON parsing
- AppleScript invocation
- Environment-aware configuration
- Graceful shutdown handling

**Key Features:**
- **Single Endpoint Design**: Focused on `/tasks` POST endpoint
- **Middleware Chain**: Authentication → Validation → Execution
- **Environment Isolation**: Separate configs for dev/test/prod
- **Signal Handling**: Graceful shutdown on SIGTERM/SIGINT

### 2. AppleScript Bridge Layer (`omnidrop.applescript`)

**Responsibilities:**
- OmniFocus 4 API interaction
- Task creation with properties
- Project hierarchy navigation
- Tag management and creation
- Error handling and recovery

**Key Features:**
- **JSON Parsing**: Python integration for reliable parsing
- **Multi-Strategy Tag Assignment**: Three fallback strategies
- **Hierarchical Project Support**: Folder navigation capability
- **Automatic Tag Creation**: Creates missing tags on-the-fly

### 3. Service Management Layer

**LaunchAgent (`init/launchd/com.oshiire.omnidrop.plist`):**
- Automatic startup on login
- Environment variable management
- Log rotation and management
- Process lifecycle control

**Makefile Build System:**
- Compilation and installation
- Service lifecycle management
- Environment configuration
- Testing infrastructure

## Data Flow

### Task Creation Flow

```
1. Client Request
   └─▶ POST /tasks with JSON payload

2. Authentication
   └─▶ Validate Bearer token

3. Validation
   └─▶ Parse JSON, validate required fields

4. Script Resolution
   └─▶ Determine AppleScript path based on environment

5. AppleScript Execution
   ├─▶ Parse JSON using Python
   ├─▶ Find/Create project
   ├─▶ Create task with properties
   ├─▶ Assign/Create tags
   └─▶ Set due date (23:59:59 today)

6. Response
   └─▶ Return success/error JSON
```

## Environment Architecture

### Production Environment
- **Port**: 8787 (protected)
- **Script**: `~/.local/share/omnidrop/omnidrop.applescript`
- **Logs**: `~/.local/log/omnidrop/`
- **Config**: `~/.config/omnidrop/`
- **Service**: LaunchAgent with auto-start

### Development Environment
- **Port**: 8788
- **Script**: `./omnidrop.applescript` (local)
- **Logs**: Console output
- **Config**: `.env` file
- **Service**: Direct execution

### Test Environment
- **Port Range**: 8788-8799
- **Script**: Temporary isolated script
- **Logs**: Temporary directory
- **Config**: Environment variables
- **Service**: Ephemeral processes

## Security Architecture

### Authentication Layer
- **Bearer Token**: Required for all task operations
- **Token Storage**: Environment variable or `.env` file
- **Token Rotation**: Supported through environment updates

### Network Security
- **Localhost Only**: Default binding to 127.0.0.1
- **Port Protection**: Production port reserved
- **HTTPS Support**: Via reverse proxy (nginx/caddy)

### Process Security
- **User-Level Execution**: No root privileges required
- **File Permissions**: Restricted access to configs
- **Log Security**: User-only readable logs

## Error Handling Strategy

### Server-Level Errors
1. **Authentication Failures**: 401 with clear message
2. **Validation Errors**: 400 with field details
3. **Script Errors**: 500 with AppleScript output

### AppleScript-Level Errors
1. **Project Not Found**: Graceful fallback to inbox
2. **Tag Creation Failure**: Continue with remaining tags
3. **OmniFocus Not Running**: Clear error message

### Recovery Mechanisms
- **Automatic Retry**: For transient failures
- **Fallback Strategies**: Multiple tag assignment methods
- **Graceful Degradation**: Partial success handling

## Scalability Considerations

### Current Limitations
- Single-threaded AppleScript execution
- Local-only operation (macOS requirement)
- Sequential task processing

### Optimization Opportunities
- Request queuing for batch processing
- Caching for project/tag lookups
- Connection pooling for AppleScript calls

## Integration Points

### Client Integration
- Standard REST API with JSON
- Bearer token authentication
- Language-agnostic interface

### OmniFocus Integration
- AppleScript automation API
- Direct property manipulation
- Native tag and project support

### System Integration
- LaunchAgent for service management
- Signal handling for lifecycle
- Standard logging to filesystem

## Monitoring and Observability

### Health Checks
- `/health` endpoint for status
- Version and environment info
- Uptime tracking

### Logging
- Structured logging to files
- Separate stdout/stderr streams
- Rotation via LaunchAgent

### Metrics (Future Enhancement)
- Request count and latency
- Success/failure rates
- AppleScript execution time