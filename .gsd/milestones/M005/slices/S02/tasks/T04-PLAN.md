---
estimated_steps: 4
estimated_files: 3
---

# T04: Document round-trip testing procedure and create test helper script

**Slice:** S02 — Email Round-Trip — Inbound & Outbound Delivery
**Milestone:** M005

## Description

The live round-trip test requires human interaction (checking external email accounts, reading headers). This task creates a helper script that automates what can be automated (sending test email, polling logs, providing header checking instructions) and clear documentation for the manual steps. Together with T03's validation script, this completes the S02 verification toolkit.

## Steps

1. **Create `scripts/test-mail-roundtrip.sh` helper script.** The script should:
   - Accept `--domain`, `--recipient` (external address), `--sender` (local user), and `--dry-run` flags
   - **Outbound test:** Use `swaks` (or fall back to the existing `dns/authtest` sender pattern) to send a DKIM-signed test email from the local mail server via SMTP submission (port 587) to the external recipient. Include a unique message ID and timestamp for tracking.
   - **Log polling:** After sending, tail Postfix mail logs (on home device) looking for delivery confirmation (`status=sent`) or errors (`status=bounced`, `status=deferred`). Timeout after 60 seconds with diagnostic output.
   - **Header check instructions:** Print instructions for the human to check the recipient's inbox for SPF/DKIM/DMARC results in the Authentication-Results header. Include what "pass" looks like for each.
   - **Inbound test prompt:** Print instructions for the human to reply or send a new email to the local address, then poll the home device mailbox (via IMAP or Maildir) for arrival.
   - Support `--dry-run` to show the test sequence without actually sending.

2. **Create `docs/validation/mail-roundtrip.md` documentation.** Include:
   - Prerequisites (working tunnel, DNS, TLS — S01 outputs)
   - Step-by-step outbound test procedure with expected results
   - Step-by-step inbound test procedure with expected results
   - How to read Authentication-Results headers (SPF, DKIM, DMARC fields)
   - Troubleshooting section: mail in spam folder, greylisting delays, blocklist hits, DKIM failures, SPF softfail, relay denied errors
   - Reference to `scripts/lib/validate-mail.sh` for pre-flight checks

3. **Add roundtrip test reference to validation docs.** Ensure the mail validation section in `scripts/validate-infrastructure.sh` comments reference the round-trip procedure doc for the human-in-the-loop steps.

4. **Test dry-run mode.** Run `bash scripts/test-mail-roundtrip.sh --dry-run --domain example.com --recipient test@gmail.com --sender alice` and verify it prints the complete test sequence without sending any mail.

## Must-Haves

- [ ] `scripts/test-mail-roundtrip.sh` sends outbound test email and polls for delivery confirmation
- [ ] Script prints clear instructions for human header verification (SPF/DKIM/DMARC)
- [ ] Script handles inbound test prompting and mailbox polling
- [ ] `docs/validation/mail-roundtrip.md` covers both directions with troubleshooting
- [ ] `--dry-run` mode works without sending mail or requiring live infrastructure
- [ ] Troubleshooting covers: spam folder, greylisting, blocklists, DKIM fail, SPF fail, relay denied

## Verification

- `bash scripts/test-mail-roundtrip.sh --dry-run --domain example.com --recipient test@gmail.com --sender alice` exits 0 and prints test sequence
- `docs/validation/mail-roundtrip.md` exists with inbound, outbound, header-reading, and troubleshooting sections
- `shellcheck scripts/test-mail-roundtrip.sh` passes (or only cosmetic warnings)

## Observability Impact

- Signals added/changed: Round-trip script outputs timestamped test events (send, poll, receive) for post-mortem review
- How a future agent inspects this: Run script with `--dry-run` to see expected flow; run with `--verbose` for real-time log tailing
- Failure state exposed: Script exits non-zero with diagnostic message on send failure or delivery timeout; points to specific logs to check

## Inputs

- `home-device/tests/test-mail-flow.sh` — existing local mail test pattern (swaks usage, log checking)
- `dns/authtest/sender.go` — existing DKIM test email sender (reference for what a test email looks like)
- T01 outputs — relay configs (script needs to know submission port and server address)
- T02 outputs — DKIM signing config (script references expected signing behavior)
- T03 outputs — validation script (referenced in documentation for pre-flight checks)

## Expected Output

- `scripts/test-mail-roundtrip.sh` — new helper script for round-trip testing
- `docs/validation/mail-roundtrip.md` — new documentation for mail round-trip procedure
- `scripts/validate-infrastructure.sh` — minor comment update referencing roundtrip docs
