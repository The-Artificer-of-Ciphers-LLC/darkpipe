---
phase: quick-02
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - README.md
  - docs/architecture.md
  - docs/quickstart.md
  - docs/configuration.md
  - docs/migration.md
  - docs/contributing.md
  - docs/security.md
  - docs/faq.md
autonomous: true

must_haves:
  truths:
    - "A new visitor to the GitHub repo immediately understands what DarkPipe is, why it exists, and how to get started"
    - "A technical user can go from zero to running DarkPipe by following the quickstart guide"
    - "A potential contributor understands the project structure, conventions, and how to submit changes"
    - "All documentation is accurate to the shipped v1.0 codebase"
  artifacts:
    - path: "README.md"
      provides: "Project overview, badges, architecture diagram, feature list, quick install, links to detailed docs"
      min_lines: 200
    - path: "docs/architecture.md"
      provides: "System architecture with ASCII diagrams, component descriptions, data flow"
      min_lines: 100
    - path: "docs/quickstart.md"
      provides: "End-to-end setup guide from VPS provisioning to sending first email"
      min_lines: 150
    - path: "docs/configuration.md"
      provides: "Complete configuration reference for all components and environment variables"
      min_lines: 100
    - path: "docs/migration.md"
      provides: "Mail migration guide for all 7 supported providers"
      min_lines: 80
    - path: "docs/contributing.md"
      provides: "Contribution guidelines, code conventions, PR process, license requirements"
      min_lines: 80
    - path: "docs/security.md"
      provides: "Security model, threat model, reporting vulnerabilities"
      min_lines: 60
    - path: "docs/faq.md"
      provides: "Frequently asked questions covering common concerns and troubleshooting"
      min_lines: 80
  key_links:
    - from: "README.md"
      to: "docs/*.md"
      via: "markdown links"
      pattern: "\\]\\(docs/"
    - from: "README.md"
      to: "deploy/platform-guides/"
      via: "markdown links"
      pattern: "platform-guides"
    - from: "docs/quickstart.md"
      to: "docs/configuration.md"
      via: "cross-reference links"
      pattern: "configuration\\.md"
---

<objective>
Generate complete user-facing documentation for DarkPipe's v1.0 public release.

Purpose: DarkPipe is going public. The repo currently has no README.md and no user-facing documentation beyond platform guides and a VPS providers guide. A polished documentation set is essential for an open-source project seeking users and contributors. Every visitor's first impression comes from the README.

Output: README.md (compelling project landing page) + 7 docs/ files covering architecture, quickstart, configuration, migration, contributing, security, and FAQ.
</objective>

<execution_context>
@/Users/trekkie/.claude/get-shit-done/workflows/execute-plan.md
@/Users/trekkie/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@home-device/docker-compose.yml
@cloud-relay/docker-compose.yml
@deploy/setup/cmd/darkpipe-setup/main.go
@dns/cmd/dns-setup/main.go
@go.mod
@.github/workflows/release.yml
@.github/workflows/build-custom.yml
@.github/workflows/build-prebuilt.yml
@deploy/platform-guides/raspberry-pi.md
@docs/vps-providers.md
@LICENSE
@THIRD-PARTY-LICENSES.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create README.md with architecture diagram, badges, and feature showcase</name>
  <files>README.md</files>
  <action>
Create a compelling README.md at the repository root. This is the single most important file for an open-source project -- it must immediately communicate what DarkPipe is, why someone should care, and how to get started. Structure:

**Header section:**
- Project name with tagline: "Your email. Your hardware. Your rules."
- Badges row: License (AGPLv3), Go version (1.25), GitHub release (latest), GitHub Actions build status, platform support (linux/amd64, linux/arm64)
- Badge URLs: Use shields.io with `github.com/trek-e/darkpipe` as the repo path
- One-paragraph description of what DarkPipe does (adapt from PROJECT.md core value)

**ASCII architecture diagram:**
Create a clear diagram showing the split architecture:
```
Internet --> Cloud Relay (VPS) --[WireGuard/mTLS]--> Home Device (your hardware)
                |                                          |
           Postfix MTA                              Mail Server
           Certbot TLS                           (Stalwart/Maddy/Postfix+Dovecot)
           Rspamd Filter                            Webmail + Calendar
           Monitoring                              Device Profiles
                                                   Offline Queue
```
Show the data flow: inbound mail arrives at cloud relay, gets filtered, transported encrypted to home device, stored locally. Outbound: home device sends via relay. Emphasize: NO mail stored on cloud relay.

