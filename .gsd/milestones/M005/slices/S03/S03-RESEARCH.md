# S03: Device Connectivity — Mobile, Desktop & Webmail — Research

**Date:** 2026-03-12

## Summary

S03 validates that real end-user devices can connect to DarkPipe from external networks. The codebase already has extensive infrastructure for device onboarding: a profile server with .mobileconfig generation, QR code token flow, autoconfig/autodiscover XML endpoints, app password management, a web UI for device management, and a monitoring dashboard with health/status/JSON APIs. The Caddy reverse proxy routes all discovery and profile endpoints through the WireGuard tunnel to the home device.

This slice is primarily a **human-in-the-loop validation exercise**, not a code-writing task. The code exists — we need to prove it works end-to-end from real devices on real external networks. The main work is: (1) create a structured validation script/checklist that sequences all device connectivity tests, (2) execute tests with real devices, (3) document results and any fixes needed.

S01 and S02 are complete (placeholder summaries, but task summaries confirm DNS/TLS/tunnel/ports validated and mail round-trip infrastructure is in place). S03 consumes that proven infrastructure.

## Recommendation

Structure S03 as a validation-and-fix slice with 4-5 tasks:

1. **iOS/macOS device onboarding** — Generate QR code via web UI, scan from phone on cellular, install .mobileconfig, verify email send/receive and CalDAV/CardDAV sync.
2. **Desktop client (Thunderbird) autoconfig** — Connect Thunderbird from external network, verify autoconfig XML auto-discovery works, verify IMAP/SMTP auth, send/receive test.
3. **Webmail external access** — Load webmail over HTTPS from external network, login, send/receive, verify mobile responsiveness.
4. **Monitoring dashboard health check** — Verify dashboard loads from external network showing healthy status, JSON API returns machine-readable status.
5. **Validation report** — Document all test results, any fixes applied, and final state.

Each task should have a dry-run verification mode (confirming endpoints respond) before the human-in-the-loop steps. This maximizes what the agent can automate.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| .mobileconfig generation | `profiles/pkg/mobileconfig/` | Full profile generator with Email+CalDAV+CardDAV payloads, already tested |
| QR code token flow | `profiles/pkg/qrcode/` | Single-use token store, 15-min expiry, crypto/rand 256-bit tokens |
| App password management | `profiles/pkg/apppassword/` | Supports stalwart, dovecot, maddy backends; bcrypt cost 12 |
| Autoconfig XML | `profiles/pkg/autoconfig/` | Mozilla/Thunderbird autoconfig with IMAP 993 + SMTP 587 |
| Autodiscover XML | `profiles/pkg/autodiscover/` | Outlook autodiscover with IMAP + SMTP protocols |
| Web UI for device mgmt | `profiles/cmd/profile-server/webui.go` | Device list, add device, platform-specific instructions, revoke |
| Monitoring dashboard | `monitoring/status/dashboard.go` + `status.html` | Health checks, queue, deliveries, certificates — with JSON API |
| Health checks | `monitoring/health/` | Liveness + readiness endpoints, pluggable check functions |
| Infrastructure validation | `scripts/validate-infrastructure.sh` | DNS, TLS, tunnel, ports, mail — sequenced validation |
| Mail round-trip test | `scripts/test-mail-roundtrip.sh` | Outbound/inbound test helper with log polling and dry-run |
| Caddy routing | `cloud-relay/caddy/Caddyfile` | Routes autoconfig, autodiscover, profile, QR, health, webmail endpoints |

## Existing Code and Patterns

- `home-device/profiles/cmd/profile-server/` — Complete HTTP server handling profile downloads, autoconfig/autodiscover, QR generation, device web UI, health endpoint. Port 8090.
- `home-device/profiles/cmd/profile-server/webui.go` — Web UI with platform-specific onboarding flows (iOS/macOS → QR+download, Android → QR+manual, Thunderbird/Outlook → autodiscovery instructions)
- `home-device/profiles/cmd/profile-server/templates/` — HTML templates: `status.html` (monitoring dashboard), `device_list.html`, `add_device.html`, `add_device_result.html`
- `cloud-relay/caddy/Caddyfile` — Routes `/.well-known/autoconfig/`, `/autodiscover/`, `/profile/*`, `/qr/*`, `/health/*`, `/status` to profile server (8090); default route to webmail (8080)
- `monitoring/status/aggregator.go` — Aggregates health, queue, delivery, certificate status into `SystemStatus` struct with JSON API
- `monitoring/health/checker.go` — Pluggable health check framework with liveness/readiness separation
- `scripts/validate-infrastructure.sh` — Existing orchestrator with `--json`, `--verbose`, `--dry-run` modes — pattern to follow for device validation
- `scripts/test-mail-roundtrip.sh` — Human-in-the-loop mail test with automated pre-flight and log polling — pattern to follow
- `home-device/tests/test-webmail-groupware.sh` — Existing test script for webmail + CalDAV/CardDAV endpoints — can be referenced
- `home-device/.env.example` — Documents all env vars including MAIL_DOMAIN, MAIL_HOSTNAME, ADMIN_EMAIL, ADMIN_PASSWORD, MAIL_SERVER_TYPE

## Constraints

