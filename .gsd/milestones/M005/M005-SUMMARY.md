---
id: M005
provides:
  - Infrastructure validation framework (validate-infrastructure.sh) with 6 sections, 35+ checks, JSON/verbose/dry-run modes
  - Device connectivity validation script (validate-device-connectivity.sh) with 8 endpoint checks
  - Mail round-trip test helper script (test-mail-roundtrip.sh) with human-in-the-loop orchestration
  - Fixed Caddyfile IP mismatch (hardcoded 10.0.0.2 → {$HOME_DEVICE_IP} env var)
  - Fixed cloud relay transport map routing loop (wildcard → domain-specific entries)
  - Outbound relay configuration for all three mail server profiles (Postfix, Maddy, Stalwart)
  - Rspamd DKIM signing configuration scoped to outbound-only traffic
  - Device connectivity validation report template with all test sections
  - Deployment documentation (deploy/README.md, mail-roundtrip.md)
  - Fixed false-positive bug in IMAP/SMTP TLS validation checks
key_decisions:
  - Infrastructure validation is a single orchestration script sequencing existing validators — no new validation logic hand-rolled
  - Domain-specific transport map entries instead of wildcard — prevents outbound mail routing loop
  - Rspamd DKIM signing for all mail server profiles via single shared config
  - Outbound relay via cloud relay WireGuard IP 10.8.0.1:25 for all three home mail server profiles
  - Round-trip verification is human-in-the-loop with helper script (live external email accounts required)
  - Device connectivity validation follows validate-infrastructure.sh patterns for consistency
  - Device validation is primarily UAT with automated pre-flight
  - Maddy uses target.smtp (transparent forwarding) for smarthost relay
  - Stalwart uses queue.strategy route + queue.route."relay" pattern per official docs
  - Used dig-based DNS checks instead of Go binary dependency for portability
  - TLS checks connect to RELAY_IP with SNI rather than resolving each subdomain
  - Port checks split relay-facing (25, 443) vs tunnel-facing (587, 993)
  - Stability section returns skip (not fail) when not root
  - Documented M005 as CANNOT FULLY VALIDATE — tooling complete, live endpoints blocked by DNS NXDOMAIN
patterns_established:
  - Validation section pattern: _check() JSON helper, _dry_run_checks(), run_<section>_validation() entry point
  - Sourced-function integration: main script sources lib scripts defining run_*() functions
  - File-based JSON result storage for bash 3.2 compatibility
  - Human-in-the-loop test scripts print structured instructions for manual steps
  - Test IDs embedded in subject and X-DarkPipe-Test-ID header for traceability
observability_surfaces:
  - scripts/validate-infrastructure.sh --json — full infrastructure validation with 6 sections (dns, tls, tunnel, ports, stability, mail)
  - scripts/validate-device-connectivity.sh --json — 8 device endpoint checks with pass/fail/detail
  - scripts/test-mail-roundtrip.sh --dry-run — mail round-trip test plan without live infrastructure
  - docs/validation/device-connectivity-report.md — persistent test state with per-endpoint and per-device results
  - deploy/README.md — deployment documentation with validation guide
requirement_outcomes: []
duration: ~4 hours
verification_result: partial — all tooling verified via dry-run; live external validation blocked by DNS NXDOMAIN for darkpipe.email
completed_at: 2026-03-12
---

# M005: Design Validation — External Access & Device Connectivity

**Built complete external validation framework and fixed critical configuration issues, but live external verification blocked by DNS NXDOMAIN — tooling is ready for re-validation when DNS is restored.**

## What Happened

M005 set out to prove the DarkPipe architecture works end-to-end from the public internet. The milestone was structured in three slices: infrastructure validation (S01), email round-trip (S02), and device connectivity (S03).

### S01: Infrastructure Validation — DNS, TLS & Tunnel

Built `scripts/validate-infrastructure.sh` — a comprehensive infrastructure validation orchestrator with 6 sections (dns, tls, tunnel, ports, stability, mail) producing structured JSON output. Each section has its own lib script following a consistent pattern: per-check JSON helpers, dry-run mock results, and live validation logic.

Key fix: The Caddyfile had 11 instances of hardcoded `10.0.0.2` that didn't match the WireGuard subnet default `10.8.0.2`. Replaced all with `{$HOME_DEVICE_IP}` env var with proper default. This would have caused all proxied traffic to fail in production.