**Key features section** (bulleted, grouped):
- Mail server choices (3 options)
- Webmail choices (2 options)
- CalDAV/CardDAV (Radicale or Stalwart built-in)
- Transport security (WireGuard or mTLS)
- DNS automation (SPF/DKIM/DMARC with Cloudflare and Route53 API)
- Encrypted offline queue with S3 overflow
- Device onboarding (Apple .mobileconfig, QR codes, autodiscovery)
- App passwords for mail clients
- Migration from 7 providers (Gmail, Outlook, iCloud, MailCow, Mailu, docker-mailserver, generic IMAP)
- Multi-arch Docker images (amd64 + arm64)
- Web monitoring dashboard with alerts
- Rspamd spam filtering with greylisting

**Quick start section:**
Show the 3-step happy path:
1. Provision a VPS with port 25 access (link to docs/vps-providers.md)
2. Download and run darkpipe-setup wizard
3. Configure DNS (link to dns-setup tool)
Include actual CLI commands. Reference the setup tool binary from GitHub releases.

**Stack configurations section:**
Table showing the two pre-built stacks:
| Stack | Mail Server | Webmail | Calendar/Contacts |
Default: Stalwart + SnappyMail + built-in CalDAV/CardDAV
Conservative: Postfix+Dovecot + Roundcube + Radicale

**Supported platforms section:**
List with links to platform guides:
- Raspberry Pi 4+ (arm64, 4GB+ RAM)
- TrueNAS Scale 24.10+
- Unraid
- Proxmox LXC
- Synology NAS (Container Manager)
- Mac Silicon (Apple M-series)
- Any Docker-capable x64/arm64 Linux host

**Documentation links section:**
Table or list linking to all docs/ files with one-line descriptions.

**Community and support section:**
- GitHub Discussions for questions and community interaction
- GitHub Issues for bug reports
- Link to contributing guide

**Sustainability section:**
Brief note: AGPLv3, funded by donations. Links to GitHub Sponsors, Open Collective, Liberapay, Ko-fi (use placeholder URLs like `https://github.com/sponsors/trek-e`).

**Footer:**
- License notice: AGPLv3 with link to LICENSE file
- Copyright: The Artificer of Ciphers, LLC
- "Built because your inbox shouldn't live on someone else's computer." (or similar closing line)

IMPORTANT: Do NOT use emojis anywhere. Use plain text, dashes, and standard markdown formatting.
IMPORTANT: All badge and link URLs should use `trek-e/darkpipe` as the GitHub path.
IMPORTANT: Do NOT invent features that don't exist. Every claim must be backed by actual code in the repo.
  </action>
  <verify>
Verify README.md exists and is well-formed:
- File exists at repo root
- Contains architecture ASCII diagram
- Contains shields.io badge URLs
- Contains links to all docs/ files
- Contains links to platform guides
- Contains quick start CLI commands
- No emojis present
- All internal links point to files that will exist after all tasks complete
  </verify>
  <done>README.md exists with 200+ lines covering project overview, architecture diagram, features, quick start, stack configs, platform support, documentation links, community info, and license -- ready for public consumption.</done>
</task>

<task type="auto">
  <name>Task 2: Create architecture, quickstart, configuration, and migration docs</name>
  <files>docs/architecture.md, docs/quickstart.md, docs/configuration.md, docs/migration.md</files>
  <action>
Create four technical documentation files in docs/:

**docs/architecture.md -- System Architecture**

Detailed architecture documentation covering:

1. "Architecture Overview" -- expanded ASCII diagram showing all components and their interactions. Show both cloud relay side and home device side with all services labeled. Include port numbers.

