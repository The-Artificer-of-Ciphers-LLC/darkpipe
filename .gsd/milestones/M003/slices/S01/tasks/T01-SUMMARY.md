---
id: T01
parent: S01
milestone: M003
provides:
  - CMD-SHELL health checks across all compose files
  - version field removed from all compose files
key_files:
  - cloud-relay/docker-compose.yml
  - home-device/docker-compose.yml
  - cloud-relay/certbot/docker-compose.certbot.yml
key_decisions:
  - Converted only health checks that had literal || in CMD arrays; left correct CMD-form checks (nc, redis-cli) unchanged since they work correctly without shell interpretation
patterns_established:
  - Health checks that need shell operators (||, &&, |) must use CMD-SHELL form
observability_surfaces:
  - none
duration: 1 step
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Fix health check CMD form bugs and remove version field

**Converted 5 health checks from buggy CMD form to CMD-SHELL and removed version field from all 3 compose files.**

## What Happened

Identified 5 health checks across 2 compose files that used `["CMD", ..., "||", "exit", "1"]` — passing `||` as a literal argument instead of a shell operator. Converted each to `["CMD-SHELL", "command || exit 1"]` form so the `||` is interpreted by the shell.

Removed `version: '3.8'` from all 3 compose files. This field is obsolete in Docker Compose v2 and triggers deprecation warnings in podman-compose.

Health checks that already used correct CMD form without `||` (nc, redis-cli) were left unchanged — no functional change to non-health-check directives.

## Verification

- `grep -rn '"||"'` across all 3 compose files → empty (exit 1) ✓
- `grep -rn "^version:"` across all 3 compose files → empty (exit 1) ✓
- `grep CMD-SHELL` confirms 1 in cloud-relay, 4 in home-device ✓
- `docker-compose -f home-device/docker-compose.yml config --quiet` → exit 0 ✓
- `docker-compose -f cloud-relay/docker-compose.yml config --quiet` → exit 1 (pre-existing tmpfs/volume conflict on relay service `/var/spool/postfix`, confirmed identical on main branch before our changes)
- `docker-compose -f cloud-relay/certbot/docker-compose.certbot.yml -f cloud-relay/docker-compose.yml config --quiet` → exit 1 (same pre-existing relay issue)

## Diagnostics

- `grep CMD-SHELL` on compose files to verify correct health check form
- `grep "^version:"` to verify no version fields remain

## Deviations

The cloud-relay compose config validation fails due to a pre-existing tmpfs/volume conflict (`/var/spool/postfix` mounted as both tmpfs and named volume). Confirmed this exists on the branch baseline before our changes. This is not a regression from T01.

## Known Issues

- `cloud-relay/docker-compose.yml` has a pre-existing `docker-compose config` failure: relay service mounts `/var/spool/postfix` as both tmpfs and a named volume. This predates T01 and should be addressed in a separate task.

## Files Created/Modified

- `cloud-relay/docker-compose.yml` — converted caddy health check to CMD-SHELL, removed version field
- `home-device/docker-compose.yml` — converted 4 health checks (roundcube, snappymail, radicale, profile-server) to CMD-SHELL, removed version field
- `cloud-relay/certbot/docker-compose.certbot.yml` — removed version field (no health checks present)
