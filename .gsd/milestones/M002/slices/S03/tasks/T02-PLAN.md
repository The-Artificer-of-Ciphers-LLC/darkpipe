---
estimated_steps: 4
estimated_files: 2
---

# T02: Add defensive HTML escaping in webui.go and verify SMTP size limit

**Slice:** S03 — TLS & Input Hardening
**Milestone:** M002

## Description

The `webui.go` handler builds an HTML string via `fmt.Sprintf` and injects it as `template.HTML(instructions)`, which bypasses `html/template` auto-escaping. While the interpolated values come from controlled sources (authenticated email, generated password, config hostname), wrapping them with `html.EscapeString()` is defense-in-depth against future changes. This task also verifies the existing SMTP `MaxMessageBytes` enforcement (no code change needed — already implemented).

## Steps

1. In `home-device/profiles/cmd/profile-server/webui.go`, add `"html"` to the import block. In each `fmt.Sprintf` call that builds `instructions` HTML, wrap interpolated values (`email`, `plainPassword`, `h.Config.Hostname`) with `html.EscapeString()`. There are multiple `fmt.Sprintf` blocks (platform-specific: thunderbird, outlook, generic) — update all of them.
2. Add a test `TestInstructionsHTMLEscaping` in `handlers_test.go` that exercises the add-device handler with an email containing HTML special characters (e.g., `test+<script>@example.com`) and verifies the response body contains the escaped form (`&lt;script&gt;`) and does NOT contain the raw `<script>` tag.
3. Verify SMTP size limit: confirm `cloud-relay/relay/smtp/server.go:47` sets `s.MaxMessageBytes = cfg.MaxMessageBytes` and `cloud-relay/relay/config/config.go:64` defaults to `50*1024*1024`. No code changes — record finding in task summary.
4. Run `go build ./home-device/...` and `go test ./home-device/profiles/cmd/profile-server/ -v` to verify everything passes. Run `go vet ./...` for full project check.

## Must-Haves

- [ ] All `fmt.Sprintf` calls building `instructions` HTML use `html.EscapeString()` on interpolated values
- [ ] Test proves HTML special characters in email are escaped in output
- [ ] SMTP `MaxMessageBytes` confirmed as already implemented (documented, no code change)
- [ ] `go build` and `go vet` pass

## Verification

- `go build ./home-device/...` passes
- `go test ./home-device/profiles/cmd/profile-server/ -run TestInstructions -v` passes
- `go vet ./...` passes
- `grep 'html.EscapeString' home-device/profiles/cmd/profile-server/webui.go` shows escaping applied

## Observability Impact

- Signals added/changed: None — defensive escaping doesn't change runtime behavior for normal inputs
- How a future agent inspects this: `grep html.EscapeString webui.go`; test proves escaping works
- Failure state exposed: Test failure if escaping is removed or bypassed

## Inputs

- `home-device/profiles/cmd/profile-server/webui.go` — file to modify (lines ~230-285)
- `home-device/profiles/cmd/profile-server/handlers_test.go` — existing test file to extend
- `cloud-relay/relay/smtp/server.go` — verify MaxMessageBytes wiring (read-only)
- `cloud-relay/relay/config/config.go` — verify env var default (read-only)
- S03 research confirming template.HTML audit findings

## Expected Output

- `home-device/profiles/cmd/profile-server/webui.go` — `html.EscapeString()` wrapping all interpolated values in instructions HTML
- `home-device/profiles/cmd/profile-server/handlers_test.go` — new `TestInstructionsHTMLEscaping` test
