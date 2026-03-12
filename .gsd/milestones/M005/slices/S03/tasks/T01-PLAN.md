---
estimated_steps: 5
estimated_files: 1
---

# T01: Build device connectivity validation script

**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

## Description

Create `scripts/validate-device-connectivity.sh` — a validation orchestrator for all device-facing endpoints. Follows the same patterns as `scripts/validate-infrastructure.sh` (--json, --verbose, --dry-run modes). This script runs automated pre-flight checks before any human-in-the-loop device testing, catching broken endpoints early.

## Steps

1. Read `scripts/validate-infrastructure.sh` to extract the structural patterns: argument parsing, color/logging helpers, section runner, JSON output formatting, dry-run mock responses, and exit code handling.
2. Create `scripts/validate-device-connectivity.sh` with the same option interface (--json, --verbose, --dry-run, --help) and environment variables (RELAY_DOMAIN, HOME_DEVICE_IP).
3. Implement endpoint checks:
   - **autoconfig**: GET `https://autoconfig.{RELAY_DOMAIN}/.well-known/autoconfig/mail/config-v1.1.xml` — expect 200 with XML content containing `<emailProvider>`
   - **autodiscover**: POST `https://autodiscover.{RELAY_DOMAIN}/autodiscover/autodiscover.xml` — expect 200 with XML content containing `<Protocol>`
   - **profile-server-health**: GET `https://mail.{RELAY_DOMAIN}/health/live` — expect 200
   - **webmail**: GET `https://mail.{RELAY_DOMAIN}/` — expect 200 with HTML content (roundcube or snappymail indicator)
   - **monitoring-dashboard**: GET `https://mail.{RELAY_DOMAIN}/status` — expect 200 with HTML content
   - **monitoring-json**: GET `https://mail.{RELAY_DOMAIN}/status?format=json` — expect 200 with JSON containing health status
   - **imap-tls**: openssl s_client to `mail.{RELAY_DOMAIN}:993` — expect successful TLS handshake with IMAP banner
   - **smtp-starttls**: openssl s_client -starttls smtp to `mail.{RELAY_DOMAIN}:587` — expect successful STARTTLS handshake
4. Add summary output (pass/fail counts, table of results) matching validate-infrastructure.sh format. Ensure JSON mode outputs a structured array of check results.
5. Run shellcheck and verify --dry-run exits 0.

## Must-Haves

- [ ] --json, --verbose, --dry-run, --help flags matching validate-infrastructure.sh patterns
- [ ] All 8 endpoint checks implemented with clear pass/fail and error details
- [ ] Dry-run mode returns mock pass results without contacting live infrastructure
- [ ] Exit code 0 on all pass, 1 on any failure, 2 on script error
- [ ] shellcheck clean

## Verification

- `bash scripts/validate-device-connectivity.sh --dry-run` exits 0
- `bash scripts/validate-device-connectivity.sh --dry-run --json | jq .` produces valid JSON
- `shellcheck scripts/validate-device-connectivity.sh` passes (or only informational notes)

## Observability Impact

- Signals added/changed: new structured pass/fail output per endpoint check with timestamps
- How a future agent inspects this: `scripts/validate-device-connectivity.sh --json` for machine-readable results; --verbose for timestamped diagnostics on stderr
- Failure state exposed: each check reports URL tested, HTTP status code, expected vs actual content match, and specific curl/openssl error message

## Inputs

- `scripts/validate-infrastructure.sh` — structural patterns (arg parsing, logging, section runner, JSON output, dry-run)
- `scripts/test-mail-roundtrip.sh` — human-in-the-loop test pattern reference
- `home-device/tests/test-webmail-groupware.sh` — webmail/groupware endpoint check patterns
- `cloud-relay/caddy/Caddyfile` — authoritative routing for endpoint URLs

## Expected Output

- `scripts/validate-device-connectivity.sh` — executable validation script for all device-facing endpoints with --json/--verbose/--dry-run support
