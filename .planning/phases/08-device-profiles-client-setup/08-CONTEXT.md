# Phase 8: Device Profiles & Client Setup - Context

**Gathered:** 2026-02-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Users onboard new devices (phones, tablets, desktops) to their DarkPipe mail server in under 2 minutes without manually entering server addresses, ports, or security settings. Includes .mobileconfig profiles, QR codes, autodiscovery endpoints, and an app-generated password system.

</domain>

<decisions>
## Implementation Decisions

### Profile delivery & content
- Profiles include everything: Email (IMAP + SMTP) + CalDAV + CardDAV in one .mobileconfig
- Users download profiles from webmail — "Add Device" button in webmail UI
- Per-user personalized profiles (pre-filled with their email address and server settings)

### QR code experience
- QR codes displayed in both webmail ("Add Device" page) and CLI (`darkpipe qr user@domain`)
- QR codes are single-use — once scanned and redeemed, the embedded token is invalidated
- Must generate a new QR code per device

### Autodiscovery protocols
- Support all major clients: Thunderbird (autoconfig), Outlook (autodiscover), Apple Mail, SRV records (RFC 6186)
- Autodiscovery served from cloud relay via Caddy — always internet-reachable
- Integrate SRV record creation with Phase 4 DNS tool (`darkpipe dns-setup` extended)

### App-generated passwords
- App passwords are the ONLY authentication method — users never set or manage mail passwords directly
- One app password per device — revoking one doesn't affect others
- Users manage passwords in both webmail (Settings > Devices) and CLI (admin override)

### Claude's Discretion
- Profile signing (unsigned vs self-signed cert from DarkPipe CA)
- Android autoconfig approach (standard XML vs enhanced app links)
- QR code content (URL-based vs inline settings)
- QR code password inclusion (embed password for zero-typing vs settings-only for security)
- Autoconfig endpoint authentication (public vs authenticated)
- App password creation flow (auto-generated during "Add Device" vs manual creation)
- App password format and strength (length, charset, grouping for readability)

</decisions>

<specifics>
## Specific Ideas

- "Under 2 minutes" is the hard target — every flow must be measured against this
- Webmail as the primary self-service portal for device setup (non-technical users don't touch CLI)
- Single-use QR codes — user explicitly chose most secure option
- App passwords as sole auth method — simplifies the mental model (no "which password do I use?")
- SRV records integrated into existing DNS tool, not a separate command

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-device-profiles-client-setup*
*Context gathered: 2026-02-14*
