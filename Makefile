# Makefile for OmniDrop Server
.PHONY: all build test install uninstall clean help start stop status logs logs-follow deps

# Variables
BINARY_NAME=omnidrop-server
BUILD_DIR=./build
BIN_DIR=$(BUILD_DIR)/bin
PACKAGE_DIR=$(BUILD_DIR)/package
CMD_DIR=./cmd/$(BINARY_NAME)
INSTALL_DIR=$(HOME)/bin
CONFIG_DIR=$(HOME)/.config/omnidrop
LOG_DIR=$(HOME)/.local/log/omnidrop
SCRIPT_DIR=$(HOME)/.local/share/omnidrop
WORK_DIR=$(HOME)/.local/var/omnidrop
LAUNCHD_PLIST=com.oshiire.omnidrop.plist
LAUNCHD_DIR=$(HOME)/Library/LaunchAgents
PLIST_TEMPLATE=./init/launchd/$(LAUNCHD_PLIST)
APPLESCRIPT_FILE=omnidrop.applescript

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build information
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## all: Build the application
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## run: Run the server in development mode
run: build
	@echo "Starting development server..."
	@echo "Press Ctrl+C to stop"
	./$(BIN_DIR)/$(BINARY_NAME)

## dev: Run the server with auto-reload (requires TOKEN env var)
dev:
	@echo "Starting development server (direct run)..."
	@if [ -z "$$TOKEN" ]; then \
		echo "Error: TOKEN environment variable required for development"; \
		echo "Usage: TOKEN=your-token make dev"; \
		exit 1; \
	fi
	$(GOCMD) run $(CMD_DIR)/main.go

## install: Install the application and LaunchAgent
install: build
	@echo "Installing $(BINARY_NAME)..."
	@# Create directories
	mkdir -p $(INSTALL_DIR) $(CONFIG_DIR) $(LOG_DIR) $(SCRIPT_DIR) $(WORK_DIR) $(LAUNCHD_DIR)

	@# Install binary
	cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

	@# Install AppleScript
	@echo "Installing AppleScript..."
	cp $(APPLESCRIPT_FILE) $(SCRIPT_DIR)/
	chmod 644 $(SCRIPT_DIR)/$(APPLESCRIPT_FILE)

	@# Install LaunchAgent plist with TOKEN replacement
	@echo "Installing LaunchAgent..."
	@if [ -f .env ]; then \
		TOKEN_VALUE=$$(grep '^TOKEN=' .env | cut -d'=' -f2- | sed 's/^"\(.*\)"$$/\1/'); \
		if [ -n "$$TOKEN_VALUE" ]; then \
			echo "Using TOKEN from .env file"; \
			sed -e 's/{{TOKEN}}/'"$$TOKEN_VALUE"'/g' -e 's|{{HOME}}|$(HOME)|g' $(PLIST_TEMPLATE) > /tmp/omnidrop-plist.$$$$; \
			cp /tmp/omnidrop-plist.$$$$ $(LAUNCHD_DIR)/$(LAUNCHD_PLIST); \
			rm -f /tmp/omnidrop-plist.$$$$; \
		else \
			echo "Warning: TOKEN not found in .env file, using template as-is"; \
			sed 's|{{HOME}}|$(HOME)|g' $(PLIST_TEMPLATE) > /tmp/omnidrop-plist.$$$$; \
			cp /tmp/omnidrop-plist.$$$$ $(LAUNCHD_DIR)/$(LAUNCHD_PLIST); \
			rm -f /tmp/omnidrop-plist.$$$$; \
		fi; \
	else \
		echo "Warning: .env file not found, using template as-is"; \
		sed 's|{{HOME}}|$(HOME)|g' $(PLIST_TEMPLATE) > /tmp/omnidrop-plist.$$$$; \
		cp /tmp/omnidrop-plist.$$$$ $(LAUNCHD_DIR)/$(LAUNCHD_PLIST); \
		rm -f /tmp/omnidrop-plist.$$$$; \
	fi
	chmod 644 $(LAUNCHD_DIR)/$(LAUNCHD_PLIST)

	@# Load LaunchAgent
	launchctl load $(LAUNCHD_DIR)/$(LAUNCHD_PLIST) 2>/dev/null || true

	@echo "Installation completed!"
	@echo "Binary: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Script: $(SCRIPT_DIR)/$(APPLESCRIPT_FILE)"
	@echo "Service will start automatically on login."
	@echo "Use 'make start' to start the service now."

## uninstall: Remove the application and LaunchAgent
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@# Stop and unload LaunchAgent
	launchctl unload $(LAUNCHD_DIR)/$(LAUNCHD_PLIST) 2>/dev/null || true
	rm -f $(LAUNCHD_DIR)/$(LAUNCHD_PLIST)

	@# Remove binary, script and directories
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -rf $(CONFIG_DIR) $(LOG_DIR) $(SCRIPT_DIR) $(WORK_DIR)

	@echo "Uninstallation completed!"

