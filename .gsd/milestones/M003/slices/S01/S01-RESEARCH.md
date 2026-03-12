# S01: Podman Compose Compatibility — Research

**Date:** 2026-03-11

## Summary

DarkPipe's compose files use compose v3.8 features that are largely compatible with podman-compose, but several specific features require attention: `extra_hosts: host-gateway` needs Podman 5.3.0+, `deploy.resources` works but podman-compose historically ignored it (now supported), `devices: /dev/net/tun` for WireGuard requires rootful Podman or specific capabilities, and privileged port binding (port 25) requires either rootful mode or `sysctl net.ipv4.ip_unprivileged_port_start=0`. The compose files are well-structured with standard OCI features — no Docker Swarm-specific constructs — which makes dual compatibility achievable with targeted fixes.

The recommended approach is: (1) validate and fix compose files for podman-compose compatibility, (2) document rootful cloud relay and rootless home device configurations, (3) add SELinux `:Z` volume label documentation for Fedora/RHEL, (4) test the full WireGuard tunnel + mail flow on Podman, and (5) update `verify-container-security.sh` to be runtime-aware.

## Recommendation

**Minimal compose changes, maximum documentation.** The compose files are already 90% compatible. Fix the few breaking issues (`host-gateway`, `version` field deprecation warning), add a podman-compose override file pattern for SELinux environments, and create comprehensive Podman setup documentation. Do NOT fork compose files — maintain a single set that works for both Docker and Podman. Use `x-podman` extensions only where absolutely necessary and keep them additive.

Cloud relay should document rootful Podman as the primary path (port 25 binding + /dev/net/tun access). Home device services can run rootless with the `sysctl` privileged port workaround.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| Compose validation | `podman-compose config` | Native validator, catches parse errors before deploy |
| Image builds | `podman build` (reads Dockerfiles natively) | OCI-compatible, no Containerfile rename needed |
| SELinux relabeling | `:Z` volume suffix | Kernel-level, correct MCS labels for bind mounts |
| Privileged port binding (rootless) | `sysctl net.ipv4.ip_unprivileged_port_start` | Kernel setting, established Podman workaround |
| Host access from container | `host.containers.internal` (Podman native) | Built-in since Podman 5.3.0, replaces host-gateway |

## Existing Code and Patterns

- `cloud-relay/docker-compose.yml` — Uses `devices: /dev/net/tun` for WireGuard, `security_opt`, `cap_drop/cap_add`, `deploy.resources.limits.memory`, `tmpfs`, health checks. All standard OCI features supported by Podman. **Risk:** `/dev/net/tun` access requires `--device` pass-through which works rootful but may need `--privileged` or specific device cgroup rules rootless.
- `home-device/docker-compose.yml` — 11 services with profiles, `extra_hosts: host-gateway` on roundcube/snappymail (needs Podman 5.3.0+), `init: true` on radicale (supported), `depends_on` on rspamd→redis (supported). **Risk:** `host-gateway` is the biggest breaking change for older Podman.
- `cloud-relay/certbot/docker-compose.certbot.yml` — Uses `external: true` network. Podman-compose supports external networks. Minimal risk.
- `scripts/verify-container-security.sh` — Hardcoded to check compose YAML structure only (no runtime checks). Needs runtime-awareness for S02 but works as-is since it checks YAML content not Docker CLI.
- `cloud-relay/Dockerfile`, all `home-device/*/Dockerfile` — Multi-stage builds, `HEALTHCHECK`, OCI labels, `ARG TARGETARCH`. All OCI-compliant, buildable with `podman build` unchanged.
- `.github/workflows/build-prebuilt.yml` — Uses `docker/build-push-action` and `docker/setup-buildx-action`. These are Docker-specific GitHub Actions; Podman CI (S04 scope) would use `podman build` directly.
- `cloud-relay/.env.example`, `home-device/.env.example` — Environment variable files. Podman-compose reads `.env` the same way.

## Constraints

- **Compose files must remain valid for both `docker compose` and `podman-compose`.** No podman-only syntax in the main files. Use override files for podman-specific settings.
- **`version: '3.8'` field** — Docker Compose v2 treats this as informational only, but podman-compose may emit deprecation warnings. Keep it for now; removing it is safe for both tools but could confuse users on older Docker Compose v1.
- **Podman minimum version: 5.3.0** — Required for `host-gateway` support in `extra_hosts`. This is a hard requirement given roundcube/snappymail depend on it for mail server access. Podman 5.3.0 released Oct 2024, widely available.
- **podman-compose minimum version: 1.x stable** — Must use podman-compose ≥1.0 for profiles, depends_on conditions, and deploy.resources support.
- **M002 security hardening must not regress** — `security_opt: no-new-privileges`, `cap_drop: ALL`, `cap_add: [specific]`, `read_only: true` all work on Podman. Rootless Podman already drops capabilities by default, so explicit `cap_drop: ALL` is redundant but harmless.
- **SELinux volume labels (`:Z`)** are destructive on non-SELinux systems — must be conditional or documented, not added to main compose files. Override file pattern is the correct approach.
- **Port 25 on cloud relay is non-negotiable** — SMTP must receive on standard port. Rootful Podman or sysctl configuration is the only path.

## Common Pitfalls

