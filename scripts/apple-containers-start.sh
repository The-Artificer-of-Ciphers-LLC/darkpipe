#!/usr/bin/env bash
# apple-containers-start.sh — Orchestrate DarkPipe cloud relay on Apple Containers.
#
# Apple Containers has no compose equivalent. This script translates the
# cloud-relay docker-compose.yml (2 services: caddy + relay) into individual
# `container` CLI commands with proper networking, volumes, environment
# variables, and startup ordering.
#
# Subcommands:  up | down | status | logs
# Flags:        --dry-run  --verbose  --help
#
# Exit codes:
#   0 — success
#   1 — general error
#   2 — usage error
#   3 — network setup failed
#   4 — image build failed
#   5 — container start failed
#   6 — readiness check failed

set -euo pipefail

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
NETWORK_NAME="darkpipe"
CADDY_IMAGE="caddy:2-alpine"
CADDY_CONTAINER="caddy"
RELAY_CONTAINER="darkpipe-relay"
RELAY_DOCKERFILE="cloud-relay/Dockerfile"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$REPO_ROOT/cloud-relay/.env"
READINESS_TIMEOUT="${READINESS_TIMEOUT:-30}"

# ---------------------------------------------------------------------------
# State
# ---------------------------------------------------------------------------
DRY_RUN=false
VERBOSE=false

# ---------------------------------------------------------------------------
# Logging (matches check-runtime.sh output style)
# ---------------------------------------------------------------------------
_ts() { date "+%Y-%m-%d %H:%M:%S"; }

log_info()  { printf "[%s] [INFO]  %s\n" "$(_ts)" "$1"; }
log_pass()  { printf "[%s] [PASS]  %s\n" "$(_ts)" "$1"; }
log_fail()  { printf "[%s] [FAIL]  %s\n" "$(_ts)" "$1"; }
log_debug() { [[ "$VERBOSE" == true ]] && printf "[%s] [DEBUG] %s\n" "$(_ts)" "$1"; return 0; }
log_cmd()   { printf "[%s] [CMD]   %s\n" "$(_ts)" "$1"; }

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
run_cmd() {
  # Execute or print a command depending on DRY_RUN mode.
  local cmd="$*"
  if [[ "$DRY_RUN" == true ]]; then
    log_cmd "$cmd"
    return 0
  fi
  log_debug "exec: $cmd"
  eval "$cmd"
}

die() {
  log_fail "$1"
  exit "${2:-1}"
}

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
  cat <<'EOF'
Usage: apple-containers-start.sh <subcommand> [options]

Orchestrate DarkPipe cloud relay services on Apple Containers.
Translates docker-compose.yml into individual `container` CLI commands.

Subcommands:
  up        Create network, build images, start containers, run readiness checks
  down      Stop and remove containers, then remove the network
  status    List running DarkPipe containers
  logs      Show logs for DarkPipe containers (pass container name to filter)

Options:
  --dry-run   Print commands without executing (verify without macOS 26)
  --verbose   Show full command details and environment variable expansion
  --help      Show this help message and exit

Environment:
  READINESS_TIMEOUT   Seconds to wait for readiness checks (default: 30)

Examples:
  # Preview what would happen
  apple-containers-start.sh --dry-run up

  # Start the cloud relay
  apple-containers-start.sh up

  # Check container status
  apple-containers-start.sh status

  # View relay logs
  apple-containers-start.sh logs darkpipe-relay

  # Tear everything down
  apple-containers-start.sh down

Requires:
  - macOS 26+ on Apple Silicon
  - Apple Containers CLI (`container`) installed and `container system start` run
  - cloud-relay/.env populated with relay configuration

Exit codes:
  0  Success
  1  General error
  2  Usage error
  3  Network setup failed
  4  Image build failed
  5  Container start failed
  6  Readiness check failed
EOF
  exit 0
}

# ---------------------------------------------------------------------------
# Environment loading
# ---------------------------------------------------------------------------
load_env() {
  # Source relay environment variables from cloud-relay/.env if it exists.
  if [[ -f "$ENV_FILE" ]]; then
    log_info "Loading environment from $ENV_FILE"
    # Export variables from .env (skip comments and blank lines)
    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a
    log_debug "RELAY_HOSTNAME=${RELAY_HOSTNAME:-<unset>}"
    log_debug "RELAY_DOMAIN=${RELAY_DOMAIN:-<unset>}"
    log_debug "RELAY_TRANSPORT=${RELAY_TRANSPORT:-<unset>}"
  else
    log_info "No .env file found at $ENV_FILE — using defaults/environment"
  fi
}

