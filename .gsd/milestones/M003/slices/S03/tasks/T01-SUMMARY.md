---
id: T01
parent: S03
milestone: M003
provides:
  - Apple Containers orchestration script with up/down/status/logs and dry-run mode
key_files:
  - scripts/apple-containers-start.sh
key_decisions:
  - mTLS is the default RELAY_TRANSPORT for Apple Containers (WireGuard kernel module unconfirmed)
  - Bind mounts to host data/ directories instead of named volumes (Apple Containers has no volume driver)
  - No --cap-add/--cap-drop flags (Apple Containers VM isolation replaces Linux capabilities)
  - No --device /dev/net/tun (unavailable in Apple's VM kernel — mTLS transport default)
patterns_established:
  - Shell orchestration pattern for Apple Containers multi-service startup (network → pull/build → run → readiness)
  - Timestamped INFO/PASS/FAIL/CMD log prefixes matching check-runtime.sh style
  - Dry-run mode that prints exact commands for verification without target hardware
observability_surfaces:
  - "--dry-run mode prints all container CLI commands without executing"
  - "--verbose mode shows env var expansion and exec details"
  - "Each step logs with [INFO]/[PASS]/[FAIL] prefixes and timestamps"
  - "Distinct exit codes per failure phase: 3=network, 4=build, 5=start, 6=readiness"
  - "status subcommand runs container list"
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Create Apple Containers orchestration script with dry-run mode

**Built shell orchestration script that translates docker-compose.yml into Apple Containers CLI commands with up/down/status/logs subcommands and dry-run mode.**

## What Happened

Created `scripts/apple-containers-start.sh` that translates the 2-service cloud-relay docker-compose.yml (caddy + relay) into individual `container` CLI commands. The script implements:

- **Subcommands:** `up` (create network, pull/build images, start containers, readiness checks), `down` (stop, remove containers and network), `status` (list containers), `logs` (show container logs)
- **Flags:** `--dry-run` prints all commands without executing, `--verbose` shows debug details, `--help` prints full usage
- **Network management:** Creates `darkpipe` network before containers, removes on `down`, skips if already exists
- **Caddy container:** ports 80/443/443udp, bind mounts for Caddyfile/data/config/logs, env vars for domains, 128M memory, read-only, on darkpipe network
- **Relay container:** port 25, env vars sourced from cloud-relay/.env with mTLS as default transport, 256M memory, read-only, bind mounts for postfix queue/certbot/queue data, optional env vars passed through only when set
- **Readiness checks:** Polls caddy admin API (curl localhost:2019/config/) and relay SMTP (nc -z localhost 25) with configurable READINESS_TIMEOUT (default 30s)

Key translation decisions from docker-compose.yml: no `--cap-add`/`--cap-drop` (VM isolation), no `--device /dev/net/tun` (mTLS default), bind mounts instead of named volumes, no restart policy (unavailable in Apple Containers).

## Verification

- `shellcheck scripts/apple-containers-start.sh` — exits 0, no warnings
- `bash -n scripts/apple-containers-start.sh` — exits 0, syntax valid
- `bash scripts/apple-containers-start.sh --help` — prints complete usage with subcommands, flags, examples, requirements, exit codes
- `bash scripts/apple-containers-start.sh --dry-run up` — prints 7 CMD lines: network create, pull, build, 2x container run (with correct ports/volumes/env/memory), 2x readiness checks
- `bash scripts/apple-containers-start.sh --dry-run down` — prints stop/rm for both containers + network rm
- `bash scripts/apple-containers-start.sh --dry-run status` — prints container list command

## Diagnostics

- Run `--dry-run up` to see exact commands that would execute
- Run `--dry-run --verbose up` to see env var expansion
- Exit codes distinguish failure phase: 2=usage, 3=network, 4=build, 5=container start, 6=readiness
- Each step logs [INFO]/[PASS]/[FAIL] with timestamps for post-mortem analysis

## Deviations

None.

## Known Issues

- DNS resolution between containers by name on Apple Containers network is unconfirmed — may need IP-based addressing (documented in research, not blocking for script)
- Script cannot be tested end-to-end without macOS 26 hardware (dry-run validates command correctness)

## Files Created/Modified

- `scripts/apple-containers-start.sh` — New orchestration script (executable, 330 lines)
