---
estimated_steps: 4
estimated_files: 1
---

# T05: Validate monitoring dashboard and finalize report

**Slice:** S03 — Device Connectivity — Mobile, Desktop & Webmail
**Milestone:** M005

## Description

Verify the monitoring dashboard shows healthy status from an external network, then finalize the complete validation report as the M005 sign-off artifact. Run one final pass of the automated validation script to confirm all endpoints remain healthy after all device testing.

## Steps

1. **Monitoring dashboard**: From an external network, load `https://mail.<domain>/status` (with Basic Auth). Verify: dashboard renders showing service health status, auto-refresh is working (30-second meta refresh), all monitored services show healthy/green status. Document dashboard state.
2. **Monitoring JSON API**: Fetch `https://mail.<domain>/status?format=json` and verify the response is valid JSON with an overall health status field. Confirm the status indicates healthy. If any service shows degraded, investigate and resolve before finalizing. Document JSON response structure and status.
3. **Final automated validation**: Run `scripts/validate-device-connectivity.sh` one last time to confirm all 8 endpoint checks still pass. Run `scripts/validate-infrastructure.sh` as well for completeness. Record results in the report.
4. **Finalize report**: Complete the Executive Summary section with overall pass/fail verdict. Ensure the Fixes Applied section documents any issues encountered and resolutions across all tasks. Add the Final State section confirming which M005 success criteria are proven. Review the complete report for completeness.

## Must-Haves

- [ ] Monitoring dashboard loads from external network showing healthy status
- [ ] Monitoring JSON API returns valid JSON with healthy overall status
- [ ] Final `validate-device-connectivity.sh` run exits 0
- [ ] Validation report has all sections populated with complete results
- [ ] Executive Summary has overall pass/fail verdict for M005

## Verification

- Monitoring section of `docs/validation/device-connectivity-report.md` has pass/fail for: dashboard load, JSON API health, all-services-healthy
- `scripts/validate-device-connectivity.sh` exits 0
- `docs/validation/device-connectivity-report.md` Executive Summary section exists with overall verdict
- Report covers all M005 success criteria: DNS, TLS, tunnel, mail round-trip, mobile device, desktop client, webmail, monitoring

## Observability Impact

- Signals added/changed: None
- How a future agent inspects this: read complete `docs/validation/device-connectivity-report.md` for M005 validation state; re-run validation scripts for current infrastructure state
- Failure state exposed: report Final State section explicitly lists which M005 criteria pass and which (if any) have caveats

## Inputs

- `docs/validation/device-connectivity-report.md` — from T02/T03/T04 (previous sections populated)
- `scripts/validate-device-connectivity.sh` — from T01
- `scripts/validate-infrastructure.sh` — existing S01 validation script
- Live DarkPipe deployment

## Expected Output

- `docs/validation/device-connectivity-report.md` — complete validation report with all sections populated, monitoring results, executive summary, and M005 sign-off verdict
