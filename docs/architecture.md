# DarkPipe System Architecture

This is an explanation of how DarkPipe is put together and *why*: the split between cloud relay and home device, what each component does, and how mail actually flows. For step-by-step instructions, see the [quick start](quickstart.md); for settings, see the [configuration reference](configuration.md).

## Why a Split Architecture

Self-hosted email has a fundamental tension:

- **Deliverability** wants your mail server on a VPS: datacenter IPs, open port 25, reputation you can manage.
- **Sovereignty** wants your mail *stored* on hardware you physically control.

Projects like Mail-in-a-Box and Mailu solve deliverability by giving you everything on a VPS — which means your mail is stored on someone else's hardware. DarkPipe resolves the tension by splitting the two roles:

- The **cloud relay** (VPS) handles everything internet-facing: receiving on port 25, sending outbound, TLS, DNS authentication. It is a *pass-through* — mail is not stored there.
- The **home device** (your hardware) stores mail and serves it: the mail server, webmail, groupware, spam filtering, onboarding, monitoring.

The two are joined by an encrypted transport (WireGuard or mTLS). The security implications of this split — and its limits — are analyzed in the [security model](security.md).

## Component Overview

```
Internet
   │  SMTP :25 (TLS)          ▲  outbound SMTP :25
   ▼                          │
┌──────────────────────────────────────────────────────┐
│  CLOUD RELAY (VPS)                                   │
│                                                      │
│  darkpipe-relay        caddy                         │
│  ┌───────────────┐     ┌──────────────────────────┐  │
│  │ Postfix MTA   │────▶│ Go relay daemon          │  │  webmail, autoconfig,
│  │ (:25, TLS)    │     │ (receives :10025)        │  │ autodiscover, dashboard
│  └───────────────┘     │ age-encrypted offline    │  │  (TLS terminated here,
│        ▲               │ queue + S3 overflow      │  │  proxied into tunnel)
│        │               └────────────┬─────────────┘  │
│  certbot sidecar (LE for relay      │                │
│  hostname) — separate compose file   │                │
└──────────────────────────────────────┼────────────────┘
                                       │  WireGuard 10.8.0.0/24  (or mTLS)
                                       ▼
┌──────────────────────────────────────────────────────┐
│  HOME DEVICE (your hardware)                         │
│                                                      │
│  Mail server (one):  stalwart :25/:587/:993/:8080   │
│                      maddy  :25/:587/:993           │
│                      postfix-dovecot :25/:587/:993  │
│  Webmail (one):      snappymail :8080  / roundcube  │
│  CalDAV/CardDAV:     radicale :5232  (or Stalwart)  │
│  rspamd :11334  redis  profile-server :8090         │
└──────────────────────────────────────────────────────┘
```

Key asymmetry: **there is no reverse proxy on the home device.** TLS for `mail.example.com` is terminated by the relay's caddy, which reverse-proxies into the tunnel to the home services. The home device itself only needs its mail and onboarding ports reachable *inside* the tunnel (and optionally on your LAN).

## Cloud Relay

| Container | Role |
|-----------|------|
| `darkpipe-relay` | Postfix MTA (internet-facing :25) + the Go relay daemon. Postfix hands accepted mail to the daemon on `127.0.0.1:10025`; the daemon forwards it to the home device over the transport |
| `caddy` | Terminates HTTPS for `mail.example.com`, `autoconfig.example.com`, `autodiscover.example.com`; proxies webmail, profile/QR downloads, and the dashboard to the home device through the tunnel |
| `darkpipe-certbot` (sidecar, separate compose file) | Obtains and renews the Let's Encrypt certificate for the relay hostname, which Postfix uses for SMTP TLS |

The relay is deliberately ephemeral:

- **Ephemeral verification** — a background check (every 60 s by default) confirms no mail persists on disk after forwarding.
- **Offline queue** — when the home device is unreachable, the daemon encrypts the message with age and queues it in RAM (200 MiB / 10,000 message caps), spilling to S3-compatible storage when full, and purges after 7 days. See [configuration](configuration.md#s3-overflow).
- **PII-safe logs** — addresses are redacted by default; `RELAY_DEBUG` opts into full PII.
- **Memory budget** — 256 MB cap, sized for a $3–6/month VPS.

The relay's own images are built from [`cloud-relay/`](../cloud-relay/) (multi-stage: Go daemon + Postfix + WireGuard tools on Alpine) or pulled from the release images.

## Transport Layer

Two options, selected by `RELAY_TRANSPORT`:

**WireGuard (recommended)** — a full IP tunnel (`10.8.0.0/24`) between the two machines. Kernel-level, simple to operate, and it needs only one outbound rule from the home side (UDP 51820 to the VPS) — no inbound exposure on the home device. Setup scripts: [`deploy/wireguard/`](../deploy/wireguard/); config templates: [`transport/wireguard/config`](../transport/wireguard/config/templates.go).

**mTLS** — the relay daemon dials the home device over mutual TLS, with certificates from an internal step-ca PKI ([`deploy/pki/step-ca-setup.sh`](../deploy/pki/step-ca-setup.sh), [`transport/mtls/`](../transport/mtls/), renewal via [`transport/pki/renewal`](../transport/pki/renewal/)). Smaller footprint, no kernel module — but the checked-in home device compose files assume the WireGuard tunnel, so mTLS is the advanced/manual path.

Both provide equivalent confidentiality for the cloud→home hop.

## Home Device

| Container | Role | Host port |
|-----------|------|-----------|
| `stalwart` / `maddy` / `postfix-dovecot` | Mail server (pick one via profile) — stores mail, serves IMAP/submission; Stalwart also serves CalDAV/CardDAV and a management UI | 25, 587, 993 (+8081 Stalwart admin) |
| `snappymail` / `roundcube` | Webmail (pick one) | 8080 |
| `radicale` | CalDAV/CardDAV (Maddy / Postfix+Dovecot stacks only) | 5232 |
| `rspamd` | Spam filtering: greylisting, Bayesian, RBLs — and **DKIM signing of outbound mail** | 11334 |
| `redis` | Rspamd state (greylist, stats) | — |
| `profile-server` | Device onboarding (mobileconfig, QR, autoconfig/autodiscover payloads, app passwords) and the monitoring dashboard/aggregator | 8090 |

**Where DKIM signing happens.** A common misconception: signing is done by Rspamd *on the home device*, not by the relay. Outbound mail is signed before it leaves the home network (see [`dkim_signing.conf`](../home-device/spam-filter/rspamd/local.d/dkim_signing.conf)); the relay then delivers it. The private key lives on the home device, named `<domain>.<selector>.key` under the rspamd volume; the selector must match the DNS record published by `dns-setup` (`DKIM_SELECTOR`).

The mail server data lives in the `mail-data` volume — this is the only state that matters to back up.

## Data Flow

### Inbound

1. An external MTA connects to the relay on :25 (opportunistic TLS by default; strict mode optional).
2. Postfix accepts the message and queues it to the relay daemon on :10025.
3. The daemon forwards it over the transport to the home mail server's :25.
4. The home mail server runs it through Rspamd (spam scan) and delivers to the mailbox.
5. The user fetches via IMAP :993 or webmail.

If the home device is unreachable at step 3, the message is age-encrypted and queued (RAM → S3), delivered when the device returns, then deleted from the queue.

### Outbound

1. The user submits via IMAP :587 (STARTTLS) or webmail.
2. Rspamd scans and **signs** the message with the domain DKIM key.
3. The mail server relays it over the transport to the relay's Postfix.
4. Postfix delivers to the recipient's MTA on :25.

### Webmail and onboarding

Browsers and mail clients talk to `mail.example.com` (and `autoconfig`/`autodiscover`), which resolve to the **relay**. Its caddy terminates TLS and proxies into the tunnel: webmail to :8080, profile/QR/autodiscover payloads to :8090, dashboard to :8090/status (Basic auth).

## Build System

- **Root Go module** (`go 1.26`): the relay daemon, DNS tooling, transport, and monitoring packages.
- **`deploy/setup`** (own module): the `darkpipe-setup` wizard and migration tool (isolated UI/IMAP dependencies).
- **`home-device/profiles`** (own module): the profile server.
- **GitHub Actions**: `release.yml` (tag → release assets + GHCR images), `build-prebuilt.yml` (default/conservative stacks), `build-custom.yml` (workflow_dispatch stack selection). See [configuration → Releases](configuration.md#releases-and-custom-builds).

## Repository Layout

```
darkpipe/
├── cloud-relay/            # relay image, compose, Caddyfile, certbot sidecar, Postfix config
│   └── relay/              # Go relay daemon (smtp, forward, queue, notify, tls, logutil)
├── home-device/            # compose + profiles, per-component config
│   ├── stalwart/ maddy/ postfix-dovecot/ webmail/ caldav-carddav/ spam-filter/
│   └── profiles/           # profile server (own module)
├── transport/              # wireguard (keygen, config, monitor), mtls, pki
├── dns/                    # dns-setup: records, providers, validator, dkim
├── monitoring/             # health, status, queue, delivery, cert (root module)
├── deploy/
│   ├── setup/              # darkpipe-setup wizard (own module)
│   ├── wireguard/ pki/     # host-level setup scripts
│   ├── platform-guides/    # Raspberry Pi, NAS, Podman, …
│   └── test/               # tunnel/outage test scripts
├── scripts/                # verification scripts (security, redaction, podman compat, runtime check)
└── docs/                   # this documentation
```

## Technology Stack

- **Go 1.26** — relay daemon, dns-setup, darkpipe-setup, profile server, monitoring
- **Postfix 3.9** (Alpine) — cloud relay MTA
- **Stalwart 0.15.4** / **Maddy 0.8.2** / **Postfix+Dovecot** — mail server options
- **SnappyMail 2.38.2** / **Roundcube 1.6.13** — webmail options
- **Radicale 3.6.0** — CalDAV/CardDAV option
- **Rspamd** (+ Redis) — spam filtering and DKIM signing
- **Caddy 2** — TLS termination on the relay
- **WireGuard** / **step-ca** — transport options
- **age** (filippo.io/age) — offline queue encryption
- **Let's Encrypt** via certbot — public TLS

---

Last Updated: 2026-11-23

License: AGPLv3 — see [LICENSE](../LICENSE)
