# Feature Research

**Domain:** Privacy-First Self-Hosted Email
**Researched:** 2026-02-08
**Confidence:** MEDIUM-HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

#### Core Email Transport (SMTP/IMAP)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Inbound SMTP relay | Fundamental email receiving | Low | Standard Postfix/mail server functionality |
| Outbound SMTP relay | Fundamental email sending | Low | Required for any email system |
| IMAP server | Standard protocol for modern email clients | Medium | Dovecot is standard; IMAP vastly better UX than POP3 |
| SMTP Submission (port 587) | Modern email sending standard | Low | Required by all clients; port 25 blocked by ISPs |
| TLS for SMTP/IMAP | Encryption in transit | Low | Let's Encrypt makes this trivial; users expect https:// everywhere |
| STARTTLS support | Opportunistic encryption | Low | Standard for SMTP; implicit in modern stacks |

#### Email Authentication & Deliverability

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| SPF record generation | Gmail/Yahoo/MS mandate for bulk senders | Low | DNS TXT record; trivial if automated |
| DKIM signing | Industry baseline for email authentication | Medium | Requires key generation, DNS publication, Postfix/OpenDKIM config |
| DMARC policy setup | Mandated by major providers in 2024-2025; baseline in 2026 | Low | DNS TXT record; policy must align with SPF/DKIM |
| Reverse DNS (PTR) setup | Deliverability requirement; missing = instant spam folder | Low | Cloud relay must handle this; user device cannot |
| MX record configuration | Email routing to your server | Low | DNS configuration; expected to be automated |

#### Basic Mail Server Features

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Multiple mailboxes/users | Multi-user support is assumed | Low | Postfix/Dovecot handle this natively |
| Multiple domains | Many users consolidate multiple identities | Medium | Postfix virtual domains; DNS delegation per-domain |
| Mail aliases | Forwarding/catch-all addresses | Low | Standard Postfix virtual alias maps |
| Folder management | IMAP folders/labels | Low | Dovecot handles this; client-driven |
| Basic search | Finding emails in mailbox | Low | Dovecot FTS (full-text search) with Xapian or Solr |
| Spam filtering | Inbound spam protection | Medium | SpamAssassin/Rspamd standard; auto-learning improves accuracy |
| Greylisting | Reduces spam via temporary deferrals | Low | Postgrey or Rspamd greylisting module |
| Virus scanning | ClamAV integration expected | Low | Performance overhead; some users disable for personal use |

#### Webmail Basics

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Web-based email access | Not everyone wants native clients | Medium | Roundcube/SOGo are standard options |
| Email composition | Writing/sending via web | Low | Built into webmail solutions |
| Attachment handling | Drag-drop upload, download | Low | Modern webmail (Roundcube 1.5+) supports drag-drop |
| Mobile-responsive UI | Mobile devices dominate email access | Medium | Roundcube/SOGo responsive by default; verify on actual devices |
| HTML email rendering | Modern emails are HTML-heavy | Low | Built into webmail clients |
| Basic contacts integration | Accessing address book from compose | Medium | Requires CardDAV integration or webmail-native contacts |

#### Security Fundamentals

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| TLS certificate auto-renewal | Let's Encrypt requires 90-day renewal | Low | Certbot handles this; failure = service outage |
| Fail2ban or equivalent | Brute force protection | Low | Standard hardening; watches auth logs and bans IPs |
| Firewall rules | Only expose required ports | Low | ufw/iptables; 25, 587, 993, 443, 22 (if SSH) |

#### Basic Administration

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Web admin panel | GUI for non-technical users | Medium | Mail-in-a-Box, Mailu, Mailcow all provide this |
| Add/remove mailboxes | User management | Low | Admin panel or CLI |
| Quota management | Disk space limits per user | Low | Dovecot quota plugin |
| Backup configuration | Automated backups expected | Medium | Mail-in-a-Box uses Duplicity to S3; critical for disaster recovery |

### Differentiators (Competitive Advantage)

Features that set product apart. Not expected, but valued.