2. "Components" section with subsections for each major component:
   - Cloud Relay: Postfix MTA, Certbot/TLS, relay service (Go), monitoring agent
   - Transport Layer: WireGuard tunnel OR mTLS, PKI (step-ca for internal certs)
   - Home Device: Mail server (Stalwart/Maddy/Postfix+Dovecot), webmail, groupware, Rspamd, Redis, queue service, profile server, Caddy reverse proxy
   - DNS Management: dns-setup CLI, SPF/DKIM/DMARC record generation, Cloudflare and Route53 API integration
   - Monitoring: health checks, web dashboard, alert notifications
   - Build System: GitHub Actions workflows, custom and pre-built image builds

3. "Data Flow" section:
   - Inbound mail flow (internet -> relay -> transport -> home device -> mailbox)
   - Outbound mail flow (mail client -> home device -> transport -> relay -> internet)
   - Offline flow (relay queues encrypted, drains when home device reconnects)

4. "Security Model" summary (brief, link to docs/security.md for details):
   - No mail stored on cloud relay (pass-through only)
   - All transport encrypted (WireGuard or mTLS)
   - Internal PKI with automatic rotation
   - SPF/DKIM/DMARC for email authentication

5. "Directory Structure" -- annotated tree showing repo layout with descriptions of each top-level directory (use the structure from planning context, add brief descriptions)

**docs/quickstart.md -- Getting Started**

Complete end-to-end guide from zero to working email. Structure:

1. "Prerequisites" -- what you need before starting:
   - A domain name you control (with DNS API access for automation, or manual DNS editing)
   - A VPS with port 25 access (link to docs/vps-providers.md)
   - A home device running Docker (link to platform guides)
   - 30-60 minutes of setup time

2. "Step 1: Provision Your Cloud Relay" -- VPS setup basics:
   - Choose provider (recommend Hetzner or Vultr, link to VPS guide)
   - Provision smallest VPS (1 vCPU, 1GB RAM, 20GB SSD)
   - Set hostname, configure reverse DNS (PTR record)
   - Verify port 25 is open (telnet test)

3. "Step 2: Set Up Transport" -- WireGuard or mTLS between cloud and home:
   - Run wireguard setup scripts (deploy/wireguard/cloud-setup.sh, home-setup.sh)
   - OR run mTLS PKI setup (deploy/pki/step-ca-setup.sh)
   - Verify tunnel connectivity

4. "Step 3: Run the Setup Wizard" -- darkpipe-setup CLI:
   - Download from GitHub releases (show curl command for latest release)
   - Run interactive wizard: `./darkpipe-setup`
   - Wizard collects: domain, mail server choice, webmail choice, transport type
   - Wizard generates: docker-compose.yml, .env, configs

5. "Step 4: Configure DNS" -- dns-setup CLI:
   - Run `dns-setup --domain yourdomain.com --relay-hostname relay.yourdomain.com --relay-ip YOUR_VPS_IP`
   - Dry-run first (default), then `--apply` to write records
   - Validate with `--validate-only`

6. "Step 5: Deploy" -- docker compose up:
   - Deploy cloud relay on VPS
   - Deploy home device on home hardware
   - Verify services are running

7. "Step 6: Test" -- send and receive test email:
   - Use dns-setup `--send-test` to send a test email
   - Check webmail for received test
   - Send outbound test from webmail

8. "Step 7: Onboard Devices" -- set up mail clients:
   - Access profile server for QR codes / .mobileconfig
   - Configure Thunderbird/Outlook via autodiscovery
   - Generate app passwords for mail clients

9. "What's Next" -- link to configuration reference, migration guide, monitoring setup

**docs/configuration.md -- Configuration Reference**

Comprehensive reference for all configurable aspects:

1. "Environment Variables" -- table of all .env variables with defaults:
   - MAIL_DOMAIN, MAIL_HOSTNAME, ADMIN_EMAIL, ADMIN_PASSWORD
   - Transport settings (TRANSPORT_TYPE, WIREGUARD_*, MTLS_*)
   - Queue settings (QUEUE_ENCRYPTION_KEY, S3_* for overflow)
   - DNS settings (DNS_PROVIDER, CLOUDFLARE_API_TOKEN, AWS_*)
   - Monitoring settings (ALERT_*, DASHBOARD_*)
   Reference the docker-compose.yml files for the authoritative list.