The DNS section validates 9 record types (MX, A, SPF, DKIM, DMARC, 2×SRV, 2×CNAME) against dual external resolvers with propagation mismatch detection. TLS validates certificate chains, expiry, and domain matches for 3 HTTPS endpoints. Ports check TCP connectivity on 25, 443, 587, 993 with correct target routing (relay vs tunnel IP). Stability wraps the outage simulation test with root-privilege gating.

### S02: Email Round-Trip — Inbound & Outbound Delivery

Fixed the cloud relay transport map which had a wildcard `*` routing ALL mail back to the relay daemon — this would have created a routing loop preventing any outbound delivery. Replaced with domain-specific `${RELAY_DOMAIN}` entries.

Configured outbound relay for all three home mail server profiles pointing to cloud relay at `10.8.0.1:25` via WireGuard tunnel. Without this, home mail servers would attempt direct delivery from the residential IP (port 25 blocked by ISPs, fails SPF).

Created Rspamd DKIM signing configuration scoped to authenticated/local/WireGuard traffic only, with key directory mounting across all compose variants. Built the mail validation section (blocklist scan, DKIM DNS, transport map, relay config, Rspamd config checks) and integrated it into the orchestrator.

Created `scripts/test-mail-roundtrip.sh` — a human-in-the-loop round-trip test helper that orchestrates outbound send via swaks, log polling, authentication header verification, and inbound delivery checking.

### S03: Device Connectivity — Mobile, Desktop & Webmail

Built `scripts/validate-device-connectivity.sh` with 8 endpoint checks (autoconfig XML, autodiscover XML, profile server health, webmail, monitoring dashboard HTML, monitoring JSON API, IMAP TLS, SMTP STARTTLS).

Found and fixed a false-positive bug in the IMAP/SMTP TLS checks — the original grep pattern matched "OK" in OpenSSL error strings like "BIO_lookup_ex", causing DNS-unreachable hosts to report as "pass".

Created the device connectivity validation report at `docs/validation/device-connectivity-report.md` with all required sections. Ran live validation — all 8 checks fail with DNS NXDOMAIN for `darkpipe.email`. The domain has no A, AAAA, or MX records configured. All human testing tasks (iOS/macOS onboarding, Thunderbird, webmail, monitoring) were documented as blocked with re-validation instructions.

### DNS Blocker

The `darkpipe.email` domain returns NXDOMAIN for all subdomains. This is an external DNS configuration issue (registrar/nameserver), not a DarkPipe code or config problem. All live validation is blocked by this. The validation tooling works correctly in dry-run mode and is ready for re-validation when DNS is restored.

## Cross-Slice Verification

### Success Criteria Assessment

| Criterion | Status | Evidence |
|---|---|---|
| DNS records resolve from external resolvers | ❌ BLOCKED | DNS validation tool built (9 record types), but darkpipe.email returns NXDOMAIN |
| TLS certificates valid and trusted | ❌ BLOCKED | TLS validation built (chain, expiry, domain match), endpoints unreachable |
| IMAP (993) / SMTP (587) accept external connections | ❌ BLOCKED | Port checks built, but DNS prevents connectivity |
| Webmail loads over HTTPS externally | ❌ BLOCKED | Webmail check in device-connectivity script, DNS prevents access |
| Full inbound round-trip | ❌ BLOCKED | Transport maps fixed, relay configs added, test script built — live test blocked |
| Full outbound round-trip | ❌ BLOCKED | All three profiles configured for relay via 10.8.0.1:25 — live test blocked |
| Mobile device syncs via .mobileconfig | ❌ BLOCKED | Device connectivity report prepared, DNS prevents device testing |
| Monitoring dashboard shows healthy | ❌ BLOCKED | Monitoring checks built, endpoints unreachable |
| Tunnel reconnects after interruption | ⏳ PARTIAL | Stability validation wraps outage-sim.sh, not exercised in live M005 context |

### What IS Verified

- **Validation tooling**: Both scripts pass dry-run with all checks (28 infrastructure + 8 device = 36 total) ✅
- **Configuration fixes**: Caddyfile IP mismatch fixed (would have broken all proxied traffic) ✅
- **Transport map fix**: Routing loop eliminated (would have prevented all outbound mail) ✅
- **Outbound relay**: All three mail server profiles configured to relay via cloud ✅
- **DKIM signing**: Rspamd configured for all profiles with correct scoping ✅
- **Script bug fix**: IMAP/SMTP false-positive detection corrected ✅
- **Documentation**: deploy/README.md, mail round-trip guide, device connectivity report ✅

### What Requires DNS Restoration

