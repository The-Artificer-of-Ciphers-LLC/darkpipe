# Mail Round-Trip Testing Procedure

End-to-end verification of bidirectional email delivery through the DarkPipe chain.

- **Outbound:** home server → WireGuard tunnel → cloud relay → external recipient
- **Inbound:** external sender → cloud relay → WireGuard tunnel → home mailbox

## Prerequisites

Before running round-trip tests, ensure the following are in place:

| Prerequisite | How to verify |
|---|---|
| WireGuard tunnel operational | `scripts/validate-infrastructure.sh` — tunnel section passes |
| DNS records configured (MX, SPF, DKIM, DMARC) | `scripts/validate-infrastructure.sh` — dns section passes |
| TLS certificates provisioned | `scripts/validate-infrastructure.sh` — tls section passes |
| Mail server profile running on home device | `docker compose --profile <profile> ps` shows healthy containers |
| Cloud relay Postfix accepting mail | Port 25 reachable on cloud relay public IP |
| `swaks` installed | `which swaks` — install via `apt install swaks` or `brew install swaks` |

### Pre-Flight Checks

Run the mail infrastructure validation before attempting round-trip tests:

```bash
RELAY_DOMAIN=yourdomain.com bash scripts/lib/validate-mail.sh --verbose
```

This checks:
- IP blocklist status (relay IP not listed on major DNSBLs)
- DKIM DNS record presence and format
- Transport map has domain-specific entries (no wildcard routing loops)
- Outbound relay configuration present for active mail server profile
- Rspamd DKIM signing configuration correct

Fix any failures before proceeding — they will cause authentication failures in round-trip tests.

## Outbound Test Procedure

### Using the Helper Script

```bash
bash scripts/test-mail-roundtrip.sh \
  --domain yourdomain.com \
  --recipient yourname@gmail.com \
  --sender alice \
  --verbose
```

The script will:
1. Run pre-flight infrastructure checks
2. Send a test email via SMTP submission (port 587) with STARTTLS and authentication
3. Poll mail logs for delivery confirmation
4. Print header verification instructions

### Manual Outbound Test

If you prefer manual testing or `swaks` is unavailable:

```bash
# Send via swaks
swaks --to recipient@gmail.com \
  --from alice@yourdomain.com \
  --server localhost:587 \
  --tls \
  --auth-user alice@yourdomain.com \
  --auth-password 'password' \
  --header "Subject: DarkPipe Outbound Test $(date +%s)" \
  --body "Test email from DarkPipe mail server"
```

### Expected Results — Outbound

1. **swaks output:** `250 2.0.0 Ok: queued` — message accepted by submission server
2. **Home device mail log:** `status=sent` entry showing relay to cloud relay (10.8.0.1)
3. **Cloud relay mail log:** `status=sent` entry showing delivery to external MX
4. **Recipient inbox:** Message arrives (may take 1–5 minutes; check spam folder)

### Verifying Authentication Headers

Open the received email's full headers (raw source) and find the `Authentication-Results` header.

**How to view full headers:**
- **Gmail:** Open message → ⋮ menu → "Show original"
- **Outlook:** Open message → ⋯ menu → "View message source"
- **Apple Mail:** View → Message → All Headers
- **ProtonMail:** Open message → ⋯ menu → "View headers"

**What passing results look like:**

```
Authentication-Results: mx.google.com;
    dkim=pass header.i=@yourdomain.com header.s=darkpipe;
    spf=pass (domain of alice@yourdomain.com designates <relay-ip> as permitted sender);
    dmarc=pass (p=REJECT sp=REJECT) header.from=yourdomain.com
```

| Field | Pass | Fail | What it means |
|---|---|---|---|
| `spf` | `pass` | `fail` / `softfail` | Sending IP authorized by SPF DNS record |
| `dkim` | `pass` | `fail` / `none` | DKIM signature verified against DNS public key |
| `dmarc` | `pass` | `fail` / `none` | DMARC policy evaluation (requires SPF + DKIM alignment) |

All three should show `pass` for fully authenticated delivery.

## Inbound Test Procedure

### Sending the Test Email

From an external email account (Gmail, Outlook, ProtonMail, etc.), send an email to your DarkPipe address:

```
To: alice@yourdomain.com
Subject: Inbound Test <timestamp>
Body: Testing inbound delivery to DarkPipe
```

### Verifying Delivery

**Via IMAP (command-line):**
```bash
curl --insecure "imaps://localhost:993/INBOX" \
  --user "alice@yourdomain.com:password" \
  --request "SEARCH UNSEEN"
```

A response containing `SEARCH 1 2 3` (message sequence numbers) means messages are present.

**Via webmail:** If Roundcube or another webmail client is configured, log in and check INBOX.

**Via Maildir (direct filesystem):**
```bash
ls -lt ~/Maildir/new/ | head -5
```

### Expected Results — Inbound

