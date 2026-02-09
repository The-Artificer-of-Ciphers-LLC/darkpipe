# DarkPipe Cloud Relay - Multi-stage Docker build
#
# Stage 1: Build the Go relay daemon
# Stage 2: Runtime container with Postfix + relay daemon + WireGuard tools
#
# Target size: <50MB (Alpine base ~5MB + Postfix ~15MB + wireguard-tools ~5MB + Go binary ~3-5MB + gettext ~2MB = ~30-35MB)

FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build relay daemon (static binary with stripped symbols)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay-daemon ./cloud-relay/cmd/relay

# Optional: UPX compression if binary exceeds 5MB (uncomment to enable)
# RUN apk add --no-cache upx && upx --best relay-daemon

# Stage 2: Runtime
FROM alpine:3.21

# Install Postfix, WireGuard tools, CA certificates, and gettext (for envsubst)
# Use --no-cache to avoid storing the apk cache layer (~20MB savings)
# Pin Postfix version for reproducibility
RUN apk add --no-cache \
    postfix~=3.9 \
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

# Labels for container metadata
LABEL org.opencontainers.image.source=https://github.com/trek-e/darkpipe
LABEL org.opencontainers.image.description="DarkPipe Cloud Relay - Ephemeral SMTP relay with WireGuard/mTLS transport"
LABEL org.opencontainers.image.licenses=AGPL-3.0

# Expose SMTP port
EXPOSE 25

# Health check using Postfix status
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD postfix status || exit 1

ENTRYPOINT ["/entrypoint.sh"]
