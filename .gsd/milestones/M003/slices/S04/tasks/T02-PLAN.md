---
estimated_steps: 4
estimated_files: 4
---

# T02: Local validation and workflow correctness audit

**Slice:** S04 — CI & Regression Validation
**Milestone:** M003

## Description

Before pushing, validate that the new workflow is syntactically correct, all referenced file paths exist, all scripts run successfully in CI-compatible mode, and existing workflows remain untouched. Fix any issues found during validation.

## Steps

1. Run `actionlint .github/workflows/validate-containers.yml` if actionlint is available; otherwise manually validate YAML structure (indentation, job/step format, action version pins)
2. Verify every file path referenced in the workflow actually exists: all 5 Dockerfiles, all compose files (base + podman + selinux overrides), all 3 scripts
3. Run all 3 verification scripts locally with CI-appropriate flags: `bash scripts/check-runtime.sh --ci`, `bash scripts/verify-podman-compat.sh`, `bash scripts/verify-container-security.sh` — all must exit 0
4. Confirm zero regression: `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml` shows empty output

## Must-Haves

- [ ] Workflow YAML is valid (passes actionlint or manual structure review)
- [ ] All file paths in workflow exist in the repo
- [ ] All 3 verification scripts exit 0 locally
- [ ] Existing workflows show zero diff

## Verification

- actionlint passes or manual YAML review confirms valid structure
- `bash scripts/check-runtime.sh --ci && echo OK` prints OK
- `bash scripts/verify-podman-compat.sh && echo OK` prints OK
- `bash scripts/verify-container-security.sh && echo OK` prints OK
- `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml | wc -l` outputs 0

## Observability Impact

- Signals added/changed: None — this task validates existing signals work correctly
- How a future agent inspects this: Re-run the same verification commands
- Failure state exposed: None new

## Inputs

- `.github/workflows/validate-containers.yml` — T01 output
- `scripts/check-runtime.sh` — T01 output (with --ci flag)
- `scripts/verify-podman-compat.sh` — existing from S01
- `scripts/verify-container-security.sh` — existing from S01

## Expected Output

- `.github/workflows/validate-containers.yml` — validated and possibly fixed
- `scripts/check-runtime.sh` — validated, possibly minor fixes
- Slice verification passing: all scripts exit 0, workflow valid, zero regression on existing CI
