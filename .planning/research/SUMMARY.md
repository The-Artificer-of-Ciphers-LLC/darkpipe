# Project Research Summary

**Project:** DarkPipe
**Domain:** Privacy-First Self-Hosted Email System
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Executive Summary

DarkPipe is a privacy-focused self-hosted email system that solves the fundamental problem preventing home-based email: residential ISPs block port 25 and blacklist dynamic IPs. The solution is a cloud relay + home device split architecture where a minimal cloud VPS (25-30MB container) receives internet mail and forwards it through an encrypted WireGuard tunnel to a home device (Raspberry Pi 4 or similar) that stores all email data. This architecture ensures email data never persists in the cloud while maintaining internet-standard deliverability.

The recommended stack balances privacy, performance, and resource constraints: Postfix for the proven minimal cloud relay, Stalwart or Maddy for the modern home mail server (single-binary deployment), WireGuard kernel module for transport (500% faster than userspace on ARM64), and Go for orchestration glue code. The key differentiator is GitHub Actions-driven customizable builds - users select their stack components (mail server, webmail, calendar) via workflow configuration, and the system generates multi-architecture Docker images. This eliminates the "one size fits all" limitation of competitors like Mail-in-a-Box and Mailcow.

Critical risks center on email deliverability (IP reputation, DNS authentication, VPS provider port 25 restrictions) and operational complexity (self-hosted email requires ongoing maintenance that drives 80%+ user abandonment). Mitigation requires careful VPS provider selection (Linode/OVH have port 25 open by default), mandatory 4-6 week IP warmup period, automated DNS validation, and exceptional UX to reduce setup/maintenance burden below competitor levels.

## Key Findings

### Recommended Stack

DarkPipe requires a split-stack architecture optimized for different deployment targets: cloud relay (minimal container on amd64 VPS) and home device (full-featured on ARM64 or amd64).

**Core technologies:**
- **Postfix (Alpine)** for cloud relay - Battle-tested SMTP relay in 15-30MB container, proven minimal footprint, decades of production hardening
- **Stalwart or Maddy** for home mail server - Single-binary deployment (Stalwart 70MB with built-in CalDAV/CardDAV, Maddy 15MB for minimal footprint)
- **WireGuard kernel module** for transport - 500% faster than userspace on ARM64 Ethernet per Nord Security benchmarks, kernel module standard since Linux 5.6
- **Go** for orchestration/glue code - 2-5MB static binaries, excellent ARM64 cross-compilation, mature SMTP libraries (emersion/go-smtp)
- **step-ca** for internal CA - Modern ACME server with short-lived certificates and auto-renewal
- **Alpine Linux** base image - 5MB base trades 3MB for operational benefits (shell, package manager, debugging tools)

**Critical version notes:**
- Stalwart 0.15.4 is pre-v1.0 (v1.0 expected Q2 2026 with stable schema)
- Postfix 3.7+ requires LMDB (BerkleyDB deprecated in Alpine)
- WireGuard kernel module requires Linux 5.6+ (standard in modern distributions)

### Expected Features

Email is uniquely complex among self-hosted services - requires perfect DNS, reputation management, anti-spam, and ongoing monitoring.

**Must have (table stakes):**
- Core SMTP/IMAP/TLS transport with port 587 submission
- SPF/DKIM/DMARC authentication (Gmail/Yahoo/Microsoft mandate since 2024-2025)
- Reverse DNS (PTR) record configuration (missing = instant spam folder)
- Basic webmail (Roundcube/SnappyMail) for non-technical users
- Multi-user/multi-domain support
- Spam filtering (Rspamd standard)
- Let's Encrypt auto-renewal (90-day certificates)

**Should have (competitive advantage):**
- **Cloud-fronted architecture** - DarkPipe's core value proposition, solves port 25 blocking
- **GitHub Actions customization** - User-selectable stack via templated workflows (unique vs competitors)
- **First-class ARM64 support** - Raspberry Pi as primary deployment target
- **DNS API integration** - Automates record creation for Cloudflare/Route53/etc.
- CalDAV/CardDAV server (required for Gmail replacement)
- MTA-STS + DANE for TLS enforcement
- Prometheus metrics export for homelab users

