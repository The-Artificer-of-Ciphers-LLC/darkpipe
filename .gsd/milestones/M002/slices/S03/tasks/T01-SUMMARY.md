---
id: T01
parent: S03
milestone: M002
provides:
  - Explicit TLS 1.2 minimum on all 7 provider IMAP clients
  - Regression test enforcing MinVersion on all provider files
key_files:
  - deploy/setup/pkg/providers/gmail.go
  - deploy/setup/pkg/providers/outlook.go
  - deploy/setup/pkg/providers/icloud.go
  - deploy/setup/pkg/providers/generic.go
  - deploy/setup/pkg/providers/dockermailserver.go
  - deploy/setup/pkg/providers/mailcow.go
  - deploy/setup/pkg/providers/mailu.go
  - deploy/setup/pkg/providers/provider_test.go
key_decisions:
  - Used string-matching test (not AST parsing) for MinVersion enforcement — simple, robust, zero dependencies
patterns_established:
  - All provider tls.Config literals must include MinVersion: tls.VersionTLS12
observability_surfaces:
  - TestProviderTLSMinVersion fails if any provider file has tls.Config without MinVersion
  - grep MinVersion deploy/setup/pkg/providers/*.go for quick inspection
duration: ~5min
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T01: Add MinVersion TLS 1.2 to all provider IMAP clients and test

**Added `MinVersion: tls.VersionTLS12` to all 7 provider IMAP `tls.Config` literals and a regression test enforcing the pattern.**

## What Happened

Added `MinVersion: tls.VersionTLS12` to the `tls.Config{}` struct in each of the 7 provider files (gmail, outlook, icloud, generic, dockermailserver, mailcow, mailu). No import changes were needed — all files already import `crypto/tls`.

Added `TestProviderTLSMinVersion` to `provider_test.go` that scans all non-test `.go` files in the providers directory, finds any containing `tls.Config{`, and asserts each also contains `MinVersion`. The test will catch any future provider that omits the setting.

## Verification

- `go build ./...` (from `deploy/setup/`) — passed
- `go vet ./...` (from `deploy/setup/`) — passed
- `go test ./pkg/providers/ -run TestProviderTLS -v` — PASS, checked 7 provider files
- `grep -c MinVersion deploy/setup/pkg/providers/*.go` — 1 per provider file (7 total)

### Slice-level checks status (intermediate task — partial expected)

- [x] `go vet ./...` passes (deploy/setup module)
- [x] `go build ./...` passes (deploy/setup module)
- [x] `go test ... -run TestTLS -v` — new test asserts MinVersion on all providers
- [ ] `go test ... -run TestHTMLEscape -v` — not yet (T02/T03 scope)
- [x] `grep -c MinVersion` returns 7
- [x] No bare `tls.Config` without `MinVersion` in provider files (verified by test)

## Diagnostics

- `grep MinVersion deploy/setup/pkg/providers/*.go` — quick check all providers have explicit TLS min version
- Test failure in `TestProviderTLSMinVersion` — names the offending file if a future provider omits MinVersion

## Deviations

Build/test commands run from `deploy/setup/` directory (separate Go module), not project root. The task plan's commands assumed root-level execution.

## Known Issues

None.

## Files Created/Modified

- `deploy/setup/pkg/providers/gmail.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/outlook.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/icloud.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/generic.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/dockermailserver.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/mailcow.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/mailu.go` — added MinVersion: tls.VersionTLS12
- `deploy/setup/pkg/providers/provider_test.go` — added TestProviderTLSMinVersion
