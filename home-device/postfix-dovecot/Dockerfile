# Postfix + Dovecot Mail Server
# Traditional split MTA (Postfix) + IMAP (Dovecot) configuration
# Documentation:
#   Postfix: http://www.postfix.org/
#   Dovecot: https://doc.dovecot.org/

FROM alpine:3.21

# Install mail server components
# - postfix: MTA for SMTP inbound (25) and submission (587)
# - dovecot: IMAP server (993) and LMTP local delivery
# - dovecot-lmtpd: LMTP daemon for Postfix -> Dovecot delivery
# - dovecot-pigeonhole: Sieve for future mail filtering
# - ca-certificates: TLS certificate support
RUN apk add --no-cache \
    postfix \
    postfix-lmdb \
    dovecot \
    dovecot-lmtpd \
    dovecot-pigeonhole \
    ca-certificates \
    openssl \
    bash

# Create vmail user for virtual mailbox ownership (UID/GID 5000)
# Virtual mailboxes don't use system users - all mail owned by vmail
RUN addgroup -g 5000 vmail && \
    adduser -D -u 5000 -G vmail -h /var/mail/vhosts -s /sbin/nologin vmail

# Create virtual mailbox base directory
RUN mkdir -p /var/mail/vhosts && \
    chown -R vmail:vmail /var/mail/vhosts && \
    chmod 0755 /var/mail/vhosts

# Create directories for configuration and TLS certificates
RUN mkdir -p /etc/ssl/mail /etc/dovecot /etc/postfix-dovecot

# Copy Postfix configuration
COPY postfix/main.cf /etc/postfix/main.cf
COPY postfix/master.cf /etc/postfix/master.cf

# Copy Dovecot configuration
COPY dovecot/dovecot.conf /etc/dovecot/dovecot.conf

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod 0755 /entrypoint.sh

# Expose mail server ports
EXPOSE 25 587 993

# Run entrypoint script
CMD ["/entrypoint.sh"]
