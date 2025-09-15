#!/bin/bash

# OmniDrop Test Pre-flight Validation Script
# Ensures safe test execution with complete production environment protection

set -e  # Exit on any error

echo "🔍 OmniDrop Test Pre-flight Validation"
echo "======================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERRORS=0

# Function to report error
report_error() {
    echo -e "${RED}❌ FATAL: $1${NC}"
    ERRORS=$((ERRORS + 1))
}

# Function to report warning
report_warning() {
    echo -e "${YELLOW}⚠️ WARNING: $1${NC}"
}

# Function to report success
report_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to report info
report_info() {
    echo -e "${BLUE}ℹ️ $1${NC}"
}

echo
echo "1. Environment Variable Validation"
echo "----------------------------------"

# Check OMNIDROP_ENV
if [[ -z "$OMNIDROP_ENV" ]]; then
    report_warning "OMNIDROP_ENV not set, will use legacy behavior"
else
    report_info "OMNIDROP_ENV: $OMNIDROP_ENV"

    # Validate environment value
    case "$OMNIDROP_ENV" in
        production|development|test)
            report_success "Valid environment: $OMNIDROP_ENV"
            ;;
        *)
            report_error "Invalid OMNIDROP_ENV value: $OMNIDROP_ENV (must be: production, development, test)"
            ;;
    esac
fi

# Check PORT
if [[ -z "$PORT" ]]; then
    report_info "PORT not set, server will use default"
else
    report_info "PORT: $PORT"

    # Critical: Protect production port
    if [[ "$PORT" == "8787" && "$OMNIDROP_ENV" != "production" ]]; then
        report_error "Port 8787 is reserved for production environment only!"
    fi

    # Validate test environment port range
    if [[ "$OMNIDROP_ENV" == "test" ]]; then
        if [[ "$PORT" -lt 8788 || "$PORT" -gt 8799 ]]; then
            report_error "Test environment must use ports 8788-8799, got: $PORT"
        fi
    fi
fi

echo
echo "2. AppleScript Path Validation"
echo "------------------------------"

# Check OMNIDROP_SCRIPT
if [[ -n "$OMNIDROP_SCRIPT" ]]; then
    report_info "OMNIDROP_SCRIPT: $OMNIDROP_SCRIPT"

    # Critical: Protect production script
    PROD_SCRIPT="$HOME/.local/share/omnidrop/omnidrop.applescript"
    if [[ "$OMNIDROP_SCRIPT" == "$PROD_SCRIPT" && "$OMNIDROP_ENV" != "production" ]]; then
        report_error "Cannot use production AppleScript ($PROD_SCRIPT) in non-production environment!"
    fi

    # Verify script exists
    if [[ ! -f "$OMNIDROP_SCRIPT" ]]; then
        report_error "AppleScript file not found: $OMNIDROP_SCRIPT"
    else
        report_success "AppleScript file exists: $OMNIDROP_SCRIPT"
    fi
else
    report_info "OMNIDROP_SCRIPT not set, server will auto-detect"
fi

echo
echo "3. Production Environment Protection"
echo "------------------------------------"

# Check if accidentally running in production context
if [[ "$PWD" == *"production"* || "$PWD" == *".local/share/omnidrop"* ]]; then
    report_error "Running from production-like directory: $PWD"
fi

# Check if production script would be used by default
PROD_SCRIPT="$HOME/.local/share/omnidrop/omnidrop.applescript"
if [[ -f "$PROD_SCRIPT" && -z "$OMNIDROP_SCRIPT" && "$OMNIDROP_ENV" != "production" ]]; then
    # Check if development script exists in current directory
    if [[ ! -f "./omnidrop.applescript" ]]; then
        report_error "No development script found, would fallback to production script!"
        report_info "Solution: Create development script or set OMNIDROP_SCRIPT explicitly"
    fi
fi

echo
echo "4. Token Validation"
echo "-------------------"

if [[ -z "$TOKEN" ]]; then
    report_error "TOKEN environment variable is required"
else
    # Check if using production-like token
    if [[ "$TOKEN" == *"prod"* || "$TOKEN" == *"live"* ]]; then
        report_warning "Token appears to be production-like: ${TOKEN:0:10}..."
    fi

    # Recommend test token format
    if [[ "$OMNIDROP_ENV" == "test" && "$TOKEN" != *"test"* ]]; then
        report_warning "Consider using test-specific token (e.g., test-token-\$\$)"
    fi

    report_success "TOKEN is set (${#TOKEN} characters)"
fi

echo
echo "5. Process Validation"
echo "---------------------"

# Check for existing OmniDrop servers
EXISTING_SERVERS=$(pgrep -f "omnidrop-server" || true)
if [[ -n "$EXISTING_SERVERS" ]]; then
    report_warning "Existing OmniDrop servers detected:"
    ps aux | grep omnidrop-server | grep -v grep || true
    echo "Consider stopping them before testing"
fi

# Check OmniFocus availability
if ! pgrep -f "OmniFocus" > /dev/null; then
    report_warning "OmniFocus not running - AppleScript execution may fail"
else
    report_success "OmniFocus is running"
fi

echo
echo "6. File System Validation"
echo "--------------------------"

# Check current directory
report_info "Current directory: $PWD"

# Check if we're in the right place for development
if [[ ! -f "./omnidrop.applescript" && "$OMNIDROP_ENV" == "development" ]]; then
    report_error "Development environment but no omnidrop.applescript in current directory"
fi

# Check available space for test files
AVAILABLE_SPACE=$(df /tmp | tail -1 | awk '{print $4}')
if [[ "$AVAILABLE_SPACE" -lt 100000 ]]; then  # Less than ~100MB
    report_warning "Low disk space in /tmp: ${AVAILABLE_SPACE}K available"
fi

echo
echo "7. Summary"
echo "----------"

if [[ "$ERRORS" -gt 0 ]]; then
    echo -e "${RED}❌ Pre-flight validation FAILED with $ERRORS error(s)${NC}"
    echo -e "${RED}🚫 Cannot proceed with testing - fix errors above${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Pre-flight validation PASSED${NC}"
    echo -e "${GREEN}🚀 Safe to proceed with testing${NC}"

    # Show configuration summary
    echo
    echo -e "${BLUE}Configuration Summary:${NC}"
    echo "Environment: ${OMNIDROP_ENV:-legacy}"
    echo "Port: ${PORT:-default}"
    echo "Script: ${OMNIDROP_SCRIPT:-auto-detect}"
    echo "Token: ${TOKEN:0:10}... (${#TOKEN} chars)"
    echo
fi

exit 0