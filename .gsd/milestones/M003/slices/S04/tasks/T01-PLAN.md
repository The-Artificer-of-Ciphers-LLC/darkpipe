---
estimated_steps: 5
estimated_files: 2
---

# T01: Add --ci flag to check-runtime.sh and create validate-containers.yml workflow

**Slice:** S04 — CI & Regression Validation
**Milestone:** M003

## Description

The port 25 availability check in `check-runtime.sh` will cause spurious failures on GitHub Actions runners where port 25 may be blocked or in use by the cloud provider. Add a `--ci` flag that skips network checks. Then create the new `validate-containers.yml` workflow with two parallel jobs: one validating Docker compose compatibility, one validating Podman builds and compose config. This is the core deliverable of the entire slice.

## Steps

1. Read `scripts/check-runtime.sh` fully to understand the argument parsing and check flow
2. Add `--ci` flag support: parse the flag alongside existing `--quiet`/`--help`, and when set, skip the `check_port_25` call. Update the usage text.
3. Verify `bash scripts/check-runtime.sh --ci` runs without error locally
4. Create `.github/workflows/validate-containers.yml` with:
   - Trigger: `push` to main branch, `pull_request` to main branch
   - `docker-validate` job on `ubuntu-latest`: checkout, run `docker compose config` on base compose files + overlay combinations, run `scripts/verify-podman-compat.sh`, run `scripts/verify-container-security.sh`, run `scripts/check-runtime.sh --ci`
   - `podman-validate` job on `ubuntu-latest`: checkout, log `podman version`, install podman-compose via pip, run `podman build --format docker` on all 5 Dockerfiles, run `podman-compose config` on base compose files (non-fatal — podman-compose config may differ), run `scripts/check-runtime.sh --ci`
5. Verify existing workflow files are untouched (diff against git)

## Must-Haves

- [ ] `--ci` flag in check-runtime.sh skips port 25 check
- [ ] `--ci` flag documented in usage/help text
- [ ] validate-containers.yml has `docker-validate` job running all 3 verification scripts
- [ ] validate-containers.yml has `podman-validate` job building all 5 Dockerfiles
- [ ] Workflow triggers on push to main and pull_request
- [ ] Existing workflows (`build-custom.yml`, `build-prebuilt.yml`, `release.yml`) are not modified

## Verification

- `bash scripts/check-runtime.sh --ci` exits 0 locally (Docker detected, port 25 skipped)
- `bash scripts/check-runtime.sh --help` shows --ci in usage text
- `.github/workflows/validate-containers.yml` exists and contains both job names
- `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml` shows no changes

## Observability Impact

- Signals added/changed: `check-runtime.sh` prints SKIP for port 25 when --ci flag is set (structured PASS/FAIL/SKIP output preserved)
- How a future agent inspects this: Run `bash scripts/check-runtime.sh --ci` and check exit code; inspect GitHub Actions run logs for job-level pass/fail
- Failure state exposed: Each CI job step is a separate script with its own exit code; GitHub Actions surfaces which step failed

## Inputs

- `scripts/check-runtime.sh` — existing script from S02 (needs --ci flag)
- `scripts/verify-podman-compat.sh` — existing script from S01 (used as-is)
- `scripts/verify-container-security.sh` — existing script from S01 (used as-is)
- S04 research identifying compose file paths, Dockerfile paths, and CI constraints
- Existing workflow files as reference for project patterns (but NOT to be modified)

## Expected Output

- `scripts/check-runtime.sh` — modified with --ci flag support
- `.github/workflows/validate-containers.yml` — new workflow file with docker-validate and podman-validate jobs
