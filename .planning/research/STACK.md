# Technology Stack

**Project:** DarkPipe
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Executive Summary

DarkPipe requires a split-stack architecture: minimal cloud relay (15-50MB container) and full-featured home server. Stalwart or Maddy for Rust/Go single-binary deployments, Postfix for battle-tested minimal relay, WireGuard kernel module for transport, step-ca for internal PKI, and Go for orchestration glue code.

---

## Cloud Relay Stack (Minimal Container)

### SMTP Relay/MTA

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Postfix (Alpine)** | 3.7.4+ | Internet-facing SMTP relay | Battle-tested, 15-30MB container, stable API, excellent documentation, standard choice for minimal relays | HIGH |
| **Maddy** | 0.8.2 | All-in-one Go mail server | Single 15MB binary, Go-native, ~15MB memory footprint, built-in DKIM/SPF/DMARC, excellent for minimal cloud relay | MEDIUM-HIGH |
| **Haraka** | Latest | Node.js SMTP relay | Async/event-driven, plugin architecture, good for custom filtering, but Node.js runtime adds ~50MB vs Go/Rust | MEDIUM |

**Recommendation:** Use **Postfix** for cloud relay. Proven minimal footprint, decades of production hardening, perfect match for "receive and forward" relay pattern. Maddy is excellent backup if Go-native stack preferred.

**Why not Stalwart for relay:** Stalwart is a full mail server (150MB+ with dependencies). Overkill for cloud relay that only needs SMTP receive + forward.

### Container Base Image

| Technology | Size | Purpose | Why Recommended | Confidence |
|------------|------|---------|-----------------|------------|
| **Alpine Linux** | ~5MB | Minimal container base | Package manager available, musl libc, broad hardware support, proven for Postfix/WireGuard, debugging tools available | HIGH |
| **Distroless (Debian)** | ~2MB | Minimal security-focused base | No shell/package manager, better security posture, but harder to debug, requires static binaries | MEDIUM-HIGH |
| **Scratch** | 0MB | Empty base for static binaries | Smallest possible, but no CA certs, no DNS resolution, requires bundling everything | MEDIUM |

**Recommendation:** Use **Alpine** for cloud relay. Trade 3MB for massive operational benefits: shell for debugging, package manager for updates, standard troubleshooting tools. Security through minimal attack surface + regular updates > distroless complexity.

**For Go glue code containers:** Use distroless or scratch with static binary.

---

## Home Server Stack (Full-Featured)

### Mail Server (User-Selectable)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Stalwart** | 0.15.4 (v1.0.0 Q2 2026) | All-in-one Rust mail server | Single binary, JMAP/IMAP4rev2/POP3/SMTP, built-in CalDAV/CardDAV, SQLite/RocksDB storage, memory-safe, REST API, excellent ARM64 support | HIGH |
| **Maddy** | 0.8.2 | All-in-one Go mail server | Single binary, replaces Postfix+Dovecot+OpenDKIM, 15MB memory, built-in DKIM/SPF/DMARC, simple config | HIGH |
| **Postfix + Dovecot** | Postfix 3.7+, Dovecot 2.3+ | Traditional split MTA+MDA | Most battle-tested, maximum flexibility, best documentation, proven ARM64 support, but complex configuration | HIGH |

**Recommendation:** Default to **Stalwart** for modern deployments. Single binary, modern protocols (JMAP), built-in calendar/contacts server, excellent security (memory-safe Rust). **Maddy** for Go preference or tighter resource constraints. **Postfix+Dovecot** for maximum compatibility or existing expertise.

**Why Stalwart over Maddy:**
- Built-in CalDAV/CardDAV (eliminates separate Radicale/Baikal)
- JMAP support (modern protocol, better than IMAP for sync)
- v1.0.0 due Q2 2026 with stable schema/auto-upgrades
- REST API for automation

