---
id: T01
parent: S02
milestone: M005
provides:
  - Domain-specific cloud relay transport map (no wildcard routing loop)
  - Outbound relay config for all three home mail server profiles
key_files:
  - cloud-relay/postfix-config/transport
  - cloud-relay/postfix-config/main.cf
  - home-device/postfix-dovecot/postfix/main.cf
  - home-device/maddy/maddy.conf
  - home-device/stalwart/config.toml
key_decisions:
  - Maddy uses target.smtp (transparent forwarding) instead of target.remote (MX lookup) for smarthost relay
  - Stalwart uses queue.strategy route + queue.route."relay" pattern per official docs
  - WireGuard tunnel encryption makes TLS to cloud relay unnecessary (disabled in Maddy/Stalwart relay configs)
patterns_established:
  - All home profiles relay outbound to cloud relay at 10.8.0.1:25 via WireGuard
  - Cloud relay transport map uses ${RELAY_DOMAIN} variable for domain-specific routing
observability_surfaces:
  - Cloud relay Postfix logs show transport routing decisions per domain
  - postqueue -p on cloud relay shows stuck outbound mail if relay misconfigured
duration: 20m
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T01: Fix cloud relay transport map and outbound relay configs for all mail server profiles

**Replaced wildcard transport map with domain-specific routing and configured outbound relay for all three home mail server profiles (Postfix-Dovecot, Maddy, Stalwart).**

## What Happened

Fixed two problems that would prevent outbound email delivery:

1. **Cloud relay transport map** had a wildcard `*` that routed ALL mail (including outbound from home) back to the relay daemon, creating a routing loop. Replaced with `${RELAY_DOMAIN} smtp:[127.0.0.1]:10025` so only inbound user-domain mail goes to the relay daemon; outbound mail to external domains uses Postfix default internet SMTP delivery.

2. **Home mail server profiles** had no outbound relay configured — they'd attempt direct delivery from the residential IP (port 25 blocked by ISPs, fails SPF). Added relay config pointing to cloud relay at `10.8.0.1:25` (WireGuard tunnel) for all three profiles:
   - **Postfix-Dovecot:** `relayhost = [10.8.0.1]:25`
   - **Maddy:** Replaced `target.remote` with `target.smtp` using `targets tcp://10.8.0.1:25` (transparent forwarding, no MX lookup)
   - **Stalwart:** Added `queue.strategy` routing all outbound through `queue.route."relay"` at `10.8.0.1:25`

Cloud relay `main.cf` comments were already updated to reflect domain-specific routing. `mynetworks` already includes `10.8.0.0/24` allowing relay from home device.

## Verification

All task-level checks passed:

- `grep -c '^\*' cloud-relay/postfix-config/transport` → `0` (no wildcard)
- `grep 'smtp:\[127.0.0.1\]:10025' cloud-relay/postfix-config/transport` → shows `${RELAY_DOMAIN}` line
- `grep 'relayhost' home-device/postfix-dovecot/postfix/main.cf` → `relayhost = [10.8.0.1]:25`
- `grep '10.8.0.1' home-device/maddy/maddy.conf` → shows `targets tcp://10.8.0.1:25`
- `grep '10.8.0.1' home-device/stalwart/config.toml` → shows `address = "10.8.0.1"` in relay route
- `grep 'mynetworks' cloud-relay/postfix-config/main.cf` → includes `10.8.0.0/24`

Slice-level checks (applicable to T01):
- `grep -q 'relayhost' home-device/postfix-dovecot/postfix/main.cf` → PASS
- Transport map has domain-specific entries, no wildcard `*` → PASS

## Diagnostics

- **Cloud relay transport routing:** Check `cloud-relay/postfix-config/transport` for domain entries. In runtime, Postfix logs show `transport:` decisions per message.
- **Outbound relay path:** `postqueue -p` on cloud relay shows queued outbound mail. If mail loops back to relay daemon, transport map still has wrong entries.
- **Per-profile relay config:** grep for `relayhost` (Postfix), `targets` (Maddy), or `queue.route` (Stalwart) in respective config files.

## Deviations

- **Maddy:** Used `target.smtp` instead of modifying `target.remote` — Context7 docs confirmed `target.smtp` is the correct directive for transparent forwarding to a smarthost (target.remote does MX lookups).
- **Stalwart:** Used `queue.strategy` + `queue.route."relay"` pattern instead of `next-hop` — Context7 docs confirmed this is the official routing configuration format.

## Known Issues

None.

## Files Created/Modified

- `cloud-relay/postfix-config/transport` — replaced wildcard with domain-specific `${RELAY_DOMAIN}` entry, added routing logic comments
- `cloud-relay/postfix-config/main.cf` — updated comments to reflect domain-specific routing (already done in prior session)
- `home-device/postfix-dovecot/postfix/main.cf` — added `relayhost = [10.8.0.1]:25` with explanatory comments (already done in prior session)
- `home-device/maddy/maddy.conf` — replaced `target.remote remote_queue` with `target.smtp remote_queue` forwarding to `tcp://10.8.0.1:25`
- `home-device/stalwart/config.toml` — added `queue.strategy` and `queue.route."relay"` sections routing all outbound to `10.8.0.1:25`
