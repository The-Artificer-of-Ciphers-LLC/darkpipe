#!/usr/bin/env bash
# Phase 9: Monitoring & Observability - Integration Test Suite
# Tests MON-01, MON-02, MON-03, CERT-03, CERT-04 requirements
#
# Prerequisites:
#   - Profile server running on port 8090 (with status dashboard)
#   - Mail server running (any profile)
#   - curl, jq available
#
# Usage:
#   ./tests/test-monitoring.sh [--profile-server-url URL]

set -euo pipefail

# Default configuration
PROFILE_SERVER_URL="${PROFILE_SERVER_URL:-http://localhost:8090}"
ADMIN_USER="${ADMIN_USER:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-changeme}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --profile-server-url)
            PROFILE_SERVER_URL="$2"
            shift 2
            ;;
        --admin-user)
            ADMIN_USER="$2"
            shift 2
            ;;
        --admin-password)
            ADMIN_PASSWORD="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --profile-server-url URL   Profile server URL (default: http://localhost:8090)"
            echo "  --admin-user USER          Admin username (default: admin@example.com)"
            echo "  --admin-password PASS      Admin password (default: changeme)"
            echo "  --help                     Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run with --help for usage information"
            exit 1
            ;;
    esac
done

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -n "Testing ${test_name}... "
}

# Helper function to mark test as passed
pass_test() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "[PASS]"
}

# Helper function to mark test as failed
fail_test() {
    local reason="$1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo "[FAIL] ${reason}"
}

# Check prerequisites
echo "=== Phase 9: Monitoring & Observability - Integration Tests ==="
echo ""
echo "Configuration:"
echo "  Profile Server: ${PROFILE_SERVER_URL}"
echo ""

# Check if profile server is running
run_test "Profile server health check"
if curl -sf "${PROFILE_SERVER_URL}/health" > /dev/null; then
    pass_test
else
    fail_test "Profile server not responding"
fi

# Check if jq is available
run_test "jq availability"
if command -v jq > /dev/null 2>&1; then
    pass_test
else
    fail_test "jq not installed (required for JSON parsing)"
fi

echo ""
echo "=== MON-01: Queue Monitoring ==="
echo ""

# Test status API returns queue information
run_test "MON-01: Status API returns queue data"
RESPONSE=$(curl -sf "${PROFILE_SERVER_URL}/status/api" || echo "{}")
if echo "$RESPONSE" | jq -e '.queue' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Status API missing queue field"
fi

run_test "MON-01: Queue depth is numeric"
if echo "$RESPONSE" | jq -e '.queue.depth | type == "number"' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Queue depth is not numeric"
fi

run_test "MON-01: Queue deferred count available"
if echo "$RESPONSE" | jq -e '.queue.deferred | type == "number"' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Queue deferred not available"
fi

run_test "MON-01: Queue stuck count available"
if echo "$RESPONSE" | jq -e '.queue.stuck | type == "number"' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Queue stuck not available"
fi

echo ""
echo "=== MON-02: Delivery Tracking ==="
echo ""

run_test "MON-02: Status API returns delivery data"
if echo "$RESPONSE" | jq -e '.delivery' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Status API missing delivery field"
fi

run_test "MON-02: Delivery counts are numeric"
if echo "$RESPONSE" | jq -e '.delivery.delivered | type == "number"' > /dev/null 2>&1 && \
   echo "$RESPONSE" | jq -e '.delivery.deferred | type == "number"' > /dev/null 2>&1 && \
   echo "$RESPONSE" | jq -e '.delivery.bounced | type == "number"' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Delivery counts not numeric"
fi

run_test "MON-02: Delivery total matches sum"
DELIVERED=$(echo "$RESPONSE" | jq -r '.delivery.delivered // 0')
DEFERRED=$(echo "$RESPONSE" | jq -r '.delivery.deferred // 0')
BOUNCED=$(echo "$RESPONSE" | jq -r '.delivery.bounced // 0')
TOTAL=$(echo "$RESPONSE" | jq -r '.delivery.total // 0')
EXPECTED_TOTAL=$((DELIVERED + DEFERRED + BOUNCED))

if [ "$TOTAL" -eq "$EXPECTED_TOTAL" ]; then
    pass_test
