#!/bin/bash

# BDD Test Runner for Resource Storage System
# This script runs all behavior-driven development tests

set -e

echo "🧪 Running BDD Tests for Resource Storage System"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to run a specific test category
run_test_category() {
    local test_name=$1
    local test_pattern=$2
    
    echo -e "\n${BLUE}📋 Running $test_name Tests${NC}"
    echo "----------------------------------------"
    
    if go test ./features/ -run "$test_pattern" -v; then
        echo -e "${GREEN}✅ $test_name tests passed${NC}"
    else
        echo -e "${RED}❌ $test_name tests failed${NC}"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    echo -e "\n${BLUE}🚀 Running All BDD Tests${NC}"
    echo "----------------------------------------"
    
    if go test ./features/... -v; then
        echo -e "${GREEN}✅ All BDD tests passed${NC}"
    else
        echo -e "${RED}❌ Some BDD tests failed${NC}"
        return 1
    fi
}

# Check if specific test category is requested
case "${1:-all}" in
    "rdf")
        run_test_category "RDF Format Support" "TestRDFFormatSupport"
        ;;
    "binary")
        run_test_category "Binary File Storage" "TestBinaryFileStorage"
        ;;
    "error")
        run_test_category "Error Handling" "TestErrorHandling"
        ;;
    "performance")
        run_test_category "Performance Requirements" "TestPerformanceRequirements"
        ;;
    "integrity")
        run_test_category "Data Integrity" "TestDataIntegrity"
        ;;
    "all")
        run_all_tests
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [category]"
        echo ""
        echo "Categories:"
        echo "  rdf         - RDF Format Support tests"
        echo "  binary      - Binary File Storage tests"
        echo "  error       - Error Handling tests"
        echo "  performance - Performance Requirements tests"
        echo "  integrity   - Data Integrity tests"
        echo "  all         - All BDD tests (default)"
        echo "  help        - Show this help message"
        exit 0
        ;;
    *)
        echo -e "${RED}❌ Unknown test category: $1${NC}"
        echo "Use '$0 help' to see available categories"
        exit 1
        ;;
esac

echo -e "\n${GREEN}🎉 BDD Test Run Complete${NC}"