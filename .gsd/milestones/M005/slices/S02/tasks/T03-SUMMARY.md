---
id: T03
parent: S02
milestone: M005
provides:
  - Mail infrastructure validation script with 5 automated checks (blocklist, DKIM, transport map, relay config, Rspamd)
  - Mail section integrated into infrastructure validation orchestrator
key_files:
  - scripts/lib/validate-mail.sh
  - scripts/validate-infrastructure.sh
key_decisions:
  - Auto-detect relay IP from DNS A record when RELAY_IP not set, with graceful skip of blocklist checks if unresolvable
  - Check all three mail server profiles when HOME_PROFILE is unset rather than requiring explicit profile selection
patterns_established:
  - Mail validation section follows same pattern as existing validators (function-per-check, JSON helpers, dry-run mocks, standalone + sourced execution)
  - DNSBL lookup via reversed-IP dig query against multiple blocklists
observability_surfaces:
  - "validate-mail.sh --verbose: per-check diagnostics with details"
  - "validate-mail.sh --json: machine-parseable JSON with name/status/detail/severity per check"
  - "validate-infrastructure.sh --dry-run: mail section appears in orchestrated output"
duration: 20m
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T03: Build mail validation script and integrate into infrastructure validator

**Created `validate-mail.sh` with 5 mail infrastructure checks (blocklist scan, DKIM DNS record, transport map, outbound relay config, Rspamd DKIM config) and integrated it as the mail section in the infrastructure validation orchestrator.**

## What Happened

Built `scripts/lib/validate-mail.sh` following the established section runner pattern from `validate-dns.sh`. The script implements 5 checks:

1. **Blocklist scan** — Queries 3 DNSBLs (Spamhaus ZEN, Barracuda, SORBS) via reversed-IP dig lookup. NXDOMAIN = clean, A record = listed.
2. **DKIM DNS record** — Verifies `${selector}._domainkey.${domain}` TXT record exists with `v=DKIM1` and `p=` tags.
3. **Transport map** — Checks `cloud-relay/postfix-config/transport` for no wildcard `*` entries and at least one domain-specific entry.
4. **Outbound relay config** — Verifies relay config in active profile (or all three if unset): `relayhost` in Postfix, `target.smtp` in Maddy, `queue.route` in Stalwart.
5. **Rspamd DKIM config** — Verifies `dkim_signing.conf` exists with required directives (`sign_authenticated`, `selector`, `path`, `use_domain`).

Integrated into `validate-infrastructure.sh` by adding `mail` to the SECTIONS list. The orchestrator's sourcing loop auto-discovers `run_mail_validation()` from the script.

## Verification

- `bash scripts/lib/validate-mail.sh --dry-run` → exits 0, shows 7 passing checks (3 blocklist + 4 config)
- `bash scripts/lib/validate-mail.sh --dry-run --verbose` → shows detail lines per check
- `bash scripts/lib/validate-mail.sh --dry-run --json` → valid JSON output verified with `python3 -m json.tool`
- `bash scripts/validate-infrastructure.sh --dry-run` → includes mail section, overall pass
- `bash scripts/validate-infrastructure.sh --dry-run --json` → valid JSON with mail section
- `shellcheck scripts/lib/validate-mail.sh` → passes clean (no warnings)
- Slice-level config checks: `relayhost` found in Postfix main.cf, transport map has domain-specific entries with no wildcard

## Diagnostics

- Run `validate-mail.sh --verbose` for per-check detail output
- Run `validate-mail.sh --json` for machine-parseable results with name/status/detail/severity fields
- Each check reports specific failure reason (e.g., "IP listed on zen.spamhaus.org", "no DKIM TXT record found")
- Set `HOME_PROFILE` env var to scope outbound relay check to a specific mail server profile

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `scripts/lib/validate-mail.sh` — New mail validation section with 5 checks, standalone + sourced execution support
- `scripts/validate-infrastructure.sh` — Added `mail` to SECTIONS list for orchestration