# ---------------------------------------------------------------------------
# Network management
# ---------------------------------------------------------------------------
network_create() {
  log_info "Creating network: $NETWORK_NAME"
  if [[ "$DRY_RUN" == false ]]; then
    # Check if network already exists
    if container network list 2>/dev/null | grep -q "$NETWORK_NAME"; then
      log_info "Network '$NETWORK_NAME' already exists — skipping"
      return 0
    fi
  fi
  run_cmd "container network create $NETWORK_NAME"
  log_pass "Network '$NETWORK_NAME' created"
}

network_remove() {
  log_info "Removing network: $NETWORK_NAME"
  run_cmd "container network rm $NETWORK_NAME"
  log_pass "Network '$NETWORK_NAME' removed"
}

# ---------------------------------------------------------------------------
# Image building
# ---------------------------------------------------------------------------
build_relay_image() {
  log_info "Building relay image from $RELAY_DOCKERFILE"
  run_cmd "container build -t darkpipe-relay -f $REPO_ROOT/$RELAY_DOCKERFILE $REPO_ROOT"
  log_pass "Relay image built"
}

pull_caddy_image() {
  log_info "Pulling Caddy image: $CADDY_IMAGE"
  run_cmd "container pull $CADDY_IMAGE"
  log_pass "Caddy image pulled"
}

# ---------------------------------------------------------------------------
# Container run: Caddy
# ---------------------------------------------------------------------------
start_caddy() {
  log_info "Starting container: $CADDY_CONTAINER"

  # Resolve env vars with defaults matching docker-compose.yml
  local relay_domain="${RELAY_DOMAIN:-example.com}"
  local webmail_domains="mail.${relay_domain}"
  local autoconfig_domains="autoconfig.${relay_domain}"
  local autodiscover_domains="autodiscover.${relay_domain}"

  # Host directories for bind mounts (Apple Containers uses bind mounts, not named volumes)
  local data_dir="$REPO_ROOT/data/caddy"
  local caddy_data="$data_dir/data"
  local caddy_config="$data_dir/config"
  local caddy_logs="$data_dir/logs"
  local caddyfile="$REPO_ROOT/cloud-relay/caddy/Caddyfile"

  if [[ "$DRY_RUN" == false ]]; then
    mkdir -p "$caddy_data" "$caddy_config" "$caddy_logs"
  fi

  run_cmd "container run -d \
    --name $CADDY_CONTAINER \
    --network $NETWORK_NAME \
    --memory 128M \
    --read-only \
    -p 80:80 \
    -p 443:443 \
    -p 443:443/udp \
    -v $caddyfile:/etc/caddy/Caddyfile:ro \
    -v $caddy_data:/data \
    -v $caddy_config:/config \
    -v $caddy_logs:/var/log/caddy \
    -e WEBMAIL_DOMAINS=$webmail_domains \
    -e AUTOCONFIG_DOMAINS=$autoconfig_domains \
    -e AUTODISCOVER_DOMAINS=$autodiscover_domains \
    $CADDY_IMAGE"

  log_pass "Container '$CADDY_CONTAINER' started"
}

