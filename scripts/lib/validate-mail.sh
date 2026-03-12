#!/usr/bin/env bash
# Mail infrastructure validation section for DarkPipe infrastructure validation.
# Validates: IP blocklist status, DKIM key consistency, transport map correctness,
#            outbound relay config presence, and Rspamd DKIM signing config.
#
# Designed to be sourced by validate-infrastructure.sh.
# Requires: RELAY_DOMAIN, DRY_RUN, VERBOSE (globals from parent script)
# Optional: RELAY_IP, DKIM_SELECTOR, HOME_PROFILE (for live checks)
#
# Usage (standalone):
#   bash scripts/lib/validate-mail.sh [--json] [--verbose] [--dry-run]
set -euo pipefail

SCRIPT_DIR_MAIL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT_MAIL="${SCRIPT_DIR_MAIL}/../.."

# DNSBLs to query for blocklist scanning
DNSBL_LISTS=("zen.spamhaus.org" "b.barracudacentral.org" "dnsbl.sorbs.net")

# External resolver for DNS queries
MAIL_DNS_RESOLVER="8.8.8.8"

# --- JSON helpers (local to this section) ---
_mail_check() {
  local name="$1" status="$2" detail="${3:-}" fix="${4:-}" severity="${5:-}"
  # Escape for valid JSON
  detail="${detail//\\/\\\\}"
  detail="${detail//\"/\\\"}"
  detail="${detail//$'\n'/\\n}"
  fix="${fix//\\/\\\\}"
  fix="${fix//\"/\\\"}"
  fix="${fix//$'\n'/\\n}"
  local json
  json="$(printf '{"name":"%s","status":"%s","detail":"%s","suggested_fix":"%s"' \
    "$name" "$status" "$detail" "$fix")"
  if [[ -n "$severity" ]]; then
    json+="$(printf ',"severity":"%s"' "$severity")"
  fi
  json+="}"
  printf '%s' "$json"
}

# --- Dry-run mock results ---
_mail_dry_run_checks() {
  local checks=""
  local names=("blocklist_zen_spamhaus_org" "blocklist_b_barracudacentral_org" "blocklist_dnsbl_sorbs_net" "dkim_dns_record" "transport_map" "outbound_relay" "rspamd_dkim_config")
  local descriptions=(
    "IP ${RELAY_IP:-203.0.113.1} clean on zen.spamhaus.org"
    "IP ${RELAY_IP:-203.0.113.1} clean on b.barracudacentral.org"
    "IP ${RELAY_IP:-203.0.113.1} clean on dnsbl.sorbs.net"
    "DKIM TXT record found for ${DKIM_SELECTOR:-darkpipe}._domainkey.${RELAY_DOMAIN:-example.com}"
    "Transport map has domain-specific entries, no wildcard"
    "Outbound relay config present"
    "Rspamd DKIM signing config present with required directives"
  )

  for i in "${!names[@]}"; do
    [[ -n "$checks" ]] && checks+=","
    checks+="$(_mail_check "${names[$i]}" "pass" "dry-run: ${descriptions[$i]}" "")"
  done
  echo "$checks"
}

# --- Individual checks (live) ---

# Reverse an IPv4 address for DNSBL lookup.
# 1.2.3.4 -> 4.3.2.1
_reverse_ip() {
  local ip="$1"
  local IFS='.'
  # shellcheck disable=SC2086
  set -- $ip
  echo "${4}.${3}.${2}.${1}"
}

# Check a single DNSBL for the relay IP.
# NXDOMAIN = clean (pass), any A record = listed (fail).
_check_blocklist() {
  local ip="$1" dnsbl="$2"
  local reversed
  reversed="$(_reverse_ip "$ip")"
  local query="${reversed}.${dnsbl}"
  local result
  result="$(dig +short +timeout=5 +tries=2 "@${MAIL_DNS_RESOLVER}" "$query" A 2>/dev/null || echo "")"

  # Sanitize DNSBL name for JSON field name (dots to underscores)
  local safe_name="blocklist_${dnsbl//\./_}"

  if [[ -z "$result" ]]; then
    _mail_check "$safe_name" "pass" "IP ${ip} not listed on ${dnsbl}" "" ""
  else
    _mail_check "$safe_name" "fail" "IP ${ip} listed on ${dnsbl} (response: ${result})" \
      "Check listing at https://${dnsbl} and request delisting if needed" "high"
  fi
}

