# Dynamic Automation Framework Design

## Overview

YAML設定駆動型の動的エンドポイント生成と自動化コマンド実行フレームワーク。

## Architecture

```
┌───────────────────────────────────────────────────────────────────┐
│                   Automation Framework                            │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────┐      ┌──────────────────┐                  │
│  │  Config Loader  │──────>│  Command Registry│                  │
│  │  (YAML Parser)  │      │  (In-Memory Map) │                  │
│  └─────────────────┘      └──────────────────┘                  │
│          │                         │                             │
│          │                         ▼                             │
│          │                ┌──────────────────┐                  │
│          │                │ Dynamic Router   │                  │
│          │                │ (Chi Subrouter)  │                  │
│          │                └──────────────────┘                  │
│          │                         │                             │
│          ▼                         ▼                             │
│  ┌─────────────────────────────────────────┐                    │
│  │       Command Executor Factory          │                    │
│  ├─────────────────────────────────────────┤                    │
│  │  ┌─────────────┐  ┌─────────────┐      │                    │
│  │  │ AppleScript │  │  Shortcuts  │      │                    │
│  │  │  Executor   │  │  Executor   │      │                    │
│  │  └─────────────┘  └─────────────┘      │                    │
│  │  ┌─────────────┐  ┌─────────────┐      │                    │
│  │  │    Shell    │  │   Future    │      │                    │
│  │  │  Executor   │  │  Executors  │      │                    │
│  │  └─────────────┘  └─────────────┘      │                    │
│  └─────────────────────────────────────────┘                    │
│          │                                                       │
│          ▼                                                       │
│  ┌─────────────────┐      ┌──────────────────┐                 │
│  │   Validator     │      │  Result Handler  │                 │
│  │  (JSON Schema)  │      │  (Formatter)     │                 │
│  └─────────────────┘      └──────────────────┘                 │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Command Definition Structure

**File**: `~/.local/share/omnidrop/commands.yaml`

```yaml
version: "1.0"

# グローバル設定
defaults:
  timeout: 30s
  working_dir: /tmp/omnidrop
  enable_logging: true

# コマンド定義
commands:
  # AppleScript実行例
  - name: open-url
    description: "Open URL in Safari browser"
    type: applescript
    endpoint: /automation/open-url
    script: |
      on run argv
        set targetURL to item 1 of argv
        tell application "Safari"
          activate
          open location targetURL
        end tell
      end run
    parameters:
      - name: url
        type: string
        required: true
        description: "URL to open"
        validation:
          pattern: "^https?://.*"
          error_message: "Invalid URL format"
    scopes:
      - automation:browser
    timeout: 10s

  # macOS Shortcuts実行例
  - name: run-shortcut
    description: "Execute macOS Shortcut with input"
    type: shortcuts
    endpoint: /automation/shortcuts/{shortcut_name}
    parameters:
      - name: shortcut_name
        type: string
        required: true
        source: path
        description: "Name of the shortcut to run"
      - name: input
        type: any
        required: false
        description: "Input data for the shortcut"
    scopes:
      - automation:shortcuts
    timeout: 60s

  # シェルコマンド実行例
  - name: video-convert
    description: "Convert video file using ffmpeg"
    type: shell
    endpoint: /automation/video-convert
    script: /usr/local/bin/convert-video.sh
    parameters:
      - name: input_path
        type: string
        required: true
        description: "Path to input video file"
        validation:
          pattern: "^/Users/.*/Videos/.*\\.(mp4|mov|avi|mkv)$"
          error_message: "Invalid video file path"
      - name: output_format
        type: string
        required: false
        default: mp4
        description: "Output video format"
        validation:
          enum: [mp4, webm, avi, mkv]
      - name: quality
        type: string
        required: false
        default: medium
        description: "Output quality preset"
        validation:
          enum: [low, medium, high, ultra]
      - name: resolution
        type: string
        required: false
        description: "Output resolution"
        validation:
          enum: ["480p", "720p", "1080p", "4k"]
    scopes:
      - automation:video-convert
    timeout: 600s
    working_dir: /tmp/omnidrop/video
    env:
      FFMPEG_PATH: /usr/local/bin/ffmpeg
      FFMPEG_THREADS: "4"

  # システム通知例
  - name: notify
    description: "Send macOS notification"
    type: applescript
    endpoint: /automation/notify
    script: |
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
    parameters:
      - name: title
        type: string
        required: true
        description: "Notification title"
        validation:
          max_length: 100
      - name: message
        type: string
        required: true
        description: "Notification message"
        validation:
          max_length: 500
      - name: sound
        type: string
        required: false
        default: ""
        description: "Notification sound"
        validation:
          enum: ["", "Basso", "Blow", "Bottle", "Frog", "Funk", "Glass", "Hero", "Morse", "Ping", "Pop", "Purr", "Sosumi", "Submarine", "Tink"]
    scopes:
      - automation:notify
    timeout: 5s

  # クリップボード操作例
  - name: set-clipboard
    description: "Set clipboard content"
    type: applescript
    endpoint: /automation/clipboard/set
    script: |
      on run argv
        set the clipboard to item 1 of argv
      end run
    parameters:
      - name: content
        type: string
        required: true
        description: "Content to set in clipboard"
        validation:
          max_length: 10000
    scopes:
      - automation:clipboard
    timeout: 5s

  # ファイル操作例（既存機能の統合）
  - name: create-file
    description: "Create file with content"
    type: builtin
    handler: files.CreateFile
    endpoint: /files
    parameters:
      - name: filename
        type: string
        required: true
        description: "File name"
      - name: content
        type: string
        required: true
        description: "File content"
      - name: directory
        type: string
        required: false
        description: "Subdirectory path"
    scopes:
      - files:write
    timeout: 30s

  # タスク作成例（既存機能の統合）
  - name: create-task
    description: "Create OmniFocus task"
    type: builtin
    handler: tasks.CreateTask
    endpoint: /tasks
    parameters:
      - name: title
        type: string
        required: true
        description: "Task title"
      - name: note
        type: string
        required: false
        description: "Task note"
      - name: project
        type: string
        required: false
        description: "Project name or path"
      - name: tags
        type: array
        items:
          type: string
        required: false
        description: "Task tags"
    scopes:
      - tasks:write
    timeout: 30s
