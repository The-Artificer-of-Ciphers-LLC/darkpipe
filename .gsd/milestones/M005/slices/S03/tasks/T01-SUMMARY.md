---
id: T01
parent: S03
milestone: M005
provides:
  - Device connectivity validation script with 8 endpoint checks
key_files:
  - scripts/validate-device-connectivity.sh
key_decisions:
  - Used file-based result storage (one JSON file per check) for bash 3.2 compatibility, matching validate-infrastructure.sh pattern
patterns_established:
  - Per-check JSON result files in temp dir with json_escape helper for safe detail embedding
  - Dual curl calls (body + status code separately) for bash 3 compatibility instead of fd redirection
observability_surfaces:
  - "scripts/validate-device-connectivity.sh --json for machine-readable structured results"
  - "scripts/validate-device-connectivity.sh --verbose for timestamped stderr diagnostics"
  - "Each check reports: name, status, url, detail, timestamp"
duration: 20m
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T01: Build device connectivity validation script

**Created `scripts/validate-device-connectivity.sh` — pre-flight validation orchestrator for all 8 device-facing endpoints with --json/--verbose/--dry-run support.**

## What Happened

Built the validation script following `validate-infrastructure.sh` structural patterns. The script checks 8 endpoints:

1. **autoconfig** — GET autoconfig XML, verify `<emailProvider>` present
2. **autodiscover** — POST autodiscover XML, verify `<Protocol>` present
3. **profile-server-health** — GET `/health/live`, expect 200
4. **webmail** — GET webmail root, expect HTML with webmail indicator
5. **monitoring-dashboard** — GET `/status`, expect HTML
6. **monitoring-json** — GET `/status?format=json`, expect JSON with health data
7. **imap-tls** — openssl s_client to port 993, verify TLS handshake
8. **smtp-starttls** — openssl s_client -starttls smtp to port 587, verify STARTTLS

Each check records name, status (pass/fail), URL tested, detail message, and timestamp. Dry-run mode returns mock passes without contacting live infrastructure.

## Verification

- `bash scripts/validate-device-connectivity.sh --dry-run` → exits 0, all 8 checks pass ✅
- `bash scripts/validate-device-connectivity.sh --dry-run --json | jq .` → valid JSON with 8 checks, overall_status "pass" ✅
- `shellcheck scripts/validate-device-connectivity.sh` → only SC2329 informational note (cleanup function invoked via trap, not directly) ✅

### Slice-level verification (partial — T01 is intermediate):
- ✅ `bash scripts/validate-device-connectivity.sh --dry-run` exits 0
- ⏳ `bash scripts/validate-device-connectivity.sh` against live infrastructure — requires T02
- ⏳ `docs/validation/device-connectivity-report.md` — created in T02
- ⏳ Monitoring JSON API healthy from external — verified in T05

## Diagnostics

- Run `scripts/validate-device-connectivity.sh --json` for machine-readable results (pipe to `jq .`)
- Run `scripts/validate-device-connectivity.sh --verbose` for timestamped diagnostic lines on stderr
- Failed checks in human mode automatically show URL and detail; pass checks show detail only with --verbose
- JSON output includes `total_checks`, `passed`, `failed` counts plus per-check array

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `scripts/validate-device-connectivity.sh` — New executable validation script for 8 device-facing endpoints
