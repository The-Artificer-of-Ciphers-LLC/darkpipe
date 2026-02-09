#!/bin/sh
set -e

echo "DarkPipe cloud relay entrypoint starting..."

# Substitute environment variables in Postfix main.cf
if [ -z "$RELAY_HOSTNAME" ] || [ -z "$RELAY_DOMAIN" ]; then
  echo "ERROR: RELAY_HOSTNAME and RELAY_DOMAIN must be set"
  exit 1
fi

echo "Configuring Postfix for hostname=$RELAY_HOSTNAME domain=$RELAY_DOMAIN"

# Use envsubst to replace placeholders in main.cf
envsubst < /etc/postfix/main.cf.template > /etc/postfix/main.cf

# Hash the transport map (LMDB format)
echo "Creating transport map..."
postmap lmdb:/etc/postfix/transport

# Validate Postfix configuration
echo "Validating Postfix configuration..."
postfix check

# Ensure queue directory permissions
mkdir -p /var/spool/postfix
chown -R postfix:postfix /var/spool/postfix

# Start relay daemon in background
echo "Starting relay daemon..."
/usr/local/bin/relay-daemon &
RELAY_PID=$!

# Trap SIGTERM to gracefully stop both processes
trap 'echo "Stopping services..."; postfix stop; kill $RELAY_PID; wait $RELAY_PID' SIGTERM SIGINT

# Start Postfix in foreground
echo "Starting Postfix..."
postfix start-fg &
POSTFIX_PID=$!

# Wait for Postfix to exit
wait $POSTFIX_PID
