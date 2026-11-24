# Configuration Reference

Reference for every configurable aspect of DarkPipe: environment variables, compose profiles, mail server config files, transport, DNS, queueing, and monitoring.

> **Container runtime:** all examples use `docker compose`. Podman is supported via `podman-compose` with override files — see the [Podman guide](../deploy/platform-guides/podman.md).

## Environment Variable Files

Both deploy targets ship `.env.example` files that document every variable with its default and a required/optional marker:

- [`cloud-relay/.env.example`](../cloud-relay/.env.example) — relay core, transport, TLS, queue, S3 overflow, caddy
- [`home-device/.env.example`](../home-device/.env.example) — domain, users, mail server, profile server, monitoring, webmail, CalDAV, Rspamd

Copy and edit them in place:

```bash
cp cloud-relay/.env.example cloud-relay/.env
cp home-device/.env.example home-device/.env
```

Environment variables are read from the `.env` file in the same directory as `docker-compose.yml` (docker compose auto-loads it) or exported in the shell.

## Cloud Relay Variables

Source: [`cloud-relay/relay/config/config.go`](../cloud-relay/relay/config/config.go) and [`cloud-relay/docker-compose.yml`](../cloud-relay/docker-compose.yml).

### Core

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_HOSTNAME` | `relay.example.com` | Public FQDN of the relay. **Must match the PTR record** |
| `RELAY_DOMAIN` | `example.com` | Primary mail domain |
| `RELAY_LISTEN_ADDR` | `127.0.0.1:10025` | Address the Go relay daemon listens on (receives from Postfix) |
| `RELAY_TRANSPORT` | `wireguard` | `wireguard` or `mtls` |
| `RELAY_HOME_ADDR` | `10.8.0.2:25` | Home device mail endpoint over the transport |
| `HOME_DEVICE_IP` | `10.8.0.2` | Home device IP on the tunnel; used by the Caddyfile and validation tooling |
| `RELAY_MAX_MESSAGE_BYTES` | `52428800` | Max inbound message size (50 MiB) |
| `RELAY_DEBUG` | `false` | Log full email addresses instead of redacted. **Troubleshooting only** |
| `RELAY_EPHEMERAL_CHECK_INTERVAL` | `60` | Seconds between ephemeral-storage verifications |

### WireGuard transport

| Variable | Default | Description |
|----------|---------|-------------|
| `WIREGUARD_CONFIG` | `/etc/wireguard/wg0.conf` | Path to the WireGuard config inside the container |
| `WIREGUARD_INTERFACE` | `wg0` | WireGuard interface name |

### mTLS transport

All three are required when `RELAY_TRANSPORT=mtls` (no defaults):

| Variable | Description |
|----------|-------------|
| `RELAY_CA_CERT` | Path to the internal CA certificate (e.g. `/certs/ca.crt`) |
| `RELAY_CLIENT_CERT` | Path to the relay's client certificate |
| `RELAY_CLIENT_KEY` | Path to the relay's private key |

Mount the certificate files into the container (e.g. `./certs:/certs:ro`) and set `RELAY_TRANSPORT=mtls`.

### TLS monitoring

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_STRICT_MODE` | `false` | If true, the relay generates a Postfix policy map that refuses plaintext-only connections |
| `RELAY_WEBHOOK_URL` | *(empty)* | Webhook URL for TLS monitoring alerts |
| `CERTBOT_EMAIL` | *(empty)* | Contact address for Let's Encrypt (certbot sidecar) |

