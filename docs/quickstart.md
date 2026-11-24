# Quick Start: How to Deploy DarkPipe

This guide takes you from zero to a working DarkPipe deployment: a cloud relay on a VPS, an encrypted tunnel, and a mail server on your home device — sending and receiving real email.

**Time required:** 30–60 minutes of active setup, plus 4–6 weeks of gradual IP warmup before your outbound volume is treated normally by receiving providers.

**What you will have at the end:**

- A cloud relay that receives mail for your domain and forwards it over an encrypted tunnel
- A mail server (Stalwart by default) on your home device, with webmail, spam filtering, and device onboarding
- SPF, DKIM, and DMARC records published for your domain
- Your devices configured via the profile server

For a risk-free first experience, start with the [First Email tutorial](tutorials/first-email.md), which walks a single default path step by step.

## Container Runtime

All examples use `docker compose`. **Podman is fully supported** via `podman-compose` with override files (`-f docker-compose.podman.yml`). See the [Podman platform guide](../deploy/platform-guides/podman.md) for details, including rootful vs rootless requirements.

Validate your environment with:

```bash
bash scripts/check-runtime.sh
```

## Before You Start

1. **A domain name you control** — with DNS API access (Cloudflare or Route53) for automation, or the ability to add DNS records manually.
2. **A VPS with port 25 access** — see [VPS provider selection](vps-providers.md). Recommended: Hetzner, Vultr, OVH, Linode. Minimum: 1 vCPU, 1 GB RAM, 20 GB SSD ($3–6/month).
3. **A home device running containers** — Raspberry Pi 4+ (4 GB RAM recommended) or any x64/arm64 Linux host with Docker 24+ (or Podman 5.3+), plus the [platform guide](../deploy/platform-guides/) for your platform if you have one (NAS, Proxmox, macOS, …).
4. **DNS names** — `dns-setup` publishes most of these in Step 4; you add the `mail` A record yourself:
   - `relay.example.com` → your VPS IP (A record)
   - `mail.example.com` → **your VPS IP** (A record, same IP as the relay) — webmail, autodiscovery, the dashboard, and profile downloads are all served by the relay under this name, so it must resolve to the relay.
   - Mail *clients* (IMAP, SMTP submission, CalDAV) connect to the **home device** directly, at the address you put in `MAIL_HOSTNAME` (Step 3): its LAN address for home-only clients, or a public name (e.g., dynamic DNS) with 993/587 forwarded for remote access.

## Step 1: Provision the Cloud Relay

1. Provision the smallest VPS that allows port 25.
2. Set the hostname:

   ```bash
   hostnamectl set-hostname relay.example.com
   ```

3. Set the reverse DNS (PTR) record in your provider's control panel so the VPS IP resolves to `relay.example.com`. Many providers require this step explicitly.
4. Update the system and install Docker:

   ```bash
   apt update && apt upgrade -y
   curl -fsSL https://get.docker.com | sh
   ```

   > **Podman users:** install Podman 5.3+ and podman-compose instead; the cloud relay must run rootful. See the [Podman guide](../deploy/platform-guides/podman.md).

5. Verify port 25 is open (outbound):

   ```bash
   apt install -y telnet
   telnet gmail-smtp-in.l.google.com 25
   ```

   You should see a `220 ... ESMTP` greeting. Type `QUIT`. If the connection times out, port 25 is blocked — contact your provider or use a different one (see [vps-providers.md](vps-providers.md)).

6. Clone the repository:

   ```bash
   git clone https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe.git
   cd darkpipe
   ```

## Step 2: Set Up the Encrypted Transport

WireGuard is the recommended transport: a full encrypted tunnel between cloud and home, with no firewall holes on the home device. The tunnel uses network `10.8.0.0/24` — cloud `10.8.0.1`, home `10.8.0.2`.

### 2.1 Generate keys

On each machine (VPS and home device):

```bash
# Install WireGuard tools (the setup scripts in 2.3 do this too)
sudo apt install -y wireguard-tools

# Generate a key pair; remember the two outputs
wg genkey            # private key
# press Enter to derive the public key:
wg pubkey
```

Write down: **cloud private key, cloud public key, home private key, home public key**.

### 2.2 Write the config files

**Cloud config** (`wg0.conf` on the VPS):

