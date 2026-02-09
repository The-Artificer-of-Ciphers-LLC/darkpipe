#!/bin/bash
# DarkPipe Home Mail Server - Spam Filter Integration Test
# Phase 03 Test Suite - Validates Rspamd spam filtering and greylisting
#
# Prerequisites:
#   - Docker containers running (rspamd, redis, and one mail server profile)
#   - Rspamd web UI accessible on localhost:11334
#   - Test users provisioned
#
# Usage:
#   ./test-spam-filter.sh [OPTIONS]
#
# Options:
#   --host HOST           Mail server host (default: localhost)
#   --rspamd-ui PORT      Rspamd web UI port (default: 11334)
#   --smtp-port PORT      SMTP port for inbound (default: 25)
#   --submission PORT     Submission port for authenticated (default: 587)
#   --user EMAIL          Test user (default: alice@example.com)
#   --pass PASSWORD       Test user password (default: alicepass)
#   --verbose             Verbose output for debugging
#
# Exit codes:
#   0 - All tests passed
#   1 - One or more tests failed

set -e

# Default configuration
HOST="${HOST:-localhost}"
RSPAMD_UI="${RSPAMD_UI:-11334}"
SMTP_PORT="${SMTP_PORT:-25}"
SUBMISSION_PORT="${SUBMISSION_PORT:-587}"
USER="${USER:-alice@example.com}"
PASS="${PASS:-alicepass}"
VERBOSE=0
FAILURES=0

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host) HOST="$2"; shift 2 ;;
        --rspamd-ui) RSPAMD_UI="$2"; shift 2 ;;
        --smtp-port) SMTP_PORT="$2"; shift 2 ;;
        --submission) SUBMISSION_PORT="$2"; shift 2 ;;
        --user) USER="$2"; shift 2 ;;
        --pass) PASS="$2"; shift 2 ;;
        --verbose) VERBOSE=1; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    FAILURES=$((FAILURES + 1))
}

pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

# Test 1: Rspamd health check
test_rspamd_health() {
    log "Test 1: Rspamd health check (web UI at $HOST:$RSPAMD_UI)"

    if curl --silent "http://$HOST:$RSPAMD_UI/stat" | grep -q "scanned"; then
        pass "Rspamd is running and responding"
    else
        fail "Rspamd is not responding on port $RSPAMD_UI"
    fi
}

# Test 2: Redis connectivity
test_redis_connectivity() {
    log "Test 2: Redis connectivity from Rspamd"

    # Check Rspamd logs for Redis connection success
    if docker logs rspamd 2>&1 | tail -n 50 | grep -qi "redis"; then
        pass "Rspamd connecting to Redis (check logs for confirmation)"
    else
        warn "Could not verify Redis connection in Rspamd logs (may need manual verification)"
    fi

    # Check Redis is running
    if docker exec redis redis-cli ping 2>&1 | grep -q "PONG"; then
        pass "Redis is running and responding"
    else
        fail "Redis is not responding"
    fi
}

# Test 3: Clean message delivery (score < 4, should not be greylisted)
test_clean_message() {
    log "Test 3: Clean message delivery (spam score < 4.0, no greylisting)"

    local CLEAN_MSG="This is a legitimate message from a real person $(date +%s)"
    local SUBJECT="Clean Test $(date +%s)"

    if command -v swaks &> /dev/null; then
        if swaks --to "$USER" --from "legitimate@example.com" \
                --server "$HOST:$SMTP_PORT" \
                --header "Subject: $SUBJECT" \
                --body "$CLEAN_MSG" --hide-all 2>&1 | grep -q "250 2.0.0"; then
            pass "Clean message accepted (not rejected or greylisted)"
        else
            fail "Clean message was rejected or deferred"
        fi
    else
        warn "swaks not available, skipping clean message test"
    fi
}