# Verify DKIM DNS TXT record exists and has valid DKIM format.
_check_dkim_dns() {
  local domain="${1:-${RELAY_DOMAIN:-example.com}}"
  local selector="${DKIM_SELECTOR:-darkpipe}"
  local dkim_name="${selector}._domainkey.${domain}"

  local result
  result="$(dig +short +timeout=5 +tries=2 "@${MAIL_DNS_RESOLVER}" "$dkim_name" TXT 2>/dev/null || echo "")"

  if [[ -z "$result" ]]; then
    _mail_check "dkim_dns_record" "fail" "No DKIM TXT record found at ${dkim_name}" \
      "Publish DKIM public key as TXT record at ${dkim_name}" "high"
    return
  fi

  # Verify it contains DKIM key marker
  if ! echo "$result" | grep -qi 'v=DKIM1'; then
    _mail_check "dkim_dns_record" "fail" "TXT record at ${dkim_name} missing v=DKIM1 marker" \
      "Verify DKIM record content starts with v=DKIM1; k=rsa; p=..." "high"
    return
  fi

  # Verify it contains a public key
  if ! echo "$result" | grep -qi 'p='; then
    _mail_check "dkim_dns_record" "fail" "DKIM record at ${dkim_name} missing public key (p= tag)" \
      "Regenerate and publish DKIM key pair" "high"
    return
  fi

  local truncated
  truncated="$(echo "$result" | head -c 60)..."
  _mail_check "dkim_dns_record" "pass" "DKIM record found: ${dkim_name} -> ${truncated}" "" ""
}

# Check cloud relay transport map for:
# 1. No wildcard (*) entry (would cause routing loop)
# 2. At least one domain-specific entry
_check_transport_map() {
  local transport_file="${PROJECT_ROOT_MAIL}/cloud-relay/postfix-config/transport"

  if [[ ! -f "$transport_file" ]]; then
    _mail_check "transport_map" "fail" "Transport map not found at ${transport_file}" \
      "Create transport map file" "high"
    return
  fi

  # Check for wildcard entries (excluding comments)
  local wildcards
  wildcards="$(grep -v '^\s*#' "$transport_file" | grep -v '^\s*$' | grep '^\*' || true)"
  if [[ -n "$wildcards" ]]; then
    _mail_check "transport_map" "fail" \
      "Transport map contains wildcard entry (causes routing loop): ${wildcards}" \
      "Remove wildcard entry; use domain-specific routing only" "critical"
    return
  fi

  # Check for at least one domain-specific entry (non-comment, non-empty line)
  local domain_entries
  domain_entries="$(grep -v '^\s*#' "$transport_file" | grep -v '^\s*$' | grep -c '.' || true)"
  if [[ "$domain_entries" -eq 0 ]]; then
    _mail_check "transport_map" "fail" \
      "Transport map has no domain-specific entries" \
      "Add domain routing entry, e.g.: example.com smtp:[127.0.0.1]:10025" "high"
    return
  fi

  _mail_check "transport_map" "pass" \
    "Transport map OK: ${domain_entries} domain-specific entry(ies), no wildcard" "" ""
}

# Check outbound relay config exists in the active profile's config file.
# Supports: postfix-dovecot (relayhost), maddy (target.smtp), stalwart (queue.route)
_check_outbound_relay() {
  local profile="${HOME_PROFILE:-}"
  local found=false
  local detail=""

  # If no profile specified, check all three and pass if any has relay config
  if [[ -z "$profile" ]]; then
    # Postfix-Dovecot
    local postfix_main="${PROJECT_ROOT_MAIL}/home-device/postfix-dovecot/postfix/main.cf"
    if [[ -f "$postfix_main" ]] && grep -q '^relayhost' "$postfix_main"; then
      found=true
      detail="postfix-dovecot: relayhost configured in main.cf"
    fi

    # Maddy
    local maddy_conf="${PROJECT_ROOT_MAIL}/home-device/maddy/maddy.conf"
    if [[ -f "$maddy_conf" ]] && grep -q 'target\.smtp' "$maddy_conf"; then
      if [[ "$found" == true ]]; then
        detail+="; "
      fi
      found=true
      detail+="maddy: target.smtp relay configured"
    fi

    # Stalwart
    local stalwart_conf="${PROJECT_ROOT_MAIL}/home-device/stalwart/config.toml"
    if [[ -f "$stalwart_conf" ]] && grep -q 'queue\.route\.' "$stalwart_conf"; then
      if [[ "$found" == true ]]; then
        detail+="; "
      fi
      found=true
      detail+="stalwart: queue.route relay configured"
    fi

    if [[ "$found" == true ]]; then
      _mail_check "outbound_relay" "pass" "Outbound relay config found: ${detail}" "" ""
    else
      _mail_check "outbound_relay" "fail" \
        "No outbound relay config found in any mail server profile" \
        "Configure relayhost (Postfix), target.smtp (Maddy), or queue.route (Stalwart)" "high"
    fi
    return
  fi

  # Profile-specific check
  case "$profile" in
    postfix-dovecot|postfix)
      local conf="${PROJECT_ROOT_MAIL}/home-device/postfix-dovecot/postfix/main.cf"
      if [[ -f "$conf" ]] && grep -q '^relayhost' "$conf"; then
        _mail_check "outbound_relay" "pass" "postfix-dovecot: relayhost configured" "" ""
      else
        _mail_check "outbound_relay" "fail" \
          "postfix-dovecot: no relayhost found in ${conf}" \
          "Add relayhost = [10.8.0.1]:25 to main.cf" "high"
      fi
      ;;
    maddy)
      local conf="${PROJECT_ROOT_MAIL}/home-device/maddy/maddy.conf"
      if [[ -f "$conf" ]] && grep -q 'target\.smtp' "$conf"; then
        _mail_check "outbound_relay" "pass" "maddy: target.smtp relay configured" "" ""
      else
        _mail_check "outbound_relay" "fail" \
          "maddy: no target.smtp relay found in ${conf}" \
          "Add target.smtp block with targets tcp://10.8.0.1:25" "high"
      fi
      ;;
    stalwart)
      local conf="${PROJECT_ROOT_MAIL}/home-device/stalwart/config.toml"
      if [[ -f "$conf" ]] && grep -q 'queue\.route\.' "$conf"; then
        _mail_check "outbound_relay" "pass" "stalwart: queue.route relay configured" "" ""
      else
        _mail_check "outbound_relay" "fail" \
          "stalwart: no queue.route relay found in ${conf}" \
          "Add [queue.route.\"relay\"] with nexthop 10.8.0.1:25" "high"
      fi
      ;;
    *)
      _mail_check "outbound_relay" "skip" \
        "Unknown profile '${profile}' — cannot check outbound relay config" \
        "Set HOME_PROFILE to postfix-dovecot, maddy, or stalwart" ""
      ;;
  esac
}