**Defer (v2+):**
- Queue encryption at rest (conflicts with spam scanning, high complexity)
- PGP/WKD support (niche user base, key management complexity)
- Full-text search with attachment indexing (resource-intensive on Pi)
- Audit logging for admin actions (enterprise feature)

**Key differentiator vs competitors:** Mail-in-a-Box, Mailu, docker-mailserver, and Mailcow all assume public IP and full VPS deployment. DarkPipe is the only solution architected for home device storage with cloud relay for internet-facing SMTP. Additionally, competitors offer opinionated stacks while DarkPipe provides user customization via GitHub Actions.

### Architecture Approach

Cloud relay is minimal, stateless, and ephemeral. Home device is full-featured with persistent storage. Transport layer (WireGuard or mTLS) provides secure, NAT-traversing connection between them.

**Major components:**

1. **Cloud Relay (SMTP Gateway)** - Minimal Postfix in relay-only mode (25-30MB container), ephemeral RAM queue with optional encrypted S3 overflow, Let's Encrypt certificates, no persistent mail storage
2. **Transport Layer** - WireGuard hub-and-spoke tunnel (primary) with persistent keepalive for NAT traversal, or mTLS persistent connection (fallback for restricted networks)
3. **Home Device Mail Stack** - Full mail server (Postfix+Dovecot or Stalwart/Maddy), optional CalDAV/CardDAV (Radicale/Baikal or built into Stalwart), optional webmail (Roundcube/SnappyMail), all storage in Docker volumes
4. **Build System** - GitHub Actions with matrix builds for component selection and multi-arch images (arm64/amd64)
5. **DNS Automation** - Record generation, provider API integration, pre-deployment validation

**Critical architectural patterns:**
- Ephemeral cloud relay with persistent home storage (privacy requirement)
- WireGuard hub-and-spoke with home device as spoke (NAT-friendly)
- Single container vs Docker Compose stack options (resource flexibility)
- Build-time configuration via GitHub Actions (eliminates runtime complexity)

**Data flow:**
- Inbound: Internet SMTP → Cloud Relay → WireGuard → Home Device → User IMAP client
- Outbound: User SMTP → Home Device → WireGuard → Cloud Relay → Internet SMTP

### Critical Pitfalls

Research identified 11 critical pitfalls with HIGH confidence. Top 5 that shape roadmap:

1. **VPS provider port 25 restrictions** - DigitalOcean/Hetzner/Vultr block SMTP by default. Solution: Choose Linode/OVH for v1 (open by default), document provider policies, budget time for unblocking requests. **Must address in Phase 0 (Infrastructure Selection).**

2. **New VPS IPs start blacklisted or zero reputation** - Fresh IPs may be on RBL blacklists from previous tenants, or have no sending history causing Gmail/Outlook rejection. Solution: Check IP reputation before launch (MXToolbox/Spamhaus), mandatory 4-6 week warmup schedule (start 2-5 emails/day, gradually increase), continuous blacklist monitoring. **Extends MVP timeline by 4-6 weeks.**

3. **Missing/misconfigured SPF/DKIM/DMARC breaks deliverability** - Even one-character DNS typo causes authentication failures. Solution: Use 2048-bit DKIM keys (1024-bit weak), start DMARC with p=none for monitoring, automated testing with mail-tester.com (must score 9+/10). **Must address in Phase 1 (MVP).**

4. **Missing PTR record triggers instant spam filtering** - Reverse DNS controlled by VPS provider, not DNS registrar. Solution: Request PTR from provider before sending any email, verify forward-confirmed reverse DNS, include in deployment checklist. **Must address in Phase 1 (MVP).**

5. **WireGuard tunnel fails after home internet drop** - Tunnel doesn't auto-reconnect after ISP outages or dynamic IP changes. Solution: PersistentKeepalive=25 setting, systemd Restart=on-failure, use IP addresses or public DNS (not local DNS), monitor handshake timestamp. **Must address in Phase 1 (MVP).**

