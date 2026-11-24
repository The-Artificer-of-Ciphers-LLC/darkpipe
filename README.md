# DarkPipe

**Your email. Your hardware. Your rules.**

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)
[![GitHub Release](https://img.shields.io/github/v/release/The-Artificer-of-Ciphers-LLC/darkpipe)](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/The-Artificer-of-Ciphers-LLC/darkpipe/release.yml)](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/actions)
[![Platform Support](https://img.shields.io/badge/platform-linux%2Famd64%20%7C%20linux%2Farm64-lightgrey)](deploy/platform-guides/)

DarkPipe is a complete self-hosted email sovereignty stack. A minimal cloud relay handles internet-facing SMTP (receiving and sending), then securely transports messages — without storing them — to a mail server running on hardware you control at home. You choose your mail server, webmail client, and groupware components through a modular build system. The system includes automated DNS authentication, encrypted offline queuing, device onboarding via QR codes and profiles, monitoring with a web dashboard, and migration tools for 7 popular providers.

Your email lives on your hardware, encrypted in transit, never stored on someone else's server — and it still works like normal email from the outside.

## Architecture

```
Internet ──> Cloud Relay (VPS) ──[WireGuard/mTLS]──> Home Device (your hardware)
                  │                                           │
             Postfix MTA + relay daemon             Mail Server
             (age-encrypted offline queue —        (Stalwart/Maddy/Postfix+Dovecot)
              never stores mail)                    Webmail (SnappyMail/Roundcube)
             Caddy (webmail, autodiscovery,         Calendar/Contacts (Radicale/Stalwart)
              dashboard — proxied into tunnel)      Rspamd (spam filter, DKIM signing)
             Certbot (Let's Encrypt)                Profile server (onboarding, dashboard)
```

**Data flow:**

- **Inbound**: Internet → Cloud Relay → Encrypted Transport → Home Device → Mailbox
- **Outbound**: Mail Client → Home Device → Encrypted Transport → Cloud Relay → Internet
- **Offline**: Cloud Relay queues mail encrypted, drains when Home Device reconnects

**Key principle:** The cloud relay never stores mail. It's a pass-through gateway that ensures deliverability while your storage remains on hardware you physically control.

## Features

**Mail Server Options**
- Stalwart (modern all-in-one with IMAP4rev2, JMAP, built-in CalDAV/CardDAV)
- Maddy (minimal Go-based single binary)
- Postfix + Dovecot (traditional battle-tested MTA + IMAP)

**Webmail Options**
- Roundcube (traditional, feature-rich, PHP-based)
- SnappyMail (modern, fast, lightweight)

**Calendar and Contacts**
- Radicale CalDAV/CardDAV server (for Maddy and Postfix+Dovecot)
- Stalwart built-in CalDAV/CardDAV (when using Stalwart mail server)
- Shared family calendars and contacts

**Transport Security**
- WireGuard full tunnel (simple setup, kernel-level encryption)
- mTLS with internal PKI (minimal footprint, certificate-based auth)
- Automatic certificate rotation (configurable: 30/60/90 days)
- TLS 1.2+ enforced on all IMAP provider connections

**Container Security**
- All containers run with `no-new-privileges` and `cap_drop: ALL`
- Read-only root filesystems with explicit tmpfs mounts
- Docker HEALTHCHECK on all custom images
- PII-redacted logging by default (email addresses masked in logs)

**DNS Automation**
- SPF, DKIM, DMARC record generation and validation
- DNS API integration (Cloudflare, Route53)
- Manual DNS guide fallback for any provider
- Automated DKIM key rotation

**Offline Queue**
- Encrypted queue with age encryption (filippo.io/age)
- S3-compatible overflow storage (Storj, AWS S3, MinIO)
- Configurable queue-or-bounce behavior
- Automatic drain when home device reconnects

**Device Onboarding**
- Apple .mobileconfig profiles (iOS/macOS one-tap setup)
- QR codes for mobile configuration
- Thunderbird/Outlook autodiscovery
- App-generated passwords for mail clients

**Mail Migration**
- Migrate from 7 providers: Gmail, Outlook/Microsoft 365, iCloud, MailCow, Mailu, docker-mailserver, generic IMAP
- OAuth2 device flow for Gmail and Outlook (no browser redirect needed)
- Dry-run mode (safe migration testing before applying)
- Progress tracking and folder mapping

**Multi-Architecture**
- Pre-built Docker images for amd64 and arm64
- GitHub Actions custom build pipeline for component selection
- Runs on Raspberry Pi 4+, x64/arm64 Docker hosts, NAS platforms

**Monitoring**
- Web-based monitoring dashboard
- Mail queue health and delivery status tracking
- Certificate expiry alerts and automatic renewal
- Alert notifications via webhook or email

**Spam Filtering**
- Rspamd spam filter with greylisting
- Redis backend for statistics and temporary storage
- Configurable thresholds and custom rules

**Multi-User, Multi-Domain**
- Support for multiple users and domains
- Email aliases and catch-all addresses
- User management via mail server admin interfaces

## Quick Start

Get from zero to running email in three steps:

### 1. Provision a VPS with Port 25 Access

Port 25 (SMTP) is required for sending and receiving email. Many cloud providers block it.

**Recommended providers:** Hetzner, Vultr, OVH, Linode

**See full provider compatibility matrix:** [docs/vps-providers.md](docs/vps-providers.md)

**Minimum VPS specs:** 1 vCPU, 1GB RAM, 20GB SSD ($3-6/month)

### 2. Download and Run Setup Wizard

```bash
# Download setup tool (replace <OS> and <ARCH> with your platform)
curl -LO https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/darkpipe-setup-<OS>-<ARCH>

# Make executable and run the interactive wizard
chmod +x darkpipe-setup-<OS>-<ARCH>
./darkpipe-setup-<OS>-<ARCH> setup
```

The wizard collects your domain, relay hostname, admin email, mail server, webmail, and transport choices (with live validation), then generates in the current directory: a single `docker-compose.yml` with your selected services, `.darkpipe.yml`, `secrets/admin_password.txt` (a generated 24-character admin password — note it down), and `secrets/dkim_private_key.pem` (a placeholder for the key generated in the DNS step).

### 3. Configure DNS

```bash
# Download DNS setup tool
curl -LO https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/dns-setup-<OS>-<ARCH>

# Make executable
chmod +x dns-setup-<OS>-<ARCH>

# Run DNS setup (dry-run by default)
./dns-setup-<OS>-<ARCH> --domain yourdomain.com --relay-hostname relay.yourdomain.com --relay-ip YOUR_VPS_IP

# Review changes, then apply
./dns-setup-<OS>-<ARCH> --domain yourdomain.com --relay-hostname relay.yourdomain.com --relay-ip YOUR_VPS_IP --apply
```

**Next steps:** Deploy containers, test email sending/receiving, onboard devices

**Full setup guide:** [docs/quickstart.md](docs/quickstart.md)

## Stack Configurations

DarkPipe provides two pre-built stack configurations:

| Stack         | Mail Server       | Webmail      | Calendar/Contacts        | Use Case                      |
|---------------|-------------------|--------------|--------------------------|-------------------------------|
| **Default**   | Stalwart 0.15.4   | SnappyMail   | Stalwart built-in        | Most users, modern features   |
| **Conservative** | Postfix + Dovecot | Roundcube    | Radicale                 | Traditional, battle-tested    |

**Custom builds:** Fork the repository and trigger the "Build Custom Stack" GitHub Actions workflow to select your own component combination.

## Supported Platforms

DarkPipe runs on any Docker-capable system. Platform-specific guides available:

- [Raspberry Pi 4+](deploy/platform-guides/raspberry-pi.md) (arm64, 4GB+ RAM recommended)
- [TrueNAS Scale 24.10+](deploy/platform-guides/truenas-scale.md)
- [Unraid](deploy/platform-guides/unraid.md)
- [Proxmox LXC](deploy/platform-guides/proxmox-lxc.md)
- [Synology NAS](deploy/platform-guides/synology-nas.md) (Container Manager)
- [Mac Silicon](deploy/platform-guides/mac-silicon.md) (Apple M-series)
- Any Docker-capable x64 or arm64 Linux host

**Minimum home device requirements:** 2GB RAM (4GB recommended), 20GB storage, Docker 24+ (or Podman 5.3+)

## Documentation

- **Learn** — [First Email tutorial](docs/tutorials/first-email.md): one safe path from zero to your first email, step by step
- **How to** — [Quick Start](docs/quickstart.md): the full deployment, VPS to first email · [Migration guide](docs/migration.md): import mail from 7 providers · [Platform guides](deploy/platform-guides/): Raspberry Pi, NAS, Proxmox, macOS
- **Reference** — [Configuration](docs/configuration.md): every variable and profile · [CLI reference](docs/reference/cli.md): `dns-setup` and `darkpipe-setup` · [VPS providers](docs/vps-providers.md): port 25 compatibility
- **Understand** — [Architecture](docs/architecture.md): the split design and why · [Security model](docs/security.md): threat model and encryption
- [FAQ](docs/faq.md): common questions and troubleshooting · [Contributing](docs/contributing.md): development setup and conventions

## Community and Support

**Questions and discussions:** [GitHub Discussions](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/discussions)

**Bug reports:** [GitHub Issues](https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/issues)

**Contributing:** See [docs/contributing.md](docs/contributing.md)

We welcome contributions of all kinds: bug reports, feature suggestions, documentation improvements, code contributions, platform testing, and helping other users in discussions.

## Sustainability

DarkPipe is AGPLv3 licensed and community-driven. Development is funded by donations:

- [GitHub Sponsors](https://github.com/sponsors/trek-e)
- [Open Collective](https://opencollective.com/darkpipe)
- [Liberapay](https://liberapay.com/darkpipe)
- [Ko-fi](https://ko-fi.com/darkpipe)

Your support helps keep this project independent and focused on user sovereignty.

## License

Copyright (C) 2026 The Artificer of Ciphers, LLC

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for full license text.

Third-party dependencies: [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md)

---

**Built because your inbox shouldn't live on someone else's computer.**
