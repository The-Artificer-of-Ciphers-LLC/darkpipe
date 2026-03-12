#!/usr/bin/env bash
# =============================================================================
# DarkPipe Device Connectivity Validation Script
# =============================================================================
#
# Pre-flight validation of all device-facing endpoints: autoconfig, autodiscover,
# profile server, webmail, monitoring, IMAP-TLS, and SMTP-STARTTLS. Run this
# before any human-in-the-loop device testing to catch broken endpoints early.
#
# Usage:
#   validate-device-connectivity.sh [OPTIONS]
#
# Options:
#   --json      Machine-readable JSON output (default: human-readable table)
#   --verbose   Emit timestamped diagnostic lines to stderr
#   --dry-run   Return mock pass results without contacting live infrastructure
#   --help      Print this help text and exit
#
# Environment Variables:
#   Name              Default          Description
#   ────────────────  ──────────────   ──────────────────────────────────────────
#   RELAY_DOMAIN      example.com      Primary mail domain to validate
#   HOME_DEVICE_IP    10.8.0.2         Home device IP on the WireGuard tunnel
#
# Prerequisites:
#   - bash 3.2+ (macOS default is fine)
#   - curl (HTTP endpoint checks)
#   - openssl (TLS handshake checks)
#   - jq (optional, for pretty-printing JSON output)
#
# Exit Codes:
#   0   All checks passed (or --dry-run)
#   1   One or more checks failed
#   2   Script error or invalid arguments
#
# Checks:
#   autoconfig            Mozilla autoconfig XML endpoint
#   autodiscover          Microsoft Autodiscover XML endpoint
#   profile-server-health Apple profile server health endpoint
#   webmail               Webmail UI (Roundcube or SnappyMail)
#   monitoring-dashboard  Monitoring status dashboard (HTML)
#   monitoring-json       Monitoring status JSON API
#   imap-tls              IMAP over TLS on port 993
#   smtp-starttls         SMTP STARTTLS on port 587
#
# Examples:
#   # Quick dry-run to verify script logic
#   ./scripts/validate-device-connectivity.sh --dry-run
#
#   # Full live check with JSON output piped to jq
#   RELAY_DOMAIN=darkpipe.email ./scripts/validate-device-connectivity.sh --json | jq .
#
#   # Verbose human-readable output
#   RELAY_DOMAIN=darkpipe.email ./scripts/validate-device-connectivity.sh --verbose
#
# =============================================================================
set -euo pipefail

# --- Configuration (from env vars with defaults) ---
RELAY_DOMAIN="${RELAY_DOMAIN:-example.com}"
HOME_DEVICE_IP="${HOME_DEVICE_IP:-10.8.0.2}"

# --- Globals ---
OUTPUT_JSON=false
VERBOSE=false
DRY_RUN=false

# Temp dir for check results (portable, bash 3 compatible)
RESULTS_DIR=""
TOTAL_PASS=0
TOTAL_FAIL=0
TOTAL_CHECKS=0

cleanup() {
  if [[ -n "$RESULTS_DIR" && -d "$RESULTS_DIR" ]]; then
    rm -rf "$RESULTS_DIR"
  fi
}
trap cleanup EXIT

# --- Argument Parsing ---
usage() {
  cat <<'EOF'
Usage: validate-device-connectivity.sh [OPTIONS]

Pre-flight validation of all DarkPipe device-facing endpoints.

Options:
  --json      Machine-readable JSON output (default: human-readable table)
  --verbose   Emit timestamped diagnostic lines to stderr
  --dry-run   Return mock pass results without contacting live infrastructure
  --help      Print this help text and exit

Environment Variables:
  Name              Default          Description
  ────────────────  ──────────────   ──────────────────────────────────────────
  RELAY_DOMAIN      example.com      Primary mail domain to validate
  HOME_DEVICE_IP    10.8.0.2         Home device IP on the WireGuard tunnel

Checks:
  autoconfig            Mozilla autoconfig XML endpoint
  autodiscover          Microsoft Autodiscover XML endpoint
  profile-server-health Apple profile server health endpoint
  webmail               Webmail UI (Roundcube or SnappyMail)
  monitoring-dashboard  Monitoring status dashboard (HTML)
  monitoring-json       Monitoring status JSON API
  imap-tls              IMAP over TLS on port 993
  smtp-starttls         SMTP STARTTLS on port 587

Exit Codes:
  0   All checks passed (or --dry-run)
  1   One or more checks failed
  2   Script error or invalid arguments

Examples:
  ./scripts/validate-device-connectivity.sh --dry-run
  RELAY_DOMAIN=darkpipe.email ./scripts/validate-device-connectivity.sh --json | jq .
  RELAY_DOMAIN=darkpipe.email ./scripts/validate-device-connectivity.sh --verbose
EOF
  exit 0
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --json)    OUTPUT_JSON=true; shift ;;
      --verbose) VERBOSE=true; shift ;;
      --dry-run) DRY_RUN=true; shift ;;
      --help)    usage ;;
      *)
        echo "Error: Unknown option '$1'" >&2
        echo "Run '$(basename "$0") --help' for usage." >&2
        exit 2
        ;;
    esac
  done
}

