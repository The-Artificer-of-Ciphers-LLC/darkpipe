#!/usr/bin/env bash
# DarkPipe Mail Round-Trip Test Helper
# Automates what can be automated for end-to-end mail testing through the
# full DarkPipe chain, and provides clear instructions for human-in-the-loop
# verification steps.
#
# Outbound: home server → WireGuard tunnel → cloud relay → external recipient
# Inbound:  external sender → cloud relay → WireGuard tunnel → home mailbox
#
# Prerequisites:
#   - Working WireGuard tunnel (validated by validate-infrastructure.sh)
#   - DNS records configured (MX, SPF, DKIM, DMARC)
#   - TLS certificates provisioned
#   - Mail server profile running on home device
#   - swaks installed (apt install swaks / brew install swaks)
#
# Usage:
#   ./scripts/test-mail-roundtrip.sh --domain example.com \
#       --recipient test@gmail.com --sender alice [OPTIONS]
#
# Options:
#   --domain DOMAIN       Mail domain under test (required)
#   --recipient EMAIL     External recipient address for outbound test (required)
#   --sender USER         Local username (without @domain) for sending (required)
#   --password PASS       SMTP submission password (prompted if omitted)
#   --host HOST           SMTP submission host (default: localhost)
#   --port PORT           SMTP submission port (default: 587)
#   --imap-host HOST      IMAP host for inbound check (default: localhost)
#   --imap-port PORT      IMAP port for inbound check (default: 993)
#   --timeout SECS        Log polling timeout in seconds (default: 60)
#   --log-file PATH       Mail log path to tail (default: /var/log/mail.log)
#   --verbose             Show detailed output during execution
#   --dry-run             Print test sequence without sending mail
#   --help                Show this help
#
# Exit Codes:
#   0   Test completed (or dry-run printed successfully)
#   1   Send failure or delivery timeout
#   2   Missing prerequisites or invalid arguments
#
# See also:
#   docs/validation/mail-roundtrip.md    Full round-trip testing procedure
#   scripts/lib/validate-mail.sh         Pre-flight infrastructure checks
set -euo pipefail

# --- Defaults ---
DOMAIN=""
RECIPIENT=""
SENDER=""
PASSWORD=""
HOST="localhost"
PORT="587"
IMAP_HOST="localhost"
IMAP_PORT="993"
TIMEOUT=60
LOG_FILE="/var/log/mail.log"
VERBOSE=0
DRY_RUN=0

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# --- Logging ---
timestamp() { date -u +%Y-%m-%dT%H:%M:%SZ; }

log_event() {
  local event="$1" detail="${2:-}"
  printf '%s [%s] %s\n' "$(timestamp)" "$event" "$detail"
}

log_info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$*"; }
log_warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*"; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$*"; }
log_step()  { printf "\n${BOLD}${CYAN}▸ %s${NC}\n" "$*"; }
log_dry()   { printf "${DIM}[DRY-RUN]${NC} %s\n" "$*"; }

verbose() {
  if [[ $VERBOSE -eq 1 ]]; then
    printf "${DIM}  %s${NC}\n" "$*"
  fi
}

# --- Argument Parsing ---
usage() {
  sed -n '/^# Usage:/,/^# See also:/p' "$0" | sed 's/^# \?//'
  exit 0
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --domain)     DOMAIN="$2"; shift 2 ;;
      --recipient)  RECIPIENT="$2"; shift 2 ;;
      --sender)     SENDER="$2"; shift 2 ;;
      --password)   PASSWORD="$2"; shift 2 ;;
      --host)       HOST="$2"; shift 2 ;;
      --port)       PORT="$2"; shift 2 ;;
      --imap-host)  IMAP_HOST="$2"; shift 2 ;;
      --imap-port)  IMAP_PORT="$2"; shift 2 ;;
      --timeout)    TIMEOUT="$2"; shift 2 ;;
      --log-file)   LOG_FILE="$2"; shift 2 ;;
      --verbose)    VERBOSE=1; shift ;;
      --dry-run)    DRY_RUN=1; shift ;;
      --help)       usage ;;
      *)            log_error "Unknown option: $1"; exit 2 ;;
    esac
  done

  if [[ -z "$DOMAIN" || -z "$RECIPIENT" || -z "$SENDER" ]]; then
    log_error "Required: --domain, --recipient, --sender"
    echo "  Run with --help for usage."
    exit 2
  fi
}

# --- Prerequisites ---
check_prerequisites() {
  local missing=0

  if ! command -v swaks &>/dev/null; then
    log_warn "swaks not found — install it: apt install swaks / brew install swaks"
    missing=1
  fi

  if [[ $DRY_RUN -eq 0 && $missing -eq 1 ]]; then
    log_error "Missing prerequisites. Install them or use --dry-run to see the test plan."
    exit 2
  fi
}

