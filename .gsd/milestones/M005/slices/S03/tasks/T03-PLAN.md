---
estimated_steps: 5
estimated_files: 1
---

# T03: Validate iOS/macOS device onboarding from cellular network

**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

## Description

Human-in-the-loop validation of the iOS/macOS .mobileconfig onboarding flow from a phone on cellular data (not home WiFi). This proves the complete device onboarding path: QR code generation → profile download → profile installation → email send/receive → CalDAV/CardDAV sync — all through the cloud relay tunnel.

## Steps

1. From the profile server web UI (admin-authenticated at `https://mail.<domain>/profile/`), navigate to "Add Device", select iOS/macOS platform, and generate a QR code for a test user.
2. On an iPhone on cellular data (WiFi off): scan the QR code with the Camera app, follow the redirect to download the .mobileconfig profile. Note: the "unverified profile" warning is expected (unsigned profiles per v1 decision).
3. Install the profile via Settings → General → VPN & Device Management → Downloaded Profile → Install. Enter device passcode when prompted. Verify the profile installs successfully with Email, CalDAV, and CardDAV payloads.
4. Test email: open Mail.app, verify the account appears, send a test email to an external address (e.g., Gmail), verify delivery. Send an email from the external address back, verify it arrives in Mail.app on the phone. Record send/receive results with timestamps.
5. Test CalDAV/CardDAV: open Calendar app, verify synced calendar appears (may take up to 15 minutes for initial sync). Open Contacts app, verify synced address book appears. If sync hasn't completed, force a manual refresh and wait. Record sync results. Document all results in the iOS/macOS section of the validation report.

## Must-Haves

- [ ] QR code generated from web UI and scanned on cellular network
- [ ] .mobileconfig profile installed successfully (unverified warning acknowledged)
- [ ] Email send from phone to external address delivered
- [ ] Email receive from external address arrives on phone
- [ ] CalDAV sync confirmed (calendar visible in Calendar app)
- [ ] CardDAV sync confirmed (address book visible in Contacts app)
- [ ] All results documented in report with timestamps

## Verification

- iOS/macOS section of `docs/validation/device-connectivity-report.md` has pass/fail for: profile install, email send, email receive, calendar sync, contacts sync
- At minimum, email send and receive must pass for the slice to succeed

## Observability Impact

- Signals added/changed: None (uses existing profile server, mail server, CalDAV/CardDAV infrastructure)
- How a future agent inspects this: read iOS/macOS section of `docs/validation/device-connectivity-report.md`
- Failure state exposed: each test step's result documented with timestamps; failures include error messages, screenshots, and root cause analysis

## Inputs

- `docs/validation/device-connectivity-report.md` — from T02 (template with pre-checks populated)
- Live profile server web UI at `https://mail.<domain>/profile/`
- iPhone with cellular data (not on home WiFi)
- External email account (Gmail/Outlook) for send/receive testing

## Expected Output

- `docs/validation/device-connectivity-report.md` — iOS/macOS section populated with complete test results
