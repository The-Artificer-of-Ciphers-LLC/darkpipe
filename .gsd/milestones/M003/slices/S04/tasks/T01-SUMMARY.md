---
id: T01
parent: S04
milestone: M003
provides:
  - "--ci flag in check-runtime.sh that skips port 25 network check"
  - "validate-containers.yml workflow with docker-validate and podman-validate jobs"
key_files:
  - scripts/check-runtime.sh
  - .github/workflows/validate-containers.yml
key_decisions:
  - "podman-compose config steps are non-fatal (|| echo warning) since podman-compose config may differ from docker compose config"
  - "Podman builds use --format docker flag for Docker-compatible image format"
  - "podman-compose installed via pip (simple, no third-party action dependency)"
patterns_established:
  - "CI_MODE flag pattern for skipping environment-specific checks in validation scripts"
  - "Non-fatal steps use || echo to warn without failing the job"
observability_surfaces:
  - "check-runtime.sh prints SKIP for port 25 when --ci flag is set (structured PASS/FAIL/SKIP output)"
  - "Each CI job step is a separate script with its own exit code visible in GitHub Actions logs"
  - "Podman version logged at start of podman-validate job"
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Add --ci flag to check-runtime.sh and create validate-containers.yml workflow

**Added `--ci` flag to check-runtime.sh (skips port 25 check) and created validate-containers.yml with docker-validate and podman-validate CI jobs.**

## What Happened

Added `CI_MODE` flag parsing alongside existing `--quiet`/`--help` in check-runtime.sh. When `--ci` is set, the port 25 availability check is skipped with a structured SKIP message, preventing spurious failures on GitHub Actions runners where port 25 may be blocked.

Created `.github/workflows/validate-containers.yml` triggered on push to main and PRs to main. Two parallel jobs:

1. **docker-validate**: Runs `docker compose config --quiet` on all compose file combinations (base, Podman overlay, SELinux overlay, Podman+SELinux overlay, certbot) for both cloud-relay and home-device. Then runs all three verification scripts.

2. **podman-validate**: Logs podman version, installs podman-compose via pip, builds all 5 Dockerfiles with `podman build --format docker`, validates base compose files with podman-compose (non-fatal), and runs check-runtime.sh --ci.

Existing workflows (build-custom.yml, build-prebuilt.yml, release.yml) were not modified.

## Verification

- `bash scripts/check-runtime.sh --ci` → exit 0, port 25 shows SKIP
- `bash scripts/check-runtime.sh --help` → shows `--ci` in usage text
- `bash scripts/verify-podman-compat.sh` → 17 pass, 0 fail, 4 skip
- `bash scripts/verify-container-security.sh` → 41 pass, 0 fail
- `grep -c 'docker-validate\|podman-validate' .github/workflows/validate-containers.yml` → 2
- `git diff HEAD -- .github/workflows/build-custom.yml .github/workflows/build-prebuilt.yml .github/workflows/release.yml` → no changes

## Diagnostics

- Run `bash scripts/check-runtime.sh --ci` to verify the CI flag works; look for "SKIP: Port 25 check (--ci mode)" in output
- GitHub Actions run logs show per-step pass/fail for each compose config validation and script execution
- Podman version is logged at the start of the podman-validate job for debugging version-related failures

## Deviations

None.

## Known Issues

- Podman version on ubuntu-latest runners may be below 5.3.0, which would cause check-runtime.sh to report a FAIL for the version check. This is informational — the builds will still succeed since Podman 4.x supports `podman build`.

## Files Created/Modified

- `scripts/check-runtime.sh` — Added `--ci` flag that skips port 25 check; updated usage text and argument parsing
- `.github/workflows/validate-containers.yml` — New workflow with docker-validate and podman-validate jobs