# --- Unique test identifiers ---
generate_test_id() {
  local ts
  ts="$(date +%s)"
  local rand
  rand="$(head -c 4 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  echo "darkpipe-rt-${ts}-${rand}"
}

# --- Step 1: Pre-flight checks ---
run_preflight() {
  log_step "Step 1: Pre-flight infrastructure checks"

  local validate_script
  validate_script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/validate-mail.sh"

  if [[ $DRY_RUN -eq 1 ]]; then
    log_dry "Would run: RELAY_DOMAIN=${DOMAIN} bash ${validate_script} --dry-run"
    log_dry "Checks: blocklist scan, DKIM DNS, transport map, relay config, Rspamd DKIM"
    return 0
  fi

  if [[ -f "$validate_script" ]]; then
    log_info "Running mail infrastructure validation..."
    verbose "RELAY_DOMAIN=${DOMAIN} bash ${validate_script}"
    if RELAY_DOMAIN="$DOMAIN" bash "$validate_script" --verbose; then
      log_info "Pre-flight checks passed"
    else
      log_warn "Some pre-flight checks failed — outbound mail may not pass authentication"
      log_warn "Review with: RELAY_DOMAIN=${DOMAIN} bash ${validate_script} --verbose"
    fi
  else
    log_warn "Mail validation script not found at ${validate_script} — skipping pre-flight"
  fi
}

# --- Step 2: Send outbound test email ---
send_outbound_test() {
  local test_id="$1"
  local sender_addr="${SENDER}@${DOMAIN}"
  local subject="DarkPipe Round-Trip Test [${test_id}]"
  local body
  body="$(cat <<EOF
This is an automated round-trip test email from DarkPipe.

Test ID: ${test_id}
Timestamp: $(timestamp)
Sender: ${sender_addr}
Recipient: ${RECIPIENT}
Domain: ${DOMAIN}

If you received this email, outbound delivery is working.
Check the Authentication-Results header for SPF/DKIM/DMARC status.
EOF
)"

  log_step "Step 2: Send outbound test email"
  log_info "From:    ${sender_addr}"
  log_info "To:      ${RECIPIENT}"
  log_info "Subject: ${subject}"
  log_info "Test ID: ${test_id}"

  if [[ $DRY_RUN -eq 1 ]]; then
    log_dry "Would send via SMTP submission:"
    log_dry "  swaks --to ${RECIPIENT} --from ${sender_addr} \\"
    log_dry "    --server ${HOST}:${PORT} --tls \\"
    log_dry "    --auth-user ${sender_addr} --auth-password *** \\"
    log_dry "    --header 'Subject: ${subject}' \\"
    log_dry "    --header 'X-DarkPipe-Test-ID: ${test_id}' \\"
    log_dry "    --body '(test message body with test ID and timestamp)'"
    return 0
  fi

  # Prompt for password if not provided
  if [[ -z "$PASSWORD" ]]; then
    read -rsp "SMTP password for ${sender_addr}: " PASSWORD
    echo
  fi

  log_event "SEND_START" "recipient=${RECIPIENT} test_id=${test_id}"

  local swaks_output
  local swaks_exit=0
  swaks_output="$(swaks \
    --to "$RECIPIENT" \
    --from "$sender_addr" \
    --server "${HOST}:${PORT}" \
    --tls \
    --auth-user "$sender_addr" \
    --auth-password "$PASSWORD" \
    --header "Subject: ${subject}" \
    --header "X-DarkPipe-Test-ID: ${test_id}" \
    --body "$body" \
    2>&1)" || swaks_exit=$?

  if [[ $swaks_exit -ne 0 ]]; then
    log_error "swaks exited with code ${swaks_exit}"
    echo "$swaks_output" >&2
    log_event "SEND_FAIL" "exit_code=${swaks_exit}"
    return 1
  fi

  log_info "Message accepted by submission server"
  log_event "SEND_OK" "recipient=${RECIPIENT}"

  if [[ $VERBOSE -eq 1 ]]; then
    echo "$swaks_output"
  fi
}

