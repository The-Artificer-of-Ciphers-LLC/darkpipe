# S04: Operational Quality — Research

**Date:** 2026-03-11

## Summary

S04 covers four housekeeping concerns: `.env.example` files, Go version alignment, structured JSON error responses, and test coverage for untested deploy/setup packages. All four are low-risk, well-scoped, and independent of each other.

The Go version mismatch is the most impactful finding: `go.mod` files declare `go 1.25.7` (root, profiles) and `go 1.24.0` (deploy/setup), while both Dockerfiles use `golang:1.24-alpine`. This means the Go toolchain in Docker builds is older than what `go.mod` expects — builds may silently lose language features or fail outright if 1.25-only syntax is used.

The test coverage gap in `deploy/setup` is significant: 6 packages (`compose`, `config`, `migrate`, `wizard`, `secrets`, `validate`) totaling ~1,920 lines have zero test files. Only `mailmigrate` (6 test files) and `providers` (1 test file) are tested. The `secrets`, `config`, and `validate` packages are the best candidates — they are pure logic with no external dependencies, making them easy to unit test.

## Recommendation

Execute four independent tasks:

1. **`.env.example` files** — Generate from `cloud-relay/relay/config/config.go` (22 env vars with defaults) and `home-device/docker-compose.yml` (env vars for mail servers, profiles). Group by feature, document defaults and required-vs-optional.

2. **Go version alignment** — Update Dockerfiles to `golang:1.25-alpine` to match `go.mod`, and update `deploy/setup/go.mod` from `go 1.24.0` to `go 1.25.7` for consistency. Verify builds still pass.

3. **Structured JSON error responses** — Replace `http.Error()` plain-text responses in `monitoring/status/dashboard.go` and `monitoring/health/server.go` with `{"error": "...", "code": N}` JSON. The profile server handlers serve HTML pages and `.mobileconfig` files to browsers/devices, so plain-text errors are appropriate there — leave them alone.

4. **Test coverage** — Add unit tests for `deploy/setup/pkg/secrets`, `deploy/setup/pkg/config`, and `deploy/setup/pkg/validate` (ports, SMTP, DNS). These are pure-logic packages with straightforward test surfaces. Skip `wizard` (interactive I/O) and `compose` (template generation, lower ROI) for this slice.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| No new dependencies needed | Standard library only | All four tasks use `os`, `testing`, `encoding/json`, `net/http` |

## Existing Code and Patterns

- `cloud-relay/relay/config/config.go` — 22 env vars with `getEnv`/`getEnvInt64`/`getEnvBool` helpers and sensible defaults. Source of truth for cloud-relay `.env.example`.
- `cloud-relay/relay/config/config_test.go` — Existing config tests; reference pattern for env-based config testing.
- `home-device/docker-compose.yml` — Contains env var references for stalwart, maddy, postfix-dovecot, and profile server. Source of truth for home-device `.env.example`.
- `cloud-relay/docker-compose.yml` — Contains env var references for relay and Caddy. Additional source for cloud-relay `.env.example`.
- `monitoring/health/server.go` — Uses `http.Error()` for plain-text errors on JSON API endpoints (`application/health+json`). Should return structured JSON errors.
- `monitoring/status/dashboard.go` — `HandleStatusAPI` returns JSON on success but plain-text on error. Inconsistent.
- `home-device/profiles/cmd/profile-server/handlers.go` — Uses `http.Error()` throughout, but these serve browser/device clients downloading profiles. Plain-text errors are fine here.
- `deploy/setup/pkg/secrets/secrets.go` — 102 lines, pure crypto/filesystem logic, easy to test (GeneratePassword, GenerateSecrets, ReadSecret, ListSecrets).
- `deploy/setup/pkg/config/config.go` — 81 lines, YAML load/save with DefaultConfig. Easy to test.
- `deploy/setup/pkg/validate/` — 264 lines across dns.go, ports.go, smtp.go. DNS validation uses external DNS servers (needs mock or skip), but ports/smtp have testable structure.

## Constraints

- **Go version in Docker**: `golang:1.25-alpine` must exist on Docker Hub. Verify before changing Dockerfiles. If not available, use latest `golang:1.24-alpine` and downgrade `go.mod` to `1.24.x` instead.
- **deploy/setup/go.mod uses `go 1.24.0`**: May be intentional (wider compatibility) or just stale. Aligning to 1.25.7 should be safe since the binary runs on the user's machine, not in Docker.
- **dns.go validation**: Tests would need to mock DNS servers or use `net.Resolver` injection. Consider testing only the pure-logic parts or using integration-style tests that dial localhost.
- **`.env.example` location**: Place at `cloud-relay/.env.example` and `home-device/.env.example` (alongside their respective `docker-compose.yml` files).

## Common Pitfalls

- **Stale `.env.example`** — If the example file drifts from config.go, it becomes misleading. Add a comment in config.go pointing to the `.env.example` and vice versa.
- **Go version tag format** — Docker Hub uses `golang:1.25-alpine` not `golang:1.25.7-alpine` (minor tags may not exist). Check available tags.
- **JSON error response breaking clients** — The monitoring JSON API is internal. Changing error format from plain-text to JSON is safe since no external clients depend on it. The health endpoints follow RFC 8674 (`application/health+json`) — errors should also be structured.
- **DNS test flakiness** — Validating real DNS records in unit tests is inherently flaky. Test the validation logic with mock resolvers, or mark DNS tests as integration tests.

## Open Risks

- `golang:1.25-alpine` may not exist yet — need to check Docker Hub. If Go 1.25 isn't released, must align go.mod down to 1.24.x instead.
- `deploy/setup/go.mod` at `go 1.24.0` might be intentionally pinned for user compatibility (setup runs on user machines, not containers). Should verify no 1.25-specific syntax is used in that module.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Go testing | python-testing-patterns | not applicable (Python) |
| Go | N/A | no Go-specific skill installed or available |
| Docker | N/A | no Docker skill found in available_skills |

No relevant installable skills for this slice — it's standard Go, Docker, and HTTP work.

## Sources

- Codebase exploration (config.go, docker-compose.yml, Dockerfiles, go.mod files, handler files, test files)
