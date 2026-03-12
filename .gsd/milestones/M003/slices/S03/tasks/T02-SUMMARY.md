---
id: T02
parent: S03
milestone: M003
provides:
  - Apple Containers platform guide with prerequisites, quick start, key differences, limitations, and troubleshooting
  - Updated mac-silicon.md forward reference from "coming soon" to live link
key_files:
  - deploy/platform-guides/apple-containers.md
  - deploy/platform-guides/mac-silicon.md
key_decisions:
  - Platform guide follows podman.md structure (prerequisites table, quick start, key differences, troubleshooting) for consistency across platform guides
  - Quick start section uses orchestration script rather than raw container commands — simplifies user workflow
  - Limitations presented as a table for scannability; troubleshooting uses symptom/cause/fix pattern matching podman.md
patterns_established:
  - Apple Containers platform guide structure matching podman.md conventions
observability_surfaces:
  - none (documentation only)
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Write Apple Containers platform guide

**Created platform guide at `deploy/platform-guides/apple-containers.md` covering prerequisites, quick start, key differences from Docker, transport configuration, limitations, and troubleshooting; updated mac-silicon.md forward reference.**

## What Happened

Wrote the complete Apple Containers platform guide following the structure established by `podman.md`. The guide covers:

- **Prerequisites table**: macOS 26 (Tahoe), Apple Silicon, `container` CLI via Homebrew, 8GB+ RAM with 16GB recommended
- **Quick start**: 5-step workflow using `apple-containers-start.sh` (from T01), plus dry-run preview commands
- **Key differences from Docker**: VM-per-container isolation (no cap_add/cap_drop/security_opt), no compose tool, no restart policy, resource defaults (1GB/4CPU per VM), bind mounts only, no host-gateway, `container system start` requirement
- **Transport configuration**: mTLS as recommended default, WireGuard documented as unknown with test instructions
- **Limitations table**: 8 limitations with impact and workaround for each
- **Troubleshooting**: 7 scenarios — CLI not found, system not started, firewall prompt, macOS 15 networking, high memory, DNS issues, port 25 blocked

Updated `mac-silicon.md` to remove "(coming soon)" from the Apple Containers forward reference, making it a live link.

## Verification

- `test -f deploy/platform-guides/apple-containers.md` — PASS
- Guide contains all required sections (Prerequisites, Quick Start, Key Differences, Limitations, Troubleshooting) — PASS (5/5)
- `grep -q "coming soon" deploy/platform-guides/mac-silicon.md` returns false — PASS
- `grep -q "apple-containers.md" deploy/platform-guides/mac-silicon.md` — PASS
- All referenced scripts exist (`apple-containers-start.sh`, `check-runtime.sh`) — PASS
- All code blocks have language tags — PASS
- Prerequisites specify macOS 26, Apple Silicon, `brew install container` — PASS
- mTLS documented as default, WireGuard as unknown — PASS
- T01 carryover: shellcheck, --help, --dry-run up, bash -n all still pass — PASS

Slice verification status (8/9 passing — T03 not yet done):
- ✅ shellcheck apple-containers-start.sh
- ✅ --help prints usage
- ✅ --dry-run up outputs commands
- ✅ bash -n syntax check
- ⬜ shellcheck check-runtime.sh (T03)
- ⬜ grep apple-containers check-runtime.sh (T03)
- ✅ apple-containers.md exists
- ✅ Guide contains all required sections
- ✅ mac-silicon.md forward reference updated

## Diagnostics

Documentation-only task. Inspect by reading `deploy/platform-guides/apple-containers.md`. Verify mac-silicon.md link with `grep apple-containers deploy/platform-guides/mac-silicon.md`.

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `deploy/platform-guides/apple-containers.md` — New platform guide (complete)
- `deploy/platform-guides/mac-silicon.md` — Removed "(coming soon)" from Apple Containers forward reference
