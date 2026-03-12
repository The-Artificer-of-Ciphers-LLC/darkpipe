---
id: T04
parent: S02
milestone: M005
provides:
  - Round-trip mail test helper script with dry-run, outbound send, log polling, inbound check
  - Documentation covering both directions with header reading and troubleshooting
  - Round-trip test reference in infrastructure validation orchestrator
key_files:
  - scripts/test-mail-roundtrip.sh
  - docs/validation/mail-roundtrip.md
  - scripts/validate-infrastructure.sh
key_decisions:
  - Use printf for colored output instead of heredocs to avoid literal escape sequence rendering issues
  - Unique test ID per run (darkpipe-rt-{timestamp}-{random}) for log correlation and mailbox search
patterns_established:
  - Human-in-the-loop test scripts print structured instructions for manual steps and automate everything else
  - Test IDs embedded in subject and X-DarkPipe-Test-ID header for traceability
observability_surfaces:
  - "Timestamped log events: SEND_START, SEND_OK, SEND_FAIL, POLL_START, DELIVERY_OK, DELIVERY_BOUNCE, POLL_TIMEOUT, INBOUND_POLL_START, INBOUND_OK, INBOUND_TIMEOUT, TEST_COMPLETE"
  - "--dry-run mode shows full test sequence without live infrastructure"
  - "--verbose mode enables real-time log tailing and detailed output"
duration: 1 step (retry — fixed heredoc escape rendering from previous attempt)
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T04: Document round-trip testing procedure and create test helper script

**Created mail round-trip test helper script and documentation covering outbound/inbound testing, authentication header verification, and troubleshooting.**

## What Happened

Built `scripts/test-mail-roundtrip.sh` that orchestrates a 5-step round-trip mail test:

1. Pre-flight infrastructure checks (delegates to `validate-mail.sh`)
2. Outbound test email via SMTP submission (swaks, port 587, TLS + auth)
3. Log polling for delivery confirmation (status=sent/bounced/deferred with timeout)
4. Authentication header verification instructions (SPF/DKIM/DMARC with examples per provider)
5. Inbound delivery test with IMAP mailbox polling

The script accepts `--domain`, `--recipient`, `--sender`, and `--dry-run` flags. Dry-run prints the complete test plan without requiring live infrastructure or swaks.

Created `docs/validation/mail-roundtrip.md` with prerequisites, step-by-step procedures for both directions, header reading guide, and troubleshooting for all six required failure modes (spam folder, greylisting, blocklists, DKIM fail, SPF fail, relay denied).

The infrastructure validation orchestrator (`scripts/validate-infrastructure.sh`) already had round-trip doc references from T03 — verified they are present and correct.

## Verification

- `bash scripts/test-mail-roundtrip.sh --dry-run --domain example.com --recipient test@gmail.com --sender alice` → exits 0, prints complete 5-step test sequence with no literal escape codes
- `shellcheck scripts/test-mail-roundtrip.sh` → only SC2059 info-level warnings (color vars in printf format strings — matches codebase pattern)
- `docs/validation/mail-roundtrip.md` exists with outbound, inbound, header-reading, and all six troubleshooting sections
- Slice-level checks: `relayhost` present in postfix-dovecot config, transport map has domain-specific entry (no wildcard)

## Diagnostics

- Run `scripts/test-mail-roundtrip.sh --dry-run` to see the full test flow without live infrastructure
- Run with `--verbose` for real-time log tailing during live tests
- Script outputs structured log events (timestamp + event name + detail) for post-mortem review
- Non-zero exit on send failure or delivery timeout, with diagnostic message pointing to specific logs

## Deviations

Previous attempt used heredocs with color variables, which rendered literal `\033[...` escape sequences. Fixed by replacing all heredocs containing color variables with printf/echo lines, matching the existing codebase pattern.

## Known Issues

None.

## Files Created/Modified

- `scripts/test-mail-roundtrip.sh` — Round-trip mail test helper (outbound send, log polling, header instructions, inbound check, dry-run)
- `docs/validation/mail-roundtrip.md` — Full round-trip testing documentation with troubleshooting
- `scripts/validate-infrastructure.sh` — Already had roundtrip doc references (verified, no changes needed)