2. "Docker Compose Profiles" -- how to select components:
   - Mail server profiles: stalwart, maddy, postfix-dovecot
   - Webmail profiles: roundcube, snappymail
   - Groupware profiles: radicale (or stalwart built-in)
   - Example compose commands for common combinations

3. "Mail Server Configuration" -- per-server config files:
   - Stalwart: home-device/stalwart/config.toml
   - Maddy: home-device/maddy/maddy.conf
   - Postfix+Dovecot: home-device/postfix-dovecot/ configs

4. "Transport Configuration":
   - WireGuard: deploy/wireguard/ scripts and config
   - mTLS: transport/mtls/ and transport/pki/

5. "DNS Configuration":
   - Supported DNS providers (Cloudflare, Route53)
   - Manual DNS setup fallback
   - DKIM key rotation

6. "Offline Queue Configuration":
   - Queue behavior (queue vs bounce)
   - S3 overflow settings
   - Encryption settings

7. "Monitoring Configuration":
   - Health check intervals
   - Alert notification channels
   - Dashboard access

8. "Custom Builds" -- GitHub Actions workflow_dispatch inputs for build-custom.yml

**docs/migration.md -- Mail Migration Guide**

Guide for migrating from existing email providers:

1. "Overview" -- what the migration tool does: IMAP-based mailbox migration with folder mapping, progress tracking, dry-run support

2. "Supported Providers" -- table:
   | Provider | Auth Method | Notes |
   Gmail: OAuth2 device flow, requires Google Cloud project with Gmail API enabled
   Outlook/Microsoft 365: OAuth2 device flow
   iCloud: App-specific password
   MailCow: IMAP credentials
   Mailu: IMAP credentials
   docker-mailserver: IMAP credentials
   Generic IMAP: IMAP credentials (any IMAP server)

3. "Before You Migrate" -- checklist:
   - DarkPipe fully set up and receiving mail
   - DNS pointing to your relay
   - Test account created on your mail server
   - Backup of source mailbox (recommended)

4. "Running a Migration" -- step-by-step:
   - `darkpipe-setup migrate` subcommand
   - Provider selection
   - Authentication flow (OAuth2 or credentials)
   - Dry-run first (default behavior)
   - Full migration with `--apply`
   - Progress monitoring

5. "Provider-Specific Notes" -- subsection for each provider with gotchas:
   - Gmail: OAuth2 device flow setup, label-to-folder mapping, Google Takeout as alternative
   - Outlook: OAuth2 setup, shared mailbox considerations
   - iCloud: How to generate app-specific password
   - Self-hosted (MailCow, Mailu, docker-mailserver): Direct IMAP access, may need to allow connections

6. "Troubleshooting" -- common issues: auth failures, timeout on large mailboxes, folder mapping conflicts

IMPORTANT for all docs: No emojis. Use `darkpipe` consistently (lowercase). All file paths relative to repo root. Include "Last Updated" footer with date 2026-02-15. Include AGPLv3 license reference in footer.
  </action>
  <verify>
Verify all four files exist and have substantive content:
- docs/architecture.md exists with ASCII diagrams and component descriptions
- docs/quickstart.md exists with numbered steps and CLI commands
- docs/configuration.md exists with environment variable tables
- docs/migration.md exists with provider table and migration steps
- All files contain cross-reference links to other docs
- No emojis present in any file
  </verify>
  <done>Four technical docs exist covering architecture (with diagrams and data flow), quickstart (end-to-end setup in 7 steps), configuration (complete env var and profile reference), and migration (7-provider guide with auth details).</done>
</task>

<task type="auto">
  <name>Task 3: Create contributing, security, and FAQ docs</name>
  <files>docs/contributing.md, docs/security.md, docs/faq.md</files>
  <action>
Create three community/operational documentation files in docs/:

**docs/contributing.md -- Contributing to DarkPipe**

1. "Welcome" -- brief welcoming message. DarkPipe is AGPLv3 open source, community contributions are valued.

2. "Ways to Contribute":
   - Report bugs (GitHub Issues with reproduction steps)
   - Suggest features (GitHub Discussions)
   - Submit code (Pull Requests)
   - Improve documentation
   - Test on new platforms and report results
   - Help other users in Discussions

