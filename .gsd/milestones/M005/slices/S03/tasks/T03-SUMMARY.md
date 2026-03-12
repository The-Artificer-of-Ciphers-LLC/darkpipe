---
id: T03
parent: S03
milestone: M005
provides:
  - iOS/macOS section of validation report updated with blocked status and re-validation instructions
key_files:
  - docs/validation/device-connectivity-report.md
key_decisions:
  - Documented task as blocked rather than skipped — DNS NXDOMAIN prevents all device onboarding testing; no workaround exists for this task without live DNS
patterns_established:
  - none
observability_surfaces:
  - Validation report iOS/macOS section documents blocker with timestamp and re-validation command
duration: 10m
verification_result: blocked
completed_at: 2026-03-12
blocker_discovered: false
---

# T03: Validate iOS/macOS device onboarding from cellular network

**Documented iOS/macOS onboarding as blocked — DNS NXDOMAIN for darkpipe.email prevents all device connectivity testing.**

## What Happened

Ran the device connectivity pre-checks against live infrastructure before attempting any human-in-the-loop testing. All 8 automated endpoint checks failed with HTTP 000 / DNS NXDOMAIN — identical to the T02 results. `dig +short mail.darkpipe.email` returns empty (NXDOMAIN). No A, AAAA, or MX records exist for darkpipe.email or any of its subdomains.

Since this task requires:
1. Reaching `https://mail.darkpipe.email/profile/` to generate a QR code
2. An iPhone on cellular data downloading a .mobileconfig over HTTPS
3. IMAP/SMTP connectivity for email send/receive
4. CalDAV/CardDAV endpoints for calendar/contacts sync

...and all of these require DNS resolution, the task is blocked. No part of the iOS/macOS onboarding flow can be attempted.

Updated the iOS/macOS section of the validation report with the blocked status, timestamp, blocker description, and re-validation instructions.

This is NOT a blocker_discovered situation — the DNS issue was already known from T02 and doesn't invalidate the slice plan. Once DNS is restored, this task can be executed as planned.

## Verification

- `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json` — confirmed 0/8 pass, all NXDOMAIN
- `dig +short mail.darkpipe.email` — returns empty (NXDOMAIN)
- `docs/validation/device-connectivity-report.md` iOS/macOS section updated with blocked status

## Diagnostics

- Re-run `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json | jq .` after DNS restoration
- Check DNS with `dig +short mail.darkpipe.email` — should return an IP when DNS is configured
- iOS/macOS section of validation report documents current state and re-validation steps

## Deviations

Task plan assumes live infrastructure is reachable. DNS has been down since T02 — no human testing was possible. Report updated to document blocked state instead of test results.

## Known Issues

- `darkpipe.email` DNS zone returns NXDOMAIN for all subdomains — blocks all device connectivity testing (T03, T04, T05)
- This is not fixable from the agent — requires DNS zone configuration at domain registrar or nameserver provider

## Files Created/Modified

- `docs/validation/device-connectivity-report.md` — Updated iOS/macOS section with blocked status, timestamp, and re-validation instructions

## Slice Verification Status

| Check | Status |
|---|---|
| `bash scripts/validate-device-connectivity.sh --dry-run` exits 0 | ✅ Pass (verified in T01) |
| `bash scripts/validate-device-connectivity.sh` exits 0 against live infra | ❌ Fail — 0/8, DNS NXDOMAIN |
| `docs/validation/device-connectivity-report.md` exists with all sections | ⏳ Partial — pre-checks and iOS sections populated, remaining sections pending T04/T05 |
| Monitoring JSON API returns healthy | ❌ Fail — unreachable |