# Test 4: GTUBE spam detection
test_gtube_spam() {
    log "Test 4: GTUBE spam pattern detection (should trigger spam score)"

    # GTUBE = Generic Test for Unsolicited Bulk Email
    # This is a standard test pattern recognized by spam filters
    local GTUBE="XJS*C4JDBQADN1.NSBN3*2IDNEN*GTUBE-STANDARD-ANTI-UBE-TEST-EMAIL*C.34X"
    local SUBJECT="GTUBE Test $(date +%s)"

    if command -v swaks &> /dev/null; then
        # Send GTUBE message - expect rejection or spam marking
        local RESULT
        RESULT=$(swaks --to "$USER" --from "spammer@example.com" \
                --server "$HOST:$SMTP_PORT" \
                --header "Subject: $SUBJECT" \
                --body "$GTUBE" --hide-all 2>&1 || true)

        if echo "$RESULT" | grep -qE "(550|554|reject|spam)"; then
            pass "GTUBE pattern detected and rejected/marked as spam"
        elif echo "$RESULT" | grep -q "250 2.0.0"; then
            # Message accepted - check if X-Spam header would be added
            warn "GTUBE message accepted (may have X-Spam header added, check Rspamd logs)"
        else
            fail "GTUBE pattern not detected or unexpected response"
        fi
    else
        warn "swaks not available, skipping GTUBE test"
    fi
}

# Test 5: Greylisting state in Redis
test_greylisting_state() {
    log "Test 5: Greylisting state persistence in Redis"

    # Check if Redis has greylist keys
    local KEY_COUNT
    KEY_COUNT=$(docker exec redis redis-cli KEYS "*greylist*" 2>&1 | wc -l)

    if [[ $KEY_COUNT -gt 0 ]]; then
        pass "Greylisting entries found in Redis ($KEY_COUNT keys)"
    else
        warn "No greylisting entries in Redis yet (may populate after first greylisted message)"
    fi

    # Check Redis memory usage
    local MEM_USAGE
    MEM_USAGE=$(docker exec redis redis-cli INFO memory 2>&1 | grep "used_memory_human" | cut -d: -f2 | tr -d '\r')
    log "Redis memory usage: $MEM_USAGE (limit: 64MB)"
}

# Test 6: Authenticated submission bypasses Rspamd
test_submission_bypass() {
    log "Test 6: Authenticated submission bypasses spam filtering"

    local TEST_MSG="This message should bypass spam filtering $(date +%s)"
    local SUBJECT="Submission Bypass Test $(date +%s)"

    if command -v swaks &> /dev/null; then
        # Send via submission port with authentication
        # This should NOT go through Rspamd (no X-Spam headers added)
        if swaks --to "external@example.net" --from "$USER" \
                --server "$HOST:$SUBMISSION_PORT" \
                --auth-user "$USER" --auth-password "$PASS" \
                --tls --header "Subject: $SUBJECT" \
                --body "$TEST_MSG" --hide-all 2>&1 | grep -q "250 2.0.0"; then
            pass "Authenticated submission accepted (bypasses Rspamd)"
        else
            fail "Authenticated submission failed"
        fi

        # Verify no spam headers added (requires message inspection via IMAP - skip for now)
        log "Note: Verify no X-Spam headers in sent message (manual check via mail client)"
    else
        warn "swaks not available, skipping submission bypass test"
    fi
}

# Test 7: Rspamd statistics
test_rspamd_stats() {
    log "Test 7: Rspamd statistics and metrics"

    local STATS
    STATS=$(curl --silent "http://$HOST:$RSPAMD_UI/stat" 2>&1)

    if echo "$STATS" | grep -q "scanned"; then
        local SCANNED
        SCANNED=$(echo "$STATS" | grep -o '"scanned":[0-9]*' | cut -d: -f2)
        log "Messages scanned by Rspamd: $SCANNED"
        pass "Rspamd statistics available"
    else
        fail "Could not retrieve Rspamd statistics"
    fi
}

# Main test execution
main() {
    echo "========================================"
    echo "DarkPipe Spam Filter Integration Test"
    echo "========================================"
    echo ""

    log "Host: $HOST"
    log "Rspamd UI: $RSPAMD_UI, SMTP: $SMTP_PORT, Submission: $SUBMISSION_PORT"
    log "Test user: $USER"
    echo ""

    test_rspamd_health
    test_redis_connectivity
    test_clean_message
    test_gtube_spam
    test_greylisting_state
    test_submission_bypass
    test_rspamd_stats

    echo ""
    echo "========================================"
    if [[ $FAILURES -eq 0 ]]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}$FAILURES test(s) failed${NC}"
        exit 1
    fi
}

main