3. "Development Setup":
   - Prerequisites: Go 1.25+, Docker, Docker Compose
   - Clone the repo
   - Build from source: `go build ./...` from repo root
   - Run tests: `go test ./...`
   - Build Docker images locally

4. "Code Conventions":
   - Go code follows standard Go conventions (gofmt, go vet)
   - Every .go file MUST include the SPDX copyright header:
     ```
     // SPDX-License-Identifier: AGPL-3.0-or-later
     // Copyright (c) The Artificer of Ciphers, LLC
     ```
   - Error handling: return errors, don't panic (except main/init)
   - Logging: use structured logging
   - Docker: multi-stage builds, minimal final images

5. "Pull Request Process":
   - Fork the repo, create a feature branch
   - Write tests for new functionality
   - Ensure `go test ./...` passes
   - Ensure `go vet ./...` passes
   - Update documentation if behavior changes
   - Update THIRD-PARTY-LICENSES.md if adding dependencies
   - Submit PR with clear description of what and why
   - PRs require review before merge

6. "License":
   - All contributions are licensed under AGPLv3
   - By submitting a PR, you agree to license your contribution under AGPLv3
   - The SPDX header is required on all new Go files

7. "Code of Conduct" -- brief statement: Be respectful, constructive, and inclusive. Harassment and discrimination are not tolerated. Maintainers reserve the right to remove contributions that violate these principles.

**docs/security.md -- Security**

1. "Security Model":
   - DarkPipe's core security promise: your mail is never stored on infrastructure you don't control
   - Cloud relay is a pass-through -- mail is relayed immediately, not stored
   - All transport between cloud and home is encrypted (WireGuard or mTLS)
   - Internal PKI with step-ca for certificate management
   - Certificates rotate automatically (configurable: 30/60/90 days)

