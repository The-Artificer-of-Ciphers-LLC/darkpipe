# M005: Design Validation — External Access & Device Connectivity — Context

**Gathered:** 2026-03-12
**Status:** Ready for planning

## Project Description

Validate that the DarkPipe architecture works end-to-end from the public internet. Verify that external devices (phones, laptops, desktop mail clients) can reach the home mail server through the cloud relay while away from the home network. Confirm DNS resolution, TLS certificates, IMAP/SMTP connectivity, webmail access, and device profile onboarding all function correctly from outside the LAN.

## Why This Milestone

DarkPipe shipped v1.0 and passed container/compose validation in CI, but no milestone has verified the full external access path — the scenario where a user is away from home and needs to send/receive email, check webmail, or sync calendar/contacts. This is the core promise of the architecture: your mail lives at home but is accessible from anywhere via the cloud relay. Without validating this, the project's central claim is untested.

## User-Visible Outcome

### When this milestone is complete, the user can:

- Send and receive email from a phone (iOS/Android) connected to cellular data (not home WiFi)
- Access webmail via HTTPS from any browser on any network
- Sync calendar and contacts from a mobile device over cellular
- Confirm IMAP (993) and SMTP (587) are reachable from the public internet through the relay
- Verify that DNS records (MX, SPF, DKIM, DMARC, autoconfig, autodiscover) resolve correctly from external DNS
- Confirm TLS certificates are valid and trusted by standard mail clients
- Verify the WireGuard/mTLS tunnel between cloud relay and home device is stable under real conditions

### Entry point / environment

- Entry point: External device (phone on cellular, laptop on coffee shop WiFi) connecting to mail.yourdomain.com
- Environment: Production-like — real VPS, real home device, real DNS, real TLS certificates
- Live dependencies involved: VPS provider, DNS (Cloudflare), Let's Encrypt, WireGuard/mTLS tunnel, home network (NAT/firewall)

## Completion Class

- Contract complete means: DNS records validate, TLS certificates pass, ports respond, mail round-trips succeed
- Integration complete means: External device sends/receives email through the full relay→tunnel→home chain
- Operational complete means: Connection survives home device restart, tunnel reconnection, and brief network interruptions

## Final Integrated Acceptance

To call this milestone complete, we must prove:

- A phone on cellular data can send an email that is delivered to an external recipient (Gmail/Outlook) and receive a reply back — full round-trip
- Webmail loads over HTTPS from an external network with a valid TLS certificate
- IMAP and SMTP connections from a desktop mail client (Thunderbird) on an external network authenticate and sync
- DNS validation passes for all required records (MX, A, SPF, DKIM, DMARC, autoconfig, autodiscover, SRV)
- The monitoring dashboard shows healthy status for all services during external access

## Risks and Unknowns

- **Home network NAT/firewall** — residential routers may block or interfere with incoming WireGuard traffic; port forwarding or UPnP may be needed
- **ISP restrictions** — some ISPs block certain ports or throttle VPN traffic; WireGuard uses UDP which is less commonly blocked
- **DNS propagation** — records may not be visible from all resolvers immediately; TTL and caching affect validation
- **TLS certificate issuance** — Let's Encrypt requires port 80/443 accessible for HTTP-01 challenge; if behind NAT, needs Caddy or DNS-01 challenge
- **IP reputation** — new VPS IP may be on blocklists; outbound email may be rejected by recipients
- **Stalwart/mail server config** — production config may differ from test scenarios; auth, TLS, and LMTP delivery need real validation

## Existing Codebase / Prior Art

- `cloud-relay/docker-compose.yml` — Cloud relay service definitions (Postfix, Caddy, relay daemon)
- `home-device/docker-compose.yml` — Home device services (mail server, webmail, profile server, Rspamd)
- `deploy/wireguard/` — WireGuard setup scripts for cloud and home
- `deploy/pki/` — mTLS PKI setup scripts
- `dns/cmd/dns-setup/` — DNS record creation and validation tool
- `docs/quickstart.md` — Step-by-step deployment guide
- `scripts/check-runtime.sh` — Runtime environment validation

> See `.gsd/DECISIONS.md` for all architectural and pattern decisions.

## Relevant Requirements

- Cloud relay receives inbound SMTP with TLS — needs external validation
- Encrypted transport (WireGuard or mTLS) — needs real tunnel under production conditions
- Device onboarding via profiles and QR codes — needs real device testing
- Webmail accessible from mobile — needs external browser access
- CalDAV/CardDAV sync — needs real device calendar/contacts sync

## Scope

### In Scope

- External DNS record validation (all records from a non-local resolver)
- TLS certificate validation (Let's Encrypt, chain, expiry)
- IMAP (993) and SMTP (587) reachability from external networks
- Webmail HTTPS access from external networks
- Full email round-trip: external → cloud relay → tunnel → home → mailbox → IMAP → client
- Full email round-trip: client → SMTP → home → tunnel → cloud relay → external recipient
- Mobile device onboarding (iOS .mobileconfig, autodiscovery)
- CalDAV/CardDAV sync from external network
- WireGuard/mTLS tunnel stability (reconnection after brief outage)
- Monitoring dashboard accessibility and health status
- Document any issues found and fixes applied

### Out of Scope / Non-Goals

- Performance benchmarking or load testing
- Multi-user provisioning beyond admin account
- Spam filter tuning
- IP warmup (time-dependent, not validation-dependent)
- Backup/restore validation
- Migration from other providers

## Technical Constraints

- Requires a live VPS with port 25 open (cannot be simulated in CI)
- Requires a live home device on a real residential network
- Requires a real domain with DNS control
- Requires a real phone or device on a non-home network for testing
- Some tests are environment-dependent (ISP, router, DNS provider)

## Integration Points

- **VPS / Cloud Relay** — Real cloud relay running on a VPS with public IP
- **Home Device** — Real home device on residential network
- **DNS Provider** — Cloudflare (or other) for record management
- **Let's Encrypt** — TLS certificate issuance and validation
- **External mail providers** — Gmail/Outlook for inbound/outbound round-trip
- **Mobile devices** — iOS/Android for profile installation and mail client testing
- **Desktop mail clients** — Thunderbird for IMAP/SMTP validation

## Open Questions

- Which VPS provider and home device hardware will be used for validation? — Depends on user's existing setup
- Is WireGuard or mTLS the transport to validate? — Validate whichever the user has configured (WireGuard is default)
- What domain will be used? — User's existing domain
