# M003: Container Runtime Compatibility — Context

**Gathered:** 2026-03-12
**Status:** Ready for planning

## Project Description

DarkPipe currently assumes Docker as the only container runtime. All documentation, compose files, Dockerfiles, verification scripts, platform guides, and CI workflows reference Docker exclusively. This milestone adds first-class support for Podman (the most mature Docker alternative) and Apple Containers (Apple's new native macOS container runtime shipping with macOS 26), and ensures the project's language, tooling, and testing account for non-Docker runtimes.

## Why This Milestone

Three forces drive this now:

1. **Podman adoption is accelerating.** Red Hat, Fedora, and RHEL default to Podman. Security-conscious users (DarkPipe's target audience) prefer its daemonless, rootless-by-default architecture. The FAQ already acknowledges Podman interest but says "not tested."

2. **Apple Containers launched with macOS 26.** Apple's native container runtime (WWDC 2025) runs each container in its own lightweight VM on Apple Silicon. DarkPipe already has a Mac Silicon platform guide — users on macOS 26 will expect native container support without Docker Desktop.

3. **DarkPipe's privacy-first users care about alternatives.** Docker Desktop's telemetry, licensing changes, and daemon-based architecture conflict with the sovereignty values DarkPipe promotes. Offering tested alternatives aligns with the project's philosophy.

## User-Visible Outcome

### When this milestone is complete, the user can:

- Deploy DarkPipe on Podman using `podman-compose` with documented steps and tested compose files
- Deploy DarkPipe on macOS 26 using Apple Containers with a dedicated platform guide
- Reference runtime-agnostic documentation that doesn't assume Docker as the only option
- Run a compatibility check script that validates their runtime meets DarkPipe requirements
- Build DarkPipe images with Podman's `podman build` (OCI-compatible Dockerfiles)

### Entry point / environment

- Entry point: `podman-compose up` / `container run` / `darkpipe-setup`
- Environment: Linux with Podman, macOS 26 with Apple Containers, existing Docker environments
- Live dependencies involved: Podman, podman-compose, Apple Containerization framework, WireGuard, mail servers

## Completion Class

- Contract complete means: compose files validated with `podman-compose config`, Containerfiles build with `podman build`, compatibility check script passes on Podman and reports correct status for Apple Containers
- Integration complete means: full mail flow (send/receive) tested on Podman, images build and start on Apple Containers
- Operational complete means: services survive restart cycles, health checks work across runtimes, security directives (cap_drop, read_only, no-new-privileges) behave correctly on Podman

## Final Integrated Acceptance

To call this milestone complete, we must prove:

- `podman-compose --profile stalwart --profile snappymail up -d` starts all services with health checks passing
- Cloud relay and home device communicate over WireGuard/mTLS on a Podman deployment
- Apple Containers platform guide enables a user to run at minimum the cloud relay components
- All existing Docker deployments continue to work identically (no regression)
- Documentation uses runtime-agnostic language with runtime-specific callouts where behavior differs

## Risks and Unknowns

- **Podman networking differences** — Podman uses slirp4netns or pasta for rootless networking; WireGuard tunnel may need different configuration for rootless Podman vs Docker bridge networking
- **Podman compose maturity** — podman-compose supports compose v3.x but some advanced features (build contexts, depends_on conditions, healthcheck integration) may behave differently
- **Apple Containers multi-container orchestration** — Apple Containers has no compose equivalent yet; each container runs in its own VM with its own IP. Multi-service orchestration may require custom scripting
- **Apple Containers WireGuard support** — Each Apple Container VM has its own network stack; WireGuard kernel module availability in Apple's custom Linux kernel is unknown
- **cap_drop / security_opt on Podman** — Podman's rootless mode already drops capabilities by default; our explicit cap_drop/security_opt directives may conflict or be redundant
- **SELinux volume labels** — Podman on Fedora/RHEL requires `:Z` volume label suffix for bind mounts; current compose files don't include this

## Existing Codebase / Prior Art

- `cloud-relay/docker-compose.yml` — Primary compose file, uses Docker-specific features (security_opt, cap_drop, tmpfs, healthcheck)
- `home-device/docker-compose.yml` — 9-service compose with profiles, security hardening from M002
- `cloud-relay/certbot/docker-compose.certbot.yml` — Certbot compose file
- `cloud-relay/Dockerfile`, `home-device/*/Dockerfile` — 5 custom Dockerfiles (OCI-compatible, should work with podman build)
- `scripts/verify-container-security.sh` — Hardcoded to check docker-compose YAML; needs runtime-awareness
- `deploy/platform-guides/mac-silicon.md` — Existing Mac guide assumes Docker Desktop
- `docs/faq.md` — "Can I use Podman instead of Docker?" section says "Probably, but not officially supported"
- `.github/workflows/` — CI uses `docker/build-push-action`; Podman build should be tested

> See `.gsd/DECISIONS.md` for all architectural and pattern decisions.

## Relevant Requirements

- No new formal requirements — this milestone extends platform compatibility for existing validated requirements

## Scope

### In Scope

- Podman compatibility testing and fixes for all compose files
- Podman-specific documentation (rootless setup, SELinux volume labels, networking)
- Apple Containers platform guide (macOS 26)
- Runtime-agnostic language in core documentation
- Runtime compatibility check script
- Podman CI testing (GitHub Actions with Podman)
- Addressing known Podman incompatibilities in compose files
- WireGuard/mTLS transport testing on Podman

### Out of Scope / Non-Goals

- Kubernetes / K8s deployment (separate milestone)
- LXC/LXD system container support (niche use case)
- Rancher Desktop or Lima support (they use Docker-compatible CLIs, should work if Docker works)
- Cloud container services (ECS, Cloud Run, etc.)
- Rewriting Dockerfiles as Containerfiles (they're OCI-compatible, renaming is cosmetic)
- Full Apple Containers compose-equivalent orchestration tool
- Podman pod-native architecture (keep compose-based for Docker compatibility)

## Technical Constraints

- Must not break existing Docker deployments
- Compose files must remain valid for both `docker compose` and `podman-compose`
- Dockerfiles must remain OCI-compatible (buildable by Docker, Podman, and Buildah)
- Security hardening directives from M002 must work or have documented alternatives on Podman
- Apple Containers support is best-effort (runtime is new, API may change)

## Integration Points

- **Podman** — compose file compatibility, volume mounts, networking, capability handling
- **Apple Containerization framework** — image pulling, container lifecycle, networking model
- **WireGuard** — kernel module availability and network namespace behavior differ across runtimes
- **GitHub Actions** — Podman is available on ubuntu-latest runners; can test builds
- **Platform guides** — Mac Silicon guide needs Apple Containers path; new Podman guide for Fedora/RHEL

## Open Questions

- Can Podman rootless mode bind to port 25? Privileged ports require `sysctl net.ipv4.ip_unprivileged_port_start=0` or running as root. DarkPipe's cloud relay needs port 25. — Current thinking: Document both rootful (for cloud relay) and rootless (for home device) Podman configurations.
- Does Apple Containers support Docker/OCI image format directly? — Research suggests yes (pulls from standard registries), but multi-service orchestration is manual.
- Should we rename Dockerfiles to Containerfiles? — Current thinking: No, keep as Dockerfile for maximum compatibility. Podman reads Dockerfiles natively.
