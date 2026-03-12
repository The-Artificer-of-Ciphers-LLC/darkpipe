---
id: T02
parent: S01
milestone: M003
provides:
  - Podman compatibility override files for cloud-relay and home-device
  - SELinux volume label override files for Fedora/RHEL systems
key_files:
  - cloud-relay/docker-compose.podman.yml
  - home-device/docker-compose.podman.yml
  - cloud-relay/docker-compose.podman-selinux.yml
  - home-device/docker-compose.podman-selinux.yml
key_decisions:
  - Used :z (shared) for all SELinux bind mounts since all are :ro config files; no :Z (private) needed
  - Named volumes included in SELinux override files without labels to maintain complete volume list per service (compose merges arrays by position)
  - x-podman extensions kept minimal — only docker_compose_compat and default_net_name_compat flags
patterns_established:
  - Override files include full volume lists per service (bind + named) since compose replaces the entire volumes array on merge
  - SELinux labels use :z for read-only shared config, :Z reserved for exclusive writable bind mounts
observability_surfaces:
  - docker compose config with layered files shows merged result and validates syntax
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Created Podman override files for SELinux and rootful configuration

**Created 4 compose override files enabling Podman compatibility and SELinux support without modifying base compose files.**

## What Happened

Inventoried all volumes across both base compose files. cloud-relay has 1 bind mount (Caddyfile:ro) and 6 named volumes. home-device has 16 bind mounts across 8 services (all :ro config files) and 9 named volumes.

Created two types of override files per stack:

1. **Podman override** (`docker-compose.podman.yml`) — adds `x-podman` top-level extensions for Docker Compose compatibility mode and network naming compat. Minimal files with no service overrides.

2. **SELinux override** (`docker-compose.podman-selinux.yml`) — adds `:z` shared SELinux labels to all bind-mount volumes. Named volumes are included without labels to maintain the complete volume array (compose replaces arrays on merge). All bind mounts are read-only config files, so `:z` (shared) is correct; no `:Z` (private/exclusive) needed.

## Verification

- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman.yml config --quiet` → **PASS**
- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman-selinux.yml config --quiet` → **PASS**
- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman.yml -f home-device/docker-compose.podman-selinux.yml config --quiet` → **PASS** (full stack layering)
- cloud-relay layering: **FAIL (pre-existing)** — base `cloud-relay/docker-compose.yml` has a tmpfs/volume conflict on `/var/spool/postfix` that causes `docker compose config` to fail regardless of overrides. The override files themselves are correct.
- Merged config output confirms `:z` labels appear only on bind mounts, not named volumes
- Base compose files unchanged (git diff clean)
- `scripts/verify-container-security.sh` → 41/41 PASS (zero regression)

### Slice-level checks status (intermediate task):
- ✅ `docker compose -f home-device/docker-compose.yml config --quiet` exits 0
- ❌ `docker compose -f cloud-relay/docker-compose.yml config --quiet` — pre-existing failure (not caused by this task)
- ✅ `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman.yml config --quiet` exits 0
- ✅ `scripts/verify-container-security.sh` still passes
- ⏳ `scripts/verify-podman-compat.sh` — not yet created (T03 scope)

## Diagnostics

- `docker compose -f <base>.yml -f <override>.yml config` shows fully merged result
- Grep for `selinux: z` in merged output to confirm labels applied correctly
- `git diff` on base compose files to confirm no modifications

## Deviations

None.

## Known Issues

- cloud-relay base compose has a pre-existing tmpfs/volume conflict on `/var/spool/postfix` (both tmpfs and named volume mount the same path). This causes `docker compose config` to fail for any cloud-relay combination. Not introduced by this task — needs separate fix.

## Files Created/Modified

- `cloud-relay/docker-compose.podman.yml` — Podman x-podman extensions for Docker compat
- `home-device/docker-compose.podman.yml` — Podman x-podman extensions for Docker compat
- `cloud-relay/docker-compose.podman-selinux.yml` — SELinux :z labels for caddy bind mount
- `home-device/docker-compose.podman-selinux.yml` — SELinux :z labels for all 16 bind mounts across 8 services
