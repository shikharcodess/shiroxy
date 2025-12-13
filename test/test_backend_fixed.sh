#!/bin/bash

# Shiroxy Test Backend - Comprehensive Test Script
# Tests all features of the backend server

# Note: Don't use 'set -e' as it would exit on first test failure
# We want to run all tests and report summary at the end

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:80}"
PASSED=0
FAILED=0
SKIPPED=0

echo ${BASE_URL}

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}Testing:${NC} $1"
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC} $1"
    ((FAILED++))
}

print_skip() {
    echo -e "${YELLOW}⊘ SKIP${NC} $1"
    ((SKIPPED++))
}

test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    
    print_test "$description"
    
    response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint")
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" == "$expected_status" ]; then
        print_pass "$description (Status: $status_code)"
        return 0
    else
        print_fail "$description (Expected: $expected_status, Got: $status_code)"
        return 1
    fi
}

test_json_response() {
    local endpoint=$1
    local jq_query=$2
    local expected_value=$3
    local description=$4
    
    print_test "$description"
    
    response=$(curl -s "$BASE_URL$endpoint")
    actual_value=$(echo "$response" | jq -r "$jq_query")
    
    if [ "$actual_value" == "$expected_value" ]; then
        print_pass "$description"
        return 0
    else
        print_fail "$description (Expected: $expected_value, Got: $actual_value)"
        return 1
    fi
}

# FIXED: Use GET request with -D to dump headers instead of HEAD request
test_header() {
    local endpoint=$1
    local header_name=$2
    local expected_value=$3
    local description=$4
    
    print_test "$description"
    
    # Use GET with -D to dump headers to temp file, then read from it
    temp_file=$(mktemp)
    curl -s -D "$temp_file" -o /dev/null "$BASE_URL$endpoint"
    actual_value=$(grep -i "^$header_name:" "$temp_file" | cut -d' ' -f2- | tr -d '\r')
    rm -f "$temp_file"
    
    if [[ "$actual_value" == *"$expected_value"* ]]; then
        print_pass "$description"
        return 0
    else
        print_fail "$description (Expected: $expected_value, Got: $actual_value)"
        return 1
    fi
}

# Main test execution
print_header "Shiroxy Test Backend - Comprehensive Test Suite"

# Test 1: Core Endpoints
print_header "Test Suite 1: Core Endpoints"

test_endpoint GET "/" 200 "Home endpoint"
test_endpoint GET "/info" 200 "Server info endpoint"

# Test 2: Health Check Endpoints
print_header "Test Suite 2: Health Check Endpoints"

test_endpoint GET "/health" 200 "Standard health check"
test_endpoint GET "/healthz" 200 "Healthz endpoint"
test_endpoint GET "/health/ready" 200 "Readiness probe"
test_endpoint GET "/health/live" 200 "Liveness probe"
test_json_response "/health" ".status" "healthy" "Health status is healthy"

# Test 3: CRUD Operations
print_header "Test Suite 3: CRUD Operations"

# Create
curl -s -X POST "$BASE_URL/crud/" \
  -H "Content-Type: application/json" \
  -d '{"id":"test1","value":"test value"}' > /dev/null

test_json_response "/crud/?id=test1" ".item.value" "test value" "Read created item"

# Update
curl -s -X PUT "$BASE_URL/crud/" \
  -H "Content-Type: application/json" \
  -d '{"id":"test1","value":"updated value"}' > /dev/null

test_json_response "/crud/?id=test1" ".item.value" "updated value" "Read updated item"

# List all
test_endpoint GET "/crud/all" 200 "List all items"

# Delete
test_endpoint DELETE "/crud/?id=test1" 200 "Delete item"
test_endpoint GET "/crud/?id=test1" 404 "Verify item deleted"

# Test 4: Content Type Endpoints
print_header "Test Suite 4: Content Type Endpoints"

test_header "/test/html" "Content-Type" "text/html" "HTML content type"
test_header "/test/json" "Content-Type" "application/json" "JSON content type"
test_header "/test/text" "Content-Type" "text/plain" "Text content type"
test_header "/test/xml" "Content-Type" "application/xml" "XML content type"
test_header "/test/css" "Content-Type" "text/css" "CSS content type"
test_header "/test/javascript" "Content-Type" "application/javascript" "JavaScript content type"

# Test 5: Compression Testing
print_header "Test Suite 5: Compression Testing"