```

### 2. Data Structures

```go
// Command represents a single automation command
type Command struct {
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Type        CommandType       `yaml:"type"`
    Endpoint    string            `yaml:"endpoint"`
    Script      string            `yaml:"script,omitempty"`
    Handler     string            `yaml:"handler,omitempty"`
    Parameters  []Parameter       `yaml:"parameters"`
    Scopes      []string          `yaml:"scopes"`
    Timeout     time.Duration     `yaml:"timeout"`
    WorkingDir  string            `yaml:"working_dir,omitempty"`
    Env         map[string]string `yaml:"env,omitempty"`
}

type CommandType string

const (
    CommandTypeAppleScript CommandType = "applescript"
    CommandTypeShortcuts   CommandType = "shortcuts"
    CommandTypeShell       CommandType = "shell"
    CommandTypeBuiltin     CommandType = "builtin"
)

type Parameter struct {
    Name        string              `yaml:"name"`
    Type        ParameterType       `yaml:"type"`
    Required    bool                `yaml:"required"`
    Default     interface{}         `yaml:"default,omitempty"`
    Source      ParameterSource     `yaml:"source,omitempty"`
    Description string              `yaml:"description,omitempty"`
    Validation  *ParameterValidation `yaml:"validation,omitempty"`
    Items       *ParameterSchema    `yaml:"items,omitempty"` // For array type
}

type ParameterType string

const (
    ParameterTypeString  ParameterType = "string"
    ParameterTypeInteger ParameterType = "integer"
    ParameterTypeFloat   ParameterType = "float"
    ParameterTypeBoolean ParameterType = "boolean"
    ParameterTypeArray   ParameterType = "array"
    ParameterTypeObject  ParameterType = "object"
    ParameterTypeAny     ParameterType = "any"
)

type ParameterSource string

const (
    ParameterSourceBody  ParameterSource = "body"  // Default: from JSON body
    ParameterSourcePath  ParameterSource = "path"  // From URL path
    ParameterSourceQuery ParameterSource = "query" // From query string
)

type ParameterValidation struct {
    Pattern      string      `yaml:"pattern,omitempty"`
    Enum         []string    `yaml:"enum,omitempty"`
    MinLength    *int        `yaml:"min_length,omitempty"`
    MaxLength    *int        `yaml:"max_length,omitempty"`
    Min          *float64    `yaml:"min,omitempty"`
    Max          *float64    `yaml:"max,omitempty"`
    ErrorMessage string      `yaml:"error_message,omitempty"`
}