# --- Logging ---
log() {
  if [[ "$VERBOSE" == true ]]; then
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" >&2
  fi
}

log_info() {
  if [[ "$OUTPUT_JSON" == false ]]; then
    echo "$*" >&2
  fi
}

# --- JSON Helpers ---
# Escape a string for safe JSON embedding (handles quotes, backslashes, newlines)
json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  printf '%s' "$s"
}

# Record a single check result to a file
record_check() {
  local name="$1" status="$2" url="$3" detail="$4"
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
  if [[ "$status" == "pass" ]]; then
    TOTAL_PASS=$((TOTAL_PASS + 1))
  else
    TOTAL_FAIL=$((TOTAL_FAIL + 1))
  fi

  local escaped_detail
  escaped_detail="$(json_escape "$detail")"
  local escaped_url
  escaped_url="$(json_escape "$url")"

  printf '{"name":"%s","status":"%s","url":"%s","detail":"%s","timestamp":"%s"}' \
    "$name" "$status" "$escaped_url" "$escaped_detail" "$timestamp" \
    > "${RESULTS_DIR}/check-$(printf '%02d' "$TOTAL_CHECKS").json"
}

# --- Check Implementations ---

check_autoconfig() {
  local url="https://autoconfig.${RELAY_DOMAIN}/.well-known/autoconfig/mail/config-v1.1.xml"
  local name="autoconfig"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock XML with <emailProvider> present"
    return
  fi

  log "Checking ${name}: GET ${url}"
  local response_body=""
  local response_code=""
  response_body="$(curl -sSL --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || true
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" != "200" ]]; then
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
    return
  fi

  if echo "$response_body" | grep -qi '<emailProvider'; then
    record_check "$name" "pass" "$url" "HTTP 200 with <emailProvider> in XML response"
  else
    record_check "$name" "fail" "$url" "HTTP 200 but <emailProvider> not found in response body"
  fi
}

check_autodiscover() {
  local url="https://autodiscover.${RELAY_DOMAIN}/autodiscover/autodiscover.xml"
  local name="autodiscover"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock XML with <Protocol> present"
    return
  fi

  log "Checking ${name}: POST ${url}"
  local autodiscover_body='<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
  <Request>
    <EMailAddress>user@'"${RELAY_DOMAIN}"'</EMailAddress>
    <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
  </Request>
</Autodiscover>'

  local response_body=""
  local response_code=""
  response_body="$(curl -sSL --connect-timeout 10 --max-time 30 \
    -X POST \
    -H 'Content-Type: text/xml' \
    -d "$autodiscover_body" \
    "$url" 2>/dev/null)" || true
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 \
    -X POST \
    -H 'Content-Type: text/xml' \
    -d "$autodiscover_body" \
    "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" != "200" ]]; then
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
    return
  fi

  if echo "$response_body" | grep -qi '<Protocol'; then
    record_check "$name" "pass" "$url" "HTTP 200 with <Protocol> in XML response"
  else
    record_check "$name" "fail" "$url" "HTTP 200 but <Protocol> not found in response body"
  fi
}

check_profile_server_health() {
  local url="https://mail.${RELAY_DOMAIN}/health/live"
  local name="profile-server-health"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock 200 OK health response"
    return
  fi

  log "Checking ${name}: GET ${url}"
  local response_code=""
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" == "200" ]]; then
    record_check "$name" "pass" "$url" "HTTP 200"
  else
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
  fi
}

