# DarkPipe Cloud Relay - Multi-stage Docker build
#
# Stage 1: Build the Go relay daemon
# Stage 2: Runtime container with Postfix + relay daemon + WireGuard tools

FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build relay daemon (static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay-daemon ./cloud-relay/cmd/relay

# Stage 2: Runtime
FROM alpine:3.21

# Install Postfix, WireGuard tools, CA certificates, and gettext (for envsubst)
RUN apk add --no-cache \
    postfix \
    ca-certificates \
    wireguard-tools \
    gettext \
    && postfix set-permissions \
    && postfix check

# Copy relay daemon binary from builder
COPY --from=builder /build/relay-daemon /usr/local/bin/relay-daemon

# Copy Postfix configuration files
COPY cloud-relay/postfix-config/main.cf /etc/postfix/main.cf.template
COPY cloud-relay/postfix-config/master.cf /etc/postfix/master.cf
COPY cloud-relay/postfix-config/transport /etc/postfix/transport

# Copy entrypoint script
COPY cloud-relay/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose SMTP port
EXPOSE 25

# Health check using Postfix status
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD postfix status || exit 1

ENTRYPOINT ["/entrypoint.sh"]
