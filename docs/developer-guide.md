# Developer Guide

## Getting Started

### Prerequisites

Before contributing to OmniDrop, ensure you have:

1. **macOS** (required for OmniFocus and AppleScript)
2. **OmniFocus 4** or later installed
3. **Go 1.25.0** or later
4. **Python 3** (for JSON parsing in AppleScript)
5. **Git** for version control

### Development Setup

1. **Clone the Repository**
   ```bash
   git clone https://github.com/sho7650/omnidrop.git
   cd omnidrop
   ```

2. **Install Dependencies**
   ```bash
   make deps
   ```

3. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env and set TOKEN="dev-token"
   ```

4. **Run Development Server**
   ```bash
   make dev-isolated
   # Server runs on port 8788
   ```

## Project Structure

```
omnidrop/
├── cmd/omnidrop-server/
│   └── main.go                # HTTP server implementation
├── scripts/
│   ├── run-isolated-test.sh   # Test runner with isolation
│   ├── test-preflight.sh      # Environment validation
│   └── test-tag-functionality.sh # Tag feature tests
├── init/launchd/
│   └── com.oshiire.omnidrop.plist # Service configuration
├── docs/                       # Documentation
│   ├── api-reference.yaml     # OpenAPI specification
│   ├── architecture.md        # System design
│   ├── developer-guide.md     # This file
│   └── deployment-guide.md    # Production setup
├── omnidrop.applescript        # OmniFocus integration
├── Makefile                    # Build automation
├── go.mod                      # Go dependencies
└── .env.example                # Environment template
```

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes
vim main.go

# Test locally
make dev-isolated

# Run tests
make test-isolated

# Commit changes
git add .
git commit -m "feat: add new feature"
```

### 2. Testing Strategy

#### Unit Testing
```bash
# Run Go unit tests
make test

# Run with coverage
go test -cover ./...
```

#### Integration Testing
```bash
# Run isolated test suite
make test-isolated

# Run specific test scenario
./scripts/test-tag-functionality.sh
```

#### Pre-flight Validation
```bash
# Always validate before testing
make test-preflight
```

### 3. Code Standards

#### Go Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly

Example:
```go
// CreateTask creates a new task in OmniFocus via AppleScript
func CreateTask(task TaskRequest) error {
    if task.Title == "" {
        return fmt.Errorf("title is required")
    }

    // Validate and process task
    // ...

    return nil
}
```

#### AppleScript Style
- Use descriptive variable names
- Add error handling for each operation
- Log important operations
- Document complex logic

Example:
```applescript
-- Create task with specified properties
on createTaskWithProperties(taskTitle, taskNote, projectPath, tagList)
    try
        -- Task creation logic
        log "Creating task: " & taskTitle

    on error errMsg number errNum
        log "Error creating task: " & errMsg
        error errMsg number errNum
    end try
end createTaskWithProperties
```

## Adding Features

### Adding a New API Endpoint

1. **Update Router** (`main.go`)
   ```go
   http.HandleFunc("/new-endpoint", authenticateMiddleware(newEndpointHandler))
   ```

2. **Implement Handler**
   ```go
   func newEndpointHandler(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

3. **Update OpenAPI** (`docs/api-reference.yaml`)
   ```yaml
   /new-endpoint:
     post:
       summary: New endpoint description
       # ...
   ```

### Adding AppleScript Features

1. **Extend AppleScript** (`omnidrop.applescript`)
   ```applescript
   -- New feature handler
   on newFeature(parameter)
       -- Implementation
   end newFeature
   ```

2. **Update Go Integration**
   ```go
   // Add new parameter to script execution
   cmd.Args = append(cmd.Args, newParameter)
   ```

3. **Test the Feature**
   ```bash
   # Add test case to test script
   vim scripts/run-isolated-test.sh
   ```

## Debugging

### Server Debugging

1. **Enable Verbose Logging**
   ```go
   log.Printf("DEBUG: %v", variable)
   ```

2. **Check Logs**
   ```bash
   make logs-follow
   tail -f ~/.local/log/omnidrop/stderr.log
   ```

3. **Use Development Mode**
   ```bash
   TOKEN="debug-token" go run main.go
   ```

### AppleScript Debugging

1. **Add Log Statements**
   ```applescript
   log "DEBUG: Project path: " & projectPath
   ```

2. **Test in Script Editor**
   - Open `omnidrop.applescript` in Script Editor
   - Run with test parameters
   - Check event log

3. **Check Console Logs**
   ```bash
   # View AppleScript logs
   log show --predicate 'process == "osascript"' --last 1m
   ```

## Common Issues and Solutions

### Issue: "AppleScript error: Project not found"

**Solution:**
```bash
# Verify project exists in OmniFocus
# Check exact project path spelling
# Test with simple project name first
```

### Issue: "Port already in use"

**Solution:**
```bash
# Check running processes
lsof -i :8788

# Kill existing process
kill -9 <PID>

# Or use different port
PORT=8789 make dev
```

### Issue: "Tags not being created"

**Solution:**
```applescript
-- Ensure tag creation strategy is working
-- Check AppleScript logs for strategy attempts
-- Verify OmniFocus 4 is running
```

## Performance Optimization

### Current Bottlenecks
- AppleScript execution is synchronous
- JSON parsing adds overhead
- Project/tag lookups are linear

### Optimization Tips
1. **Cache Frequently Used Data**
   ```go
   var projectCache = make(map[string]bool)
   ```

2. **Batch Operations** (Future Enhancement)
   ```go
   // Queue multiple tasks for batch processing
   taskQueue <- task
   ```

3. **Profile Performance**
   ```bash
   go test -bench=. -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   ```

## Contributing Guidelines

### Pull Request Process

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/omnidrop.git
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make Changes**
   - Write clean, documented code
   - Add tests for new features
   - Update documentation

4. **Test Thoroughly**
   ```bash
   make test-preflight
   make test-isolated
   ```

5. **Submit PR**
   - Clear description of changes
   - Reference any related issues
   - Include test results

### Code Review Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass successfully
- [ ] Documentation updated
- [ ] No hardcoded values
- [ ] Error handling implemented
- [ ] Security considerations addressed
- [ ] Performance impact assessed

## Advanced Topics

### Custom Environment Setup

```bash
# Create custom environment
export OMNIDROP_ENV=custom
export PORT=8795
export OMNIDROP_SCRIPT=/path/to/custom.applescript
./omnidrop-server
```

### Integration with CI/CD

```yaml
# GitHub Actions example
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: make test
```

### Extending for Other Task Managers

While OmniDrop is OmniFocus-specific, the architecture supports adaptation:

1. **Abstract Task Interface**
   ```go
   type TaskManager interface {
       CreateTask(task TaskRequest) error
   }
   ```

2. **Implement Adapter**
   ```go
   type ThingsAdapter struct{}
   func (t *ThingsAdapter) CreateTask(task TaskRequest) error {
       // Things-specific implementation
   }
   ```

## Resources

- [OmniFocus AppleScript Reference](https://www.omnigroup.com/omnifocus/applescript)
- [Go Documentation](https://golang.org/doc/)
- [AppleScript Language Guide](https://developer.apple.com/library/archive/documentation/AppleScript/Conceptual/AppleScriptLangGuide/)
- [Project Repository](https://github.com/sho7650/omnidrop)