### Queue (offline buffering)

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_QUEUE_ENABLED` | `true` | `false` = bounce immediately when the home device is unreachable |
| `RELAY_QUEUE_KEY_PATH` | `/data/queue-keys/identity` | age identity file used to encrypt queued messages |
| `RELAY_QUEUE_MAX_RAM` | `209715200` | Max in-memory queue size in bytes (200 MiB) |
| `RELAY_QUEUE_MAX_MESSAGES` | `10000` | Max queued messages |
| `RELAY_QUEUE_TTL_HOURS` | `168` | Max age before a queued message is purged (7 days) |
| `RELAY_QUEUE_SNAPSHOT_PATH` | `/data/queue-state/snapshot.json` | Queue metadata snapshot |

### S3 overflow

Spill the queue to object storage when the local queue is full.

| Variable | Default | Description |
|----------|---------|-------------|
| `RELAY_OVERFLOW_ENABLED` | `false` | Enable S3 overflow |
| `RELAY_OVERFLOW_ENDPOINT` | *(required if enabled)* | S3-compatible endpoint, e.g. `gateway.storjshare.io` |
| `RELAY_OVERFLOW_BUCKET` | `darkpipe-queue` | Bucket name |
| `RELAY_OVERFLOW_ACCESS_KEY` | *(required if enabled)* | S3 access key |
| `RELAY_OVERFLOW_SECRET_KEY` | *(required if enabled)* | S3 secret key |
| `RELAY_OVERFLOW_USE_SSL` | `true` | Use TLS for S3 connections |

The entrypoint also supports `_FILE` variants (`RELAY_WEBHOOK_URL_FILE`, `CERTBOT_EMAIL_FILE`, `RELAY_OVERFLOW_ACCESS_KEY_FILE`, `RELAY_OVERFLOW_SECRET_KEY_FILE`) that read the value from a file — useful with docker secrets.

### Caddy reverse proxy (cloud relay)

The relay's caddy terminates TLS for `mail.example.com` (the default `WEBMAIL_DOMAINS`) and proxies webmail, autodiscovery, profiles, and the dashboard into the tunnel. These domain names must resolve to the relay: caddy obtains their Let's Encrypt certificates itself, and they are how clients reach everything DarkPipe serves over HTTPS.

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBMAIL_DOMAINS` | `mail.${RELAY_DOMAIN}` | Site block for webmail (proxied to `{$HOME_DEVICE_IP}:8080`) |
| `AUTOCONFIG_DOMAINS` | `autoconfig.${RELAY_DOMAIN}` | Thunderbird autoconfig site block |
| `AUTODISCOVER_DOMAINS` | `autodiscover.${RELAY_DOMAIN}` | Outlook autodiscover site block |
| `ADMIN_USER` / `ADMIN_PASSWORD_HASH` | — | Basic auth for the proxied `/health/*`, `/status*` routes |

## Home Device Variables

Source: [`home-device/docker-compose.yml`](../home-device/docker-compose.yml) and [`home-device/profiles`](../home-device/profiles/).

### Domain and identity

| Variable | Default | Description |
|----------|---------|-------------|
| `MAIL_DOMAIN` | `example.com` | Primary mail domain (required) |
| `MAIL_HOSTNAME` | `mail.example.com` | Name placed into onboarding payloads (autoconfig/autodiscover/mobileconfig) — the address your mail clients connect to directly (LAN name, or a public name with 993/587 forwarded). Mail ports are never proxied by the relay, so this is usually *not* the relay hostname. (required) |
| `ADMIN_EMAIL` | `admin@example.com` | First admin user (required) |
| `ADMIN_PASSWORD` | `changeme` | Admin password — **change this** (required) |
| `MAIL_SERVER_TYPE` | `stalwart` | `stalwart`, `maddy`, or `postfix-dovecot` (used by the profile server) |
| `VIRTUAL_DOMAINS` | — | Comma-separated extra domains (Postfix+Dovecot only) |

### Profile server

| Variable | Default | Description |
|----------|---------|-------------|
| `PROFILE_SERVER_PORT` | `8090` | Port the profile server listens on |
| `CALDAV_URL` | *(auto for Stalwart)* | CalDAV URL embedded in generated profiles (set for Radicale) |
| `CARDDAV_URL` | *(auto for Stalwart)* | CardDAV URL embedded in generated profiles (set for Radicale) |
| `CALDAV_PORT` / `CARDDAV_PORT` | `443` | Ports used when constructing CalDAV/CardDAV URLs |
| `APP_PASSWORD_STORE_PATH` | `/data/app-passwords.json` | Where the profile server persists app passwords |
| `PROFILE_DEBUG` | `false` | Log full email addresses in profile server logs. Read directly by the profile server — add it to the `profile-server` service's `environment:` block in the compose file. **Troubleshooting only** |
| `STALWART_ADMIN_USER` / `STALWART_ADMIN_PASSWORD` | from `ADMIN_EMAIL`/`ADMIN_PASSWORD` | Stalwart management UI credentials (compose passes them through) |

