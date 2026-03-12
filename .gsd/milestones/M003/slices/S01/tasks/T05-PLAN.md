---
estimated_steps: 3
estimated_files: 3
---

# T05: Run full verification suite and fix any regressions

**Slice:** S01 — Podman Compose Compatibility
**Milestone:** M003

## Description

Final integration gate — run both `verify-podman-compat.sh` and `verify-container-security.sh` together. Fix any issues found. This proves the slice is complete: compose files are dual-compatible AND the M002 security hardening hasn't regressed.

## Steps

1. Run `bash scripts/verify-podman-compat.sh` — all checks must pass. Fix any failures.
2. Run `bash scripts/verify-container-security.sh` — all checks must pass. Fix any regressions caused by T01-T04 changes.
3. Run `docker compose config` on all compose file combinations: base only, base+podman, base+podman+selinux for both cloud-relay and home-device. Verify each produces valid output.

## Must-Haves

- [ ] `verify-podman-compat.sh` passes with 0 failures
- [ ] `verify-container-security.sh` passes with 0 failures
- [ ] All compose file layering combinations produce valid config
- [ ] No changes to existing functionality or security posture

## Verification

- `bash scripts/verify-podman-compat.sh && bash scripts/verify-container-security.sh` exits 0
- `docker compose -f cloud-relay/docker-compose.yml -f cloud-relay/docker-compose.podman.yml -f cloud-relay/docker-compose.podman-selinux.yml config --quiet` exits 0
- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman.yml -f home-device/docker-compose.podman-selinux.yml config --quiet` exits 0

## Observability Impact

- Signals added/changed: None (verification only)
- How a future agent inspects this: Rerun both scripts
- Failure state exposed: Any regression prints specific failing check

## Inputs

- T01-T04 outputs — all modified and new files
- `scripts/verify-container-security.sh` — existing security verification (must not regress)
- `scripts/verify-podman-compat.sh` — new Podman compatibility verification from T03

## Expected Output

- All verification scripts passing — slice is complete
- Any regression fixes applied to compose files or scripts