#### Privacy & Control Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Cloud-fronted architecture | Internet-facing SMTP without exposing home IP | High | **DarkPipe's core value prop**: cloud relay + home device split |
| User-owned hardware for storage | Email stored on user's device, not cloud provider | Medium | Differentiates from hosted solutions; requires secure transport |
| No vendor lock-in for storage | Users control their own hardware/data | Low | Natural consequence of architecture |
| Transparent relay operation | Users see what cloud relay does (audit logs) | Medium | Builds trust; critical for privacy-focused users |
| Minimal cloud footprint | Cloud relay doesn't store mail (forward-only) | Medium | Reduces privacy exposure; requires reliable home device |
| Data sovereignty | Email never leaves user's jurisdiction | Low | Marketing/trust angle; practical for some use cases |

#### TLS Enforcement & Security Visibility

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| MTA-STS policy publication | Modern TLS enforcement standard | Low | DNS + HTTPS-hosted policy file; Mailu/Mailcow have this |
| DANE TLSA records | DNSSEC-based certificate verification | Medium | Requires DNSSEC-enabled DNS provider; high security users want this |
| TLS-RPT (TLS reporting) | Visibility into TLS failures with peers | Low | DNS TXT record + report receiver |
| Strict mode: refuse plaintext peers | Reject mail from servers without TLS | Medium | High privacy users want this; may break mail from legacy systems |
| Notification of insecure peers | Alert when peer doesn't support TLS | Medium | Transparency into security posture; rspamd can log this |
| Queue encryption at rest | Encrypt mail spool on disk | High | **Rarely implemented**; requires content filter or filesystem encryption |

#### DNS Automation & Deliverability

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Auto-generate DNS records | Eliminates manual DNS configuration errors | Medium | Mail-in-a-Box does this if it's your nameserver; DarkPipe likely needs instructions |
| DNS validation checker | Verify SPF/DKIM/DMARC/MX setup | Low | Mail-in-a-Box includes this; critical for deliverability troubleshooting |
| DNS API integration | Programmatic DNS updates (Cloudflare, Route53, etc.) | Medium | Enables true one-click setup for supported providers |
| DNS record templates | Copy-paste instructions for manual setup | Low | Fallback for unsupported DNS providers |

#### Calendar & Contacts (Groupware)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| CalDAV server | Standards-based calendar sync | Medium | Nextcloud, Baïkal, SOGo, Radicale; required for full Gmail replacement |
| CardDAV server | Standards-based contacts sync | Medium | Same implementations as CalDAV; often bundled |
| Shared calendars | Family/team calendar sharing | Medium | SOGo supports this well; Nextcloud too |
| Calendar web UI | View/edit calendars without client | Medium | SOGo provides this; Nextcloud Calendar app |
| Contacts web UI | Manage contacts without client | Low | SOGo, Nextcloud Contacts |

#### Advanced Mail Features

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Sieve filtering rules | Server-side mail rules (folder sorting, auto-reply) | Medium | ManageSieve protocol; Roundcube plugin for GUI management |
| Vacation/auto-reply | Out-of-office messages | Low | Sieve script or Dovecot pigeonhole plugin |
| Mail forwarding rules | Forward to external addresses | Low | Postfix aliases or Sieve |
| Catch-all addresses | domain@example.com forwards all unknown addresses | Low | Postfix virtual alias wildcard |
| Full-text search with attachments | Search email body + PDF/DOCX content | High | Dovecot FTS with Apache Tika for attachment indexing |

#### Build/Deployment Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| GitHub Actions build customization | Users choose their stack via workflow config | High | **DarkPipe's unique approach**: templated workflows, user customizes |
| Multi-architecture images | ARM64 (Raspberry Pi, Mac) + x86_64 support | Medium | Docker buildx with `--platform linux/amd64,linux/arm64` |
| One-click deploy templates | Deploy to cloud providers via marketplace | High | DigitalOcean App Platform, AWS Lightsail, etc.; high initial effort |
| Documented stack alternatives | Choose Postfix vs Stalwart, Roundcube vs SOGo | Medium | Flexibility vs "opinionated simplicity" tradeoff |
| Reproducible builds | GitOps for email server config | Medium | Natural with GitHub Actions; versioned infrastructure |

