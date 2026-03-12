---
estimated_steps: 4
estimated_files: 2
---

# T04: Document Podman deployment prerequisites and known differences

**Slice:** S01 — Podman Compose Compatibility
**Milestone:** M003

## Description

Create concise Podman deployment notes alongside each compose directory. These cover minimum versions, rootful vs rootless guidance, override file usage, and known differences. These are operational reference docs for users deploying on Podman today — the full platform guide (S02 scope) will expand on these.

## Steps

1. Create `cloud-relay/PODMAN.md` covering: Podman 5.3.0+ and podman-compose 1.x+ requirements, rootful requirement (port 25 binding + /dev/net/tun device access), override file usage (`podman-compose -f docker-compose.yml -f docker-compose.podman.yml up`), SELinux override for Fedora/RHEL, known difference re: `host.containers.internal` vs `host-gateway`
2. Create `home-device/PODMAN.md` covering: same version requirements, rootless option with `sysctl net.ipv4.ip_unprivileged_port_start=0`, override file usage, SELinux override, profile-specific startup commands, Podman pod mode warning (must NOT use `--in-pod` — breaks DNS)
3. Verify all referenced file names match actual override files from T02
4. Cross-reference version numbers and commands against S01-RESEARCH.md findings

## Must-Haves

- [ ] Both PODMAN.md files exist with accurate content
- [ ] Podman minimum version (5.3.0+) and podman-compose minimum version (1.x+) stated
- [ ] Rootful vs rootless guidance is correct per component
- [ ] Override file paths and usage commands are correct
- [ ] No incorrect or unsourced claims

## Verification

- Files exist at `cloud-relay/PODMAN.md` and `home-device/PODMAN.md`
- Referenced override files (`docker-compose.podman.yml`, `docker-compose.podman-selinux.yml`) exist
- Version requirements match research (Podman 5.3.0+, podman-compose 1.x+)
- sysctl command is correct: `sysctl net.ipv4.ip_unprivileged_port_start=0`

## Observability Impact

- Signals added/changed: None (static documentation)
- How a future agent inspects this: Read the PODMAN.md files
- Failure state exposed: None

## Inputs

- `.gsd/milestones/M003/slices/S01/S01-RESEARCH.md` — version requirements, constraints, pitfalls
- T02 output — override file names and paths
- `.gsd/DECISIONS.md` — existing decisions about compose structure

## Expected Output

- `cloud-relay/PODMAN.md` — cloud relay Podman deployment notes
- `home-device/PODMAN.md` — home device Podman deployment notes
