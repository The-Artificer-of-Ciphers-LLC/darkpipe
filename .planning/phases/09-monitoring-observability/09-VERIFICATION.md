---
phase: 09-monitoring-observability
verified: 2026-02-14T16:45:19Z
status: gaps_found
score: 3/5 success criteria verified
gaps:
  - truth: "User can see mail queue depth and stuck/deferred message count at a glance"
    status: partial
    reason: "Libraries exist but not wired into user-facing interface (profile server)"
    artifacts:
      - path: "home-device/profiles/cmd/profile-server/main.go"
        issue: "No /status routes registered, no StatusAggregator initialization"
    missing:
      - "Register dashboard routes in profile server main.go"
      - "Initialize StatusAggregator with real dependencies"
      - "Start background monitoring services (log parser, cert watcher, alert evaluator)"
  - truth: "User can check delivery status of recent outbound messages (delivered, deferred, bounced)"
    status: partial
    reason: "Same as queue monitoring - libraries exist but not integrated into profile server"
    artifacts:
      - path: "home-device/profiles/cmd/profile-server/main.go"
        issue: "No status API endpoint wired"
    missing:
      - "Wire status dashboard as documented in Plan 09-03"
  - truth: "User receives an alert at least 14 days before any certificate expires, and again at 7 days if not renewed"
    status: partial
    reason: "Alert system exists but not running (no daemon started)"
    artifacts: []
    missing:
      - "Start alert evaluator periodic loop in profile server"
      - "Configure alert channels (email/webhook/CLI)"
---

# Phase 09: Monitoring & Observability Verification Report

**Phase Goal:** Users have clear visibility into whether their email system is healthy -- mail is flowing, queues are clear, certificates are valid, and the cloud relay container is running properly

**Verified:** 2026-02-14T16:45:19Z  
**Status:** gaps_found  
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can see mail queue depth and stuck/deferred message count at a glance | ⚠️ PARTIAL | Libraries exist (`monitoring/queue/mailq.go`), CLI command exists (`monitoring/status/cli.go`), dashboard template exists, but NOT wired into profile server main.go |
| 2 | User can check delivery status of recent outbound messages (delivered, deferred, bounced) | ⚠️ PARTIAL | Libraries exist (`monitoring/delivery/tracker.go`), aggregator exists, but profile server doesn't serve /status routes |
| 3 | Cloud relay container exposes health check endpoints that return pass/fail status | ✓ VERIFIED | Docker health check configured in cloud-relay/docker-compose.yml line 112-115 (`nc -z localhost 25`), Caddy routes with Basic Auth exist (lines 47-66) |
| 4 | Certificate rotation is configurable and rotations happen automatically without service interruption | ⚠️ PARTIAL | Libraries exist (`monitoring/cert/rotator.go` with 2/3 lifetime rule, exponential backoff), but no daemon running periodic checks |
| 5 | User receives an alert at least 14 days before any certificate expires, and again at 7 days if not renewed | ⚠️ PARTIAL | Alert system exists (`monitoring/alert/triggers.go` has 14-day/7-day thresholds), but alert evaluator not started in profile server |

