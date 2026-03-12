---
estimated_steps: 4
estimated_files: 2
---

# T03: Extend check-runtime.sh with Apple Containers detection and validate all artifacts

**Slice:** S03 — Apple Containers Support
**Milestone:** M003

## Description

Extend `scripts/check-runtime.sh` to detect Apple Containers (`container` CLI) as a recognized runtime alongside Docker and Podman. Then run the full slice verification suite to confirm all artifacts are complete and correct. This is the final task — it closes the loop on the slice by wiring Apple Containers into the existing tooling and validating everything together.

## Steps

1. Read the existing `check-runtime.sh` detection flow to understand the pattern (Docker first, then Podman fallback). Add Apple Containers as a third detection path:
   - Check for `container` CLI availability (`command -v container`)
   - Gate on macOS (`uname -s` == Darwin) — Apple Containers only runs on macOS
   - Parse version from `container --version` output
   - Set `DETECTED_RUNTIME="apple-containers"` with version
   - Skip compose tool check (no compose equivalent — print SKIP with explanation)
   - Skip SELinux check (not applicable on macOS — print SKIP)
   - Port 25 check still applies
2. Add Apple Containers to the environment summary output section
3. Test that existing Docker and Podman detection is unaffected — the Apple Containers check should only trigger when neither Docker nor Podman is found AND `container` CLI exists on macOS
4. Run full slice verification:
   - `shellcheck scripts/apple-containers-start.sh` and `scripts/check-runtime.sh`
   - `bash scripts/apple-containers-start.sh --dry-run up` outputs correct commands
   - `deploy/platform-guides/apple-containers.md` exists with required sections
   - `mac-silicon.md` forward reference is updated
   - `grep` confirms Apple Containers detection in check-runtime.sh

## Must-Haves

- [ ] Apple Containers detected when `container` CLI exists on macOS and neither Docker nor Podman is found
- [ ] Version parsed from `container --version` output
- [ ] Compose check outputs SKIP with "Apple Containers has no compose equivalent" message
- [ ] SELinux check outputs SKIP on macOS
- [ ] Existing Docker and Podman detection unchanged (Apple Containers is lowest priority fallback)
- [ ] `shellcheck` passes on modified check-runtime.sh
- [ ] All slice verification checks pass

## Verification

- `shellcheck scripts/check-runtime.sh` exits 0
- `grep -q 'apple-containers\|Apple Containers' scripts/check-runtime.sh` confirms detection logic
- `grep -q 'container --version' scripts/check-runtime.sh` confirms version check
- Existing PASS/FAIL/SKIP output pattern preserved
- All T01 and T02 verification checks still pass (regression check)

## Observability Impact

- Signals added/changed: check-runtime.sh now reports `Runtime: apple-containers` and version when detected; SKIP messages explain why compose and SELinux checks are not applicable
- How a future agent inspects this: run `bash scripts/check-runtime.sh` on macOS with `container` CLI installed; output shows detected runtime and version
- Failure state exposed: FAIL if `container` CLI exists but version cannot be parsed; SKIP with reason for inapplicable checks

## Inputs

- `scripts/check-runtime.sh` — existing runtime detection script (357 lines, Docker + Podman detection)
- `scripts/apple-containers-start.sh` — from T01 (verified in final validation)
- `deploy/platform-guides/apple-containers.md` — from T02 (verified in final validation)

## Expected Output

- `scripts/check-runtime.sh` — extended with Apple Containers detection (macOS-gated, version parsing, SKIP for compose/SELinux)
- All slice artifacts validated together; slice verification suite passes
