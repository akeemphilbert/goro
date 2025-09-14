#!/bin/bash

# Container Integration Test Runner
# This script runs all container-related BDD and integration tests

set -e

echo "=== Container Integration Test Suite ==="
echo "Running comprehensive container tests..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${YELLOW}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Function to run a test and capture results
run_test() {
    local test_name=$1
    local test_command=$2
    
    print_status "INFO" "Running $test_name..."
    
    if eval "$test_command"; then
        print_status "SUCCESS" "$test_name passed"
        return 0
    else
        print_status "ERROR" "$test_name failed"
        return 1
    fi
}

# Initialize test results
total_tests=0
passed_tests=0
failed_tests=0

# Test categories
declare -a test_categories=(
    "BDD Feature Tests"
    "End-to-End Integration Tests"
    "Performance Tests"
    "Unit Tests"
)

echo "Available test categories:"
for i in "${!test_categories[@]}"; do
    echo "  $((i+1)). ${test_categories[$i]}"
done
echo

# Check if specific category was requested
if [ $# -eq 1 ]; then
    case $1 in
        "1"|"bdd")
            CATEGORY="bdd"
            ;;
        "2"|"e2e")
            CATEGORY="e2e"
            ;;
        "3"|"perf")
            CATEGORY="perf"
            ;;
        "4"|"unit")
            CATEGORY="unit"
            ;;
        "all")
            CATEGORY="all"
            ;;
        *)
            echo "Invalid category. Use: bdd, e2e, perf, unit, or all"
            exit 1
            ;;
    esac
else
    CATEGORY="all"
fi

print_status "INFO" "Running category: $CATEGORY"
echo

# BDD Feature Tests
if [ "$CATEGORY" = "all" ] || [ "$CATEGORY" = "bdd" ]; then
    echo "=== BDD Feature Tests ==="
    
    # Container Basic Integration Tests
    total_tests=$((total_tests + 1))
    if run_test "Container Basic Integration" "go test -v ./features -run TestContainerBasicIntegration"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # Container Domain Logic Tests
    total_tests=$((total_tests + 1))
    if run_test "Container Domain Logic" "go test -v ./features -run TestContainerDomainLogic"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo
fi

# End-to-End Integration Tests
if [ "$CATEGORY" = "all" ] || [ "$CATEGORY" = "e2e" ]; then
    echo "=== End-to-End Integration Tests ==="
    
    total_tests=$((total_tests + 1))
    if run_test "Container Basic Integration" "go test -v ./features -run TestContainerBasicIntegration"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo
fi

# Performance Tests
if [ "$CATEGORY" = "all" ] || [ "$CATEGORY" = "perf" ]; then
    echo "=== Performance Tests ==="
    
    total_tests=$((total_tests + 1))
    if run_test "Container Domain Logic" "go test -v ./features -run TestContainerDomainLogic"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo
fi

# Unit Tests (existing container tests)
if [ "$CATEGORY" = "all" ] || [ "$CATEGORY" = "unit" ]; then
    echo "=== Unit Tests ==="
    
    # Domain layer tests
    total_tests=$((total_tests + 1))
    if run_test "Container Domain Tests" "go test -v ./internal/ldp/domain -run TestContainer"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # Application layer tests
    total_tests=$((total_tests + 1))
    if run_test "Container Service Tests" "go test -v ./internal/ldp/application -run TestContainer"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # Infrastructure layer tests
    total_tests=$((total_tests + 1))
    if run_test "Container Repository Tests" "go test -v ./internal/ldp/infrastructure -run TestContainer"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # HTTP handler tests
    total_tests=$((total_tests + 1))
    if run_test "Container Handler Tests" "go test -v ./internal/infrastructure/transport/http/handlers -run TestContainer"; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo
fi

# Test Summary
echo "=== Test Summary ==="
echo "Total tests: $total_tests"
echo "Passed: $passed_tests"
echo "Failed: $failed_tests"
echo

if [ $failed_tests -eq 0 ]; then
    print_status "SUCCESS" "All container tests passed! 🎉"
    exit 0
else
    print_status "ERROR" "$failed_tests test(s) failed"
    exit 1
fi