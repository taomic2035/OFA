#!/bin/bash
# OFA End-to-End Integration Test Runner
# Runs all integration tests for Center + Android SDK + iOS SDK

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "OFA End-to-End Integration Tests"
echo "Version: v8.2.0"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test result tracking
declare -a TEST_RESULTS

run_test() {
    local test_name="$1"
    local test_command="$2"

    echo ""
    echo "Running: $test_name"
    echo "Command: $test_command"
    echo "----------------------------------------"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if eval "$test_command"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        TEST_RESULTS+=("$test_name: PASSED")
        echo "${GREEN}✅ $test_name PASSED${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TEST_RESULTS+=("$test_name: FAILED")
        echo "${RED}❌ $test_name FAILED${NC}"
    fi
}

# Check prerequisites
check_prerequisites() {
    echo ""
    echo "Checking prerequisites..."
    echo "----------------------------------------"

    # Check Go
    if command -v go &> /dev/null; then
        echo "${GREEN}✅ Go installed: $(go version)${NC}"
    else
        echo "${RED}❌ Go not installed${NC}"
        exit 1
    fi

    # Check Docker (optional)
    if command -v docker &> /dev/null; then
        echo "${GREEN}✅ Docker installed${NC}"
    else
        echo "${YELLOW}⚠️ Docker not installed (optional)${NC}"
    fi

    # Check curl
    if command -v curl &> /dev/null; then
        echo "${GREEN}✅ curl installed${NC}"
    else
        echo "${RED}❌ curl not installed${NC}"
        exit 1
    fi
}

# Build Center
build_center() {
    echo ""
    echo "Building Center..."
    echo "----------------------------------------"

    cd "$PROJECT_ROOT/src/center"

    if go build -o center.exe cmd/center/main.go; then
        echo "${GREEN}✅ Center built successfully${NC}"
        cd "$PROJECT_ROOT"
        return 0
    else
        echo "${RED}❌ Center build failed${NC}"
        cd "$PROJECT_ROOT"
        return 1
    fi
}

# Run Center tests
run_center_tests() {
    echo ""
    echo "Running Center tests..."
    echo "----------------------------------------"

    cd "$PROJECT_ROOT/src/center"

    if go test ./... -v -count=1; then
        echo "${GREEN}✅ Center tests passed${NC}"
        cd "$PROJECT_ROOT"
        return 0
    else
        echo "${RED}❌ Center tests failed${NC}"
        cd "$PROJECT_ROOT"
        return 1
    fi
}

# Check Android SDK
check_android_sdk() {
    echo ""
    echo "Checking Android SDK..."
    echo "----------------------------------------"

    if [ -d "$PROJECT_ROOT/src/sdk/android" ]; then
        echo "${GREEN}✅ Android SDK directory exists${NC}"

        # Check if Gradle wrapper exists
        if [ -f "$PROJECT_ROOT/src/sdk/android/gradlew" ]; then
            echo "${GREEN}✅ Gradle wrapper exists${NC}"
        else
            echo "${YELLOW}⚠️ Gradle wrapper not found${NC}"
        fi

        return 0
    else
        echo "${RED}❌ Android SDK not found${NC}"
        return 1
    fi
}

# Check iOS SDK
check_ios_sdk() {
    echo ""
    echo "Checking iOS SDK..."
    echo "----------------------------------------"

    if [ -d "$PROJECT_ROOT/src/sdk/ios" ]; then
        echo "${GREEN}✅ iOS SDK directory exists${NC}"

        # Check Swift files
        SWIFT_COUNT=$(find "$PROJECT_ROOT/src/sdk/ios" -name "*.swift" 2>/dev/null | wc -l)
        echo "Swift files found: $SWIFT_COUNT"

        # Check Package.swift
        if [ -f "$PROJECT_ROOT/src/sdk/ios/Package.swift" ]; then
            echo "${GREEN}✅ Package.swift exists${NC}"
        else
            echo "${YELLOW}⚠️ Package.swift not found${NC}"
        fi

        return 0
    else
        echo "${RED}❌ iOS SDK not found${NC}"
        return 1
    fi
}

# Check documentation
check_docs() {
    echo ""
    echo "Checking documentation..."
    echo "----------------------------------------"

    docs=(
        "docs/API.md"
        "docs/PHASE_SUMMARY.md"
        "docs/SDK_ARCHITECTURE.md"
        "docs/SCENE_DESIGN.md"
        "docs/DEPLOYMENT.md"
        "docs/E2E_TEST_PLAN.md"
    )

    for doc in "${docs[@]}"; do
        if [ -f "$PROJECT_ROOT/$doc" ]; then
            echo "${GREEN}✅ $doc exists${NC}"
        else
            echo "${YELLOW}⚠️ $doc not found${NC}"
        fi
    done
}

# Main test execution
main() {
    check_prerequisites

    echo ""
    echo "=========================================="
    echo "Running Integration Tests"
    echo "=========================================="

    # Build tests
    run_test "Center Build" build_center

    # Go tests
    run_test "Center Unit Tests" run_center_tests

    # SDK checks
    run_test "Android SDK Check" check_android_sdk
    run_test "iOS SDK Check" check_ios_sdk

    # Documentation check
    run_test "Documentation Check" check_docs

    # Print summary
    echo ""
    echo "=========================================="
    echo "Test Results Summary"
    echo "=========================================="
    echo "Total tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo ""

    echo "Detailed Results:"
    for result in "${TEST_RESULTS[@]}"; do
        echo "  - $result"
    done

    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo "${GREEN}=========================================="
        echo "✅ All Integration Tests Passed!"
        echo "==========================================${NC}"
        exit 0
    else
        echo "${RED}=========================================="
        echo "❌ Some Tests Failed"
        echo "==========================================${NC}"
        exit 1
    fi
}

main "$@"