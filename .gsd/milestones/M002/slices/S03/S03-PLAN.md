# S03: TLS & Input Hardening

**Goal:** All 7 provider IMAP clients specify `MinVersion: tls.VersionTLS12` explicitly, template.HTML usage has defensive escaping, and SMTP size limit is verified documented.
**Demo:** `go vet ./...` and `go build ./...` pass, new unit tests confirm TLS config and HTML escaping, grep confirms no bare `tls.Config` without `MinVersion` in provider files.

## Must-Haves

- All 7 provider files in `deploy/setup/pkg/providers/` set `MinVersion: tls.VersionTLS12` in their `tls.Config`
- `webui.go` wraps interpolated values in `html.EscapeString()` before injecting into the `template.HTML` instructions string
- SMTP `MaxMessageBytes` enforcement is verified (existing — no code change, just confirmation)
- `go vet ./...` and `go build ./...` pass cleanly
- Existing tests continue to pass

## Proof Level

- This slice proves: contract
- Real runtime required: no
- Human/UAT required: no

## Verification

- `go vet ./...` passes
- `go build ./...` passes
- `go test ./deploy/setup/pkg/providers/ -run TestTLS -v` — new test asserts MinVersion on all providers
- `go test ./home-device/profiles/cmd/profile-server/ -run TestHTMLEscape -v` — new test asserts escaping in instructions
- `grep -c 'MinVersion' deploy/setup/pkg/providers/*.go` returns 7 (one per provider file)
- `grep 'tls.Config{' deploy/setup/pkg/providers/*.go | grep -v MinVersion` returns empty (no bare configs)

## Observability / Diagnostics

- Runtime signals: None — these are compile-time/config-level hardening changes, not runtime behavior changes
- Inspection surfaces: `grep MinVersion` on provider files; test output
- Failure visibility: Build failure or test failure if TLS config is misconfigured
- Redaction constraints: None

## Integration Closure

- Upstream surfaces consumed: `transport/mtls/server/listener.go` pattern for `MinVersion` usage (reference only)
- New wiring introduced in this slice: None — all changes are to existing config literals and string construction
- What remains before the milestone is truly usable end-to-end: S04 (Operational Quality) — `.env.example`, Go version alignment, JSON error responses, test coverage

## Tasks

- [x] **T01: Add MinVersion TLS 1.2 to all provider IMAP clients and test** `est:25m`
  - Why: 7 provider files use bare `tls.Config{ServerName: ...}` without explicit MinVersion, relying on Go defaults. Making it explicit matches the project's own `transport/mtls/` pattern and makes security posture self-documenting.
  - Files: `deploy/setup/pkg/providers/gmail.go`, `deploy/setup/pkg/providers/outlook.go`, `deploy/setup/pkg/providers/icloud.go`, `deploy/setup/pkg/providers/generic.go`, `deploy/setup/pkg/providers/dockermailserver.go`, `deploy/setup/pkg/providers/mailcow.go`, `deploy/setup/pkg/providers/mailu.go`, `deploy/setup/pkg/providers/provider_test.go`
  - Do: Add `MinVersion: tls.VersionTLS12` to each `&tls.Config{}` literal (one line per file). Add a test in `provider_test.go` that uses `go/ast` or source scanning to verify all provider files contain `MinVersion` in their `tls.Config`. Verify build and existing tests pass.
  - Verify: `go build ./deploy/setup/...` passes; `go test ./deploy/setup/pkg/providers/ -v` passes including new TLS test; `grep -c MinVersion deploy/setup/pkg/providers/*.go` returns 7
  - Done when: All 7 provider files have explicit `MinVersion: tls.VersionTLS12` and a test enforces it

- [x] **T02: Add defensive HTML escaping in webui.go and verify SMTP size limit** `est:20m`
  - Why: `template.HTML(instructions)` bypasses auto-escaping — wrapping interpolated values with `html.EscapeString()` is defense-in-depth. SMTP size limit is already implemented but needs verification confirmation.
  - Files: `home-device/profiles/cmd/profile-server/webui.go`, `home-device/profiles/cmd/profile-server/handlers_test.go`
  - Do: Wrap `email`, `plainPassword`, and `h.Config.Hostname` with `html.EscapeString()` in the `fmt.Sprintf` calls that build `instructions`. Add `"html"` import. Add a test that verifies HTML special characters in email/hostname are escaped in the rendered output. Verify SMTP `MaxMessageBytes` is wired (confirm `server.go:47` sets it from config, config default is 50MB) — document in slice summary, no code change needed.
  - Verify: `go build ./home-device/...` passes; `go test ./home-device/profiles/cmd/profile-server/ -v` passes including new escaping test; `go vet ./...` clean
  - Done when: `html.EscapeString` wraps all interpolated values in instructions HTML, test proves escaping works, SMTP size limit verified as already implemented

## Files Likely Touched

- `deploy/setup/pkg/providers/gmail.go`
- `deploy/setup/pkg/providers/outlook.go`
- `deploy/setup/pkg/providers/icloud.go`
- `deploy/setup/pkg/providers/generic.go`
- `deploy/setup/pkg/providers/dockermailserver.go`
- `deploy/setup/pkg/providers/mailcow.go`
- `deploy/setup/pkg/providers/mailu.go`
- `deploy/setup/pkg/providers/provider_test.go`
- `home-device/profiles/cmd/profile-server/webui.go`
- `home-device/profiles/cmd/profile-server/handlers_test.go`
