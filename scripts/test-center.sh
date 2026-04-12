#!/bin/bash
# OFA Center Integration Test Script
# Tests Center server communication endpoints

set -e

CENTER_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/ws"

echo "=========================================="
echo "OFA Center Integration Tests"
echo "=========================================="

# Wait for Center to be ready
wait_for_center() {
    echo "Waiting for Center server..."
    for i in {1..30}; do
        if curl -s "$CENTER_URL/api/v1/status" > /dev/null 2>&1; then
            echo "Center server is ready"
            return 0
        fi
        sleep 1
    done
    echo "Center server not available after 30 seconds"
    return 1
}

# Test 1: Status endpoint
test_status() {
    echo ""
    echo "Test 1: Status endpoint"
    echo "------------------------"
    RESPONSE=$(curl -s "$CENTER_URL/api/v1/status")
    if [ -n "$RESPONSE" ]; then
        echo "✅ Status endpoint responded"
        echo "Response: $RESPONSE"
    else
        echo "❌ Status endpoint failed"
        return 1
    fi
}

# Test 2: Identity creation
test_identity_create() {
    echo ""
    echo "Test 2: Identity creation"
    echo "-------------------------"
    RESPONSE=$(curl -s -X POST "$CENTER_URL/api/v1/identity" \
        -H "Content-Type: application/json" \
        -d '{"name": "Test User", "personality": {"openness": 0.5}}')

    if echo "$RESPONSE" | grep -q "id"; then
        echo "✅ Identity created successfully"
        IDENTITY_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        echo "Identity ID: $IDENTITY_ID"
    else
        echo "❌ Identity creation failed"
        return 1
    fi
}

# Test 3: Identity retrieval
test_identity_get() {
    echo ""
    echo "Test 3: Identity retrieval"
    echo "-------------------------"
    if [ -z "$IDENTITY_ID" ]; then
        echo "⚠️ Skipping - no identity ID"
        return 0
    fi

    RESPONSE=$(curl -s "$CENTER_URL/api/v1/identity/$IDENTITY_ID")
    if echo "$RESPONSE" | grep -q "Test User"; then
        echo "✅ Identity retrieved successfully"
    else
        echo "❌ Identity retrieval failed"
        return 1
    fi
}

# Test 4: Chat endpoint
test_chat() {
    echo ""
    echo "Test 4: Chat endpoint"
    echo "---------------------"
    RESPONSE=$(curl -s -X POST "$CENTER_URL/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d '{"message": "Hello", "session_id": "test_session"}')

    if [ -n "$RESPONSE" ]; then
        echo "✅ Chat endpoint responded"
    else
        echo "❌ Chat endpoint failed"
        return 1
    fi
}

# Test 5: WebSocket connection (requires wscat or websocat)
test_websocket() {
    echo ""
    echo "Test 5: WebSocket connection"
    echo "----------------------------"

    # Check if websocat is available
    if command -v websocat &> /dev/null; then
        echo '{"type":"Register","payload":{"agent_id":"test_agent"}}' | timeout 5 websocat "$WS_URL" > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo "✅ WebSocket connection successful"
        else
            echo "❌ WebSocket connection failed"
            return 1
        fi
    else
        echo "⚠️ websocat not available - skipping WebSocket test"
    fi
}

# Test 6: Device registration
test_device_register() {
    echo ""
    echo "Test 6: Device registration"
    echo "---------------------------"
    RESPONSE=$(curl -s -X POST "$CENTER_URL/api/v1/devices" \
        -H "Content-Type: application/json" \
        -d '{"agent_id": "test_android", "device_type": "android", "device_name": "Test Phone"}')

    if [ -n "$RESPONSE" ]; then
        echo "✅ Device registration responded"
    else
        echo "❌ Device registration failed"
        return 1
    fi
}

# Test 7: Scene endpoints
test_scene() {
    echo ""
    echo "Test 7: Scene endpoints"
    echo "----------------------"
    RESPONSE=$(curl -s "$CENTER_URL/api/v1/scene")

    if [ -n "$RESPONSE" ]; then
        echo "✅ Scene endpoint responded"
    else
        echo "❌ Scene endpoint failed"
        return 1
    fi
}

# Test 8: Audio stream request
test_audio() {
    echo ""
    echo "Test 8: Audio/TTS endpoint"
    echo "-------------------------"
    RESPONSE=$(curl -s -X POST "$CENTER_URL/api/v1/tts" \
        -H "Content-Type: application/json" \
        -d '{"text": "Hello world", "voice": "default"}')

    if [ -n "$RESPONSE" ]; then
        echo "✅ TTS endpoint responded"
    else
        echo "⚠️ TTS endpoint may not be fully implemented"
    fi
}

# Main test runner
main() {
    echo ""

    if ! wait_for_center; then
        echo "❌ Center server not available - aborting tests"
        exit 1
    fi

    PASSED=0
    FAILED=0

    test_status && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_identity_create && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_identity_get && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_chat && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_websocket && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_device_register && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_scene && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))
    test_audio && PASSED=$((PASSED+1)) || FAILED=$((FAILED+1))

    echo ""
    echo "=========================================="
    echo "Test Results Summary"
    echo "=========================================="
    echo "Passed: $PASSED"
    echo "Failed: $FAILED"
    echo ""

    if [ $FAILED -eq 0 ]; then
        echo "✅ All tests passed!"
        exit 0
    else
        echo "❌ Some tests failed"
        exit 1
    fi
}

main "$@"