All 9 success criteria require live DNS resolution for `darkpipe.email`. Re-validation steps are documented in `docs/validation/device-connectivity-report.md` Executive Summary.

**Re-validation commands:**
```bash
RELAY_DOMAIN=darkpipe.email scripts/validate-infrastructure.sh --json
RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json
RELAY_DOMAIN=darkpipe.email scripts/test-mail-roundtrip.sh --domain darkpipe.email --recipient <external@gmail.com> --sender admin
```

## Requirement Changes

No requirements changed status during this milestone. M005 was a validation milestone — it aimed to prove existing requirements work end-to-end from external networks. Since live validation was blocked by DNS, no requirement can be promoted from "validated (internal)" to "validated (external)". All requirements retain their prior status from M001.

## Forward Intelligence

### What the next milestone should know
- DNS for `darkpipe.email` must be configured before any external validation can proceed. This is a prerequisite for completing the live verification portion of M005.
- Three critical config fixes were made (Caddyfile IP, transport map wildcard, outbound relay) that would have caused failures in any real deployment. These are now fixed in the codebase.
- All validation tooling is ready — running the three scripts above against live infrastructure will complete the validation.
- The validation report at `docs/validation/device-connectivity-report.md` has step-by-step re-validation instructions for each test category.

### What's fragile
- **DNS dependency** — All external validation is gated on `darkpipe.email` DNS records existing. No workaround.
- **Placeholder slice summaries** — S01/S02/S03 summaries were doctor-generated placeholders. Task summaries are the authoritative source.
- **Rspamd DKIM keys** — The `dkim-keys` directory has a .gitkeep and README but no actual keys. Key provisioning is a deployment step.
- **Stalwart config format** — Used `queue.strategy` + `queue.route."relay"` based on Context7 docs. Stalwart is pre-v1.0 and schema may change.

### Authoritative diagnostics
- `scripts/validate-infrastructure.sh --dry-run --json` — proves all 6 validation sections work correctly
- `scripts/validate-device-connectivity.sh --dry-run --json` — proves all 8 device checks work correctly
- Task summaries in `.gsd/milestones/M005/slices/*/tasks/` — authoritative record of what was built and verified
- `docs/validation/device-connectivity-report.md` — documents current live state and re-validation steps

### What assumptions changed
- **Assumed DNS would be live** — M005 planning assumed `darkpipe.email` had working DNS. It does not. This turned a validation milestone into a tooling+config-fix milestone.
- **Found config bugs** — The Caddyfile IP mismatch and transport map routing loop were not anticipated. Both would have been showstoppers in any real deployment.
- **IMAP/SMTP validation had a false-positive bug** — OpenSSL error strings containing "OK" as a substring caused unreachable hosts to appear connected. Fixed.

## Files Created/Modified

- `scripts/validate-infrastructure.sh` — Infrastructure validation orchestrator (6 sections, 35+ checks)
- `scripts/lib/validate-dns.sh` — DNS validation (9 record types, dual resolvers)
- `scripts/lib/validate-tls.sh` — TLS validation (chain, expiry, domain match)
- `scripts/lib/validate-tunnel.sh` — Tunnel health validation
- `scripts/lib/validate-ports.sh` — Port reachability (25, 443, 587, 993)
- `scripts/lib/validate-stability.sh` — Stability/reconnection validation
- `scripts/lib/validate-mail.sh` — Mail infrastructure validation (5 checks)
- `scripts/validate-device-connectivity.sh` — Device endpoint validation (8 checks)
- `scripts/test-mail-roundtrip.sh` — Round-trip mail test helper
- `cloud-relay/caddy/Caddyfile` — Fixed hardcoded IP → env var
- `cloud-relay/docker-compose.yml` — Added HOME_DEVICE_IP env var
- `cloud-relay/.env.example` — Added HOME_DEVICE_IP documentation
- `cloud-relay/postfix-config/transport` — Fixed wildcard routing loop
- `home-device/maddy/maddy.conf` — Added outbound relay config
- `home-device/stalwart/config.toml` — Added outbound relay config
- `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` — DKIM signing config
- `home-device/docker-compose.yml` — Added DKIM key volume mount
- `home-device/docker-compose.podman-selinux.yml` — Added DKIM key volume mount with SELinux label
- `home-device/spam-filter/rspamd/dkim-keys/README.md` — Key provisioning docs
- `deploy/README.md` — Deployment documentation with validation guide
- `docs/validation/mail-roundtrip.md` — Round-trip testing documentation
- `docs/validation/device-connectivity-report.md` — Device connectivity validation report
