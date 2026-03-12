# S03: Apple Containers Support — Research

**Date:** 2026-03-11

## Summary

Apple Containers (`container` CLI) is Apple's native container runtime shipping with macOS 26, running each OCI container in its own lightweight VM on Apple Silicon. The CLI supports standard Docker/OCI images, `container build` (BuildKit-based), port publishing, volume mounts, environment variables, and resource limits — all with a Docker-like interface. macOS 26 adds inter-container networking via `container network create`, meaning DarkPipe's cloud relay services (Caddy + relay daemon) can communicate over a shared virtual network.

The primary challenge is that Apple Containers has **no compose equivalent** — every container must be started individually via `container run`. DarkPipe's cloud relay uses 2 services (Caddy reverse proxy + relay daemon) that need a shared network, shared volumes, and coordinated startup. A shell-based orchestration script will be needed to replicate what `docker-compose up` does. WireGuard support is the critical unknown: Apple's containers use a custom minimal Linux kernel, and WireGuard kernel module availability is unconfirmed — the relay daemon requires `/dev/net/tun` and `NET_ADMIN` capability equivalent.

The scope is **cloud relay only** (development/testing use case, matching the existing mac-silicon.md positioning). Home device's 9-service profile-based compose is too complex for manual orchestration on Apple Containers and isn't the intended deployment target for macOS anyway.

## Recommendation

Write a platform guide (`deploy/platform-guides/apple-containers.md`) that:

1. Documents prerequisites (macOS 26, Apple Silicon, `container` CLI install via Homebrew or pkg)
2. Provides a shell script (`scripts/apple-containers-start.sh`) that starts cloud relay services with proper networking, volumes, env vars, and startup order
3. Verifies SMTP connectivity (`container exec` + telnet/nc to port 25)
4. Clearly documents limitations: no compose, manual orchestration, WireGuard unknown, dev/testing only
5. Updates `check-runtime.sh` to detect Apple Containers as a recognized runtime

Focus on the cloud relay (2 services: Caddy + relay) — this is tractable for manual orchestration and matches the mac-silicon guide's dev/testing positioning.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| Container CLI reference | [apple/container docs](https://github.com/apple/container/blob/main/docs/command-reference.md) | Official command reference with all flags |
| Networking model | `container network create` + `--network` flag | macOS 26 vmnet supports container-to-container communication on shared networks |
| Image building | `container build` (BuildKit-based) | Reads Dockerfiles natively, supports multi-stage, build args, multi-arch |
| Platform guide structure | `deploy/platform-guides/podman.md` | Follow same structure: prerequisites table, quick start, key differences, troubleshooting |

## Existing Code and Patterns

- `deploy/platform-guides/mac-silicon.md` — Existing Mac guide; already has a forward reference to `apple-containers.md` ("coming soon"). Positions macOS as dev/testing only (not production due to port 25 blocking, NAT). Apple Containers guide should follow this same positioning.
- `deploy/platform-guides/podman.md` — Best template for structure: prerequisites table, quick start commands, key differences section, troubleshooting. Follow this pattern.
- `cloud-relay/docker-compose.yml` — 2 services: `caddy` (image: caddy:2-alpine) and `relay` (custom build). Relay needs `NET_ADMIN`, `/dev/net/tun`, port 25, extensive env vars, tmpfs mounts, and volume mounts. These must be translated to `container run` flags.
- `scripts/check-runtime.sh` — Currently detects Docker and Podman. Needs Apple Containers detection (`container --version` check). Exit code / PASS/FAIL pattern already established.
- `cloud-relay/Dockerfile` — Multi-stage Go build + Alpine runtime. OCI-compatible, should work with `container build` directly (supports `--file`, `--build-arg`, `--target`).

## Constraints

- **macOS 26 required** — Container-to-container networking only works on macOS 26+ (vmnet on macOS 15 isolates containers from each other)
- **Apple Silicon only** — No Intel Mac support; Apple Containers uses Virtualization.framework which requires Apple Silicon
- **No compose tool** — Each container must be `container run` individually; orchestration is manual or scripted
- **Each container is a separate VM** — Higher memory overhead than Docker (default: 4 CPUs, 1GB RAM per container); DarkPipe's 128MB/256MB memory limits translate to `--memory` flags but minimum granularity is 1MB
- **No `host-gateway` equivalent** — Docker's `extra_hosts: host-gateway` doesn't exist; containers on the same `container network` can reach each other by IP; DNS resolution between containers by name needs verification
- **No `security_opt`, `cap_drop`, `cap_add` equivalents** — Apple Containers' VM isolation provides stronger isolation than Linux capabilities, but explicit capability control doesn't exist in the CLI; `--read-only` flag is supported
- **No `depends_on` or health check orchestration** — Startup ordering must be handled by the orchestration script (wait for Caddy before relay, or vice versa)
- **Port 25 binding** — macOS may block inbound port 25 by default; same limitation as Docker on Mac (dev/testing only, not production)
- **Volume mount syntax** — Uses `-v host:container` like Docker; no SELinux `:Z` labels needed (macOS)

## Common Pitfalls