type ParameterSchema struct {
    Type ParameterType `yaml:"type"`
}

type CommandConfig struct {
    Version  string             `yaml:"version"`
    Defaults CommandDefaults    `yaml:"defaults"`
    Commands []Command          `yaml:"commands"`
}

type CommandDefaults struct {
    Timeout       time.Duration `yaml:"timeout"`
    WorkingDir    string        `yaml:"working_dir"`
    EnableLogging bool          `yaml:"enable_logging"`
}
```

### 3. Config Loader

```go
package automation

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type ConfigLoader struct {
    configPath string
}

func NewConfigLoader(configPath string) *ConfigLoader {
    if configPath == "" {
        configPath = defaultConfigPath()
    }
    return &ConfigLoader{configPath: configPath}
}

func defaultConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".local/share/omnidrop/commands.yaml")
}

func (cl *ConfigLoader) Load() (*CommandConfig, error) {
    data, err := os.ReadFile(cl.configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var config CommandConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    // Apply defaults
    for i := range config.Commands {
        cmd := &config.Commands[i]
        if cmd.Timeout == 0 {
            cmd.Timeout = config.Defaults.Timeout
        }
        if cmd.WorkingDir == "" {
            cmd.WorkingDir = config.Defaults.WorkingDir
        }
    }

    if err := cl.validate(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return &config, nil
}

func (cl *ConfigLoader) validate(config *CommandConfig) error {
    // Check for duplicate names
    names := make(map[string]bool)
    for _, cmd := range config.Commands {
        if names[cmd.Name] {
            return fmt.Errorf("duplicate command name: %s", cmd.Name)
        }
        names[cmd.Name] = true
    }

    // Check for duplicate endpoints
    endpoints := make(map[string]bool)
    for _, cmd := range config.Commands {
        if endpoints[cmd.Endpoint] {
            return fmt.Errorf("duplicate endpoint: %s", cmd.Endpoint)
        }
        endpoints[cmd.Endpoint] = true
    }

    // Validate each command
    for _, cmd := range config.Commands {
        if err := cl.validateCommand(&cmd); err != nil {
            return fmt.Errorf("command %s: %w", cmd.Name, err)
        }
    }

    return nil
}

func (cl *ConfigLoader) validateCommand(cmd *Command) error {
    if cmd.Name == "" {
        return fmt.Errorf("command name is required")
    }
    if cmd.Type == "" {
        return fmt.Errorf("command type is required")
    }
    if cmd.Endpoint == "" {
        return fmt.Errorf("endpoint is required")
    }

    // Type-specific validation
    switch cmd.Type {
    case CommandTypeAppleScript:
        if cmd.Script == "" {
            return fmt.Errorf("script is required for applescript type")
        }
    case CommandTypeShell:
        if cmd.Script == "" {
            return fmt.Errorf("script path is required for shell type")
        }
        // Check if script file exists
        if _, err := os.Stat(cmd.Script); err != nil {
            return fmt.Errorf("script file not found: %s", cmd.Script)
        }
    case CommandTypeBuiltin:
        if cmd.Handler == "" {
            return fmt.Errorf("handler is required for builtin type")
        }
    }

    // Validate parameters
    for _, param := range cmd.Parameters {
        if err := cl.validateParameter(&param); err != nil {
            return fmt.Errorf("parameter %s: %w", param.Name, err)
        }
    }

    return nil
}

func (cl *ConfigLoader) validateParameter(param *Parameter) error {
    if param.Name == "" {
        return fmt.Errorf("parameter name is required")
    }
    if param.Type == "" {
        return fmt.Errorf("parameter type is required")
    }

    // Validate enum values
    if param.Validation != nil && len(param.Validation.Enum) > 0 {
        if param.Default != nil {
            defaultStr, ok := param.Default.(string)
            if !ok {
                return fmt.Errorf("default value must be string for enum validation")
            }
            found := false
            for _, v := range param.Validation.Enum {
                if v == defaultStr {
                    found = true
                    break
                }
            }
            if !found {
                return fmt.Errorf("default value not in enum: %v", param.Default)
            }
        }
    }

    return nil
}
```

### 4. Command Registry

```go
package automation

import (
    "fmt"
    "sync"
)

type Registry struct {
    mu       sync.RWMutex
    commands map[string]*Command
}

func NewRegistry() *Registry {
    return &Registry{
        commands: make(map[string]*Command),
    }
}

func (r *Registry) Register(cmd *Command) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.commands[cmd.Name]; exists {
        return fmt.Errorf("command already registered: %s", cmd.Name)
    }

    r.commands[cmd.Name] = cmd
    return nil
}

func (r *Registry) Get(name string) (*Command, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    cmd, ok := r.commands[name]
    return cmd, ok
}

func (r *Registry) GetByEndpoint(endpoint string) (*Command, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, cmd := range r.commands {
        if cmd.Endpoint == endpoint {
            return cmd, true
        }
    }
    return nil, false
}

func (r *Registry) List() []*Command {
    r.mu.RLock()
    defer r.mu.RUnlock()

    commands := make([]*Command, 0, len(r.commands))
    for _, cmd := range r.commands {
        commands = append(commands, cmd)
    }
    return commands
}

func (r *Registry) Reload(config *CommandConfig) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // Clear existing commands
    r.commands = make(map[string]*Command)

    // Register new commands
    for i := range config.Commands {
        cmd := &config.Commands[i]
        r.commands[cmd.Name] = cmd
    }

    return nil
}
```

### 5. Dynamic Router

```go
package automation

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

