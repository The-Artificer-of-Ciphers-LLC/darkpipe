# Device Connectivity Validation Report

**Domain:** darkpipe.email
**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

---

## Executive Summary

| Metric | Value |
|---|---|
| Report created | 2026-03-12T15:08Z |
| Last updated | 2026-03-12T15:14Z |
| Automated pre-checks | ❌ 0/8 passed — DNS not resolving |
| iOS/macOS device onboarding | ❌ Blocked — DNS not resolving |
| Desktop client (Thunderbird) | ❌ Blocked — DNS not resolving |
| Webmail access | ❌ Blocked — DNS not resolving |
| Monitoring dashboard | ❌ Blocked — DNS not resolving |
| Monitoring JSON API | ❌ Blocked — DNS not resolving |
| Infrastructure validation | ❌ DNS/TLS/tunnel/ports all failing |
| Overall status | **❌ BLOCKED — DNS resolution failure** |

### M005 Verdict

**Result: ❌ CANNOT VALIDATE — DNS blocker prevents all external connectivity testing.**

All 8 device connectivity endpoints and all infrastructure validation checks fail because `darkpipe.email` has no DNS records (NXDOMAIN for A, AAAA, MX, CNAME on all subdomains). This is an external configuration dependency — the domain's DNS zone must be provisioned at the registrar/nameserver before any M005 success criteria can be proven from external networks.

**What is proven:**
- ✅ Validation tooling is complete and correct (`validate-device-connectivity.sh` dry-run passes 8/8)
- ✅ Infrastructure validation tooling exists (`validate-infrastructure.sh` with DNS/TLS/tunnel/ports/mail sections)
- ✅ False-positive bug in IMAP/SMTP TLS checks was found and fixed
- ✅ Report template covers all M005 criteria with structured test matrices

**What requires DNS restoration to validate:**
- DNS records (MX, A, SPF, DKIM, DMARC, SRV, CNAME)
- TLS certificates (Caddy auto-provisioned via ACME — requires DNS)
- WireGuard tunnel connectivity to relay
- All 8 device connectivity endpoints (autoconfig, autodiscover, profile server, webmail, monitoring dashboard, monitoring JSON, IMAP TLS, SMTP STARTTLS)
- iOS/macOS device onboarding (profile download, email, calendar, contacts)
- Thunderbird autoconfig and IMAP/SMTP authentication
- Webmail HTTPS login and send/receive
- Monitoring dashboard and JSON API health status

### Re-validation Instructions

Once DNS is configured for `darkpipe.email`:

```bash
# 1. Verify DNS resolves
dig +short mail.darkpipe.email   # Should return an IP

# 2. Run infrastructure validation
RELAY_DOMAIN=darkpipe.email scripts/validate-infrastructure.sh --verbose

# 3. Run device connectivity pre-checks
RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --verbose

# 4. If all 8 pass, proceed with manual device testing (T03/T04 steps)
# 5. Update this report with results and set overall verdict
```

---

## Automated Pre-Checks

Run via `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --verbose`

| # | Check | Status | URL / Target | Detail | Timestamp |
|---|---|---|---|---|---|
| 1 | autoconfig | ❌ FAIL | `https://autoconfig.darkpipe.email/.well-known/autoconfig/mail/config-v1.1.xml` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 2 | autodiscover | ❌ FAIL | `https://autodiscover.darkpipe.email/autodiscover/autodiscover.xml` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 3 | profile-server-health | ❌ FAIL | `https://mail.darkpipe.email/health/live` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 4 | webmail | ❌ FAIL | `https://mail.darkpipe.email/` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 5 | monitoring-dashboard | ❌ FAIL | `https://mail.darkpipe.email/status` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 6 | monitoring-json | ❌ FAIL | `https://mail.darkpipe.email/status?format=json` | HTTP 000 — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 7 | imap-tls | ❌ FAIL | `imaps://mail.darkpipe.email:993` | TLS handshake failed — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |
| 8 | smtp-starttls | ❌ FAIL | `smtp://mail.darkpipe.email:587` | STARTTLS handshake failed — DNS NXDOMAIN, host not resolvable | 2026-03-12T15:08:35Z |

**Pre-check result:** ❌ FAIL — 0 passed, 8 failed
**Checks run at:** 2026-03-12T15:08:35Z

### Root Cause

All 8 checks fail because `darkpipe.email` has no DNS A/AAAA records. `dig +short mail.darkpipe.email`, `dig +short autoconfig.darkpipe.email`, and `dig +short autodiscover.darkpipe.email` all return NXDOMAIN. No MX record exists either (`dig +short MX darkpipe.email` returns empty). The domain's DNS zone is either not configured, expired, or the registrar nameservers are not responding.

This blocks all device connectivity testing — HTTP endpoints cannot be reached and TLS handshakes cannot be initiated without DNS resolution.

### Re-validation

Once DNS is restored, re-run:
```bash
RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --verbose
```
All 8 checks should pass before proceeding to human testing tasks (T03–T05).

---

## iOS/macOS Device Onboarding

