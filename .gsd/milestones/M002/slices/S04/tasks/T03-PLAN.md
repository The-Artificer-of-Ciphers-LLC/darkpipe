---
estimated_steps: 4
estimated_files: 2
---

# T03: Replace plain-text HTTP errors with structured JSON in monitoring endpoints

**Slice:** S04 — Operational Quality
**Milestone:** M002

## Description

The monitoring health and status endpoints serve JSON APIs (`application/health+json`) but return plain-text error responses via `http.Error()`. This is inconsistent and not machine-parseable. Replace these with structured `{"error":"...","code":N}` JSON responses. Leave HTML dashboard handler errors as-is since those serve browser clients.

## Steps

1. Read `monitoring/health/server.go` and `monitoring/status/dashboard.go` to identify all `http.Error()` call sites and which serve JSON vs HTML
2. Add a local `jsonError(w http.ResponseWriter, message string, code int)` helper to each file (separate modules, can't share) that writes `Content-Type: application/json` and encodes `{"error":"...","code":N}`
3. Replace `http.Error()` calls on JSON API paths with `jsonError()` — health endpoints (both 405s), status API endpoint (500s for get-status and encode failures). Leave `HandleDashboard` template-render error as `http.Error` (it serves HTML)
4. Build and vet: `go build ./monitoring/... && go vet ./monitoring/...`

## Must-Haves

- [ ] Health endpoint 405 errors return JSON with correct Content-Type
- [ ] Status API 500 errors return JSON with correct Content-Type
- [ ] HTML dashboard errors remain plain-text (appropriate for browser clients)
- [ ] `go build` and `go vet` pass cleanly

## Verification

- `go build ./monitoring/...` — compiles
- `go vet ./monitoring/...` — clean
- Code review: all JSON API error paths use jsonError helper

## Observability Impact

- Signals added/changed: monitoring error responses now machine-parseable JSON with explicit error code
- How a future agent inspects this: curl health/status endpoints, parse JSON error response
- Failure state exposed: structured error code and message in response body

## Inputs

- `monitoring/health/server.go` — 2 http.Error calls (405s)
- `monitoring/status/dashboard.go` — 4 http.Error calls (500s on API path, 500s on dashboard path)

## Expected Output

- `monitoring/health/server.go` — jsonError helper + updated error calls
- `monitoring/status/dashboard.go` — jsonError helper + updated error calls on API path only
