---
estimated_steps: 5
estimated_files: 4
---

# T02: Create Podman override files for SELinux and rootful configuration

**Slice:** S01 — Podman Compose Compatibility
**Milestone:** M003

## Description

Create compose override files that Podman users layer on top of the base compose files. Two types: (1) general Podman overrides with `x-podman` extensions for Docker compatibility flags, and (2) SELinux-specific overrides that add `:Z` volume labels for Fedora/RHEL. These keep the base compose files unchanged and Docker-compatible while enabling Podman-specific features through standard compose file layering.

## Steps

1. Read both base compose files to inventory all bind-mount volumes (candidates for `:Z` labels) vs named volumes (no `:Z` needed)
2. Create `cloud-relay/docker-compose.podman.yml` — add `x-podman` top-level extensions (`docker_compose_compat: true`, `default_net_name_compat: true`). No service overrides needed unless specific Podman tweaks emerge.
3. Create `home-device/docker-compose.podman.yml` — same `x-podman` extensions. Note: `extra_hosts: host-gateway` works on Podman 5.3.0+ with no override needed.
4. Create `cloud-relay/docker-compose.podman-selinux.yml` — override bind-mount volumes with `:z` (shared) for read-only config mounts. Only affects bind mounts, not named volumes.
5. Create `home-device/docker-compose.podman-selinux.yml` — same pattern for home-device bind mounts. Use `:z` for `:ro` config files, `:Z` for exclusive writable bind mounts (if any).

## Must-Haves

- [ ] Podman override files exist for both cloud-relay and home-device
- [ ] SELinux override files add `:z`/`:Z` only to bind-mount volumes (not named volumes)
- [ ] Override files layer cleanly: `docker compose -f base.yml -f override.yml config` passes
- [ ] Base compose files are NOT modified
- [ ] Override files are minimal — only what differs from base

## Verification

- `docker compose -f cloud-relay/docker-compose.yml -f cloud-relay/docker-compose.podman.yml config --quiet` exits 0
- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman.yml config --quiet` exits 0
- `docker compose -f cloud-relay/docker-compose.yml -f cloud-relay/docker-compose.podman-selinux.yml config --quiet` exits 0
- `docker compose -f home-device/docker-compose.yml -f home-device/docker-compose.podman-selinux.yml config --quiet` exits 0
- Full stack layering: `docker compose -f docker-compose.yml -f docker-compose.podman.yml -f docker-compose.podman-selinux.yml config` passes

## Observability Impact

- Signals added/changed: None (override files are static configuration)
- How a future agent inspects this: `docker compose config` with layered files shows merged result
- Failure state exposed: None

## Inputs

- `cloud-relay/docker-compose.yml` — base cloud relay compose (bind mounts: Caddyfile:ro, relay config files:ro)
- `home-device/docker-compose.yml` — base home device compose (bind mounts: config.toml:ro, maddy.conf:ro, setup scripts:ro, etc.)
- T01 output — health checks and version field already fixed

## Expected Output

- `cloud-relay/docker-compose.podman.yml` — Podman compatibility extensions
- `home-device/docker-compose.podman.yml` — Podman compatibility extensions
- `cloud-relay/docker-compose.podman-selinux.yml` — SELinux volume labels for cloud-relay bind mounts
- `home-device/docker-compose.podman-selinux.yml` — SELinux volume labels for home-device bind mounts
