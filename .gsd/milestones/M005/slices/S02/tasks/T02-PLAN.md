---
estimated_steps: 4
estimated_files: 6
---

# T02: Configure Rspamd DKIM signing and provision key mounting

**Slice:** S02 — Email Round-Trip — Inbound & Outbound Delivery
**Milestone:** M005

## Description

Outbound mail must carry a valid DKIM signature for recipients to verify authenticity. The Go DKIM library (`dns/dkim/signer.go`) exists but is not wired into the mail pipeline. Rspamd (already running as spam filter for all three mail server profiles) has a built-in `dkim_signing` module that can sign outbound mail — this is the correct integration point because it works identically for all profiles.

This task creates the Rspamd DKIM signing configuration and ensures the DKIM private key is volume-mounted into the Rspamd container across all compose profiles.

## Steps

1. **Create Rspamd DKIM signing configuration.** Write `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` with: `sign_authenticated = true`, `sign_local = true`, `sign_networks = "10.8.0.0/24"` (scope to outbound from authenticated users and local/WireGuard network), DKIM selector from env or default (`darkpipe`), domain extraction from `From:` header, key path pointing to `/var/lib/rspamd/dkim/${domain}.${selector}.key`, and `allow_username_mismatch = true` (users may send from any configured domain).

2. **Update Rspamd container volume mounts.** Edit the Rspamd service in the base `home-device/docker-compose.yml` to mount the DKIM private key directory into the container at `/var/lib/rspamd/dkim/`. The key file follows the naming convention `${domain}.${selector}.key` (e.g., `example.com.darkpipe.key`). Add this volume mount consistently.

3. **Verify key provisioning documentation.** Add comments in the DKIM signing config and a note in the compose file explaining the key lifecycle: (a) generate key with `dns/dkim/keygen.go`, (b) place private key at expected path, (c) publish public key as DNS TXT record (done in S01). Reference the existing `dns/dkim/keygen.go` utility.

4. **Validate no signing on inbound.** Confirm the signing config scopes correctly: `sign_authenticated = true` + `sign_local = true` + `sign_networks` ensures only mail from authenticated submission (port 587) or local/WireGuard sources gets signed. Inbound mail from the internet arriving on port 25 must NOT be re-signed. Add a comment explaining this scoping.

## Must-Haves

- [ ] `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` exists with correct signing scope
- [ ] DKIM key path convention documented and consistent (`/var/lib/rspamd/dkim/${domain}.${selector}.key`)
- [ ] Rspamd container has volume mount for DKIM key directory
- [ ] Inbound mail (port 25, external sources) is NOT signed
- [ ] Outbound mail (authenticated/local/WireGuard) IS signed
- [ ] Key provisioning lifecycle documented in config comments

## Verification

- `cat home-device/spam-filter/rspamd/local.d/dkim_signing.conf` shows signing config with correct scoping
- `grep -r 'dkim' home-device/docker-compose.yml` shows volume mount for DKIM keys
- Config has `sign_authenticated = true` and `sign_networks` containing `10.8.0.0/24`

## Observability Impact

- Signals added/changed: Rspamd logs will show DKIM signing events for outbound mail (success/failure with key path and selector)
- How a future agent inspects this: Check Rspamd web UI (port 11334) for DKIM signing statistics; check Rspamd logs for `dkim_signing` module output
- Failure state exposed: Missing key file → Rspamd logs error with expected path; key mismatch → recipient Authentication-Results shows `dkim=fail`

## Inputs

- `home-device/spam-filter/rspamd/local.d/` — existing Rspamd config directory (has actions.conf, greylist.conf, etc. but no dkim_signing.conf)
- `dns/dkim/keygen.go` — key generation utility (defines selector and key format)
- `dns/dkim/signer.go` — Go DKIM library (reference for canonicalization and signing parameters)
- `home-device/docker-compose.yml` — base compose with Rspamd service definition
- S02-RESEARCH.md — DKIM signing scope recommendations

## Expected Output

- `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` — new file, Rspamd DKIM signing config
- `home-device/docker-compose.yml` — updated with DKIM key volume mount on Rspamd service
- Potentially `home-device/postfix-dovecot/docker-compose.yml`, `home-device/maddy/docker-compose.yml`, `home-device/stalwart/docker-compose.yml` — if profile-specific compose files override Rspamd volumes