- **`host-gateway` on Podman < 5.3.0** — `extra_hosts: mail-server:host-gateway` will produce `invalid IP address` error. Minimum Podman version must be enforced or an alternative documented. Since Podman 5.3.0 is 18 months old, requiring it is reasonable.
- **SELinux `:Z` on non-SELinux systems** — Adding `:Z` unconditionally is safe (it's a no-op on non-SELinux) but `:Z` recursively relabels the entire mount point which is slow on large volumes and destructive if the same host path is shared between containers. Use `:z` (lowercase, shared) for read-only mounts, `:Z` (uppercase, private) for exclusive mounts. Since DarkPipe uses named volumes (not bind mounts) for data, SELinux relabeling mainly affects the `:ro` config file bind mounts.
- **`deploy.resources` silently ignored** — Older podman-compose versions silently skip `deploy.resources.limits.memory`. Current stable versions support it, but users on stale installs may run without memory limits. Document minimum podman-compose version.
- **Rootless networking performance** — Podman rootless uses pasta (formerly slirp4netns) for networking, which has measurable overhead for high-throughput workloads. SMTP relay traffic is low-bandwidth, so this shouldn't matter in practice.
- **Podman pod networking confusion** — podman-compose can optionally create pods (where containers share localhost). Default is `--in-pod=false` (each container gets its own network namespace, matching Docker behavior). Do NOT enable pod mode — it breaks service name DNS resolution.
- **`init: true` on radicale** — Supported by Podman. Uses `catatonit` or `podman --init`. No issue expected.
- **Build context paths** — `cloud-relay/Dockerfile` uses `context: ..` to access root-level `go.mod`. Podman build handles this, but paths must be relative to compose file location. Verified: this pattern works with podman-compose.

## Open Risks

- **WireGuard `/dev/net/tun` in rootless Podman** — The cloud relay requires `devices: /dev/net/tun` and `cap_add: NET_ADMIN`. Rootless Podman may not grant device access without `--privileged`. Cloud relay likely requires rootful Podman. Must test empirically.
- **Postfix setuid helpers with `no-new-privileges`** — Already documented as a known risk in DECISIONS.md. Podman enforces `no-new-privileges` at the seccomp level. If Postfix queue management helpers fail, we need the same workaround as Docker (skip the directive). Existing decision: "document and skip if empirically proven to break."
- **`host-gateway` as hard dependency** — Roundcube and SnappyMail use `extra_hosts: mail-server:host-gateway` to reach the host's mail server. On Podman, `host.containers.internal` is automatically available, but the compose `extra_hosts` directive with `host-gateway` needs Podman 5.3.0+. Alternative: replace with `host.containers.internal` in a podman override, but this changes the hostname the app uses.
- **Health check `||` in CMD form** — Some health checks use `["CMD", "wget", ..., "||", "exit", "1"]`. The `||` is passed as a literal argument in CMD form, not interpreted as shell OR. This is a pre-existing bug in both Docker and Podman (the health check likely works despite this because wget returns non-zero on failure). Should be verified and potentially fixed to use CMD-SHELL form.
- **Podman firewall interaction on Fedora** — Podman rootless on Fedora doesn't auto-configure firewalld like Docker does. Ports below 1024 may need explicit `firewall-cmd --add-port` in addition to the sysctl workaround. Must document.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Podman | `knoopx/pi@podman` (56 installs) | available — directly relevant for Podman compatibility work |
| Podman security | `bagelhole/devops-security-agent-skills@podman` (15 installs) | available — may help with security hardening verification |
| Docker Compose | `manutej/luxor-claude-marketplace@docker-compose-orchestration` (399 installs) | available — less relevant since compose files already exist |

## Sources

- podman-compose `deploy.resources.limits.memory` works with cgroup v2 delegation (source: [braedach.com](https://www.braedach.com/memory-in-podman/))
- `host-gateway` support added in Podman 5.3.0 via pasta `--map-guest-addr` (source: [podman#24133](https://github.com/containers/podman/discussions/24133))
- `host.containers.internal` is Podman's built-in equivalent to `host.docker.internal` (source: [stackoverflow](https://stackoverflow.com/questions/58678983/accessing-host-from-inside-container))
- Rootless Podman port 25 requires `sysctl net.ipv4.ip_unprivileged_port_start=0` (source: [Red Hat Solutions](https://access.redhat.com/solutions/7044059))
- podman-compose feature table: deploy section "not supported" is outdated for current versions; recent sources confirm it works (source: [oneuptime.com comparison](https://oneuptime.com/blog/post/2026-03-04-podman-compose-docker-compose-alternative-rhel-9/view))
- SELinux `:Z` volume labels on Fedora/RHEL are mandatory for bind mounts, no-op elsewhere (source: [betterstack tutorial](https://betterstack.com/community/guides/scaling-docker/podman-compose/))
- podman-compose `x-podman` extensions: `docker_compose_compat`, `default_net_name_compat`, `default_net_behavior_compat` for Docker parity (source: [deepwiki compatibility docs](https://deepwiki.com/containers/podman-compose/5.1-compatibility-issues))
- Podman rootless override file pattern for compose compatibility (source: [jesperronn gist](https://gist.github.com/jesperronn/6612a7604cb3a76bf696fc1338aa1161))
- podman-compose now supports profiles, depends_on conditions, build labels, cache_from/cache_to (source: [podman-compose releases](https://github.com/containers/podman-compose/releases))
