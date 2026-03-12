---
estimated_steps: 4
estimated_files: 4
---

# T02: Align Go version across go.mod files and Dockerfiles

**Slice:** S04 — Operational Quality
**Milestone:** M002

## Description

The root `go.mod` and `home-device/profiles/go.mod` declare `go 1.25.7`, but `deploy/setup/go.mod` uses `go 1.24.0` and both Dockerfiles use `golang:1.24-alpine`. This mismatch means Docker builds use an older toolchain than what go.mod expects — builds may silently lose language features or fail if 1.25-only syntax is used. Align everything to the same version.

## Steps

1. Check if `golang:1.25-alpine` exists on Docker Hub (search tags or pull test). If it doesn't exist, determine the correct alignment strategy (downgrade go.mod or use latest available tag).
2. Update `deploy/setup/go.mod` from `go 1.24.0` to match the target version
3. Update `cloud-relay/Dockerfile` and `home-device/profiles/Dockerfile` FROM lines to use the matching `golang:X.XX-alpine` tag
4. Verify: run `go build ./...` from repo root and `cd deploy/setup && go build ./...` to confirm builds pass with aligned versions

## Must-Haves

- [ ] All three go.mod files specify the same Go minor version
- [ ] Both Dockerfiles use a golang alpine image matching the go.mod version
- [ ] `go build ./...` passes from repo root
- [ ] `cd deploy/setup && go build ./...` passes

## Verification

- `grep "^go " go.mod deploy/setup/go.mod home-device/profiles/go.mod` — all show same version
- `grep "golang:" cloud-relay/Dockerfile home-device/profiles/Dockerfile` — matching tags
- `go build ./...` — clean build
- `cd deploy/setup && go build ./...` — clean build

## Observability Impact

- Signals added/changed: None
- How a future agent inspects this: grep go.mod and Dockerfile for version consistency
- Failure state exposed: None

## Inputs

- `go.mod` — current root version (go 1.25.7)
- `home-device/profiles/go.mod` — current profiles version (go 1.25.7)
- `deploy/setup/go.mod` — current setup version (go 1.24.0)
- `cloud-relay/Dockerfile` — current Docker image (golang:1.24-alpine)
- `home-device/profiles/Dockerfile` — current Docker image (golang:1.24-alpine)

## Expected Output

- `deploy/setup/go.mod` — updated to matching Go version
- `cloud-relay/Dockerfile` — updated golang image tag
- `home-device/profiles/Dockerfile` — updated golang image tag