#### Monitoring & Observability

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Mail delivery status dashboard | See which emails sent/received, delivery state | Medium | Postfix logs + parsing; Mailcow has queue viewer |
| Queue health monitoring | Detect stuck/deferred mail | Low | Postfix queue monitoring (mailq); alert on buildup |
| Certificate expiry alerts | Proactive notification before Let's Encrypt expires | Low | Simple cron job + email; critical to avoid outages |
| SMTP/IMAP connection logs | Audit who's connecting, from where | Low | Standard Postfix/Dovecot logs; privacy users want visibility |
| Prometheus metrics export | Integration with monitoring stack | Medium | Postfix exporter + Dovecot exporter exist; homelab users want this |
| Deliverability scoring | Check reputation of your sending IP | Medium | Integration with mail-tester.com or similar APIs |

#### Advanced Security Features

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| PGP/WKD support | End-to-end encryption for power users | Medium | Stalwart has this; Web Key Directory for automatic key discovery |
| S/MIME support | Certificate-based email signing/encryption | Medium | Standards-based alternative to PGP; enterprise users may want this |
| 2FA for webmail/admin | Protect web interfaces | Low | TOTP via Google Authenticator; Mail-in-a-Box includes this |
| Audit logging | Detailed logs of admin actions | Medium | Mailcow has this; critical for multi-admin environments |
| Rate limiting outbound mail | Prevent account takeover abuse | Low | Postfix policyd-rate-limit or Rspamd ratelimit module |

### Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Built-in webmail from scratch | Massive undertaking; solved problem | Use Roundcube or SOGo; focus on integration, not reinvention |
| Custom email client | Out of scope; existing clients work | Support standard protocols (IMAP, SMTP, CalDAV, CardDAV) |
| AI-powered features (spam, categorization) | Adds complexity, dependencies, privacy concerns | Use proven tools (Rspamd has Bayes learning); let users opt into external AI |
| Built-in VPN/Tor routing | Scope creep; separate concern | Document how to run behind VPN if desired; don't bundle |
| Blockchain-based identity | Complexity, immaturity, user confusion | Use standard DNS + TLS certificates |
| Built-in file sharing (Nextcloud-style) | Massive scope; solved problem | CalDAV/CardDAV only; if users want files, they install Nextcloud separately |
| Real-time collaborative editing | Wrong product category | Email is asynchronous; leave collaboration to Nextcloud/office suites |
| Built-in chat/messaging | Different problem domain | Email server, not Signal/Matrix |
| Support for legacy protocols (POP3, insecure SMTP) | Security liability; 2026 baseline is IMAP+TLS | IMAP only; refuse plaintext auth |
| Automatic migration from Gmail/Outlook | High complexity, API changes break it | Provide clear manual migration guide; let users use Thunderbird tools |
| Native mobile apps | Requires ongoing maintenance for iOS/Android | Users use native Mail apps or K-9/FairEmail; support standard protocols |
| Self-hosted DNS server | Increases attack surface, complexity | Use existing DNS providers with API integration or manual setup |
| Catch-all relay (forward all mail to Gmail) | Defeats privacy purpose | If users want Gmail as backup, they can configure forwarding manually |
| Machine learning spam training UI | Users won't use it; Bayesian auto-learn works | Rspamd auto-learns from user actions (move to spam folder) |
| Multi-tenant SaaS mode | DarkPipe is self-hosted; multi-tenancy adds vast complexity | Single-user or family-scale; if users want SaaS, recommend hosted providers |
| Windows/macOS native server support | Linux is email server standard; porting is massive effort | Document Docker Desktop for local testing; production = Linux VPS + home Linux |

## Feature Dependencies

