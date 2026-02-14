#!/usr/bin/env bash
# DarkPipe Radicale User Sync Script
# Synchronizes mail server users to Radicale htpasswd file
# Usage: ./sync-users.sh <mail-server-type> [--user <user@domain>] [--password <password>]
#   <mail-server-type>: maddy | postfix-dovecot
#   --user <user@domain>: Sync only this user (default: sync all users)
#   --password <password>: Password for user (non-interactive mode)
#
# NOTE: Not needed for Stalwart — Stalwart uses the same user database for mail, CalDAV, and CardDAV.

set -euo pipefail

USERS_FILE="${USERS_FILE:-./caldav-carddav/radicale/users}"
SETUP_SCRIPT="${SETUP_SCRIPT:-./caldav-carddav/setup-collections.sh}"

usage() {
    echo "Usage: $0 <mail-server-type> [--user <user@domain>] [--password <password>]"
    echo ""
    echo "  <mail-server-type>  Mail server type: maddy | postfix-dovecot"
    echo "  --user <user>       Sync only this user (default: sync all users)"
    echo "  --password <pass>   Password for user (non-interactive mode)"
    echo ""
    echo "Examples:"
    echo "  $0 maddy                                        # Sync all Maddy users (interactive)"
    echo "  $0 postfix-dovecot --user alice@example.com    # Sync one user (interactive password)"
    echo "  $0 maddy --user bob@example.com --password secret  # Non-interactive sync"
    echo ""
    echo "Environment:"
    echo "  USERS_FILE     Path to Radicale users file (default: ./caldav-carddav/radicale/users)"
    echo "  SETUP_SCRIPT   Path to setup-collections.sh (default: ./caldav-carddav/setup-collections.sh)"
    echo ""
    echo "NOTE: Not needed for Stalwart — Stalwart uses the same user database for mail, CalDAV, and CardDAV."
    exit 1
}

sync_user_maddy() {
    local user="$1"
    local password="$2"

    # Use htpasswd to create bcrypt hash and append to users file
    # -B flag uses bcrypt encryption
    if command -v htpasswd >/dev/null 2>&1; then
        echo "$password" | htpasswd -iB -c "${USERS_FILE}.tmp" "$user"
        # Extract the user line and append to main file (remove duplicates first)
        grep -v "^${user}:" "$USERS_FILE" > "${USERS_FILE}.new" 2>/dev/null || touch "${USERS_FILE}.new"
        grep "^${user}:" "${USERS_FILE}.tmp" >> "${USERS_FILE}.new"
        mv "${USERS_FILE}.new" "$USERS_FILE"
        rm "${USERS_FILE}.tmp"
    else
        echo "Error: htpasswd command not found. Install apache2-utils (Debian/Ubuntu) or httpd-tools (RHEL/Fedora)"
        exit 1
    fi

    echo "Synced user: ${user}"
}

get_maddy_users() {
    # Get list of users from Maddy container
    if command -v docker >/dev/null 2>&1; then
        docker exec maddy maddy creds list 2>/dev/null || {
            echo "Error: Could not list Maddy users. Is the maddy container running?"
            exit 1
        }
    else
        echo "Error: docker command not found"
        exit 1
    fi
}

get_postfix_dovecot_users() {
    # Get list of users from Postfix+Dovecot users file
    local dovecot_users="./postfix-dovecot/dovecot/users"

    if [ ! -f "$dovecot_users" ]; then
        echo "Error: Dovecot users file not found at ${dovecot_users}"
        exit 1
    fi

    # Extract usernames (before the colon)
    cut -d: -f1 "$dovecot_users"
}

create_collections_for_user() {
    local user="$1"

    # Check if collections already exist
    local collections_dir="${COLLECTIONS_DIR:-/data/collections/collection-root}"
    if [ -d "${collections_dir}/${user}" ]; then
        echo "Collections already exist for ${user}, skipping creation"
        return
    fi

    # Run setup script to create default collections
    if [ -x "$SETUP_SCRIPT" ]; then
        "$SETUP_SCRIPT" "$user"
    else
        echo "Warning: Setup script not found or not executable at ${SETUP_SCRIPT}"
        echo "Run manually: ${SETUP_SCRIPT} ${user}"
    fi
}

# Parse arguments
if [ $# -eq 0 ]; then
    usage
fi

MAIL_SERVER_TYPE="$1"
SINGLE_USER=""
PASSWORD=""

shift
while [ $# -gt 0 ]; do
    case "$1" in
        --user)
            SINGLE_USER="$2"
            shift 2
            ;;
        --password)
            PASSWORD="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate mail server type
case "$MAIL_SERVER_TYPE" in
    maddy|postfix-dovecot)
        ;;
    stalwart)
        echo "NOTE: Stalwart uses built-in CalDAV/CardDAV with the same user database."
        echo "This script is not needed for Stalwart."
        exit 0
        ;;
    *)
        echo "Error: Unknown mail server type: ${MAIL_SERVER_TYPE}"
        usage
        ;;
esac

# Get list of users to sync
if [ -n "$SINGLE_USER" ]; then
    USERS_TO_SYNC=("$SINGLE_USER")
else
    # Get all users from mail server
    case "$MAIL_SERVER_TYPE" in
        maddy)
            mapfile -t USERS_TO_SYNC < <(get_maddy_users)
            ;;
        postfix-dovecot)
            mapfile -t USERS_TO_SYNC < <(get_postfix_dovecot_users)
            ;;
    esac
fi

# Sync each user
for user in "${USERS_TO_SYNC[@]}"; do
    # Skip empty lines
    [ -z "$user" ] && continue

    # Get password (interactive if not provided)
    if [ -z "$PASSWORD" ]; then
        echo -n "Password for ${user}: "
        read -rs user_password
        echo
    else
        user_password="$PASSWORD"
    fi

    # Sync user to Radicale
    case "$MAIL_SERVER_TYPE" in
        maddy|postfix-dovecot)
            sync_user_maddy "$user" "$user_password"
            ;;
    esac

    # Create default collections
    create_collections_for_user "$user"
done

echo ""
echo "User sync complete!"
echo ""
echo "To create shared family collections, run:"
echo "  ${SETUP_SCRIPT} <any-user@domain> --shared"