**Status:** ❌ Blocked — DNS not resolving (validated 2026-03-12T11:56Z)
**Blocker:** `darkpipe.email` returns NXDOMAIN for all subdomains. Profile server at `https://mail.darkpipe.email/profile/` is unreachable. Cannot generate QR code, download .mobileconfig, or test any device onboarding flow.
**Re-validation:** Once DNS is restored, re-run `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json` — all 8 pre-checks must pass before attempting iOS/macOS onboarding.

### Profile Installation

| Step | Expected | Result | Notes |
|---|---|---|---|
| Download .mobileconfig from profile server | Profile downloads over HTTPS | ❌ BLOCKED | DNS NXDOMAIN — `mail.darkpipe.email` not resolvable (2026-03-12T11:56Z) |
| Install profile on iOS device | Profile installs without errors | ⏳ | Depends on profile download |
| Mail account appears in Settings | IMAP/SMTP account configured | ⏳ | Depends on profile install |
| Calendar/Contacts sync configured | CalDAV/CardDAV endpoints set | ⏳ | Depends on profile install |

### Send/Receive Test (iOS)

| Test | Expected | Result | Notes |
|---|---|---|---|
| Send email from iOS to external address | Delivered within 60s | ⏳ | Blocked on DNS |
| Receive email from external address on iOS | Appears in inbox within 60s | ⏳ | Blocked on DNS |
| Reply to received email | Outbound delivery succeeds | ⏳ | Blocked on DNS |

### macOS Mail.app

| Test | Expected | Result | Notes |
|---|---|---|---|
| Install .mobileconfig on macOS | Profile installs, account appears | ⏳ | Blocked on DNS |
| Send/receive email via Mail.app | Bidirectional delivery works | ⏳ | Blocked on DNS |

---

## Desktop Client (Thunderbird)

**Status:** ❌ Blocked — DNS not resolving (validated 2026-03-12T12:56Z)
**Blocker:** `darkpipe.email` returns NXDOMAIN for all subdomains including `autoconfig.darkpipe.email` and `mail.darkpipe.email`. Thunderbird autoconfig endpoint at `https://autoconfig.darkpipe.email/.well-known/autoconfig/mail/config-v1.1.xml` returns HTTP 000 (DNS failure). IMAP port 993 and SMTP port 587 TLS handshakes fail — host not resolvable. No part of the Thunderbird validation flow can be attempted.
**Re-validation:** Once DNS is restored, re-run `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json` — autoconfig, imap-tls, and smtp-starttls checks must pass before attempting Thunderbird testing.

### Autoconfig Discovery

| Step | Expected | Result | Notes |
|---|---|---|---|
| Add account in Thunderbird | Autoconfig XML fetched automatically | ❌ BLOCKED | DNS NXDOMAIN — `autoconfig.darkpipe.email` not resolvable (2026-03-12T12:56Z) |
| IMAP server auto-populated | `mail.darkpipe.email:993` SSL/TLS | ❌ BLOCKED | Depends on autoconfig; DNS not resolving |
| SMTP server auto-populated | `mail.darkpipe.email:587` STARTTLS | ❌ BLOCKED | Depends on autoconfig; DNS not resolving |
| Authentication method | Normal password | ❌ BLOCKED | Depends on account setup |

### Send/Receive Test (Thunderbird)

| Test | Expected | Result | Notes |
|---|---|---|---|
| Send email from Thunderbird to external | Delivered within 60s | ❌ BLOCKED | DNS NXDOMAIN — SMTP unreachable |
| Receive email from external in Thunderbird | Appears in inbox within 60s | ❌ BLOCKED | DNS NXDOMAIN — IMAP unreachable |
| Folder sync (IMAP) | Folders visible and in-sync | ❌ BLOCKED | DNS NXDOMAIN — IMAP unreachable |

---

## Webmail Access

**Status:** ❌ Blocked — DNS not resolving (validated 2026-03-12T12:56Z)
**Blocker:** `mail.darkpipe.email` returns NXDOMAIN. Webmail at `https://mail.darkpipe.email/` returns HTTP 000 (DNS failure). Cannot test HTTPS certificate, login, send/receive, or mobile responsiveness.
**Re-validation:** Once DNS is restored, re-run `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json` — webmail check must pass before attempting webmail testing.

| Test | Expected | Result | Notes |
|---|---|---|---|
| Load webmail URL over HTTPS | Login page renders | ❌ BLOCKED | DNS NXDOMAIN — `mail.darkpipe.email` not resolvable (2026-03-12T12:56Z) |
| HTTPS certificate valid | No browser warnings | ❌ BLOCKED | Cannot check — host not resolvable |
| Login with app password | Inbox loads | ❌ BLOCKED | Depends on HTTPS load |
| Send email from webmail | Outbound delivery succeeds | ❌ BLOCKED | Depends on login |
| Receive email in webmail | Inbound mail visible in inbox | ❌ BLOCKED | Depends on login |
| Mobile-responsive layout | Usable at mobile viewport | ❌ BLOCKED | Depends on HTTPS load |

