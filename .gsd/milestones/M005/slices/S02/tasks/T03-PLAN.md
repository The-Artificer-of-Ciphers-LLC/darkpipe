---
estimated_steps: 5
estimated_files: 3
---

# T03: Build mail validation script and integrate into infrastructure validator

**Slice:** S02 — Email Round-Trip — Inbound & Outbound Delivery
**Milestone:** M005

## Description

S01 established a validation framework with section runner scripts (`scripts/lib/validate-*.sh`) orchestrated by `scripts/validate-infrastructure.sh`. This task adds a mail-specific validation section that checks: IP blocklist status, DKIM key pair consistency, transport map correctness, outbound relay config presence, and Rspamd DKIM signing config. These are automated config/infrastructure checks — the actual round-trip send/receive test is in T04.

## Steps

1. **Create `scripts/lib/validate-mail.sh` following the section runner pattern.** Study one existing validator (e.g., `validate-dns.sh`) for the pattern: function-per-check, consistent output formatting, support for `--json`, `--verbose`, `--dry-run` flags passed from orchestrator. Implement these checks:
   - **Blocklist scan:** Query Spamhaus ZEN (`zen.spamhaus.org`), Barracuda (`b.barracudacentral.org`), and SORBS (`dnsbl.sorbs.net`) via `dig` for the cloud relay IP. NXDOMAIN = clean, any A record = listed.
   - **DKIM key consistency:** Extract the public key from DNS TXT record (`${selector}._domainkey.${domain}`) and compare with the private key file if accessible (or just verify the DNS record exists and has valid format).
   - **Transport map check:** Verify `cloud-relay/postfix-config/transport` has no wildcard `*` entry and has at least one domain-specific entry.
   - **Outbound relay config:** Based on active profile, verify the relayhost/smarthost config exists in the correct config file.
   - **Rspamd DKIM config:** Verify `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` exists and contains required directives.

2. **Add `--dry-run` mock results.** When `--dry-run` is set, return mock pass results for all checks without querying DNS or reading live configs. Follow the pattern from existing validators.

3. **Add `--json` output mode.** Each check outputs a JSON object with `name`, `status` (pass/fail/skip), `detail`, and optional `severity`. Follow existing validator JSON schema.

4. **Integrate into `scripts/validate-infrastructure.sh`.** Source `scripts/lib/validate-mail.sh` and add it as a section in the orchestration script, after the existing sections (dns, tls, tunnel, ports, stability). Pass through the global flags.

5. **Test the script in dry-run mode.** Run `bash scripts/lib/validate-mail.sh --dry-run` and `bash scripts/validate-infrastructure.sh --dry-run` to verify integration and output format.

## Must-Haves

- [ ] `scripts/lib/validate-mail.sh` exists with all 5 checks implemented
- [ ] Supports `--json`, `--verbose`, `--dry-run` flags
- [ ] Blocklist check queries at least 3 DNSBLs
- [ ] DKIM key presence verified via DNS query
- [ ] Transport map checked for no-wildcard + domain-specific entries
- [ ] Integrated into `scripts/validate-infrastructure.sh` orchestrator
- [ ] `--dry-run` mode returns mock results without external queries

## Verification

- `bash scripts/lib/validate-mail.sh --dry-run` exits 0 and outputs check results
- `bash scripts/lib/validate-mail.sh --dry-run --json` outputs valid JSON per check
- `bash scripts/validate-infrastructure.sh --dry-run` includes mail section in output
- `shellcheck scripts/lib/validate-mail.sh` passes (or only cosmetic warnings)

## Observability Impact

- Signals added/changed: Validation script outputs structured check results (pass/fail with details) for all mail infrastructure checks
- How a future agent inspects this: Run `validate-mail.sh --verbose` for detailed per-check diagnostics; run with `--json` for machine-parseable output
- Failure state exposed: Each check reports specific failure reason (e.g., "IP 1.2.3.4 listed on zen.spamhaus.org", "no DKIM TXT record found for darkpipe._domainkey.example.com")

## Inputs

- `scripts/lib/validate-dns.sh` (or any existing validator) — pattern reference for section runner format
- `scripts/validate-infrastructure.sh` — orchestration script to integrate into
- T01 outputs — transport map and relay configs to validate
- T02 outputs — Rspamd DKIM config to validate
- S02-RESEARCH.md — blocklist query technique (DNSBL via `dig`)

## Expected Output

- `scripts/lib/validate-mail.sh` — new validation section script with 5 checks
- `scripts/validate-infrastructure.sh` — updated to include mail section