check_webmail() {
  local url="https://mail.${RELAY_DOMAIN}/"
  local name="webmail"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock HTML with webmail indicator"
    return
  fi

  log "Checking ${name}: GET ${url}"
  local response_body=""
  local response_code=""
  response_body="$(curl -sSL --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || true
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" != "200" ]]; then
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
    return
  fi

  if echo "$response_body" | grep -qi 'roundcube\|rcmlogin\|snappymail\|rainloop'; then
    record_check "$name" "pass" "$url" "HTTP 200 with webmail indicator in HTML"
  else
    # Still pass if we get HTML — the webmail may have changed branding
    if echo "$response_body" | grep -qi '<html'; then
      record_check "$name" "pass" "$url" "HTTP 200 with HTML content (no specific webmail indicator matched)"
    else
      record_check "$name" "fail" "$url" "HTTP 200 but response does not appear to be HTML"
    fi
  fi
}

check_monitoring_dashboard() {
  local url="https://mail.${RELAY_DOMAIN}/status"
  local name="monitoring-dashboard"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock HTML dashboard"
    return
  fi

  log "Checking ${name}: GET ${url}"
  local response_body=""
  local response_code=""
  response_body="$(curl -sSL --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || true
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" != "200" ]]; then
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
    return
  fi

  if echo "$response_body" | grep -qi '<html\|<!doctype'; then
    record_check "$name" "pass" "$url" "HTTP 200 with HTML content"
  else
    record_check "$name" "fail" "$url" "HTTP 200 but response does not appear to be HTML"
  fi
}

check_monitoring_json() {
  local url="https://mail.${RELAY_DOMAIN}/status?format=json"
  local name="monitoring-json"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock JSON with health status"
    return
  fi

  log "Checking ${name}: GET ${url}"
  local response_body=""
  local response_code=""
  response_body="$(curl -sSL --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || true
  response_code="$(curl -sSL -o /dev/null -w '%{http_code}' --connect-timeout 10 --max-time 30 "$url" 2>/dev/null)" || response_code="000"

  if [[ "$response_code" != "200" ]]; then
    record_check "$name" "fail" "$url" "HTTP ${response_code} (expected 200)"
    return
  fi

  # Check for JSON with health/status indicator
  if echo "$response_body" | grep -qi '"health"\|"status"\|"overall"\|"healthy"'; then
    record_check "$name" "pass" "$url" "HTTP 200 with JSON health data"
  else
    # If it starts with { or [, it's probably JSON even without a known key
    if echo "$response_body" | grep -q '^[[:space:]]*[\[{]'; then
      record_check "$name" "pass" "$url" "HTTP 200 with JSON response (no standard health key matched)"
    else
      record_check "$name" "fail" "$url" "HTTP 200 but response does not appear to be JSON"
    fi
  fi
}

check_imap_tls() {
  local host="mail.${RELAY_DOMAIN}"
  local port=993
  local url="imaps://${host}:${port}"
  local name="imap-tls"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock TLS handshake with IMAP banner"
    return
  fi

  log "Checking ${name}: openssl s_client ${host}:${port}"
  local tls_output=""
  tls_output="$(echo "" | openssl s_client -connect "${host}:${port}" -servername "$host" 2>&1)" || true

  # Check for successful TLS handshake — require explicit CONNECTED indicator
  # Note: grep for bare 'OK' would false-positive on error strings like 'BIO_lookup_ex'
  if echo "$tls_output" | grep -q 'CONNECTED('; then
    # Check for IMAP banner
    if echo "$tls_output" | grep -q '^\* OK\|IMAP'; then
      record_check "$name" "pass" "$url" "TLS handshake successful with IMAP banner"
    else
      # TLS connected but no IMAP banner visible (may have been consumed)
      record_check "$name" "pass" "$url" "TLS handshake successful (IMAP service on port ${port})"
    fi
  else
    local err_detail
    err_detail="$(echo "$tls_output" | grep -i 'error\|errno\|verify return code' | head -3)"
    record_check "$name" "fail" "$url" "TLS handshake failed: ${err_detail}"
  fi
}

