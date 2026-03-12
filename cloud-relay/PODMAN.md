# Cloud Relay — Podman Deployment

## Prerequisites

| Requirement | Minimum Version | Why |
|-------------|----------------|-----|
| Podman | 5.3.0+ | `extra_hosts: host-gateway` support, pasta networking |
| podman-compose | 1.x+ | Profiles, `depends_on` conditions, `deploy.resources` |

## Rootful Podman Required

The cloud relay **must run under rootful Podman** for two reasons:

1. **Port 25 binding** — SMTP requires binding to a privileged port.
2. **`/dev/net/tun` device access** — The WireGuard container needs the TUN device for tunnel creation. Rootless Podman cannot pass through `/dev/net/tun` without `--privileged`.

Run rootful with:

```bash
sudo podman-compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

## Override Files

### Base Podman override (always use)

`docker-compose.podman.yml` enables Docker Compose compatibility flags (`x-podman` extensions) so podman-compose correctly interprets `depends_on` conditions, health checks, and network naming.

```bash
sudo podman-compose -f docker-compose.yml -f docker-compose.podman.yml up -d
```

### SELinux override (Fedora/RHEL/CentOS)

`docker-compose.podman-selinux.yml` adds `:z` labels to bind-mount volumes so containers can read host config files under SELinux enforcing mode. Layer it on top of the base override:

```bash
sudo podman-compose \
  -f docker-compose.yml \
  -f docker-compose.podman.yml \
  -f docker-compose.podman-selinux.yml \
  up -d
```

> **Note:** `:z` (shared label) is safe on non-SELinux systems — it's a no-op. Named volumes are managed by Podman and don't need SELinux labels.

## Known Differences from Docker

### `host-gateway` resolution

Docker resolves `extra_hosts: host-gateway` to the host IP automatically. Podman 5.3.0+ supports this via pasta networking. On Podman, `host.containers.internal` is also available as a built-in hostname inside every container — no `extra_hosts` entry needed.

### Networking

Podman rootful uses netavark for networking (similar to Docker's bridge driver). Container-to-container DNS resolution works the same way as Docker when `--in-pod` is not used (the default).

### Security defaults

Rootless Podman runs with reduced capabilities by default. Since the cloud relay runs rootful, the security model is similar to Docker. The compose file's `cap_drop: ALL` + explicit `cap_add` is honored correctly.

### Memory limits

`deploy.resources.limits.memory` requires cgroup v2 delegation, which is the default on modern Fedora/RHEL and Ubuntu 22.04+. If memory limits appear to have no effect, verify cgroup v2 is active:

```bash
mount | grep cgroup2
```

## Firewall (Fedora/RHEL)

Podman rootful on Fedora does not auto-configure firewalld the way Docker does. You may need to open ports manually:

```bash
sudo firewall-cmd --add-port=25/tcp --permanent
sudo firewall-cmd --add-port=443/tcp --permanent
sudo firewall-cmd --add-port=51820/udp --permanent   # WireGuard
sudo firewall-cmd --reload
```
