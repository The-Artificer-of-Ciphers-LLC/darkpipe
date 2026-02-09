#!/bin/bash
# Maddy User Setup Script
#
# This script demonstrates how to create users across multiple domains
# using Maddy's command-line tools (maddy creds, maddy imap-acct).
#
# Prerequisites:
# - Maddy container must be running
# - Execute this script inside the Maddy container:
#   docker compose exec maddy /data/setup-users.sh
#
# Usage:
#   docker compose --profile maddy exec maddy /data/setup-users.sh
#
# Note: This script creates example users with default passwords.
# Modify email addresses and passwords for production use.

set -e

echo "==> Maddy User Setup"
echo ""

# ============================================================================
# Create Users
# ============================================================================

echo "==> Creating users"

# Create alice@example.com (user on first domain)
if ! maddy creds list | grep -q "alice@example.com"; then
  echo "changeme" | maddy creds create --password alice@example.com
  maddy imap-acct create alice@example.com
  echo "    Created: alice@example.com (password: changeme)"
else
  echo "    Exists: alice@example.com"
fi

# Create bob@example.org (user on second domain)
if ! maddy creds list | grep -q "bob@example.org"; then
  echo "changeme" | maddy creds create --password bob@example.org
  maddy imap-acct create bob@example.org
  echo "    Created: bob@example.org (password: changeme)"
else
  echo "    Exists: bob@example.org"
fi

echo ""

# ============================================================================
# Verify Setup
# ============================================================================

echo "==> Verifying user accounts"
maddy creds list
echo ""

# ============================================================================
# Summary
# ============================================================================

echo "==> Setup complete"
echo ""
echo "Users created:"
echo "  - alice@example.com (password: changeme)"
echo "  - bob@example.org (password: changeme)"
echo ""
echo "Domains configured (in maddy.conf):"
echo "  - example.com"
echo "  - example.org"
echo ""
echo "IMPORTANT: Change default passwords in production!"
echo "  maddy creds passwd alice@example.com"
echo "  maddy creds passwd bob@example.org"
echo ""
echo "Test IMAP login:"
echo "  openssl s_client -connect localhost:993 -quiet"
echo "  a001 login alice@example.com changeme"
echo "  a002 list \"\" \"*\""
echo "  a003 logout"