```
Email Fundamentals
├── SMTP Inbound/Outbound → Required for everything
├── IMAP → Required for webmail, clients
└── TLS Certificates → Required for SMTP/IMAP encryption, HTTPS

Email Authentication (Deliverability)
├── SPF → Required before DMARC
├── DKIM → Required before DMARC
├── DMARC → Requires SPF + DKIM
├── MTA-STS → Requires TLS certs + HTTPS hosting
└── DANE TLSA → Requires DNSSEC-enabled DNS

DNS Automation
├── DNS API Integration → Enables auto-setup of SPF/DKIM/DMARC/MTA-STS
└── DNS Validation → Requires readable DNS records (after setup)

Webmail
├── Web Server (Nginx/Apache) → Hosts webmail interface
├── IMAP Server → Webmail backend
├── TLS Certificates → HTTPS for webmail
└── Contacts Integration → Requires CardDAV server

Groupware (Calendar/Contacts)
├── CalDAV Server → Calendar sync
├── CardDAV Server → Contacts sync
└── Web Server → Optional web UI for calendar/contacts

Advanced Mail Features
├── Sieve Filtering → Requires Dovecot + Pigeonhole
├── ManageSieve → Requires Sieve + protocol server for remote management
└── Full-Text Search → Requires Dovecot FTS plugin + Xapian/Solr

Security Features
├── Queue Encryption → Requires content filter or filesystem encryption
├── Strict TLS Mode → Requires MTA-STS or custom Postfix policy
├── PGP/WKD → Requires key management + WKD hosting
└── 2FA → Requires TOTP library + session management

Monitoring
├── Prometheus Metrics → Requires exporters for Postfix/Dovecot
├── Delivery Status Dashboard → Requires log parsing + database
└── Certificate Expiry Alerts → Requires cron job + notification system

DarkPipe-Specific
├── Cloud Relay → Required for inbound SMTP (port 25)
├── Secure Transport (Cloud → Home) → Requires VPN or authenticated relay
└── GitHub Actions Build → Required for multi-stack support
```

### Dependency Notes

- **SPF/DKIM before DMARC:** DMARC policy is meaningless without SPF and DKIM configured. Must implement in order.
- **TLS Certificates before HTTPS-dependent features:** MTA-STS, webmail, admin panel all require valid TLS certificates. Let's Encrypt must be working first.
- **DNS API integration enhances but doesn't replace validation:** Even with automated DNS updates, validation checks are required to catch provider API issues.
- **Sieve requires Dovecot:** If you swap Dovecot for another IMAP server, Sieve support may not be available.
- **CalDAV/CardDAV are bundled:** Solutions like Nextcloud, Baïkal, Radicale provide both; rarely makes sense to split them.
- **Queue encryption conflicts with content filters:** If you encrypt queue at rest, Amavis/SpamAssassin/ClamAV can't read messages. Must decrypt for scanning or scan before encryption.
- **Cloud relay is DarkPipe's architectural constraint:** Home devices can't receive inbound SMTP (port 25 blocked, dynamic IP). Cloud relay is mandatory for receiving mail.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **Inbound SMTP relay (cloud)** — Receives mail from internet
- [x] **Outbound SMTP relay (home device)** — Sends mail to internet
- [x] **IMAP server (home device)** — Users access mail via clients
- [x] **SPF/DKIM/DMARC setup** — Deliverability baseline
- [x] **TLS encryption (STARTTLS)** — Required for modern email
- [x] **Let's Encrypt auto-renewal** — Avoid certificate expiry outages
- [x] **Basic webmail (Roundcube)** — Web access for non-technical users
- [x] **Spam filtering (Rspamd/SpamAssassin)** — Inbound spam protection
- [x] **DNS validation tool** — Check if setup is correct
- [x] **Multi-user support** — Family/small team use case
- [x] **GitHub Actions build system** — Core differentiator: user customizes stack
- [x] **Multi-arch Docker images (ARM64 + x86_64)** — Raspberry Pi + VPS support

**Rationale:** This is the minimum feature set to replace Gmail for a privacy-focused user. Without cloud relay + home device architecture, DarkPipe has no value prop. Without DNS validation, users will fail at setup. Without webmail, non-technical family members can't participate. Without multi-arch images, Raspberry Pi users (core audience) can't run it.

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **CalDAV/CardDAV server** — Trigger: users request calendar/contacts sync (likely immediate)
- [ ] **Sieve filtering with ManageSieve** — Trigger: power users want server-side rules
- [ ] **MTA-STS + DANE** — Trigger: privacy-focused users want TLS enforcement
- [ ] **DNS API integration (Cloudflare, Route53)** — Trigger: users struggle with manual DNS setup
- [ ] **TLS-RPT logging** — Trigger: users want visibility into TLS failures
- [ ] **Queue health monitoring** — Trigger: users experience stuck mail, don't know why
- [ ] **Certificate expiry alerts** — Trigger: user's cert expires, service goes down
- [ ] **Prometheus metrics export** — Trigger: homelab users want Grafana dashboards
- [ ] **Strict TLS mode (reject plaintext peers)** — Trigger: high-privacy users want this hardening
- [ ] **Multiple domain support** — Trigger: users want to consolidate identities

