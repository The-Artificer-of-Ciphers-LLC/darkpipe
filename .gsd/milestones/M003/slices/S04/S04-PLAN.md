# S04: CI & Regression Validation

**Goal:** GitHub Actions validates container compatibility on every push/PR — Podman builds pass, Docker compose validates, and existing CI is untouched.
**Demo:** Push a commit → `validate-containers.yml` runs two jobs (docker-validate, podman-validate) → both green. Existing `build-custom.yml`, `build-prebuilt.yml`, `release.yml` are unmodified.

## Must-Haves

- New workflow `.github/workflows/validate-containers.yml` with `docker-validate` and `podman-validate` jobs
- Docker job: `docker compose config` on all compose files, `verify-podman-compat.sh`, `verify-container-security.sh`, `check-runtime.sh`
- Podman job: `podman-compose` installed, `podman build` on all 5 Dockerfiles, `podman-compose config` on base files, `check-runtime.sh`
- `check-runtime.sh` gains `--ci` flag to skip port 25 check (spurious failure in CI)
- Zero changes to existing workflow files (`build-custom.yml`, `build-prebuilt.yml`, `release.yml`)
- Workflow triggers on push to main and on pull requests

## Proof Level

- This slice proves: contract (compose config validation, Dockerfile build, script execution — no runtime containers)
- Real runtime required: no (CI runners validate builds and config, not live services)
- Human/UAT required: no

## Verification

- `bash scripts/check-runtime.sh --ci` exits 0 on a machine with Docker installed (skips port 25)
- `bash scripts/check-runtime.sh --ci` exits 0 with `--ci` flag even when port 25 is in use
- `actionlint .github/workflows/validate-containers.yml` passes (if actionlint available) OR manual YAML structure review
- `grep -c 'validate-containers' .github/workflows/validate-containers.yml` confirms file exists
- `diff .github/workflows/build-custom.yml` / `build-prebuilt.yml` / `release.yml` show zero changes vs main
- All 3 verification scripts run locally without error: `verify-podman-compat.sh`, `verify-container-security.sh`, `check-runtime.sh --ci`

## Observability / Diagnostics

- Runtime signals: GitHub Actions job logs with PASS/FAIL/SKIP output from each script
- Inspection surfaces: Workflow run summary in GitHub Actions UI; each script prints structured summary line
- Failure visibility: Scripts exit non-zero on failure with specific check name and remediation hint; Podman version logged in job output
- Redaction constraints: None (no secrets in validation-only workflow)

## Integration Closure

- Upstream surfaces consumed: `scripts/verify-podman-compat.sh` (S01), `scripts/verify-container-security.sh` (S01), `scripts/check-runtime.sh` (S02), all compose files and Dockerfiles (S01)
- New wiring introduced in this slice: GitHub Actions workflow that runs validation scripts on CI events
- What remains before the milestone is truly usable end-to-end: nothing — this is the final slice of M003

## Tasks

- [x] **T01: Add --ci flag to check-runtime.sh and create validate-containers.yml workflow** `est:45m`
  - Why: check-runtime.sh's port 25 check will spuriously fail on CI runners; the workflow file is the core deliverable of this slice
  - Files: `scripts/check-runtime.sh`, `.github/workflows/validate-containers.yml`
  - Do: Add `--ci` flag to check-runtime.sh that skips port 25 network check. Create validate-containers.yml with two jobs: `docker-validate` (runs compose config, verify-podman-compat.sh, verify-container-security.sh, check-runtime.sh --ci) and `podman-validate` (installs podman-compose, runs podman build on all 5 Dockerfiles with --format docker, runs podman-compose config, runs check-runtime.sh --ci). Both jobs run on ubuntu-latest. Trigger on push to main and PRs. Do NOT modify existing workflows.
  - Verify: `bash scripts/check-runtime.sh --ci` exits 0 locally; `cat .github/workflows/validate-containers.yml` shows both jobs; `md5 .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml` unchanged
  - Done when: Workflow file exists with correct structure, check-runtime.sh --ci skips port 25, existing workflows untouched

- [x] **T02: Local validation and workflow correctness audit** `est:20m`
  - Why: Catch syntax errors, missing paths, and script failures before the workflow runs on GitHub
  - Files: `.github/workflows/validate-containers.yml`, `scripts/check-runtime.sh`, `scripts/verify-podman-compat.sh`, `scripts/verify-container-security.sh`
  - Do: Run actionlint (if available) on the new workflow. Run all three scripts locally with --ci/appropriate flags. Verify YAML structure (job names, step ordering, correct file paths for Dockerfiles and compose files). Verify podman-compose install step syntax. Fix any issues found.
  - Verify: All scripts exit 0 locally; workflow YAML is valid; no references to nonexistent files
  - Done when: All local validation passes, workflow is syntactically correct, and a commit is ready for push

## Files Likely Touched

- `scripts/check-runtime.sh`
- `.github/workflows/validate-containers.yml`
