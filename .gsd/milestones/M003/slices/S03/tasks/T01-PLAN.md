---
estimated_steps: 5
estimated_files: 1
---

# T01: Create Apple Containers orchestration script with dry-run mode

**Slice:** S03 — Apple Containers Support
**Milestone:** M003

## Description

Apple Containers has no compose equivalent — every container must be started individually via `container run`. This task creates a shell orchestration script that translates the cloud-relay `docker-compose.yml` (2 services: caddy + relay) into equivalent `container` CLI commands with proper networking, volumes, environment variables, and startup ordering. The script includes a `--dry-run` mode so the full command translation can be verified without macOS 26 hardware.

## Steps

1. Create `scripts/apple-containers-start.sh` with argument parsing (`up`, `down`, `status`, `logs` subcommands, `--dry-run`, `--verbose`, `--help` flags)
2. Implement network management: `container network create darkpipe` on `up`, `container network rm darkpipe` on `down`; skip if already exists
3. Implement image building: `container build` for relay (context: repo root, dockerfile: cloud-relay/Dockerfile); caddy uses `caddy:2-alpine` pulled image
4. Implement `container run` commands for both services, translating from docker-compose.yml:
   - **caddy**: ports 80/443/443udp, volumes (bind mounts for Caddyfile, host dirs for data/config/logs), env vars (WEBMAIL_DOMAINS, AUTOCONFIG_DOMAINS, AUTODISCOVER_DOMAINS), `--memory 128M`, `--network darkpipe`, `--read-only`, `--name caddy`, detached
   - **relay**: port 25, env vars from `.env` file (RELAY_HOSTNAME, RELAY_DOMAIN, RELAY_TRANSPORT=mtls as default, etc.), `--memory 256M`, `--network darkpipe`, `--read-only`, `--name darkpipe-relay`, detached; note: no `--cap-add`/`--cap-drop` (Apple Containers VM isolation replaces Linux capabilities); no `--device /dev/net/tun` (unavailable — mTLS transport default)
5. Implement readiness checks in `up`: poll caddy admin API (curl localhost:2019/config/) and relay SMTP (nc -z localhost 25) with timeout; implement `status` subcommand using `container list`; implement `down` using `container stop` + `container rm`; implement `logs` using `container logs`

## Must-Haves

- [ ] `--dry-run` mode prints all commands without executing (enables verification without macOS 26)
- [ ] `--help` prints complete usage information
- [ ] Script sources `cloud-relay/.env` when it exists for relay environment variables
- [ ] Caddy container run command includes all ports, volumes, env vars from compose
- [ ] Relay container run command includes port 25, env vars, memory limit; uses mTLS transport by default
- [ ] Network created before containers, destroyed on `down`
- [ ] Readiness polling with configurable timeout (default 30s) and descriptive failure messages
- [ ] Passes `shellcheck` and `bash -n` syntax validation

## Verification

- `shellcheck scripts/apple-containers-start.sh` exits 0
- `bash -n scripts/apple-containers-start.sh` exits 0
- `bash scripts/apple-containers-start.sh --help` prints usage with subcommands and flags
- `bash scripts/apple-containers-start.sh --dry-run up` prints `container network create`, `container build`, and two `container run` commands with correct flags
- `bash scripts/apple-containers-start.sh --dry-run down` prints `container stop`, `container rm`, and `container network rm` commands
- `bash scripts/apple-containers-start.sh --dry-run status` prints `container list` command

## Observability Impact

- Signals added/changed: script prints timestamped step-by-step progress (`[INFO]`, `[PASS]`, `[FAIL]` prefixes) matching check-runtime.sh output style
- How a future agent inspects this: `--dry-run` shows exact commands; `--verbose` shows full env var expansion; `status` subcommand checks running state
- Failure state exposed: each step (network create, build, container start, readiness check) fails independently with descriptive error and exit code

## Inputs

- `cloud-relay/docker-compose.yml` — source of truth for services, ports, volumes, env vars, resource limits
- `cloud-relay/Dockerfile` — built by `container build` command
- S03-RESEARCH.md — Apple Containers CLI flags, constraints, pitfalls

## Expected Output

- `scripts/apple-containers-start.sh` — complete orchestration script with up/down/status/logs subcommands and dry-run/verbose/help flags; passes shellcheck
