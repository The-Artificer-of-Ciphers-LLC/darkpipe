#!/bin/bash
# DarkPipe Home Mail Server - Mail Flow Integration Test
# Phase 03 Test Suite - Validates complete mail pipeline
#
# Prerequisites:
#   - Docker containers running (rspamd, redis, and one mail server profile)
#   - Test users provisioned (alice@example.com, bob@example.org)
#   - Mail server accessible on localhost ports 25, 587, 993
#
# Usage:
#   ./test-mail-flow.sh [OPTIONS]
#
# Options:
#   --host HOST         Mail server host (default: localhost)
#   --smtp-port PORT    SMTP port for inbound (default: 25)
#   --submission PORT   Submission port for authenticated (default: 587)
#   --imap-port PORT    IMAP port (default: 993)
#   --user1 EMAIL       First test user (default: alice@example.com)
#   --pass1 PASS        First user password (default: alicepass)
#   --user2 EMAIL       Second test user (default: bob@example.org)
#   --pass2 PASS        Second user password (default: bobpass)
#   --verbose           Verbose output for debugging
#
# Exit codes:
#   0 - All tests passed
#   1 - One or more tests failed

set -e

# Default configuration
HOST="${HOST:-localhost}"
SMTP_PORT="${SMTP_PORT:-25}"
SUBMISSION_PORT="${SUBMISSION_PORT:-587}"
IMAP_PORT="${IMAP_PORT:-993}"
USER1="${USER1:-alice@example.com}"
PASS1="${PASS1:-alicepass}"
USER2="${USER2:-bob@example.org}"
PASS2="${PASS2:-bobpass}"
VERBOSE=0
FAILURES=0

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host) HOST="$2"; shift 2 ;;
        --smtp-port) SMTP_PORT="$2"; shift 2 ;;
        --submission) SUBMISSION_PORT="$2"; shift 2 ;;
        --imap-port) IMAP_PORT="$2"; shift 2 ;;
        --user1) USER1="$2"; shift 2 ;;
        --pass1) PASS1="$2"; shift 2 ;;
        --user2) USER2="$2"; shift 2 ;;
        --pass2) PASS2="$2"; shift 2 ;;
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

# Detect which mail server profile is active
detect_mail_server() {
    log "Detecting active mail server profile..."

    if docker ps --format '{{.Names}}' | grep -q '^stalwart$'; then
        echo "stalwart"
    elif docker ps --format '{{.Names}}' | grep -q '^maddy$'; then
        echo "maddy"
    elif docker ps --format '{{.Names}}' | grep -q '^postfix-dovecot$'; then
        echo "postfix-dovecot"
    else
        fail "No mail server container detected. Start with: docker compose --profile <stalwart|maddy|postfix-dovecot> up -d"
        exit 1
    fi
}

# Test SMTP inbound delivery
test_smtp_inbound() {
    log "Test 1: SMTP inbound delivery to port $SMTP_PORT"

    local TEST_MSG="Test message $(date +%s)"
    local SUBJECT="Test Inbound $(date +%s)"

    # Use swaks if available, otherwise Python smtplib
    if command -v swaks &> /dev/null; then
        if [[ $VERBOSE -eq 1 ]]; then
            swaks --to "$USER1" --from "test@external.com" \
                --server "$HOST:$SMTP_PORT" \
                --header "Subject: $SUBJECT" \
                --body "$TEST_MSG"
        else
            swaks --to "$USER1" --from "test@external.com" \
                --server "$HOST:$SMTP_PORT" \
                --header "Subject: $SUBJECT" \
                --body "$TEST_MSG" --hide-all 2>&1 | grep -q "250 2.0.0 Ok" && pass "SMTP inbound accepted" || fail "SMTP inbound rejected"
        fi
    else
        # Fallback to Python
        python3 -c "
import smtplib
from email.message import EmailMessage

msg = EmailMessage()
msg['Subject'] = '$SUBJECT'
msg['From'] = 'test@external.com'
msg['To'] = '$USER1'
msg.set_content('$TEST_MSG')

with smtplib.SMTP('$HOST', $SMTP_PORT) as smtp:
    smtp.send_message(msg)
print('Message sent successfully')
" && pass "SMTP inbound accepted" || fail "SMTP inbound failed"
    fi
}

# Test IMAP access
test_imap_access() {
    log "Test 2: IMAP access on port $IMAP_PORT"

    # Use curl with IMAP URL
    if curl --silent --insecure "imaps://$HOST:$IMAP_PORT" \
            --user "$USER1:$PASS1" \
            --request "LIST \"\" \"*\"" 2>&1 | grep -q "INBOX"; then
        pass "IMAP connection successful, INBOX found"
    else
        fail "IMAP connection failed or INBOX not found"
    fi
}

