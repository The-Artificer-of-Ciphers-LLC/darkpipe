---
id: T03
parent: S04
milestone: M002
provides:
  - Structured JSON error responses on all monitoring JSON API paths
key_files:
  - monitoring/health/server.go
  - monitoring/status/dashboard.go
key_decisions:
  - Added jsonError helper as package-local function in each file (separate Go modules, can't share)
  - Left HandleDashboard http.Error calls unchanged — HTML-serving path, plain-text errors appropriate for browser clients
patterns_established:
  - jsonError(w, message, code) helper pattern for JSON API error responses — writes Content-Type application/json + {"error":"...","code":N}
observability_surfaces:
  - Health endpoint 405 errors now return machine-parseable JSON with error message and HTTP status code
  - Status API 500 errors now return machine-parseable JSON with error message and HTTP status code
duration: 5m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T03: Replace plain-text HTTP errors with structured JSON in monitoring endpoints

**Replaced http.Error() with structured JSON error responses on all monitoring JSON API paths while preserving plain-text errors on HTML dashboard path.**

## What Happened

Added a `jsonError(w, message, code)` helper to both `monitoring/health/server.go` and `monitoring/status/dashboard.go`. The helper writes `Content-Type: application/json`, sets the HTTP status code, and encodes `{"error":"...","code":N}`.

Replaced 4 `http.Error()` calls total:
- `health/server.go`: 2 calls (405 Method Not Allowed on liveness and readiness handlers)
- `status/dashboard.go`: 2 calls (500 Internal Server Error on HandleStatusAPI — get-status failure and JSON encode failure)

Left 2 `http.Error()` calls unchanged in `HandleDashboard` — these serve HTML to browser clients where plain-text errors are appropriate.

## Verification

- `go build ./monitoring/...` — passed
- `go vet ./monitoring/...` — passed
- Code review: grep confirms all JSON API paths use `jsonError`, HTML dashboard paths retain `http.Error`
- Slice-level checks: .env.example files exist, Go versions aligned, monitoring build/vet clean — all passing

## Diagnostics

- `curl -X POST <host>/health/live` → `{"error":"Method not allowed","code":405}` with `Content-Type: application/json`
- `curl <host>/api/status` on failure → `{"error":"...","code":500}` with `Content-Type: application/json`
- HTML dashboard errors remain plain-text via `http.Error`

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `monitoring/health/server.go` — Added jsonError helper, replaced 2 http.Error calls with jsonError on 405 paths
- `monitoring/status/dashboard.go` — Added jsonError helper, replaced 2 http.Error calls with jsonError on API 500 paths
