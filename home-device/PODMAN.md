# Home Device — Podman Deployment

## Prerequisites

| Requirement | Minimum Version | Why |
|-------------|----------------|-----|
| Podman | 5.3.0+ | `extra_hosts: host-gateway` support (used by Roundcube/SnappyMail) |
| podman-compose | 1.x+ | Profiles, `depends_on` conditions, `deploy.resources` |

## Rootless Option

Unlike the cloud relay, the home device **can run rootless** — it does not require `/dev/net/tun` access. However, mail servers bind privileged ports (25, 587, 993), so you must lower the unprivileged port threshold:

```bash
# Allow rootless containers to bind ports starting from 0
sudo sysctl net.ipv4.ip_unprivileged_port_start=0

# Make persistent across reboots
echo 'net.ipv4.ip_unprivileged_port_start=0' | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
```

Then run without `sudo`:

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile stalwart --profile roundcube up -d
```

Alternatively, run rootful with `sudo` — no sysctl change needed.

## Override Files

### Base Podman override (always use)

`docker-compose.podman.yml` enables Docker Compose compatibility flags (`x-podman` extensions) so podman-compose correctly interprets `depends_on` conditions, health checks, and network naming.

### SELinux override (Fedora/RHEL/CentOS)

`docker-compose.podman-selinux.yml` adds `:z` labels to bind-mount volumes so containers can read host config files under SELinux enforcing mode.

## Startup Commands

The home device uses [Docker Compose profiles](https://docs.docker.com/compose/profiles/) to select mail server, webmail, and CalDAV/CardDAV components. Always include the Podman override file.

**Stalwart + Roundcube (built-in CalDAV/CardDAV):**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile stalwart --profile roundcube up -d
```

**Maddy + SnappyMail + Radicale:**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile maddy --profile snappymail --profile radicale up -d
```

**Postfix+Dovecot + Roundcube + Radicale:**

```bash
podman-compose -f docker-compose.yml -f docker-compose.podman.yml \
  --profile postfix-dovecot --profile roundcube --profile radicale up -d
```

**With SELinux (add the SELinux override to any of the above):**

```bash
podman-compose \
  -f docker-compose.yml \
  -f docker-compose.podman.yml \
  -f docker-compose.podman-selinux.yml \
  --profile stalwart --profile roundcube up -d
```

## Pod Mode Warning

> **Do NOT use `podman-compose --in-pod`** (or `x-podman: in_pod: true`). Pod mode places all containers in a shared network namespace where they communicate via `localhost` — this breaks service-name DNS resolution that DarkPipe relies on for inter-container communication (e.g., rspamd → redis, webmail → mail server).

## Known Differences from Docker

### `host-gateway` resolution

Roundcube and SnappyMail use `extra_hosts: mail-server:host-gateway` to reach the mail server on the host. Podman 5.3.0+ resolves `host-gateway` correctly. On Podman, `host.containers.internal` is also available as a built-in hostname — no extra configuration needed.

### Memory limits

`deploy.resources.limits.memory` requires cgroup v2 delegation. Verify:

```bash
mount | grep cgroup2
```

### Rootless networking

Rootless Podman uses pasta for networking, which has slightly more overhead than Docker's bridge driver. For a home mail server's traffic volume, this is not noticeable.

## Firewall (Fedora/RHEL)

Podman does not auto-configure firewalld. Open the required ports:

```bash
sudo firewall-cmd --add-port=25/tcp --permanent
sudo firewall-cmd --add-port=587/tcp --permanent
sudo firewall-cmd --add-port=993/tcp --permanent
sudo firewall-cmd --add-port=443/tcp --permanent    # Webmail
sudo firewall-cmd --add-port=5232/tcp --permanent   # Radicale (if used)
sudo firewall-cmd --reload
```