check_smtp_starttls() {
  local host="mail.${RELAY_DOMAIN}"
  local port=587
  local url="smtp://${host}:${port}"
  local name="smtp-starttls"

  if [[ "$DRY_RUN" == true ]]; then
    log "DRY-RUN: mock pass for ${name}"
    record_check "$name" "pass" "$url" "dry-run: mock STARTTLS handshake successful"
    return
  fi

  log "Checking ${name}: openssl s_client -starttls smtp ${host}:${port}"
  local tls_output=""
  tls_output="$(echo "" | openssl s_client -starttls smtp -connect "${host}:${port}" -servername "$host" 2>&1)" || true

  # Check for successful TLS handshake — require explicit CONNECTED indicator
  # Note: grep for bare 'OK' would false-positive on error strings like 'BIO_lookup_ex'
  if echo "$tls_output" | grep -q 'CONNECTED('; then
    if echo "$tls_output" | grep -qi '250 \|220 '; then
      record_check "$name" "pass" "$url" "STARTTLS handshake successful with SMTP banner"
    else
      record_check "$name" "pass" "$url" "STARTTLS handshake successful"
    fi
  else
    local err_detail
    err_detail="$(echo "$tls_output" | grep -i 'error\|errno\|verify return code' | head -3)"
    record_check "$name" "fail" "$url" "STARTTLS handshake failed: ${err_detail}"
  fi
}

# --- Output ---
emit_json() {
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  local overall_status="pass"
  if [[ "$TOTAL_FAIL" -gt 0 ]]; then
    overall_status="fail"
  fi

  # Build checks array from result files
  local checks_json=""
  local f
  for f in "${RESULTS_DIR}"/check-*.json; do
    [[ -f "$f" ]] || continue
    if [[ -n "$checks_json" ]]; then
      checks_json+=","
    fi
    checks_json+="$(cat "$f")"
  done

  printf '{"overall_status":"%s","timestamp":"%s","config":{"relay_domain":"%s","home_device_ip":"%s","dry_run":%s},"total_checks":%d,"passed":%d,"failed":%d,"checks":[%s]}\n' \
    "$overall_status" "$timestamp" "$RELAY_DOMAIN" "$HOME_DEVICE_IP" "$DRY_RUN" \
    "$TOTAL_CHECKS" "$TOTAL_PASS" "$TOTAL_FAIL" "$checks_json"
}

emit_human() {
  echo ""
  echo "=== DarkPipe Device Connectivity Validation ==="
  echo ""
  echo "  Domain:      ${RELAY_DOMAIN}"
  echo "  Home Device: ${HOME_DEVICE_IP}"
  [[ "$DRY_RUN" == true ]] && echo "  Mode:        dry-run (no live checks)"
  echo ""

  local f
  for f in "${RESULTS_DIR}"/check-*.json; do
    [[ -f "$f" ]] || continue
    local content
    content="$(cat "$f")"
    local name status detail url
    name="$(echo "$content" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)"
    status="$(echo "$content" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)"
    detail="$(echo "$content" | grep -o '"detail":"[^"]*"' | cut -d'"' -f4)"
    url="$(echo "$content" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)"

    local icon="✓"
    local color="\033[0;32m"
    if [[ "$status" == "fail" ]]; then
      icon="✗"
      color="\033[0;31m"
    fi
    local nc="\033[0m"

    printf "  ${color}%s${nc} %-25s %s\n" "$icon" "$name" "$status"
    if [[ "$VERBOSE" == true || "$status" == "fail" ]]; then
      printf "    URL:    %s\n" "$url"
      printf "    Detail: %s\n" "$detail"
    fi
  done

  echo ""
  echo "  ────────────────────────────────────────"
  local overall_icon="✓"
  local overall_color="\033[0;32m"
  if [[ "$TOTAL_FAIL" -gt 0 ]]; then
    overall_icon="✗"
    overall_color="\033[0;31m"
  fi
  local nc="\033[0m"
  printf "  ${overall_color}%s${nc} Results: %d passed, %d failed out of %d checks\n" \
    "$overall_icon" "$TOTAL_PASS" "$TOTAL_FAIL" "$TOTAL_CHECKS"
  echo ""
}

# --- Main ---
main() {
  parse_args "$@"

  RESULTS_DIR="$(mktemp -d)"

  log_info "DarkPipe Device Connectivity Validation"
  [[ "$DRY_RUN" == true ]] && log_info "  (dry-run mode — no live checks)"
  log "Config: RELAY_DOMAIN=${RELAY_DOMAIN} HOME_DEVICE_IP=${HOME_DEVICE_IP}"

  # Run all 8 endpoint checks
  check_autoconfig
  check_autodiscover
  check_profile_server_health
  check_webmail
  check_monitoring_dashboard
  check_monitoring_json
  check_imap_tls
  check_smtp_starttls

  if [[ "$OUTPUT_JSON" == true ]]; then
    emit_json
  else
    emit_human
  fi

  if [[ "$TOTAL_FAIL" -gt 0 ]]; then
    exit 1
  fi
  exit 0
}

main "$@"
