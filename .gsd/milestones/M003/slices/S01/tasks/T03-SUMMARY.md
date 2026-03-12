---
id: T03
parent: S01
milestone: M003
provides:
  - Podman compatibility verification script with structured PASS/FAIL output
  - 7 check categories covering compose validation, health check syntax, version field, override layering, Swarm directives, podman-compose
key_files:
  - scripts/verify-podman-compat.sh
key_decisions:
  - Graceful degradation when docker or podman-compose not installed (skip, not fail) — same pattern as podman-compose check
  - Scans all 7 compose files (3 base + 4 override) for static checks; runtime validation only for base files
patterns_established:
  - Verification scripts skip tool-dependent checks with SKIP counter when the tool is absent — never fails on missing optional tooling
observability_surfaces:
  - Run `bash scripts/verify-podman-compat.sh` — structured PASS/FAIL per check with file:line context on failures, summary count at end
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T03: Build Podman compatibility verification script

**Created `scripts/verify-podman-compat.sh` — structured PASS/FAIL verification across 7 check categories for dual Docker/Podman compose compatibility.**

## What Happened

Built a verification script following the exact output pattern of `scripts/verify-container-security.sh` (pass/fail counters, emoji, summary line). The script checks 7 categories:

1. **Compose validation** — `docker compose config --quiet` on all 3 base files (skips if docker not installed)
2. **Health check syntax** — no CMD-form arrays with literal `||` across all 7 compose files
3. **Version field** — no deprecated `version:` field in any compose file
4. **Podman override layering** — base + podman override passes `docker compose config` (skips if docker not installed)
5. **SELinux override layering** — base + selinux override passes `docker compose config` (skips if docker not installed)
6. **Swarm directives** — no `deploy.mode`, `deploy.replicas`, `deploy.placement` in base files
7. **Podman-compose validation** — `podman-compose config` on base files (skips if podman-compose not installed)

All runtime-dependent checks (docker, podman-compose) degrade gracefully with SKIP status rather than failing, making the script usable in any environment.

## Verification

- `bash scripts/verify-podman-compat.sh` → exit 0, 17/17 pass, 4 skip (docker/podman-compose not installed on dev machine)
- Injected `version: "3.8"` into `cloud-relay/docker-compose.yml` → script exits 1 with `FAIL: cloud-relay/docker-compose.yml:1: deprecated version field present` → reverted
- `bash scripts/verify-container-security.sh` → exit 0, 41/41 pass (no regression)

## Diagnostics

- Run `bash scripts/verify-podman-compat.sh` — self-documenting output with PASS/FAIL/SKIP per check
- On failure: prints file path, line number, and specific field that failed
- Summary line: `Total: N | Pass: N | Fail: N | Skip: N`

### Slice-level verification status

| Check | Status |
|-------|--------|
| `scripts/verify-podman-compat.sh` exits 0 | ✅ pass (static checks; runtime checks skipped — docker not installed) |
| `docker compose -f cloud-relay/docker-compose.yml config --quiet` | ⏭️ skip (docker not installed) |
| `docker compose -f home-device/docker-compose.yml config --quiet` | ⏭️ skip (docker not installed) |
| `docker compose ... -f docker-compose.podman.yml config --quiet` | ⏭️ skip (docker not installed) |
| `scripts/verify-container-security.sh` still passes | ✅ pass |
| Health checks use CMD-SHELL form | ✅ pass (verified by script) |

## Deviations

- Task plan mentioned an `extra_hosts: host-gateway` check category but the T03-PLAN steps did not include it as a required check. Omitted from this task — it's a documentation concern better suited for T04 (PODMAN.md).

## Known Issues

- Runtime validation checks (docker compose config, podman-compose config, override layering) can only run in environments with docker/podman installed. The static checks (health check syntax, version field, Swarm directives) always run.

## Files Created/Modified

- `scripts/verify-podman-compat.sh` — new executable verification script (7 check categories, PASS/FAIL/SKIP output)
