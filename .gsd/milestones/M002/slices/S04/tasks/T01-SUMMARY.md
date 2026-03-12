---
id: T01
parent: S04
milestone: M002
provides:
  - cloud-relay/.env.example with all 22 config.go vars + 6 compose-level vars
  - home-device/.env.example with all compose-level vars across all profiles
key_files:
  - cloud-relay/.env.example
  - home-device/.env.example
key_decisions:
  - Commented out conditional vars (mTLS, overflow) rather than showing empty values — operator sees what to uncomment
  - Included container-internal vars (Roundcube, SnappyMail, Radicale) as commented sections for completeness
patterns_established:
  - .env.example format: section headers with dashes, Required/Optional markers, source cross-references, defaults as values
observability_surfaces:
  - none (documentation files only)
duration: 10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Create .env.example files for cloud-relay and home-device

**Created .env.example files documenting all environment variables for both deployment targets with grouped sections, defaults, and required/optional markers.**

## What Happened

Extracted all environment variables from three sources:
- `cloud-relay/relay/config/config.go`: 22 vars (relay core, mTLS, TLS monitoring, queue, S3 overflow)
- `cloud-relay/docker-compose.yml`: 6 additional vars (RELAY_HOSTNAME, RELAY_DOMAIN, RELAY_EPHEMERAL_CHECK_INTERVAL, WEBMAIL_DOMAINS, AUTOCONFIG_DOMAINS, AUTODISCOVER_DOMAINS, CERTBOT_EMAIL)
- `home-device/docker-compose.yml`: vars across all profile services (stalwart, maddy, postfix-dovecot, roundcube, snappymail, radicale, profile-server)

Cloud-relay .env.example has 6 sections: Relay Core, mTLS Transport, TLS Monitoring, Queue, S3 Overflow, Caddy Reverse Proxy.

Home-device .env.example has 8 sections: Domain & Identity, User Configuration, Mail Server Selection, Profile Server, Monitoring & Alerting, Webmail (Roundcube), Webmail (SnappyMail), Radicale.

## Verification

- `test -f cloud-relay/.env.example && test -f home-device/.env.example` — PASS
- `grep -q "RELAY_" cloud-relay/.env.example` — PASS
- `grep -c "=" cloud-relay/.env.example` → 35 (≥20 required) — PASS
- All 22 config.go vars present — PASS (verified each individually)
- All compose-level vars present in appropriate files — PASS
- Section headers present in both files — PASS

Slice-level checks (T01-relevant):
- Both .env.example files exist — PASS
- Cloud-relay vars documented — PASS

## Diagnostics

Read `.env.example` files to understand available configuration. Each var has a comment indicating Required/Optional status and which source file defines it.

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `cloud-relay/.env.example` — complete env var documentation for cloud relay (35 var lines, 6 sections)
- `home-device/.env.example` — complete env var documentation for home device (21 var lines, 8 sections)
