# DKIM Private Keys

This directory holds DKIM private keys for Rspamd outbound signing.

## Key Naming Convention

```
{domain}.{selector}.key
```

Example: `example.com.darkpipe.key`

## Key Provisioning

1. Generate a key pair:
   ```bash
   go run dns/dkim/keygen.go
   ```
   This produces `{selector}.private.pem` and `{selector}.public.pem`.

2. Copy/rename the private key to this directory:
   ```bash
   cp darkpipe.private.pem home-device/spam-filter/rspamd/dkim-keys/example.com.darkpipe.key
   ```

3. Publish the public key as a DNS TXT record:
   ```
   darkpipe._domainkey.example.com. IN TXT "v=DKIM1; k=rsa; p=<base64-public-key>"
   ```

## Security

- Private keys should have restrictive permissions (0600)
- Do NOT commit real private keys to version control
- The directory is mounted read-only into the Rspamd container
