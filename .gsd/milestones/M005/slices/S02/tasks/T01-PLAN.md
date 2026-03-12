---
estimated_steps: 5
estimated_files: 5
---

# T01: Fix cloud relay transport map and outbound relay configs for all mail server profiles

**Slice:** S02 — Email Round-Trip — Inbound & Outbound Delivery
**Milestone:** M005

## Description

The cloud relay transport map uses a wildcard `* smtp:[127.0.0.1]:10025` that routes ALL mail to the Go relay daemon — including outbound mail from the home device. This creates a routing loop: home sends outbound to cloud relay, cloud relay sends it back to home via relay daemon. Additionally, none of the three home mail server profiles have a relayhost configured, so they attempt direct internet delivery from the residential IP (blocked by ISPs, fails SPF).

This task fixes both problems: domain-specific transport on cloud relay, and relayhost config on all three home profiles.

## Steps

1. **Replace cloud relay transport wildcard with domain-specific entries.** Edit `cloud-relay/postfix-config/transport` to replace `* smtp:[127.0.0.1]:10025` with `${RELAY_DOMAIN} smtp:[127.0.0.1]:10025` (and comment explaining that additional domains use the same pattern). Update the file header comments to explain the routing logic: user domains → relay daemon → home device; all other destinations → normal internet SMTP delivery.

2. **Update cloud relay main.cf comments.** Adjust the transport section comments in `cloud-relay/postfix-config/main.cf` to reflect the new domain-specific routing behavior instead of "all mail goes to relay daemon."

3. **Add relayhost to Postfix-Dovecot profile.** Edit `home-device/postfix-dovecot/postfix/main.cf` to add `relayhost = [10.8.0.1]:25` — routing all outbound mail through the cloud relay via WireGuard tunnel. Add comments explaining why (residential IP port 25 blocked, SPF alignment requires cloud relay IP).

4. **Configure Maddy to relay outbound through cloud relay.** Edit `home-device/maddy/maddy.conf` to change the `target.remote remote_queue` block to use `[10.8.0.1]:25` as the MX override / smarthost destination instead of direct delivery.

5. **Configure Stalwart outbound relay.** Edit `home-device/stalwart/config.toml` to configure the `[queue.outbound]` section with next-hop or relay pointing to `[10.8.0.1]:25` for all outbound mail.

## Must-Haves

- [ ] Cloud relay transport map uses domain-specific entries, NOT wildcard `*`
- [ ] Postfix-Dovecot has `relayhost = [10.8.0.1]:25`
- [ ] Maddy routes outbound through cloud relay at 10.8.0.1:25
- [ ] Stalwart routes outbound through cloud relay at 10.8.0.1:25
- [ ] Cloud relay `mynetworks` includes 10.8.0.0/24 (already present, verify not removed)
- [ ] No mail routing loop possible (outbound from home hits cloud relay, cloud relay delivers to internet, NOT back to relay daemon)

## Verification

- `grep -c '^\*' cloud-relay/postfix-config/transport` returns 0 (no wildcard)
- `grep 'smtp:\[127.0.0.1\]:10025' cloud-relay/postfix-config/transport` shows domain-specific line(s)
- `grep 'relayhost' home-device/postfix-dovecot/postfix/main.cf` shows `[10.8.0.1]:25`
- `grep -i 'relay\|10.8.0.1' home-device/maddy/maddy.conf` shows smarthost config
- `grep -i 'relay\|10.8.0.1' home-device/stalwart/config.toml` shows relay config

## Observability Impact

- Signals added/changed: Cloud relay Postfix logs will now show different routing for user-domain vs external-domain mail; transport map decisions visible in `postcat` / `maillog`
- How a future agent inspects this: Check transport map file for expected domain entries; check Postfix logs for `transport:` routing decisions
- Failure state exposed: Mail to external domains stuck in cloud relay queue if transport map still routes to relay daemon; visible via `postqueue -p`

## Inputs

- `cloud-relay/postfix-config/transport` — current wildcard transport map
- `cloud-relay/postfix-config/main.cf` — cloud relay Postfix config with `mynetworks` and `transport_maps`
- `home-device/postfix-dovecot/postfix/main.cf` — home Postfix config, no relayhost
- `home-device/maddy/maddy.conf` — Maddy config with direct delivery
- `home-device/stalwart/config.toml` — Stalwart config with no relay
- S02-RESEARCH.md — routing gap analysis and correct config patterns

## Expected Output

- `cloud-relay/postfix-config/transport` — domain-specific entries replacing wildcard
- `cloud-relay/postfix-config/main.cf` — updated comments
- `home-device/postfix-dovecot/postfix/main.cf` — relayhost added
- `home-device/maddy/maddy.conf` — smarthost relay configured
- `home-device/stalwart/config.toml` — outbound relay configured