# Verify Rspamd DKIM signing config exists and contains required directives.
_check_rspamd_dkim_config() {
  local config_file="${PROJECT_ROOT_MAIL}/home-device/spam-filter/rspamd/local.d/dkim_signing.conf"

  if [[ ! -f "$config_file" ]]; then
    _mail_check "rspamd_dkim_config" "fail" \
      "Rspamd DKIM signing config not found at ${config_file}" \
      "Create dkim_signing.conf with signing directives" "high"
    return
  fi

  # Check required directives
  local missing=""
  local required_directives=("sign_authenticated" "selector" "path" "use_domain")
  for directive in "${required_directives[@]}"; do
    if ! grep -q "^${directive}\s*=" "$config_file"; then
      if [[ -n "$missing" ]]; then
        missing+=", "
      fi
      missing+="$directive"
    fi
  done

  if [[ -n "$missing" ]]; then
    _mail_check "rspamd_dkim_config" "fail" \
      "Rspamd DKIM config missing directives: ${missing}" \
      "Add missing directives to ${config_file}" "medium"
    return
  fi

  _mail_check "rspamd_dkim_config" "pass" \
    "Rspamd DKIM signing config present with all required directives" "" ""
}

# --- Main entry point ---
# Outputs: comma-separated JSON check objects (the checks array content)
# Returns: 0 if all pass, 1 if any fail
run_mail_validation() {
  local domain="${RELAY_DOMAIN:-example.com}"

  # Dry-run: return mocks without network calls or file reads
  if [[ "${DRY_RUN:-false}" == "true" ]]; then
    _mail_dry_run_checks
    return 0
  fi

  # Resolve relay IP if not set
  local relay_ip="${RELAY_IP:-}"
  if [[ -z "$relay_ip" ]]; then
    relay_ip="$(dig +short +timeout=5 +tries=2 "@${MAIL_DNS_RESOLVER}" "${domain}" A 2>/dev/null | head -1 || echo "")"
    if [[ -z "$relay_ip" ]]; then
      # Can't do blocklist checks without an IP
      local skip_check
      skip_check="$(_mail_check "blocklist_scan" "skip" \
        "Cannot resolve ${domain} A record — skipping blocklist checks" \
        "Set RELAY_IP or ensure ${domain} has an A record" "")"
      # Continue with other checks
      local checks="$skip_check"
      # DKIM, transport, relay, rspamd checks
      local check
      check="$(_check_dkim_dns "$domain")"
      checks+=",${check}"
      check="$(_check_transport_map)"
      checks+=",${check}"
      check="$(_check_outbound_relay)"
      checks+=",${check}"
      check="$(_check_rspamd_dkim_config)"
      checks+=",${check}"
      echo "$checks"
      # Check for any failures
      if echo "$checks" | grep -q '"status":"fail"'; then
        return 1
      fi
      return 0
    fi
  fi

  # Live mode
  local checks=""
  local any_fail=false

  # 1. Blocklist scan — query each DNSBL
  for dnsbl in "${DNSBL_LISTS[@]}"; do
    local result
    result="$(_check_blocklist "$relay_ip" "$dnsbl")"
    [[ -n "$checks" ]] && checks+=","
    checks+="$result"
    if echo "$result" | grep -q '"status":"fail"'; then
      any_fail=true
    fi
  done

  # 2. DKIM DNS record check
  local result
  result="$(_check_dkim_dns "$domain")"
  checks+=",${result}"
  if echo "$result" | grep -q '"status":"fail"'; then
    any_fail=true
  fi

  # 3. Transport map check
  result="$(_check_transport_map)"
  checks+=",${result}"
  if echo "$result" | grep -q '"status":"fail"'; then
    any_fail=true
  fi

  # 4. Outbound relay config check
  result="$(_check_outbound_relay)"
  checks+=",${result}"
  if echo "$result" | grep -q '"status":"fail"'; then
    any_fail=true
  fi

  # 5. Rspamd DKIM signing config check
  result="$(_check_rspamd_dkim_config)"
  checks+=",${result}"
  if echo "$result" | grep -q '"status":"fail"'; then
    any_fail=true
  fi

  echo "$checks"

  if [[ "$any_fail" == "true" ]]; then
    return 1
  fi
  return 0
}

