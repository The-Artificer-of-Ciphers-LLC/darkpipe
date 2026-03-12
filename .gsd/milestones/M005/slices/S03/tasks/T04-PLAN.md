---
estimated_steps: 5
estimated_files: 1
---

# T04: Validate Thunderbird autoconfig and webmail from external network

**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

## Description

Validate desktop mail client (Thunderbird) connectivity via autoconfig and webmail access from an external network. Thunderbird should auto-discover IMAP/SMTP settings and authenticate with an app password. Webmail should load over HTTPS with a trusted certificate and support login, send, and receive.

## Steps

1. **Thunderbird autoconfig**: From a machine on an external network, open Thunderbird and add a new email account. Enter the user's email address and app password. Verify that Thunderbird auto-discovers settings from the autoconfig XML endpoint (IMAP 993 SSL + SMTP 587 STARTTLS). If autoconfig doesn't trigger, check DNS for autoconfig CNAME and try manual account setup. Document discovery results.
2. **Thunderbird send/receive**: With the account configured, send a test email to an external address. Send an email from the external address back. Verify both arrive. Document send/receive results with timestamps.
3. **Webmail HTTPS**: From an external network, navigate to `https://mail.<domain>/` in a browser. Verify: HTTPS loads without certificate warnings, page renders login UI (Roundcube or SnappyMail). Document TLS status and page load results.
4. **Webmail send/receive**: Log in with user credentials. Compose and send a test email to an external address. Send an email from the external address and verify it appears in webmail inbox (may need manual refresh). Document login and send/receive results.
5. **Webmail mobile responsiveness**: Resize browser to mobile viewport (or use phone browser) and verify the webmail UI is usable (Roundcube Elastic skin or SnappyMail responsive layout per decision). Document responsiveness observation. Record all results in the Thunderbird and Webmail sections of the validation report.

## Must-Haves

- [ ] Thunderbird auto-discovers IMAP/SMTP settings via autoconfig
- [ ] Thunderbird authenticates and sends/receives email from external network
- [ ] Webmail loads over HTTPS with trusted certificate (no warnings)
- [ ] Webmail login, send, and receive work from external network
- [ ] All results documented in report

## Verification

- Thunderbird and Webmail sections of `docs/validation/device-connectivity-report.md` have pass/fail for: autoconfig discovery, IMAP/SMTP auth, send, receive, webmail HTTPS, webmail login, webmail send/receive, mobile responsiveness
- At minimum, Thunderbird IMAP/SMTP and webmail HTTPS+login must pass

## Observability Impact

- Signals added/changed: None (uses existing autoconfig, webmail, mail server infrastructure)
- How a future agent inspects this: read Thunderbird and Webmail sections of `docs/validation/device-connectivity-report.md`
- Failure state exposed: each test step documented with results; Thunderbird autoconfig failures include DNS dig output and XML response; webmail failures include HTTP status and certificate details

## Inputs

- `docs/validation/device-connectivity-report.md` — from T02/T03 (previous sections populated)
- Thunderbird on a machine on external network
- Browser on external network
- External email account for send/receive testing
- App password for mail account authentication

## Expected Output

- `docs/validation/device-connectivity-report.md` — Thunderbird and Webmail sections populated with complete test results
