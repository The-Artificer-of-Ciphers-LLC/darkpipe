---
estimated_steps: 4
estimated_files: 3
---

# T01: Fix health check CMD form bugs and remove version field

**Slice:** S01 — Podman Compose Compatibility
**Milestone:** M003

## Description

Multiple health checks across compose files use `["CMD", "wget", ..., "||", "exit", "1"]` which passes `||` as a literal argument to wget instead of being interpreted as a shell OR operator. This is a pre-existing bug on both Docker and Podman (the checks work by accident because wget returns non-zero on failure, making the `||` unnecessary — but the literal `||` argument could cause unexpected behavior). Convert all to `CMD-SHELL` form for correctness.

The `version: '3.8'` field is informational-only on Docker Compose v2 but triggers deprecation warnings on podman-compose. Remove it from all compose files.

## Steps

1. Read all three compose files and identify every health check using CMD form with `||`
2. Convert each to CMD-SHELL form: `["CMD-SHELL", "wget --quiet --tries=1 --spider http://... || exit 1"]`
3. Remove the `version: '3.8'` line from `cloud-relay/docker-compose.yml`, `home-device/docker-compose.yml`, and `cloud-relay/certbot/docker-compose.certbot.yml`
4. Validate all compose files with `docker compose config`

## Must-Haves

- [ ] All health checks use CMD-SHELL form (no literal `||` in CMD arrays)
- [ ] No `version:` field in any compose file
- [ ] `docker compose config` passes for all three compose files
- [ ] No functional change to any non-health-check directive

## Verification

- `grep -rn '"||"' cloud-relay/docker-compose.yml home-device/docker-compose.yml cloud-relay/certbot/docker-compose.certbot.yml` returns empty
- `grep -rn "^version:" cloud-relay/docker-compose.yml home-device/docker-compose.yml cloud-relay/certbot/docker-compose.certbot.yml` returns empty
- `docker compose -f cloud-relay/docker-compose.yml config --quiet` exits 0
- `docker compose -f home-device/docker-compose.yml config --quiet` exits 0
- `docker compose -f cloud-relay/certbot/docker-compose.certbot.yml -f cloud-relay/docker-compose.yml config --quiet` exits 0

## Observability Impact

- Signals added/changed: None (health checks already existed, just fixing syntax)
- How a future agent inspects this: `grep CMD-SHELL` on compose files to verify correct form
- Failure state exposed: None

## Inputs

- `cloud-relay/docker-compose.yml` — has 1 CMD-form health check with `||` (caddy)
- `home-device/docker-compose.yml` — has 4 CMD-form health checks with `||` (roundcube, snappymail, radicale, profile-server)
- `cloud-relay/certbot/docker-compose.certbot.yml` — may have health checks to check

## Expected Output

- `cloud-relay/docker-compose.yml` — health checks use CMD-SHELL, version field removed
- `home-device/docker-compose.yml` — health checks use CMD-SHELL, version field removed
- `cloud-relay/certbot/docker-compose.certbot.yml` — version field removed (if present)
