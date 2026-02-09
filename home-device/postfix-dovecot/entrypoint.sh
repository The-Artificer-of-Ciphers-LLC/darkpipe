#!/bin/bash
# Postfix+Dovecot Entrypoint Script
# This script:
# 1. Generates self-signed TLS certificate if none exists
# 2. Creates initial empty map files for Postfix
# 3. Creates default admin user in Dovecot users file
# 4. Starts Dovecot in background
# 5. Starts Postfix in foreground

set -e

# Environment variables with defaults
MAIL_DOMAIN="${MAIL_DOMAIN:-example.com}"
MAIL_HOSTNAME="${MAIL_HOSTNAME:-mail.example.com}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-changeme}"

echo "==> Starting Postfix+Dovecot mail server"
echo "    Domain: ${MAIL_DOMAIN}"
echo "    Hostname: ${MAIL_HOSTNAME}"
echo "    Admin email: ${ADMIN_EMAIL}"

# ============================================================================
# Generate Self-Signed TLS Certificate
# ============================================================================

if [ ! -f /etc/ssl/mail/cert.pem ] || [ ! -f /etc/ssl/mail/key.pem ]; then
  echo "==> Generating self-signed TLS certificate"
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout /etc/ssl/mail/key.pem \
    -out /etc/ssl/mail/cert.pem \
    -subj "/C=US/ST=State/L=City/O=DarkPipe/CN=${MAIL_HOSTNAME}" \
    2>/dev/null

  chmod 0600 /etc/ssl/mail/key.pem
  chmod 0644 /etc/ssl/mail/cert.pem
  echo "    Generated certificate for ${MAIL_HOSTNAME}"
else
  echo "==> Using existing TLS certificate"
fi

# ============================================================================
# Substitute Environment Variables in Postfix Configuration
# ============================================================================

echo "==> Configuring Postfix"

# Substitute myhostname
sed -i "s/^myhostname = .*/myhostname = ${MAIL_HOSTNAME}/" /etc/postfix/main.cf

# Substitute mydomain
sed -i "s/^mydomain = .*/mydomain = ${MAIL_DOMAIN}/" /etc/postfix/main.cf

# Substitute virtual_mailbox_domains
sed -i "s/^virtual_mailbox_domains = .*/virtual_mailbox_domains = ${MAIL_DOMAIN}/" /etc/postfix/main.cf

# Substitute postmaster_address in Dovecot config
sed -i "s/postmaster_address = .*/postmaster_address = postmaster@${MAIL_DOMAIN}/" /etc/dovecot/dovecot.conf

# ============================================================================
# Create Initial Postfix Map Files
# ============================================================================

echo "==> Initializing Postfix map files"

# Create virtual mailbox map if it doesn't exist
if [ ! -f /etc/postfix/vmailbox ]; then
  # Extract username from admin email (before @)
  ADMIN_USER="${ADMIN_EMAIL%%@*}"

  # Add admin user to virtual mailbox map
  # Format: user@domain  domain/user/
  echo "${ADMIN_EMAIL}  ${MAIL_DOMAIN}/${ADMIN_USER}/" > /etc/postfix/vmailbox
  echo "    Created virtual mailbox for ${ADMIN_EMAIL}"
fi

# Create virtual alias map if it doesn't exist
if [ ! -f /etc/postfix/virtual ]; then
  # Create empty virtual alias map (aliases added in Phase 03 Plan 02)
  touch /etc/postfix/virtual
  echo "    Created empty virtual alias map"
fi

# Convert text map files to LMDB database format
postmap lmdb:/etc/postfix/vmailbox
postmap lmdb:/etc/postfix/virtual

# ============================================================================
# Create Initial Dovecot Users File
# ============================================================================

echo "==> Initializing Dovecot users"

if [ ! -f /etc/dovecot/users ]; then
  # Create admin user with plaintext password
  # Format: user@domain:{PLAIN}password
  # WARNING: Change default password after first deployment!
  echo "${ADMIN_EMAIL}:{PLAIN}${ADMIN_PASSWORD}" > /etc/dovecot/users
  chmod 0600 /etc/dovecot/users
  echo "    Created default admin user: ${ADMIN_EMAIL}"
  echo "    WARNING: Change default password after deployment!"
else
  echo "    Using existing users file"
fi

# ============================================================================
# Create Maildir for Admin User
# ============================================================================

ADMIN_USER="${ADMIN_EMAIL%%@*}"
ADMIN_MAILDIR="/var/mail/vhosts/${MAIL_DOMAIN}/${ADMIN_USER}/Maildir"

if [ ! -d "${ADMIN_MAILDIR}" ]; then
  echo "==> Creating maildir for ${ADMIN_EMAIL}"
  mkdir -p "${ADMIN_MAILDIR}"/{cur,new,tmp}
  chown -R vmail:vmail "/var/mail/vhosts/${MAIL_DOMAIN}"
  chmod -R 0700 "/var/mail/vhosts/${MAIL_DOMAIN}"
fi

# ============================================================================
# Fix Postfix Queue Directory Permissions
# ============================================================================

# Ensure Postfix queue directory exists and has correct permissions
if [ ! -d /var/spool/postfix/private ]; then
  mkdir -p /var/spool/postfix/private
fi

# Create pid directory for Postfix
if [ ! -d /var/spool/postfix/pid ]; then
  mkdir -p /var/spool/postfix/pid
fi

# Set correct permissions for Postfix directories
chown -R postfix:postfix /var/spool/postfix
chmod 0755 /var/spool/postfix
chmod 0700 /var/spool/postfix/private

# ============================================================================
# Start Dovecot in Background
# ============================================================================

echo "==> Starting Dovecot"
dovecot -F &
DOVECOT_PID=$!

# Wait for Dovecot to create LMTP socket
sleep 2

if [ ! -S /var/spool/postfix/private/dovecot-lmtp ]; then
  echo "ERROR: Dovecot LMTP socket not created"
  kill $DOVECOT_PID 2>/dev/null || true
  exit 1
fi

echo "    Dovecot started (PID: ${DOVECOT_PID})"

# ============================================================================
# Start Postfix in Foreground
# ============================================================================

echo "==> Starting Postfix"

# Initialize Postfix if needed
if [ ! -f /var/spool/postfix/pid/master.pid ]; then
  postfix check || true
fi

# Trap SIGTERM for graceful shutdown
trap "echo 'Shutting down...'; postfix stop; kill ${DOVECOT_PID} 2>/dev/null || true; exit 0" SIGTERM SIGINT

# Start Postfix in foreground mode
exec postfix start-fg
