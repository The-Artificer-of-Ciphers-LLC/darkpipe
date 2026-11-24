# First Email Tutorial

In this tutorial, you will deploy a complete DarkPipe email system — a cloud relay on a Hetzner VPS, an encrypted WireGuard tunnel, and a mail server on a Raspberry Pi 4 at home — and send and receive a real email through it.

This is the single recommended starting path. Every choice below is a default that "just works"; there are no alternatives to choose. When you're ready to customize, the [Quick Start guide](../quickstart.md) covers all options (mail servers, transports, DNS providers, platforms).

## What you will need

- **A domain name** you control, with the ability to add DNS records (a Cloudflare account speeds this up, but manual entry works with any provider).
- **A Hetzner VPS** — the CX11 is enough (1 vCPU, 2 GB RAM, ~€3.79/month). Hetzner allows port 25, which your relay needs.
- **A Raspberry Pi 4 (4 GB)** with a USB SSD and Raspberry Pi OS 64-bit, on your home network, with Docker 24+ installed (`curl -fsSL https://get.docker.com | sh`).
- **About an hour** of active setup.

**At the end, you will have:**

- Mail for your domain arriving at a relay on Hetzner and forwarded over an encrypted tunnel to your Pi
- A Stalwart mail server storing everything on your Pi — the VPS stores nothing
- Webmail at `https://mail.example.com` (replace `example.com` with your domain throughout)
- SPF, DKIM, and DMARC published for your domain
- A working email account on your phone

