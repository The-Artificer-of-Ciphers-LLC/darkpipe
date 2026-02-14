#!/usr/bin/env bash
# DarkPipe Radicale Collection Setup Script
# Creates default calendar and address book for users
# Usage: ./setup-collections.sh <username> [--shared]
#   <username>: Create default personal calendar and address book for this user (e.g., alice@example.com)
#   --shared: Also create shared family calendar and address book (run once on initial setup)

set -euo pipefail

COLLECTIONS_DIR="${COLLECTIONS_DIR:-/data/collections/collection-root}"

usage() {
    echo "Usage: $0 <username> [--shared]"
    echo ""
    echo "  <username>    Create default calendar and address book for this user"
    echo "  --shared      Also create shared family calendar and address book"
    echo ""
    echo "Examples:"
    echo "  $0 alice@example.com          # Create personal collections for alice"
    echo "  $0 alice@example.com --shared # Create personal + shared collections"
    echo ""
    echo "Environment:"
    echo "  COLLECTIONS_DIR  Base directory for collections (default: /data/collections/collection-root)"
    exit 1
}

create_calendar() {
    local user="$1"
    local calendar_dir="${COLLECTIONS_DIR}/${user}/calendar.ics"

    mkdir -p "$calendar_dir"

    cat > "${calendar_dir}/.Radicale.props" <<EOF
{"tag": "VCALENDAR", "{DAV:}displayname": "My Calendar", "{urn:ietf:params:xml:ns:caldav}supported-calendar-component-set": "VEVENT,VTODO,VJOURNAL"}
EOF

    echo "Created calendar for ${user}"
}

create_addressbook() {
    local user="$1"
    local contacts_dir="${COLLECTIONS_DIR}/${user}/contacts.vcf"

    mkdir -p "$contacts_dir"

    cat > "${contacts_dir}/.Radicale.props" <<EOF
{"tag": "VADDRESSBOOK", "{DAV:}displayname": "My Contacts"}
EOF

    echo "Created address book for ${user}"
}

create_shared_calendar() {
    local shared_dir="${COLLECTIONS_DIR}/shared/family-calendar.ics"

    mkdir -p "$shared_dir"

    cat > "${shared_dir}/.Radicale.props" <<EOF
{"tag": "VCALENDAR", "{DAV:}displayname": "Family Calendar", "{urn:ietf:params:xml:ns:caldav}supported-calendar-component-set": "VEVENT,VTODO,VJOURNAL"}
EOF

    echo "Created shared family calendar"
}

create_shared_addressbook() {
    local shared_dir="${COLLECTIONS_DIR}/shared/family-contacts.vcf"

    mkdir -p "$shared_dir"

    cat > "${shared_dir}/.Radicale.props" <<EOF
{"tag": "VADDRESSBOOK", "{DAV:}displayname": "Family Contacts"}
EOF

    echo "Created shared family address book"
}

# Parse arguments
if [ $# -eq 0 ]; then
    usage
fi

USERNAME="$1"
CREATE_SHARED=false

shift
while [ $# -gt 0 ]; do
    case "$1" in
        --shared)
            CREATE_SHARED=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate username format (basic check for email-like string)
if [[ ! "$USERNAME" =~ ^[^@]+@[^@]+$ ]]; then
    echo "Error: Username must be in email format (user@domain)"
    exit 1
fi

# Create personal collections
echo "Creating collections for ${USERNAME}..."
create_calendar "$USERNAME"
create_addressbook "$USERNAME"

# Create shared collections if requested
if [ "$CREATE_SHARED" = true ]; then
    echo "Creating shared collections..."
    create_shared_calendar
    create_shared_addressbook
fi

echo "Done!"