**Rationale:** These features are high-value but not launch-blockers. CalDAV/CardDAV will likely be requested immediately (Gmail replacement requires calendar/contacts). Monitoring features become critical as users scale up. TLS enforcement features serve the privacy-focused niche.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Queue encryption at rest** — Why defer: high complexity, niche demand, conflicts with spam scanning
- [ ] **PGP/WKD support** — Why defer: niche user base, requires key management UX
- [ ] **S/MIME support** — Why defer: enterprise feature, less relevant for personal email
- [ ] **Full-text search with attachment indexing** — Why defer: high resource usage, not critical for small mailboxes
- [ ] **Audit logging for admin actions** — Why defer: relevant for multi-admin orgs, not v1 audience
- [ ] **Rate limiting outbound mail** — Why defer: anti-abuse feature, less relevant for trusted family users
- [ ] **Deliverability scoring integration** — Why defer: nice-to-have diagnostics, not core functionality
- [ ] **One-click deploy to cloud marketplaces** — Why defer: high partnership/integration effort
- [ ] **SOGo webmail (alternative to Roundcube)** — Why defer: dual webmail support adds maintenance burden

**Rationale:** These are "nice-to-have" features that serve narrow use cases or require disproportionate effort. Queue encryption and PGP serve extreme privacy users (niche within niche). Deliverability scoring and audit logs are diagnostic tools that can wait. Full-text search with attachments is resource-intensive and most users search by sender/subject anyway.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Cloud relay architecture | HIGH | HIGH | P0 (MVP blocker) |
| SMTP/IMAP/TLS | HIGH | LOW | P0 (MVP blocker) |
| SPF/DKIM/DMARC | HIGH | LOW | P0 (MVP blocker) |
| DNS validation tool | HIGH | LOW | P0 (MVP blocker) |
| Webmail (Roundcube) | HIGH | MEDIUM | P0 (MVP blocker) |
| Multi-arch builds | HIGH | MEDIUM | P0 (MVP blocker) |
| GitHub Actions customization | HIGH | HIGH | P0 (core differentiator) |
| CalDAV/CardDAV | HIGH | MEDIUM | P1 (post-launch) |
| Sieve filtering | MEDIUM | MEDIUM | P1 (post-launch) |
| MTA-STS + DANE | MEDIUM | MEDIUM | P1 (post-launch) |
| DNS API integration | HIGH | MEDIUM | P1 (reduces support burden) |
| Certificate expiry alerts | HIGH | LOW | P1 (prevents outages) |
| Queue health monitoring | MEDIUM | LOW | P1 (prevents support tickets) |
| Prometheus metrics | MEDIUM | LOW | P1 (homelab users expect this) |
| Strict TLS mode | MEDIUM | MEDIUM | P2 (niche privacy feature) |
| TLS-RPT | LOW | LOW | P2 (diagnostics) |
| Multiple domain support | MEDIUM | MEDIUM | P2 (defer until requested) |
| Queue encryption at rest | LOW | HIGH | P3 (niche, high complexity) |
| PGP/WKD | LOW | HIGH | P3 (niche, key mgmt complexity) |
| Full-text search + attachments | MEDIUM | HIGH | P3 (resource-intensive) |
| Deliverability scoring | LOW | MEDIUM | P3 (nice-to-have diagnostics) |
| One-click cloud deploy | MEDIUM | HIGH | P3 (partnership effort) |
| Audit logging | LOW | MEDIUM | P3 (enterprise feature) |

**Priority key:**
- **P0:** Must have for launch (MVP blockers)
- **P1:** Should have, add in first 3 months post-launch
- **P2:** Could have, add based on user feedback
- **P3:** Won't have initially, consider for v2.0+

## Competitor Feature Analysis