1. **Cloud relay mail log:** Inbound message routed via transport map to home device (10.8.0.2)
2. **Home device mail log:** Message accepted and delivered to local mailbox
3. **Mailbox:** Message appears in INBOX within 1–5 minutes

## Troubleshooting

### Message in Spam/Junk Folder

**Symptoms:** Email delivered but lands in spam instead of inbox.

**Causes and fixes:**
- **New domain/IP:** Fresh domains and IPs have no reputation. Send legitimate email for a few weeks to build reputation. Consider starting with `p=none` DMARC policy.
- **SPF softfail:** Check that the relay IP is included in the SPF record: `dig TXT yourdomain.com | grep spf`
- **Missing DKIM:** Verify DKIM signing is active: check Rspamd logs for `dkim_signing` events.
- **Content triggers:** Avoid spam-like content in test messages (all caps, excessive links).

### Greylisting Delays

**Symptoms:** First email to a recipient takes 5–15 minutes; subsequent emails arrive quickly.

**Explanation:** Some receiving servers temporarily reject mail from unknown sender/IP pairs. Postfix retries automatically. This is normal for first-time delivery to a new recipient.

**What to check:**
```bash
# Check for deferred messages on cloud relay
postqueue -p

# Look for 4xx temporary rejections in logs
sudo grep 'status=deferred' /var/log/mail.log | tail -10
```

### Blocklist Hits

**Symptoms:** Mail rejected with 5xx errors mentioning blocklist or RBL.

**Diagnosis:**
```bash
RELAY_DOMAIN=yourdomain.com bash scripts/lib/validate-mail.sh --verbose
# Check blocklist section output
```

**Fixes:**
- Check your relay IP on [MXToolbox Blocklist Check](https://mxtoolbox.com/blacklists.aspx)
- Request delisting from the specific blocklist
- If using a VPS provider, contact support — shared IP ranges sometimes get listed

### DKIM Failures

**Symptoms:** `dkim=fail` in Authentication-Results header.

**Causes:**
- **Key mismatch:** The DNS public key doesn't match the private key used for signing
  ```bash
  # Check DNS record
  dig TXT darkpipe._domainkey.yourdomain.com
  
  # Check Rspamd signing config
  cat home-device/rspamd/local.d/dkim_signing.conf
  ```
- **Selector mismatch:** Signing with selector X but DNS record is under selector Y
- **Message modification:** An intermediary (mailing list, forwarding) altered the message body after signing
- **Missing key file:** Rspamd can't find the private key at the expected path
  ```bash
  # Check Rspamd logs for key path errors
  docker logs rspamd 2>&1 | grep -i dkim
  ```

### SPF Failures

**Symptoms:** `spf=fail` or `spf=softfail` in Authentication-Results.

**Causes:**
- **Relay IP not in SPF record:** The cloud relay's public IP must be in the SPF TXT record
  ```bash
  dig TXT yourdomain.com | grep spf
  # Should include the relay IP or an include: that covers it
  ```
- **Multiple SPF records:** DNS should have exactly one SPF TXT record per domain
- **SPF lookup limit:** SPF allows max 10 DNS lookups; too many `include:` directives cause `permerror`

### Relay Denied Errors

**Symptoms:** `relay access denied` in mail logs.

**Causes:**
- **Cloud relay transport map:** Missing or incorrect entry for your domain
  ```bash
  cat cloud-relay/postfix-config/transport
  # Should have: yourdomain.com smtp:[10.8.0.2]:25
  ```
- **Home server not accepting relay:** Check that Postfix `mynetworks` includes the WireGuard subnet
- **Tunnel down:** Verify WireGuard connectivity
  ```bash
  ping -c 3 10.8.0.2   # From cloud relay
  ping -c 3 10.8.0.1   # From home device
  ```

### Delivery Timeout (No status=sent)

**Symptoms:** Script times out without finding delivery confirmation in logs.

**Diagnosis:**
```bash
# Check mail queue for stuck messages
postqueue -p

# Check recent log entries
sudo tail -50 /var/log/mail.log

# Look for connection errors
sudo grep 'connect to\|Connection refused\|Connection timed out' /var/log/mail.log | tail -10
```

**Common causes:**
- Tunnel down (connection refused/timed out to 10.8.0.x)
- Recipient MX server rejecting connections (check their DNS)
- Postfix paused (`postfix status` / `systemctl status postfix`)

## Reference

| Resource | Purpose |
|---|---|
| `scripts/test-mail-roundtrip.sh` | Automated helper script (this procedure) |
| `scripts/lib/validate-mail.sh` | Pre-flight infrastructure checks |
| `scripts/validate-infrastructure.sh` | Full infrastructure validation orchestrator |
| `home-device/tests/test-mail-flow.sh` | Local mail flow integration tests (no external delivery) |
| `dns/authtest/sender.go` | Go-based DKIM test email sender |
