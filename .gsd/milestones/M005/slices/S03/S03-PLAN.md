# S03: Device Connectivity — Mobile, Desktop & Webmail

**Goal:** Prove that real end-user devices can connect to DarkPipe from external networks — phones on cellular, desktop mail clients, and webmail browsers — using the infrastructure validated in S01 and mail delivery proven in S02.
**Demo:** A phone on cellular data installs a .mobileconfig profile, sends/receives email, and syncs calendar/contacts. Thunderbird connects via autoconfig from an external network. Webmail loads over HTTPS externally. Monitoring dashboard shows all healthy.

## Must-Haves

- Validation script that pre-checks all device connectivity endpoints (autoconfig, autodiscover, profile server, webmail, monitoring) from the agent before human tests
- iOS/macOS .mobileconfig onboarding tested from a phone on cellular data (email send/receive + CalDAV/CardDAV sync)
- Desktop mail client (Thunderbird) connected via autoconfig from external network with IMAP/SMTP verified
- Webmail HTTPS access verified from external network with login and send/receive
- Monitoring dashboard health check verified from external network with JSON API returning healthy status
- Validation report documenting all test results, fixes applied, and final state

## Proof Level

- This slice proves: operational + UAT (real devices on real external networks)
- Real runtime required: yes (live DarkPipe deployment with tunnel active)
- Human/UAT required: yes (iOS profile installation, device email send/receive, webmail login)

## Verification

- `bash scripts/validate-device-connectivity.sh --dry-run` exits 0 (script is well-formed)
- `bash scripts/validate-device-connectivity.sh` exits 0 against live infrastructure (all endpoint pre-checks pass)
- `docs/validation/device-connectivity-report.md` exists with all test sections completed and results recorded
- Monitoring JSON API returns `"overall":"healthy"` or equivalent from external network

## Observability / Diagnostics

- Runtime signals: validate-device-connectivity.sh emits structured pass/fail per endpoint with timestamps; supports `--json` and `--verbose` modes matching existing validate-infrastructure.sh patterns
- Inspection surfaces: `scripts/validate-device-connectivity.sh --json` for machine-readable results; monitoring dashboard `/status` JSON endpoint for service health; validation report as persistent artifact
- Failure visibility: each endpoint check reports URL tested, HTTP status, expected vs actual, and specific error; monitoring JSON API exposes per-service health with timestamps
- Redaction constraints: admin passwords not logged; app passwords shown only as `****` in report; email addresses may appear in test documentation

## Integration Closure

- Upstream surfaces consumed: S01's DNS/TLS/tunnel/ports (`scripts/validate-infrastructure.sh`), S02's proven mail delivery (`scripts/test-mail-roundtrip.sh`), profile server endpoints (port 8090 via Caddy), webmail (port 8080 via Caddy), monitoring dashboard (`/status` endpoint)
- New wiring introduced in this slice: `scripts/validate-device-connectivity.sh` (new validation orchestrator for device endpoints), `docs/validation/device-connectivity-report.md` (persistent test results)
- What remains before the milestone is truly usable end-to-end: nothing — S03 is the final slice; completing it means all M005 success criteria are proven

## Tasks

- [x] **T01: Build device connectivity validation script** `est:1h`
  - Why: Automates pre-flight checks for all device-facing endpoints before human testing — catches broken endpoints early and provides structured pass/fail output for the validation report
  - Files: `scripts/validate-device-connectivity.sh`
  - Do: Create validation script following validate-infrastructure.sh patterns (--json, --verbose, --dry-run). Check: autoconfig XML at `/.well-known/autoconfig/mail/config-v1.1.xml`, autodiscover XML at `/autodiscover/autodiscover.xml`, profile server health at `/health/live`, webmail HTTPS response, monitoring dashboard at `/status`, monitoring JSON API, IMAP port 993 TLS handshake, SMTP port 587 STARTTLS handshake. Each check reports pass/fail with URL, status code, and details.
  - Verify: `bash scripts/validate-device-connectivity.sh --dry-run` exits 0; script validates cleanly with `shellcheck`
  - Done when: script exits 0 in dry-run mode and each endpoint check has clear pass/fail output with structured error reporting

