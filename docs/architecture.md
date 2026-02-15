# DarkPipe System Architecture

This document provides a detailed technical overview of DarkPipe's architecture, components, and data flows.

## Architecture Overview

DarkPipe uses a split architecture that separates the internet-facing relay from persistent mail storage:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              INTERNET                                    │
└──────────────────────────┬──────────────────────────────────────────────┘
                           │
                           │ SMTP (port 25)
                           │ TLS (Let's Encrypt)
                           │
┌──────────────────────────▼──────────────────────────────────────────────┐
│                       CLOUD RELAY (VPS)                                  │
│                                                                           │
│  ┌─────────────┐    ┌──────────────┐    ┌────────────────┐             │
│  │   Postfix   │───▶│  Go Relay    │───▶│   Certbot      │             │
│  │     MTA     │    │   Service    │    │ (Let's Encrypt)│             │
│  │  (port 25)  │    │ (port 10025) │    │                │             │
│  └─────────────┘    └──────┬───────┘    └────────────────┘             │
│                             │                                            │
│  ┌─────────────────────────┴────────────────┐                           │
│  │        Encrypted Transport Layer         │                           │
│  │    WireGuard (10.8.0.0/24) OR mTLS       │                           │
│  └─────────────────────┬────────────────────┘                           │
└────────────────────────┼─────────────────────────────────────────────────┘
                         │ Encrypted tunnel
                         │ (WireGuard UDP or mTLS TCP)
                         │
┌────────────────────────▼─────────────────────────────────────────────────┐
│                  HOME DEVICE (User Hardware)                              │
│                                                                           │
│  ┌────────────────────────────────────────────────────────┐              │
│  │              Mail Server (user-selectable)             │              │
│  │  Stalwart OR Maddy OR Postfix+Dovecot                 │              │
│  │  - SMTP: port 25 (inbound), 587 (submission)          │              │
│  │  - IMAP: port 993 (IMAPS)                             │              │
│  │  - Management API (Stalwart: 8080, Maddy: N/A)        │              │
│  └────────────────────────────────────────────────────────┘              │
│                                                                           │
│  ┌──────────────┐  ┌───────────────┐  ┌─────────────────┐               │
│  │   Webmail    │  │ CalDAV/CardDAV│  │   Rspamd        │               │
│  │  Roundcube/  │  │  Radicale or  │  │ Spam Filter     │               │
│  │  SnappyMail  │  │ Stalwart built│  │ (port 11333)    │               │
│  │ (port 8080)  │  │  (port 5232)  │  └─────────────────┘               │
│  └──────────────┘  └───────────────┘                                     │
│                                        ┌─────────────────┐               │
│  ┌──────────────┐  ┌───────────────┐  │   Redis         │               │
│  │    Caddy     │  │  Profile      │  │  (Rspamd data)  │               │
│  │ Reverse Proxy│  │   Server      │  │                 │               │
│  │(443/80)      │  │ (port 8090)   │  └─────────────────┘               │
│  └──────────────┘  └───────────────┘                                     │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

## Components

### Cloud Relay (VPS)

The cloud relay runs on a VPS with port 25 access and provides internet-facing SMTP services without storing mail.

**Postfix MTA**
- Receives inbound SMTP connections from the internet (port 25)
- Sends outbound SMTP to external mail servers (port 25)
- Configured as a relay-only server (no local delivery)
- TLS required on all connections with optional strict mode
- Forwards mail to internal Go relay service (port 10025)

**Go Relay Service**
- Receives mail from Postfix (port 10025)
- Encrypts and forwards mail to home device via WireGuard or mTLS
- Handles offline queueing when home device is unreachable
- Manages encrypted queue with age encryption
- S3-compatible overflow storage for large queues
- Ephemeral storage verification (checks every 60 seconds that no mail remains on disk)
- TLS monitoring and notifications for delivery failures

**Certbot**
- Automates Let's Encrypt certificate acquisition and renewal
- Provides TLS certificates for Postfix (SMTP TLS)
- Monitors certificate expiry and triggers alerts

**Monitoring Agent**
- Health checks for all cloud relay services
- Mail queue monitoring and alerts
- Certificate expiry tracking
- Reports status to home device monitoring dashboard

**Resource Limits**
- Maximum memory: 256MB (per UX-02 requirement)
- Designed for minimal VPS ($3-6/month)

### Transport Layer

The transport layer securely connects the cloud relay to the home device using one of two mechanisms:

**WireGuard (Recommended for Most Users)**
- Full tunnel VPN between cloud and home
- Kernel-level encryption (ChaCha20-Poly1305)
- NAT traversal built-in
- Simple setup with wg-quick
- Network: 10.8.0.0/24 (cloud: 10.8.0.1, home: 10.8.0.2)
- Configuration: deploy/wireguard/cloud-setup.sh and home-setup.sh

**mTLS (Minimal Footprint Alternative)**
- Mutual TLS authentication between cloud and home
- Certificate-based identity verification
- Internal PKI managed by step-ca (Smallstep)
- Automatic certificate rotation (configurable: 30/60/90 days)
- No kernel dependencies (userspace only)
- Configuration: transport/mtls/ and transport/pki/

Both options provide equivalent security guarantees. WireGuard is simpler to set up, mTLS has a smaller footprint.

### Home Device

The home device runs all persistent mail storage and user-facing services on hardware the user controls.

**Mail Server Options**

*Stalwart (Default for Most Users)*
- Modern all-in-one mail server written in Rust
- IMAP4rev2, JMAP, SMTP submission
- Built-in CalDAV/CardDAV (no separate groupware needed)
- Web-based management UI (port 8080)
- Multi-user, multi-domain support
- Version: 0.15.4 (pre-v1.0, monitor for updates)
- Configuration: home-device/stalwart/config.toml

*Maddy (Minimal Footprint)*
- Single Go binary, minimal dependencies
- IMAP, SMTP submission
- Suitable for low-resource environments (2GB RAM systems)
- No built-in CalDAV/CardDAV (use Radicale)
- Version: 0.8.2
- Configuration: home-device/maddy/maddy.conf

*Postfix + Dovecot (Battle-Tested Traditional)*
- Postfix MTA + Dovecot IMAP
- Decades of production use
- Extensive documentation and community support
- No built-in CalDAV/CardDAV (use Radicale)
- Configuration: home-device/postfix-dovecot/

All mail servers support multi-user, multi-domain, aliases, catch-all, and SMTP submission authentication.

**Webmail Options**

*Roundcube (Traditional, Feature-Rich)*
- PHP-based webmail
- Rich feature set, extensive plugin ecosystem
- Mobile-responsive interface
- SQLite backend (no database server needed)
- Version: 1.6.13
- Configuration: home-device/webmail/roundcube/config.inc.php

*SnappyMail (Modern, Fast)*
- Modern JavaScript webmail
- Faster and lighter than Roundcube
- Mobile-first design
- Lower memory footprint (128MB limit vs 256MB)
- Version: 2.38.2
- Configuration: home-device/webmail/snappymail/domains/default.json

**CalDAV/CardDAV Options**

*Radicale (Standalone Server)*
- Lightweight Python-based CalDAV/CardDAV server
- Used with Maddy or Postfix+Dovecot configurations
- Shared family calendars and contacts
- File-based storage (no database)
- Version: 3.6.0
- Configuration: home-device/caldav-carddav/radicale/

*Stalwart Built-in*
- When using Stalwart mail server, CalDAV/CardDAV is built-in
- No separate service needed
- Unified user management

**Rspamd Spam Filter**
- Modern spam filtering engine
- Greylisting support (temporary rejection of unknown senders)
- Bayesian filtering with Redis backend
- Configurable thresholds and custom rules
- Web UI for statistics (port 11334)
- Configuration: home-device/spam-filter/rspamd/

**Redis**
- Backend for Rspamd greylisting and statistics
- Minimal memory footprint (64MB limit)
- Persistent storage for greylist data

**Caddy Reverse Proxy**
- Automatic HTTPS with Let's Encrypt
- Reverse proxy for webmail and CalDAV/CardDAV
- HTTP/3 (QUIC) support
- Configuration: home-device/caddy/Caddyfile

**Profile Server**
- Go service for device onboarding
- Generates Apple .mobileconfig profiles
- QR codes for mobile configuration
- Thunderbird/Outlook autodiscover.xml
- App password management
- Port 8090
- Built from: home-device/profiles/

### DNS Management

**dns-setup CLI**
- Go command-line tool for DNS record management
- Generates SPF, DKIM, DMARC records
- Validates existing DNS configuration
- API integration for Cloudflare and Route53
- Dry-run mode by default (--apply to commit changes)
- Built from: dns/cmd/dns-setup/

**Supported DNS Providers**
- Cloudflare (API token required)
- AWS Route53 (AWS credentials required)
- Manual fallback (prints records for manual entry)

**DKIM Key Management**
- Automatic DKIM key generation
- Key rotation support
- 2048-bit RSA keys

### Monitoring

**Health Checks**
- All services include Docker healthcheck definitions
- 30-second intervals, 3 retries, 10-second timeout
- Specific checks: nc for SMTP/IMAP, curl/wget for HTTP services

**Monitoring Dashboard**
- Web-based dashboard showing:
  - Mail queue status (cloud and home)
  - Service health (all containers)
  - Certificate expiry dates
  - Recent delivery status
- Configuration via profile-server environment variables

**Alert Notifications**
- Email alerts (configurable recipient)
- Webhook notifications (for integration with monitoring platforms)
- Healthcheck.io integration (MONITOR_HEALTHCHECK_URL)

### Build System

**GitHub Actions Workflows**

*.github/workflows/release.yml*
- Triggered on version tags (v*)
- Builds multi-arch Docker images (amd64, arm64)
- Publishes to GitHub Container Registry
- Creates GitHub release with binaries and release notes

*.github/workflows/build-prebuilt.yml*
- Builds two pre-configured stack images:
  - Default: Stalwart + SnappyMail
  - Conservative: Postfix+Dovecot + Roundcube + Radicale
- Published as ghcr.io/trek-e/darkpipe/home-stalwart:TAG-default
- Published as ghcr.io/trek-e/darkpipe/home-postfix-dovecot:TAG-conservative

*.github/workflows/build-custom.yml*
- Manual workflow_dispatch trigger
- Inputs for component selection (mail server, webmail, groupware)
- Builds custom stack based on user selection
- Published with user-selected tag

**Docker Compose Profiles**
- Single docker-compose.yml with profile-based service activation
- Profile selection: --profile stalwart --profile roundcube
- Allows flexible component combinations without multiple compose files

## Data Flow

### Inbound Mail Flow

1. External mail server connects to cloud relay on port 25
2. Postfix receives connection, negotiates TLS
3. Postfix accepts message, passes to Go relay service (port 10025)
4. Go relay service encrypts message and sends via WireGuard/mTLS to home device
5. Home device mail server receives and stores message in mailbox
6. Rspamd scans message for spam (post-delivery filtering)
7. User retrieves message via IMAP (port 993) or webmail

**If home device offline:**
1. Go relay service detects unreachable home device
2. Message encrypted with age (filippo.io/age)
3. Encrypted message queued on local disk
4. If queue exceeds threshold, overflow to S3-compatible storage
5. When home device reconnects, queue drains automatically
6. User configurable: queue (default) or bounce after timeout

### Outbound Mail Flow

1. User sends mail via IMAP submission (port 587) or webmail
2. Home device mail server authenticates user
3. Rspamd scans outbound message
4. Mail server relays message via WireGuard/mTLS to cloud relay
5. Cloud relay Postfix receives message
6. Postfix signs with DKIM
7. Postfix delivers directly to recipient mail server (port 25)
8. Delivery status tracked and logged

### Offline Queue Flow

**Queue Encryption (age)**
- Each message encrypted with age recipient public key
- Recipient private key stored securely on cloud relay
- Encrypted messages are safe to store on disk or S3

**S3 Overflow**
- Triggered when local queue exceeds size threshold
- Encrypted messages uploaded to S3-compatible endpoint
- Supports Storj, AWS S3, MinIO, or any S3-compatible service
- When home device reconnects, S3 messages downloaded and delivered
- Local and S3 messages deleted after successful delivery

**Configurable Behavior**
- RELAY_QUEUE_ENABLED: true (queue) or false (bounce immediately)
- Queue timeout configurable (default: 7 days)
- Overflow threshold configurable (default: 100MB local queue)

## Security Model

See [docs/security.md](security.md) for detailed threat model and security guarantees.

**Key principles:**
- No mail stored unencrypted on cloud relay (pass-through only)
- All transport encrypted (WireGuard or mTLS)
- Internal PKI with automatic rotation
- SPF/DKIM/DMARC for email authentication
- TLS required on all SMTP connections
- Offline queue encrypted at rest

## Directory Structure

```
darkpipe/
├── cloud-relay/              # Cloud relay Docker image and config
│   ├── Dockerfile            # Multi-stage build (Postfix + Go relay service)
│   ├── docker-compose.yml    # Cloud relay deployment
│   └── caddy/                # Reverse proxy config
├── home-device/              # Home device Docker compose and configs
│   ├── docker-compose.yml    # Home device deployment with profiles
│   ├── stalwart/             # Stalwart mail server config
│   ├── maddy/                # Maddy mail server config
│   ├── postfix-dovecot/      # Postfix+Dovecot config and Dockerfile
│   ├── webmail/              # Roundcube and SnappyMail configs
│   ├── caldav-carddav/       # Radicale config
│   ├── spam-filter/          # Rspamd and Redis configs
│   └── profiles/             # Profile server source code
├── transport/                # Transport layer implementation
│   ├── wireguard/            # WireGuard setup scripts
│   ├── mtls/                 # mTLS implementation
│   └── pki/                  # step-ca PKI setup
├── deploy/                   # Deployment tools and guides
│   ├── setup/                # darkpipe-setup CLI source
│   └── platform-guides/      # Platform-specific deployment guides
├── dns/                      # DNS management
│   └── cmd/dns-setup/        # dns-setup CLI source
├── monitoring/               # Monitoring and alerting
│   └── cmd/monitor/          # Monitoring agent source
├── migration/                # Mail migration tools
│   └── cmd/migrate/          # Migration CLI source
├── .github/workflows/        # GitHub Actions CI/CD
├── .planning/                # GSD planning and execution metadata
├── docs/                     # User-facing documentation
├── go.mod                    # Root Go module (relay service, shared packages)
├── LICENSE                   # AGPLv3 license text
├── THIRD-PARTY-LICENSES.md   # Third-party dependency licenses
└── README.md                 # Project overview
```

**Go Module Structure:**
- Root module: Cloud relay service and shared packages
- deploy/setup/go.mod: Setup wizard CLI (isolated dependencies)
- home-device/profiles/go.mod: Profile server (isolated dependencies)

## Technology Stack

- **Go 1.25**: Relay service, CLIs, monitoring, migration tools
- **Postfix**: SMTP MTA
- **Stalwart 0.15.4**: Modern all-in-one mail server (Rust)
- **Maddy 0.8.2**: Minimal Go mail server
- **Dovecot**: IMAP server (with Postfix)
- **Rspamd**: Spam filtering engine
- **Redis**: Rspamd backend
- **Roundcube 1.6.13**: Traditional webmail (PHP)
- **SnappyMail 2.38.2**: Modern webmail (JavaScript)
- **Radicale 3.6.0**: CalDAV/CardDAV server (Python)
- **Caddy 2**: Reverse proxy with automatic HTTPS
- **WireGuard**: VPN transport layer
- **step-ca**: Internal PKI for mTLS
- **age (filippo.io/age)**: Encryption for offline queue
- **Docker**: Containerization
- **Docker Compose**: Service orchestration

---

Last Updated: 2026-02-15

License: AGPLv3 - See [LICENSE](../LICENSE)
