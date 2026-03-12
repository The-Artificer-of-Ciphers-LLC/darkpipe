---
id: T02
parent: S02
milestone: M005
provides:
  - Rspamd DKIM signing configuration for all mail server profiles
  - DKIM private key volume mount into Rspamd container
  - Key provisioning lifecycle documentation
key_files:
  - home-device/spam-filter/rspamd/local.d/dkim_signing.conf
  - home-device/docker-compose.yml
  - home-device/docker-compose.podman-selinux.yml
  - home-device/spam-filter/rspamd/dkim-keys/README.md
key_decisions:
  - Use Rspamd dkim_signing module (not Go signer) as the signing integration point — works identically across all three mail server profiles
  - Key path convention /var/lib/rspamd/dkim/${domain}.${selector}.key matches Rspamd template syntax for automatic domain resolution
  - sign_headers explicitly lists same headers as dns/dkim/signer.go for consistency
patterns_established:
  - DKIM keys stored in host directory mounted read-only into Rspamd container
  - Signing scoped to authenticated/local/WireGuard traffic only — inbound mail from port 25 passes through unsigned
observability_surfaces:
  - Rspamd logs show dkim_signing module events (success with selector/domain, failure with missing key path)
  - Rspamd web UI (port 11334) exposes DKIM signing statistics
  - Missing key file produces Rspamd error log with expected path for easy diagnosis
duration: 12min
verification_result: passed
completed_at: 2026-03-12
blocker_discovered: false
---

# T02: Configure Rspamd DKIM signing and provision key mounting

**Created Rspamd DKIM signing config with scoped outbound-only signing and provisioned key directory mounting across all compose variants.**

## What Happened

Created `dkim_signing.conf` in the Rspamd local.d directory with signing scoped to authenticated submission (port 587), local Docker network, and WireGuard tunnel (10.8.0.0/24). Inbound mail from external sources on port 25 is explicitly not signed — documented with comments explaining the scoping logic.

The DKIM selector defaults to `darkpipe`, domain is extracted from the From: header, and the key path uses Rspamd's template syntax (`/var/lib/rspamd/dkim/${domain}.${selector}.key`) for automatic per-domain key resolution.

Added volume mount for the `dkim-keys` host directory into the Rspamd container at `/var/lib/rspamd/dkim:ro` in both the base `docker-compose.yml` and the `docker-compose.podman-selinux.yml` (with `:z` SELinux label). The podman base override file needed no changes (it only sets compose compat flags, not volumes).

Created the `dkim-keys` directory with a README documenting key naming convention and provisioning steps. The directory is `.gitkeep`'d since actual keys should not be committed to version control.

Canonicalization (relaxed/relaxed) and signed headers match the Go DKIM signer in `dns/dkim/signer.go` for consistency.

## Verification

All 10 must-have checks passed:
- `dkim_signing.conf` exists with `sign_authenticated = true`, `sign_local = true`, `sign_networks` containing `10.8.0.0/24`
- Key path convention `/var/lib/rspamd/dkim/${domain}.${selector}.key` present in config
- Volume mount for dkim-keys in `docker-compose.yml` and `docker-compose.podman-selinux.yml`
- Key provisioning lifecycle documented in config comments and README
- Inbound mail NOT signed — scoping comments present
- `dkim-keys` directory exists with README

Slice-level checks (partial — this is T02 of slice):
- ✅ DKIM config exists and is consistent
- ✅ docker-compose.yml has dkim volume mount
- ✅ Outbound relay configs from T01 still intact (relayhost present in postfix main.cf)
- ⏳ `scripts/lib/validate-mail.sh` — not yet created (future task)
- ⏳ Manual send/receive tests — future task

## Diagnostics

- **DKIM signing events:** Check Rspamd logs for `dkim_signing` module output — shows domain, selector, and signing result per message
- **Missing key:** Rspamd logs error with expected key path (e.g., `/var/lib/rspamd/dkim/example.com.darkpipe.key not found`)
- **Signing stats:** Rspamd web UI at port 11334 shows signing success/failure counts
- **Key mismatch:** Recipient's Authentication-Results header shows `dkim=fail` with reason

## Deviations

- Also updated `docker-compose.podman-selinux.yml` for SELinux compatibility — not explicitly in task plan but necessary for consistency across all compose variants
- Created `dkim-keys/README.md` with provisioning instructions — supplements the config file comments

## Known Issues

None.

## Files Created/Modified

- `home-device/spam-filter/rspamd/local.d/dkim_signing.conf` — new: Rspamd DKIM signing configuration
- `home-device/docker-compose.yml` — updated: added DKIM key volume mount to Rspamd service
- `home-device/docker-compose.podman-selinux.yml` — updated: added DKIM key volume mount with SELinux label
- `home-device/spam-filter/rspamd/dkim-keys/.gitkeep` — new: placeholder for key directory
- `home-device/spam-filter/rspamd/dkim-keys/README.md` — new: key provisioning documentation
