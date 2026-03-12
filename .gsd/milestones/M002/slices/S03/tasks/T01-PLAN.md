---
estimated_steps: 4
estimated_files: 8
---

# T01: Add MinVersion TLS 1.2 to all provider IMAP clients and test

**Slice:** S03 — TLS & Input Hardening
**Milestone:** M002

## Description

All 7 provider files use `&tls.Config{ServerName: "..."}` without explicit `MinVersion`, relying on Go's default (TLS 1.2 since Go 1.18). While secure today, explicit configuration matches the project's own pattern in `transport/mtls/` and makes the security posture self-documenting. This task adds `MinVersion: tls.VersionTLS12` to each provider and adds a test to enforce the pattern.

## Steps

1. Add `MinVersion: tls.VersionTLS12,` to the `&tls.Config{}` literal in each of the 7 provider files, directly after the `ServerName` field. The exact lines are:
   - `deploy/setup/pkg/providers/gmail.go:43`
   - `deploy/setup/pkg/providers/outlook.go:42`
   - `deploy/setup/pkg/providers/icloud.go:39`
   - `deploy/setup/pkg/providers/generic.go:59`
   - `deploy/setup/pkg/providers/dockermailserver.go:39`
   - `deploy/setup/pkg/providers/mailcow.go:73`
   - `deploy/setup/pkg/providers/mailu.go:64`
2. Verify all 7 files already import `crypto/tls` (they do — they use `tls.Dial`). No import changes needed.
3. Add a test `TestProviderTLSMinVersion` in `provider_test.go` that reads all `.go` files in the providers directory, finds `tls.Config` literals, and asserts each contains `MinVersion`. Use `os.ReadDir` + string matching (simple and robust — no need for AST parsing on a one-line check).
4. Run `go build ./deploy/setup/...` and `go test ./deploy/setup/pkg/providers/ -v` to verify everything passes.

## Must-Haves

- [ ] All 7 provider files have `MinVersion: tls.VersionTLS12` in their `tls.Config`
- [ ] Test exists that will catch any future provider file missing `MinVersion`
- [ ] `go build` and existing tests still pass

## Verification

- `go build ./deploy/setup/...` passes
- `go test ./deploy/setup/pkg/providers/ -run TestProviderTLS -v` passes
- `grep 'tls.Config{' deploy/setup/pkg/providers/*.go | grep -v MinVersion` returns empty

## Observability Impact

- Signals added/changed: None — compile-time config change
- How a future agent inspects this: `grep MinVersion deploy/setup/pkg/providers/*.go`
- Failure state exposed: Test failure if a new provider file omits MinVersion

## Inputs

- `transport/mtls/server/listener.go:62` — reference pattern for `MinVersion: tls.VersionTLS12`
- `deploy/setup/pkg/providers/*.go` — 7 files to modify
- `deploy/setup/pkg/providers/provider_test.go` — existing test file to extend

## Expected Output

- `deploy/setup/pkg/providers/gmail.go` — `MinVersion: tls.VersionTLS12` added
- `deploy/setup/pkg/providers/outlook.go` — same
- `deploy/setup/pkg/providers/icloud.go` — same
- `deploy/setup/pkg/providers/generic.go` — same
- `deploy/setup/pkg/providers/dockermailserver.go` — same
- `deploy/setup/pkg/providers/mailcow.go` — same
- `deploy/setup/pkg/providers/mailu.go` — same
- `deploy/setup/pkg/providers/provider_test.go` — new `TestProviderTLSMinVersion` test
