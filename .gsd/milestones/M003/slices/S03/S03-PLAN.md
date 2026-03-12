# S03: Apple Containers Support

**Goal:** macOS 26 platform guide enables running DarkPipe cloud relay on Apple Containers, with a shell orchestration script that starts Caddy + relay on a shared network, and SMTP connectivity verified.
**Demo:** `bash scripts/apple-containers-start.sh up` starts both cloud relay containers on a shared Apple Containers network; `container exec` verifies SMTP port 25 is listening; the platform guide documents all steps, limitations, and WireGuard/mTLS transport considerations.

## Must-Haves

- Platform guide at `deploy/platform-guides/apple-containers.md` following Podman guide structure
- Shell orchestration script `scripts/apple-containers-start.sh` that creates network, builds images, starts Caddy + relay with correct flags
- Orchestration script handles startup ordering and basic readiness checks
- `scripts/check-runtime.sh` detects Apple Containers as a recognized runtime
- WireGuard limitation documented with mTLS as recommended transport
- All constraints from research documented: no compose, no restart policy, no cap_add/cap_drop, VM isolation model, resource defaults
- mac-silicon.md forward reference updated to link to completed guide

## Proof Level

- This slice proves: contract (script syntax, flag translation, documentation completeness) + partial integration (check-runtime.sh detection logic tested)
- Real runtime required: no — Apple Containers requires macOS 26 on Apple Silicon; verification is structural (shellcheck, flag correctness, documentation review) with UAT deferred
- Human/UAT required: yes — final walkthrough on macOS 26 hardware needed (tracked in S03-UAT.md)

## Verification

- `shellcheck scripts/apple-containers-start.sh` passes with no errors
- `bash scripts/apple-containers-start.sh --help` prints usage without errors
- `bash scripts/apple-containers-start.sh --dry-run up` prints all `container` commands it would execute without running them
- `bash -n scripts/apple-containers-start.sh` (syntax check passes)
- `shellcheck scripts/check-runtime.sh` still passes after Apple Containers additions
- `grep -q "apple-containers" scripts/check-runtime.sh` confirms detection logic exists
- `test -f deploy/platform-guides/apple-containers.md` confirms guide exists
- Platform guide contains required sections: Prerequisites, Quick Start, Key Differences, Limitations, Troubleshooting
- mac-silicon.md forward reference updated (no longer says "coming soon")

## Observability / Diagnostics

- Runtime signals: orchestration script logs each step with timestamps and PASS/FAIL indicators; `--verbose` flag shows full `container` commands
- Inspection surfaces: `--dry-run` mode for verifying command translation without runtime; `--status` flag checks if containers are running
- Failure visibility: script exits with descriptive error on each failure point (network creation, build, container start, readiness check); exit codes distinguish failure phases
- Redaction constraints: none (no secrets in Apple Containers orchestration — env vars are passed by reference from .env file, not logged)

## Integration Closure

- Upstream surfaces consumed: `cloud-relay/docker-compose.yml` (source of truth for services, ports, volumes, env vars), `cloud-relay/Dockerfile` (built by orchestration script), `scripts/check-runtime.sh` (extended with Apple Containers detection), `deploy/platform-guides/podman.md` (structural template)
- New wiring introduced in this slice: `scripts/apple-containers-start.sh` orchestrates Apple Containers startup; `check-runtime.sh` gains Apple Containers detection; `apple-containers.md` guide links from `mac-silicon.md`
- What remains before the milestone is truly usable end-to-end: S04 (CI validation) must complete; UAT walkthrough on macOS 26 hardware must validate the guide and script work end-to-end

## Tasks

- [x] **T01: Create Apple Containers orchestration script with dry-run mode** `est:1h`
  - Why: Apple Containers has no compose equivalent — a shell script must translate docker-compose.yml services into individual `container` CLI commands with proper networking, volumes, env vars, and startup ordering
  - Files: `scripts/apple-containers-start.sh`
  - Do: Translate cloud-relay compose services (caddy + relay) into `container run` commands; implement `up`, `down`, `status`, `logs` subcommands; add `--dry-run` that prints commands without executing; add `--verbose` for debug output; create shared network via `container network create`; handle readiness checks (curl caddy admin, nc relay port 25); source env vars from `cloud-relay/.env`; use mTLS as default transport (WireGuard unknown on Apple's kernel)
  - Verify: `shellcheck scripts/apple-containers-start.sh` passes; `bash -n scripts/apple-containers-start.sh` passes; `bash scripts/apple-containers-start.sh --help` prints usage; `bash scripts/apple-containers-start.sh --dry-run up` outputs expected `container` CLI commands
  - Done when: script passes shellcheck, dry-run outputs correct `container network create`, `container build`, and `container run` commands for both services with all required flags

- [x] **T02: Write Apple Containers platform guide** `est:45m`
  - Why: Users need a complete guide to run DarkPipe cloud relay on Apple Containers, following the established platform guide structure from podman.md
  - Files: `deploy/platform-guides/apple-containers.md`, `deploy/platform-guides/mac-silicon.md`
  - Do: Write guide with Prerequisites table (macOS 26, Apple Silicon, `container` CLI, Homebrew install), Quick Start (system start, build, run via script), Key Differences from Docker (VM isolation, no compose, no restart policy, no cap_add/cap_drop, resource defaults), Limitations section (WireGuard unknown — recommend mTLS, no health check orchestration, manual restart, dev/testing only), Troubleshooting section (system start forgotten, firewall prompt, networking on macOS 15 vs 26, resource usage); update mac-silicon.md forward reference from "coming soon" to actual link
  - Verify: guide exists at correct path; contains all required sections; mac-silicon.md link is updated; no broken relative links
  - Done when: `deploy/platform-guides/apple-containers.md` has Prerequisites, Quick Start, Key Differences, Limitations, and Troubleshooting sections; `mac-silicon.md` references the completed guide

- [x] **T03: Extend check-runtime.sh with Apple Containers detection and validate all artifacts** `est:30m`
  - Why: The runtime compatibility check script must recognize Apple Containers as a supported runtime so users get correct validation output; final task validates all slice artifacts together
  - Files: `scripts/check-runtime.sh`, `scripts/apple-containers-start.sh`
  - Do: Add Apple Containers detection to check-runtime.sh (`container --version` check, version parsing, macOS-only gate); add to environment summary output; ensure existing Docker/Podman detection is unaffected; run full slice verification suite
  - Verify: `shellcheck scripts/check-runtime.sh` passes; `grep -q "apple-containers\|Apple Containers\|container --version" scripts/check-runtime.sh` confirms detection; dry-run of orchestration script still works; platform guide is complete
  - Done when: check-runtime.sh detects Apple Containers when `container` CLI is available, reports version, and all slice verification checks pass

## Files Likely Touched

- `scripts/apple-containers-start.sh` (new)
- `deploy/platform-guides/apple-containers.md` (new)
- `deploy/platform-guides/mac-silicon.md` (update forward reference)
- `scripts/check-runtime.sh` (extend with Apple Containers detection)