2. "Encryption in Transit":
   - Internet to cloud relay: TLS (Let's Encrypt via Certbot)
   - Cloud relay to home device: WireGuard (full tunnel) or mTLS (certificate-based)
   - Home device to mail clients: TLS (IMAPS port 993, SMTPS port 465/587)
   - No unencrypted mail transport at any stage

3. "Encryption at Rest":
   - Offline queue: encrypted with age (filippo.io/age) before writing to disk or S3
   - Mail storage: depends on chosen mail server and filesystem encryption
   - Recommendation: use full-disk encryption (LUKS) on home device

4. "Email Authentication":
   - SPF: Authorized senders published in DNS
   - DKIM: Cryptographic message signing with automatic key rotation
   - DMARC: Policy enforcement with reporting
   - All configured automatically by dns-setup tool

5. "Spam Filtering":
   - Rspamd with greylisting
   - Runs on home device (not cloud relay)
   - Configurable thresholds and rules

6. "Threat Model" -- honest assessment:
   - What DarkPipe protects against: cloud provider reading your mail, mass surveillance of stored mail, third-party data breaches of mail storage
   - What DarkPipe does NOT protect against: compromised home device, nation-state targeting your specific VPS traffic, malware on your endpoints, social engineering
   - Cloud relay is a trusted component -- if compromised, an attacker could read mail in transit (but not stored mail, because there is none)

7. "Reporting Vulnerabilities":
   - Email security@darkpipe.org (placeholder -- note this will be updated with actual contact)
   - OR use GitHub Security Advisories (private vulnerability reporting)
   - Do NOT open public issues for security vulnerabilities
   - Expected response time: 48 hours for acknowledgment
   - Coordinated disclosure: 90 days before public disclosure

8. "Dependencies":
   - Link to THIRD-PARTY-LICENSES.md
   - All dependencies are open source with compatible licenses
   - go-imap v2 is beta -- monitor for security updates
   - Stalwart is pre-v1.0 -- monitor for security updates

**docs/faq.md -- Frequently Asked Questions**

Organize into sections:

1. "General":
   - What is DarkPipe? (one-paragraph answer)
   - Why not just use Gmail/Outlook/ProtonMail? (privacy argument, control argument)
   - How is this different from Mail-in-a-Box / Mailu / docker-mailserver? (split architecture -- those are all single-server; DarkPipe separates relay from storage for privacy)
   - Is self-hosted email hard? (DarkPipe automates the hard parts -- DNS, TLS, deliverability -- but requires basic Docker knowledge)
   - Will my emails go to spam? (IP warmup takes 4-6 weeks, proper DNS auth helps, link to deliverability tips)

2. "Setup and Deployment":
   - What VPS provider should I use? (link to docs/vps-providers.md, recommend Hetzner or Vultr)
   - Can I run everything on one server? (No -- DarkPipe's architecture requires cloud relay + home device. This is the core privacy design.)
   - What home hardware do I need? (Raspberry Pi 4+ with 4GB RAM, or any Docker-capable x64/arm64 machine)
   - Can I use Podman instead of Docker? (Podman with docker-compose compatibility should work but is not officially tested)
   - How much does it cost? (VPS: $3-6/month. Home hardware: whatever you already own. Domain: ~$10/year.)

3. "Mail Servers":
   - Which mail server should I choose? (Stalwart for most users -- modern, built-in CalDAV/CardDAV. Postfix+Dovecot for traditional/battle-tested. Maddy for minimal footprint.)
   - Can I switch mail servers later? (Yes, re-run setup wizard. Mail data needs migration between server formats.)
   - Does DarkPipe support POP3? (No. POP3 is a security liability and IMAP is the modern standard. Explicitly out of scope.)

4. "Security and Privacy":
   - Is my email encrypted end-to-end? (Transport is encrypted. At-rest encryption depends on your home device setup. DarkPipe is not E2EE like PGP -- it's about storage sovereignty, not content encryption.)
   - What happens if my VPS is compromised? (Attacker could read mail in transit but cannot access stored mail -- that's on your home device.)
   - What happens if my home device is offline? (Cloud relay queues mail encrypted. When home device reconnects, queue drains. Configurable: queue or bounce.)

5. "Operations":
   - How do I add users/domains? (Through mail server's admin interface or setup wizard)
   - How do I back up my mail? (Back up mail server data directory on home device. Cloud relay has no mail to back up.)
   - How do I update DarkPipe? (Pull new Docker images, docker compose up -d. Setup wizard handles config migrations.)
   - How do I migrate from another email provider? (Link to docs/migration.md)

6. "Troubleshooting":
   - Mail isn't being delivered (check DNS, port 25, IP reputation, transport tunnel)
   - Webmail isn't loading (check Caddy reverse proxy, container logs)
   - Transport tunnel is down (check WireGuard/mTLS status, monitoring dashboard)
   - Link to GitHub Issues for bug reports

IMPORTANT for all docs: No emojis. Include "Last Updated: 2026-02-15" footer. Include link back to README.md. Include AGPLv3 license reference.
  </action>
  <verify>
Verify all three files exist and have substantive content:
- docs/contributing.md exists with development setup and PR process
- docs/security.md exists with threat model and vulnerability reporting
- docs/faq.md exists with categorized Q&A
- All files link back to README.md or other docs
- No emojis present in any file
- Security doc includes vulnerability reporting process
- Contributing doc includes SPDX header requirement
  </verify>
  <done>Three community/operational docs exist: contributing guide (with code conventions and SPDX requirements), security policy (with honest threat model and vulnerability reporting), and FAQ (covering general, setup, mail server, security, operations, and troubleshooting questions).</done>
</task>

</tasks>

<verification>
After all tasks complete:
1. Verify README.md exists at repo root with badges, architecture diagram, and feature list
2. Verify all 7 docs/ files exist: architecture.md, quickstart.md, configuration.md, migration.md, contributing.md, security.md, faq.md
3. Verify all internal links in README.md resolve to existing files
4. Verify no emojis appear in any documentation file
5. Verify all docs reference AGPLv3 license
6. Verify ASCII architecture diagram is present and readable
</verification>

<success_criteria>
- README.md is compelling and complete for a v1.0 open-source project going public
- 7 documentation files in docs/ cover all aspects a user, contributor, or evaluator needs
- All cross-references between docs are valid
- Documentation accurately reflects the shipped v1.0 codebase (no invented features)
- No emojis anywhere in documentation
- Total: 8 new files, ~1500+ lines of documentation
</success_criteria>

<output>
After completion, create `.planning/quick/2-generate-all-project-documentation/2-SUMMARY.md`
</output>
