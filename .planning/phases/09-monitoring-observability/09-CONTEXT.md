# Phase 9: Monitoring & Observability - Context

**Gathered:** 2026-02-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Users have clear visibility into whether their email system is healthy — mail is flowing, queues are clear, certificates are valid, and the cloud relay container is running properly. Includes CLI status command, web dashboard, alerting system, certificate lifecycle automation, and container health check endpoints.

</domain>

<decisions>
## Implementation Decisions

### Health visibility & status display
- Both CLI (`darkpipe status`) and web dashboard for system health
- Four metric categories: mail queue depth & stuck messages, recent delivery status, certificate expiry countdown, tunnel/transport health
- JSON output via `--json` flag for scripting and Home Assistant/monitoring tool integration
- CLI for power users, web dashboard for household members

### Alert & notification behavior
- All delivery methods: email to admin, webhook (HTTP POST), CLI warning on next command
- All four trigger conditions: certificate expiry approaching, queue backup threshold, delivery failure spike, transport tunnel down
- Rate-limit per alert type (same alert type at most once per hour) to prevent notification storms during extended outages
- Certificate alerts at 14 days and 7 days before expiry (CERT-04 requirement)

### Certificate lifecycle management
- Renew at 2/3 of certificate lifetime (future-proof for Let's Encrypt moving from 90-day to 45-day certs through 2026-2028)
- For 90-day LE certs: renew at 60 days. For 45-day: renew at 30 days. For step-ca internal: configurable
- DKIM key rotation automated quarterly (matches Phase 4 selector format {prefix}-{YYYY}q{Q})
- Retry with exponential backoff on renewal failure (3 retries), then alert admin. Keep using old cert until actual expiry
- Let's Encrypt timeline awareness: 90-day default now, 45-day opt-in May 2026, 64-day default Feb 2027, 45-day default Feb 2028

### Container health checks & endpoints
- Deep readiness checks (actual service health: can Postfix accept mail? Is IMAP responding? Is tunnel connected?)
- Both per-container healthcheck endpoints (for Docker) AND unified aggregation endpoint (for user-facing status)
- Public health endpoint via Caddy with Basic Auth for remote monitoring
- Push-based pings to external uptime services (Healthchecks.io, UptimeRobot) via outbound HTTP — no inbound exposure needed

### Claude's Discretion
- Web dashboard location (profile server /status vs separate container)
- CLI auto-refresh behavior (one-shot vs watch mode)
- Per-user vs system-wide delivery stats
- Inbound vs outbound delivery tracking scope
- Delivery history retention period
- Alert severity levels (warn/critical vs single level)
- Certificate rotation service interruption strategy (hot reload vs brief restart)
- Health check aggregation architecture

</decisions>

<specifics>
## Specific Ideas

- "Under 2 minutes" device setup extends to monitoring — health should be glanceable, not require investigation
- JSON output enables the homelab ecosystem (Home Assistant, Grafana, Prometheus via json_exporter)
- Push-based monitoring avoids exposing additional inbound ports (security-first)
- LE cert timeline is actively changing — design for shorter lifetimes from day one

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-monitoring-observability*
*Context gathered: 2026-02-14*
