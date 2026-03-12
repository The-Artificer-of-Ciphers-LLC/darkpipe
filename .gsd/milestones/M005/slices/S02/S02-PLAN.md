# S02: Email Round-Trip — Inbound & Outbound Delivery

**Goal:** Prove bidirectional email delivery through the full DarkPipe chain: external → cloud relay → tunnel → home mailbox (inbound), and home → tunnel → cloud relay → external recipient (outbound), with SPF/DKIM/DMARC passing on outbound.

**Demo:** Send an email from an external provider to the home mailbox and confirm arrival. Send an email from the home mail server to an external recipient and confirm delivery with SPF/DKIM/DMARC pass.

## Must-Haves

- Cloud relay transport map routes only user domains to relay daemon (no wildcard), allowing outbound mail to flow to internet
- All three home mail server profiles (Postfix-Dovecot, Maddy, Stalwart) configured to relay outbound through cloud relay via WireGuard
- DKIM signing configured in Rspamd for outbound mail (applies to all profiles)
- DKIM private key provisioned and mounted into Rspamd container
- Round-trip validation script proving inbound delivery arrives in mailbox
- Round-trip validation script proving outbound delivery reaches external recipient
- SPF/DKIM/DMARC pass confirmation on outbound mail headers
- IP blocklist check for the cloud relay VPS IP
- No mail routing loops between cloud relay and home device

## Proof Level

- This slice proves: integration
- Real runtime required: yes (live VPS, home device, real DNS, real external email accounts)
- Human/UAT required: yes (external email account interaction to verify delivery and check headers)

## Verification

- `scripts/lib/validate-mail.sh` — mail validation section (blocklist checks, DKIM key consistency, relay config checks)
- `scripts/validate-infrastructure.sh` — orchestration includes mail section
- Manual: send test email outbound from home server, verify delivery + SPF/DKIM/DMARC pass in recipient headers
- Manual: send test email inbound from external provider, verify arrival in home mailbox
- Config verification: `grep -q 'relayhost' home-device/postfix-dovecot/postfix/main.cf` confirms outbound relay
- Config verification: transport map has domain-specific entries, NOT wildcard `*`

## Observability / Diagnostics

- Runtime signals: Postfix mail logs (`/var/log/mail.log`) on both cloud relay and home device show routing decisions; Rspamd logs show DKIM signing events
- Inspection surfaces: `scripts/lib/validate-mail.sh --verbose` outputs per-check diagnostics; `postqueue -p` shows stuck mail on either side
- Failure visibility: transport map routing errors appear in cloud relay Postfix logs; DKIM signing failures in Rspamd logs; SPF/DKIM/DMARC results in recipient email headers (Authentication-Results)
- Redaction constraints: no email content logged; test email addresses in validation script only

## Integration Closure

- Upstream surfaces consumed: S01's validated DNS (MX, SPF, DKIM TXT, DMARC), TLS certificates, WireGuard tunnel, port reachability (25, 587, 993)
- New wiring introduced in this slice: outbound relay path (home → cloud relay), DKIM signing pipeline (Rspamd), domain-specific transport routing on cloud relay
- What remains before the milestone is truly usable end-to-end: S03 device connectivity (mobile onboarding, desktop client, webmail from external network, monitoring dashboard)

## Tasks

- [x] **T01: Fix cloud relay transport map and outbound relay configs for all mail server profiles** `est:1h`
  - Why: The wildcard `*` transport map causes a routing loop for outbound mail, and no home server profile has a relayhost configured — outbound mail cannot reach external recipients
  - Files: `cloud-relay/postfix-config/transport`, `cloud-relay/postfix-config/main.cf`, `home-device/postfix-dovecot/postfix/main.cf`, `home-device/maddy/maddy.conf`, `home-device/stalwart/config.toml`
  - Do: Replace transport wildcard with domain-specific entries using `RELAY_DOMAIN`; add `relayhost = [10.8.0.1]:25` to Postfix-Dovecot; configure Maddy remote target to use cloud relay as smarthost; configure Stalwart outbound next-hop to cloud relay; verify cloud relay `mynetworks` already allows 10.8.0.0/24 relay
  - Verify: `grep -v '^\*' cloud-relay/postfix-config/transport` shows domain entries; all three profiles have relay config pointing to 10.8.0.1 (cloud relay WireGuard IP)
  - Done when: all config files have correct outbound routing and no wildcard transport remains

