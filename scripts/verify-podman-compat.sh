#!/usr/bin/env bash
# verify-podman-compat.sh — Verify all compose files are compatible with both
# Docker Compose and Podman Compose. Exit non-zero if any check fails.
# Designed for CI and agent-driven verification.
#
# Check categories:
#   1. Compose validation: docker compose config --quiet passes for all base files
#   2. Health check syntax: no CMD-form arrays containing literal "||"
#   3. Version field: no deprecated version: field in compose files
#   4. Override layering: base + podman override passes docker compose config
#   5. Override layering: base + selinux override passes docker compose config
#   6. Swarm directives: no deploy.mode, deploy.replicas, deploy.placement
#   7. Podman-compose validation: if available, run podman-compose config on base files

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Base compose files
BASE_COMPOSE_FILES=(
  "cloud-relay/docker-compose.yml"
  "home-device/docker-compose.yml"
  "cloud-relay/certbot/docker-compose.certbot.yml"
)

# Override file pairs: base,override
PODMAN_OVERRIDES=(
  "cloud-relay/docker-compose.yml,cloud-relay/docker-compose.podman.yml"
  "home-device/docker-compose.yml,home-device/docker-compose.podman.yml"
)

SELINUX_OVERRIDES=(
  "cloud-relay/docker-compose.yml,cloud-relay/docker-compose.podman-selinux.yml"
  "home-device/docker-compose.yml,home-device/docker-compose.podman-selinux.yml"
)

# All compose files to scan (base + overrides)
ALL_COMPOSE_FILES=(
  "cloud-relay/docker-compose.yml"
  "home-device/docker-compose.yml"
  "cloud-relay/certbot/docker-compose.certbot.yml"
  "cloud-relay/docker-compose.podman.yml"
  "cloud-relay/docker-compose.podman-selinux.yml"
  "home-device/docker-compose.podman.yml"
  "home-device/docker-compose.podman-selinux.yml"
)

PASS=0
FAIL=0
SKIP=0
TOTAL=0

