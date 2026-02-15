# DarkPipe Mail Migration Guide

This guide covers migrating existing email from 7 popular providers to your DarkPipe installation.

## Overview

DarkPipe includes a mail migration tool that uses IMAP to copy messages from your current email provider to your new DarkPipe mail server. The tool:

- Copies all messages and folders from source to destination
- Preserves folder structure (with mapping for provider-specific folders)
- Maintains message flags (read, starred, etc.)
- Supports dry-run mode (preview before applying)
- Shows progress tracking for large mailboxes
- Handles OAuth2 authentication for Gmail and Outlook (no app passwords needed)

**Important:** Migration does NOT delete mail from the source provider. Your original mail remains untouched.

## Supported Providers

| Provider | Authentication | Notes |
|----------|----------------|-------|
| **Gmail** | OAuth2 device flow | Requires Google Cloud project with Gmail API enabled |
| **Outlook / Microsoft 365** | OAuth2 device flow | Personal and organizational accounts supported |
| **iCloud** | App-specific password | Generate via appleid.apple.com |
| **MailCow** | IMAP credentials | Direct IMAP access |
| **Mailu** | IMAP credentials | Direct IMAP access |
| **docker-mailserver** | IMAP credentials | Direct IMAP access |
| **Generic IMAP** | IMAP credentials | Any standards-compliant IMAP server |

## Before You Migrate

**Complete DarkPipe Setup:**
1. Cloud relay and home device fully deployed
2. DNS configured (MX, SPF, DKIM, DMARC)
3. Test email sending and receiving working
4. Admin account created on your mail server

**Prepare Source Provider:**
1. For Gmail/Outlook: Set up OAuth2 application (see provider-specific sections below)
2. For iCloud: Generate app-specific password
3. For self-hosted: Ensure IMAP is enabled and accessible

**Backup Recommendation:**
- Back up source mailbox before migration (export or local backup)
- Migration tool does not delete source mail, but backups are always wise

**Estimate Time:**
- Small mailbox (< 1GB, < 10,000 messages): 30-60 minutes
- Medium mailbox (1-5GB, 10,000-50,000 messages): 2-6 hours
- Large mailbox (> 5GB, > 50,000 messages): 6-24+ hours

Network speed and provider rate limits affect migration time.

## Running a Migration

### Step 1: Download Migration Tool

Download the darkpipe-setup tool, which includes the migrate subcommand:

```bash
# Linux (amd64)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-amd64
chmod +x darkpipe-setup-linux-amd64

# Linux (arm64)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-linux-arm64
chmod +x darkpipe-setup-linux-arm64

# macOS (Intel)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-amd64
chmod +x darkpipe-setup-darwin-amd64

# macOS (Apple Silicon)
curl -LO https://github.com/trek-e/darkpipe/releases/latest/download/darkpipe-setup-darwin-arm64
chmod +x darkpipe-setup-darwin-arm64
```

### Step 2: Run Migration Wizard

```bash
./darkpipe-setup-linux-amd64 migrate
```

The wizard will prompt for:

1. **Source provider**: Select from supported providers
2. **Source authentication**: OAuth2 or credentials depending on provider
3. **Destination mail server**: Your DarkPipe mail server details
   - Hostname: mail.example.com
   - Port: 993 (IMAPS)
   - Username: your-email@example.com
   - Password: your DarkPipe password
4. **Migration options**:
   - Dry-run (default): Preview what would be migrated
   - Folder mapping: Customize folder name mapping
   - Message limit: Migrate only recent messages (optional)

### Step 3: Review Dry-Run Output

The tool runs in dry-run mode by default, showing:

- Total messages to migrate
- Total mailbox size
- Folder mapping (source → destination)
- Estimated time

Example output:
```
Source: Gmail (user@gmail.com)
Destination: DarkPipe (user@example.com)

Folders to migrate:
  INBOX → INBOX (3,421 messages, 1.2 GB)
  [Gmail]/Sent Mail → Sent (1,205 messages, 450 MB)
  [Gmail]/Drafts → Drafts (12 messages, 45 KB)
  Work → Work (892 messages, 320 MB)
  Personal → Personal (1,544 messages, 680 MB)

Total: 7,074 messages, 2.67 GB
Estimated time: 2-3 hours

Dry-run complete. No changes made.
Run with --apply to perform migration.
```

### Step 4: Run Actual Migration

If dry-run looks correct, run again with `--apply`:

```bash
./darkpipe-setup-linux-amd64 migrate --apply
```

Progress display:
```
Migrating INBOX... [=====>    ] 1,234/3,421 (36%) - ETA 45 minutes
```

**Migration can be interrupted and resumed.** The tool tracks progress and skips already-migrated messages if restarted.

### Step 5: Verify Migration

After migration completes:

1. Log into DarkPipe webmail
2. Check all folders were created
3. Spot-check messages in various folders
4. Verify attachments are intact
5. Check message flags (read/unread, starred)

## Provider-Specific Instructions

### Gmail

Gmail uses OAuth2 device flow for authentication, which requires setting up a Google Cloud project.