```ini
[Interface]
Address = 10.8.0.1/24
PrivateKey = <cloud-private-key>
ListenPort = 51820

[Peer]
PublicKey = <home-public-key>
AllowedIPs = 10.8.0.2/32
```

**Home config** (`wg0.conf` on the home device):

```ini
[Interface]
Address = 10.8.0.2/24
PrivateKey = <home-private-key>

[Peer]
PublicKey = <cloud-public-key>
Endpoint = <VPS-public-IP>:51820
AllowedIPs = 10.8.0.1/32
PersistentKeepalive = 25
```

These match the templates in [`transport/wireguard/config`](../../transport/wireguard/config/templates.go). `PersistentKeepalive` matters for NAT traversal behind your home router.

### 2.3 Install and start

Both helper scripts take a path to a ready `wg0.conf`, install it, enable the systemd service with auto-restart, and (on the cloud) enable IP forwarding and open UDP 51820 in ufw.

**On the VPS:**

```bash
curl -LO https://raw.githubusercontent.com/The-Artificer-of-Ciphers-LLC/darkpipe/main/deploy/wireguard/cloud-setup.sh
chmod +x cloud-setup.sh
sudo ./cloud-setup.sh /path/to/wg0.conf
```

**On the home device:**

```bash
curl -LO https://raw.githubusercontent.com/The-Artificer-of-Ciphers-LLC/darkpipe/main/deploy/wireguard/home-setup.sh
chmod +x home-setup.sh
sudo ./home-setup.sh /path/to/wg0.conf
```

If you open a router port: forward UDP 51820 to the home device. The home side initiates the connection, so no inbound rule is strictly required if the VPS is the hub.

### 2.4 Verify

```bash
# On the VPS
ping -c 3 10.8.0.2
# On the home device
ping -c 3 10.8.0.1
```

Both directions must succeed. If not: check `sudo wg show` on both sides, confirm the public keys in each config match the *other* machine, and confirm UDP 51820 is reachable (firewall on the VPS).

### mTLS alternative

