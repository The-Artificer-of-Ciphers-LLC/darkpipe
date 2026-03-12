---
estimated_steps: 5
estimated_files: 4
---

# T04: Add unit tests for deploy/setup secrets, config, and validate packages

**Slice:** S04 — Operational Quality
**Milestone:** M002

## Description

The deploy/setup module has 6 packages with zero test coverage. The `secrets`, `config`, and `validate` packages are pure logic with straightforward test surfaces — best candidates for unit tests. Skip `wizard` (interactive I/O), `compose` (template generation, lower ROI), and `dns.go` (external DNS dependency, flaky).

## Steps

1. Read `deploy/setup/pkg/secrets/secrets.go` to understand exported functions (GeneratePassword, GenerateSecrets, ReadSecret, ListSecrets) and write table-driven tests in `secrets_test.go` covering: password length/charset, GenerateSecrets creates expected files in temp dir, ReadSecret round-trips, ListSecrets returns created secrets
2. Read `deploy/setup/pkg/config/config.go` to understand exported functions (DefaultConfig, Load, Save) and write tests in `config_test.go` covering: DefaultConfig returns valid struct, YAML marshal/unmarshal round-trip, Load from nonexistent file returns error or default
3. Read `deploy/setup/pkg/validate/ports.go` and `smtp.go` to understand validation functions and write tests: ports_test.go tests CheckPort with localhost listener (open) and random high port (closed); smtp_test.go tests SMTP validation structure
4. Use `t.TempDir()` for all filesystem operations (auto-cleanup)
5. Run `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v -count=1 && go vet ./pkg/...`

## Must-Haves

- [ ] `secrets_test.go` tests GeneratePassword, GenerateSecrets, ReadSecret, ListSecrets
- [ ] `config_test.go` tests DefaultConfig and YAML round-trip
- [ ] `ports_test.go` tests port checking with real localhost listener
- [ ] `smtp_test.go` tests SMTP validation structure
- [ ] All tests pass with `go test -v -count=1`
- [ ] `go vet` clean on all packages

## Verification

- `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v -count=1` — all pass
- `cd deploy/setup && go vet ./pkg/...` — clean
- Test count: minimum 3 tests per package (secrets, config, validate)

## Observability Impact

- Signals added/changed: None (test files only)
- How a future agent inspects this: run `go test ./pkg/... -v` to see test results
- Failure state exposed: test failures with descriptive names indicate which function broke

## Inputs

- `deploy/setup/pkg/secrets/secrets.go` — functions to test
- `deploy/setup/pkg/config/config.go` — functions to test
- `deploy/setup/pkg/validate/ports.go` — port check functions
- `deploy/setup/pkg/validate/smtp.go` — SMTP validation functions

## Expected Output

- `deploy/setup/pkg/secrets/secrets_test.go` — unit tests for secrets package
- `deploy/setup/pkg/config/config_test.go` — unit tests for config package
- `deploy/setup/pkg/validate/ports_test.go` — unit tests for port validation
- `deploy/setup/pkg/validate/smtp_test.go` — unit tests for SMTP validation
