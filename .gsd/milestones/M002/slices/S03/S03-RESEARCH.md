# S03: TLS & Input Hardening — Research

**Date:** 2026-03-11

## Summary

S03 covers three concerns: (1) explicit TLS minimum version on provider IMAP clients, (2) SMTP DATA message size limiting, and (3) template.HTML safety audit.

The TLS hardening is the primary deliverable — 7 provider files use `tls.Config{}` with only `ServerName` set, relying on Go's default MinVersion (TLS 1.2 since Go 1.18). While the current Go default is secure, explicit `MinVersion: tls.VersionTLS12` matches the project's own pattern in `transport/mtls/` and makes the security posture self-documenting. This is a straightforward, low-risk change.

The SMTP size limit is already implemented — `emersion/go-smtp` v0.24.0 enforces `MaxMessageBytes` at the connection level before the `Data()` reader sees bytes. The config already exposes `RELAY_MAX_MESSAGE_BYTES` with a 50MB default. No code changes needed; just verify and document.

The template.HTML audit is clean — the only usage is in `webui.go` where `fmt.Sprintf` builds HTML instructions from controlled inputs (authenticated email, system-generated password, server config hostname, platform from a closed if/else chain). The template file uses `html/template` which auto-escapes all fields except `.Instructions` (typed as `template.HTML`). Risk is low but a defensive `html.EscapeString` on interpolated values would be belt-and-suspenders.

## Recommendation

1. **TLS: Add `MinVersion: tls.VersionTLS12` to all 7 provider files.** One-line addition per file inside the existing `&tls.Config{}` literal. Follow the exact pattern from `transport/mtls/server/listener.go:62`.

2. **SMTP size limit: No code change needed.** The `emersion/go-smtp` library enforces `MaxMessageBytes` at the protocol level (SIZE extension advertisement + byte counting on DATA). The `RELAY_MAX_MESSAGE_BYTES` env var is already configurable with a 50MB default. Add a defense-in-depth `io.LimitReader` in `session.go` Data() as a secondary safeguard (optional, low priority).

3. **template.HTML: Audit complete, add defensive escaping.** Wrap `email` and any form-sourced values with `html.EscapeString()` inside the `fmt.Sprintf` calls that produce the `instructions` string. This guards against future changes that might introduce less controlled inputs.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| SMTP message size limit | `emersion/go-smtp` MaxMessageBytes | Library enforces at connection level with SIZE extension + byte counting |
| TLS config constants | `crypto/tls` stdlib | `tls.VersionTLS12` constant, no external deps |
| HTML escaping | `html.EscapeString()` stdlib | Standard library function for escaping HTML entities |

## Existing Code and Patterns

- `transport/mtls/server/listener.go:62` — Reference pattern for explicit `MinVersion: tls.VersionTLS12` in `tls.Config`
- `transport/mtls/client/connector.go:60` — Same pattern on client side
- `transport/health/checker.go:223` — Same pattern in health checker
- `deploy/setup/pkg/providers/gmail.go:43` — Example of current gap: `&tls.Config{ServerName: "imap.gmail.com"}` with no MinVersion
- `cloud-relay/relay/smtp/server.go:47` — `s.MaxMessageBytes = cfg.MaxMessageBytes` already wires config to server
- `cloud-relay/relay/config/config.go:74` — `RELAY_MAX_MESSAGE_BYTES` env var with 50MB default
- `cloud-relay/relay/smtp/session.go:72` — `io.Copy(buf, r)` reads from library-bounded reader
- `home-device/profiles/cmd/profile-server/webui.go:283` — `template.HTML(instructions)` — the only template.HTML usage

## Affected Provider Files (7 files, same change pattern)

1. `deploy/setup/pkg/providers/gmail.go:43`
2. `deploy/setup/pkg/providers/outlook.go:42`
3. `deploy/setup/pkg/providers/icloud.go:39`
4. `deploy/setup/pkg/providers/generic.go:59`
5. `deploy/setup/pkg/providers/dockermailserver.go:39`
6. `deploy/setup/pkg/providers/mailcow.go:73`
7. `deploy/setup/pkg/providers/mailu.go:64`

## Constraints

- Must add `"crypto/tls"` import to any provider file that doesn't already import it (all 7 already import `crypto/tls` for `tls.Dial`)
- `emersion/go-smtp` v0.24.0 is the pinned version — SIZE enforcement behavior is version-specific
- `html/template` auto-escapes `{{.DeviceName}}` and other fields but NOT `{{.Instructions}}` (typed as `template.HTML`)
- Web UI is behind Basic Auth — the email value is the authenticated user's own address, not arbitrary input

## Common Pitfalls

- **Assuming io.Copy is unbounded** — The `emersion/go-smtp` library wraps the DATA reader with byte counting before passing it to the `Data()` callback. The `io.Copy` in session.go is already bounded by `MaxMessageBytes`. Adding `io.LimitReader` is defense-in-depth, not a fix for an actual vulnerability.
- **Over-rotating on template.HTML** — The instructions HTML is built from controlled values (authenticated email, generated password, config hostname, closed set of platform strings). The XSS surface is theoretical, not practical. Defensive escaping is good practice but this is not a live vulnerability.
- **Breaking TLS for older servers** — All 7 providers connect to well-known IMAP servers (Gmail, Outlook, iCloud) or self-hosted servers (mailcow, mailu, docker-mailserver). All support TLS 1.2+. Setting MinVersion to 1.2 will not break any real connection.

## Open Risks

- None — this is a low-risk mechanical slice with well-understood changes and an existing reference pattern.

## template.HTML Audit Detail

The `template.HTML(instructions)` in `webui.go:283` accepts `fmt.Sprintf`-built HTML. The interpolated values and their sources:

| Value | Source | Risk |
|-------|--------|------|
| `email` | Basic Auth credential (authenticated user) | None — user controls their own email |
| `plainPassword` | `apppassword.GenerateAppPassword()` (crypto/rand) | None — system-generated alphanumeric |
| `h.Config.Hostname` | Server configuration (env var) | None — admin-controlled |
| `platformTitle` | Derived from `platform` only when `== "thunderbird"` or `== "outlook"` | None — closed set |
| `platform` (else branch) | Not interpolated into HTML in the else branch | N/A |

**Verdict:** Safe. Add `html.EscapeString()` as defense-in-depth.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| Go TLS | N/A — stdlib `crypto/tls` | No skill needed |
| Go SMTP | N/A — `emersion/go-smtp` library | No skill needed |
| Go templates | N/A — stdlib `html/template` | No skill needed |

## Sources

- `emersion/go-smtp` v0.24.0 source at `/Users/trekkie/go/pkg/mod/github.com/emersion/go-smtp@v0.24.0/` — confirmed MaxMessageBytes enforcement via byte counting in `conn.go` and SIZE extension advertisement
- Go stdlib `crypto/tls` — TLS 1.2 is the default MinVersion since Go 1.18 (source: Go documentation)
- Go stdlib `html/template` — auto-escapes all pipeline values except those typed as `template.HTML` (source: Go documentation)