### Monitoring

| Variable | Default | Description |
|----------|---------|-------------|
| `MONITOR_ALERT_EMAIL` | *(empty)* | Email address for health alerts |
| `MONITOR_WEBHOOK_URL` | *(empty)* | Webhook URL for alerts (Slack, Discord, …) |
| `MONITOR_HEALTHCHECK_URL` | *(empty)* | External uptime pings (e.g. Healthchecks.io) |
| `MONITOR_CERT_PATHS` | *(empty)* | Comma-separated certificate paths to watch for expiry |
| `MONITOR_LOG_PATH` | `/var/log/mail.log` | Mail delivery log to parse for delivery status |
| `MONITOR_CLI_ALERT_PATH` | `/data/monitoring/cli-alerts.json` | CLI alert state file |

### Webmail and CalDAV

| Variable | Default | Description |
|----------|---------|-------------|
| `ROUNDCUBEMAIL_DB_TYPE` | `sqlite` | Roundcube database type |
| `ROUNDCUBEMAIL_UPLOAD_MAX_FILESIZE` | `25M` | Roundcube upload limit |
| `UPLOAD_MAX_SIZE` | `25M` | SnappyMail upload limit |
| `LOG_TO_STDOUT` | `true` | SnappyMail log to stdout |
| `MEMORY_LIMIT` | `128M` | SnappyMail PHP memory limit |
| `RADICALE_CONFIG` | `/config/config` | Radicale config path |
| `DKIM_SELECTOR` | `darkpipe` | Rspamd DKIM signing selector — set to the selector published by `dns-setup` (e.g. `darkpipe-2026q4`) |

### Mail server config files

Each mail server reads files under `home-device/` (mounted read-only):

| Mail server | Config file(s) |
|-------------|----------------|
| Stalwart | [`home-device/stalwart/config.toml`](../home-device/stalwart/config.toml) — listeners: SMTP 25, submission 587 (STARTTLS), IMAPS 993, management 8080 |
| Maddy | [`home-device/maddy/maddy.conf`](../home-device/maddy/maddy.conf) |
| Postfix+Dovecot | [`home-device/postfix-dovecot/`](../home-device/postfix-dovecot/) — `postfix/vmailbox`, `postfix/virtual`, `dovecot/users` |
| Radicale | [`home-device/caldav-carddav/radicale/`](../home-device/caldav-carddav/radicale/) — `config`, `rights`, `users` |
| Rspamd | [`home-device/spam-filter/rspamd/local.d/`](../home-device/spam-filter/rspamd/local.d/) — thresholds, greylisting, DKIM signing |
| Redis | [`home-device/spam-filter/redis/redis.conf`](../home-device/spam-filter/redis/redis.conf) |

## Container Security Configuration

All compose services carry hardening directives (do not remove them):

- `security_opt: [no-new-privileges:true]`
- `cap_drop: [ALL]` with selective `cap_add` only where required (e.g. `NET_BIND_SERVICE`)
- `read_only: true` with explicit `tmpfs` mounts for writable paths
- `HEALTHCHECK` in every custom Dockerfile

Audit them with:

```bash
bash scripts/verify-container-security.sh
bash scripts/verify-log-redaction.sh
```

## Compose Profiles

The home device compose file uses profiles. Select exactly one mail server and one webmail; add `radicale` only when the mail server has no built-in CalDAV/CardDAV.

| Profile | Component |
|---------|-----------|
| `stalwart` | Stalwart 0.15.4 (IMAP4rev2/JMAP, built-in CalDAV/CardDAV) |
| `maddy` | Maddy 0.8.2 (minimal Go binary) |
| `postfix-dovecot` | Postfix + Dovecot (traditional) |
| `roundcube` | Roundcube 1.6.13 |
| `snappymail` | SnappyMail 2.38.2 |
| `radicale` | Radicale 3.6.0 (Maddy / Postfix+Dovecot only) |

`rspamd`, `redis`, and `profile-server` always run.