- [x] **T02: Configure Rspamd DKIM signing and provision key mounting** `est:45m`
  - Why: DKIM signing is required for outbound mail to pass authentication at recipients — the Go DKIM signer exists but Rspamd (which processes mail for all profiles) has no signing config
  - Files: `home-device/spam-filter/rspamd/local.d/dkim_signing.conf`, `home-device/docker-compose.yml`, `home-device/postfix-dovecot/docker-compose.yml`, `home-device/maddy/docker-compose.yml`, `home-device/stalwart/docker-compose.yml`
  - Do: Create `dkim_signing.conf` with `sign_authenticated = true; sign_local = true; sign_networks = "10.8.0.0/24"` scoping to outbound only; configure DKIM key path, selector, and domain; mount DKIM private key volume into Rspamd container across all compose profiles; document key provisioning in comments
  - Verify: `cat home-device/spam-filter/rspamd/local.d/dkim_signing.conf` shows correct signing config; compose files mount the key volume
  - Done when: Rspamd DKIM signing configured for outbound-only, key path consistent across all compose profiles

- [x] **T03: Build mail validation script and integrate into infrastructure validator** `est:1h`
  - Why: Need automated verification of mail routing configs, DKIM key consistency, and IP blocklist status — completes the validation framework started in S01
  - Files: `scripts/lib/validate-mail.sh`, `scripts/validate-infrastructure.sh`
  - Do: Create `validate-mail.sh` following S01 section runner pattern with checks: (1) IP blocklist scan against Spamhaus/Barracuda/SORBS via DNS, (2) DKIM key pair consistency (private key matches DNS TXT record), (3) transport map has domain-specific entries (no wildcard), (4) outbound relay config present in active profile, (5) Rspamd DKIM signing config exists; add mail section to orchestration script; support `--json`, `--verbose`, `--dry-run` modes
  - Verify: `bash scripts/lib/validate-mail.sh --dry-run` exits 0 with mock results; `bash scripts/validate-infrastructure.sh --dry-run` includes mail section
  - Done when: validation script runs all mail checks and integrates into infrastructure orchestrator

- [x] **T04: Document round-trip testing procedure and create test helper script** `est:45m`
  - Why: Live round-trip requires human interaction with external email accounts — need clear procedure and a helper script that automates what can be automated (sending test email, checking logs, parsing auth results)
  - Files: `scripts/test-mail-roundtrip.sh`, `docs/validation/mail-roundtrip.md`
  - Do: Create helper script that: (1) sends a DKIM-signed test email outbound via SMTP submission using `dns/authtest` sender or swaks, (2) polls Postfix logs for delivery confirmation, (3) provides instructions for checking recipient headers (SPF/DKIM/DMARC), (4) waits for inbound test reply and checks mailbox arrival; create documentation with step-by-step procedure, expected header values, and troubleshooting for common failures (spam folder, greylisting delays, blocklist hits)
  - Verify: `bash scripts/test-mail-roundtrip.sh --dry-run` shows the test sequence without sending; documentation covers both inbound and outbound paths
  - Done when: helper script and docs provide a complete, repeatable round-trip test procedure that a future operator can execute

## Files Likely Touched

- `cloud-relay/postfix-config/transport`
- `cloud-relay/postfix-config/main.cf`
- `home-device/postfix-dovecot/postfix/main.cf`
- `home-device/maddy/maddy.conf`
- `home-device/stalwart/config.toml`
- `home-device/spam-filter/rspamd/local.d/dkim_signing.conf`
- `home-device/docker-compose.yml`
- `home-device/postfix-dovecot/docker-compose.yml`
- `home-device/maddy/docker-compose.yml`
- `home-device/stalwart/docker-compose.yml`
- `scripts/lib/validate-mail.sh`
- `scripts/validate-infrastructure.sh`
- `scripts/test-mail-roundtrip.sh`
- `docs/validation/mail-roundtrip.md`