#### Set Up Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (e.g., "DarkPipe Migration")
3. Enable Gmail API:
   - APIs & Services > Library
   - Search "Gmail API"
   - Click "Enable"
4. Create OAuth2 credentials:
   - APIs & Services > Credentials
   - Create Credentials > OAuth client ID
   - Application type: Desktop app
   - Name: DarkPipe Migration
5. Download client ID JSON file

#### Run Migration with OAuth2

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider gmail \
  --oauth-client-id YOUR_CLIENT_ID \
  --oauth-client-secret YOUR_CLIENT_SECRET
```

The tool will:
1. Display a URL and device code
2. Prompt you to visit the URL in a browser
3. Enter the device code
4. Authorize DarkPipe Migration app
5. Migration begins after authorization

#### Gmail Folder Mapping

Gmail uses labels, not traditional folders. Common mappings:

| Gmail Label | DarkPipe Folder |
|-------------|-----------------|
| INBOX | INBOX |
| [Gmail]/Sent Mail | Sent |
| [Gmail]/Drafts | Drafts |
| [Gmail]/Trash | Trash |
| [Gmail]/Spam | Junk |
| [Gmail]/Starred | Starred |
| [Gmail]/All Mail | (skipped - duplicate of other folders) |
| Custom labels | Same name |

#### Gmail Notes

- **All Mail is skipped by default** to avoid duplicates (all messages are already in INBOX or other labels)
- **Conversation view**: Gmail groups messages in threads; DarkPipe shows individual messages
- **Rate limits**: Google limits IMAP operations; large mailboxes may take longer
- **Alternative: Google Takeout**: For very large mailboxes, use Google Takeout to export mbox files, then import via IMAP

### Outlook / Microsoft 365

Outlook uses OAuth2 device flow for both personal (outlook.com, hotmail.com) and organizational (Microsoft 365) accounts.

#### Set Up Azure AD Application

1. Go to [Azure Portal](https://portal.azure.com/)
2. Azure Active Directory > App registrations
3. New registration:
   - Name: DarkPipe Migration
   - Supported account types: Personal Microsoft accounts and organizational accounts
   - Redirect URI: (leave blank)
4. After creation, note Application (client) ID
5. Certificates & secrets > New client secret
   - Description: Migration tool
   - Expiry: 6 months (or longer if needed)
   - Note the secret value
6. API permissions:
   - Add a permission > Microsoft Graph > Delegated permissions
   - Add: Mail.Read, IMAP.AccessAsUser.All
   - Grant admin consent (if organizational account)

#### Run Migration with OAuth2

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider outlook \
  --oauth-client-id YOUR_CLIENT_ID \
  --oauth-client-secret YOUR_CLIENT_SECRET
```

Follow the device code authorization flow (same as Gmail).

#### Outlook Folder Mapping

| Outlook Folder | DarkPipe Folder |
|----------------|-----------------|
| Inbox | INBOX |
| Sent Items | Sent |
| Drafts | Drafts |
| Deleted Items | Trash |
| Junk Email | Junk |
| Archive | Archive |
| Custom folders | Same name |

#### Outlook Notes

- **Shared mailboxes**: Supported for Microsoft 365 organizational accounts (requires additional permissions)
- **Rate limits**: Microsoft throttles IMAP connections; expect slower migration for large mailboxes
- **Focused Inbox**: Messages in "Focused" and "Other" tabs both migrate to INBOX

### iCloud

iCloud uses app-specific passwords for IMAP access.

#### Generate App-Specific Password

