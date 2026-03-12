# DarkPipe on Apple Containers

Run the DarkPipe cloud relay on macOS 26+ using Apple Containers — Apple's native lightweight container runtime for Apple Silicon.

> **General macOS guidance:** For Docker Desktop or OrbStack on Mac, see [Mac Silicon Guide](mac-silicon.md). This guide covers the Apple Containers runtime specifically.

> **Development/Testing Only:** Like all macOS deployments, Apple Containers is intended for development and testing. macOS blocks inbound port 25 by default and is typically behind NAT — see [Mac Silicon Guide → Purpose](mac-silicon.md#purpose-developmenttesting-only) for details.

## Prerequisites

| Requirement | Detail | Why |
|-------------|--------|-----|
| macOS 26+ (Tahoe) | Apple Containers requires macOS 26 | Container-to-container networking (`container network create`) requires macOS 26; macOS 15 isolates containers from each other |
| Apple Silicon | M1, M2, M3, M4, or later | Apple Containers uses Virtualization.framework (Apple Silicon only, no Intel support) |
| `container` CLI | `brew install container` | Apple's native container CLI for building, running, and managing containers |
| RAM | 8GB minimum, 16GB recommended | Each container runs in its own lightweight VM with ~1GB RAM default; two services = ~2GB minimum |

**Install and verify:**

```bash
# Install Apple Containers CLI
brew install container

# Start the container runtime daemon (required before first use)
container system start

# Verify installation
container --version
```

> **First run:** `container system start` downloads a minimal Linux kernel on first invocation. This is a one-time setup step.

## Quick Start

DarkPipe's cloud relay requires two services (Caddy reverse proxy + relay daemon) on a shared network. Since Apple Containers has no compose tool, an orchestration script handles startup.

```bash
# 1. Start the container runtime (if not already running)
container system start

# 2. Start the cloud relay (creates network, builds images, starts containers)
bash scripts/apple-containers-start.sh up

# 3. Check container status
bash scripts/apple-containers-start.sh status

# 4. View logs
bash scripts/apple-containers-start.sh logs

# 5. Tear everything down
bash scripts/apple-containers-start.sh down
```

**Preview without executing (no macOS 26 required):**

```bash
# Print all commands without running them
bash scripts/apple-containers-start.sh --dry-run up

# Print with full environment variable expansion
bash scripts/apple-containers-start.sh --dry-run --verbose up
```

The orchestration script translates `cloud-relay/docker-compose.yml` into individual `container` CLI commands. See `bash scripts/apple-containers-start.sh --help` for all options.

## Key Differences from Docker

### VM-per-container isolation

Every Apple Container runs in its own lightweight VM with a dedicated Linux kernel. This provides stronger isolation than Docker's shared-kernel model, but with higher per-container resource overhead.

**Implications:**
- No `security_opt`, `cap_add`, or `cap_drop` flags — VM boundary replaces Linux capability control
- No `--device /dev/net/tun` pass-through — each VM has its own device tree
- Sub-second startup despite VM overhead

### No compose tool

Apple Containers has no Docker Compose equivalent. Multi-service orchestration requires scripting each `container run` command individually.

DarkPipe provides [`scripts/apple-containers-start.sh`](../../scripts/apple-containers-start.sh) to automate this for the cloud relay (Caddy + relay daemon).

### No restart policy

There is no `restart: unless-stopped` equivalent. If a container exits, restart it manually or re-run the orchestration script:

```bash
bash scripts/apple-containers-start.sh down
bash scripts/apple-containers-start.sh up
```

### Resource defaults

Each container VM defaults to 4 CPUs and 1GB RAM. The orchestration script sets explicit memory limits (`--memory 128M` for Caddy, `--memory 256M` for the relay) to reduce footprint.

Override if needed:
```bash
# Check current resource usage
container list
```

On 8GB Macs, two containers at default settings (1GB each) consume 25% of total RAM. The orchestration script's reduced limits keep usage under 500MB total.

### Bind mounts only (no named volumes)

Apple Containers has no volume driver. All persistent data uses host bind mounts (`-v /host/path:/container/path`). The orchestration script creates host directories under `data/` automatically:

```
data/
├── caddy/
│   ├── data/
│   ├── config/
│   └── logs/
└── relay/
    ├── postfix-queue/
    ├── certbot-etc/
    └── queue-data/
```

### No `host-gateway`

Docker's `extra_hosts: host-gateway` has no Apple Containers equivalent. Containers on the same `container network` can reach each other by IP address. If container name DNS resolution is unavailable, the orchestration script uses direct IP addressing.

### `container system start` required

The container runtime daemon must be started before any `container` commands work:

```bash
container system start
```

This is not automatic on install — the orchestration script does not start it for you (it requires user consent for kernel download on first run).

## Transport Configuration

### mTLS (recommended)

mTLS is the **recommended default transport** for Apple Containers. Set in `cloud-relay/.env`:

```bash
RELAY_TRANSPORT=mtls
```

The orchestration script defaults to `mtls` if `RELAY_TRANSPORT` is not set.

### WireGuard (unknown compatibility)

Apple Containers runs a custom minimal Linux kernel. **WireGuard kernel module availability is unconfirmed** — the kernel configuration may omit it. If WireGuard is unavailable:

- The relay daemon cannot create WireGuard tunnels
- `/dev/net/tun` device access may not work inside the container VM
- No `--cap-add NET_ADMIN` equivalent exists (VM isolation replaces capabilities)

**Do not use `RELAY_TRANSPORT=wireguard` on Apple Containers** unless you have confirmed WireGuard works on your macOS 26 installation. To test:

```bash
# Inside a running container, check for WireGuard module
container exec darkpipe-relay -- ls /sys/module/wireguard
```

If WireGuard is needed, consider Docker Desktop or OrbStack instead (see [Mac Silicon Guide](mac-silicon.md)).

## Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No compose tool | Each container started individually | Use orchestration script |
| Manual orchestration | No automatic dependency ordering or health check integration | Script handles startup order and readiness checks |
| WireGuard uncertain | Tunnel transport may not work | Use mTLS transport (default) |
| Dev/testing only | Port 25 blocked on macOS, typically behind NAT | Same as Docker on Mac — deploy cloud relay on VPS for production |
| No restart policy | Containers don't auto-restart on crash | Re-run `up` or monitor externally |
| No health check integration | Compose `HEALTHCHECK` directives don't apply | Script implements its own readiness checks |
| API may change | Apple Containers is a first-release product (macOS 26) | Pin to tested versions; check release notes on upgrade |
| Higher memory overhead | VM-per-container model uses more RAM than Docker | Set `--memory` flags (orchestration script does this) |

## Troubleshooting

### `container: command not found`

**Cause:** Apple Containers CLI not installed.

**Fix:**
```bash
brew install container
```

### Containers fail to start — "system not started"

**Cause:** The container runtime daemon is not running.

**Fix:**
```bash
container system start
```

This must be run after every macOS reboot. The first invocation downloads a Linux kernel (~50MB).

### macOS firewall prompt on first run

**Cause:** macOS prompts for network access permission when containers first bind ports.

**Fix:** Click **Allow** when the macOS firewall dialog appears. If you dismissed it, go to **System Settings → Network → Firewall → Options** and add the `container` process.

### Containers can't reach each other (macOS 15)

**Cause:** Container-to-container networking requires macOS 26. On macOS 15, each container is network-isolated — there is no shared network.

**Fix:** Upgrade to macOS 26 (Tahoe). There is no workaround on macOS 15 for multi-service deployments.

### High memory usage

**Cause:** Each container VM defaults to 1GB RAM. Two containers = 2GB minimum at default settings.

**Fix:** The orchestration script already sets reduced limits (`--memory 128M` for Caddy, `--memory 256M` for relay). To adjust further, edit the `--memory` values in `scripts/apple-containers-start.sh`.

### Container-to-container DNS not resolving

**Cause:** Container name DNS on shared networks may not work in all macOS 26 builds.

**Fix:** If containers can't resolve each other by name, use IP addresses. Check container IPs with:

```bash
container inspect caddy
container inspect darkpipe-relay
```

The orchestration script uses the `--network` flag to place containers on the same network, which should enable IP-level connectivity.

### Port 25 not reachable from outside

**Cause:** macOS blocks inbound port 25 by default (anti-spam). This is a macOS platform limitation, not an Apple Containers issue.

**Fix:** This cannot be fixed on macOS. For production mail flow, deploy the cloud relay on a Linux VPS with port 25 open. Apple Containers on Mac is for development and testing only.

## Resources

- **Orchestration script:** [`scripts/apple-containers-start.sh`](../../scripts/apple-containers-start.sh) — start/stop/status/logs for cloud relay
- **Mac Silicon guide:** [`deploy/platform-guides/mac-silicon.md`](mac-silicon.md) — general macOS guidance, Docker Desktop / OrbStack
- **Runtime validation:** [`scripts/check-runtime.sh`](../../scripts/check-runtime.sh) — detects Apple Containers alongside Docker and Podman
- **Apple Containers docs:** [github.com/apple/container](https://github.com/apple/container) — official CLI reference
- **Technical overview:** [Apple Containers architecture](https://github.com/apple/container/blob/main/docs/technical-overview.md) — VM model, networking, kernel