# ---------------------------------------------------------------------------
# Container run: Relay
# ---------------------------------------------------------------------------
start_relay() {
  log_info "Starting container: $RELAY_CONTAINER"

  # Resolve env vars with defaults — mTLS is default transport for Apple Containers
  # (WireGuard kernel module availability is unconfirmed in Apple's custom Linux kernel)
  local relay_hostname="${RELAY_HOSTNAME:-relay.example.com}"
  local relay_domain="${RELAY_DOMAIN:-example.com}"
  local relay_listen="${RELAY_LISTEN_ADDR:-127.0.0.1:10025}"
  local relay_transport="${RELAY_TRANSPORT:-mtls}"
  local relay_home="${RELAY_HOME_ADDR:-10.8.0.2:25}"
  local relay_max_bytes="${RELAY_MAX_MESSAGE_BYTES:-52428800}"
  local relay_ephemeral="${RELAY_EPHEMERAL_CHECK_INTERVAL:-60}"

  # Host directories for bind mounts
  local data_dir="$REPO_ROOT/data/relay"
  local postfix_queue="$data_dir/postfix-queue"
  local certbot_etc="$data_dir/certbot-etc"
  local queue_data="$data_dir/queue-data"

  if [[ "$DRY_RUN" == false ]]; then
    mkdir -p "$postfix_queue" "$certbot_etc" "$queue_data"
  fi

  # Build env var flags — core set always included
  local env_flags=""
  env_flags="$env_flags -e RELAY_HOSTNAME=$relay_hostname"
  env_flags="$env_flags -e RELAY_DOMAIN=$relay_domain"
  env_flags="$env_flags -e RELAY_LISTEN_ADDR=$relay_listen"
  env_flags="$env_flags -e RELAY_TRANSPORT=$relay_transport"
  env_flags="$env_flags -e RELAY_HOME_ADDR=$relay_home"
  env_flags="$env_flags -e RELAY_MAX_MESSAGE_BYTES=$relay_max_bytes"
  env_flags="$env_flags -e RELAY_EPHEMERAL_CHECK_INTERVAL=$relay_ephemeral"

  # Optional env vars — pass through only if set
  [[ -n "${RELAY_CA_CERT:-}" ]]       && env_flags="$env_flags -e RELAY_CA_CERT=$RELAY_CA_CERT"
  [[ -n "${RELAY_CLIENT_CERT:-}" ]]   && env_flags="$env_flags -e RELAY_CLIENT_CERT=$RELAY_CLIENT_CERT"
  [[ -n "${RELAY_CLIENT_KEY:-}" ]]    && env_flags="$env_flags -e RELAY_CLIENT_KEY=$RELAY_CLIENT_KEY"
  [[ -n "${RELAY_STRICT_MODE:-}" ]]   && env_flags="$env_flags -e RELAY_STRICT_MODE=$RELAY_STRICT_MODE"
  [[ -n "${RELAY_WEBHOOK_URL:-}" ]]   && env_flags="$env_flags -e RELAY_WEBHOOK_URL=$RELAY_WEBHOOK_URL"
  [[ -n "${CERTBOT_EMAIL:-}" ]]       && env_flags="$env_flags -e CERTBOT_EMAIL=$CERTBOT_EMAIL"
  [[ -n "${RELAY_QUEUE_ENABLED:-}" ]]        && env_flags="$env_flags -e RELAY_QUEUE_ENABLED=$RELAY_QUEUE_ENABLED"
  [[ -n "${RELAY_QUEUE_KEY_PATH:-}" ]]       && env_flags="$env_flags -e RELAY_QUEUE_KEY_PATH=$RELAY_QUEUE_KEY_PATH"
  [[ -n "${RELAY_QUEUE_SNAPSHOT_PATH:-}" ]]  && env_flags="$env_flags -e RELAY_QUEUE_SNAPSHOT_PATH=$RELAY_QUEUE_SNAPSHOT_PATH"
  [[ -n "${RELAY_OVERFLOW_ENABLED:-}" ]]     && env_flags="$env_flags -e RELAY_OVERFLOW_ENABLED=$RELAY_OVERFLOW_ENABLED"
  [[ -n "${RELAY_OVERFLOW_ENDPOINT:-}" ]]    && env_flags="$env_flags -e RELAY_OVERFLOW_ENDPOINT=$RELAY_OVERFLOW_ENDPOINT"
  [[ -n "${RELAY_OVERFLOW_BUCKET:-}" ]]      && env_flags="$env_flags -e RELAY_OVERFLOW_BUCKET=$RELAY_OVERFLOW_BUCKET"
  [[ -n "${RELAY_OVERFLOW_ACCESS_KEY:-}" ]]  && env_flags="$env_flags -e RELAY_OVERFLOW_ACCESS_KEY=$RELAY_OVERFLOW_ACCESS_KEY"
  [[ -n "${RELAY_OVERFLOW_SECRET_KEY:-}" ]]  && env_flags="$env_flags -e RELAY_OVERFLOW_SECRET_KEY=$RELAY_OVERFLOW_SECRET_KEY"

  # Note: No --cap-add/--cap-drop (Apple Containers VM isolation replaces Linux capabilities)
  # Note: No --device /dev/net/tun (unavailable — mTLS transport is default)
  run_cmd "container run -d \
    --name $RELAY_CONTAINER \
    --network $NETWORK_NAME \
    --memory 256M \
    --read-only \
    -p 25:25 \
    -v $postfix_queue:/var/spool/postfix \
    -v $certbot_etc:/etc/letsencrypt:ro \
    -v $queue_data:/data \
    $env_flags \
    darkpipe-relay"

  log_pass "Container '$RELAY_CONTAINER' started"
}