# --- Standalone execution support ---
# When run directly (not sourced), parse args and execute with formatted output.
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  # Set defaults for globals that the orchestrator normally provides
  RELAY_DOMAIN="${RELAY_DOMAIN:-example.com}"
  DRY_RUN="${DRY_RUN:-false}"
  VERBOSE="${VERBOSE:-false}"
  OUTPUT_JSON="${OUTPUT_JSON:-false}"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --json)    OUTPUT_JSON=true; shift ;;
      --verbose) VERBOSE=true; shift ;;
      --dry-run) DRY_RUN=true; shift ;;
      --help)
        echo "Usage: validate-mail.sh [--json] [--verbose] [--dry-run]"
        echo ""
        echo "Validate mail infrastructure: blocklist status, DKIM, transport map,"
        echo "outbound relay config, and Rspamd DKIM signing."
        echo ""
        echo "Environment Variables:"
        echo "  RELAY_DOMAIN     Primary mail domain (default: example.com)"
        echo "  RELAY_IP         Cloud relay public IPv4 (auto-detected if unset)"
        echo "  DKIM_SELECTOR    DKIM selector name (default: darkpipe)"
        echo "  HOME_PROFILE     Mail server profile: postfix-dovecot, maddy, stalwart"
        exit 0
        ;;
      *)
        echo "Error: Unknown option '$1'" >&2
        exit 2
        ;;
    esac
  done

  # Logging helper (standalone mode)
  log() {
    if [[ "$VERBOSE" == true ]]; then
      echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" >&2
    fi
  }

  log "Running mail validation (domain=${RELAY_DOMAIN}, dry_run=${DRY_RUN})"

  # Run validation
  exit_code=0
  checks_output="$(run_mail_validation)" || exit_code=$?
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  if [[ "$OUTPUT_JSON" == true ]]; then
    # JSON output: wrap checks in a section envelope
    json_status="pass"
    [[ $exit_code -ne 0 ]] && json_status="fail"
    printf '{"section":"mail","status":"%s","timestamp":"%s","checks":[%s]}\n' \
      "$json_status" "$timestamp" "$checks_output"
  else
    # Human-readable output
    echo ""
    echo "=== Mail Infrastructure Validation ==="
    [[ "$DRY_RUN" == true ]] && echo "  (dry-run mode)"
    echo ""

    # Parse and display each check from JSON
    echo "$checks_output" | grep -oE '"name":"[^"]*","status":"[^"]*","detail":"[^"]*"' | while IFS= read -r line; do
      c_name="$(echo "$line" | grep -oE '"name":"[^"]*"' | cut -d'"' -f4)"
      c_status="$(echo "$line" | grep -oE '"status":"[^"]*"' | cut -d'"' -f4)"
      c_detail="$(echo "$line" | grep -oE '"detail":"[^"]*"' | cut -d'"' -f4)"
      c_icon="✓"
      [[ "$c_status" == "fail" ]] && c_icon="✗"
      [[ "$c_status" == "skip" ]] && c_icon="○"
      printf "  %s %-40s %s\n" "$c_icon" "$c_name" "$c_status"
      if [[ "$VERBOSE" == true ]]; then
        echo "    → ${c_detail}"
      fi
    done

    echo ""
    if [[ $exit_code -eq 0 ]]; then
      echo "  ✓ All mail checks passed"
    else
      echo "  ✗ Some mail checks failed"
    fi
    echo ""
  fi

  exit "$exit_code"
fi