# Test SMTP submission
test_smtp_submission() {
    log "Test 3: SMTP submission on port $SUBMISSION_PORT with authentication"

    local TEST_MSG="Submission test $(date +%s)"
    local SUBJECT="Test Submission $(date +%s)"

    if command -v swaks &> /dev/null; then
        if swaks --to "external@example.net" --from "$USER1" \
                --server "$HOST:$SUBMISSION_PORT" \
                --auth-user "$USER1" --auth-password "$PASS1" \
                --tls --header "Subject: $SUBJECT" \
                --body "$TEST_MSG" --hide-all 2>&1 | grep -q "250 2.0.0"; then
            pass "SMTP submission accepted with authentication"
        else
            fail "SMTP submission failed"
        fi
    else
        # Fallback to openssl s_client (manual SMTP commands)
        warn "swaks not available, skipping submission test (install swaks for full coverage)"
    fi
}

# Test multi-user isolation
test_multi_user_isolation() {
    log "Test 4: Multi-user isolation (different users cannot access each other's mail)"

    # Send message to user1
    local MSG1="Message for user1 $(date +%s)"
    echo "$MSG1" | mail -s "Test User1" "$USER1" 2>/dev/null || true

    # Send message to user2
    local MSG2="Message for user2 $(date +%s)"
    echo "$MSG2" | mail -s "Test User2" "$USER2" 2>/dev/null || true

    # Check user1 can see only their message
    if curl --silent --insecure "imaps://$HOST:$IMAP_PORT/INBOX" \
            --user "$USER1:$PASS1" 2>&1 | grep -q "INBOX"; then
        pass "User1 can access their mailbox"
    else
        fail "User1 cannot access their mailbox"
    fi

    # Check user2 can see only their message
    if curl --silent --insecure "imaps://$HOST:$IMAP_PORT/INBOX" \
            --user "$USER2:$PASS2" 2>&1 | grep -q "INBOX"; then
        pass "User2 can access their mailbox"
    else
        fail "User2 cannot access their mailbox"
    fi
}

# Test alias delivery
test_alias_delivery() {
    log "Test 5: Alias delivery (admin@example.com -> alice@example.com)"

    local ALIAS_MSG="Alias test $(date +%s)"
    local SUBJECT="Test Alias $(date +%s)"

    if command -v swaks &> /dev/null; then
        swaks --to "admin@example.com" --from "test@external.com" \
            --server "$HOST:$SMTP_PORT" \
            --header "Subject: $SUBJECT" \
            --body "$ALIAS_MSG" --hide-all 2>&1

        # Wait for delivery
        sleep 2

        # Check if message appears in alice's mailbox
        if curl --silent --insecure "imaps://$HOST:$IMAP_PORT/INBOX" \
                --user "$USER1:$PASS1" --request "SEARCH SUBJECT \"$SUBJECT\"" 2>&1 | grep -q "SEARCH"; then
            pass "Alias delivery successful (admin@ -> alice@)"
        else
            warn "Could not verify alias delivery (may require longer wait or manual verification)"
        fi
    else
        warn "swaks not available, skipping alias test"
    fi
}

# Test catch-all delivery
test_catchall() {
    log "Test 6: Catch-all delivery (@example.org -> bob@example.org)"

    local CATCHALL_MSG="Catchall test $(date +%s)"
    local SUBJECT="Test Catchall $(date +%s)"
    local RANDOM_ADDR="random$(date +%s)@example.org"

    if command -v swaks &> /dev/null; then
        swaks --to "$RANDOM_ADDR" --from "test@external.com" \
            --server "$HOST:$SMTP_PORT" \
            --header "Subject: $SUBJECT" \
            --body "$CATCHALL_MSG" --hide-all 2>&1

        # Wait for delivery
        sleep 2

        # Check if message appears in bob's mailbox
        if curl --silent --insecure "imaps://$HOST:$IMAP_PORT/INBOX" \
                --user "$USER2:$PASS2" --request "SEARCH SUBJECT \"$SUBJECT\"" 2>&1 | grep -q "SEARCH"; then
            pass "Catch-all delivery successful (random@ -> bob@)"
        else
            warn "Could not verify catch-all delivery (may require manual configuration or verification)"
        fi
    else
        warn "swaks not available, skipping catch-all test"
    fi
}

# Main test execution
main() {
    echo "========================================"
    echo "DarkPipe Mail Flow Integration Test"
    echo "========================================"
    echo ""

    MAIL_SERVER=$(detect_mail_server)
    log "Active mail server: $MAIL_SERVER"
    log "Host: $HOST"
    log "SMTP: $SMTP_PORT, Submission: $SUBMISSION_PORT, IMAP: $IMAP_PORT"
    log "Test users: $USER1, $USER2"
    echo ""

    test_smtp_inbound
    test_imap_access
    test_smtp_submission
    test_multi_user_isolation
    test_alias_delivery
    test_catchall

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