| Feature | Mail-in-a-Box | Mailu | docker-mailserver | Mailcow | DarkPipe (Planned) |
|---------|---------------|-------|-------------------|---------|-------------------|
| **Architecture** | Monolithic Ubuntu script | Docker Compose | Single Docker image | Docker Compose | **Cloud relay + home device** |
| **SMTP/IMAP/TLS** | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ Postfix/Dovecot | ✅ User-chosen via GitHub Actions |
| **SPF/DKIM/DMARC** | ✅ Auto-generated | ✅ Supported | ✅ Supported | ✅ Supported | ✅ Auto-generated + validation |
| **MTA-STS** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **DANE** | ✅ | ✅ | ❌ | ❌ | ✅ (planned) |
| **Webmail** | ✅ Roundcube | ✅ Roundcube/Rainloop | ❌ (BYO) | ✅ SOGo/Roundcube | ✅ Roundcube (default) |
| **CalDAV/CardDAV** | ✅ Nextcloud | ❌ (separate install) | ❌ (BYO) | ✅ SOGo | ✅ (planned: Baïkal or SOGo) |
| **Admin Panel** | ✅ Custom | ✅ Custom | ❌ (CLI only) | ✅ Custom + API | ✅ (planned: minimal web UI) |
| **DNS Management** | ✅ Can host DNS | ❌ Manual | ❌ Manual | ❌ Manual | **✅ API integration (Cloudflare, Route53)** |
| **Spam Filtering** | ✅ SpamAssassin | ✅ Rspamd | ✅ Rspamd | ✅ Rspamd | ✅ Rspamd (default) |
| **Greylisting** | ✅ Postgrey | ✅ Rspamd | ✅ Postgrey | ✅ Rspamd | ✅ Rspamd |
| **Virus Scanning** | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV | ✅ ClamAV (optional) |
| **Backups** | ✅ Duplicity to S3 | ❌ (manual) | ❌ (manual) | ✅ Via admin panel | ✅ (planned: backup to user's cloud) |
| **Multi-Domain** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **Sieve Filtering** | ✅ | ✅ | ✅ | ✅ | ✅ (planned) |
| **2FA** | ✅ TOTP | ❌ | ❌ | ✅ TOTP | ✅ (planned) |
| **Multi-Arch (ARM64)** | ❌ (x86_64 only) | ✅ | ✅ | ⚠️ (experimental) | **✅ First-class ARM64 support** |
| **User Customization** | ❌ Opinionated | ❌ Opinionated | ⚠️ Config files | ⚠️ Limited | **✅ GitHub Actions: choose your stack** |
| **Home Device Support** | ❌ Assumes public IP | ❌ Assumes public IP | ❌ Assumes public IP | ❌ Assumes public IP | **✅ Cloud relay for inbound** |
| **Privacy Focus** | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | ⚠️ Self-hosted (good) | **✅ Storage on user device** |

### Key Differentiators vs Competitors

1. **DarkPipe's unique value:** Cloud relay + home device split solves the "home ISP blocks port 25" problem that kills other self-hosted solutions.
2. **GitHub Actions customization:** No other solution lets users choose stack components (Postfix vs Stalwart, Roundcube vs SOGo) via templated builds.
3. **First-class ARM64 support:** Mail-in-a-Box doesn't support ARM. Mailcow's ARM support is experimental. DarkPipe treats Raspberry Pi as primary deployment target.
4. **DNS API integration:** Competitors require manual DNS setup or (Mail-in-a-Box) full DNS hosting. DarkPipe automates via provider APIs.
5. **Privacy architecture:** Competitors assume VPS = trusted. DarkPipe assumes cloud relay = untrusted, storage = home device only.

### What Competitors Do Well (Learn From)

- **Mail-in-a-Box:** Excellent DNS validation UI. Copy this. Comprehensive status checks. One-command install.
- **Mailu:** Clean Docker Compose structure. Role-based admin delegation. Good anti-spam (Rspamd).
- **docker-mailserver:** Configuration via files (GitOps-friendly). `setup.sh` utility for CLI admin tasks.
- **Mailcow:** Polished admin UI. SOGo integration (best webmail + groupware). Queue viewer for debugging.

### What Competitors Struggle With (Avoid)

- **Mail-in-a-Box:** Monolithic bash script. Can't customize components. No Docker (harder to deploy/update). x86_64 only.
- **Mailu:** No CalDAV/CardDAV (deal-breaker for Gmail replacement). Manual DNS setup (high failure rate).
- **docker-mailserver:** No admin panel (CLI-only intimidates users). No webmail (BYO). Steep learning curve.
- **Mailcow:** Heavy resource usage (runs many containers). Complex to customize. Assumes public IP (no home device support).

## Helm's Feature Set & Failure Points

### What Helm Offered

Helm was a **$499 hardware device** (personal email server) that launched in 2018 and shut down in December 2022. It targeted non-technical users wanting email privacy.

**Feature Set:**
- Email server (SMTP/IMAP) with custom domain support
- Contacts and calendar (CalDAV/CardDAV)
- Notes and file storage
- **120GB SSD storage** (expandable to 5TB)
- **Proximity-based security:** Bluetooth token for 2FA
- **Full disk encryption** with Secure Enclave-managed keys
- **Cloud backup:** Nightly encrypted backups
- DMARC/SPF/DKIM authentication
- TLS encryption (client-server, server-server)
- PGP/S/MIME support (at MUA level, not MTA)

### Why Helm Failed

**Critical Failure Points:**

1. **Supply chain issues:** Shifted manufacturing from Mexico to China in late 2019 to reduce costs. COVID-19 devastated supply chain. Unable to manufacture/ship devices.

2. **Hardware dependence:** Physical device = capital expenditure, inventory risk, shipping logistics. Software-only competitors (Proton Mail) scaled without these constraints.

3. **Subscription dependency:** Service required ongoing cloud backup subscription. When revenue dropped (couldn't ship new devices), cloud costs became unsustainable.

4. **No webmail:** Users had to configure IMAP clients. Non-technical users struggled. Helm's target audience expected "just works" web access.

5. **Home network complexity:** Users had to configure port forwarding, DDNS, or use Helm's relay. Many failed at network setup.

6. **Locked-in hardware:** $499 device became e-waste when service shut down. Users lost investment. (Helm did release Armbian conversion firmware, but most users didn't convert.)

7. **Single point of failure:** Helm the company = Helm the service. No open-source alternative to migrate to.

### Lessons for DarkPipe

**Do:**
- ✅ **Separate hardware and software:** DarkPipe runs on user's existing hardware (Raspberry Pi, old laptop). No inventory/shipping risk.
- ✅ **Open-source + self-hosted:** If DarkPipe project ends, users keep running their servers. No vendor lock-in.
- ✅ **Cloud relay for inbound:** Solves Helm's "home network complexity" problem. Users don't configure port forwarding.
- ✅ **Include webmail:** Roundcube/SOGo required for non-technical users.
- ✅ **Donation-funded, not subscription:** No recurring revenue pressure. Cloud relay costs covered by donations or minimal fees.

**Don't:**
- ❌ **Don't sell hardware:** Helm's $499 device = capital risk. DarkPipe = bring your own device.
- ❌ **Don't create service dependencies:** Helm's cloud backup was a recurring cost center. DarkPipe users manage their own backups.
- ❌ **Don't target "zero-configuration":** Helm promised simplicity, delivered complexity. DarkPipe targets "homelab-adjacent" users who tolerate config.
- ❌ **Don't hide relay operation:** Helm's relay was opaque. DarkPipe should show users exactly what cloud relay does (logs, audit trail).

**Helm's Legacy:**
- Proved market demand for privacy-focused email (1000s of units sold).
- Showed that hardware + subscription model is fragile.
- Validated need for "works behind home ISP" solution (port 25 blocking).
- Demonstrated that non-technical users will pay for privacy (if UX is good enough).

## Sources

### Competitor Analysis
- [Mail-in-a-Box](https://mailinabox.email/)
- [Mail-in-a-Box GitHub](https://github.com/mail-in-a-box/mailinabox)
- [Mailu](https://mailu.io/)
- [Mailu GitHub](https://github.com/Mailu/Mailu)
- [docker-mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/)
- [docker-mailserver GitHub](https://github.com/docker-mailserver/docker-mailserver)
- [Mailcow January 2026 Update](https://mailcow.email/posts/2026/release-2026-01/)
- [Mailcow GitHub](https://github.com/mailcow/mailcow-dockerized)
- [Mailcow features analysis](https://www.servercow.de/mailcow)

### Helm Analysis
- [Helm Shutdown FAQ](https://support.thehelm.com/hc/en-us/articles/10831596925203-Helm-Shutdown-FAQ)
- [Helm Email Server Review](https://blog.strom.com/wp/?p=6990)
- [Fortune: Helm's Private Email Server](https://fortune.com/2019/03/09/helm-server-gmail-privacy/)
- [SlashGear: Helm Personal Email Server](https://www.slashgear.com/helm-personal-email-server-promises-perfect-privacy-17550402/)
- [GeekWire: Seattle startup vets take on Google with Helm](https://www.geekwire.com/2018/seattle-startup-vets-take-tech-giants-helm-new-personal-email-server/)

### Technical Standards & Protocols
- [How to Host Your Own Email Server in 2026](https://elementor.com/blog/how-to-host-your-own-email-server/)
- [Self-Hosting Email in 2026: Is Running a Linux Mail Server Still Worth It?](https://securityboulevard.com/2025/12/self-hosting-email-in-2026-is-running-a-linux-mail-server-still-worth-it/)
- [SPF DKIM DMARC Explained 2026](https://skynethosting.net/blog/spf-dkim-dmarc-explained-2026/)
- [Email Security Market: 8 Things to Know for 2026 Planning](https://abnormal.ai/blog/email-security-market)
- [MTA-STS Guide](https://redsift.com/guides/email-security-guide/mta-sts)
- [DMARC, DANE, MTA-STS, TLS, and DKIM Explained](https://www.anubisnetworks.com/blog/dmarc_dane_explained)

### CalDAV/CardDAV
- [Calendar & Contacts - awesome-selfhosted](https://awesome-selfhosted.net/tags/calendar--contacts.html)
- [Baïkal](https://sabre.io/baikal/)
- [Best 11 Open-source CalDAV Self-hosted Servers](https://medevel.com/11-caldav-os-servers/)
- [Building a self-hosted CalDAV server](https://rfrancocantero.medium.com/building-a-self-hosted-caldav-server-the-technical-reality-behind-calendar-sharing-9a930af28ff0)

### Sieve Filtering
- [Sieve (mail filtering language) - Wikipedia](https://en.wikipedia.org/wiki/Sieve_(mail_filtering_language))
- [Advanced Email Filtering with Sieve - Docker Mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/config/advanced/mail-sieve/)
- [Proton Mail Sieve Filters](https://proton.me/support/sieve-advanced-custom-filters)

### Multi-Architecture & Deployment
- [How to Build Multi-Architecture Docker Images (ARM64 + AMD64)](https://oneuptime.com/blog/post/2026-01-06-docker-multi-architecture-images/view)
- [How to Build Multi Architecture Docker Images](https://devopscube.com/build-multi-arch-docker-image/)
- [GitHub Actions: Early February 2026 updates](https://github.blog/changelog/2026-02-05-github-actions-early-february-2026-updates/)

### Monitoring & Security
- [SSL Certificate Monitor - Dynatrace](https://www.dynatrace.com/hub/detail/ssl-certificate-monitor/)
- [TLS Certificate Monitoring with OpenTelemetry](https://www.elastic.co/observability-labs/blog/edot-certificate-monitoring)
- [NETSCOUT SSL certificate monitoring](https://www.helpnetsecurity.com/2026/01/27/netscout-ngeniusone-enhancements/)
- [Postfix Hardening Guide](https://linux-audit.com/postfix-hardening-guide-for-security-and-privacy/)
- [Postfix TLS Support](http://www.postfix.org/TLS_README.html)

### Privacy-Focused Email
- [Privacy Guides: Self-Hosting Email](https://www.privacyguides.org/en/self-hosting/email-servers/)
- [Self Hosted Email: Privacy, Security and Full Control](https://www.smartertools.com/blog/2025/01/self-hosting-email)
- [Forward Email - 100% open-source and privacy-focused](https://forwardemail.net/en/blog/best-private-email-service)

### DNS Automation
- [DNSimple TLSA Record Support](https://blog.dnsimple.com/2026/01/tlsa-record-support/)
- [Best Email API For Developers](https://mailtrap.io/blog/best-email-api/)

---
*Feature research for: DarkPipe (Privacy-First Self-Hosted Email)*
*Researched: 2026-02-08*
*Confidence: MEDIUM-HIGH (Web search verified with official docs; some gaps around queue encryption and bleeding-edge features)*
