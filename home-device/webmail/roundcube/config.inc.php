<?php
// DarkPipe Roundcube configuration
// IMAP passthrough authentication (mail server is source of truth)

// IMAP server (connects to whichever mail server profile is active)
// All three mail servers (stalwart, maddy, postfix-dovecot) expose IMAP on port 993
$config['imap_host'] = 'ssl://mail-server:993';
$config['imap_conn_options'] = array(
    'ssl' => array(
        'verify_peer' => false,       // Self-signed cert within WireGuard tunnel
        'verify_peer_name' => false,
    ),
);

// SMTP submission for sending
$config['smtp_host'] = 'tls://mail-server:587';
$config['smtp_port'] = 587;
$config['smtp_conn_options'] = array(
    'ssl' => array(
        'verify_peer' => false,
        'verify_peer_name' => false,
    ),
);
$config['smtp_user'] = '%u';         // Use IMAP username
$config['smtp_pass'] = '%p';         // Use IMAP password

// Auto-create user on first IMAP login
$config['auto_create_user'] = true;

// Database (SQLite for settings/contacts - not auth)
$config['db_dsnw'] = 'sqlite:////var/roundcube/db/sqlite.db?mode=0640';

// UI settings
$config['skin'] = 'elastic';         // Mobile-responsive skin (WEB-02)
$config['language'] = null;           // Auto-detect browser language
$config['product_name'] = 'DarkPipe Mail';

// Session settings (avoid timeout frustration - Pitfall 9)
$config['session_lifetime'] = 60;     // 60 minutes inactivity timeout
$config['refresh_interval'] = 60;     // Auto-refresh every 60 seconds (keeps session alive)

// Identity
$config['identities_level'] = 0;     // User can edit all identity fields

// Upload/message limits
$config['upload_max_filesize'] = '25M';
$config['max_message_size'] = '25M';

// Plugins (minimal set - avoid Pitfall 7)
$config['plugins'] = array(
    'archive',                         // Archive button
    'zipdownload',                     // Download attachments as zip
    'managesieve',                     // Sieve filter management (if server supports)
);

// Security
$config['ip_check'] = false;          // Behind reverse proxy, IP changes
$config['use_https'] = true;          // Force HTTPS (Caddy terminates TLS)
$config['des_key'] = 'CHANGE_THIS_24CHAR_KEY!';  // 24-char encryption key for session
