# S01: Podman Compose Compatibility

**Goal:** All DarkPipe compose files work with both `docker compose` and `podman-compose`, with health checks passing and a verified mail flow path on Podman.
**Demo:** `podman-compose up` starts cloud-relay and home-device services, health checks pass, and compose validation succeeds for both runtimes. Compose files remain a single set — no forks.

## Must-Haves

- Health check `CMD` form bugs fixed to `CMD-SHELL` across all compose files
- `extra_hosts: host-gateway` compatibility ensured (Podman 5.3.0+ requirement documented)
- `version: '3.8'` handling addressed (suppress podman-compose deprecation warnings)
- SELinux `:Z` volume label strategy via override file (not in main compose files)
- Podman-compose override file for SELinux environments
- Cloud relay rootful Podman documented (port 25 + /dev/net/tun)
- `podman-compose config` validation passes for all compose files
- `docker compose config` validation still passes (zero regression)
- Verification script that validates compose files against both runtimes

## Proof Level

- This slice proves: contract + partial integration (compose parse/validate, health check syntax, override file layering)
- Real runtime required: no — compose config validation proves parse compatibility; actual Podman runtime testing deferred to CI (S04) and manual verification
- Human/UAT required: no

## Verification

- `scripts/verify-podman-compat.sh` — validates all compose files with both `docker compose config` and `podman-compose config` (when available), checks health check syntax, validates override file layering
- `docker compose -f cloud-relay/docker-compose.yml config --quiet` exits 0
- `docker compose -f home-device/docker-compose.yml config --quiet` exits 0
- `docker compose -f cloud-relay/docker-compose.yml -f cloud-relay/docker-compose.podman.yml config --quiet` exits 0 (override layering)
- `scripts/verify-container-security.sh` still passes (zero regression)
- Health checks use `CMD-SHELL` form with proper shell `||` interpretation

## Observability / Diagnostics

- Runtime signals: verification script outputs structured PASS/FAIL per check (same pattern as existing `verify-container-security.sh`)
- Inspection surfaces: `scripts/verify-podman-compat.sh` runnable by any agent or CI
- Failure visibility: script prints specific failing compose file, service name, and field on failure
- Redaction constraints: none (no secrets in compose files)

## Integration Closure

- Upstream surfaces consumed: existing `cloud-relay/docker-compose.yml`, `home-device/docker-compose.yml`, `cloud-relay/certbot/docker-compose.certbot.yml`
- New wiring introduced in this slice: Podman override files (`docker-compose.podman.yml`), compatibility verification script
- What remains before the milestone is truly usable end-to-end: S02 (documentation + runtime check script), S03 (Apple Containers), S04 (CI validation with actual Podman runtime)

## Tasks

- [x] **T01: Fix health check CMD form bugs and remove version field** `est:30m`
  - Why: Health checks using `["CMD", "wget", ..., "||", "exit", "1"]` pass `||` as a literal string argument — broken on both Docker and Podman. The `version: '3.8'` field triggers deprecation warnings on podman-compose. These are blocking compatibility issues.
  - Files: `cloud-relay/docker-compose.yml`, `home-device/docker-compose.yml`
  - Do: Convert all health checks from CMD form with `||` to CMD-SHELL form (`test: ["CMD-SHELL", "wget ... || exit 1"]`). Remove or comment the `version: '3.8'` field from both main compose files and the certbot override. Keep all other health check parameters unchanged.
  - Verify: `docker compose -f cloud-relay/docker-compose.yml config --quiet` and `docker compose -f home-device/docker-compose.yml config --quiet` exit 0. `grep -r '"||"' cloud-relay/docker-compose.yml home-device/docker-compose.yml` returns no matches.
  - Done when: All health checks use CMD-SHELL form, no compose file has `version:` field, and `docker compose config` passes for all files.

