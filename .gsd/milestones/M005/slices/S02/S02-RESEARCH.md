# S02: Email Round-Trip — Inbound & Outbound Delivery — Research

**Date:** 2026-03-12

## Summary

S02 must prove bidirectional email delivery through the full DarkPipe chain: external sender → cloud relay → WireGuard tunnel → home mailbox (inbound), and home mail client → SMTP submission → home server → cloud relay → external recipient (outbound). The codebase has strong inbound routing (cloud relay Postfix → Go relay daemon → WireGuard → home device port 25), but **outbound mail has a critical routing gap**: none of the three home mail server profiles (Stalwart, Maddy, Postfix-Dovecot) are configured to relay outbound mail through the cloud relay. They attempt direct delivery to the internet, which will fail from residential IPs (ISP port 25 blocks, poor IP reputation, SPF mismatch). Additionally, DKIM signing is implemented as a Go library (`dns/dkim/signer.go`) but is not wired into the mail pipeline for any mail server profile.

The existing `dns/authtest` package provides a test email sender with DKIM signing and an Authentication-Results header parser — ideal for building the verification step. The `scripts/validate-infrastructure.sh` framework from S01 provides the orchestration pattern for validation scripts. The `home-device/tests/test-mail-flow.sh` script tests local mail flow but does not cover the external round-trip path.

## Recommendation

Structure S02 as three phases:

1. **Fix outbound routing** — Configure all three home mail server profiles to relay outbound mail through the cloud relay via WireGuard tunnel. Fix the cloud relay transport map (`* smtp:[127.0.0.1]:10025`) to only route inbound domains to the relay daemon, allowing outbound mail to flow directly to the internet from the relay's public IP.
2. **Wire DKIM signing** — Configure Rspamd DKIM signing (preferred for all profiles) or use mail-server-native signing, so outbound mail gets DKIM headers before leaving the cloud relay.
3. **Build round-trip verification** — Create a validation script that sends a test email outbound, checks delivery, sends inbound, checks arrival, and validates SPF/DKIM/DMARC pass on both directions. Include IP blocklist checking.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| DKIM signing | `dns/dkim/signer.go` (wraps `emersion/go-msgauth`) | Already tested, RFC 6376 compliant, relaxed/relaxed canonicalization |
| Auth-Results parsing | `dns/authtest/parser.go` (wraps `emersion/go-msgauth/authres`) | RFC 8601 compliant, already extracts SPF/DKIM/DMARC pass/fail |
| Test email sending | `dns/authtest/sender.go` | Sends DKIM-signed test emails via SMTP relay, includes verification instructions |
| DKIM signing in mail pipeline | Rspamd `dkim_signing` module | Built-in to Rspamd (already running), works for all 3 mail server profiles |
| Infrastructure validation | `scripts/validate-infrastructure.sh` | S01 established section runner pattern — add a `validate-mail.sh` section |
| Blocklist checking | DNS-based DNSBL queries | Standard `dig` against zen.spamhaus.org, b.barracudacentral.org etc. — no external tool needed |

## Existing Code and Patterns

- `cloud-relay/postfix-config/main.cf` — Cloud relay Postfix config. `transport_maps = lmdb:/etc/postfix/transport` routes ALL mail to relay daemon. Must be changed to route only user domains inbound; outbound mail from `mynetworks` (10.8.0.0/24) should flow to the internet directly.
- `cloud-relay/postfix-config/transport` — Currently `* smtp:[127.0.0.1]:10025` (wildcard). Needs domain-specific entries instead: `example.com smtp:[127.0.0.1]:10025` so outbound to external domains uses normal SMTP.
- `home-device/postfix-dovecot/postfix/main.cf` — No `relayhost` configured. Needs `relayhost = [10.8.0.1]:25` to route outbound through cloud relay via WireGuard.
- `home-device/maddy/maddy.conf` — `target.remote remote_queue` delivers directly to remote MTAs. Needs `relay` or `remote_mta` configuration to use cloud relay as smarthost.
- `home-device/stalwart/config.toml` — `[queue.outbound]` has `hostname` but no relay/smarthost config. Needs `next-hop` or equivalent to route through cloud relay.
- `home-device/spam-filter/rspamd/local.d/` — No `dkim_signing.conf` exists. Need to add DKIM signing configuration pointing to the private key.
- `dns/dkim/keygen.go` — DKIM key generation utility. Keys need to be provisioned and mounted into Rspamd/mail server containers.
- `home-device/tests/test-mail-flow.sh` — Local mail flow tests using swaks. Pattern to follow for external round-trip script, but uses localhost targets only.
- `scripts/validate-infrastructure.sh` + `scripts/lib/validate-*.sh` — Section runner pattern from S01. Add `validate-mail.sh` for email round-trip validation.
- `cloud-relay/docker-compose.yml` — Relay service config with `mynetworks: 10.8.0.0/24` already allowing WireGuard-originated mail to relay.

## Constraints