If a step fails, the [Quick Start troubleshooting section](../quickstart.md#troubleshooting) covers every failure mode below.

## 1. Prepare the VPS

SSH into your fresh Hetzner VPS and run:

```bash
hostnamectl set-hostname relay.example.com
```

Then, in the Hetzner robot web console, set the **reverse DNS (PTR)** for the VPS IP to `relay.example.com`. Mail servers check this; without it, other providers treat your mail as spam.

Install Docker and verify port 25 is open:

```bash
apt update && apt upgrade -y
curl -fsSL https://get.docker.com | sh
apt install -y telnet
telnet gmail-smtp-in.l.google.com 25
```

**You will see** a `220 ... ESMTP` greeting (type `QUIT` to exit). If you get a timeout instead, your VPS can't send — see [vps-providers.md](../vps-providers.md).

Clone the repository:

```bash
git clone https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe.git
cd darkpipe
```

## 2. Build the tunnel

WireGuard connects the two machines over an encrypted tunnel (`10.8.0.0/24`). First, on **both** the VPS and the Pi, generate a key pair:

```bash
sudo apt install -y wireguard-tools
wg genkey        # writes the private key; press Enter to print the public key
```

Write down all four values: cloud private key, cloud public key, home private key, home public key.

**On the VPS**, create `/etc/wireguard/wg0.conf`:

```ini
[Interface]
Address = 10.8.0.1/24
PrivateKey = <cloud-private-key>
ListenPort = 51820

[Peer]
PublicKey = <home-public-key>
AllowedIPs = 10.8.0.2/32
```

**On the Pi**, create `/etc/wireguard/wg0.conf`:

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

Now install and start the tunnel on each machine using the repo's helper script (it takes the path to the config you just wrote):

```bash
# On the VPS
sudo ./deploy/wireguard/cloud-setup.sh /etc/wireguard/wg0.conf

# On the Pi
sudo ./deploy/wireguard/home-setup.sh /etc/wireguard/wg0.conf
```

**You will see** "WireGuard interface status" with the peer listed on both sides. Verify from the VPS:

```bash
ping -c 3 10.8.0.2
```

Packets must get through in both directions before continuing.

## 3. Configure both halves

Both deploy targets ship a documented `.env.example`. Copy and edit it on each machine:

```bash
# On the VPS
cd darkpipe/cloud-relay && cp .env.example .env

# On the Pi
cd darkpipe/home-device && cp .env.example .env
```

**In `cloud-relay/.env`**, set:

```bash
RELAY_HOSTNAME=relay.example.com
RELAY_DOMAIN=example.com
RELAY_HOME_ADDR=10.8.0.2:25
HOME_DEVICE_IP=10.8.0.2
```

**In `home-device/.env`**, set:

```bash
MAIL_DOMAIN=example.com
MAIL_HOSTNAME=<your Pi's LAN address>   # what your mail clients will connect to (see below)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=<a-strong-password>      # note this — you'll use it for webmail
DKIM_SELECTOR=darkpipe-2026q4          # will be confirmed in step 4
```

**About `MAIL_HOSTNAME`:** mail clients connect to this address directly — the relay never proxies mail ports. In this tutorial your clients are on the home network, so use the Pi's LAN address (for example `192.168.1.50`). If you later want access from outside the home network, change it to a public name for the Pi with ports 993/587 port-forwarded. The webmail, dashboard, and device onboarding use `mail.example.com`, which points at the relay — a different address, on purpose.

Also add this A record to your domain's DNS now (you did the PTR in step 1; this is the forward record):

- `mail.example.com` → **your VPS IP** (the same IP as `relay.example.com`)

## 4. Publish DNS records

On the machine where you'll keep the DKIM key (the Pi is fine), download the DNS tool:

```bash
curl -LO https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/dns-setup-linux-arm64
chmod +x dns-setup-linux-arm64
```

Preview what it will do (dry-run — nothing is changed):

```bash
./dns-setup-linux-arm64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip <VPS-public-IP>
```

**You will see** every record listed: MX, A, SPF, DKIM, DMARC, mail SRV, and autoconfig/autodiscover CNAMEs. Note the DKIM selector it prints (for example `darkpipe-2026q4`) — put that value in `home-device/.env` as `DKIM_SELECTOR`.

Apply them. The tool detects your DNS provider from the domain's NS records — Cloudflare needs a token, Route53 standard AWS credentials, any other provider manual entry from the printed list:

```bash
export CLOUDFLARE_API_TOKEN=<your-token>     # only for Cloudflare
./dns-setup-linux-arm64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip <VPS-public-IP> \
  --apply
```

Install the generated DKIM key on the Pi, where Rspamd signs your outgoing mail:

```bash
# From the machine where dns-setup ran, copy the key over:
scp /etc/darkpipe/dkim/darkpipe-2026q4.private.pem you@<pi-lan-address>:~/

# On the Pi, from the repo clone's root directory — the key must be
# named <domain>.<selector>.key in Rspamd's key directory:
cd ~/darkpipe   # wherever you cloned the repository
mkdir -p home-device/spam-filter/rspamd/dkim-keys
mv ~/darkpipe-2026q4.private.pem home-device/spam-filter/rspamd/dkim-keys/example.com.darkpipe-2026q4.key
```

DNS takes 5–60 minutes to propagate. When it's done, validate:

```bash
./dns-setup-linux-arm64 \
  --domain example.com \
  --relay-hostname relay.example.com \
  --relay-ip <VPS-public-IP> \
  --validate-only
```

**You will see** every check pass. Only move on once they do.

## 5. Deploy the services

**On the VPS:**

```bash
cd darkpipe/cloud-relay
# The relay runs without TLS until a certificate appears — start it first
docker compose up -d relay
# The certbot sidecar (separate compose file) obtains the Let's Encrypt cert
docker compose -f docker-compose.yml -f certbot/docker-compose.certbot.yml up -d certbot
docker compose logs certbot | tail   # expect "Certificate obtained successfully"
# Then caddy, which serves webmail through the tunnel
docker compose up -d caddy
docker compose ps
```

**You will see** `darkpipe-relay` and `caddy` both Up and healthy.

**On the Pi:**

```bash
cd darkpipe/home-device
docker compose --profile stalwart --profile snappymail up -d
docker compose ps
```

**You will see** five healthy containers: `stalwart`, `snappymail`, `rspamd`, `redis`, and `profile-server`. There is no Caddy on the Pi — by design; the relay's Caddy is your front door.

## 6. Send and receive your first email

1. From an outside account (your Gmail), send a message to `admin@example.com`.
2. **You will see** it appear in webmail. Open `https://mail.example.com` in a browser and log in with `admin@example.com` and the `ADMIN_PASSWORD` you chose in step 3.
3. Reply to that message from webmail. Check the Gmail inbox — it should arrive (check spam first; new IPs have no reputation yet).
4. Forward the Gmail-→-you message to `check-auth@verifier.port25.com`. **You will see** "SPF: Pass, DKIM: Pass, DMARC: Pass" — your domain is now authenticated on the internet.

New sending IPs need 4–6 weeks of gradual warmup before receiving providers fully trust them. For the first weeks, send only to people who reply — that's what builds reputation.

## 7. Put mail on your phone

From a device on your home Wi-Fi, open the profile server's web UI:

```
http://<Pi-LAN-address>:8090/devices
```

- **iPhone/Mac:** use *Add Device*, then download the generated profile — one tap installs IMAP, SMTP, and calendar/contacts.
- **Android:** generate a QR code in the UI and scan it with your mail app.

The profiles point the account at the `MAIL_HOSTNAME` address from step 3, so on your home network everything connects directly to the Pi. The mail server's certificate is self-signed, so your phone will ask you to trust it once.

## You're done

You now run real email on your own hardware, authenticated end to end. Next, pick up where this tutorial left off:

- The [Quick Start](../quickstart.md) — every option this tutorial skipped (other mail servers, mTLS, manual DNS, S3 overflow, …)
- [Migrate existing mail](../migration.md) — import your old inbox with `darkpipe-setup migrate`
- [Configuration reference](../configuration.md) — every knob on both halves

---

**Last Updated**: 2026-11-23

**License**: AGPLv3 — see [LICENSE](../../LICENSE)
