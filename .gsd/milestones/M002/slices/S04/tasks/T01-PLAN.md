---
estimated_steps: 4
estimated_files: 5
---

# T01: Create .env.example files for cloud-relay and home-device

**Slice:** S04 — Operational Quality
**Milestone:** M002

## Description

Generate `.env.example` files for both deployment targets so operators can configure DarkPipe without reading source code. Extract all environment variables from `config.go` (22 vars with defaults) and `docker-compose.yml` files. Group by feature, document defaults, and mark required vs optional.

## Steps

1. Read `cloud-relay/relay/config/config.go` to extract all env var names, types, and defaults from `getEnv`/`getEnvInt64`/`getEnvBool` calls
2. Read `cloud-relay/docker-compose.yml` and `home-device/docker-compose.yml` for additional env vars not in config.go (Caddy, mail server, profile server vars)
3. Write `cloud-relay/.env.example` with vars grouped by feature (relay core, TLS, queue, WireGuard, monitoring, Caddy), defaults shown as values, required vars commented as `# Required`, optional vars with their defaults
4. Write `home-device/.env.example` with vars grouped by feature (mail server selection, domain config, user config, profile server, Radicale), cross-referencing docker-compose.yml

## Must-Haves

- [ ] All env vars from config.go are present in cloud-relay/.env.example
- [ ] All env vars from both docker-compose.yml files are present in the appropriate .env.example
- [ ] Variables grouped by feature with section headers
- [ ] Defaults documented inline as values
- [ ] Required vs optional clearly marked
- [ ] Cross-reference comment pointing to source (config.go or docker-compose.yml)

## Verification

- `test -f cloud-relay/.env.example && test -f home-device/.env.example`
- `grep -q "RELAY_" cloud-relay/.env.example`
- `grep -c "=" cloud-relay/.env.example` shows ≥20 vars
- Both files have section header comments

## Observability Impact

- Signals added/changed: None (documentation files only)
- How a future agent inspects this: read .env.example to understand available configuration
- Failure state exposed: None

## Inputs

- `cloud-relay/relay/config/config.go` — source of truth for relay env vars and defaults
- `cloud-relay/docker-compose.yml` — Caddy and relay compose-level env vars
- `home-device/docker-compose.yml` — mail server and profile server env vars

## Expected Output

- `cloud-relay/.env.example` — complete env var documentation for cloud relay deployment
- `home-device/.env.example` — complete env var documentation for home device deployment