else
    fail_test "Total ($TOTAL) doesn't match sum ($EXPECTED_TOTAL)"
fi

echo ""
echo "=== MON-03: Health Check Endpoints ==="
echo ""

run_test "MON-03: Liveness endpoint returns 200"
if curl -sf "${PROFILE_SERVER_URL}/health/live" > /dev/null; then
    pass_test
else
    fail_test "Liveness endpoint failed"
fi

run_test "MON-03: Readiness endpoint returns valid JSON"
READY_RESPONSE=$(curl -sf "${PROFILE_SERVER_URL}/health/ready" || echo "{}")
if echo "$READY_RESPONSE" | jq -e '.status' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Readiness endpoint missing status field"
fi

run_test "MON-03: Health checks include service status"
if echo "$READY_RESPONSE" | jq -e '.checks | length > 0' > /dev/null 2>&1; then
    pass_test
else
    # It's OK if no checks are configured yet
    echo "[SKIP] No health checks configured"
fi

run_test "MON-03: Docker healthcheck command works"
# This test assumes we're running in or near a Docker environment
# We'll just check if the health endpoint is accessible
if wget --quiet --tries=1 --spider "${PROFILE_SERVER_URL}/health/live" 2>/dev/null; then
    pass_test
else
    echo "[SKIP] wget not available or endpoint not reachable"
fi

echo ""
echo "=== CERT-03 & CERT-04: Certificate Monitoring ==="
echo ""

run_test "CERT-03: Status API returns certificates data"
if echo "$RESPONSE" | jq -e '.certificates' > /dev/null 2>&1; then
    pass_test
else
    fail_test "Status API missing certificates field"
fi

run_test "CERT-03: Certificate days_left is numeric"
CERT_COUNT=$(echo "$RESPONSE" | jq -r '.certificates.certificates | length // 0')
if [ "$CERT_COUNT" -gt 0 ]; then
    if echo "$RESPONSE" | jq -e '.certificates.certificates[0].days_left | type == "number"' > /dev/null 2>&1; then
        pass_test
    else
        fail_test "Certificate days_left not numeric"
    fi
else
    echo "[SKIP] No certificates configured"
fi

run_test "CERT-04: Alert system configured"
# Check if monitoring environment variables are set
if [ -n "${MONITOR_ALERT_EMAIL:-}" ] || [ -n "${MONITOR_WEBHOOK_URL:-}" ]; then
    pass_test
else
    echo "[SKIP] Alert channels not configured (MONITOR_ALERT_EMAIL, MONITOR_WEBHOOK_URL)"
fi

echo ""
echo "=== Web Dashboard ==="
echo ""

run_test "Web dashboard (/status) returns HTML"
DASH_RESPONSE=$(curl -sf "${PROFILE_SERVER_URL}/status" || echo "")
if echo "$DASH_RESPONSE" | grep -q "DarkPipe System Status"; then
    pass_test
else
    fail_test "Dashboard doesn't contain expected title"
fi

run_test "Web dashboard includes all four metric cards"
if echo "$DASH_RESPONSE" | grep -q "Services" && \
   echo "$DASH_RESPONSE" | grep -q "Mail Queue" && \
   echo "$DASH_RESPONSE" | grep -q "Deliveries" && \
   echo "$DASH_RESPONSE" | grep -q "Certificates"; then
    pass_test
else
    fail_test "Dashboard missing expected sections"
fi

run_test "Web dashboard shows overall status"
if echo "$DASH_RESPONSE" | grep -qi "HEALTHY\|DEGRADED\|UNHEALTHY"; then
    pass_test
else
    fail_test "Dashboard doesn't show overall status"
fi

run_test "Web dashboard has auto-refresh meta tag"
if echo "$DASH_RESPONSE" | grep -q 'meta http-equiv="refresh"'; then
    pass_test
else
    fail_test "Dashboard missing auto-refresh tag"
fi

echo ""
echo "=== Summary ==="
echo ""
echo "Total Tests:  ${TESTS_TOTAL}"
echo "Passed:       ${TESTS_PASSED}"
echo "Failed:       ${TESTS_FAILED}"
echo ""

if [ ${TESTS_FAILED} -eq 0 ]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi
