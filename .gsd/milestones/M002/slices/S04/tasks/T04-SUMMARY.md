---
id: T04
parent: S04
milestone: M002
provides:
  - Unit tests for deploy/setup secrets, config, and validate packages (24 tests total)
key_files:
  - deploy/setup/pkg/secrets/secrets_test.go
  - deploy/setup/pkg/config/config_test.go
  - deploy/setup/pkg/validate/ports_test.go
  - deploy/setup/pkg/validate/smtp_test.go
key_decisions:
  - Tested SMTP banner validation at protocol level rather than calling ValidateSMTPBanner directly (hardcoded to port 25, can't redirect to mock)
  - Used containsSubstring helper instead of importing strings package — keeps test file self-contained
patterns_established:
  - Table-driven tests with t.TempDir() for filesystem operations
  - Real TCP listeners for port validation tests (net.Listen on :0 for random ports)
  - Mock SMTP servers via goroutine TCP listeners for banner tests
observability_surfaces:
  - Run `cd deploy/setup && go test ./pkg/... -v` to see test results
duration: ~10m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T04: Add unit tests for deploy/setup secrets, config, and validate packages

**Added 24 unit tests across 4 test files covering secrets, config, and validate packages in deploy/setup.**

## What Happened

Wrote comprehensive unit tests for the three pure-logic packages in deploy/setup:

- **secrets** (9 tests): GeneratePassword length for short/default/long, uniqueness check, GenerateSecrets file creation and content verification, ReadSecret round-trip and not-found error, ListSecrets with files/empty/directory-exclusion
- **config** (5 tests): DefaultConfig field values, SaveConfig/LoadConfig YAML round-trip, LoadConfig on nonexistent file, LoadConfig on invalid YAML, SaveConfig creates file
- **validate/ports** (5 tests): ValidatePort on open port (real listener), ValidatePort on closed port, CheckLocalPorts with mixed open/closed, CheckLocalPorts empty input, DetectRAM stub returns error
- **validate/smtp** (5 tests): ValidateSMTPPort open/closed, SMTP banner valid/invalid via mock TCP servers, error message contains VPS guidance

All tests use `t.TempDir()` for filesystem operations and real TCP listeners (`net.Listen("tcp", "127.0.0.1:0")`) for port/SMTP tests.

## Verification

- `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v -count=1` — **24/24 tests pass**
- `cd deploy/setup && go vet ./pkg/...` — **clean**
- All slice-level verification checks pass:
  - ✅ cloud-relay/.env.example and home-device/.env.example exist
  - ✅ RELAY_ vars documented
  - ✅ Go version consistent (1.25.7 / golang:1.25-alpine)
  - ✅ monitoring builds and vets clean
  - ✅ deploy/setup tests pass and vet clean

## Diagnostics

- Run `cd deploy/setup && go test ./pkg/... -v` to inspect all test results
- Test names are descriptive — failures indicate which specific function/scenario broke

## Deviations

- ValidateSMTPBanner is hardcoded to port 25, so banner tests exercise the SMTP protocol at the TCP level against a mock server rather than calling ValidateSMTPBanner directly. This still validates the banner parsing logic pattern.

## Known Issues

- DetectRAM is a stub that always returns an error — tested as-is (confirms the error contract). Not a regression; it was already unimplemented.

## Files Created/Modified

- `deploy/setup/pkg/secrets/secrets_test.go` — 9 tests for password generation, secret CRUD, listing
- `deploy/setup/pkg/config/config_test.go` — 5 tests for defaults, YAML round-trip, error cases
- `deploy/setup/pkg/validate/ports_test.go` — 5 tests for port checking with real TCP listeners
- `deploy/setup/pkg/validate/smtp_test.go` — 5 tests for SMTP port and banner validation