---

## Monitoring Dashboard

**Status:** ❌ Blocked — DNS not resolving (validated 2026-03-12T15:14Z)
**Blocker:** `mail.darkpipe.email` returns NXDOMAIN. Monitoring dashboard at `https://mail.darkpipe.email/status` returns HTTP 000 (DNS failure). Monitoring JSON API at `https://mail.darkpipe.email/status?format=json` also returns HTTP 000. Cannot verify dashboard rendering, auto-refresh, service health, or JSON API response structure.
**Re-validation:** Once DNS is restored, re-run `RELAY_DOMAIN=darkpipe.email scripts/validate-device-connectivity.sh --json` — monitoring-dashboard and monitoring-json checks must pass before manual dashboard verification.

| Test | Expected | Result | Notes |
|---|---|---|---|
| Load monitoring dashboard over HTTPS | Status page renders with Basic Auth | ❌ BLOCKED | DNS NXDOMAIN — `mail.darkpipe.email` not resolvable (HTTP 000, 2026-03-12T15:14Z) |
| All services show healthy | Green/healthy status for each service | ❌ BLOCKED | Depends on dashboard load |
| Auto-refresh working | 30-second meta refresh active | ❌ BLOCKED | Depends on dashboard load |
| JSON API returns structured health | `overall` field present, services enumerated | ❌ BLOCKED | `https://mail.darkpipe.email/status?format=json` returns HTTP 000 (2026-03-12T15:14Z) |
| Dashboard accessible from external network | No VPN required for status page | ❌ BLOCKED | Cannot test without DNS |

---

## Fixes Applied

| # | Issue | Root Cause | Fix Applied | Verified |
|---|---|---|---|---|
| 1 | IMAP/SMTP checks reported false-positive pass when DNS failed | `grep -qi 'OK'` matched "lookup" in OpenSSL error `BIO_lookup_ex` | Changed grep pattern from `'Verify return code: 0\|OK\|CONNECTED'` to `'CONNECTED('` — matches only actual TLS connection indicator | ✅ Dry-run passes; live run correctly reports fail for DNS-unresolvable hosts |
| 2 | All 8 endpoints unreachable | `darkpipe.email` DNS zone returns NXDOMAIN for all subdomains (mail, autoconfig, autodiscover) — no A/AAAA/MX records | **Not fixable from agent** — requires DNS zone configuration at domain registrar or nameserver provider | ❌ Awaiting DNS restoration |

---

## Final Automated Validation Run

**Timestamp:** 2026-03-12T15:14:50Z

### `validate-device-connectivity.sh`

```
Result: ❌ FAIL — 0 passed, 8 failed out of 8 checks
Root cause: DNS NXDOMAIN for all darkpipe.email subdomains
```

All 8 checks fail with HTTP 000 / TLS handshake failure due to DNS not resolving.

### `validate-infrastructure.sh`

```
Result: ❌ FAIL — DNS (0/9), TLS (0/3 + 6 skipped), Tunnel (1/3), Ports (0/1 + 3 skipped), Mail (3/5)
Only passing: tunnel_script_exists, transport_map, outbound_relay, rspamd_dkim_config
```

Infrastructure validation confirms the same DNS root cause cascading through all sections.

### `validate-device-connectivity.sh --dry-run`

```
Result: ✅ PASS — 8/8 checks pass in dry-run mode
Script is well-formed and ready for live validation when DNS is restored.
```

---

## Final State

| M005 Success Criterion | Status | Evidence |
|---|---|---|
| DNS records configured | ❌ BLOCKED | NXDOMAIN for all subdomains — no A/MX/CNAME/SRV/TXT records |
| TLS certificates valid | ❌ BLOCKED | Cannot provision — depends on DNS |
| WireGuard tunnel active | ❌ BLOCKED | Cannot verify — relay IP unknown |
| Mail round-trip delivery | ❌ BLOCKED | Cannot test — DNS and tunnel required |
| iOS/macOS device onboarding | ❌ BLOCKED | Cannot test — profile server unreachable |
| Desktop client (Thunderbird) | ❌ BLOCKED | Cannot test — autoconfig/IMAP/SMTP unreachable |
| Webmail access | ❌ BLOCKED | Cannot test — HTTPS unreachable |
| Monitoring dashboard healthy | ❌ BLOCKED | Cannot test — /status endpoint unreachable |
| Validation tooling | ✅ PASS | `validate-device-connectivity.sh` dry-run 8/8; `validate-infrastructure.sh` exists with full section coverage |
| Validation report | ✅ PASS | Complete report with all sections, test matrices, re-validation instructions |
| Bug fixes | ✅ PASS | IMAP/SMTP false-positive bug found and fixed |

**Overall M005 Verdict:** ❌ **CANNOT VALIDATE** — All external connectivity criteria blocked by DNS NXDOMAIN. Validation tooling and report are complete; re-run after DNS restoration per instructions above.

**Sign-off:** Pending DNS restoration
**Date:** 2026-03-12T15:14Z