type Router struct {
    registry *Registry
    executor *Executor
}

func NewRouter(registry *Registry, executor *Executor) *Router {
    return &Router{
        registry: registry,
        executor: executor,
    }
}

func (r *Router) Setup(router chi.Router) {
    commands := r.registry.List()
    
    for _, cmd := range commands {
        // Create handler for this command
        handler := r.createHandler(cmd)
        
        // Parse endpoint for path parameters
        // Example: /automation/shortcuts/{shortcut_name}
        endpoint := cmd.Endpoint
        
        // Register route with scope middleware
        router.With(RequireScopes(cmd.Scopes...)).Post(endpoint, handler)
    }
}

func (r *Router) createHandler(cmd *Command) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        // Extract parameters
        params, err := r.extractParameters(req, cmd)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Execute command
        result, err := r.executor.Execute(req.Context(), cmd, params)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Return result
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }
}

func (r *Router) extractParameters(req *http.Request, cmd *Command) (map[string]interface{}, error) {
    params := make(map[string]interface{})

    // Parse request body
    var body map[string]interface{}
    if req.Body != nil {
        if err := json.NewDecoder(req.Body).Decode(&body); err != nil && err != io.EOF {
            return nil, fmt.Errorf("invalid JSON body: %w", err)
        }
    }

    // Extract each parameter
    for _, paramDef := range cmd.Parameters {
        var value interface{}
        var found bool

        switch paramDef.Source {
        case ParameterSourcePath, "": // Default to path for path params
            if strings.Contains(cmd.Endpoint, "{"+paramDef.Name+"}") {
                value = chi.URLParam(req, paramDef.Name)
                found = value != ""
            }
        case ParameterSourceQuery:
            value = req.URL.Query().Get(paramDef.Name)
            found = value != ""
        case ParameterSourceBody, "": // Default to body
            value, found = body[paramDef.Name]
        }

        // Check required
        if !found {
            if paramDef.Required {
                return nil, fmt.Errorf("required parameter missing: %s", paramDef.Name)
            }
            if paramDef.Default != nil {
                value = paramDef.Default
            } else {
                continue
            }
        }

        // Validate parameter
        if err := r.validateParameter(value, &paramDef); err != nil {
            return nil, fmt.Errorf("parameter %s: %w", paramDef.Name, err)
        }

        params[paramDef.Name] = value
    }

    return params, nil
}

