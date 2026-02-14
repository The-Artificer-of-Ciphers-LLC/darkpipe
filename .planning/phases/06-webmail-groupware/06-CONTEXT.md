# Phase 6: Webmail & Groupware - Context

**Gathered:** 2026-02-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Web-based email access for non-technical household members plus CalDAV/CardDAV sync for calendars and contacts on phones and computers. Users read, compose, and send email in a browser. Users sync calendars and contacts bidirectionally with iOS/macOS/Android devices.

Does NOT include: calendar web UI (GRP-01), contacts web UI (GRP-02), shared calendars web UI (GRP-03) — those are v2.

</domain>

<decisions>
## Implementation Decisions

### Webmail selection & access
- User-selectable webmail client: both Roundcube and SnappyMail available (matches MAIL-01 pattern of offering choices via Docker compose profiles)
- Webmail accessed via subdomain: `mail.example.com` (requires DNS record per domain)
- Webmail runs on home device (keeps mail content local, aligns with DarkPipe privacy model)
- Webmail accessible remotely through the cloud relay tunnel (non-technical family members don't need VPN)

### CalDAV/CardDAV server choice
- Use Stalwart built-in CalDAV/CardDAV when user picked Stalwart as mail server; deploy standalone server for Maddy/Postfix+Dovecot setups
- Same credentials as mail account — user@domain + mail password works for CalDAV/CardDAV (one set of credentials per person)
- Auto-create default calendar + address book per user on account creation (ready to sync immediately)
- Standard well-known URLs: `/.well-known/caldav` and `/.well-known/carddav` for auto-discovery on iOS/macOS/Android

### Multi-user experience
- Basic calendar sharing between household members (read-only or read-write)
- Shared family address book visible and editable by all household members, plus individual private address books
- CalDAV/CardDAV remotely accessible through tunnel (phone syncs whether at home or away)

### Claude's Discretion
- Webmail authentication approach (same IMAP credentials vs SSO — pick based on mail server capabilities)
- Standalone CalDAV/CardDAV server for non-Stalwart setups (Radicale vs Baikal — pick best fit for lightweight philosophy)
- Account onboarding flow (how household members get accounts — align with Phase 3 user management)
- Email isolation model (strictly private vs admin-accessible — pick based on privacy model)
- Reverse proxy choice (Caddy vs Nginx vs Traefik — pick best fit for DarkPipe's Go stack and auto-TLS needs)
- Webmail hosting location rationale and tunnel routing architecture

</decisions>

<specifics>
## Specific Ideas

- The user-selectable pattern should mirror Phase 3 (MAIL-01) — Docker compose profiles for webmail choice
- Calendar sharing is basic (v1 scope) — read-only or read-write per calendar, not granular event-level permissions
- Shared family address book is a good default for households (doctor, dentist, school contacts everyone needs)
- Well-known URLs are critical for zero-config device setup (feeds into Phase 8 device profiles)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-webmail-groupware*
*Context gathered: 2026-02-14*
