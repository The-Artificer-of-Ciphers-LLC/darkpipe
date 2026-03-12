---
id: T04
parent: S01
milestone: M003
provides:
  - Podman deployment documentation for cloud-relay and home-device
key_files:
  - cloud-relay/PODMAN.md
  - home-device/PODMAN.md
key_decisions:
  - Cloud relay documented as rootful-only (port 25 + /dev/net/tun); home device documented with rootless option via sysctl
patterns_established:
  - Per-component PODMAN.md files alongside compose directories for operational reference
observability_surfaces:
  - none (static documentation)
duration: 1 step
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T04: Document Podman deployment prerequisites and known differences

**Created `cloud-relay/PODMAN.md` and `home-device/PODMAN.md` with version requirements, rootful/rootless guidance, override file usage, profile-specific commands, and known Podman differences.**

## What Happened

Created two operational reference docs for Podman deployment:

- **`cloud-relay/PODMAN.md`** — Documents rootful Podman as required (port 25 + /dev/net/tun), override file layering commands, SELinux override usage, known differences (host-gateway, networking, security defaults, memory limits), and firewalld configuration for Fedora/RHEL.

- **`home-device/PODMAN.md`** — Documents rootless option with `sysctl net.ipv4.ip_unprivileged_port_start=0`, all three profile startup commands with Podman override files, SELinux override layering, pod mode warning (must NOT use `--in-pod`), and known differences.

All content cross-referenced against S01-RESEARCH.md findings and T02 override file names.

## Verification

- Both files exist: `cloud-relay/PODMAN.md`, `home-device/PODMAN.md` ✅
- Podman 5.3.0+ and podman-compose 1.x+ stated in both files ✅
- Rootful required for cloud-relay, rootless option for home-device ✅
- Override file names match actual files from T02 (`docker-compose.podman.yml`, `docker-compose.podman-selinux.yml`) ✅
- sysctl command correct: `sysctl net.ipv4.ip_unprivileged_port_start=0` ✅
- `host-gateway` vs `host.containers.internal` difference documented ✅
- Pod mode warning present in home-device doc ✅
- `scripts/verify-podman-compat.sh` passes: 17/17 pass, 0 fail, 4 skip ✅

### Slice-level verification status

| Check | Status |
|-------|--------|
| `scripts/verify-podman-compat.sh` exits 0 | ✅ pass (static checks; runtime checks skipped — docker not installed) |
| `docker compose -f cloud-relay/docker-compose.yml config --quiet` | ⏭️ skip (docker not installed) |
| `docker compose -f home-device/docker-compose.yml config --quiet` | ⏭️ skip (docker not installed) |
| `docker compose ... -f docker-compose.podman.yml config --quiet` | ⏭️ skip (docker not installed) |
| `scripts/verify-container-security.sh` still passes | ✅ pass (verified by T03) |
| Health checks use CMD-SHELL form | ✅ pass (verified by script) |

## Diagnostics

Read the PODMAN.md files directly. No runtime diagnostics — these are static documentation.

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `cloud-relay/PODMAN.md` — Podman deployment guide for cloud relay (rootful, override files, known differences)
- `home-device/PODMAN.md` — Podman deployment guide for home device (rootless option, profiles, pod mode warning)