mTLS is a minimal-footprint alternative that needs no tunnel: the relay dials the home device over mutual TLS using certificates from an internal step-ca PKI. The relay side is ready (`RELAY_TRANSPORT=mtls` plus `RELAY_CA_CERT`/`RELAY_CLIENT_CERT`/`RELAY_CLIENT_KEY` — see [configuration](configuration.md#mtls-transport)), and `deploy/pki/step-ca-setup.sh` provisions the internal CA on the home device. The checked-in home device compose files assume the WireGuard tunnel, so **WireGuard is the supported path for the standard deployment**; treat mTLS as an advanced, manual integration (see [`transport/mtls`](../../transport/mtls/)).

## Step 3: Configure the Deployment

There are two ways to produce deployment configuration. Pick one:

### Option A (recommended): the `.env` files

Both deploy targets ship `.env.example` files documenting every variable with defaults:

```bash
# On the VPS
cd darkpipe/cloud-relay
cp .env.example .env

# On the home device
cd darkpipe/home-device
cp .env.example .env
```

Edit each `.env` with your real values. The variables that must change:

**cloud-relay/.env**

```bash
RELAY_HOSTNAME=relay.example.com     # must match your PTR record
RELAY_DOMAIN=example.com
RELAY_HOME_ADDR=10.8.0.2:25          # home device on the tunnel
HOME_DEVICE_IP=10.8.0.2
```

**home-device/.env**

```bash
MAIL_DOMAIN=example.com
MAIL_HOSTNAME=mail.example.com      # set to the address your clients reach — see note below
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-me             # CHANGE THIS
DKIM_SELECTOR=darkpipe-2026q4        # the selector dns-setup publishes (Step 4)
```

> **Note on `MAIL_HOSTNAME`.** The onboarding profiles and the autoconfig/autodiscover payloads use this name, and mail *clients* connect to it directly — the relay never proxies mail ports (993/587/5232) into the tunnel. Set it to the address your clients actually reach: the home device's LAN name or IP for home-only clients, or a public name (e.g., a dynamic-DNS hostname) with 993/587 port-forwarded for remote clients. The relay-side webmail/dashboard/autodiscovery services are addressed separately by `WEBMAIL_DOMAINS` in `cloud-relay/.env` (default `mail.${RELAY_DOMAIN}`) — that name must resolve to the relay.

Full variable reference: [configuration.md](configuration.md).

### Option B: the setup wizard

Download the `darkpipe-setup` release asset for your platform and run the `setup` subcommand:

```bash
curl -LO https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64
./darkpipe-setup-linux-amd64 setup
```

The wizard asks (with live validation where possible):

1. Setup mode — Quick (recommended defaults) or Advanced
2. Primary mail domain
3. Cloud relay hostname (tests port 25)
4. Admin email
5. Mail server — `stalwart` (default), `maddy`, or `postfix-dovecot`
6. Webmail — `snappymail` (default), `roundcube`, or none
7. Calendar/Contacts — Stalwart built-in, Radicale, or none
8. Transport — `wireguard` (default) or `mtls`

It then generates in the current directory:

- `docker-compose.yml` — a single combined compose file with your selected services
- `.darkpipe.yml` — the saved configuration
- `secrets/admin_password.txt` — a **generated** 24-character admin password (note it; the wizard does not take a password as input)
- `secrets/dkim_private_key.pem` — a placeholder to replace with the key from Step 4
- `.darkpipe-configured` — completion marker

Re-running `setup` in the same directory offers to upgrade the existing configuration.

## Step 4: Configure DNS

The `dns-setup` tool generates and validates all required records: MX, A, SPF, DKIM, DMARC, mail SRV records, and autoconfig/autodiscover CNAMEs.

Download the release asset:

```bash
curl -LO https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/dns-setup-linux-amd64
chmod +x dns-setup-linux-amd64
```

### 4.1 Preview (dry-run — the default)

```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip YOUR_VPS_IP
```

The run prints every record, writes a `DNS-RECORDS.md` guide for manual entry, and generates the DKIM key pair in `/etc/darkpipe/dkim/` (selector format: `darkpipe-<year>q<quarter>`, e.g. `darkpipe-2026q4`).

### 4.2 Apply

The tool auto-detects your DNS provider from the domain's NS records. If it detects **Cloudflare**, export the token first; for **Route53**, provide standard AWS credentials. There is no `--provider` flag:

```bash
export CLOUDFLARE_API_TOKEN=your_token   # Cloudflare hosts
# or: export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=...   # Route53 hosts

./dns-setup-linux-amd64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip YOUR_VPS_IP \
  --apply
```

For any other provider, add the records from the printed list (or `DNS-RECORDS.md`) in your DNS control panel.

### 4.3 Install the DKIM key on the home device

Outbound mail is DKIM-signed by **Rspamd on the home device** (not on the relay). After generating the key:

```bash
# On the machine where dns-setup ran — note the selector it printed, e.g. darkpipe-2026q4
scp /etc/darkpipe/dkim/darkpipe-2026q4.private.pem user@home-device:~/

# On the home device, from the repo clone's root directory — the key must be
# named <domain>.<selector>.key:
cd ~/darkpipe   # wherever you cloned the repository
mkdir -p home-device/spam-filter/rspamd/dkim-keys
mv ~/darkpipe-2026q4.private.pem home-device/spam-filter/rspamd/dkim-keys/example.com.darkpipe-2026q4.key
```

And set the same selector in `home-device/.env` (`DKIM_SELECTOR=darkpipe-2026q4`) so Rspamd signs with the selector published in DNS.

### 4.4 Validate

After propagation (5–60 minutes):

```bash
./dns-setup-linux-amd64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip YOUR_VPS_IP \
  --validate-only
```

All checks (SPF, DKIM, DMARC, MX, SRV, autoconfig/autodiscover CNAMEs, PTR) should pass.

You can also have the tool send a DKIM-signed test message:

```bash
./dns-setup-linux-amd64 --domain example.com --relay-hostname relay.example.com --relay-ip YOUR_VPS_IP --send-test you@external-mail.com
```

## Step 5: Deploy the Services

### 5.1 Cloud relay (on the VPS)

```bash
cd darkpipe/cloud-relay
docker compose up -d
docker compose ps
```

Expected containers: `darkpipe-relay` (Postfix + Go relay daemon) and `caddy` (reverse proxy for webmail and autodiscovery).

**Let's Encrypt for the relay hostname.** Postfix obtains its TLS certificate from the certbot sidecar, which is a *separate* compose file (it also needs port 80, which caddy holds). Use this sequence:

```bash
# 1. Start the relay first (it runs without TLS until a certificate appears)
docker compose up -d relay

# 2. Run the certbot sidecar (set CERTBOT_EMAIL in .env first)
docker compose -f docker-compose.yml -f certbot/docker-compose.certbot.yml up -d certbot
docker compose logs certbot | tail   # expect "Certificate obtained successfully"

# 3. Then bring up caddy
docker compose up -d caddy
```

> **Port 80 note:** caddy and the certbot sidecar both want port 80 for HTTP-01 challenges. If the sidecar logs a bind failure, `docker compose stop caddy` while the initial certificate is obtained, then restart caddy. For unattended renewal alongside caddy, switch the sidecar to a DNS-01 challenge (`--dns-cloudflare` with `CLOUDFLARE_API_TOKEN`).

The relay entrypoint watches the certificate directory and reloads Postfix automatically when the certificate appears.

### 5.2 Home device

```bash
cd darkpipe/home-device
docker compose --profile stalwart --profile snappymail up -d
docker compose ps
```

Other common combinations:

```bash
# Conservative stack
docker compose --profile postfix-dovecot --profile roundcube --profile radicale up -d

# Minimal stack (low RAM)
docker compose --profile maddy --profile snappymail --profile radicale up -d
```

> **Podman users:** add `-f docker-compose.podman.yml` (and `-f docker-compose.podman-selinux.yml` on SELinux systems) after the base compose file. The home device can run rootless if you set `net.ipv4.ip_unprivileged_port_start=0`. See the [Podman guide](../deploy/platform-guides/podman.md).

Expected containers for the default stack: `stalwart`, `snappymail`, `rspamd`, `redis`, `profile-server`. There is **no caddy on the home device** — TLS for `mail.example.com` is terminated by the cloud relay's caddy, which proxies into the tunnel.

Host ports on the home device: 25/587/993 (mail), 8080 (webmail), 8081 (Stalwart admin UI), 8090 (profile server), 5232 (Radicale, if used), 11334 (Rspamd admin UI).

## Step 6: Test Email Delivery

1. **Inbound:** from an external account (Gmail, …), send a message to `admin@example.com`.
2. **Check it:** open webmail at `https://mail.example.com` (served by the relay's caddy through the tunnel) or `http://<HOME_DEVICE_IP>:8080` directly. Log in with `admin@example.com` and the `ADMIN_PASSWORD` from `home-device/.env` (or the password the wizard generated).
3. **Outbound:** from webmail, send a message to your personal inbox. Check the spam folder first — a new IP has no reputation yet.
4. **Authentication:** forward the outbound message to `check-auth@verifier.port25.com`. Expect SPF pass, DKIM pass, DMARC pass. Any failure: re-run `dns-setup --validate-only` and check `DKIM_SELECTOR` matches the published record.
5. **Score:** send a test to [mail-tester.com](https://www.mail-tester.com/) and aim for 9/10 or better.

## Step 7: Onboard Your Devices

The profile server (host port **8090**) configures mail clients. Its web UI is at:

- `http://<HOME_DEVICE_IP>:8090/devices` on the LAN, or
- `https://mail.example.com/profile/` … `https://mail.example.com/qr/` via the relay (Basic auth with your admin credentials)

1. **iOS/macOS:** use the Add Device flow, download the generated `.mobileconfig` — one tap installs IMAP (993), SMTP (587), and CalDAV/CardDAV where applicable.
2. **Android:** generate a QR code from the web UI and scan it with K-9 Mail, FairEmail, or the Gmail app.
3. **Thunderbird / Outlook:** point them at `autoconfig.example.com` / `autodiscover.example.com`; the relay's caddy proxies the generated payloads. Manual settings: incoming = the `MAIL_HOSTNAME` address from Step 3, port 993 (SSL/TLS); outgoing = same address, port 587 (STARTTLS); password = the app password from step 1.
4. **App passwords:** issue per-device app passwords from the web UI (`/devices`) instead of sharing your main password.

> **Client TLS note:** the home mail server ships a self-signed TLS certificate, so a client connecting to it directly will ask you to trust it once. For remote clients, you can instead mount a Let's Encrypt (DNS-01) certificate for `MAIL_HOSTNAME` in the mail server config.

## Step 8: Monitor and Maintain

**Dashboard.** The monitoring dashboard is served by the profile server at:

- `http://<HOME_DEVICE_IP>:8090/status` on the LAN
- `https://mail.example.com/status` via the relay (Basic auth)

It shows queue status, service health, certificate expiry, and recent delivery results. Liveness/readiness endpoints: `/health/live`, `/health/ready`.

**Alerts** (`home-device/.env`, then `docker compose up -d`):

```bash
MONITOR_ALERT_EMAIL=you@personal-mail.com
MONITOR_WEBHOOK_URL=https://hooks.slack.com/services/...
MONITOR_HEALTHCHECK_URL=https://hc-ping.com/your-uuid
```

**Logs.** Email addresses are redacted in logs by default (`s***r@example.com`). For active troubleshooting only:

```bash
# cloud-relay/.env
RELAY_DEBUG=true
# home-device: add PROFILE_DEBUG=true to the profile-server service in docker-compose.yml
```

**Restart** the affected containers after changing debug settings, and **turn debug off again** afterwards — it logs full PII.

**IP warmup.** For the first 4–6 weeks, ramp outbound volume gradually: week 1: 10–20 mail/day; week 2: 50–100; week 3: 200–500; weeks 4–6: your normal volume. Send to engaged recipients first, and monitor spam-folder placement and [MXToolbox](https://mxtoolbox.com/blacklists.aspx) blocklist status.

## What's Next

- **Add users and domains** — Stalwart: admin UI at `http://<HOME_DEVICE_IP>:8081` (or via `https://mail.example.com` if you add it to the relay Caddyfile). Maddy: edit `home-device/maddy/setup-users.sh` and `docker compose restart maddy`. Postfix+Dovecot: edit the `vmailbox`/`virtual`/`users` files and `docker compose restart postfix-dovecot`.
- **Migrate existing mail** — see the [migration guide](migration.md) (`darkpipe-setup migrate`).
- **Tune spam filtering** — Rspamd admin UI at `http://<HOME_DEVICE_IP>:11334`; train the Bayesian filter with spam/ham samples.
- **Back up** — the `mail-data` volume on the home device is everything; the cloud relay has nothing worth backing up. Snapshot it daily.
- **Customize** — every variable is documented in [configuration.md](configuration.md); the CLI tools in [reference/cli.md](reference/cli.md).

## Troubleshooting

**Mail not being received**

1. `dns-setup ... --validate-only` — all records pass? (MX, A, PTR especially)
2. Is port 25 reachable on the VPS? `telnet relay.example.com 25` from outside.
3. Cloud relay logs: `docker compose logs relay` — look for delivery to `10.8.0.2`.
4. Is the tunnel up? `ping 10.8.0.2` from the VPS, `sudo wg show` on both sides.

**Mail not being sent**

1. Does `https://mail.example.com` load? (caddy on the relay)
2. Home device mail server logs: `docker compose logs stalwart` (or `maddy` / `postfix-dovecot`).
3. DKIM: forward a sample to `check-auth@verifier.port25.com`; check `DKIM_SELECTOR` and the key file name.
4. Blocklists: check the relay IP on [MXToolbox](https://mxtoolbox.com/blacklists.aspx).

**Webmail not loading**

1. On the **VPS**: `docker compose ps caddy` and `docker compose logs caddy` — TLS errors? ACME failures?
2. From the VPS: `wget -qO- http://10.8.0.2:8080/` — reachable through the tunnel?
3. On the home device: `docker compose ps` — is the webmail container healthy?

**Transport tunnel down**

```bash
sudo wg show                    # both sides
sudo wg-quick down wg0 && sudo wg-quick up wg0
sudo iptables -L | grep 51820   # UDP 51820 must be allowed on the VPS
```

**Services using too much memory**

`docker stats` on the home device. For 2 GB systems: switch Stalwart → Maddy and Roundcube → SnappyMail, and review the per-service limits in `home-device/docker-compose.yml`.

**Need help?** [GitHub Discussions](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/discussions) for questions, [GitHub Issues](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/issues) for bugs.

---

Last Updated: 2026-11-23

License: AGPLv3 — see [LICENSE](../LICENSE)