- [x] **T02: Create Podman override files for SELinux and rootful configuration** `est:45m`
  - Why: SELinux `:Z` volume labels are required on Fedora/RHEL but destructive if added unconditionally. Cloud relay needs rootful Podman documentation. Override files let Podman users layer these on without changing the base compose files.
  - Files: `cloud-relay/docker-compose.podman.yml`, `home-device/docker-compose.podman.yml`, `cloud-relay/docker-compose.podman-selinux.yml`, `home-device/docker-compose.podman-selinux.yml`
  - Do: Create cloud-relay override with `x-podman` extensions for Docker compat flags. Create SELinux override files that add `:Z` suffix to bind-mount volumes (not named volumes). Document the layering: `podman-compose -f docker-compose.yml -f docker-compose.podman.yml up`. Keep overrides minimal — only what differs from base.
  - Verify: `docker compose -f cloud-relay/docker-compose.yml -f cloud-relay/docker-compose.podman.yml config --quiet` exits 0. Override files parse cleanly.
  - Done when: Override files exist, layer correctly on base compose files, and `docker compose config` accepts the layered result.

- [x] **T03: Build Podman compatibility verification script** `est:45m`
  - Why: The slice needs a repeatable, agent-runnable proof that compose files are dual-compatible. This script is the primary verification artifact for S01 and will be consumed by S04 for CI.
  - Files: `scripts/verify-podman-compat.sh`
  - Do: Create bash script following the pattern of `verify-container-security.sh` (structured PASS/FAIL output). Checks: (1) all compose files pass `docker compose config`, (2) health checks use CMD-SHELL form, (3) no `version:` field present, (4) override files layer cleanly, (5) no Docker Swarm-only directives, (6) `extra_hosts: host-gateway` usage documented with Podman 5.3.0+ note, (7) `podman-compose config` if podman-compose is available (graceful skip if not installed). Exit non-zero on any failure.
  - Verify: `bash scripts/verify-podman-compat.sh` exits 0. Intentionally break a compose file and confirm the script catches it.
  - Done when: Script passes on current compose files, catches intentional breakage, and follows the existing verification script pattern.

- [x] **T04: Document Podman deployment prerequisites and known differences** `est:30m`
  - Why: Users need to know Podman minimum version (5.3.0+), rootful vs rootless guidance, sysctl for privileged ports, and how to use override files. This documentation is the "deployment guide" part of the slice and feeds into S02's full platform guide.
  - Files: `cloud-relay/PODMAN.md`, `home-device/PODMAN.md`
  - Do: Create concise Podman notes files alongside each compose directory. Cover: minimum Podman/podman-compose versions, rootful requirement for cloud relay (port 25 + /dev/net/tun), rootless option for home device with sysctl, override file usage, SELinux override for Fedora/RHEL, `host-gateway` note (Podman 5.3.0+ required), known differences from Docker. Keep focused — full platform guide is S02 scope.
  - Verify: Files exist, reference correct override file names, and version requirements match research findings.
  - Done when: Both PODMAN.md files exist with accurate, actionable content covering all prerequisites from research.

- [x] **T05: Run full verification suite and fix any regressions** `est:20m`
  - Why: Final integration check — both the new Podman compatibility script and the existing container security script must pass together, proving zero regression.
  - Files: `scripts/verify-podman-compat.sh`, `scripts/verify-container-security.sh`
  - Do: Run both verification scripts. Fix any issues found. Verify override file layering doesn't break security checks. Run `docker compose config` on all compose file combinations (base, base+podman, base+podman+selinux).
  - Verify: `bash scripts/verify-podman-compat.sh && bash scripts/verify-container-security.sh` exits 0.
  - Done when: Both scripts pass, all compose file combinations validate, and no existing behavior has regressed.

## Files Likely Touched

- `cloud-relay/docker-compose.yml`
- `home-device/docker-compose.yml`
- `cloud-relay/certbot/docker-compose.certbot.yml`
- `cloud-relay/docker-compose.podman.yml`
- `home-device/docker-compose.podman.yml`
- `cloud-relay/docker-compose.podman-selinux.yml`
- `home-device/docker-compose.podman-selinux.yml`
- `scripts/verify-podman-compat.sh`
- `cloud-relay/PODMAN.md`
- `home-device/PODMAN.md`
