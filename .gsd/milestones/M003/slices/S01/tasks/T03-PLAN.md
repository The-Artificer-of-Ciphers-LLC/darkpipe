---
estimated_steps: 5
estimated_files: 2
---

# T03: Build Podman compatibility verification script

**Slice:** S01 — Podman Compose Compatibility
**Milestone:** M003

## Description

Create `scripts/verify-podman-compat.sh` — a structured verification script following the same PASS/FAIL pattern as `scripts/verify-container-security.sh`. This script is the primary proof artifact for S01 and will be consumed by S04 for CI integration. It validates compose file dual-compatibility without requiring Podman to be installed (graceful degradation).

## Steps

1. Read `scripts/verify-container-security.sh` to match the output pattern (pass/fail counters, emoji, summary line)
2. Create `scripts/verify-podman-compat.sh` with these check categories:
   - **Compose validation:** `docker compose config --quiet` passes for all base compose files
   - **Health check syntax:** No CMD-form arrays containing literal `||` strings
   - **Version field:** No `version:` field in compose files (deprecated)
   - **Override layering:** Base + podman override passes `docker compose config`
   - **Override layering:** Base + selinux override passes `docker compose config`
   - **Swarm directives:** No `deploy.mode`, `deploy.replicas`, `deploy.placement` (Swarm-only)
   - **Podman-compose validation:** If `podman-compose` is available, run `podman-compose config` on all base files (skip with note if not installed)
3. Each check: increment PASS or FAIL counter, print result with service/file context
4. Summary line at end: `N/M checks passed` with exit 1 if any failures
5. Mark executable (`chmod +x`)

## Must-Haves

- [ ] Script follows `verify-container-security.sh` output pattern
- [ ] All check categories implemented with clear PASS/FAIL output
- [ ] Graceful skip when `podman-compose` is not installed (not a failure)
- [ ] Script exits 0 on current (post-T01/T02) compose files
- [ ] Script catches intentional breakage (e.g., re-adding `version:` field)
- [ ] Script is idempotent and side-effect-free

## Verification

- `bash scripts/verify-podman-compat.sh` exits 0 with all PASS
- Temporarily add `version: '3.8'` to a compose file → script exits 1 → revert
- `bash scripts/verify-container-security.sh` still passes (no interference)

## Observability Impact

- Signals added/changed: Structured PASS/FAIL output per check, summary count
- How a future agent inspects this: Run `bash scripts/verify-podman-compat.sh` — output is self-documenting
- Failure state exposed: Failing check prints file path, service name, and specific field that failed

## Inputs

- `scripts/verify-container-security.sh` — pattern to follow for output format
- T01 output — compose files with fixed health checks and no version field
- T02 output — override files that must layer cleanly

## Expected Output

- `scripts/verify-podman-compat.sh` — executable verification script, all checks passing