**Score:** 1/5 truths fully verified (only #3), 4/5 partial (libraries exist, not wired)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `monitoring/health/checker.go` | Unified health check interface | ✓ EXISTS | Contains `type Checker struct` line 43, `Liveness()` and `Readiness()` methods |
| `monitoring/health/server.go` | HTTP handlers for /health/live and /health/ready | ✓ EXISTS | `LivenessHandler` at line 10, returns application/health+json |
| `monitoring/queue/mailq.go` | Postfix queue parser via postqueue -j | ✓ EXISTS | `GetQueueStats()` line 53, executes `postqueue -j` line 40 |
| `monitoring/delivery/tracker.go` | Ring buffer for recent deliveries | ✓ EXISTS | `type DeliveryTracker struct` line 11, thread-safe with RWMutex |
| `monitoring/alert/notifier.go` | Multi-channel alert dispatch | ✓ EXISTS | `type AlertNotifier struct` line 196, fan-out to email/webhook/CLI |
| `monitoring/alert/ratelimit.go` | Per-type rate limiting (1-hour window) | ✓ EXISTS | `type RateLimiter struct` line 10, `ShouldSend()` method |
| `monitoring/cert/watcher.go` | Certificate expiry monitoring | ✓ EXISTS | `type CertWatcher struct` line 26, 2/3-lifetime renewal rule |
| `monitoring/cert/rotator.go` | Automated renewal with exponential backoff | ✓ EXISTS | `RenewWithRetry` line 34, uses cenkalti/backoff/v4 line 47 |
| `monitoring/cert/dkim.go` | Quarterly DKIM rotation | ✓ EXISTS | `RotateDKIM` line 31, reuses Phase 4 selector format |
| `monitoring/status/aggregator.go` | Unified status collector | ✓ EXISTS | `type StatusAggregator struct` line 69, interfaces at lines 52-66 |
| `monitoring/status/cli.go` | darkpipe status command | ✓ EXISTS | `RunStatusCommand` line 15, supports --json and --watch flags |
| `monitoring/status/dashboard.go` | Web dashboard handler | ✓ EXISTS | `HandleDashboard` line 24, serves HTML + JSON API |
| `monitoring/status/push.go` | Push-based uptime pinger | ✓ EXISTS | `type HealthchecksPinger struct` line 14, Dead Man's Switch pattern |
| `home-device/profiles/cmd/profile-server/templates/status.html` | Dashboard template | ✓ EXISTS | Contains "Queue Depth" line 265, four-card layout |
| `tests/test-monitoring.sh` | Phase integration test suite | ✓ EXISTS | 285 lines, covers MON-01/02/03, CERT-03/04 |

**All 15 key artifacts exist and are substantive** (not stubs).

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `monitoring/alert/notifier.go` | `monitoring/alert/ratelimit.go` | `rateLimiter.ShouldSend` check | ✓ WIRED | Line 214: `if a.rateLimiter != nil && !a.rateLimiter.ShouldSend` |
| `monitoring/cert/rotator.go` | `cenkalti/backoff/v4` | Exponential backoff retry | ✓ WIRED | Line 47: `err := backoff.RetryNotify` |
| `monitoring/status/aggregator.go` | `monitoring/health/checker.go` | Readiness check | ✓ WIRED | Line 96: `healthStatus := s.health.Readiness(ctx)` via HealthChecker interface |
| `monitoring/status/aggregator.go` | `monitoring/queue/mailq.go` | Queue stats | ✓ WIRED | Line 103: `queueStats, err := s.queue()` via func field |
| `monitoring/status/aggregator.go` | `monitoring/delivery/tracker.go` | Delivery stats | ✓ WIRED | Lines 112, 118: `s.delivery.GetStats()`, `s.delivery.GetRecent(10)` via DeliveryTracker interface |
| `monitoring/status/aggregator.go` | `monitoring/cert/watcher.go` | Cert info | ✓ WIRED | Line 122: `certInfos, err := s.certs.CheckAll()` via CertWatcher interface |
| `home-device/docker-compose.yml` | `monitoring/health/server.go` | Docker HEALTHCHECK | ✓ WIRED | Line 296: healthcheck calls `/health/live` endpoint |
| `cloud-relay/docker-compose.yml` | Postfix SMTP | Docker HEALTHCHECK | ✓ WIRED | Line 113: `nc -z localhost 25` |
| `cloud-relay/caddy/Caddyfile` | Profile server monitoring | Reverse proxy routes | ✓ WIRED | Lines 47-66: `/health/*` and `/status/*` with Basic Auth |
| `home-device/profiles/cmd/profile-server/main.go` | `monitoring/status/dashboard.go` | Status routes | ✗ NOT_WIRED | **No /status routes registered in main.go** |
| `home-device/profiles/cmd/profile-server/main.go` | `monitoring/status/aggregator.go` | Status aggregator init | ✗ NOT_WIRED | **StatusAggregator not initialized** |

**9/11 key links wired**, 2/11 not wired (profile server integration missing).

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| **MON-01** (Mail queue health monitoring) | ⚠️ BLOCKED | Libraries complete, CLI works, but profile server doesn't serve /status API |
| **MON-02** (Mail delivery status visibility) | ⚠️ BLOCKED | Same - delivery tracker works, but no web interface |
| **MON-03** (Cloud relay container health checks) | ✓ SATISFIED | Docker health check configured, Caddy routes exist with Basic Auth |
| **CERT-03** (Certificate rotation) | ⚠️ BLOCKED | Rotator library complete (2/3 lifetime, exponential backoff), but no periodic execution daemon |
| **CERT-04** (Certificate expiry monitoring) | ⚠️ BLOCKED | Alert triggers exist (14-day warn, 7-day critical), but alert evaluator not running |

**1/5 requirements satisfied**, 4/5 blocked by missing integration.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| N/A | N/A | None found | N/A | All code is substantive, no TODOs/FIXMEs/placeholders in monitoring packages |

**No anti-patterns detected.** All code is production-ready, fully tested (47 tests in Plan 09-01, ~5 tests in Plan 09-02, 23 tests in Plan 09-03).

### Human Verification Required

#### 1. Profile Server Status Dashboard (after integration)

**Test:** Start profile server, navigate to http://localhost:8090/status in browser  
**Expected:** Dashboard shows four cards (Services, Queue, Deliveries, Certificates) with real-time data, colored status indicators (green/yellow/red), auto-refresh every 30 seconds  
**Why human:** Visual appearance, layout responsiveness, color coding accuracy

#### 2. CLI Status Command

**Test:** Run `darkpipe status` from command line  
**Expected:** Human-readable output with colored status (green=OK, yellow=WARN, red=ERROR), shows all four sections  
**Why human:** Terminal color rendering, output formatting, user experience

#### 3. Alert Delivery

**Test:** Trigger cert expiry condition (set cert to expire in 10 days), wait for alert evaluator loop  
**Expected:** Email/webhook/CLI alert delivered with correct severity (critical for 7 days, warn for 14 days)  
**Why human:** Email delivery, webhook POST reception, alert content accuracy

#### 4. Push Monitoring

**Test:** Configure MONITOR_HEALTHCHECK_URL to Healthchecks.io or UptimeRobot, wait 5 minutes  
**Expected:** External service receives ping, shows "up" status  
**Why human:** External service integration, network connectivity

#### 5. Docker Health Checks

**Test:** Run `docker ps` after containers start  
**Expected:** All containers show "healthy" status  
**Why human:** Docker orchestration behavior, health check timing

### Gaps Summary

**Three core gaps blocking Phase 9 completion:**

1. **Profile server integration missing:** All monitoring libraries exist and are tested, but `home-device/profiles/cmd/profile-server/main.go` doesn't initialize the StatusAggregator or register /status routes. This blocks MON-01 and MON-02 from being user-accessible.

2. **No monitoring daemon:** Alert evaluator and cert watcher need to run in background loops (periodic checks every 5 minutes for alerts, every 6 hours for certs). Currently no daemon starts these services.

3. **Requirements not marked complete:** REQUIREMENTS.md still shows CERT-03, CERT-04, MON-01, MON-02, MON-03 as unchecked despite libraries being complete.

**These are integration gaps, not implementation gaps.** All code exists and is tested. The gaps are wiring/plumbing:
- Add ~30 lines to profile-server/main.go to initialize aggregator and register routes
- Start 2-3 background goroutines for periodic monitoring tasks
- Update REQUIREMENTS.md checkboxes

**Plan 09-03 summary explicitly notes this:** "Remaining Work: Wire dashboard into profile-server main.go (integration step)"

---

_Verified: 2026-02-14T16:45:19Z_  
_Verifier: Claude (gsd-verifier)_