- [x] **T02: Create validation report template and run automated pre-checks** `est:30m`
  - Why: Establishes the persistent test report where all human and automated results are recorded; runs the automated pre-checks against live infrastructure to confirm all endpoints are reachable before human testing begins
  - Files: `docs/validation/device-connectivity-report.md`, `scripts/validate-device-connectivity.sh`
  - Do: Create report template with sections for: automated pre-checks, iOS/macOS onboarding, Thunderbird autoconfig, webmail access, monitoring dashboard, and fixes applied. Run `validate-device-connectivity.sh` against live infrastructure and record results in the report. If any endpoint fails, diagnose and fix before proceeding.
  - Verify: `docs/validation/device-connectivity-report.md` exists with automated pre-check section populated; all automated checks pass or failures are documented with fixes
  - Done when: report exists with pre-check results recorded; all automated endpoint checks pass against live infrastructure

- [x] **T03: Validate iOS/macOS device onboarding from cellular network** `est:45m`
  - Why: Proves the core UAT requirement — a real phone on cellular data can onboard via .mobileconfig and use email, calendar, and contacts through the DarkPipe relay
  - Files: `docs/validation/device-connectivity-report.md`
  - Do: From the web UI (admin-authenticated), generate a QR code for a test device. On an iPhone on cellular data: scan QR with Camera, follow redirect to profile download, install profile via Settings → General → VPN & Device Management (expect "unverified" warning per decision — unsigned v1 profiles). Verify: email account appears in Mail.app, send test email to external address, receive test email from external address, check Calendar app for synced calendar, check Contacts for synced address book. Document each step's result with screenshots/notes in the report. Account for CalDAV/CardDAV sync delay (up to 15 minutes per research).
  - Verify: report iOS section populated with pass/fail for: profile install, email send, email receive, calendar sync, contacts sync
  - Done when: all five iOS checks documented in report — any failures include root cause and fix applied

- [x] **T04: Validate Thunderbird autoconfig and webmail from external network** `est:45m`
  - Why: Proves desktop mail client connectivity and webmail access from external networks — the remaining device connectivity requirements
  - Files: `docs/validation/device-connectivity-report.md`
  - Do: (1) Thunderbird: from a machine on an external network, add account in Thunderbird — verify autoconfig auto-discovers IMAP 993 + SMTP 587 settings, authenticate with app password, send test email to external address, receive test email from external. (2) Webmail: from external network, load `https://mail.<domain>` in browser, verify HTTPS with trusted certificate (no warnings), login with user credentials, send test email, receive test email, verify mobile-responsive layout. Document all results in report.
  - Verify: report Thunderbird and webmail sections populated with pass/fail for: autoconfig discovery, IMAP/SMTP auth, send, receive, webmail HTTPS, webmail login, webmail send/receive
  - Done when: all Thunderbird and webmail checks documented — any failures include root cause and fix applied

- [x] **T05: Validate monitoring dashboard and finalize report** `est:30m`
  - Why: Confirms monitoring shows healthy status during external access (operational verification) and produces the final validation report proving all M005 success criteria
  - Files: `docs/validation/device-connectivity-report.md`
  - Do: (1) From external network, load monitoring dashboard at `/status`, verify it shows all services healthy, verify JSON API returns machine-readable status with healthy overall state. (2) Review the complete validation report — ensure all sections are populated, all failures have documented fixes, and the final state summary confirms M005 success criteria. (3) Run `validate-device-connectivity.sh` one final time to confirm all endpoints still pass. Record final results.
  - Verify: monitoring dashboard loads showing healthy; JSON API returns healthy status; `validate-device-connectivity.sh` exits 0; complete report exists at `docs/validation/device-connectivity-report.md` with all sections populated
  - Done when: monitoring verified healthy, final validation script passes, report is complete with overall pass/fail verdict for M005

## Files Likely Touched

- `scripts/validate-device-connectivity.sh` (new — device endpoint validation orchestrator)
- `docs/validation/device-connectivity-report.md` (new — persistent test results and M005 sign-off)