# ---------------------------------------------------------------------------
# Readiness checks
# ---------------------------------------------------------------------------
check_readiness() {
  log_info "Running readiness checks (timeout: ${READINESS_TIMEOUT}s)"

  if [[ "$DRY_RUN" == true ]]; then
    log_cmd "curl -sf --max-time 5 http://localhost:2019/config/  # caddy admin API"
    log_cmd "nc -z localhost 25  # relay SMTP port"
    log_pass "Readiness checks (dry-run — skipped)"
    return 0
  fi

  local deadline=$((SECONDS + READINESS_TIMEOUT))

  # Check Caddy admin API
  log_info "Waiting for Caddy admin API (localhost:2019)..."
  while [[ $SECONDS -lt $deadline ]]; do
    if curl -sf --max-time 5 http://localhost:2019/config/ > /dev/null 2>&1; then
      log_pass "Caddy admin API is responding"
      break
    fi
    sleep 2
  done
  if [[ $SECONDS -ge $deadline ]]; then
    die "Caddy admin API did not respond within ${READINESS_TIMEOUT}s — check 'container logs $CADDY_CONTAINER'" 6
  fi

  # Reset deadline for relay check
  deadline=$((SECONDS + READINESS_TIMEOUT))

  # Check relay SMTP port
  log_info "Waiting for relay SMTP (localhost:25)..."
  while [[ $SECONDS -lt $deadline ]]; do
    if nc -z localhost 25 2>/dev/null; then
      log_pass "Relay SMTP port 25 is listening"
      break
    fi
    sleep 2
  done
  if [[ $SECONDS -ge $deadline ]]; then
    die "Relay SMTP port 25 did not open within ${READINESS_TIMEOUT}s — check 'container logs $RELAY_CONTAINER'" 6
  fi

  log_pass "All readiness checks passed"
}

# ---------------------------------------------------------------------------
# Subcommands
# ---------------------------------------------------------------------------
cmd_up() {
  log_info "=== DarkPipe Cloud Relay — Apple Containers: UP ==="

  load_env
  network_create
  pull_caddy_image
  build_relay_image
  start_caddy
  start_relay
  check_readiness

  log_pass "=== All services started ==="
}

cmd_down() {
  log_info "=== DarkPipe Cloud Relay — Apple Containers: DOWN ==="

  # Stop containers (ignore errors if not running)
  log_info "Stopping containers..."
  run_cmd "container stop $CADDY_CONTAINER" || true
  run_cmd "container stop $RELAY_CONTAINER" || true
  log_pass "Containers stopped"

  # Remove containers
  log_info "Removing containers..."
  run_cmd "container rm $CADDY_CONTAINER" || true
  run_cmd "container rm $RELAY_CONTAINER" || true
  log_pass "Containers removed"

  # Remove network
  network_remove

  log_pass "=== All services torn down ==="
}

cmd_status() {
  log_info "=== DarkPipe Cloud Relay — Apple Containers: STATUS ==="
  run_cmd "container list"
}

cmd_logs() {
  local target="${1:-}"
  if [[ -n "$target" ]]; then
    log_info "Showing logs for: $target"
    run_cmd "container logs $target"
  else
    log_info "Showing logs for all DarkPipe containers"
    log_info "--- $CADDY_CONTAINER ---"
    run_cmd "container logs $CADDY_CONTAINER" || true
    log_info "--- $RELAY_CONTAINER ---"
    run_cmd "container logs $RELAY_CONTAINER" || true
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
  local subcommand=""
  local logs_target=""

  # Parse flags and subcommand
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run)  DRY_RUN=true;  shift ;;
      --verbose)  VERBOSE=true;  shift ;;
      --help|-h)  usage ;;
      up|down|status|logs)
        subcommand="$1"; shift
        # Capture optional target for logs subcommand
        if [[ "$subcommand" == "logs" && $# -gt 0 && ! "$1" =~ ^-- ]]; then
          logs_target="$1"; shift
        fi
        ;;
      *)
        log_fail "Unknown argument: $1"
        printf "\nRun '%s --help' for usage.\n" "$(basename "$0")" >&2
        exit 2
        ;;
    esac
  done

  if [[ -z "$subcommand" ]]; then
    log_fail "No subcommand specified"
    printf "\nRun '%s --help' for usage.\n" "$(basename "$0")" >&2
    exit 2
  fi

  [[ "$DRY_RUN" == true ]] && log_info "DRY-RUN mode — commands will be printed, not executed"
  [[ "$VERBOSE" == true ]] && log_info "VERBOSE mode enabled"

  case "$subcommand" in
    up)     cmd_up ;;
    down)   cmd_down ;;
    status) cmd_status ;;
    logs)   cmd_logs "$logs_target" ;;
  esac
}

main "$@"
