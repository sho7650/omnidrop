#!/bin/bash

# OmniDrop Isolated Test Execution Script
# Creates completely isolated test environment and executes comprehensive tests

set -e  # Exit on any error

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 OmniDrop Isolated Test Execution${NC}"
echo "===================================="

# Configuration
TEST_PORT=${TEST_PORT:-8789}
TEST_TOKEN="test-token-$(date +%s)-$$"
TEST_DIR="/tmp/omnidrop-test-$(date +%s)-$$"
SERVER_PID=""

# Cleanup function
cleanup() {
    echo
    echo -e "${YELLOW}🧹 Cleaning up test environment...${NC}"

    # Stop server if running
    if [[ -n "$SERVER_PID" ]]; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi

    # Kill any remaining omnidrop-server processes on test port
    pkill -f "omnidrop-server.*$TEST_PORT" 2>/dev/null || true

    # Remove test directory
    if [[ -d "$TEST_DIR" ]]; then
        echo "Removing test directory: $TEST_DIR"
        rm -rf "$TEST_DIR"
    fi

    # Reset environment variables
    unset OMNIDROP_ENV OMNIDROP_SCRIPT PORT TOKEN

    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Set trap for cleanup on exit
trap cleanup EXIT

echo
echo "1. Creating Isolated Test Environment"
echo "-------------------------------------"

# Create isolated test directory
echo "Creating test directory: $TEST_DIR"
mkdir -p "$TEST_DIR"

# Copy development AppleScript to isolated location
if [[ ! -f "./omnidrop.applescript" ]]; then
    echo -e "${RED}❌ Development AppleScript not found in current directory${NC}"
    exit 1
fi

echo "Copying AppleScript to isolated environment..."
cp "./omnidrop.applescript" "$TEST_DIR/"

# Verify copy
if [[ ! -f "$TEST_DIR/omnidrop.applescript" ]]; then
    echo -e "${RED}❌ Failed to copy AppleScript to test directory${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Isolated environment created${NC}"

echo
echo "2. Pre-flight Validation"
echo "------------------------"

# Set test environment variables for validation
export OMNIDROP_ENV=test
export OMNIDROP_SCRIPT="$TEST_DIR/omnidrop.applescript"
export PORT=$TEST_PORT
export TOKEN=$TEST_TOKEN

# Run pre-flight validation
if ! ./scripts/test-preflight.sh; then
    echo -e "${RED}❌ Pre-flight validation failed${NC}"
    exit 1
fi
echo "Test directory: $TEST_DIR"
echo "Test port: $TEST_PORT"
echo "Test token: ${TEST_TOKEN:0:15}..."

echo
echo "3. Building Server"
echo "------------------"

# Build server if not exists or if source is newer
if [[ ! -f "./omnidrop-server" ]] || [[ "cmd/omnidrop-server/main.go" -nt "./omnidrop-server" ]]; then
    echo "Building server..."
    go build -o omnidrop-server ./cmd/omnidrop-server
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Server build failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Server built successfully${NC}"
else
    echo -e "${GREEN}✅ Server already built and up-to-date${NC}"
fi

echo
echo "4. Starting Test Server"
echo "-----------------------"

# Set final environment variables
export OMNIDROP_ENV=test
export OMNIDROP_SCRIPT="$TEST_DIR/omnidrop.applescript"
export PORT=$TEST_PORT
export TOKEN=$TEST_TOKEN

echo "Environment configuration:"
echo "  OMNIDROP_ENV=$OMNIDROP_ENV"
echo "  OMNIDROP_SCRIPT=$OMNIDROP_SCRIPT"
echo "  PORT=$PORT"
echo "  TOKEN=${TOKEN:0:15}..."

# Start server in background
echo "Starting server..."
./omnidrop-server &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"

# Wait for server to start
echo "Waiting for server to initialize..."
sleep 3

# Verify server is running
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo -e "${RED}❌ Server failed to start${NC}"
    exit 1
fi

# Test server connectivity
echo "Testing server connectivity..."
for i in {1..5}; do
    if curl -s "http://localhost:$TEST_PORT/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is responding${NC}"
        break
    fi
    if [[ $i -eq 5 ]]; then
        echo -e "${RED}❌ Server not responding after 5 attempts${NC}"
        exit 1
    fi
    echo "Attempt $i/5 - waiting..."
    sleep 2
done

echo
echo "5. Executing Test Suite"
echo "-----------------------"

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run test
run_test() {
    local test_name="$1"
    local test_data="$2"
    local expected_status="$3"

    echo -e "${BLUE}Test: $test_name${NC}"
    TESTS_RUN=$((TESTS_RUN + 1))

    # Make API call
    local response
    response=$(curl -s -X POST "http://localhost:$TEST_PORT/tasks" \
        -H "Authorization: Bearer $TEST_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$test_data")

    local status
    status=$(echo "$response" | jq -r '.status // "unknown"')

    if [[ "$status" == "$expected_status" ]]; then
        echo -e "${GREEN}✅ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo "Expected status: $expected_status, Got: $status"
        echo "Response: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
}

# Test 1: Simple task creation (no tags)
run_test "Simple task creation" \
    '{"title":"Test Task 1","note":"Simple task without tags"}' \
    "ok"

# Test 2: Task with existing tag
run_test "Task with existing tag" \
    '{"title":"Test Task 2","note":"Testing existing tag","tags":["urgent"]}' \
    "ok"

# Test 3: Task with new tag (auto-creation)
run_test "Task with new tag auto-creation" \
    '{"title":"Test Task 3","note":"Testing new tag creation","tags":["test-auto-created-'.$(date +%s)'"]}' \
    "ok"

# Test 4: Task with multiple mixed tags
run_test "Task with multiple mixed tags" \
    '{"title":"Test Task 4","note":"Testing multiple tags","tags":["urgent","new-tag-'.$(date +%s)'","work"]}' \
    "ok"

# Test 5: Task with hierarchical project
run_test "Task with hierarchical project" \
    '{"title":"Test Task 5","note":"Testing project assignment","project":"Getting Things Done/Projects/Test"}' \
    "ok"

# Test 6: Complex task with everything
run_test "Complex task with all features" \
    '{"title":"Test Task 6","note":"Full feature test","project":"Test Project","tags":["complex-test-'.$(date +%s)'","urgent"]}' \
    "ok"

# Test 7: Invalid request (no authorization)
echo -e "${BLUE}Test: Invalid request (no authorization)${NC}"
TESTS_RUN=$((TESTS_RUN + 1))
response=""
response=$(curl -s -X POST "http://localhost:$TEST_PORT/tasks" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test Task"}')

if [[ $(echo "$response" | grep -ci "authorization") -gt 0 ]]; then
    echo -e "${GREEN}✅ PASSED (properly rejected)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}❌ FAILED (should have rejected)${NC}"
    echo "Response: $response"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

echo
echo "6. Test Results Summary"
echo "----------------------"

echo "Tests run: $TESTS_RUN"
echo -e "${GREEN}Tests passed: $TESTS_PASSED${NC}"
if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}Tests failed: $TESTS_FAILED${NC}"
else
    echo -e "${GREEN}Tests failed: $TESTS_FAILED${NC}"
fi

# Calculate success rate
if [[ $TESTS_RUN -gt 0 ]]; then
    success_rate=$((TESTS_PASSED * 100 / TESTS_RUN))
    echo "Success rate: $success_rate%"

    if [[ $success_rate -ge 90 ]]; then
        echo -e "${GREEN}🎉 Excellent test results!${NC}"
    elif [[ $success_rate -ge 70 ]]; then
        echo -e "${YELLOW}⚠️ Good test results, some issues to investigate${NC}"
    else
        echo -e "${RED}❌ Poor test results, significant issues detected${NC}"
    fi
fi

echo
echo -e "${BLUE}💡 Next Steps:${NC}"
echo "1. Check OmniFocus to verify tasks were created with correct tags"
echo "2. Review server logs for any warnings or errors"
echo "3. Run additional manual tests if needed"

# Final result
if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}🎉 All tests passed! Tag functionality is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Review results and fix issues.${NC}"
    exit 1
fi