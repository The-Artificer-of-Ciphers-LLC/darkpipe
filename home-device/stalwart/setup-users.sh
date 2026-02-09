#!/bin/bash
# Stalwart User and Domain Setup Script
#
# This script demonstrates how to create domains, users, and aliases
# using Stalwart's REST API.
#
# Prerequisites:
# - Stalwart container must be running
# - Management API accessible at http://localhost:8080
# - Default admin credentials: admin:changeme (CHANGE THIS!)
#
# Usage:
#   ./setup-users.sh
#
# Note: This script creates example users across multiple domains.
# Modify email addresses and passwords for production use.

set -e

# Configuration
STALWART_API="${STALWART_API:-http://localhost:8080}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-changeme}"

# Basic auth header (base64 encoded admin:password)
AUTH="$ADMIN_USER:$ADMIN_PASSWORD"

echo "==> Stalwart User Setup"
echo "    API: ${STALWART_API}"
echo "    Auth: ${ADMIN_USER}:***"
echo ""

# ============================================================================
# Add Domains
# ============================================================================

echo "==> Adding domains"

# Add example.com domain
curl -X POST "${STALWART_API}/api/v1/domain" \
  -u "${AUTH}" \
  -H "Content-Type: application/json" \
  -d '{"name": "example.com"}' \
  -w "\n" || echo "    (domain may already exist)"

# Add example.org domain
curl -X POST "${STALWART_API}/api/v1/domain" \
  -u "${AUTH}" \
  -H "Content-Type: application/json" \
  -d '{"name": "example.org"}' \
  -w "\n" || echo "    (domain may already exist)"

echo "    Domains added: example.com, example.org"
echo ""

# ============================================================================
# Add Users
# ============================================================================

echo "==> Adding users"

# Add alice@example.com (user on first domain)
curl -X POST "${STALWART_API}/api/v1/account" \
  -u "${AUTH}" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "changeme",
    "quota": 10737418240,
    "description": "Alice Example (example.com)"
  }' \
  -w "\n" || echo "    (user may already exist)"

# Add bob@example.org (user on second domain)
curl -X POST "${STALWART_API}/api/v1/account" \
  -u "${AUTH}" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bob@example.org",
    "password": "changeme",
    "quota": 10737418240,
    "description": "Bob Example (example.org)"
  }' \
  -w "\n" || echo "    (user may already exist)"

echo "    Users added: alice@example.com, bob@example.org"
echo "    Quota: 10GB per user"
echo ""

# ============================================================================
# Add Aliases (will be configured in Task 2)
# ============================================================================

echo "==> Alias setup will be added in Task 2"
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
echo "Domains configured:"
echo "  - example.com"
echo "  - example.org"
echo ""
echo "IMPORTANT: Change default passwords after deployment!"
echo ""
echo "Test IMAP login:"
echo "  openssl s_client -connect localhost:993 -quiet"
echo "  a001 login alice@example.com changeme"
echo "  a002 list \"\" \"*\""
echo "  a003 logout"
echo ""
echo "Web UI: ${STALWART_API}"
echo "  Login: ${ADMIN_USER} / ${ADMIN_PASSWORD}"