**Why Maddy over Stalwart:**
- Smaller memory footprint (~15MB vs Stalwart's ~50MB)
- Simpler if only need email (no calendar/contacts)
- Go codebase if team prefers Go

**Why Postfix+Dovecot:**
- 20+ years production hardening
- Most documentation/examples available
- Maximum flexibility for complex routing
- ISP/enterprise familiarity

### Calendar/Contacts Server (If Not Using Stalwart)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Radicale** | 3.x | Lightweight CalDAV/CardDAV | Python, minimal dependencies, file-system storage, 5-10MB memory, simple setup, GPLv3 | MEDIUM-HIGH |
| **Baikal** | Latest | PHP CalDAV/CardDAV with web UI | Admin interface, MySQL/SQLite, more polished UI, 5M+ Docker pulls, multi-arch support | MEDIUM-HIGH |

**Recommendation:** Skip separate calendar server if using **Stalwart** (built-in). Otherwise use **Baikal** for better admin UX and proven Docker support. **Radicale** if minimizing dependencies (Python vs PHP).

**Skip if:** Using Stalwart (built-in CalDAV/CardDAV).

### Webmail Client (Optional)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **SnappyMail** | Latest | Modern lightweight webmail | 138KB download (Brotli), 99% Lighthouse score, no database required, actively maintained RainLoop fork, significantly faster than Roundcube | MEDIUM-HIGH |
| **Roundcube** | Latest | Traditional PHP webmail | Most mature, extensive plugin ecosystem, fpm-alpine variant available, 58x more popular than SnappyMail, but heavier | HIGH |

**Recommendation:** Use **SnappyMail** for modern, minimal deployment. Dramatically faster, no database, minimal resource usage. Use **Roundcube** only if need specific plugins or organizational familiarity.

**Why SnappyMail over Roundcube:**
- 138KB vs multi-MB page loads
- No database required (simpler deployment)
- Better mobile experience
- Actively maintained (Roundcube slower development)

**Why Roundcube over SnappyMail:**
- Massive plugin ecosystem
- 20+ years of production use
- More organizational adoption
- Better documentation

---

## Transport Security Stack

### Encrypted Relay ↔ Home Transport

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **WireGuard (Kernel)** | Kernel module | VPN tunnel cloud ↔ home | Lowest latency, best throughput on ARM64, ~500% faster than userspace on Ethernet, kernel merged since Linux 5.6, standard WireGuard tools | HIGH |
| **mTLS over SMTP** | TLS 1.3 | Application-layer mutual auth | No separate tunnel, SMTP-native, works with Postfix relay_clientcerts, simpler architecture, but requires certificate distribution | MEDIUM-HIGH |

**Recommendation:** Use **WireGuard kernel module** for primary transport. Proven performance advantage on ARM64 (Nord Security study: kernel dramatically faster than userspace on RPi4). mTLS over SMTP as fallback/alternative for users unable to run kernel modules.

**Implementation:** WireGuard kernel module + custom Go relay daemon using emersion/go-smtp for SMTP protocol handling.

**Why WireGuard kernel over userspace:**
- 500% throughput improvement on ARM64 Ethernet (Nord Security testing)
- Lower power consumption (critical for RPi4)
- Better latency under load
- Kernel module standard since 5.6, widely available

**Why WireGuard over pure mTLS:**
- Simpler than certificate distribution to Postfix
- Tunnel isolates all relay ↔ home traffic (not just SMTP)
- Better for future multi-protocol support
- Easier NAT traversal

**Why mTLS as alternative:**
- No kernel dependency (works in restricted environments)
- SMTP-native (no separate tunnel)
- Standard Postfix configuration
- Better for environments blocking VPN ports

### Go WireGuard Libraries

| Library | Purpose | Why Recommended | Confidence |
|---------|---------|-----------------|------------|
| **golang.zx2c4.com/wireguard/wgctrl** | Kernel WireGuard control | Official Go bindings, standard for kernel module management | HIGH |
| **wireguard-go** | Userspace fallback | Official userspace implementation, use only when kernel unavailable | HIGH |

**Recommendation:** Use **wgctrl** for kernel control. Only use **wireguard-go** userspace as fallback for environments without kernel module access.

---

## Certificate Management Stack

### Public Certificates (Internet-Facing Relay)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **Certbot** | Latest | Let's Encrypt ACME client | Official EFF tool, Alpine Docker image available, standard automation, supports DNS-01 and HTTP-01 challenges | HIGH |
| **acme.sh** | Latest | Alternative ACME client | Lightweight shell script, broader DNS provider support, smaller footprint than Certbot | MEDIUM-HIGH |

**Recommendation:** Use **Certbot** in sidecar container. Standard choice, excellent documentation, official Docker image (certbot/certbot), proven DNS-01 automation for wildcard certs.

**DNS-01 challenge pattern:** CNAME _acme-challenge subdomain to validation-specific server (security best practice per Let's Encrypt docs).

### Internal CA (Relay ↔ Home Transport)

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| **step-ca** | 0.29.0 | Private certificate authority | Modern ACME server, short-lived certs with auto-renewal, SSH CA support, OAuth/OIDC integration, multiple database backends, REST API, Apache 2.0 license | HIGH |
| **cfssl** | Latest | CloudFlare's PKI toolkit | Simple JSON API, proven in production, but less active development than step-ca | MEDIUM |

**Recommendation:** Use **step-ca** for internal CA. Modern design, built-in ACME server (works with Certbot/Caddy), short-lived certificate best practices, active development, excellent documentation.

**Why step-ca over cfssl:**
- Built-in ACME server (automated renewal)
- Short-lived certificates (security best practice)
- Active development (SmallStep Labs)
- Better OAuth/SSO integration options
- Badger/BoltDB/Postgres/MySQL backends

**Why step-ca over manual OpenSSL:**
- Automated renewal eliminates manual process
- ACME protocol standardization
- REST API for automation
- Short-lived certs by default (better security)

### Go mTLS Libraries

| Library | Purpose | Why Recommended | Confidence |
|---------|---------|-----------------|------------|
| **crypto/tls (stdlib)** | Standard library TLS | Built-in, zero dependencies, RequireAndVerifyClientCert support, sufficient for mTLS | HIGH |
| **github.com/stephen-fox/mtls** | Certificate generation helper | Simplifies cert/key pair generation for testing/development | MEDIUM |

**Recommendation:** Use **crypto/tls** from Go standard library. Zero dependencies, well-tested, sufficient for production mTLS. Use mtls library only for development/testing certificate generation.

---

## Orchestration & Glue Code Stack

### Primary Language

| Language | Binary Size | ARM64 Support | SMTP Ecosystem | Why Recommended | Confidence |
|----------|-------------|---------------|----------------|-----------------|------------|
| **Go** | 2-5MB static | Excellent | Excellent (emersion/go-smtp) | Simple deployment, fast compilation, excellent stdlib, standard for cloud-native tools, mature SMTP libraries, great ARM64 cross-compilation | HIGH |
| **Rust** | 2-5MB static | Excellent | Growing (minismtp, Stalwart SMTP) | Smaller binaries with optimization, memory safety, but slower compilation, steeper learning curve | MEDIUM-HIGH |
| **Python** | N/A (interpreted) | Excellent | Excellent (aiosmtpd) | Rapid development, but requires runtime (~50MB+), slower performance, not ideal for minimal containers | MEDIUM |

**Recommendation:** Use **Go** for all orchestration and glue code. Single static binary deployment, excellent ARM64 cross-compilation, mature SMTP ecosystem (emersion/go-smtp), fast compilation, standard for cloud-native tools (similar to Docker, Kubernetes).

**Why Go over Rust:**
- Faster compilation (critical for CI/CD iteration)
- Simpler syntax (easier contributor onboarding)
- Larger SMTP library ecosystem
- Standard choice for cloud-native infrastructure
- Better error handling for network operations
- Comparable binary size (2-5MB range)

**Why Go over Python:**
- Static binary (no runtime dependency)
- ~10x smaller container footprint
- Better performance for network I/O
- Easier cross-compilation for ARM64

### Go SMTP Libraries

| Library | Version | Purpose | Why Recommended | Confidence |
|---------|---------|---------|-----------------|------------|
| **github.com/emersion/go-smtp** | 0.24.0 | SMTP client & server | Active development (1.1K projects depend on it), RFC 5321 compliant, ESMTP extensions, UTF-8 support, LMTP, MIT license, 2K+ stars | HIGH |
| **github.com/mhale/smtpd** | Latest | Minimal SMTP server | Simple API, but inactive (no updates in 12+ months), not recommended for new projects | LOW |

**Recommendation:** Use **emersion/go-smtp** exclusively. Active maintenance, comprehensive feature set, production-proven (1.1K dependents), RFC compliant, excellent for building custom relay logic.

---

## Build & CI/CD Stack

### Multi-Architecture Docker Builds

| Technology | Purpose | Why Recommended | Confidence |
|------------|---------|-----------------|------------|
| **docker/setup-qemu-action** | ARM64 emulation on amd64 runners | Standard GitHub Actions approach, works but slower (~3-5x) | HIGH |
| **Native ARM64 runners** | Build ARM64 on ARM64 hardware | Fastest (no emulation overhead), but requires ARM64 runner access (GitHub-hosted ARM64 runners or self-hosted) | HIGH |
| **docker/build-push-action** | Multi-platform builds | Official Docker action, supports buildx, push to registries | HIGH |

**Recommendation:** Use **QEMU emulation** for simplicity (no cost, standard GitHub Actions). Optimize to **native ARM64 runners** if build time becomes bottleneck. Use matrix strategy to build each platform separately and merge manifests.

**Pattern:**
```yaml
strategy:
  matrix:
    platform: [linux/amd64, linux/arm64]
```

Build each platform on dedicated runner (or QEMU), push by digest, merge job creates manifest list with `docker buildx imagetools create`.

### User-Selectable Components

**Build Strategy:** Use GitHub Actions build matrix with boolean inputs for component selection:

```yaml
on:
  workflow_dispatch:
    inputs:
      mail_server:
        type: choice
        options: [stalwart, maddy, postfix-dovecot]
      calendar_server:
        type: choice
        options: [none, radicale, baikal]
      webmail:
        type: choice
        options: [none, snappymail, roundcube]
```

Use Docker multi-stage builds with build args to select components. Each component as separate build stage, final image only includes selected components.

---

## Supporting Libraries (Go)

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| **golang.zx2c4.com/wireguard/wgctrl** | Latest | WireGuard kernel control | Managing WireGuard interfaces from Go | HIGH |
| **github.com/emersion/go-smtp** | 0.24.0 | SMTP protocol implementation | Custom relay logic, SMTP server/client | HIGH |
| **crypto/tls** (stdlib) | stdlib | TLS/mTLS implementation | Secure connections, certificate handling | HIGH |
| **github.com/spf13/cobra** | Latest | CLI framework | Building user-facing CLI tools | HIGH |
| **github.com/spf13/viper** | Latest | Configuration management | Loading config from files/env/flags | HIGH |

---

## Installation & Quickstart

### Cloud Relay Container (Postfix + Go Glue)

**Dockerfile pattern:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay ./cmd/relay

FROM alpine:3.21
RUN apk add --no-cache postfix ca-certificates wireguard-tools
COPY --from=builder /build/relay /usr/local/bin/relay
COPY postfix-config/ /etc/postfix/
EXPOSE 25
CMD ["/usr/local/bin/relay"]
```

**Result:** ~25-30MB container (Alpine 5MB + Postfix 15MB + WireGuard tools 5MB + Go binary 2-5MB).

### Home Server Container (Stalwart)

**Dockerfile pattern:**
```dockerfile
FROM stalwartlabs/stalwart:0.15.4
# Stalwart single binary, ~50MB total
# Add WireGuard for relay connection
RUN apt-get update && apt-get install -y wireguard-tools && rm -rf /var/lib/apt/lists/*
```

**Result:** ~70MB container (Stalwart includes all mail + calendar/contacts).

### Development Dependencies

```bash
# Go development
go install github.com/emersion/go-smtp@latest
go install golang.zx2c4.com/wireguard/wgctrl@latest

# Container building
docker buildx create --use

# Certificate management (local testing)
brew install step  # or: wget https://dl.step.sm/gh-release/cli/docs-ca-install/v0.29.0/step_linux_0.29.0_amd64.tar.gz
```

---

## Alternatives Considered

| Category | Recommended | Alternative | When to Use Alternative | Confidence |
|----------|-------------|-------------|------------------------|------------|
| Cloud Relay | Postfix | Haraka | Need custom JavaScript plugins, async event processing | MEDIUM |
| Cloud Relay | Postfix | Maddy | Prefer Go-native stack, comfortable with less battle-tested option | MEDIUM-HIGH |
| Home Mail Server | Stalwart | Maddy | Tighter memory constraints (<50MB), don't need CalDAV/CardDAV | HIGH |
| Home Mail Server | Stalwart | Postfix+Dovecot | Maximum compatibility, existing expertise, complex routing rules | HIGH |
| Calendar/Contacts | Baikal (if not Stalwart) | Radicale | Minimizing dependencies (Python vs PHP), file-based storage preference | MEDIUM-HIGH |
| Webmail | SnappyMail | Roundcube | Need specific plugins, organizational standard, more conservative choice | HIGH |
| Transport | WireGuard Kernel | mTLS-only | Kernel module unavailable, restricted environment, VPN ports blocked | MEDIUM-HIGH |
| CA | step-ca | cfssl | Simpler JSON API preference, less feature-rich CA sufficient | MEDIUM |
| Language | Go | Rust | Team Rust expertise, willing to accept slower compile times for memory safety | MEDIUM-HIGH |
| Base Image | Alpine | Distroless | Maximum security posture, comfortable with no-shell debugging constraints | MEDIUM-HIGH |

---

## What NOT to Use

| Avoid | Why | Use Instead | Confidence |
|-------|-----|-------------|------------|
| **Exim** | Complex configuration, less modern, smaller community than Postfix | Postfix | HIGH |
| **Sendmail** | Ancient, notorious configuration complexity, security history | Postfix | HIGH |
| **RainLoop (original)** | Abandoned project (forked to SnappyMail), security concerns | SnappyMail | HIGH |
| **Mail-in-a-Box** | Opinionated all-in-one with Ubuntu dependency, not containerizable, forces specific stack choices | Stalwart or Maddy | MEDIUM-HIGH |
| **BoringTun (Rust WireGuard)** | Userspace implementation, slower than kernel on ARM64, use only if kernel unavailable | WireGuard kernel module | HIGH |
| **wireguard-go (userspace)** | 3-5x slower than kernel, higher CPU usage, acceptable only as fallback | WireGuard kernel module | HIGH |
| **BerkleyDB** | Deprecated in Alpine (Oracle license change to AGPL-3.0), replaced with LMDB | LMDB (default in Alpine Postfix) | HIGH |
| **Python for relay glue** | Requires 50MB+ runtime, slower, larger containers | Go | HIGH |
| **Node.js for relay glue** | Requires 50MB+ runtime, async complexity for simple relay logic | Go | HIGH |

---

## Stack Patterns by Deployment

### Pattern 1: Minimal (Best for RPi4, 2GB RAM)
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Maddy (15MB binary) + SnappyMail (no DB) = ~40MB total
**Transport:** WireGuard kernel module
**Total footprint:** ~70MB containers + ~15MB RAM (Maddy)

### Pattern 2: Modern (Recommended for RPi4 4GB+)
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Stalwart 0.15.4 (70MB) with built-in CalDAV/CardDAV
**Transport:** WireGuard kernel module
**Total footprint:** ~100MB containers + ~50MB RAM (Stalwart)

### Pattern 3: Maximum Compatibility
**Cloud:** Postfix relay (25MB) + Go glue (5MB) = 30MB total
**Home:** Postfix + Dovecot (~80MB) + Baikal (~30MB) + Roundcube (~40MB) = 150MB
**Transport:** mTLS over SMTP (no WireGuard)
**Total footprint:** ~180MB containers + ~100MB RAM

---

## Version Compatibility Notes

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Postfix 3.7+ | Alpine 3.20+ | Requires LMDB (BerkleyDB deprecated) |
| Stalwart 0.15.4 | Pre-v1.0 schema | Breaking changes from 0.14.x, read upgrade docs |
| step-ca 0.29.0 | Certbot latest | Full ACME server compatibility |
| WireGuard kernel | Linux 5.6+ | Kernel module mainlined, standard in modern kernels |
| emersion/go-smtp 0.24.0 | Go 1.20+ | Requires generics support |

---

## Sources

### Official Documentation (HIGH Confidence)
- [Stalwart Mail Server](https://stalw.art/mail-server/) - Official docs and latest release (v0.15.4)
- [Stalwart GitHub Releases](https://github.com/stalwartlabs/stalwart/releases) - Version 0.15.4, 2026-01-19
- [Maddy Mail Server](https://maddy.email/) - Official documentation
- [Maddy GitHub](https://github.com/foxcpp/maddy) - Version 0.8.2, 2026-01-14
- [emersion/go-smtp GitHub](https://github.com/emersion/go-smtp) - Version 0.24.0, 2025-08-05
- [step-ca GitHub](https://github.com/smallstep/certificates) - Version 0.29.0, 2025-12-03
- [Docker Multi-Platform Builds Official Docs](https://docs.docker.com/build/ci/github-actions/multi-platform/)
- [Let's Encrypt Challenge Types](https://letsencrypt.org/docs/challenge-types/) - ACME DNS-01 documentation
- [WireGuard Official](https://www.wireguard.com/) - Kernel module documentation

### Docker Hub & Container Images (HIGH Confidence)
- [Postfix Alpine Containers](https://github.com/bokysan/docker-postfix) - Multi-arch Alpine/Debian/Ubuntu
- [Stalwart Docker Hub](https://hub.docker.com/r/stalwartlabs/stalwart) - Official multi-arch images
- [Maddy Docker Hub](https://hub.docker.com/r/foxcpp/maddy) - Official image
- [Certbot Docker Hub](https://hub.docker.com/r/certbot/certbot) - Official EFF image
- [step-ca Docker Hub](https://hub.docker.com/r/smallstep/step-ca/) - Official Smallstep image
- [Roundcube Docker Hub](https://hub.docker.com/r/roundcube/roundcubemail/) - Official fpm-alpine variant
- [Baikal Docker Hub](https://hub.docker.com/r/ckulka/baikal) - Multi-arch support

### Community & Comparisons (MEDIUM Confidence)
- [Alpine, Distroless, or Scratch Comparison (Medium)](https://medium.com/@cloudwithusama/alpine-distroless-or-scratch-choosing-the-right-lightweight-base-image-f5b12dc5d4f6) - 2026 container base image comparison
- [Docker Image Size Reduction (OneUpTime Blog)](https://oneuptime.com/blog/post/2026-01-16-docker-reduce-image-size/view) - Alpine vs Distroless vs Scratch
- [WireGuard Kernel vs Userspace Performance (Nord Security)](https://nordsecurity.com/blog/wireguard-kernel-module-vs-user-space) - ARM64 benchmarks on Raspberry Pi 4
- [SnappyMail vs Roundcube (Forward Email Blog)](https://forwardemail.net/en/blog/open-source/webmail-email-clients) - 2026 webmail comparison
- [Go vs Rust in 2026 (Bitfield Consulting)](https://bitfieldconsulting.com/posts/rust-vs-go)
- [Building Multi-Platform Docker Images (Blacksmith)](https://www.blacksmith.sh/blog/building-multi-platform-docker-images-for-arm64-in-github-actions)
- [GotaTun Rust WireGuard (Mullvad)](https://mullvad.net/en/blog/announcing-gotatun-the-future-of-wireguard-at-mullvad-vpn) - 2026 Rust WireGuard developments

### GitHub & Library Documentation (MEDIUM-HIGH Confidence)
- [Haraka SMTP Server GitHub](https://github.com/haraka/Haraka) - Node.js SMTP relay
- [Radicale GitHub](https://github.com/Kozea/Radicale) - CalDAV/CardDAV Python server
- [Baikal GitHub](https://github.com/sabre-io/Baikal) - CalDAV/CardDAV PHP server
- [minismtp Rust Library](https://github.com/saefstroem/minismtp) - Minimal Rust SMTP server
- [mhale/smtpd Go Library](https://github.com/mhale/smtpd) - Inactive, not recommended
- [acme-dns GitHub](https://github.com/joohoi/acme-dns) - DNS-01 challenge automation

---

*Stack research for: DarkPipe Privacy Email*
*Researched: 2026-02-08*
*Researcher: GSD Project Researcher*