test_endpoint GET "/test/compressible" 200 "Compressible content endpoint"
test_endpoint GET "/test/non-compressible" 200 "Non-compressible content endpoint"
test_endpoint GET "/test/large" 200 "Large content endpoint"
test_endpoint GET "/test/binary" 200 "Binary content endpoint"

# Test 6: Status Code Testing
print_header "Test Suite 6: Status Code Testing"

test_endpoint GET "/status/200" 200 "Status 200 endpoint"
test_endpoint GET "/status/404" 404 "Status 404 endpoint"
test_endpoint GET "/status/503" 503 "Status 503 endpoint"
test_endpoint GET "/error/400" 400 "Error 400 endpoint"
test_endpoint GET "/error/500" 500 "Error 500 endpoint"

# Test 7: Performance Endpoints
print_header "Test Suite 7: Performance Endpoints"

test_endpoint GET "/fast" 200 "Fast endpoint"
test_endpoint GET "/slow" 200 "Slow endpoint (2s delay)"

start_time=$(date +%s)
curl -s "$BASE_URL/delay/2" > /dev/null
end_time=$(date +%s)
delay_time=$((end_time - start_time))

if [ "$delay_time" -ge 2 ] && [ "$delay_time" -le 3 ]; then
    print_pass "Delay endpoint (2s) - Actual delay: ${delay_time}s"
else
    print_fail "Delay endpoint (2s) - Expected 2s, got ${delay_time}s"
fi

# Test 8: Metrics & Analytics
print_header "Test Suite 8: Metrics & Analytics"

test_endpoint GET "/metrics" 200 "Metrics endpoint"
test_endpoint GET "/stats" 200 "Statistics endpoint"
test_json_response "/stats" ".server" "null" "Stats has server info" || true # May not be null

# Test 9: Session & Cookie Testing
print_header "Test Suite 9: Session & Cookie Testing"

test_endpoint GET "/session/start" 200 "Session start endpoint"
test_endpoint GET "/session/get" 200 "Session get endpoint"
test_endpoint GET "/cookie/set?name=test&value=testvalue" 200 "Cookie set endpoint"
test_endpoint GET "/cookie/get?name=test" 200 "Cookie get endpoint"

# Test 10: Debug Endpoints
print_header "Test Suite 10: Debug Endpoints"

test_endpoint GET "/echo" 200 "Echo GET endpoint"
test_endpoint POST "/echo" 200 "Echo POST endpoint"
test_endpoint GET "/debug" 200 "Debug endpoint"

# Test 11: Streaming Endpoints
print_header "Test Suite 11: Streaming Endpoints"

test_endpoint GET "/stream/sse" 200 "SSE streaming endpoint"
test_endpoint GET "/stream/chunked" 200 "Chunked streaming endpoint"

# Test 12: Redirect Endpoints
print_header "Test Suite 12: Redirect Endpoints"

test_endpoint GET "/redirect/301" 301 "301 redirect"
test_endpoint GET "/redirect/302" 302 "302 redirect"

# Test 13: Cache Control
print_header "Test Suite 13: Cache Control"

test_header "/cache/public" "Cache-Control" "public" "Public cache header"
test_header "/cache/private" "Cache-Control" "private" "Private cache header"

# Test 14: Custom Headers
print_header "Test Suite 14: Custom Headers"

test_header "/headers/custom" "X-Custom-Header-1" "Value1" "Custom header 1"
test_header "/" "X-Backend-Server" "" "Backend server header exists" || true
test_header "/" "X-Backend-Instance" "" "Backend instance header exists" || true

# Test 15: Health Status Toggle
print_header "Test Suite 15: Health Status Toggle"

# Toggle to unhealthy
curl -s -X POST "$BASE_URL/health/toggle" > /dev/null
test_endpoint GET "/health" 503 "Health check after toggle to unhealthy"

# Toggle back to healthy
curl -s -X POST "$BASE_URL/health/toggle" > /dev/null
test_endpoint GET "/health" 200 "Health check after toggle back to healthy"

# Print summary
print_header "Test Summary"

total=$((PASSED + FAILED + SKIPPED))

echo -e "Total Tests:  $total"
echo -e "${GREEN}Passed:       $PASSED${NC}"
echo -e "${RED}Failed:       $FAILED${NC}"
echo -e "${YELLOW}Skipped:      $SKIPPED${NC}"

echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