- **Assuming macOS 15 works** — Apple Containers runs on macOS 15 but with **no container-to-container networking**. The guide must require macOS 26 since cloud relay services need to communicate.
- **Forgetting `container system start`** — The container runtime daemon must be started first (`container system start`); this is not automatic on install. First run also prompts for Linux kernel download.
- **macOS firewall prompt** — First container run may trigger a macOS permission prompt for local network access; guide must warn users to click "Allow".
- **WireGuard `/dev/net/tun` access** — Apple's custom Linux kernel may not include the WireGuard module or expose `/dev/net/tun`. The relay daemon needs both. If unavailable, the relay can still work with mTLS transport, but WireGuard transport will fail. The guide must document this as a known limitation with mTLS as fallback.
- **No named volume management** — Apple Containers uses bind mounts (`-v`); there's no volume driver. Persistent data directories must be created on the host explicitly.
- **Resource defaults are generous** — Default 4 CPUs + 1GB RAM per container VM; two containers = 2GB RAM minimum. Users with 8GB Macs may need to set `--memory 512M` or lower.
- **No restart policy** — No `restart: unless-stopped` equivalent; if a container exits, the user must restart it manually or the script must handle it.

## Open Risks

- **WireGuard kernel module in Apple's custom Linux kernel** — This is the #1 unknown. Apple's container VMs use a minimal custom kernel. WireGuard has been in mainline Linux since 5.6, but Apple's kernel config may omit it. If absent, the relay daemon cannot create WireGuard tunnels. Mitigation: document mTLS as the recommended transport for Apple Containers, test WireGuard if/when macOS 26 is available.
- **Custom init image limitations** — Apple Containers supports `--init` and `--init-image` for boot customization, but it's unclear if this enables kernel module loading. The `--kernel` flag allows custom kernels, which could theoretically include WireGuard, but building custom kernels is out of scope for a platform guide.
- **DNS resolution between containers** — On a shared `container network`, containers get IPs (192.168.x.x). It's confirmed containers can reach each other by IP, but it's unconfirmed whether container names resolve via DNS on the same network. The orchestration script may need to use `container inspect` to get IPs and pass them as environment variables.
- **API stability** — Apple Containers is a first-release product (macOS 26). CLI flags, networking behavior, and build semantics may change in future releases. The guide should be versioned and note this.
- **`container build` vs pre-built images** — Building locally with `container build` works but is slower (BuildKit in a VM). Pulling pre-built arm64 images from a registry would be faster but requires DarkPipe to publish images. Current setup assumes local builds. Guide should support both paths.
- **No health check integration** — Compose health checks (`HEALTHCHECK` in Dockerfile) don't translate to any orchestration primitive. The startup script must implement its own readiness checks (curl/wget against endpoints).

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Apple Containers | — | none found (too new, no community skills exist) |
| Container runtimes | — | none directly relevant (k8s-resource-optimizer is Kubernetes-focused) |
| Docker/OCI | — | no general Docker skill found |

No relevant installable skills exist for Apple Containers — the technology is too new (first shipped macOS 26, WWDC 2025).

## Sources

- Apple Containers CLI supports `run`, `build`, `create`, `exec`, `logs`, `network create`, port publishing, volume mounts, env vars, resource limits, read-only filesystem, tmpfs, detach mode (source: [apple/container command-reference.md](https://github.com/apple/container/blob/main/docs/command-reference.md))
- macOS 26 adds `container network create` for isolated networks; containers on same network can communicate; macOS 15 has no inter-container networking (source: [apple/container technical-overview.md](https://github.com/apple/container/blob/main/docs/technical-overview.md))
- Each Apple Container runs in its own lightweight VM with dedicated IP, sub-second startup, custom Linux kernel, EXT4 block devices (source: [InfoQ — Apple Containerization](https://www.infoq.com/news/2025/06/apple-container-linux/))
- Install via `brew install container` or signed pkg from GitHub releases; requires `container system start` before first use; first run downloads Linux kernel (source: [suraj.io — Running Linux Containers Natively](https://suraj.io/post/2026/using-osx-containerization/))
- `container build` uses BuildKit, reads Dockerfile/Containerfile, supports multi-stage builds, build args, multi-arch, cache control (source: [apple/container command-reference.md](https://github.com/apple/container/blob/main/docs/command-reference.md))
- Default per-container resources: 4 CPUs, 1GB RAM; configurable via `--cpus` and `--memory` flags (source: [4sysops — Apple Container vs Docker Desktop](https://4sysops.com/archives/apple-container-vs-docker-desktop/))
- Container-to-container networking confirmed working on macOS 26 via `container network create` + `--network` flag; each network is isolated (source: [apple/container how-to.md](https://github.com/apple/container/blob/main/docs/how-to.md))
- WireGuard is in mainline Linux kernel since 5.6; Apple's custom kernel config is available on GitHub but WireGuard module inclusion is unconfirmed (source: [apple/containerization kernel config](https://github.com/apple/containerization/blob/0.5.0/kernel/config-arm64))
- No compose equivalent exists; multi-service orchestration requires manual scripting (source: project M003-CONTEXT.md risk analysis)
