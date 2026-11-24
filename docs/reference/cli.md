# CLI Reference

Reference for the two DarkPipe command-line tools: `dns-setup` (DNS record management) and `darkpipe-setup` (setup wizard and migration).

Both tools are provided as release assets for Linux (amd64/arm64) and macOS (Intel/Apple Silicon):

```
https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/dns-setup-<os>-<arch>
https://github.com/The-Artificer-of-Ciphers-LLC/darkpipe/releases/latest/download/darkpipe-setup-<os>-<arch>
```

where `<os>` is `linux` or `darwin` and `<arch>` is `amd64` or `arm64`.

See [How to deploy DarkPipe](../quickstart.md) for where these tools fit in a deployment, and [Migrate existing mail](../migration.md) for migration workflows.

## dns-setup

`dns-setup` generates, applies, and validates the DNS records DarkPipe requires. Built from [`dns/cmd/dns-setup`](../../dns/cmd/dns-setup/main.go).

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--domain` | *(required)* | Primary mail domain, e.g. `example.com` |
| `--relay-hostname` | *(required)* | Cloud relay FQDN, e.g. `relay.example.com` |
| `--relay-ip` | *(required)* | Cloud relay public IPv4 address |
| `--apply` | `false` | Create the records via the detected DNS provider. Without it, the run is a dry-run |
| `--validate-only` | `false` | Only check existing records; do not generate or apply |
| `--rotate-dkim` | `false` | Generate a new DKIM key pair with a new selector and print the new record |
| `--dkim-key-dir` | `/etc/darkpipe/dkim` | Directory for DKIM private/public key files |
| `--dkim-selector-prefix` | `darkpipe` | Prefix for time-based selectors (see DKIM selectors below) |
| `--dkim-key-bits` | `2048` | RSA key size in bits |
| `--dmarc-policy` | `none` | DMARC policy: `none`, `quarantine`, or `reject` |
| `--dmarc-rua` | *(empty)* | DMARC aggregate report address (`rua=`) |
| `--dmarc-ruf` | *(empty)* | DMARC forensic report address (`ruf=`) |
| `--send-test` | *(empty)* | If set to an email address, send a DKIM-signed test message to that address after setup |
| `--records-file` | `DNS-RECORDS.md` | Path of the Markdown guide written with all records for manual setup |
| `--json` | `false` | Emit machine-readable JSON instead of formatted text |
| `--verbose` | `false` | Include per-record apply details |

### Environment variables

Flags can be omitted when the corresponding environment variable is set:

| Variable | Equivalent flag |
|----------|----------------|
| `DARKPIPE_DOMAIN` | `--domain` |
| `RELAY_HOSTNAME` | `--relay-hostname` |
| `RELAY_IP` | `--relay-ip` |
| `DKIM_KEY_DIR` | `--dkim-key-dir` |
| `DMARC_RUA` | `--dmarc-rua` |
| `DMARC_RUF` | `--dmarc-ruf` |

Provider credentials are always read from the environment:

| Variable | Used by |
|----------|---------|
| `CLOUDFLARE_API_TOKEN` | Cloudflare provider (Zone.DNS.Edit permission) |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` (or any credential source from the standard AWS SDK chain) | Route53 provider |

### Mode: full setup (default)

Generates all records, prints them, writes the `--records-file` guide, and applies them only with `--apply`.

```
dns-setup --domain example.com --relay-hostname relay.example.com --relay-ip 203.0.113.10
```

Records in the plan:

| Record | Name | Content |
|--------|------|---------|
| MX | `example.com.` | `10 relay.example.com.` |
| A | `relay.example.com.` | the relay IP |
| SPF (TXT) | `example.com.` | `v=spf1 ip4:<relay-ip> -all` |
| DKIM (TXT) | `<selector>._domainkey.example.com.` | `v=DKIM1; k=rsa; p=<public key>` |
| DMARC (TXT) | `_dmarc.example.com.` | `v=DMARC1; p=<policy>; sp=quarantine` plus `rua`/`ruf` when set |
| SRV | `_imaps._tcp.example.com.`, `_imap._tcp.example.com.`, `_submission._tcp.example.com.` | point at the relay hostname per RFC 6186; the `_imap` record has weight 0 (service unavailable — DarkPipe is TLS-only) |
| CNAME | `autoconfig.example.com.`, `autodiscover.example.com.` | target the relay hostname (Thunderbird/Outlook autodiscovery) |