- **Real devices required** — iOS/macOS .mobileconfig installation cannot be automated or simulated; a physical phone on cellular data is needed
- **External network required** — Tests must run from outside the home network (cellular data or coffee-shop WiFi) to prove relay path works
- **S01/S02 must be operational** — DNS, TLS, tunnel, ports, and mail delivery must be working; S01/S02 have placeholder summaries but task summaries confirm completion
- **Caddy auto-HTTPS** — TLS certificates for webmail/autoconfig/autodiscover subdomains are managed by Caddy via Let's Encrypt; must be working before device tests
- **Profile server port 8090** — Not directly exposed to internet; traffic reaches it via Caddy reverse proxy through WireGuard tunnel
- **Unsigned .mobileconfig profiles** — Per decision, profiles are unsigned for v1; iOS will show "unverified" warning during installation — this is expected and must be documented
- **Basic Auth for admin endpoints** — QR generation and device management web UI require admin credentials; monitoring dashboard requires Basic Auth through Caddy
- **CalDAV/CardDAV routing varies by mail server** — Radicale on port 5232 for maddy/postfix-dovecot; Stalwart uses built-in on port 443; Caddyfile has conditional routing
- **Single-use QR tokens** — 15-minute expiry, marked used immediately on validation; if install fails, a new token must be generated

## Common Pitfalls

- **iOS profile installation requires Settings app** — After scanning QR code in Camera app, user must go to Settings → General → VPN & Device Management → install the downloaded profile. This is non-obvious and must be in instructions.
- **STARTTLS vs SSL confusion** — Outgoing SMTP uses port 587 with STARTTLS (not SSL); .mobileconfig correctly sets `OutgoingMailServerUseSSL: false`. Some client UIs label this confusingly.
- **CalDAV/CardDAV principal URL format** — Must match the mail server backend. Radicale uses `/radicale/<user>/` paths; Stalwart uses built-in URLs. Wrong URL = silent sync failure.
- **Thunderbird autoconfig caching** — Thunderbird aggressively caches autoconfig responses. If testing with a changed config, clear Thunderbird's cached provider data.
- **DNS propagation lag** — If autoconfig/autodiscover CNAME or SRV records were recently created, clients may not discover them immediately. Use `dig` to verify before testing.
- **App password vs main password** — Device onboarding generates app passwords via the store. If testing manually, must use the app password (not admin password) for IMAP/SMTP auth, unless testing via the web UI's QR flow which handles this automatically.
- **Monitoring dashboard auto-refresh** — The status.html has `<meta http-equiv="refresh" content="30">` for auto-refresh. This is fine for dashboard viewing but could cause confusion during debugging.

## Open Risks

- **S01/S02 placeholder summaries** — Both dependency slices have doctor-created placeholder summaries instead of real compressed summaries. Task summaries in those slices confirm work was done, but forward intelligence may be incomplete. If something was left in a broken state, it would only be discoverable by re-running infrastructure validation.
- **Unsigned profile warning on iOS** — Users may be alarmed by the "unverified" warning. Acceptance criteria should explicitly acknowledge this as expected for v1.
- **CalDAV/CardDAV sync timing** — iOS syncs calendars/contacts on a schedule (not real-time push). Initial sync may take up to 15 minutes. Verification must account for this delay.
- **Android has no .mobileconfig** — Android users must configure manually or use autoconfig-capable apps (K-9 Mail supports autoconfig, Gmail does not). The web UI provides manual instructions but the experience is degraded vs iOS.
- **Monitoring health checks depend on running services** — If any mail service is down during the dashboard check, it will show degraded/unhealthy. This could be confused with a dashboard bug vs. actual service issue.
- **Port 587/993 behind NAT** — These ports are on the home device accessed via WireGuard tunnel. If the tunnel drops during testing, IMAP/SMTP connections will timeout with no clear error for the end user.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Email/IMAP client config | (searched: "email client configuration IMAP") | none directly relevant — results were auth0, himalaya CLI, database sync |
| Apple .mobileconfig | (searched: "Apple mobileconfig") | none found — search timed out with only unrelated results |
| Go HTTP servers | developing-ios-apps | installed (but for iOS apps, not Go servers) |
| QA testing | qa-expert | installed — potentially useful for structuring the validation checklist |

No external skills are needed for this slice. The work is validation of existing code against real infrastructure, not building new technology integrations.

## Sources

- Profile server handles all device onboarding flows (source: `home-device/profiles/cmd/profile-server/handlers.go`, `webui.go`)
- .mobileconfig generation with Email+CalDAV+CardDAV (source: `home-device/profiles/pkg/mobileconfig/generator.go`)
- Autoconfig/autodiscover XML for Thunderbird/Outlook (source: `home-device/profiles/pkg/autoconfig/autoconfig.go`, `home-device/profiles/pkg/autodiscover/autodiscover.go`)
- Caddy routes all discovery/profile traffic through WireGuard tunnel (source: `cloud-relay/caddy/Caddyfile`)
- Monitoring dashboard with accessible HTML, JSON API, health endpoints (source: `monitoring/status/dashboard.go`, `monitoring/health/server.go`, template `status.html`)
- Infrastructure validation pattern with --json/--verbose/--dry-run (source: `scripts/validate-infrastructure.sh`)
- Round-trip test helper with human-in-the-loop steps (source: `scripts/test-mail-roundtrip.sh`, `docs/validation/mail-roundtrip.md`)
- Existing webmail+groupware test script (source: `home-device/tests/test-webmail-groupware.sh`)
- Unsigned profiles decision (source: `.gsd/DECISIONS.md` — "Apple profiles are UNSIGNED for v1")
- QR token 15-min expiry, single-use (source: `.gsd/DECISIONS.md`)
