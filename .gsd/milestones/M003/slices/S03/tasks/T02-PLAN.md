---
estimated_steps: 4
estimated_files: 2
---

# T02: Write Apple Containers platform guide

**Slice:** S03 — Apple Containers Support
**Milestone:** M003

## Description

Create the platform guide at `deploy/platform-guides/apple-containers.md` following the established structure from `podman.md`. The guide covers prerequisites, quick start with the orchestration script, key differences from Docker, known limitations (WireGuard, no compose, no restart policy), and troubleshooting. Also updates the forward reference in `mac-silicon.md` from "coming soon" to an actual link.

## Steps

1. Write `deploy/platform-guides/apple-containers.md` with these sections:
   - **Header**: title, scope statement (cloud relay only, dev/testing), link to mac-silicon.md for general macOS guidance
   - **Prerequisites table**: macOS 26+ (Tahoe), Apple Silicon, `container` CLI (`brew install container`), 8GB+ RAM (16GB recommended — each container VM uses ~1GB default)
   - **Quick Start**: `container system start`, build + run via `scripts/apple-containers-start.sh up`, verify with `scripts/apple-containers-start.sh status`, teardown with `down`
   - **Key Differences from Docker**: VM-per-container isolation model, no compose tool, no `security_opt`/`cap_add`/`cap_drop` (stronger VM isolation instead), no `restart: unless-stopped`, no named volumes (bind mounts only), no `host-gateway`, `--memory` flag (default 1GB vs Docker's unlimited), `container system start` required
   - **Transport Configuration**: recommend mTLS as default transport; document WireGuard as unknown (Apple's custom kernel may omit module); show how to switch RELAY_TRANSPORT in .env
   - **Limitations**: no compose, manual orchestration, WireGuard uncertain, dev/testing only (same port 25/NAT limitations as Docker on Mac), no health check integration, API may change (first-release product)
   - **Troubleshooting**: forgot `container system start`, macOS firewall prompt, macOS 15 vs 26 networking, high memory usage (reduce with `--memory`), container-to-container DNS (use IPs if names don't resolve)
2. Update `deploy/platform-guides/mac-silicon.md` forward reference — change "coming soon" parenthetical to link to the completed guide
3. Review guide for consistency with podman.md structure and mac-silicon.md positioning (dev/testing only)
4. Verify all internal links resolve, code blocks have language tags, and prerequisites are accurate per research sources

## Must-Haves

- [ ] Guide follows podman.md structure (prerequisites table, quick start, key differences, troubleshooting)
- [ ] Prerequisites specify macOS 26, Apple Silicon, `container` CLI with install method
- [ ] Quick start uses orchestration script (not raw `container` commands) for main workflow
- [ ] WireGuard limitation clearly documented with mTLS as recommended fallback
- [ ] Dev/testing-only positioning matches mac-silicon.md
- [ ] mac-silicon.md forward reference updated from "coming soon"
- [ ] All constraints from S03-RESEARCH.md documented (no compose, VM model, resource defaults, no restart policy, etc.)

## Verification

- `test -f deploy/platform-guides/apple-containers.md` confirms file exists
- Guide contains sections: Prerequisites, Quick Start, Key Differences, Limitations, Troubleshooting (grep check)
- `grep -q "coming soon" deploy/platform-guides/mac-silicon.md` returns false (forward reference updated)
- `grep -q "apple-containers.md" deploy/platform-guides/mac-silicon.md` confirms link exists
- No broken relative links in guide (all referenced scripts and paths exist)

## Observability Impact

- Signals added/changed: None (documentation only)
- How a future agent inspects this: read the guide file; check mac-silicon.md link
- Failure state exposed: None

## Inputs

- `deploy/platform-guides/podman.md` — structural template to follow
- `deploy/platform-guides/mac-silicon.md` — forward reference to update; positioning context
- `scripts/apple-containers-start.sh` — from T01, referenced in quick start section
- S03-RESEARCH.md — all constraints, pitfalls, and sources

## Expected Output

- `deploy/platform-guides/apple-containers.md` — complete platform guide with all required sections
- `deploy/platform-guides/mac-silicon.md` — updated forward reference (one line change)