```bash
docker compose --profile stalwart --profile snappymail up -d                 # default stack
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d
docker compose --profile maddy --profile snappymail --profile radicale up -d  # low-RAM stack
```

## Transport Configuration

### WireGuard

Config files live at `/etc/wireguard/wg0.conf` on each machine (installed by the setup scripts — see [quick start](quickstart.md#step-2-set-up-the-encrypted-transport)). Cloud: `10.8.0.1/24`, `ListenPort 51820`. Home: `10.8.0.2/24`, `PersistentKeepalive 25`.

```bash
sudo wg-quick up wg0      # start
sudo wg-quick down wg0    # stop
sudo wg show              # status
```

### mTLS

```bash
# cloud-relay/.env
RELAY_TRANSPORT=mtls
RELAY_CA_CERT=/certs/ca.crt
RELAY_CLIENT_CERT=/certs/relay.crt
RELAY_CLIENT_KEY=/certs/relay.key
```

Certificates come from the internal step-ca CA (`deploy/pki/step-ca-setup.sh`); the relay mounts them read-only. Automatic renewal is available via the `cert-renewer@` systemd timer in [`transport/pki/renewal`](../transport/pki/renewal/).

## DNS

The `dns-setup` tool is documented in full in [reference/cli.md](reference/cli.md). Records it manages: MX, A, SPF (`v=spf1 ip4:<relay-ip> -all`), DKIM (time-based selector `<prefix>-<year>q<quarter>`), DMARC (`p=` + `sp=quarantine` + `rua`/`ruf`), mail SRV records, and autoconfig/autodiscover CNAMEs.

## Offline Queue

When the home device is unreachable, inbound mail is encrypted with age and queued; when it reconnects, the queue drains.

- Queue vs bounce: `RELAY_QUEUE_ENABLED` (default `true`)
- Purge after `RELAY_QUEUE_TTL_HOURS` (default 168 = 7 days)
- RAM cap `RELAY_QUEUE_MAX_RAM` / message cap `RELAY_QUEUE_MAX_MESSAGES`; beyond that, S3 overflow when enabled
- The age identity key is generated locally on the relay at `RELAY_QUEUE_KEY_PATH`; public key is derived from it

## Monitoring

- Dashboard: profile server `/status` (Basic auth when proxied via the relay)
- API: `/status/api`
- Liveness/readiness: `/health/live`, `/health/ready`
- Alert channels: `MONITOR_ALERT_EMAIL`, `MONITOR_WEBHOOK_URL`, `MONITOR_HEALTHCHECK_URL`
- Certificate expiry monitoring: `MONITOR_CERT_PATHS` (alerts within 7 days of expiry)
- Compose healthchecks: 30 s interval, 10 s timeout, 3 retries (profile server: 5 s timeout, 10 s start period)

## Releases and Custom Builds

### Pre-built images (from releases)

Version tags publish multi-arch images to GitHub Container Registry:

```
ghcr.io/<owner>/darkpipe/cloud-relay:vX.Y.Z
ghcr.io/<owner>/darkpipe/home-stalwart:vX.Y.Z-default
ghcr.io/<owner>/darkpipe/home-postfix-dovecot:vX.Y.Z-conservative
```

plus `darkpipe-setup-<os>-<arch>` and `dns-setup-<os>-<arch>` binaries. `<owner>` is the repository's GitHub owner.

### Custom builds

Fork the repository, then run Actions → **Build Custom Stack** with:

- `mail_server`: `stalwart` / `maddy` / `postfix-dovecot`
- `webmail`: `none` / `roundcube` / `snappymail`
- `calendar`: `none` / `builtin` / `radicale`
- `tag_suffix`: optional (e.g. `custom-01`)

It publishes `ghcr.io/<your-fork>/darkpipe/home-<mail_server>:vX.Y.Z[-suffix]` and a matching `cloud-relay` image. Point your compose file's `image:` field at them.

### Updating

```bash
docker pull ghcr.io/<owner>/darkpipe/cloud-relay:vX.Y.Z
cd cloud-relay && docker compose up -d
# same for the home device image(s)
docker compose ps   # verify health
```

---

Last Updated: 2026-11-23

License: AGPLv3 — see [LICENSE](../LICENSE)