## start: Start the LaunchAgent service
start:
	@echo "Starting $(BINARY_NAME) service..."
	launchctl load $(LAUNCHD_DIR)/$(LAUNCHD_PLIST) 2>/dev/null || \
	launchctl start com.oshiire.omnidrop

## stop: Stop the LaunchAgent service
stop:
	@echo "Stopping $(BINARY_NAME) service..."
	launchctl stop com.oshiire.omnidrop 2>/dev/null || true

## status: Check service status
status:
	@echo "Service status:"
	@launchctl list | grep com.oshiire.omnidrop || echo "Service not running"

## logs: Show service logs
logs:
	@echo "=== STDOUT Log ==="
	@tail -20 $(LOG_DIR)/stdout.log 2>/dev/null || echo "No stdout log found"
	@echo ""
	@echo "=== STDERR Log ==="
	@tail -20 $(LOG_DIR)/stderr.log 2>/dev/null || echo "No stderr log found"

## logs-follow: Follow service logs in real-time
logs-follow:
	@echo "Following logs (Ctrl+C to stop)..."
	@tail -f $(LOG_DIR)/stdout.log $(LOG_DIR)/stderr.log 2>/dev/null || \
	echo "Log files not found. Service may not be running."

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

## deps: Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

## test-isolated: Run isolated test suite with complete environment separation
test-isolated: build
	@echo "🧪 Running isolated test suite (port 8789)..."
	@./scripts/run-isolated-test.sh

## test-preflight: Run pre-flight validation checks
test-preflight:
	@echo "🔍 Running pre-flight validation..."
	@./scripts/test-preflight.sh

## dev-isolated: Run development server with explicit environment control (port 8788)
dev-isolated: build
	@echo "🚀 Starting isolated development server (port 8788)..."
	@echo "Environment: development"
	@echo "Port: 8788"
	@echo "Script: ./omnidrop.applescript"
	@echo ""
	@echo "Press Ctrl+C to stop"
	@OMNIDROP_ENV=development \
	 OMNIDROP_SCRIPT=./omnidrop.applescript \
	 PORT=8788 \
	 TOKEN=$${TOKEN:-dev-token} \
	 $(BIN_DIR)/$(BINARY_NAME)

## staging: Run staging environment (production-like, port 8790)
staging: build
	@echo "🎭 Starting staging environment (port 8790)..."
	@if [ -z "$$TOKEN" ]; then \
		echo "Error: TOKEN environment variable required for staging"; \
		echo "Usage: TOKEN=your-token make staging"; \
		exit 1; \
	fi
	@echo "Environment: staging (test mode)"
	@echo "Port: 8790"
	@echo "Script: ./omnidrop.applescript"
	@echo ""
	@echo "Press Ctrl+C to stop"
	@OMNIDROP_ENV=test \
	 OMNIDROP_SCRIPT=./omnidrop.applescript \
	 PORT=8790 \
	 TOKEN=$$TOKEN \
	 $(BIN_DIR)/$(BINARY_NAME)

## production-run: Run production server (PROTECTED - port 8787)
production-run:
	@echo "🚨 PRODUCTION ENVIRONMENT - Port 8787"
	@echo "======================================"
	@echo ""
	@echo "⚠️  This will start the production server"
	@echo "⚠️  Port: 8787"
	@echo "⚠️  Script: $(SCRIPT_DIR)/$(APPLESCRIPT_FILE)"
	@echo ""
	@read -p "Are you ABSOLUTELY SURE? Type 'yes' to continue: " confirm; \
	if [ "$$confirm" != "yes" ]; then \
		echo "Production run cancelled"; \
		exit 1; \
	fi
	@if [ -z "$$TOKEN" ]; then \
		echo "Error: TOKEN environment variable required for production"; \
		exit 1; \
	fi
	@OMNIDROP_ENV=production \
	 PORT=8787 \
	 TOKEN=$$TOKEN \
	 $(BIN_DIR)/$(BINARY_NAME)

## help: Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@grep -E '^## (build|run|dev|test):' $(MAKEFILE_LIST) | sed 's/^## /  /'
	@echo ""
	@echo "Testing & Environments:"
	@grep -E '^## (test-isolated|test-preflight|dev-isolated|staging|production-run):' $(MAKEFILE_LIST) | sed 's/^## /  /'
	@echo ""
	@echo "Build & Install:"
	@grep -E '^## (all|install|uninstall|clean|deps):' $(MAKEFILE_LIST) | sed 's/^## /  /'
	@echo ""
	@echo "Service Management:"
	@grep -E '^## (start|stop|status|logs):' $(MAKEFILE_LIST) | sed 's/^## /  /'