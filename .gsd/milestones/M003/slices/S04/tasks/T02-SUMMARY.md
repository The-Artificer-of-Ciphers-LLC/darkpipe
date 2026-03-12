---
id: T02
parent: S04
milestone: M003
provides:
  - "Local validation confirming workflow YAML is syntactically valid"
  - "All referenced file paths verified to exist in repo (5 Dockerfiles, 7 compose files, 3 scripts)"
  - "All 3 verification scripts exit 0 in CI-compatible mode"
  - "Zero regression on existing workflows confirmed"
key_files:
  - .github/workflows/validate-containers.yml
  - scripts/check-runtime.sh
  - scripts/verify-podman-compat.sh
  - scripts/verify-container-security.sh
key_decisions: []
patterns_established: []
observability_surfaces:
  - "Re-run the same verification commands to re-validate"
duration: 10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Local validation and workflow correctness audit

**Validated workflow YAML, all referenced paths, all scripts exit 0, and zero regression on existing CI workflows.**

## What Happened

Ran a four-step validation audit before the workflow is pushed to GitHub:

1. **YAML validation** — Ruby YAML parser confirmed `.github/workflows/validate-containers.yml` is syntactically valid. actionlint was not available locally; manual review confirmed correct job/step structure, action version pins (`actions/checkout@v4`), and consistent indentation.

2. **File path verification** — Confirmed all 15 files referenced in the workflow exist: 5 Dockerfiles (`cloud-relay/Dockerfile`, `home-device/maddy/Dockerfile`, `home-device/postfix-dovecot/Dockerfile`, `home-device/profiles/Dockerfile`, `home-device/stalwart/Dockerfile`), 7 compose files (3 base + 4 overlay), and 3 scripts.

3. **Script execution** — All 3 verification scripts ran successfully:
   - `check-runtime.sh --ci`: 4 checks, 3 pass, 0 fail, 2 skip (port 25 skipped via --ci, SELinux N/A on macOS)
   - `verify-podman-compat.sh`: 17 pass, 0 fail, 4 skip (docker/podman-compose not available locally)
   - `verify-container-security.sh`: 41 pass, 0 fail

4. **Regression check** — `git diff HEAD` on `build-custom.yml`, `build-prebuilt.yml`, `release.yml` shows 0 lines changed.

No issues found — no fixes were needed.

## Verification

- `ruby -ryaml -e "YAML.safe_load(...)"` → "YAML: valid" ✅
- All 15 referenced file paths exist in repo ✅
- `bash scripts/check-runtime.sh --ci && echo OK` → OK ✅
- `bash scripts/verify-podman-compat.sh && echo OK` → OK ✅
- `bash scripts/verify-container-security.sh && echo OK` → OK ✅
- `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml | wc -l` → 0 ✅

### Slice-level verification

- `check-runtime.sh --ci` exits 0, skips port 25 ✅
- YAML structure review passed (no actionlint, manual + parser validation) ✅
- Workflow file exists ✅
- Existing workflows show zero diff ✅
- All 3 verification scripts run locally without error ✅

## Diagnostics

Re-run the same verification commands to re-validate:
- `bash scripts/check-runtime.sh --ci`
- `bash scripts/verify-podman-compat.sh`
- `bash scripts/verify-container-security.sh`
- `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml`

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

No files were modified — this was a validation-only task. All artifacts passed validation as-is.