- **Live infrastructure required** — Round-trip testing needs a real VPS, real home device, real DNS, and real external email accounts (Gmail/Outlook). Cannot be simulated in CI.
- **Residential IP port 25 blocked** — Most ISPs block outbound port 25 from residential networks. Outbound mail MUST route through the cloud relay's public IP.
- **SPF alignment** — SPF records point to the cloud relay IP. Outbound mail must originate from the cloud relay IP or SPF will fail.
- **Transport map change affects inbound routing** — Changing `*` wildcard to domain-specific entries means the cloud relay must know which domains to route to the home device. This comes from `RELAY_DOMAIN` env var.
- **Three mail server profiles** — Outbound relay config differs for each profile (Postfix relayhost, Maddy remote target, Stalwart next-hop). All three must be updated.
- **DKIM key lifecycle** — Private key must exist and be mounted into the signing component (Rspamd or mail server). Key must match the DNS TXT record published in S01.
- **Cloud relay Postfix must not loop** — When home device sends outbound to cloud relay, the transport map must NOT route it back to the relay daemon (which would send it back to home device). Domain-specific transport entries solve this.

## Common Pitfalls

- **Mail routing loop** — If cloud relay transport map uses `*` wildcard and home device uses cloud relay as relayhost, outbound mail gets forwarded back to home device indefinitely. Fix: use domain-specific transport entries, not wildcard.
- **SPF softfail on outbound** — If outbound mail leaves from home device IP instead of cloud relay IP, SPF will fail. All outbound MUST transit through cloud relay.
- **DKIM key mismatch** — If the DKIM private key in Rspamd doesn't match the public key in DNS, DKIM verification fails at the recipient. Validate key pair before testing.
- **Rspamd DKIM signing on inbound** — Rspamd should only sign outbound mail (from authenticated users), not inbound. Use `sign_authenticated = true; sign_local = true; sign_networks = "10.8.0.0/24";` in Rspamd config to scope signing.
- **Greylisting delays** — Rspamd greylisting (5-min delay, score >= 4.0) may delay inbound test mail. Private network whitelist (10.8.0.0/24) should bypass greylisting for relay-originated mail, but verify.
- **DNS propagation** — If S01's DNS records haven't fully propagated, SPF/DKIM/DMARC checks at external providers may fail. Verify records resolve from 8.8.8.8 before testing.
- **Cloud relay TLS certificates** — If Let's Encrypt certs aren't provisioned, the cloud relay may accept but not properly present TLS for outbound SMTP. Check `smtp_tls_security_level = may` on cloud relay.

## Open Risks

- **VPS IP on blocklists** — New VPS IPs may be listed on Spamhaus, Barracuda, or other DNSBLs. Outbound mail would be rejected by recipients. Must check blocklists before testing delivery. MXToolbox or direct DNSBL queries needed.
- **Gmail/Outlook strict filtering** — Major providers may delay or spam-folder mail from new/low-reputation IPs even with correct SPF/DKIM/DMARC. First test emails may land in spam.
- **Stalwart outbound relay config** — Stalwart 0.15.4's smarthost/relay configuration syntax needs verification against current docs. The `next-hop` or equivalent may have changed between versions.
- **Maddy smarthost config** — Maddy's `target.remote` block needs `relay` or similar directive for smarthost delivery. Syntax needs verification.
- **DKIM key provisioning across profiles** — Each mail server profile mounts volumes differently. The DKIM key path must be consistent and accessible to whichever component signs (Rspamd or mail server).
- **Test email accounts** — Need a real external email account (Gmail, Outlook) both to send inbound test mail and to receive outbound test mail. These are user-provided, not automatable.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Email/SMTP systems | `sickn33/antigravity-awesome-skills@email-systems` | available (281 installs) — may have relevant Postfix/DKIM patterns |
| Agent email CLI | `zaddy6/agent-email-skill@agent-email-cli` | available (20.7K installs) — general email agent skill, less relevant to mail server config |
| Postfix/Dovecot | — | none found |
| Rspamd DKIM | — | none found |
| Stalwart mail | — | none found |

No installed skills are directly relevant to this slice's mail server configuration work. The `email-systems` skill (281 installs) might contain useful Postfix/DKIM patterns but low install count suggests limited coverage.

## Sources

- Outbound routing gap identified by code inspection: `home-device/postfix-dovecot/postfix/main.cf` has no `relayhost`, `home-device/maddy/maddy.conf` delivers directly to remote, `home-device/stalwart/config.toml` has no smarthost config
- Transport map loop risk identified from `cloud-relay/postfix-config/transport` wildcard `*` entry combined with WireGuard `mynetworks` relay permission
- DKIM signing gap identified: `dns/dkim/signer.go` exists but no Rspamd `dkim_signing.conf` or mail-server-native DKIM config found
- S01 validation framework pattern from task summaries (T01-T04) in `.gsd/milestones/M005/slices/S01/tasks/`
- Existing test tooling from `dns/authtest/sender.go` and `home-device/tests/test-mail-flow.sh`
