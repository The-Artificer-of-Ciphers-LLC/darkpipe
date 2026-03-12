---
id: T02
parent: S03
milestone: M002
provides:
  - Defense-in-depth HTML escaping on all template.HTML instructions in webui.go
  - Test proving XSS-like email inputs are escaped in all platform instruction blocks
  - Documented verification that SMTP MaxMessageBytes is already wired (50MB default)
key_files:
  - home-device/profiles/cmd/profile-server/webui.go
  - home-device/profiles/cmd/profile-server/handlers_test.go
key_decisions:
  - Wrapped all fmt.Sprintf interpolated values with html.EscapeString() including email, plainPassword, hostname, platformTitle, profileURL, and expiry — even controlled values, for defense-in-depth
  - Test constructs WebUIHandler directly (not via NewWebUIHandler) to avoid status.html template parsing which requires custom FuncMap not registered in the constructor
patterns_established:
  - All values interpolated into template.HTML instructions must be wrapped with html.EscapeString()
observability_surfaces:
  - grep html.EscapeString webui.go — confirms escaping applied
  - TestInstructionsHTMLEscaping — test failure if escaping is removed or bypassed
duration: ~10min
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Add defensive HTML escaping in webui.go and verify SMTP size limit

**Wrapped all interpolated values in webui.go instruction HTML with html.EscapeString() and confirmed SMTP MaxMessageBytes is already enforced at 50MB default.**

## What Happened

1. Added `"html"` import to `webui.go` and wrapped every interpolated value in all 4 `fmt.Sprintf` blocks (iOS/macOS, Android, Thunderbird/Outlook, Manual/Other) with `html.EscapeString()`. This covers: `email`, `plainPassword`, `h.Config.Hostname`, `platformTitle`, `profileURL`, and `expiry.Format()`.

2. Added `TestInstructionsHTMLEscaping` test in `handlers_test.go` that exercises the add-device handler with an email containing `<script>alert(1)</script>` across all 4 non-iOS platforms. Verifies the response contains `&lt;script&gt;` (escaped) and does NOT contain raw `<script>alert(1)</script>`.

3. Verified SMTP size limit: `cloud-relay/relay/smtp/server.go:47` sets `s.MaxMessageBytes = cfg.MaxMessageBytes` and `cloud-relay/relay/config/config.go:64` defaults to `50*1024*1024` (50MB). No code changes needed.

## Verification

- `go build ./...` — passes (home-device/profiles module)
- `go test ./cmd/profile-server/ -v` — all 12 tests pass including new `TestInstructionsHTMLEscaping` (4 subtests)
- `go vet ./...` — passes across all 3 modules (root, deploy/setup, home-device/profiles)
- `grep html.EscapeString webui.go` — shows escaping in all 4 instruction blocks
- Slice-level: `TestProviderTLSMinVersion` passes (T01), `TestInstructionsHTMLEscaping` passes (T02)

## Diagnostics

- `grep html.EscapeString home-device/profiles/cmd/profile-server/webui.go` — quick check escaping is applied
- Test failure in `TestInstructionsHTMLEscaping` names the platform if escaping is removed

## Deviations

- Test uses direct `WebUIHandler` construction instead of `NewWebUIHandler` because the latter calls `template.ParseFS(embedFS, "templates/*.html")` which includes `status.html` — that template uses a `mul` custom function registered only in `main.go`, not in the constructor. This is a pre-existing issue, not introduced by this task.
- iOS/macOS platform not tested in the HTML escaping test because it doesn't interpolate the email into the instructions HTML (it uses a QR code/profile URL flow instead).

## Known Issues

- `NewWebUIHandler` will panic if `status.html` is included in the embed glob because it references `mul` template function not registered in the `FuncMap`. This pre-dates this task and doesn't affect production (status template is loaded separately in `main.go`).

## Files Created/Modified

- `home-device/profiles/cmd/profile-server/webui.go` — Added `"html"` import; wrapped all interpolated values in 4 fmt.Sprintf instruction blocks with html.EscapeString()
- `home-device/profiles/cmd/profile-server/handlers_test.go` — Added `TestInstructionsHTMLEscaping` with 4 platform subtests; added setupTestWebUIHandler helper