# --- Step 3: Poll logs for delivery confirmation ---
poll_delivery_logs() {
  local test_id="$1"

  log_step "Step 3: Poll mail logs for delivery confirmation"

  if [[ $DRY_RUN -eq 1 ]]; then
    log_dry "Would tail ${LOG_FILE} for up to ${TIMEOUT}s"
    log_dry "Looking for: status=sent (success), status=bounced/deferred (failure)"
    log_dry "Grep pattern: '${test_id}' or 'status=' in log lines"
    return 0
  fi

  if [[ ! -r "$LOG_FILE" ]]; then
    log_warn "Cannot read ${LOG_FILE} — you may need sudo or a different --log-file path"
    log_info "Check delivery manually:"
    echo "  sudo grep '${SENDER}' ${LOG_FILE} | tail -20"
    echo "  sudo grep 'status=' ${LOG_FILE} | tail -20"
    return 0
  fi

  log_info "Tailing ${LOG_FILE} for ${TIMEOUT}s (looking for delivery status)..."
  log_event "POLL_START" "log=${LOG_FILE} timeout=${TIMEOUT}s"

  local deadline
  deadline=$(( $(date +%s) + TIMEOUT ))
  local found=0

  # Use tail -f with a timeout loop
  while [[ $(date +%s) -lt $deadline ]]; do
    local recent
    recent="$(tail -20 "$LOG_FILE" 2>/dev/null | grep -i "${SENDER}\|${RECIPIENT}\|status=" || true)"

    if echo "$recent" | grep -qi "status=sent"; then
      log_info "Delivery confirmed: status=sent"
      log_event "DELIVERY_OK" "found status=sent in logs"
      found=1
      break
    elif echo "$recent" | grep -qi "status=bounced"; then
      log_error "Message bounced!"
      echo "$recent" | grep -i "status=bounced"
      log_event "DELIVERY_BOUNCE" "found status=bounced in logs"
      return 1
    elif echo "$recent" | grep -qi "status=deferred"; then
      log_warn "Message deferred (will retry) — check logs for reason"
      echo "$recent" | grep -i "status=deferred"
      verbose "Deferred messages will be retried automatically by Postfix"
    fi

    sleep 3
  done

  if [[ $found -eq 0 ]]; then
    log_warn "Delivery not confirmed within ${TIMEOUT}s"
    log_info "This may be normal — check manually:"
    echo "  sudo grep '${RECIPIENT}' ${LOG_FILE} | tail -20"
    echo "  postqueue -p    # check for queued messages"
    log_event "POLL_TIMEOUT" "no status=sent found within ${TIMEOUT}s"
  fi
}

# --- Step 4: Header verification instructions ---
print_header_instructions() {
  log_step "Step 4: Verify authentication headers (manual)"
  echo ""
  printf "${BOLD}Check the recipient's inbox (%s) for the test email.${NC}\n" "$RECIPIENT"
  echo ""
  echo "Open the message and view the full headers (raw source). Look for the"
  printf "${BOLD}Authentication-Results${NC} header. Here is what each result means:\n"
  echo ""
  printf "  ${GREEN}spf=pass${NC}      SPF record authorized the sending IP\n"
  printf "  ${GREEN}dkim=pass${NC}     DKIM signature verified against DNS public key\n"
  printf "  ${GREEN}dmarc=pass${NC}    DMARC policy evaluation passed (requires SPF + DKIM alignment)\n"
  echo ""
  printf "${BOLD}Example of a fully passing Authentication-Results header:${NC}\n"
  echo ""
  echo "  Authentication-Results: mx.google.com;"
  printf "    dkim=pass header.i=@%s header.s=darkpipe;\n" "$DOMAIN"
  printf "    spf=pass (google.com: domain of %s@%s designates <relay-ip> as permitted sender);\n" "$SENDER" "$DOMAIN"
  printf "    dmarc=pass (p=REJECT sp=REJECT) header.from=%s\n" "$DOMAIN"
  echo ""
  printf "${BOLD}How to view headers in common providers:${NC}\n"
  echo "  Gmail:     Open message → ⋮ menu → \"Show original\""
  echo "  Outlook:   Open message → ⋯ menu → \"View message source\""
  echo "  Apple Mail: View → Message → All Headers"
  echo ""
  printf "${RED}If any result shows \"fail\" or \"softfail\":${NC}\n"
  echo "  spf=fail     → Relay IP not in SPF record, or SPF record missing"
  echo "  dkim=fail    → DKIM signature mismatch (key rotation? selector wrong?)"
  echo "  dmarc=fail   → Alignment failure (From domain vs SPF/DKIM domain mismatch)"
  echo ""
  printf "  Run pre-flight checks: RELAY_DOMAIN=%s bash scripts/lib/validate-mail.sh --verbose\n" "$DOMAIN"
  echo "  See: docs/validation/mail-roundtrip.md for troubleshooting"
  echo ""
}

