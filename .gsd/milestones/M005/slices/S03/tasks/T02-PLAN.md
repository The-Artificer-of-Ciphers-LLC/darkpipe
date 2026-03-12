---
estimated_steps: 4
estimated_files: 2
---

# T02: Create validation report template and run automated pre-checks

**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

## Description

Create the persistent validation report at `docs/validation/device-connectivity-report.md` with structured sections for every test category. Run the automated validation script against live infrastructure and populate the pre-check section. If any endpoint fails, diagnose and fix before proceeding to human testing.

## Steps

1. Create `docs/validation/device-connectivity-report.md` with sections: Executive Summary, Automated Pre-Checks, iOS/macOS Device Onboarding, Desktop Client (Thunderbird), Webmail Access, Monitoring Dashboard, Fixes Applied, Final State. Each section has a table or checklist for results.
2. Run `scripts/validate-device-connectivity.sh --verbose` against the live infrastructure. Capture output.
3. Populate the Automated Pre-Checks section with results for each endpoint (autoconfig, autodiscover, profile server health, webmail, monitoring dashboard, monitoring JSON, IMAP TLS, SMTP STARTTLS). Record pass/fail, response details, and timestamp.
4. If any check fails: diagnose the root cause (DNS, Caddy routing, tunnel, service health), apply fix, re-run validation, and document the fix in the Fixes Applied section. Iterate until all automated checks pass.

## Must-Haves

- [ ] Report template created with all required sections and result placeholders
- [ ] Automated pre-checks section populated with real results from live infrastructure
- [ ] All 8 automated endpoint checks pass (or failures are documented with applied fixes)
- [ ] Report timestamps show when checks were run

## Verification

- `docs/validation/device-connectivity-report.md` exists and has all section headers
- Automated Pre-Checks section has pass/fail for all 8 endpoints
- `scripts/validate-device-connectivity.sh` exits 0 against live infrastructure

## Observability Impact

- Signals added/changed: None (script from T01 already provides signals)
- How a future agent inspects this: read `docs/validation/device-connectivity-report.md` for persistent test state; re-run `scripts/validate-device-connectivity.sh` for current state
- Failure state exposed: Fixes Applied section documents what broke and how it was resolved

## Inputs

- `scripts/validate-device-connectivity.sh` — from T01
- Live DarkPipe deployment — S01 infrastructure + S02 mail delivery must be operational

## Expected Output

- `docs/validation/device-connectivity-report.md` — report template with automated pre-check section populated with live results