1. Go to [appleid.apple.com](https://appleid.apple.com/)
2. Sign in with your Apple ID
3. Security section > App-Specific Passwords
4. Generate a password:
   - Label: DarkPipe Migration
   - Copy the generated password (format: xxxx-xxxx-xxxx-xxxx)

#### Run Migration with App-Specific Password

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider icloud \
  --source-user your-email@icloud.com \
  --source-password xxxx-xxxx-xxxx-xxxx
```

#### iCloud Folder Mapping

| iCloud Folder | DarkPipe Folder |
|---------------|-----------------|
| INBOX | INBOX |
| Sent Messages | Sent |
| Drafts | Drafts |
| Deleted Messages | Trash |
| Junk | Junk |
| Archive | Archive |
| Custom folders | Same name |

#### iCloud Notes

- **IMAP server**: imap.mail.me.com (port 993)
- **No OAuth2**: iCloud does not support OAuth2 for IMAP; app-specific passwords required
- **Two-factor authentication**: If enabled, you MUST use app-specific password (regular password won't work)

### MailCow

MailCow uses standard IMAP authentication.

#### Run Migration

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider mailcow \
  --source-host mail.yourmailcow.com \
  --source-port 993 \
  --source-user your-email@domain.com \
  --source-password your-password
```

#### MailCow Notes

- **IMAP must be enabled** in MailCow mailbox settings
- **Firewall**: Ensure IMAP port 993 is accessible from your migration location
- **Self-signed certificates**: Use `--insecure-skip-verify` if MailCow uses self-signed TLS certificate (not recommended for production)

### Mailu

Mailu uses standard IMAP authentication.

#### Run Migration

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider mailu \
  --source-host mail.yourmailu.com \
  --source-port 993 \
  --source-user your-email@domain.com \
  --source-password your-password
```

#### Mailu Notes

- **IMAP access**: Enabled by default in Mailu
- **Admin vs user**: Use user credentials (admin credentials won't work for IMAP)

### docker-mailserver

docker-mailserver uses standard IMAP authentication.

#### Run Migration

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider docker-mailserver \
  --source-host mail.yourdomain.com \
  --source-port 993 \
  --source-user your-email@domain.com \
  --source-password your-password
```

#### docker-mailserver Notes

- **Dovecot**: docker-mailserver uses Dovecot for IMAP (same as DarkPipe Postfix+Dovecot option)
- **Folder structure**: Should migrate cleanly with minimal folder mapping needed

### Generic IMAP

For any other IMAP-compliant server not listed above.

#### Run Migration

```bash
./darkpipe-setup-linux-amd64 migrate \
  --provider generic \
  --source-host imap.yourprovider.com \
  --source-port 993 \
  --source-user your-email@domain.com \
  --source-password your-password
```

#### Generic IMAP Notes

- **Test connectivity first**: `telnet imap.yourprovider.com 993` should connect
- **TLS required**: Migration tool requires TLS (port 993, not plain port 143)
- **Folder names**: Generic provider folder names copied as-is

## Troubleshooting

### Authentication Failures

**Gmail/Outlook OAuth2 fails:**
- Verify client ID and secret are correct
- Check that Gmail API (Gmail) or Graph API (Outlook) is enabled
- Ensure redirect URI is configured (can be blank for device flow)
- Try deleting and recreating OAuth2 credentials

**iCloud authentication fails:**
- Ensure you're using app-specific password, not regular password
- Verify two-factor authentication is enabled on Apple ID
- Generate a new app-specific password if old one doesn't work

**IMAP credentials fail:**
- Double-check username and password
- Verify IMAP is enabled on source server
- Try logging into source webmail to confirm credentials

### Timeout on Large Mailboxes

**Symptoms:** Migration stalls or times out on folders with many messages.

**Solutions:**
- Run migration from a system with good network connection (not over VPN)
- Migrate in batches: `--folder INBOX` to migrate one folder at a time
- Use `--limit 10000` to migrate only recent messages first, then run again without limit

### Folder Mapping Conflicts

**Symptoms:** Folders created with unexpected names.

**Solutions:**
- Use `--folder-map` to specify custom mapping:
  ```bash
  --folder-map "[Gmail]/Sent Mail:Sent" \
  --folder-map "Work Projects:Work"
  ```
- Run dry-run first to preview folder mapping
- After migration, manually merge folders via IMAP client if needed

### Missing Messages

**Symptoms:** Message counts don't match between source and destination.

**Causes and solutions:**
- **Duplicate detection**: Tool skips messages with identical Message-ID (not a bug, prevents duplicates)
- **All Mail folder**: Gmail's All Mail is skipped by default (messages are in other folders)
- **Corrupted messages**: Some providers have corrupted messages that fail to migrate (logged in output)

**Verify:**
- Check migration log output for skipped messages
- Compare folder-by-folder message counts (not total across all folders, which may include duplicates)

### Slow Migration Speed

**Causes:**
- Provider rate limiting (Gmail and Outlook throttle IMAP)
- Large attachments (takes time to transfer)
- Network latency
- Source server performance

**Improvements:**
- Run migration from a high-bandwidth location (not mobile hotspot)
- Migrate during off-peak hours (less provider throttling)
- Be patient - large mailboxes take time

### Migration Interrupted

**Resume migration:**
- Simply run the same migrate command again
- Tool tracks progress and skips already-migrated messages
- Safe to interrupt and resume multiple times

## Post-Migration

### Update Mail Clients

After migration, update your mail clients (phone, desktop) to connect to DarkPipe:

- See [docs/quickstart.md Step 7](quickstart.md#step-7-onboard-devices) for device onboarding
- Use profile server for easy configuration

### Keep Source Provider (Optional)

**Recommended approach:** Keep source provider account active for 30-60 days after migration.

- Set up forwarding from old account to DarkPipe
- Monitor both inboxes during transition
- Gives time to catch any services still sending to old address
- After transition period, close or downgrade old account

### Update Email Address

Don't forget to update your email address with:

- Important services (banks, utilities, subscriptions)
- Social media accounts
- Online shopping accounts
- Professional contacts
- Email signatures

## Alternative: Manual Migration

If the migration tool doesn't work for your provider, manual IMAP-to-IMAP migration is possible with Thunderbird or other IMAP clients:

1. Configure source account in Thunderbird
2. Configure destination (DarkPipe) account in Thunderbird
3. Select all messages in source folder
4. Drag and drop to destination folder
5. Repeat for each folder

This is slower and more manual but works for any IMAP provider.

---

Last Updated: 2026-02-15

License: AGPLv3 - See [LICENSE](../LICENSE)
