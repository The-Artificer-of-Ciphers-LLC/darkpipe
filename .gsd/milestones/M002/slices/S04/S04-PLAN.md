# S04: Operational Quality

**Goal:** .env.example files document all config, Go versions are consistent, monitoring API errors return structured JSON, and deploy/setup has meaningful test coverage.
**Demo:** `.env.example` files exist for cloud-relay and home-device with grouped/documented vars; `go.mod` and Dockerfiles agree on Go version; monitoring endpoints return `{"error":"...","code":N}` on failure; `go test ./deploy/setup/pkg/...` covers secrets, config, and validate packages.

## Must-Haves

- `.env.example` at `cloud-relay/.env.example` and `home-device/.env.example` with all env vars grouped by feature, defaults documented, required vs optional marked
- Go version consistent across `go.mod` (root), `home-device/profiles/go.mod`, `deploy/setup/go.mod`, `cloud-relay/Dockerfile`, and `home-device/profiles/Dockerfile`
- `monitoring/health/server.go` and `monitoring/status/dashboard.go` return `Content-Type: application/json` with `{"error":"...","code":N}` on errors
- Unit tests for `deploy/setup/pkg/secrets`, `deploy/setup/pkg/config`, and `deploy/setup/pkg/validate` (ports, smtp; dns pure-logic only)

## Proof Level

- This slice proves: contract (unit tests, build checks, file existence assertions)
- Real runtime required: no (build + unit test + file checks sufficient)
- Human/UAT required: no

## Verification

- `test -f cloud-relay/.env.example && test -f home-device/.env.example` — both files exist
- `grep -q "RELAY_" cloud-relay/.env.example` — cloud-relay vars documented
- Go version grep: all `go.mod` files and Dockerfiles reference the same Go version
- `go build ./monitoring/...` — structured error changes compile
- `go vet ./monitoring/...` — no issues
- `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v` — all tests pass
- `go vet ./deploy/setup/...` — no issues

## Observability / Diagnostics

- Runtime signals: monitoring JSON error responses now include structured `{"error":"...","code":N}` instead of plain text — machine-parseable by agents and monitoring tools
- Inspection surfaces: health endpoint errors are self-describing JSON; `.env.example` files serve as documentation surface for config inspection
- Failure visibility: JSON error `code` field maps directly to HTTP status, making failure categorization unambiguous
- Redaction constraints: none (no secrets in error responses or .env.example values)

## Integration Closure

- Upstream surfaces consumed: `cloud-relay/relay/config/config.go` (env var names/defaults), `home-device/docker-compose.yml` (env var references), `cloud-relay/docker-compose.yml` (env var references)
- New wiring introduced in this slice: none (housekeeping and documentation improvements)
- What remains before the milestone is truly usable end-to-end: nothing — S04 is the final slice of M002

## Tasks

- [x] **T01: Create .env.example files for cloud-relay and home-device** `est:30m`
  - Why: Every env var should have a documented default so operators can configure DarkPipe without reading source code
  - Files: `cloud-relay/.env.example`, `home-device/.env.example`, `cloud-relay/relay/config/config.go`, `home-device/docker-compose.yml`, `cloud-relay/docker-compose.yml`
  - Do: Extract all env vars from config.go (22 vars with defaults) and docker-compose files. Group by feature (relay, TLS, queue, WireGuard, etc.). Mark required vs optional. Document defaults inline. Add cross-reference comments pointing to config.go / docker-compose.yml.
  - Verify: `test -f cloud-relay/.env.example && test -f home-device/.env.example && grep -q "RELAY_" cloud-relay/.env.example`
  - Done when: both files exist, all env vars from config.go and compose files are documented with defaults and grouping

- [x] **T02: Align Go version across go.mod files and Dockerfiles** `est:20m`
  - Why: go.mod declares go 1.25.7 (root, profiles) but deploy/setup uses 1.24.0, and Dockerfiles use golang:1.24-alpine — builds may silently lose language features
  - Files: `deploy/setup/go.mod`, `cloud-relay/Dockerfile`, `home-device/profiles/Dockerfile`
  - Do: Check if `golang:1.25-alpine` exists on Docker Hub. If yes, update both Dockerfiles to `golang:1.25-alpine` and update `deploy/setup/go.mod` to `go 1.25.7`. If Go 1.25 image doesn't exist, downgrade root and profiles go.mod to match available Docker image. Verify all build.
  - Verify: `grep "go 1\." go.mod deploy/setup/go.mod home-device/profiles/go.mod` shows same version; `grep "golang:" cloud-relay/Dockerfile home-device/profiles/Dockerfile` shows matching tag; `go build ./...` passes
  - Done when: all go.mod files and Dockerfiles reference the same Go minor version and builds pass

- [x] **T03: Replace plain-text HTTP errors with structured JSON in monitoring endpoints** `est:25m`
  - Why: monitoring/health and monitoring/status serve JSON APIs but return plain-text errors — inconsistent and not machine-parseable
  - Files: `monitoring/health/server.go`, `monitoring/status/dashboard.go`
  - Do: Replace `http.Error()` calls on JSON API paths with a helper that writes `{"error":"...","code":N}` with `Content-Type: application/json`. Leave the HTML dashboard handler's template-render error as-is (it serves HTML). Add the helper as a local function in each file (separate go.mod modules).
  - Verify: `go build ./monitoring/... && go vet ./monitoring/...` passes
  - Done when: all JSON API error paths return structured JSON with correct Content-Type; HTML-serving paths unchanged; builds clean

- [x] **T04: Add unit tests for deploy/setup secrets, config, and validate packages** `est:45m`
  - Why: 6 packages in deploy/setup have zero tests — secrets, config, and validate are pure logic with the best test surface
  - Files: `deploy/setup/pkg/secrets/secrets_test.go`, `deploy/setup/pkg/config/config_test.go`, `deploy/setup/pkg/validate/ports_test.go`, `deploy/setup/pkg/validate/smtp_test.go`
  - Do: Write table-driven tests for secrets (GeneratePassword length/charset, GenerateSecrets creates files, ReadSecret/ListSecrets round-trip). Test config (DefaultConfig values, YAML marshal/unmarshal round-trip). Test validate/ports (CheckPort on known-open and known-closed ports) and validate/smtp (SMTP check structure). Skip dns.go tests (external DNS dependency, flaky). Use t.TempDir() for filesystem tests.
  - Verify: `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v -count=1` — all pass; `go vet ./pkg/...`
  - Done when: all new tests pass, go vet clean, tests cover core functions in each package

## Files Likely Touched

- `cloud-relay/.env.example`
- `home-device/.env.example`
- `deploy/setup/go.mod`
- `cloud-relay/Dockerfile`
- `home-device/profiles/Dockerfile`
- `monitoring/health/server.go`
- `monitoring/status/dashboard.go`
- `deploy/setup/pkg/secrets/secrets_test.go`
- `deploy/setup/pkg/config/config_test.go`
- `deploy/setup/pkg/validate/ports_test.go`
- `deploy/setup/pkg/validate/smtp_test.go`
