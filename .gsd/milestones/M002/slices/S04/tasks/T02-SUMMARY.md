---
id: T02
parent: S04
milestone: M002
provides:
  - Go version alignment across all go.mod files and Dockerfiles (go 1.25.7 / golang:1.25-alpine)
key_files:
  - deploy/setup/go.mod
  - cloud-relay/Dockerfile
  - home-device/profiles/Dockerfile
key_decisions:
  - Aligned to go 1.25.7 (existing root/profiles version) rather than upgrading to 1.25.8 — matches what the project already declared
  - Used golang:1.25-alpine (latest alpine variant) for Docker images — confirmed tag exists on Docker Hub
patterns_established:
  - none
observability_surfaces:
  - none
duration: 5m
verification_result: passed
completed_at: 2026-03-11
blocker_discovered: false
---

# T02: Aligned Go version to 1.25.7 across all go.mod files and golang:1.25-alpine in Dockerfiles

**Unified Go toolchain version across all three go.mod files and both Dockerfiles to eliminate version mismatch.**

## What Happened

The root `go.mod` and `home-device/profiles/go.mod` declared `go 1.25.7`, but `deploy/setup/go.mod` used `go 1.24.0` and both Dockerfiles used `golang:1.24-alpine`. Confirmed `golang:1.25-alpine` exists on Docker Hub (with alpine 3.21/3.22/3.23 variants). Updated:

1. `deploy/setup/go.mod`: `go 1.24.0` → `go 1.25.7`
2. `cloud-relay/Dockerfile`: `golang:1.24-alpine` → `golang:1.25-alpine`
3. `home-device/profiles/Dockerfile`: `golang:1.24-alpine` → `golang:1.25-alpine`

## Verification

- `grep "^go " go.mod deploy/setup/go.mod home-device/profiles/go.mod` — all show `go 1.25.7` ✅
- `grep "golang:" cloud-relay/Dockerfile home-device/profiles/Dockerfile` — both show `golang:1.25-alpine` ✅
- `go build ./...` from repo root — clean (no output) ✅
- `cd deploy/setup && go build ./...` — clean (no output) ✅

### Slice-level checks (cumulative):
- `.env.example` files exist for cloud-relay and home-device ✅ (T01)
- `grep -q "RELAY_" cloud-relay/.env.example` ✅ (T01)
- Go version alignment across go.mod and Dockerfiles ✅ (this task)
- `go build ./monitoring/...` — not yet (T03)
- `go vet ./monitoring/...` — not yet (T03)
- `cd deploy/setup && go test ./pkg/secrets/ ./pkg/config/ ./pkg/validate/ -v` — not yet (T04)
- `go vet ./deploy/setup/...` — not yet (T04)

## Diagnostics

Future agents can verify version consistency with:
```
grep "^go " go.mod deploy/setup/go.mod home-device/profiles/go.mod
grep "golang:" cloud-relay/Dockerfile home-device/profiles/Dockerfile
```

## Deviations

None.

## Known Issues

None.

## Files Created/Modified

- `deploy/setup/go.mod` — go directive updated from 1.24.0 to 1.25.7
- `cloud-relay/Dockerfile` — FROM image updated from golang:1.24-alpine to golang:1.25-alpine
- `home-device/profiles/Dockerfile` — FROM image updated from golang:1.24-alpine to golang:1.25-alpine
