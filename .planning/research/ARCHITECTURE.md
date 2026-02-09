# Architecture Research: DarkPipe

**Domain:** Cloud-Relay + Home-Device Email System
**Researched:** 2026-02-08
**Confidence:** MEDIUM

## Standard Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           INTERNET                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  Inbound SMTP (port 25)           Outbound SMTP (port 25/587)            │
│          ↓                                     ↑                          │
│  ┌───────────────────────────────────────────────────────────┐           │
│  │              CLOUD RELAY (Minimal Gateway)                 │           │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │           │
│  │  │ SMTP Receive │→ │ Ephemeral    │→ │ SMTP Forward │    │           │
│  │  │ (Postfix)    │  │ Queue (RAM)  │  │ (Postfix)    │    │           │
│  │  └──────────────┘  └──────────────┘  └──────────────┘    │           │
│  │         ↓                   ↓                                          │
│  │  ┌──────────────┐  ┌──────────────┐                                   │
│  │  │ TLS/Certbot  │  │ Overflow to  │                                   │
│  │  │ (Let's Enc)  │  │ S3 Storage   │                                   │
│  │  └──────────────┘  └──────────────┘                                   │
│  └───────────────────────────────────────────────────────────┘           │
│                                ↕                                          │
│                    TRANSPORT LAYER                                        │
│              (WireGuard Tunnel OR mTLS Connection)                        │
│                                ↕                                          │
│  ┌───────────────────────────────────────────────────────────┐           │
│  │         HOME DEVICE (Full Mail Stack on RPi4+)             │           │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │           │
│  │  │ Transport    │  │ Mail Server  │  │ Webmail      │    │           │
│  │  │ Endpoint     │  │ (Postfix/    │  │ (Roundcube/  │    │           │
│  │  │ (WG/mTLS)    │→ │  Dovecot)    │  │  others)     │    │           │
│  │  └──────────────┘  └──────┬───────┘  └──────────────┘    │           │
│  │  ┌──────────────┐         │                                           │
│  │  │ CalDAV/      │←────────┘                                           │
│  │  │ CardDAV      │                                                      │
│  │  │ (Radicale)   │                                                      │
│  │  └──────────────┘                                                      │
│  │         ↓                                                              │
│  │  ┌──────────────────────────────────────────┐                         │
│  │  │ Persistent Storage (Docker Volumes)      │                         │
│  │  └──────────────────────────────────────────┘                         │
│  └───────────────────────────────────────────────────────────┐           │
└──────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Cloud Relay (SMTP Gateway)** | Minimal SMTP gateway that receives mail and immediately forwards via transport layer | Postfix in relay-only mode, smallest possible container |
| **Ephemeral Queue** | In-memory queue for active forwarding; overflow to encrypted S3-compatible storage | RAM-based spool with configurable overflow threshold |
| **Transport Layer** | Secure, NAT-traversing persistent connection between cloud and home | WireGuard tunnel (primary) or mTLS persistent connection |
| **Home Device Mail Stack** | Full-featured mail server with IMAP, SMTP, CalDAV, CardDAV | Docker Mailserver or component stack (Postfix + Dovecot + Radicale + webmail) |
| **Certificate Management** | TLS certificates for cloud relay (public) and relay↔home (internal) | Certbot for public certs, internal CA for relay-to-home authentication |
| **Build System** | User-configurable CI/CD that produces custom multi-arch Docker images | GitHub Actions with matrix builds for arm64/amd64 |
| **DNS Management** | Auto-generate and validate required DNS records (MX, SPF, DKIM, DMARC) | DNS provider API integration with validation |
| **Configuration** | User-friendly config format that drives all components | YAML primary, env vars for secrets, CLI wizard for initial setup |

## Recommended Project Structure

```
darkpipe/
├── cloud-relay/              # Minimal SMTP gateway
│   ├── Dockerfile            # Multi-stage build for smallest image
│   ├── postfix/              # Relay-only Postfix config
│   │   ├── main.cf.template  # Template with variable substitution
│   │   └── master.cf         # Process configuration
│   ├── transport/            # WireGuard or mTLS endpoint
│   │   ├── wireguard/        # WireGuard config templates
│   │   └── mtls/             # mTLS connection handler
│   ├── queue/                # Ephemeral queue with overflow
│   │   ├── ram-spool/        # In-memory queue manager
│   │   └── s3-overflow/      # S3-compatible backup storage
│   └── certbot/              # Let's Encrypt automation
│       └── renewal-hooks/    # Certificate rotation scripts
├── home-device/              # Home mail stack
│   ├── single-container/     # All-in-one Docker Mailserver approach
│   │   └── Dockerfile        # Extended docker-mailserver image
│   ├── compose-stack/        # Multi-container alternative
│   │   ├── docker-compose.yml
│   │   ├── postfix/          # MTA configuration
│   │   ├── dovecot/          # IMAP/POP3 server
│   │   ├── radicale/         # CalDAV/CardDAV server
│   │   └── roundcube/        # Webmail interface
│   └── transport/            # WireGuard or mTLS client
│       ├── wireguard/        # WireGuard peer config
│       └── mtls/             # mTLS client with reconnection
├── build-system/             # GitHub Actions workflows
│   ├── .github/workflows/
│   │   ├── build-cloud.yml   # Cloud relay build (amd64)
│   │   ├── build-home.yml    # Home device build (arm64/amd64)
│   │   └── user-config.yml   # User-triggered custom build
│   └── config-templates/     # Component selection templates
├── dns-tools/                # DNS automation
│   ├── generator/            # DNS record generation from config
│   ├── validators/           # Pre-deployment validation
│   └── providers/            # DNS provider API integrations
│       ├── cloudflare.go
│       ├── route53.go
│       └── generic-api.go
├── config/                   # Configuration management
│   ├── darkpipe.yaml.example # Main configuration template
│   ├── schema.json           # JSON schema for validation
│   └── wizard/               # Interactive CLI setup
│       └── setup.go
└── docs/                     # User documentation
    ├── architecture.md       # This document
    ├── deployment.md         # Deployment guide
    └── troubleshooting.md    # Common issues
```

### Structure Rationale

- **cloud-relay/**: Separated by deployment target (cloud VPS). Single-purpose design for minimal attack surface and resource usage.
- **home-device/**: Provides both single-container and compose-stack options to support different use cases (resource-constrained RPi4 vs. more powerful home servers).
- **build-system/**: User-driven CI/CD allows users to select components without maintaining their own build infrastructure.
- **dns-tools/**: DNS automation is critical for email delivery; separating this allows pre-deployment validation and multi-provider support.
- **config/**: Single source of truth (YAML) that drives all other components, reducing configuration drift.

## Architectural Patterns

### Pattern 1: Ephemeral Cloud Relay with Persistent Home Storage

**What:** Cloud relay stores nothing persistently (except optional encrypted overflow queue). All permanent storage lives on home device.

**When to use:** When privacy is paramount and cloud infrastructure must be minimal and stateless.

**Trade-offs:**
- **Pro:** Zero persistent cloud storage means no data breach exposure; minimal cloud costs.
- **Pro:** Cloud relay container can be ultra-small (50-100MB) and horizontally scalable.
- **Con:** Home device must be highly available or mail delivery fails.
- **Con:** Transport layer becomes single point of failure; requires robust reconnection logic.

**Example:**
```yaml
# cloud-relay postfix main.cf
# Relay immediately, no local delivery
mydestination =
local_recipient_maps =
relay_domains = $mydomain
transport_maps = hash:/etc/postfix/transport

# Minimal queue
queue_run_delay = 30s
minimal_backoff_time = 60s
maximal_queue_lifetime = 4h

# Forward via WireGuard tunnel to home device
# /etc/postfix/transport
example.com    smtp:[10.8.0.2]:25
```

### Pattern 2: WireGuard Hub-and-Spoke with NAT Traversal

**What:** Cloud relay acts as WireGuard hub (static IP, always reachable). Home device acts as spoke (behind NAT, connects outbound to hub). Uses persistent keepalive for NAT hole punching.

**When to use:** When home device is behind residential NAT and cannot accept inbound connections directly.

**Trade-offs:**
- **Pro:** Works with any residential ISP, no port forwarding required.
- **Pro:** WireGuard's performance is excellent (minimal overhead, uses kernel cryptography).
- **Pro:** Simple configuration; WireGuard handles roaming IPs automatically.
- **Con:** Home device must maintain persistent connection (battery consideration for mobile devices).
- **Con:** Cellular NATs with strict UDP randomization may require fallback.

**Example:**
```ini
# Cloud relay WireGuard config
[Interface]
PrivateKey = <cloud-private-key>
Address = 10.8.0.1/24
ListenPort = 51820

[Peer]
PublicKey = <home-device-public-key>
AllowedIPs = 10.8.0.2/32

# Home device WireGuard config
[Interface]
PrivateKey = <home-private-key>
Address = 10.8.0.2/24

[Peer]
PublicKey = <cloud-public-key>
Endpoint = relay.example.com:51820
AllowedIPs = 10.8.0.1/32
PersistentKeepalive = 25  # NAT hole punch every 25s
```

### Pattern 3: mTLS Persistent Connection with Reconnection Handling

**What:** Alternative to WireGuard. Cloud relay and home device maintain persistent mTLS connection with mutual certificate authentication. Home device initiates connection (NAT-friendly). Connection pooling and automatic reconnection.

**When to use:** When WireGuard is unavailable (restrictive networks) or when HTTP/2 multiplexing over single connection is beneficial.

**Trade-offs:**
- **Pro:** Works through HTTP-aware proxies and restrictive firewalls.
- **Pro:** Certificate-based authentication eliminates pre-shared keys.
- **Pro:** Connection pooling allows multiple mail streams over single connection.
- **Con:** More complex to implement than WireGuard.
- **Con:** Higher CPU overhead for TLS handshakes (mitigated by session resumption).
- **Con:** Requires careful certificate expiration and renewal automation.

**Example:**
```go
// Home device mTLS client with reconnection
func maintainConnection(ctx context.Context) {
    backoff := 1 * time.Second
    for {
        select {
        case <-ctx.Done():
            return
        default:
            conn, err := dialMTLS()
            if err != nil {
                log.Printf("Connection failed: %v, retrying in %v", err, backoff)
                time.Sleep(backoff)
                backoff = min(backoff*2, 5*time.Minute)
                continue
            }
            backoff = 1 * time.Second
            handleConnection(conn) // Blocks until connection drops
        }
    }
}
```

### Pattern 4: Single Container vs. Docker Compose Stack

**What:** Two deployment approaches for home device:
1. **Single container**: Extends docker-mailserver with all components in one image.
2. **Compose stack**: Separate containers for Postfix, Dovecot, Radicale, webmail.

**When to use:**
- **Single container**: Resource-constrained devices (RPi4 with 4GB RAM), simplicity priority.
- **Compose stack**: More powerful home servers (8GB+ RAM), component flexibility priority.

**Trade-offs:**

**Single Container:**
- **Pro:** Lower memory overhead (shared processes, no duplicate libraries).
- **Pro:** Simpler deployment (single `docker run` command).
- **Pro:** Easier updates (single image pull).
- **Con:** All components must restart together (no isolated updates).
- **Con:** Harder to customize individual components.

**Compose Stack:**
- **Pro:** Independent scaling and updates per component.
- **Pro:** Easier to swap components (e.g., replace Roundcube with Rainloop).
- **Pro:** Resource limits per container for better control.
- **Con:** Higher memory overhead (duplicate base images, separate processes).
- **Con:** More complex networking (inter-container communication).

### Pattern 5: Build-Time Configuration via GitHub Actions

**What:** Users fork repository, configure YAML file, trigger GitHub Actions workflow that builds custom multi-arch Docker images with their configuration baked in.

**When to use:** When distributing a configurable product that users deploy to their own infrastructure.

**Trade-offs:**
- **Pro:** No runtime configuration complexity; images are ready-to-run.
- **Pro:** User-specific builds can optimize for their exact component selection.
- **Pro:** Multi-arch builds (amd64 for cloud, arm64 for RPi) handled automatically.
- **Con:** Re-build required for configuration changes (not runtime reconfigurable).
- **Con:** GitHub Actions minutes consumption (mitigated by caching).
- **Con:** Users need GitHub account and basic understanding of workflows.

**Example workflow:**
```yaml
# .github/workflows/build-home.yml
name: Build Home Device Image
on:
  workflow_dispatch:
    inputs:
      enable_caldav:
        description: 'Enable CalDAV/CardDAV'
        required: true
        type: boolean
      webmail_choice:
        description: 'Webmail interface'
        required: true
        type: choice
        options:
          - roundcube
          - rainloop
          - none

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        platform: [linux/amd64, linux/arm64]
    steps:
      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v5
        with:
          platforms: ${{ matrix.platform }}
          build-args: |
            ENABLE_CALDAV=${{ inputs.enable_caldav }}
            WEBMAIL=${{ inputs.webmail_choice }}
          tags: ghcr.io/${{ github.repository }}/darkpipe-home:latest
          push: true
```

### Pattern 6: DNS Automation with Pre-Deployment Validation

**What:** CLI tool generates required DNS records (MX, SPF, DKIM, DMARC) from configuration, integrates with DNS provider APIs to apply them, and validates before deploying mail services.

**When to use:** Always. Email delivery fails without correct DNS configuration.

**Trade-offs:**
- **Pro:** Eliminates most common deployment error (incorrect DNS).
- **Pro:** Provider-agnostic abstraction over various DNS APIs.
- **Con:** Requires DNS provider API credentials (security consideration).
- **Con:** Not all DNS providers have APIs (fallback to manual instructions).

**Example:**
```bash
# Generate DNS records from config
$ darkpipe dns generate --config darkpipe.yaml
Generated DNS records for example.com:
  MX     10 relay.example.com
  TXT    "v=spf1 ip4:203.0.113.10 -all"
  TXT    (DKIM key for selector 'default')
  TXT    "_dmarc" "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"

# Validate before deployment
$ darkpipe dns validate --config darkpipe.yaml
✓ MX record points to relay.example.com (203.0.113.10)
✓ SPF record authorizes relay IP
✗ DKIM record not found (deployment will fail)
✗ DMARC record missing

# Apply via DNS provider API
$ darkpipe dns apply --config darkpipe.yaml --provider cloudflare
Applied 4 DNS records to Cloudflare
Waiting for propagation (this may take up to 5 minutes)...
✓ All records validated
```

## Data Flow

### Inbound Mail Flow (Internet → Mailbox)

```
┌─────────────┐
│ Sender's    │
│ Mail Server │
└──────┬──────┘
       │ SMTP (port 25)
       ↓
┌─────────────────────────────────────────────┐
│ Cloud Relay                                  │
│ 1. Postfix receives on port 25              │
│ 2. TLS negotiation (STARTTLS)               │
│ 3. Anti-spam checks (basic)                 │
│ 4. Enqueue to RAM spool                     │
│ 5. Dequeue immediately (if transport up)    │
└──────┬──────────────────────────────────────┘
       │ SMTP over WireGuard tunnel (10.8.0.2:25)
       │ OR SMTP over mTLS persistent connection
       ↓
┌─────────────────────────────────────────────┐
│ Home Device                                  │
│ 6. Transport endpoint receives              │
│ 7. Forward to local Postfix (localhost:25)  │
│ 8. Postfix applies local rules              │
│ 9. MDA (Dovecot LMTP) delivers to mailbox   │
│ 10. Mail stored in Docker volume             │
└──────┬──────────────────────────────────────┘
       │ IMAP/POP3 (ports 993/995)
       ↓
┌─────────────┐
│ User's Mail │
│ Client      │
└─────────────┘
```

**Critical decision points:**
- **Step 4-5**: If transport is down, queue in RAM. If RAM threshold exceeded, overflow to encrypted S3 storage.
- **Step 7**: Authentication happens here via WireGuard tunnel (network-level) or mTLS (application-level).
- **Step 8**: Home device applies user-specific filtering, spam rules, etc. Cloud relay does minimal processing.

### Outbound Mail Flow (Mailbox → Internet)

```
┌─────────────┐
│ User's Mail │
│ Client      │
└──────┬──────┘
       │ SMTP Submission (port 587 with AUTH)
       ↓
┌─────────────────────────────────────────────┐
│ Home Device                                  │
│ 1. Postfix receives on port 587             │
│ 2. User authentication (SASL)               │
│ 3. DKIM signing                              │
│ 4. Route via transport to cloud relay       │
└──────┬──────────────────────────────────────┘
       │ SMTP over WireGuard tunnel (10.8.0.1:587)
       │ OR SMTP over mTLS persistent connection
       ↓
┌─────────────────────────────────────────────┐
│ Cloud Relay                                  │
│ 5. Receive from home via transport          │
│ 6. Verify source (WireGuard peer/mTLS cert) │
│ 7. Rewrite headers (hide home IP)           │
│ 8. Send to destination via SMTP (port 25)   │
└──────┬──────────────────────────────────────┘
       │ SMTP (port 25)
       ↓
┌─────────────┐
│ Recipient's │
│ Mail Server │
└─────────────┘
```

**Critical decision points:**
- **Step 4**: Home device trusts cloud relay to send on its behalf. Cloud relay IP must have good reputation.
- **Step 6**: Cloud relay MUST verify source via transport authentication to prevent open relay abuse.
- **Step 7**: Cloud relay rewrites `Received:` headers to show relay IP, not home IP (privacy).

### Calendar/Contact Sync Flow (CalDAV/CardDAV)

```
┌─────────────┐
│ User's      │
│ Calendar    │
│ Client      │
└──────┬──────┘
       │ HTTPS (CalDAV/CardDAV)
       ↓
┌─────────────────────────────────────────────┐
│ Home Device (via WireGuard or VPN)          │
│ 1. Radicale receives HTTPS request          │
│ 2. Authentication (HTTP Basic/Digest)       │
│ 3. Read/write iCal/vCard files              │
│ 4. Store in Docker volume                   │
└─────────────────────────────────────────────┘
```

**Note:** CalDAV/CardDAV do NOT route through cloud relay. Direct connection to home device required (WireGuard tunnel or VPN).

### Update/Upgrade Flow

```
┌─────────────────────────────────────────────┐
│ 1. User triggers update                     │
│    $ docker pull darkpipe/home:latest       │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 2. Docker pulls new image                   │
│    (configuration baked into image)         │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 3. Health check on new container            │
│    Wait until new container responds        │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 4. Route traffic to new container           │
│    (load balancer or Docker networking)     │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 5. Graceful shutdown of old container       │
│    Wait for in-flight mail to complete      │
└──────┬──────────────────────────────────────┘
       ↓
┌─────────────────────────────────────────────┐
│ 6. Remove old container                     │
│    Data persists in Docker volumes          │
└─────────────────────────────────────────────┘
```

**Critical considerations:**
- **Configuration persistence**: User config stored in Docker volumes (`.env` files or YAML), not baked into images. Alternative: rebuild custom image via GitHub Actions.
- **Zero-downtime strategy**: Use Docker Compose with `update_config` for rolling updates, or blue-green deployment with Nginx routing.
- **Database migrations**: Mail formats rarely change, but if schema updates required, include migration scripts in entrypoint.

## Scaling Considerations

| Concern | Single User (10 emails/day) | Power User (100 emails/day) | Small Organization (1000 emails/day) |
|---------|----------------------------|------------------------------|--------------------------------------|
| **Cloud Relay** | 512MB VPS ($3-5/mo) | 1GB VPS ($5-10/mo) | 2GB VPS ($10-20/mo) with horizontal scaling |
| **Home Device** | RPi4 4GB | RPi4 8GB or NUC | Dedicated server or TrueNAS Scale |
| **Transport Layer** | Single WireGuard tunnel | Single WireGuard tunnel | Multiple WireGuard tunnels or mTLS connection pool |
| **Storage** | 10GB Docker volume | 50GB Docker volume | 500GB+ with S3 archival |
| **Container Strategy** | Single container | Single container or Compose | Compose stack with resource limits |

### Scaling Priorities

1. **First bottleneck: Transport layer reconnection**
   - **What breaks first:** If home device loses connection frequently (unreliable network), cloud relay queue fills up.
   - **How to fix:** Implement robust reconnection logic with exponential backoff. Add monitoring/alerting for transport state. Consider overflow to S3-compatible storage (e.g., Storj) for temporary queue persistence.

2. **Second bottleneck: Home device CPU (spam filtering)**
   - **What breaks next:** SpamAssassin and ClamAV are CPU-intensive on RPi4.
   - **How to fix:** Move heavy spam filtering to cloud relay (breaks privacy model) OR disable ClamAV (rely on sender's scanning) OR upgrade to more powerful home device (NUC with i5+).

3. **Third bottleneck: Cloud relay reputation**
   - **What breaks next:** If cloud relay IP gets blacklisted (compromised account sending spam), all mail delivery fails.
   - **How to fix:** Implement rate limiting, strict authentication, DMARC monitoring. Use multiple cloud relays with DNS round-robin. Consider dedicated IP from VPS provider.

## Anti-Patterns

### Anti-Pattern 1: Storing Mail Persistently in Cloud Relay

**What people do:** Configure cloud relay with local mailboxes, thinking it's a backup.

**Why it's wrong:** Defeats the entire privacy model of DarkPipe. Cloud provider has access to plaintext mail. Increases attack surface. Regulatory compliance issues (GDPR, etc.).

**Do this instead:** Cloud relay should be 100% stateless. Only ephemeral queue in RAM with optional encrypted overflow to S3. If user wants cloud backup, they should use encrypted backup of home device volumes.

### Anti-Pattern 2: Using Let's Encrypt Certificates for Internal Transport

**What people do:** Request Let's Encrypt certs for both cloud relay (correct) and internal WireGuard IPs (wrong).

**Why it's wrong:** Let's Encrypt requires public DNS validation. Internal IPs (10.8.0.x) cannot be validated. Also, WireGuard doesn't use TLS (has its own encryption), and mTLS should use internal CA for client certs.

**Do this instead:**
- **Cloud relay**: Let's Encrypt certs via Certbot for public SMTP (STARTTLS).
- **WireGuard transport**: No TLS needed; WireGuard provides encryption.
- **mTLS transport**: Internal CA for mutual authentication certificates.

### Anti-Pattern 3: Running Certbot Inside Docker Container Without Volume Persistence

**What people do:** Include Certbot in Docker image, run cert renewal inside container, forget to persist `/etc/letsencrypt`.

**Why it's wrong:** Certificates are lost on container restart. Let's Encrypt has rate limits (5 certs per domain per week). Container restart triggers new cert request, hitting rate limit, breaking mail delivery.

**Do this instead:** Mount `/etc/letsencrypt` as Docker volume. Use Certbot renewal hooks to reload Postfix. Alternative: use external cert management (certbot on host, mount certs into container).

### Anti-Pattern 4: Baking Secrets Into Docker Images

**What people do:** Include DKIM private keys, WireGuard private keys, passwords in Dockerfile or config files committed to Git.

**Why it's wrong:** Secrets exposed in image layers (even if deleted in later layer). Anyone with image access has secrets. GitHub Actions logs may leak secrets.

**Do this instead:** Secrets passed via environment variables at runtime or Docker secrets (Swarm) or mounted from encrypted volumes. Build system generates keys at deployment time, never stores in Git.

### Anti-Pattern 5: No Health Checks in Docker Compose

**What people do:** Deploy containers without `healthcheck` configuration, assuming "if container runs, it's healthy."

**Why it's wrong:** Postfix/Dovecot may start but fail to bind to ports (port conflict). Service may be running but rejecting connections (config error). Load balancer routes traffic to unhealthy container.

**Do this instead:**
```yaml
services:
  mail:
    image: darkpipe/home:latest
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "25"]  # Check SMTP port
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s  # Allow time for Postfix to start
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **DNS Providers** | REST API (Cloudflare, Route53, etc.) | Required for automated record creation; fallback to manual instructions if no API |
| **S3-Compatible Storage** | AWS SDK (works with Storj, Backblaze B2, MinIO) | For encrypted queue overflow; optional feature |
| **Let's Encrypt** | ACME protocol via Certbot | Cloud relay only; automatic renewal with DNS-01 or HTTP-01 challenge |
| **Monitoring/Alerting** | Prometheus metrics + Grafana OR simple health endpoint | For transport layer status, queue depth, delivery rates |
| **VPS Providers** | Manual deployment or Terraform | User deploys cloud relay to their chosen provider |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| **Cloud Relay ↔ Home Device** | SMTP over WireGuard (encrypted tunnel) OR SMTP over mTLS (persistent connection) | Primary data path; must be highly reliable |
| **Postfix ↔ Dovecot (Home)** | LMTP (Local Mail Transfer Protocol) over localhost | Standard MTA-to-MDA handoff; no authentication needed (localhost trusted) |
| **Webmail ↔ Dovecot (Home)** | IMAP over localhost | Roundcube/Rainloop connects via IMAP to Dovecot on same host |
| **CalDAV/CardDAV ↔ Storage** | File I/O to Docker volume | Radicale stores iCal/vCard files directly; no database |
| **GitHub Actions ↔ User Config** | YAML file in user's forked repo | User commits config, triggers workflow, receives custom image |

## Build Order Implications

Based on dependency analysis, recommended build order for development:

### Phase 1: Foundation (No Dependencies)
1. **Configuration schema and validation** - Defines data structures for all other components.
2. **DNS automation library** - Standalone; no dependencies; critical for all deployments.
3. **Build system (GitHub Actions workflows)** - Enables user-driven builds from day one.

### Phase 2: Transport Layer (Depends on Config)
4. **WireGuard tunnel setup** - Simpler than mTLS; implement first.
5. **mTLS persistent connection** - Alternative transport; shares config schema with WireGuard.

### Phase 3: Cloud Relay (Depends on Transport)
6. **Minimal Postfix relay container** - Core cloud relay; depends on transport to forward mail.
7. **Ephemeral queue with RAM spool** - Cloud relay needs queue; start with RAM-only (simpler).
8. **S3-compatible overflow storage** - Optional feature; add after RAM queue works.
9. **Certbot automation** - Cloud relay needs public TLS; integrate Let's Encrypt.

### Phase 4: Home Device (Depends on Transport)
10. **Postfix + Dovecot single container** - Core home mail server; simplest deployment.
11. **Docker Compose stack alternative** - Multi-container option; shares config with single container.
12. **CalDAV/CardDAV integration (Radicale)** - Optional component; depends on home device foundation.
13. **Webmail interface (Roundcube)** - Optional component; depends on Dovecot IMAP.

### Phase 5: User Experience (Depends on All)
14. **CLI wizard for initial setup** - Guides user through config creation; needs understanding of all components.
15. **Update/upgrade automation** - Zero-downtime updates; needs full system understanding.
16. **Monitoring and health checks** - Observability across all components.

### Dependency Graph

```
Configuration Schema
  ├─→ DNS Automation
  ├─→ Build System
  ├─→ WireGuard Setup
  │     ├─→ Cloud Relay (Postfix)
  │     │     ├─→ Ephemeral Queue
  │     │     │     └─→ S3 Overflow
  │     │     └─→ Certbot
  │     └─→ Home Device (Postfix+Dovecot)
  │           ├─→ Compose Stack Alternative
  │           ├─→ CalDAV/CardDAV
  │           └─→ Webmail
  ├─→ mTLS Setup (Alternative to WireGuard)
  │     └─→ (same downstream as WireGuard)
  └─→ CLI Wizard (needs all components)
        └─→ Update/Upgrade System
              └─→ Monitoring
```

**Critical path:** Configuration → WireGuard → Cloud Relay → Home Device → CLI Wizard

**Parallel development opportunities:**
- DNS Automation and Build System can be developed in parallel.
- WireGuard and mTLS are alternatives; pick one for MVP, add the other later.
- CalDAV/CardDAV and Webmail are optional; can be developed after core mail works.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| **Cloud Relay (Postfix)** | HIGH | Well-documented standard practice. [Postfix docs](https://www.postfix.org/BASIC_CONFIGURATION_README.html) authoritative. |
| **WireGuard Transport** | HIGH | Mature protocol with excellent documentation. [Official WireGuard docs](https://www.wireguard.com/quickstart/) and [NAT traversal research](https://nordvpn.com/blog/achieving-nat-traversal-with-wireguard/) confirm patterns. |
| **mTLS Patterns** | MEDIUM | Well-established pattern but implementation details vary. [Medium article on connection handling](https://medium.com/@wolfroma/handling-mtls-connection-spikes-haproxy-tomcat-httpclient-ecab9f18707e) provides real-world guidance. |
| **Home Device (Docker Mailserver)** | HIGH | [docker-mailserver](https://github.com/docker-mailserver/docker-mailserver) is production-ready and well-documented for single-container deployment. |
| **CalDAV/CardDAV Integration** | MEDIUM-LOW | Separate applications (like Radicale) are the norm. No evidence of unified Postfix+Dovecot+CalDAV stack. [Dovecot mailing list discussions](https://dovecot.org/pipermail/dovecot/2022-October/125533.html) confirm this. |
| **S3-Compatible Queue** | MEDIUM | [Maddy mail server](https://maddy.email/reference/blob/s3/) demonstrates S3 blob storage for mail. Stalwart also supports S3. Pattern is proven but not mainstream. |
| **GitHub Actions Multi-Arch** | HIGH | [Official Docker docs](https://docs.docker.com/build/ci/github-actions/multi-platform/) and [multiple community examples](https://github.com/sredevopsorg/multi-arch-docker-github-workflow) provide clear patterns. |
| **DNS Automation** | MEDIUM | Provider APIs are well-documented (Cloudflare, Route53), but [DNS-PERSIST-01](https://www.certkit.io/blog/dns-persist-01) is new for 2026 and may not be widely supported yet. |
| **Zero-Downtime Updates** | MEDIUM | [Docker update patterns](https://oneuptime.com/blog/post/2026-01-06-docker-update-without-downtime/view) are documented, but mail-specific considerations (queue draining) need validation. |
| **RPi4 Resource Constraints** | MEDIUM | [Community reports](https://forums.raspberrypi.com/viewtopic.php?t=310308) confirm Docker Mailserver works on RPi4, but resource limits need empirical testing. |

## Open Questions Requiring Phase-Specific Research

1. **Queue overflow threshold**: What RAM threshold triggers S3 overflow? How to calculate based on average email size and delivery latency?

2. **Reconnection backoff strategy**: What backoff algorithm balances quick reconnection vs. not hammering cloud relay? Needs simulation/testing.

3. **CalDAV/CardDAV without separate service**: Is there a way to serve CalDAV/CardDAV from Dovecot directly, or is Radicale/Sabre always required?

4. **Certificate rotation without downtime**: How to rotate mTLS certificates (both cloud and home) without breaking persistent connection? Needs protocol-level investigation.

5. **Multi-user support on single home device**: Architecture assumes single user. What changes needed for family (5-10 users) on one RPi4?

6. **ARM64 performance**: Quantify SpamAssassin and ClamAV performance on RPi4 arm64 vs. amd64 NUC. May influence default component selection.

7. **DNS propagation delays**: How to handle initial deployment when DNS records take 24-48 hours to propagate? Does cloud relay need to queue mail for extended period?

8. **Backup strategy**: Where should automated backups fit in architecture? Cloud relay? Home device? User's responsibility?

## Sources

### High Confidence Sources (Official Documentation)
- [Postfix Basic Configuration](http://www.postfix.org/BASIC_CONFIGURATION_README.html)
- [Postfix Standard Configuration Examples](https://www.postfix.org/STANDARD_CONFIGURATION_README.html)
- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [Docker Multi-Platform Images with GitHub Actions](https://docs.docker.com/build/ci/github-actions/multi-platform/)
- [docker-mailserver GitHub Repository](https://github.com/docker-mailserver/docker-mailserver)
- [Certbot User Guide](https://eff-certbot.readthedocs.io/en/stable/using.html)
- [Let's Encrypt Challenge Types](https://letsencrypt.org/docs/challenge-types/)
- [Docker Compose Environment Variables](https://docs.docker.com/compose/how-tos/environment-variables/set-environment-variables/)
- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

### Medium Confidence Sources (Verified with Multiple Sources)
- [SMTP Relay Explained - Mailtrap 2026](https://mailtrap.io/blog/smtp-relay/)
- [How NAT Traversal Works - NordVPN](https://nordvpn.com/blog/achieving-nat-traversal-with-wireguard/)
- [WireGuard NAT Traversal - Nettica](https://nettica.com/nat-traversal-hole-punch/)
- [Email Forwarding FAQ - Forward Email](https://forwardemail.net/en/faq)
- [Docker Mailserver Basic Installation Tutorial](https://docker-mailserver.github.io/docker-mailserver/latest/examples/tutorials/basic-installation/)
- [Handling mTLS Connection Spikes - Medium](https://medium.com/@wolfroma/handling-mtls-connection-spikes-haproxy-tomcat-httpclient-ecab9f18707e)
- [Certificate Rotation Best Practices - Workik](https://workik.com/certificate-rotation-script-generator)
- [Multi-Arch Docker Images with GitHub Actions - Red Hat](https://developers.redhat.com/articles/2023/12/08/build-multi-architecture-container-images-github-actions)
- [DNS-PERSIST-01 Validation - CertKit](https://www.certkit.io/blog/dns-persist-01)
- [How to Update Docker Container with Zero Downtime - Atlantic.Net](https://www.atlantic.net/vps-hosting/how-to-update-docker-container-with-zero-downtime/)
- [How Email Works: MUA, MSA, MTA, MDA - Oxilor](https://oxilor.com/blog/how-does-email-work)
- [Docker Container Resource Limits - OneUpTime](https://oneuptime.com/blog/post/2026-01-30-docker-container-resource-limits/view)

### Community Sources (Lower Confidence, Needs Validation)
- [Raspberry Pi Email Server Options - Forward Email Blog](https://forwardemail.net/en/blog/open-source/raspberry-pi-email-server)
- [Raspberry Pi docker-mailserver Discussion - RPi Forums](https://forums.raspberrypi.com/viewtopic.php?t=310308)
- [Adding CalDAV/CardDAV Next to Dovecot - Dovecot Mailing List](https://dovecot.org/pipermail/dovecot/2022-October/125533.html)
- [S3-Compatible Storage Solutions 2026 - Cloudian](https://cloudian.com/guides/s3-storage/best-s3-compatible-storage-solutions-top-5-in-2026/)
- [Maddy Mail Server S3 Storage](https://maddy.email/reference/blob/s3/)

---

**Architecture research for:** Cloud-Relay + Home-Device Email System (DarkPipe)
**Researched:** 2026-02-08
**Overall Confidence:** MEDIUM - Core patterns (Postfix, WireGuard, Docker) are HIGH confidence from official sources. Advanced patterns (S3 queue overflow, CalDAV integration, zero-downtime updates) are MEDIUM confidence requiring validation during implementation.