# --- Step 5: Inbound test instructions and polling ---
run_inbound_test() {
  local sender_addr="${SENDER}@${DOMAIN}"

  log_step "Step 5: Inbound delivery test"
  echo ""
  printf "${BOLD}To test inbound delivery, send an email from an external account to:${NC}\n"
  echo ""
  printf "  ${CYAN}%s${NC}\n" "$sender_addr"
  echo ""
  echo "You can:"
  printf "  1. Reply to the outbound test email from %s\n" "$RECIPIENT"
  printf "  2. Or compose a new message to %s\n" "$sender_addr"
  echo ""
  printf "Include a recognizable subject (e.g., \"Inbound Test %s\")\n" "$(date +%Y%m%d)"
  echo "so it's easy to find in the mailbox."
  echo ""

  if [[ $DRY_RUN -eq 1 ]]; then
    log_dry "Would then poll for inbound arrival via IMAP or Maildir"
    log_dry "  IMAP check: curl --insecure imaps://${IMAP_HOST}:${IMAP_PORT}/INBOX"
    log_dry "    --user '${sender_addr}:***' --request 'SEARCH UNSEEN'"
    log_dry "  Or Maildir: ls -lt ~/Maildir/new/ | head -5"
    log_dry "Would poll for up to ${TIMEOUT}s"
    return 0
  fi

  read -rp "Press Enter once you've sent the inbound test email (or 'skip' to skip): " response
  if [[ "$response" == "skip" ]]; then
    log_info "Skipping inbound delivery check"
    return 0
  fi

  log_info "Polling for inbound delivery..."
  log_event "INBOUND_POLL_START" "mailbox=${sender_addr}"

  local deadline
  deadline=$(( $(date +%s) + TIMEOUT ))
  local found=0

  while [[ $(date +%s) -lt $deadline ]]; do
    # Try IMAP SEARCH for unseen messages
    local imap_result
    imap_result="$(curl --silent --insecure \
      "imaps://${IMAP_HOST}:${IMAP_PORT}/INBOX" \
      --user "${sender_addr}:${PASSWORD}" \
      --request "SEARCH UNSEEN" 2>&1 || true)"

    if echo "$imap_result" | grep -q "SEARCH [0-9]"; then
      log_info "New message(s) found in INBOX!"
      log_event "INBOUND_OK" "unseen messages found via IMAP"
      found=1
      break
    fi

    verbose "No new messages yet, polling..."
    sleep 5
  done

  if [[ $found -eq 0 ]]; then
    log_warn "No new messages detected within ${TIMEOUT}s"
    log_info "Check manually:"
    echo "  curl --insecure 'imaps://${IMAP_HOST}:${IMAP_PORT}/INBOX' \\"
    echo "    --user '${sender_addr}:<password>' --request 'SEARCH ALL'"
    echo ""
    echo "  # Or check Postfix logs on cloud relay for inbound routing:"
    echo "  sudo grep '${DOMAIN}' /var/log/mail.log | tail -20"
    log_event "INBOUND_TIMEOUT" "no unseen messages within ${TIMEOUT}s"
  fi
}

# --- Step 6: Summary ---
print_summary() {
  local test_id="$1"

  log_step "Test Summary"
  echo ""
  printf "Test ID:  %s\n" "$test_id"
  printf "Domain:   %s\n" "$DOMAIN"
  printf "Outbound: %s@%s → %s\n" "$SENDER" "$DOMAIN" "$RECIPIENT"
  printf "Inbound:  %s → %s@%s\n" "$RECIPIENT" "$SENDER" "$DOMAIN"
  echo ""
  printf "${BOLD}Checklist:${NC}\n"
  printf "  □ Outbound email received by %s\n" "$RECIPIENT"
  echo "  □ Authentication-Results: spf=pass, dkim=pass, dmarc=pass"
  echo "  □ Email NOT in spam/junk folder"
  printf "  □ Inbound email arrived in %s@%s mailbox\n" "$SENDER" "$DOMAIN"
  echo "  □ No relay denied errors in mail logs"
  echo ""
  printf "${BOLD}If something failed, see:${NC}\n"
  echo "  docs/validation/mail-roundtrip.md          Troubleshooting guide"
  echo "  scripts/lib/validate-mail.sh --verbose     Pre-flight diagnostics"
  echo ""
  log_event "TEST_COMPLETE" "test_id=${test_id}"
}

# --- Main ---
main() {
  parse_args "$@"
  check_prerequisites

  local test_id
  test_id="$(generate_test_id)"

  echo ""
  echo "╔══════════════════════════════════════════════╗"
  echo "║   DarkPipe Mail Round-Trip Test              ║"
  echo "╚══════════════════════════════════════════════╝"
  echo ""
  log_info "Test ID: ${test_id}"
  log_info "Domain:  ${DOMAIN}"
  [[ $DRY_RUN -eq 1 ]] && log_info "Mode:    DRY-RUN (no mail will be sent)"
  echo ""

  run_preflight
  send_outbound_test "$test_id"
  poll_delivery_logs "$test_id"
  print_header_instructions
  run_inbound_test
  print_summary "$test_id"

  return 0
}

main "$@"
