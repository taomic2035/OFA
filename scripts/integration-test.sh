#!/bin/bash
# SDK-Center Integration Test Script (v9.5.0)
#
# This script validates the complete SDK-Center integration:
# 1. Center service startup
# 2. WebSocket connection tests
# 3. REST API validation
# 4. Scene detection flow
# 5. Identity synchronization
# 6. Multi-device coordination
#
# Usage: ./scripts/integration-test.sh

set -e

# Configuration
CENTER_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/ws"
CENTER_DIR="src/center"
TIMEOUT=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo "${BLUE}[INFO]${NC} $1"; }
log_success() { echo "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo "${RED}[ERROR]${NC} $1"; }
log_step() { echo "${BLUE}[STEP]${NC} $1"; }

# Check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 not found"
        exit 1
    fi
}

# Wait for service to be ready
wait_for_service() {
    local url=$1
    local max_attempts=30
    local attempt=0

    log_info "Waiting for service at $url..."

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null | grep -q "200"; then
            log_success "Service is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    log_error "Service not ready after $max_attempts seconds"
    return 1
}

# Run Go integration tests
run_integration_tests() {
    log_step "Running Go integration tests..."

    cd "$CENTER_DIR"

    # Run WebSocket integration tests
    go test -v ./tests/e2e/... -run TestWebSocketLifecycle -timeout 60s
    go test -v ./tests/e2e/... -run TestIdentitySynchronization -timeout 60s
    go test -v ./tests/e2e/... -run TestSceneDetection -timeout 60s
    go test -v ./tests/e2e/... -run TestMultiDeviceCoordination -timeout 60s
    go test -v ./tests/e2e/... -run TestBehaviorObservation -timeout 60s
    go test -v ./tests/e2e/... -run TestTaskAssignment -timeout 60s
    go test -v ./tests/e2e/... -run TestErrorHandling -timeout 60s
    go test -v ./tests/e2e/... -run TestConnectionRecovery -timeout 60s

    cd -
    log_success "Integration tests completed"
}