The DKIM private key is written to `--dkim-key-dir` as `<selector>.private.pem` (mode 0600). Copy it to the home device and install it for Rspamd signing — see [DKIM signing](../security.md#dkim).

### Mode: validate

Checks the live zone against the expected records (SPF, DKIM, DMARC, MX, SRV, autodiscover CNAMEs, plus the PTR/reverse-DNS record):

```
dns-setup --domain example.com --relay-hostname relay.example.com --relay-ip 203.0.113.10 --validate-only
```

Exit code is non-zero when any check fails. `--json` emits a machine-readable report.

### Mode: DKIM rotation

```
dns-setup --domain example.com --relay-hostname relay.example.com --relay-ip 203.0.113.10 --rotate-dkim
```

Generates a key pair for the next selector, prints the new TXT record, and lists the follow-up steps (publish the record, update the signing key, wait 7 days, remove the old record).

### Provider detection

`--apply` auto-detects the domain's DNS provider from its NS records:

- NS records containing `cloudflare.com` → Cloudflare API (requires `CLOUDFLARE_API_TOKEN`)
- NS records containing `awsdns` → Route53 API (standard AWS credential chain)
- Anything else → no provider; the tool prints the manual setup guide and exits

There is no `--provider` flag. If the domain is hosted by Cloudflare or Route53, set the credential environment variables and re-run with `--apply`; otherwise add the records from the printed guide.

### DKIM selectors

Selectors are time-based to support automated quarterly rotation:

```
{prefix}-{YYYY}q{Q}        e.g. darkpipe-2026q4
```

`--rotate-dkim` moves to the next quarter's selector. The mail server (Rspamd) must be updated to sign with the new selector after the new record propagates.

## darkpipe-setup

`darkpipe-setup` is the setup wizard and migration tool. Built from [`deploy/setup/cmd/darkpipe-setup`](../../deploy/setup/cmd/darkpipe-setup/main.go). It is a cobra CLI with three subcommands: `version`, `setup`, and `migrate`. Running it without a subcommand prints usage.

### darkpipe-setup setup

Interactive wizard that validates configuration and generates deployment files in the current directory.

Flow:

1. **Setup mode** — `Quick (recommended defaults)` or `Advanced (customize everything)`
2. **Required questions** (both modes):
   - Primary mail domain (validates DNS resolution live)
   - Cloud relay hostname (tests port 25 reachability live)
   - Admin email address
   - Mail server: `stalwart` (default), `maddy`, or `postfix-dovecot`
   - Webmail: `snappymail` (default), `roundcube`, or `none`
   - Calendar/Contacts: Stalwart built-in, Radicale, or none
   - Transport: `wireguard` (default) or `mtls`
3. **Advanced questions** (advanced mode only): message queuing, TLS strict mode
4. **Summary and confirmation**
5. **Generated output**:

| File | Contents |
|------|----------|
| `docker-compose.yml` | A single compose file with the selected services (relay + mail server + webmail + caddy + rspamd + redis) |
| `.darkpipe.yml` | Saved configuration (re-run the wizard to upgrade it) |
| `secrets/admin_password.txt` | A freshly generated 24-character admin password (printed once) |
| `secrets/dkim_private_key.pem` | Placeholder; the real key comes from `dns-setup` |
| `.darkpipe-configured` | Marker file indicating setup completed |

Notes:

- The wizard generates the admin password; it does not take one as input. Note it when the wizard finishes.
- Re-running `darkpipe-setup setup` in a directory containing `.darkpipe.yml` offers to upgrade the existing configuration.

### darkpipe-setup migrate

Migrates mail, contacts, and calendars from an existing provider into your DarkPipe installation.

```
darkpipe-setup migrate --from <provider> [options]
```

Running `migrate` without `--from` lists supported providers.

**Provider slugs (`--from`):**

| Slug | Authentication | Migrated data |
|------|----------------|---------------|
| `gmail` | OAuth2 device flow | IMAP, CalDAV, CardDAV |
| `outlook` | OAuth2 device flow (personal and M365) | IMAP, CalDAV, CardDAV |
| `icloud` | App-specific password | IMAP, CalDAV, CardDAV |
| `mailcow` | IMAP credentials | IMAP + user export |
| `mailu` | IMAP credentials | IMAP + user export |
| `docker-mailserver` | IMAP credentials | IMAP only |
| `generic` | IMAP credentials | IMAP, CalDAV, CardDAV (flexible) |

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--from` | *(required)* | Provider slug from the table above |
| `--apply` | `false` | Execute the migration. Without it, the run is a dry-run |
| `--folder-map` | *(empty)* | Custom folder mappings as one comma-separated value: `Source:Dest,Source2:Dest2` |
| `--labels-as-folders` | `false` | Map Gmail labels to folders instead of keywords (fallback for non-Stalwart destinations) |
| `--state-file` | `/data/migration-state.json` | Progress state file; enables interrupt/resume |
| `--contacts-mode` | `append` | Contact merge mode: `append`, `overwrite`, or `skip` |
| `--imap-dest` | from DarkPipe config | Destination IMAP server override |
| `--caldav-dest` | from DarkPipe config | Destination CalDAV server override |
| `--carddav-dest` | from DarkPipe config | Destination CardDAV server override |
| `--dest-user` | from DarkPipe config | Destination mailbox username |
| `--dest-pass` | from DarkPipe config | Destination mailbox password |
| `--vcf-file` | *(empty)* | Import contacts from a `.vcf` file |
| `--ics-file` | *(empty)* | Import calendar events from an `.ics` file |
| `--batch-size` | `50` | IMAP fetch batch size |

Source credentials (host, port, username, password) and OAuth2 client details are requested interactively by the wizard; there are no `--source-*` or `--oauth-*` flags.

The tool does not delete mail from the source provider. Migrations can be interrupted and resumed: state is tracked in `--state-file`, and already-migrated messages are skipped on re-run.

## Output conventions

Both tools:

- Default to safe behavior: `dns-setup` dry-runs, `migrate` dry-runs, `--apply` opts in.
- Print human-readable, colorized output by default; `--json` (dns-setup) emits a machine-readable report.
- Exit non-zero on validation or apply failures.