func (r *Router) validateParameter(value interface{}, param *Parameter) error {
    if param.Validation == nil {
        return nil
    }

    val := param.Validation

    // Type-specific validation
    switch param.Type {
    case ParameterTypeString:
        str, ok := value.(string)
        if !ok {
            return fmt.Errorf("expected string, got %T", value)
        }

        // Pattern validation
        if val.Pattern != "" {
            matched, err := regexp.MatchString(val.Pattern, str)
            if err != nil {
                return fmt.Errorf("invalid pattern: %w", err)
            }
            if !matched {
                if val.ErrorMessage != "" {
                    return fmt.Errorf(val.ErrorMessage)
                }
                return fmt.Errorf("does not match pattern: %s", val.Pattern)
            }
        }

        // Enum validation
        if len(val.Enum) > 0 {
            found := false
            for _, enumVal := range val.Enum {
                if str == enumVal {
                    found = true
                    break
                }
            }
            if !found {
                return fmt.Errorf("must be one of: %v", val.Enum)
            }
        }

        // Length validation
        if val.MinLength != nil && len(str) < *val.MinLength {
            return fmt.Errorf("minimum length is %d", *val.MinLength)
        }
        if val.MaxLength != nil && len(str) > *val.MaxLength {
            return fmt.Errorf("maximum length is %d", *val.MaxLength)
        }

    case ParameterTypeInteger:
        // Handle both float64 (from JSON) and int
        var num float64
        switch v := value.(type) {
        case float64:
            num = v
        case int:
            num = float64(v)
        default:
            return fmt.Errorf("expected integer, got %T", value)
        }

        if val.Min != nil && num < *val.Min {
            return fmt.Errorf("minimum value is %f", *val.Min)
        }
        if val.Max != nil && num > *val.Max {
            return fmt.Errorf("maximum value is %f", *val.Max)
        }
    }

    return nil
}
```

### 6. Command Executors

**Base Executor Interface:**
```go
package automation

import "context"

type Executor interface {
    Execute(ctx context.Context, cmd *Command, params map[string]interface{}) (*ExecutionResult, error)
}

type ExecutionResult struct {
    Success  bool        `json:"success"`
    Output   string      `json:"output,omitempty"`
    Error    string      `json:"error,omitempty"`
    ExitCode int         `json:"exit_code,omitempty"`
    Duration string      `json:"duration"`
    Data     interface{} `json:"data,omitempty"`
}
```

**AppleScript Executor:**
```go
package automation

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

type AppleScriptExecutor struct{}

func (e *AppleScriptExecutor) Execute(ctx context.Context, cmd *Command, params map[string]interface{}) (*ExecutionResult, error) {
    start := time.Now()

    // Build AppleScript arguments
    args := make([]string, 0, len(cmd.Parameters))
    for _, paramDef := range cmd.Parameters {
        if val, ok := params[paramDef.Name]; ok {
            args = append(args, fmt.Sprintf("%v", val))
        }
    }

    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
    defer cancel()

    // Execute AppleScript
    osascriptCmd := exec.CommandContext(execCtx, "osascript", "-e", cmd.Script)
    if len(args) > 0 {
        osascriptCmd.Args = append(osascriptCmd.Args, args...)
    }

    output, err := osascriptCmd.CombinedOutput()
    duration := time.Since(start)

    result := &ExecutionResult{
        Duration: duration.String(),
    }

    if err != nil {
        result.Success = false
        result.Error = err.Error()
        result.Output = string(output)
        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        }
        return result, nil
    }

    result.Success = true
    result.Output = strings.TrimSpace(string(output))
    result.ExitCode = 0
    return result, nil
}
```

**Shortcuts Executor:**
```go
package automation

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

type ShortcutsExecutor struct{}

func (e *ShortcutsExecutor) Execute(ctx context.Context, cmd *Command, params map[string]interface{}) (*ExecutionResult, error) {
    start := time.Now()

    // Extract shortcut name
    shortcutName, ok := params["shortcut_name"].(string)
    if !ok {
        return nil, fmt.Errorf("shortcut_name is required")
    }

    // Build shortcuts command
    args := []string{"run", shortcutName}

    // Add input if provided
    if input, ok := params["input"]; ok {
        inputJSON, err := json.Marshal(input)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal input: %w", err)
        }
        args = append(args, "-i", string(inputJSON))
    }

    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
    defer cancel()

    // Execute shortcuts command
    shortcutsCmd := exec.CommandContext(execCtx, "shortcuts", args...)
    output, err := shortcutsCmd.CombinedOutput()
    duration := time.Since(start)

    result := &ExecutionResult{
        Duration: duration.String(),
    }

    if err != nil {
        result.Success = false
        result.Error = err.Error()
        result.Output = string(output)
        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        }
        return result, nil
    }

    result.Success = true
    result.Output = strings.TrimSpace(string(output))
    result.ExitCode = 0

    // Try to parse output as JSON
    var data interface{}
    if err := json.Unmarshal(output, &data); err == nil {
        result.Data = data
    }

    return result, nil
}
```

**Shell Executor:**
```go
package automation

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

type ShellExecutor struct{}