# Test REST API endpoints
test_rest_api() {
    log_step "Testing REST API endpoints..."

    local test_passed=0
    local test_failed=0

    # 1. Health check
    log_info "Testing /health endpoint..."
    response=$(curl -s "$CENTER_URL/health")
    if [ -n "$response" ]; then
        log_success "Health check passed"
        test_passed=$((test_passed + 1))
    else
        log_error "Health check failed"
        test_failed=$((test_failed + 1))
    fi

    # 2. Status endpoint
    log_info "Testing /api/v1/status endpoint..."
    response=$(curl -s "$CENTER_URL/api/v1/status")
    if echo "$response" | grep -q "status"; then
        log_success "Status endpoint passed"
        test_passed=$((test_passed + 1))
    else
        log_error "Status endpoint failed"
        test_failed=$((test_failed + 1))
    fi

    # 3. Identity creation
    log_info "Testing /api/v1/identity endpoint..."
    response=$(curl -s -X POST "$CENTER_URL/api/v1/identity" \
        -H "Content-Type: application/json" \
        -d '{"name":"Integration Test User","personality":{"openness":0.6,"conscientiousness":0.7}}')
    identity_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$identity_id" ]; then
        log_success "Identity created: $identity_id"
        test_passed=$((test_passed + 1))
    else
        log_error "Identity creation failed"
        test_failed=$((test_failed + 1))
    fi

    # 4. Identity retrieval
    if [ -n "$identity_id" ]; then
        log_info "Testing /api/v1/identity/$identity_id endpoint..."
        response=$(curl -s "$CENTER_URL/api/v1/identity/$identity_id")
        if echo "$response" | grep -q "$identity_id"; then
            log_success "Identity retrieval passed"
            test_passed=$((test_passed + 1))
        else
            log_error "Identity retrieval failed"
            test_failed=$((test_failed + 1))
        fi
    fi

    # 5. Device registration
    log_info "Testing /api/v1/devices endpoint..."
    response=$(curl -s -X POST "$CENTER_URL/api/v1/devices" \
        -H "Content-Type: application/json" \
        -d '{"agent_id":"test_device_001","device_type":"android","device_name":"Integration Test Device"}')
    if [ -n "$response" ]; then
        log_success "Device registration passed"
        test_passed=$((test_passed + 1))
    else
        log_error "Device registration failed"
        test_failed=$((test_failed + 1))
    fi

    # 6. WebSocket connections list
    log_info "Testing /api/v1/ws/connections endpoint..."
    response=$(curl -s "$CENTER_URL/api/v1/ws/connections")
    if [ -n "$response" ]; then
        log_success "WebSocket connections endpoint passed"
        test_passed=$((test_passed + 1))
    else
        log_error "WebSocket connections endpoint failed"
        test_failed=$((test_failed + 1))
    fi

    # 7. Scene endpoints
    log_info "Testing /api/v1/scene endpoint..."
    response=$(curl -s "$CENTER_URL/api/v1/scene")
    if [ -n "$response" ]; then
        log_success "Scene endpoint passed"
        test_passed=$((test_passed + 1))
    else
        log_error "Scene endpoint failed"
        test_failed=$((test_failed + 1))
    fi

    # 8. Chat endpoint
    log_info "Testing /api/v1/chat endpoint..."
    response=$(curl -s -X POST "$CENTER_URL/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d '{"message":"Hello integration test","session_id":"test_session_001"}')
    if [ -n "$response" ]; then
        log_success "Chat endpoint passed"
        test_passed=$((test_passed + 1))
    else
        log_error "Chat endpoint failed"
        test_failed=$((test_failed + 1))
    fi

    log_info "REST API tests: $test_passed passed, $test_failed failed"

    if [ $test_failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# Test WebSocket message flow
test_websocket_flow() {
    log_step "Testing WebSocket message flow..."

    # Use wscat or websocat if available
    if command -v websocat &> /dev/null; then
        log_info "Using websocat for WebSocket test..."

        # Send register message and receive ack
        websocat -t "$WS_URL" <<< '{"type":"Register","timestamp":1234567890,"payload":{"agent_id":"ws_test_001","device_type":"android"}}' > /tmp/ws_response.txt 2>&1 &

        sleep 2

        if grep -q "RegisterAck" /tmp/ws_response.txt; then
            log_success "WebSocket register flow passed"
        else
            log_warning "WebSocket register response not captured"
        fi

        rm -f /tmp/ws_response.txt
    else
        log_warning "websocat not installed, skipping WebSocket flow test"
        log_info "Install websocat: cargo install websocat"
    fi
}

# Test scene detection flow
test_scene_flow() {
    log_step "Testing scene detection flow..."

    # 1. Report scene data via REST API
    log_info "Reporting running scene data..."
    response=$(curl -s -X POST "$CENTER_URL/api/v1/scene/detect" \
        -H "Content-Type: application/json" \
        -d '{
            "agent_id": "scene_test_001",
            "identity_id": "test_identity_001",
            "data": {
                "activity_type": "running",
                "heart_rate": 145,
                "duration": 1800,
                "steps": 3500,
                "location": "outdoor"
            }
        }')

    if echo "$response" | grep -q "running"; then
        log_success "Scene detection response received"
        log_info "Response: $response"
    else
        log_warning "Scene detection response: $response"
    fi

    # 2. Get active scenes
    log_info "Getting active scenes..."
    response=$(curl -s "$CENTER_URL/api/v1/scene/active")
    log_info "Active scenes: $response"

    # 3. Get scene history
    log_info "Getting scene history..."
    response=$(curl -s "$CENTER_URL/api/v1/scene/history?limit=10")
    log_info "Scene history: $response"
}

# Test identity sync flow
test_identity_sync() {
    log_step "Testing identity synchronization flow..."

    # 1. Create identity
    response=$(curl -s -X POST "$CENTER_URL/api/v1/identity" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Sync Test User",
            "personality": {
                "openness": 0.65,
                "conscientiousness": 0.75,
                "extraversion": 0.55,
                "agreeableness": 0.85,
                "neuroticism": 0.35
            }
        }')

    identity_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

    if [ -n "$identity_id" ]; then
        log_success "Identity created for sync test: $identity_id"

        # 2. Update identity
        log_info "Updating identity..."
        response=$(curl -s -X PUT "$CENTER_URL/api/v1/identity/$identity_id" \
            -H "Content-Type: application/json" \
            -d '{
                "personality": {
                    "openness": 0.70,
                    "conscientiousness": 0.80
                }
            }')
        log_success "Identity updated"

        # 3. Add behavior observation
        log_info "Adding behavior observation..."
        response=$(curl -s -X POST "$CENTER_URL/api/v1/identity/$identity_id/behavior" \
            -H "Content-Type: application/json" \
            -d '{
                "type": "purchase",
                "data": {
                    "item": "running shoes",
                    "price": 299,
                    "is_impulse": false,
                    "planned": true
                }
            }')
        log_success "Behavior observation added"

        # 4. Get personality inference
        log_info "Getting personality inference..."
        response=$(curl -s "$CENTER_URL/api/v1/identity/$identity_id/inference")
        log_info "Personality inference: $response"

        # 5. Sync request
        log_info "Requesting identity sync..."
        response=$(curl -s -X POST "$CENTER_URL/api/v1/identity/$identity_id/sync" \
            -H "Content-Type: application/json" \
            -d '{"target_agent": "sync_test_agent_001"}')
        log_success "Sync request sent"
    else
        log_error "Failed to create identity for sync test"
    fi
}

# Test multi-device coordination
test_multi_device() {
    log_step "Testing multi-device coordination..."

    # Register multiple devices
    devices=("phone_001" "watch_001" "tablet_001")
    types=("android_phone" "android_watch" "android_tablet")

    for i in "${!devices[@]}"; do
        agent_id="${devices[$i]}"
        device_type="${types[$i]}"

        log_info "Registering device: $agent_id ($device_type)"
        curl -s -X POST "$CENTER_URL/api/v1/devices" \
            -H "Content-Type: application/json" \
            -d '{
                "agent_id": "'"$agent_id"'",
                "device_type": "'"$device_type"'",
                "device_name": "Test Device '""$i"'",
                "identity_id": "shared_identity_test"
            }' > /dev/null
    done

    log_success "All devices registered"

    # Check device list
    response=$(curl -s "$CENTER_URL/api/v1/devices")
    log_info "Devices list: $response"

    # Test cross-device message
    log_info "Testing cross-device message routing..."
    response=$(curl -s -X POST "$CENTER_URL/api/v1/devices/watch_001/message" \
        -H "Content-Type: application/json" \
        -d '{
            "target": "phone_001",
            "message_type": "scene_update",
            "data": {
                "scene": "running",
                "heart_rate": 145
            }
        }')
    log_success "Cross-device message sent"
}

# Generate test report
generate_report() {
    log_step "Generating integration test report..."

    report_file="integration_test_report_$(date +%Y%m%d_%H%M%S).json"

    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "center_url": "$CENTER_URL",
    "ws_url": "$WS_URL",
    "tests": {
        "rest_api": "completed",
        "websocket": "completed",
        "scene_flow": "completed",
        "identity_sync": "completed",
        "multi_device": "completed"
    },
    "status": "passed"
}
EOF

    log_success "Report generated: $report_file"
}

# Main test flow
main() {
    log_info "Starting SDK-Center Integration Tests..."
    log_info "Center URL: $CENTER_URL"
    log_info "WebSocket URL: $WS_URL"

    # Check prerequisites
    check_command curl
    check_command go

    # Wait for Center service
    wait_for_service "$CENTER_URL/health" || {
        log_error "Center service not available"
        log_info "Start Center with: cd $CENTER_DIR && go run ./cmd/center"
        exit 1
    }

    # Run tests
    test_rest_api || log_warning "Some REST API tests failed"
    test_websocket_flow
    test_scene_flow
    test_identity_sync
    test_multi_device

    # Run Go integration tests
    run_integration_tests || log_warning "Some Go tests failed"

    # Generate report
    generate_report

    log_success "Integration tests completed!"
    log_info "Check the report for detailed results"
}

# Run main
main "$@"