---
created: 2026-02-14T16:47:09.201Z
title: Mail migration tool for existing providers
area: general
files: []
---

## Problem

Users migrating to DarkPipe from existing email setups need a way to import their existing mailboxes, contacts, and calendars. Currently there's no migration path from:

- **MailCow Docker** — popular self-hosted mail solution with SOGo groupware
- **iCloud Mail** — Apple's cloud email service
- **Gmail** — Google's email with contacts and calendar
- **Outlook/Microsoft 365** — Microsoft's cloud email suite
- **Other self-hosted** — Mailu, Mail-in-a-Box, docker-mailserver, etc.

Without migration tooling, users face manual IMAP copy (loses folder structure metadata), no calendar/contact migration, and potential data loss.

## Solution

Potential approaches (TBD — needs research phase):

- **IMAP sync tool** — Use `imapsync` or Go-based IMAP-to-IMAP copy preserving folder hierarchy, flags, and dates
- **MailCow-specific** — MailCow API export + DarkPipe import (users, aliases, mailboxes)
- **iCloud/Gmail/Outlook** — OAuth-based IMAP access or Google Takeout / Apple Data & Privacy export parsing
- **CalDAV/CardDAV migration** — Export/import .ics and .vcf files between providers
- **CLI wizard** — `darkpipe migrate --from mailcow` interactive flow

This is v2 scope — capture for future milestone planning.