func (e *ShellExecutor) Execute(ctx context.Context, cmd *Command, params map[string]interface{}) (*ExecutionResult, error) {
    start := time.Now()

    // Build command arguments
    args := e.buildArgs(cmd, params)

    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
    defer cancel()

    // Execute shell script
    shellCmd := exec.CommandContext(execCtx, cmd.Script, args...)

    // Set working directory
    if cmd.WorkingDir != "" {
        shellCmd.Dir = cmd.WorkingDir
    }

    // Set environment variables
    if len(cmd.Env) > 0 {
        shellCmd.Env = append(os.Environ(), e.buildEnv(cmd.Env)...)
    }

    output, err := shellCmd.CombinedOutput()
    duration := time.Since(start)

    result := &ExecutionResult{
        Duration: duration.String(),
    }

    if err != nil {
        result.Success = false
        result.Error = err.Error()
        result.Output = string(output)
        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        }
        return result, nil
    }

    result.Success = true
    result.Output = strings.TrimSpace(string(output))
    result.ExitCode = 0
    return result, nil
}

func (e *ShellExecutor) buildArgs(cmd *Command, params map[string]interface{}) []string {
    args := make([]string, 0)
    
    for _, paramDef := range cmd.Parameters {
        if val, ok := params[paramDef.Name]; ok {
            // Convert to command-line argument
            argName := "--" + strings.ReplaceAll(paramDef.Name, "_", "-")
            args = append(args, argName, fmt.Sprintf("%v", val))
        }
    }
    
    return args
}

func (e *ShellExecutor) buildEnv(envMap map[string]string) []string {
    env := make([]string, 0, len(envMap))
    for k, v := range envMap {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }
    return env
}
```

## Security Considerations

### 1. Script Execution Safety
- **ホワイトリスト方式**: 設定ファイルに定義されたコマンドのみ実行可能
- **パス検証**: シェルスクリプトは絶対パスで指定、実行前に存在確認
- **サンドボックス化**: `working_dir`で実行ディレクトリを制限

### 2. Parameter Validation
- 正規表現によるパターンマッチング
- Enum値の厳密な検証
- 長さ・範囲制限

### 3. Timeout Protection
- すべてのコマンドに timeout 設定必須
- デフォルト 30秒、最大推奨 10分

### 4. Resource Limits
将来的な実装候補:
- メモリ使用量制限
- CPU使用率制限
- 同時実行数制限

## Metrics

```go
var (
    CommandExecutionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "omnidrop_command_executions_total",
            Help: "Total number of command executions",
        },
        []string{"command_name", "status"},
    )

    CommandExecutionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "omnidrop_command_execution_duration_seconds",
            Help:    "Command execution duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"command_name", "type"},
    )

    CommandTimeouts = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "omnidrop_command_timeouts_total",
            Help: "Total number of command timeouts",
        },
        []string{"command_name"},
    )
)
```

## Testing Strategy

### Unit Tests
```go
func TestParameterExtraction(t *testing.T) {
    // Test parameter extraction from various sources
}

func TestParameterValidation(t *testing.T) {
    // Test validation rules
}

func TestCommandRegistry(t *testing.T) {
    // Test registry operations
}
```

### Integration Tests
```go
func TestAppleScriptExecution(t *testing.T) {
    // Test actual AppleScript execution
}

func TestShortcutsExecution(t *testing.T) {
    // Test Shortcuts execution
}
```

## Example Requests

**Open URL:**
```bash
curl -X POST http://localhost:8787/automation/open-url \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

**Run Shortcut:**
```bash
curl -X POST http://localhost:8787/automation/shortcuts/MyWorkflow \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"input": "some data"}'
```

**Send Notification:**
```bash
curl -X POST http://localhost:8787/automation/notify \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task Complete",
    "message": "Your task has been completed successfully",
    "sound": "Glass"
  }'
```

## Configuration Reload

動的な設定リロード機能:

```go
// HTTP endpoint
POST /admin/reload-config
Authorization: Bearer ${ADMIN_TOKEN}

// CLI
omnidrop-cli config reload
```

実装:
```go
func (s *Server) ReloadConfig(w http.ResponseWriter, r *http.Request) {
    // Require admin scope
    claims := r.Context().Value("oauth_claims").(*Claims)
    if !hasRequiredScopes(claims.Scopes, []string{"admin"}) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Reload configuration
    config, err := s.configLoader.Load()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update registry
    if err := s.registry.Reload(config); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Re-setup router
    s.setupDynamicRoutes()

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "ok",
        "message": "Configuration reloaded successfully",
    })
}
```