**Additional critical pitfalls:**
- Residential/dynamic IP from home device gets blacklisted (architecture prevents this by design - home never sends direct SMTP)
- TLS certificate expiration breaks email silently (automated renewal + monitoring required)
- Becoming an open relay leads to immediate blacklisting (require SASL auth, test with MXToolbox)
- Docker volume management loses mail data on updates (use named volumes, test restore)
- Raspberry Pi ARM64 runs out of memory under load (don't run ClamAV/SpamAssassin on Pi, filter on cloud relay)
- Users give up due to complex setup (UX must exceed competitors, automate DNS/cert management)

## Implications for Roadmap

Based on dependency analysis and pitfall research, suggested phase structure:

### Phase 0: Infrastructure Selection & Validation
**Rationale:** VPS provider port 25 restrictions are absolute blockers. Must research and validate provider before any development. IP reputation baseline must be established.

**Delivers:**
- VPS provider compatibility list (port 25 open: Linode, OVH; unblocking available: BuyVM, Vultr)
- IP reputation validation checklist (MXToolbox, Spamhaus, multi-RBL)
- Provider selection guide for users

**Addresses:** Pitfall #1 (port 25 restrictions), Pitfall #2 (IP blacklisting)

**Research needed:** None - provider policies change frequently, maintain living document

### Phase 1: Foundation (Configuration & Transport)
**Rationale:** Configuration schema drives all components. Transport layer (WireGuard) is architectural requirement that both cloud relay and home device depend on.

**Delivers:**
- Configuration schema (darkpipe.yaml) and validation
- WireGuard tunnel setup with NAT traversal and auto-reconnection
- DNS automation library (record generation, validation, provider APIs)

**Uses:** Go for config/DNS tools, WireGuard kernel module

**Addresses:** Architecture foundation, Pitfall #8 (tunnel reconnection)

**Research needed:** LOW - WireGuard well-documented, Go stdlib sufficient

### Phase 2: Cloud Relay (Minimal Gateway)
**Rationale:** Cloud relay is simpler than home device (relay-only vs full mail server). Depends on transport layer from Phase 1.

**Delivers:**
- Postfix relay container (Alpine-based, 25-30MB)
- Ephemeral RAM queue with configurable overflow
- Let's Encrypt/Certbot automation
- Health checks and basic monitoring

**Uses:** Postfix 3.7+, Alpine Linux, Certbot, WireGuard client

**Addresses:** Pitfall #6 (certificate expiration), Pitfall #7 (open relay prevention)

**Research needed:** MEDIUM - S3 overflow queue pattern needs validation

### Phase 3: Home Device (Mail Storage)
**Rationale:** Home device provides persistent storage and full mail server functionality. Depends on transport layer and benefits from cloud relay for testing.

**Delivers:**
- Single-container option (docker-mailserver extended or Stalwart/Maddy)
- Docker Compose alternative (Postfix + Dovecot + Radicale)
- Volume management and persistence
- Basic webmail integration (Roundcube or SnappyMail)

**Uses:** Stalwart/Maddy/Postfix+Dovecot (user-selectable), Dovecot, Radicale or Baikal

**Addresses:** Pitfall #9 (Docker volume data loss), Pitfall #10 (Pi memory limits)

**Research needed:** MEDIUM - CalDAV/CardDAV integration patterns, ARM64 performance testing

### Phase 4: DNS & Authentication (Deliverability)
**Rationale:** Email authentication (SPF/DKIM/DMARC) and DNS automation are critical for deliverability but depend on functional mail stack for testing.

**Delivers:**
- Automated SPF/DKIM/DMARC record generation
- DNS provider API integrations (Cloudflare, Route53)
- Pre-deployment validation (mail-tester.com integration)
- PTR record verification and documentation

**Uses:** DNS automation library from Phase 1, OpenDKIM or built-in signing

**Addresses:** Pitfall #3 (SPF/DKIM/DMARC), Pitfall #4 (PTR records)

**Research needed:** LOW - Standards well-documented, provider APIs mature

### Phase 5: Build System (User Customization)
**Rationale:** GitHub Actions builds enable DarkPipe's key differentiator (user-selectable components) but require complete stack to be implemented first.

**Delivers:**
- Multi-arch build workflows (arm64, amd64)
- Component selection inputs (mail server, webmail, calendar)
- Build matrix for parallel builds
- Image publishing to GHCR

**Uses:** GitHub Actions, Docker buildx, QEMU for emulation

**Addresses:** Core differentiator vs competitors

**Research needed:** LOW - Multi-arch Docker builds well-documented

### Phase 6: Deployment & UX (Launch Readiness)
**Rationale:** Exceptional UX required to overcome "self-hosted email is too hard" barrier. All components must be functional for end-to-end testing.

**Delivers:**
- CLI wizard for initial setup
- Pre-flight checks (DNS, ports, IP reputation)
- Status dashboard (authentication, reputation, queue health)
- Plain-language error messages
- Zero-downtime update process

**Uses:** Go CLI (cobra/viper), web framework for dashboard

**Addresses:** Pitfall #11 (complex setup UX)

**Research needed:** MEDIUM - Update strategies need validation, UX testing required

### Phase 7: IP Warmup & Production Validation (4-6 weeks)
**Rationale:** IP warmup cannot be skipped or accelerated. This is a time-based phase, not development.

**Delivers:**
- IP warmup schedule execution (2-5 emails/day → 25-50/day over 4-6 weeks)
- Deliverability monitoring (bounce rates, spam folder placement)
- Blacklist monitoring automation
- DMARC report collection and analysis

**Uses:** Existing mail stack, external monitoring tools

**Addresses:** Pitfall #2 (IP reputation)

**Research needed:** None - warmup schedules standard, execution required

### Phase 8: Scaling & Operations (Post-Launch)
**Rationale:** Operational features that reduce ongoing maintenance burden. Can be added after MVP validates core concept.

**Delivers:**
- Automated blacklist monitoring with alerts
- Queue management and cleanup
- Certificate expiry monitoring
- Backup/restore automation
- Rate limiting and abuse prevention
- Prometheus metrics export

**Addresses:** Reduces ongoing maintenance burden (Pitfall #11)

**Research needed:** LOW - Standard patterns, needs integration testing

### Phase Ordering Rationale

**Why this order:**
- **Phase 0 before anything:** Provider restrictions are absolute blockers
- **Configuration/transport before implementations:** Shared dependencies
- **Cloud relay before home device:** Simpler component validates transport layer
- **DNS after functional stack:** Needs working mail server for testing
- **Build system after components:** Requires complete stack to parameterize
- **UX after functionality:** Can't polish non-existent features
- **IP warmup is time-based:** 4-6 weeks regardless of development speed
- **Operations last:** Nice-to-have after core validation

**How this avoids pitfalls:**
- Provider selection (Phase 0) prevents port 25 issues
- Transport resilience (Phase 1) prevents tunnel failures
- Volume management (Phase 3) prevents data loss
- DNS automation (Phase 4) prevents authentication failures
- UX focus (Phase 6) reduces abandonment
- IP warmup (Phase 7) ensures deliverability

**Why grouping makes sense:**
- Phase 1 groups foundational dependencies (config, transport, DNS library)
- Phases 2-3 group deployment targets (cloud vs home)
- Phase 4 groups deliverability requirements (DNS, auth)
- Phase 6 groups user-facing features (UX, dashboard)
- Phase 8 groups operational maturity features

### Research Flags

Phases likely needing deeper research during planning:

- **Phase 3 (Home Device):** CalDAV/CardDAV integration patterns unclear - Stalwart has built-in but Postfix+Dovecot requires separate service. ARM64 performance benchmarks needed for spam filtering decision.
- **Phase 6 (UX/Updates):** Zero-downtime update strategies for mail servers need validation - queue draining and mail-in-flight handling during container swap.
- **Phase 8 (Operations):** S3-compatible queue overflow pattern proven by Maddy/Stalwart but implementation details sparse.

Phases with standard patterns (skip research-phase):

- **Phase 1 (Configuration):** YAML schemas and Go CLI tooling well-documented
- **Phase 2 (Cloud Relay):** Postfix relay configuration extremely well-documented
- **Phase 4 (DNS):** Email authentication standards (SPF/DKIM/DMARC) mature with clear specs
- **Phase 5 (Build System):** GitHub Actions multi-arch builds have official docs and many examples
- **Phase 7 (IP Warmup):** Warmup schedules standardized across industry

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Core technologies (Postfix, WireGuard, Docker) verified with official documentation. Stalwart/Maddy recommended based on feature comparison and current versions. |
| Features | MEDIUM-HIGH | Table stakes validated against competitors and standards (SPF/DKIM/DMARC mandates). Differentiators based on competitor analysis. Anti-features informed by Helm failure analysis. |
| Architecture | MEDIUM | Core patterns (relay, WireGuard tunnel, Docker deployment) HIGH confidence. Advanced patterns (S3 overflow, CalDAV integration) MEDIUM - proven by other projects but implementation details need validation. |
| Pitfalls | HIGH | Critical pitfalls verified with official sources (Spamhaus, provider docs, Let's Encrypt). Community pitfalls (UX complexity, IP warmup) validated by multiple 2026 blog posts and user abandonment reports. |

**Overall confidence:** MEDIUM-HIGH

### Gaps to Address

Areas where research was inconclusive or needs validation during implementation:

- **CalDAV/CardDAV without separate service:** Stalwart provides built-in CalDAV/CardDAV but research found no evidence of Dovecot-native implementation. Need to decide: recommend Stalwart for integrated approach, or Postfix+Dovecot+Radicale for flexibility. Phase 3 planning should resolve this.

- **Queue overflow threshold calculation:** No authoritative source on when to trigger S3 overflow from RAM queue. Needs simulation/testing based on average email size and delivery latency. Phase 2 planning should establish thresholds.

- **Certificate rotation without downtime:** How to rotate mTLS certificates (if mTLS transport used) without breaking persistent connection. WireGuard doesn't have this issue (uses pre-shared keys). Phase 1 planning for mTLS fallback should investigate.

- **Multi-user resource budgeting on Pi:** Architecture assumes single user or small family (5-10 users). What are actual memory/CPU requirements per user on RPi4? Phase 3 needs empirical testing.

- **ARM64 spam filtering performance:** Should ClamAV/SpamAssassin run on cloud relay (defeats privacy model by scanning plaintext in cloud) or home device (may exceed Pi4 resources)? Phase 3 needs benchmarking to inform default configuration.

- **DNS propagation delays on initial deployment:** How to handle 24-48 hour DNS propagation when mail may arrive immediately after MX record creation? Does cloud relay need extended queue retention for first deployment? Phase 4 planning should establish strategy.

- **Backup strategy ownership:** Should DarkPipe provide automated backup (additional complexity) or document user's responsibility? Where should backups be stored (user's cloud account, local USB drive)? Phase 8 planning should decide scope.

## Sources

### Primary (HIGH confidence)

Research drew from official documentation for core technologies:

- **STACK.md sources:** Stalwart/Maddy/Postfix official docs, Docker multi-platform build documentation, WireGuard quickstart, step-ca GitHub, emersion/go-smtp library documentation
- **FEATURES.md sources:** Mail-in-a-Box/Mailu/Mailcow official sites and GitHub repositories, Helm shutdown FAQ and reviews, email authentication standards (SPF/DKIM/DMARC RFCs), CalDAV/CardDAV specifications
- **ARCHITECTURE.md sources:** Postfix configuration README, WireGuard documentation, Docker Compose documentation, docker-mailserver GitHub repository
- **PITFALLS.md sources:** VPS provider official SMTP policies (DigitalOcean, Linode, OVH, Hetzner, Vultr), Spamhaus PBL documentation, Let's Encrypt challenge types, email deliverability guides from major providers

### Secondary (MEDIUM confidence)

Community resources and comparative analyses:

- Alpine vs Distroless vs Scratch container base image comparisons (Medium, OneUptime blog)
- WireGuard kernel vs userspace performance benchmarks (Nord Security blog with RPi4 testing)
- SnappyMail vs Roundcube webmail comparison (Forward Email blog)
- Multi-architecture Docker image build guides (Blacksmith, Red Hat developer articles)
- Self-hosting email sustainability discussions (2026 blog posts on email self-hosting challenges)
- IP warmup best practices (Mailwarm, Iterable, Mailtrap guides)

### Tertiary (LOW confidence, needs validation)

- S3-compatible queue overflow pattern (proven by Maddy/Stalwart but implementation details sparse)
- CalDAV/CardDAV integration approaches (Dovecot mailing list discussions, community examples)
- Raspberry Pi resource constraints under mail server load (forum discussions, community reports)
- Zero-downtime Docker update strategies for mail servers (generic Docker guides, needs mail-specific validation)

---
*Research completed: 2026-02-08*
*Ready for roadmap: yes*