pass() {
  PASS=$((PASS + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ✅ PASS: %s\n" "$1"
}

fail() {
  FAIL=$((FAIL + 1))
  TOTAL=$((TOTAL + 1))
  printf "  ❌ FAIL: %s\n" "$1"
}

skip() {
  SKIP=$((SKIP + 1))
  printf "  ⏭️  SKIP: %s\n" "$1"
}

# ---------------------------------------------------------------------------
# 1. Compose validation: docker compose config --quiet
# ---------------------------------------------------------------------------
check_compose_validation() {
  printf "\n📋 Compose File Validation (docker compose config)\n"

  if ! command -v docker &>/dev/null; then
    skip "docker not installed — skipping docker compose config checks"
    return
  fi

  for file in "${BASE_COMPOSE_FILES[@]}"; do
    local filepath="${REPO_ROOT}/${file}"

    if [[ ! -f "$filepath" ]]; then
      fail "${file}: file not found"
      continue
    fi

    if docker compose -f "$filepath" config --quiet 2>/dev/null; then
      pass "${file}: docker compose config valid"
    else
      fail "${file}: docker compose config failed"
    fi
  done
}

# ---------------------------------------------------------------------------
# 2. Health check syntax: no CMD-form with literal "||"
# ---------------------------------------------------------------------------
check_healthcheck_syntax() {
  printf "\n🏥 Health Check Syntax (no CMD-form with shell operators)\n"

  for file in "${ALL_COMPOSE_FILES[@]}"; do
    local filepath="${REPO_ROOT}/${file}"

    if [[ ! -f "$filepath" ]]; then
      continue
    fi

    # Find CMD-form test arrays that contain literal "||" or "&&"
    # CMD-SHELL is fine — it's the CMD form with shell operators that's broken
    local bad_lines
    bad_lines=$(grep -n 'test:.*\["CMD"' "$filepath" | grep -E '\|\||&&' || true)

    if [[ -n "$bad_lines" ]]; then
      while IFS= read -r line; do
        local lineno
        lineno=$(echo "$line" | cut -d: -f1)
        fail "${file}:${lineno}: CMD-form health check contains shell operators"
      done <<< "$bad_lines"
    else
      pass "${file}: health check syntax OK"
    fi
  done
}

# ---------------------------------------------------------------------------
# 3. Version field: deprecated, breaks podman-compose
# ---------------------------------------------------------------------------
check_version_field() {
  printf "\n🏷️  Version Field (must not be present)\n"

  for file in "${ALL_COMPOSE_FILES[@]}"; do
    local filepath="${REPO_ROOT}/${file}"

    if [[ ! -f "$filepath" ]]; then
      continue
    fi

    if grep -qE '^version:' "$filepath"; then
      local lineno
      lineno=$(grep -nE '^version:' "$filepath" | head -1 | cut -d: -f1)
      fail "${file}:${lineno}: deprecated version field present"
    else
      pass "${file}: no version field"
    fi
  done
}

# ---------------------------------------------------------------------------
# 4. Override layering: base + podman override
# ---------------------------------------------------------------------------
check_podman_overrides() {
  printf "\n🔧 Podman Override Layering\n"

  if ! command -v docker &>/dev/null; then
    skip "docker not installed — skipping podman override layering checks"
    return
  fi

  for pair in "${PODMAN_OVERRIDES[@]}"; do
    IFS=',' read -r base override <<< "$pair"
    local base_path="${REPO_ROOT}/${base}"
    local override_path="${REPO_ROOT}/${override}"

    if [[ ! -f "$base_path" ]]; then
      fail "${base}: base file not found"
      continue
    fi
    if [[ ! -f "$override_path" ]]; then
      fail "${override}: override file not found"
      continue
    fi

    if docker compose -f "$base_path" -f "$override_path" config --quiet 2>/dev/null; then
      pass "${base} + ${override}: layering valid"
    else
      fail "${base} + ${override}: layering failed"
    fi
  done
}

# ---------------------------------------------------------------------------
# 5. Override layering: base + selinux override
# ---------------------------------------------------------------------------
check_selinux_overrides() {
  printf "\n🔒 SELinux Override Layering\n"

  if ! command -v docker &>/dev/null; then
    skip "docker not installed — skipping SELinux override layering checks"
    return
  fi

  for pair in "${SELINUX_OVERRIDES[@]}"; do
    IFS=',' read -r base override <<< "$pair"
    local base_path="${REPO_ROOT}/${base}"
    local override_path="${REPO_ROOT}/${override}"

    if [[ ! -f "$base_path" ]]; then
      fail "${base}: base file not found"
      continue
    fi
    if [[ ! -f "$override_path" ]]; then
      fail "${override}: override file not found"
      continue
    fi

    if docker compose -f "$base_path" -f "$override_path" config --quiet 2>/dev/null; then
      pass "${base} + ${override}: layering valid"
    else
      fail "${base} + ${override}: layering failed"
    fi
  done
}

# ---------------------------------------------------------------------------
# 6. Swarm directives: no deploy.mode, deploy.replicas, deploy.placement
# ---------------------------------------------------------------------------
check_swarm_directives() {
  printf "\n🐝 Swarm Directives (must not be present)\n"

  local swarm_keys=("mode:" "replicas:" "placement:")

  for file in "${BASE_COMPOSE_FILES[@]}"; do
    local filepath="${REPO_ROOT}/${file}"

    if [[ ! -f "$filepath" ]]; then
      continue
    fi

    local found_swarm=false

    # Look for Swarm-only keys inside deploy: blocks
    for key in "${swarm_keys[@]}"; do
      # Find deploy blocks and check for swarm keys within them
      local matches
      matches=$(awk '
        /^[[:space:]]+deploy:/ { in_deploy=1; indent=0; for(i=1;i<=length($0);i++) { if(substr($0,i,1)==" ") indent++; else break } next }
        in_deploy && /^[[:space:]]+[a-z]/ {
          cur_indent=0; for(i=1;i<=length($0);i++) { if(substr($0,i,1)==" ") cur_indent++; else break }
          if(cur_indent <= indent) { in_deploy=0; next }
        }
        in_deploy && /'"$key"'/ { print NR": "$0 }
      ' "$filepath" || true)

      if [[ -n "$matches" ]]; then
        found_swarm=true
        while IFS= read -r match; do
          fail "${file}: Swarm directive ${key} found at line ${match%%:*}"
        done <<< "$matches"
      fi
    done

    if [[ "$found_swarm" == "false" ]]; then
      pass "${file}: no Swarm directives"
    fi
  done
}

# ---------------------------------------------------------------------------
# 7. Podman-compose validation (optional — graceful skip)
# ---------------------------------------------------------------------------
check_podman_compose() {
  printf "\n🦭 Podman Compose Validation (optional)\n"

  if ! command -v podman-compose &>/dev/null; then
    skip "podman-compose not installed — skipping podman-compose config checks"
    return
  fi

  for file in "${BASE_COMPOSE_FILES[@]}"; do
    local filepath="${REPO_ROOT}/${file}"

    if [[ ! -f "$filepath" ]]; then
      fail "${file}: file not found"
      continue
    fi

    if podman-compose -f "$filepath" config --quiet 2>/dev/null; then
      pass "${file}: podman-compose config valid"
    else
      fail "${file}: podman-compose config failed"
    fi
  done
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
printf "=== DarkPipe Podman Compatibility Verification ===\n"
printf "Checking %d compose files (%d base, %d override)\n" \
  "${#ALL_COMPOSE_FILES[@]}" "${#BASE_COMPOSE_FILES[@]}" \
  $(( ${#ALL_COMPOSE_FILES[@]} - ${#BASE_COMPOSE_FILES[@]} ))

check_compose_validation
check_healthcheck_syntax
check_version_field
check_podman_overrides
check_selinux_overrides
check_swarm_directives
check_podman_compose

printf "\n=== Summary ===\n"
printf "Total: %d | Pass: %d | Fail: %d | Skip: %d\n" "$TOTAL" "$PASS" "$FAIL" "$SKIP"

if [[ $FAIL -gt 0 ]]; then
  printf "\n⚠️  %d check(s) failed. See details above.\n" "$FAIL"
  exit 1
else
  printf "\n✅ All checks passed.\n"
  exit 0
fi
