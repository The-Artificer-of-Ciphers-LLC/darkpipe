---
id: T03
parent: S03
milestone: M003
provides:
  - Apple Containers detection in check-runtime.sh (macOS-gated, version parsing, SKIP for compose/SELinux)
  - Full slice verification suite passing for all S03 artifacts
key_files:
  - scripts/check-runtime.sh
key_decisions:
  - Apple Containers detection is lowest-priority fallback (Docker → Podman → Apple Containers) to avoid interfering with existing runtimes
  - No minimum version enforced for Apple Containers (versioning tracks macOS releases, no established baseline yet)
  - SELinux check skips on all macOS systems (not just Apple Containers) since SELinux is never applicable on Darwin
patterns_established:
  - macOS-gated runtime detection pattern (uname -s == Darwin guard before CLI check)
  - SKIP-with-explanation pattern for inapplicable checks (compose, SELinux)
observability_surfaces:
  - "check-runtime.sh reports Runtime: apple-containers with version when detected on macOS"
  - "SKIP messages explain why compose and SELinux checks are not applicable"
  - "FAIL if container CLI exists but version cannot be parsed"
duration: 15m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T03: Extended check-runtime.sh with Apple Containers detection and validated all S03 artifacts

**Added Apple Containers as a third runtime detection path in check-runtime.sh with macOS gate, version parsing, and SKIP for inapplicable checks; all slice verification checks pass.**

## What Happened

Extended `scripts/check-runtime.sh` to detect Apple Containers (`container` CLI) as a recognized runtime alongside Docker and Podman. The detection is:

1. **Lowest priority** — only triggers when neither Docker nor Podman is found
2. **macOS-gated** — checks `uname -s == Darwin` before probing for `container` CLI
3. **Version-aware** — parses version from `container --version` output using the existing `parse_version()` helper
4. **SKIP-aware** — compose check outputs SKIP with explanation (no compose equivalent); SELinux check outputs SKIP on macOS (not applicable)

Also improved the SELinux check to skip early on all macOS systems (not just when `getenforce` is missing), since SELinux is never present on Darwin.

Updated script header comments and usage text to reflect the three supported runtimes.

Ran the full slice verification suite confirming all T01, T02, and T03 artifacts are correct.

## Verification

All slice-level verification checks pass:

- `shellcheck scripts/apple-containers-start.sh` — exit 0 ✅
- `bash scripts/apple-containers-start.sh --help` — prints usage ✅
- `bash scripts/apple-containers-start.sh --dry-run up` — prints correct container commands ✅
- `bash -n scripts/apple-containers-start.sh` — syntax check passes ✅
- `shellcheck scripts/check-runtime.sh` — exit 0 ✅
- `grep -q "apple-containers" scripts/check-runtime.sh` — confirms detection logic ✅
- `grep -q "container --version" scripts/check-runtime.sh` — confirms version check ✅
- `test -f deploy/platform-guides/apple-containers.md` — guide exists ✅
- Platform guide contains: Prerequisites, Quick Start, Key Differences, Limitations, Troubleshooting ✅
- `mac-silicon.md` forward reference updated (no "coming soon") ✅

## Diagnostics

- Run `bash scripts/check-runtime.sh` on macOS with `container` CLI installed — output shows `Runtime: apple-containers` and version
- On systems without `container` CLI or not on macOS, Apple Containers detection is silently skipped (Docker/Podman paths unchanged)
- FAIL output if `container` CLI exists but version cannot be parsed

## Deviations

- SELinux skip was broadened to all macOS systems (not just when runtime is apple-containers) since SELinux is never applicable on Darwin regardless of runtime. This is a minor improvement, not a plan violation.
- Added `DETECTED_OS` state variable to cache `uname -s` result (used by both runtime detection and SELinux check).

## Known Issues

None.

## Files Created/Modified

- `scripts/check-runtime.sh` — Extended with Apple Containers detection (macOS-gated, version parsing, SKIP for compose/SELinux), updated comments and usage text
