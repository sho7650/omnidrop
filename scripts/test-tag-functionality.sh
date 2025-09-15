#!/bin/bash

# Test script for improved tag handling functionality
# This script tests both existing tag registration and new tag auto-creation

set -e  # Exit on any error

echo "🧪 Testing Tag Handling Improvements"
echo "===================================="

# Check if server is built
if [ ! -f "./omnidrop-server" ]; then
    echo "❌ Server not found. Building..."
    cd cmd/omnidrop-server && go build -o ../../omnidrop-server . && cd ../..
fi

# Check for required environment variables
if [ -z "$TOKEN" ]; then
    echo "❌ TOKEN environment variable is required"
    echo "Please set TOKEN=your-secret-token"
    exit 1
fi

# Start server in background on TEST port (NOT production 8787)
echo "🚀 Starting OmniDrop server on test port 8788..."
PORT=8788 TOKEN="$TOKEN" ./omnidrop-server &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Function to cleanup on exit
cleanup() {
    echo "🧹 Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test 1: Existing tag (should work without error)
echo ""
echo "📝 Test 1: Using existing tag"
echo "----------------------------"
curl -s -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task with Existing Tag","note":"Testing existing tag functionality","tags":["urgent"]}' \
  | jq '.'

echo ""
echo "✅ Test 1 completed (check OmniFocus to verify tag was applied)"

# Test 2: New tag creation (should auto-create tag)
echo ""
echo "📝 Test 2: Auto-creating new tag"
echo "-------------------------------"
curl -s -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task with New Tag","note":"Testing new tag auto-creation","tags":["auto-created-tag"]}' \
  | jq '.'

echo ""
echo "✅ Test 2 completed (check OmniFocus to verify new tag was created and applied)"

# Test 3: Mixed existing and new tags
echo ""
echo "📝 Test 3: Mixed existing and new tags"
echo "------------------------------------"
curl -s -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task with Mixed Tags","note":"Testing mixed tag scenario","tags":["urgent","brand-new-tag","work"]}' \
  | jq '.'

echo ""
echo "✅ Test 3 completed (check OmniFocus to verify all tags were processed)"

# Test 4: Invalid tag names (empty, whitespace)
echo ""
echo "📝 Test 4: Edge cases with invalid tag names"
echo "------------------------------------------"
curl -s -X POST http://localhost:8788/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Task with Edge Cases","note":"Testing edge case handling","tags":["valid-tag","","  whitespace-only  "]}' \
  | jq '.'

echo ""
echo "✅ Test 4 completed (should handle edge cases gracefully)"

echo ""
echo "🎉 All tests completed!"
echo "Please check OmniFocus to verify:"
echo "1. Existing tags are properly applied"
echo "2. New tags are automatically created"
echo "3. All tasks were created successfully"
echo ""
echo "Check the AppleScript logs for detailed processing information."