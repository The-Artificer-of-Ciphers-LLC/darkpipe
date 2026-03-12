# S04: CI & Regression Validation — Research

**Date:** 2026-03-11

## Summary

S04 adds a Podman build and compose validation job to GitHub Actions, ensures existing Docker CI continues passing (zero regression), and integrates the runtime compatibility check script into CI. The work is low-risk because Podman comes pre-installed on `ubuntu-latest` (Ubuntu 24.04) runners, the existing verification scripts (`verify-podman-compat.sh`, `check-runtime.sh`) are already designed for CI (exit codes, structured output), and the compose files were already validated as dual-compatible in S01.

The primary challenge is that S01's verification was done without Docker or Podman installed (all compose config checks were skipped). This slice is where those runtime-dependent checks actually execute for the first time in a real environment. Any latent compose incompatibilities will surface here.

The approach should be a single new workflow file (e.g., `validate-containers.yml`) with two jobs: one using Docker (existing behavior + regression check) and one using Podman (new). Both jobs run the verification scripts and `compose config` validation. This avoids modifying existing build/release workflows, keeping regression risk near zero.

## Recommendation

Create a new workflow `.github/workflows/validate-containers.yml` triggered on push to main and PRs. Two jobs:

1. **`docker-validate`** — Runs `docker compose config` on all compose files (base, podman override layering, selinux override layering), runs `verify-podman-compat.sh` (which uses `docker compose config` internally), runs `verify-container-security.sh`, and runs `check-runtime.sh`.

2. **`podman-validate`** — Installs `podman-compose` via pip (`webgtx/setup-podman-compose@v1` or `pip install podman-compose`), runs `podman version` to log the version, runs `podman build` on all 5 Dockerfiles, runs `podman-compose config` on base compose files, and runs `check-runtime.sh` (which will auto-detect Podman).

Keep the existing `build-custom.yml`, `build-prebuilt.yml`, and `release.yml` workflows untouched — they handle Docker image building and pushing. The new workflow is purely validation/linting.

## Don't Hand-Roll

| Problem | Existing Solution | Why Use It |
|---------|------------------|------------|
| Podman-compose setup in CI | `webgtx/setup-podman-compose@v1` action or `pip install podman-compose` | Avoids manual install steps; podman-compose is a Python package |
| Compose file validation | `scripts/verify-podman-compat.sh` | Already validates health checks, version fields, swarm directives, override layering — purpose-built for this |
| Container security audit | `scripts/verify-container-security.sh` | Already checks cap_drop, no-new-privileges, read_only, HEALTHCHECK |
| Runtime prereq checks | `scripts/check-runtime.sh` | Detects runtime, validates version, checks compose tool, SELinux state |
| Podman login to GHCR | `redhat-actions/podman-login@v1` | Official Red Hat action for registry auth (if pushing from Podman job — may not be needed for validation-only) |

## Existing Code and Patterns

- `.github/workflows/build-custom.yml` — Docker-based build using `docker/build-push-action@v6` with QEMU for multi-arch. Don't modify; the new workflow is additive.
- `.github/workflows/build-prebuilt.yml` — Matrix build for default+conservative stacks. Same — don't modify.
- `.github/workflows/release.yml` — Tag-triggered release with Go binary builds (go 1.25). Don't modify.
- `scripts/verify-podman-compat.sh` — 7 check categories, PASS/FAIL/SKIP output, exit 1 on failures. Ready for CI as-is. Note: gracefully skips `podman-compose` checks if not installed.
- `scripts/verify-container-security.sh` — Checks all compose services and Dockerfiles. Ready for CI.
- `scripts/check-runtime.sh` — Detects Docker/Podman/Apple Containers, validates versions, checks compose tool. Has `--quiet` flag for CI-friendly minimal output.
- `cloud-relay/docker-compose.podman.yml` — Podman override with `x-podman` compat flags.
- `home-device/docker-compose.podman.yml` — Same pattern for home device.
- `cloud-relay/docker-compose.podman-selinux.yml` — SELinux `:z` label override (CI runner won't have SELinux enforcing, but layering should still validate).
- `home-device/docker-compose.podman-selinux.yml` — Same for home device.

## Constraints

- **Podman pre-installed on ubuntu-latest** — Ubuntu 24.04 runners come with Podman pre-installed, but the version may or may not meet the 5.3.0+ minimum. Need a version check step and potential upgrade.
- **podman-compose is NOT pre-installed** — Must be installed via pip or the `webgtx/setup-podman-compose@v1` action.
- **No runtime container testing** — S04 is compose validation and build checks only (per decision: "Podman compatibility verified via contract-level compose config validation, not runtime testing"). Do NOT try to `podman-compose up` in CI.
- **Go 1.25 required** — If adding Go vet/build steps, need `actions/setup-go@v5` with `go-version: '1.25'`.
- **5 Dockerfiles to build** — `cloud-relay/Dockerfile`, `home-device/maddy/Dockerfile`, `home-device/postfix-dovecot/Dockerfile`, `home-device/profiles/Dockerfile`, `home-device/stalwart/Dockerfile`.
- **Existing workflows must not change** — Zero regression means the three existing workflow files stay untouched.
- **SELinux not enforcing on Ubuntu runners** — SELinux checks in `check-runtime.sh` will show "not installed" / skip, which is fine.

## Common Pitfalls

- **Podman version too old on ubuntu-latest** — Ubuntu 24.04 may ship Podman 4.x, which is below the 5.3.0+ requirement. Fix: check version first, and if needed, install from Podman's official PPA or use `podmanlabs/setup-podman` to get a recent version. Alternatively, accept the runner's version for build validation (Podman 4.x can still `podman build`) and note the version gap.
- **podman-compose config behaves differently than docker compose config** — `podman-compose` is less mature; `config` subcommand may not support `--quiet` or may parse differently. The verify script already handles this gracefully (skip if not available, test if available).
- **Trying to run containers in CI** — The decision explicitly says "contract-level compose config validation, not runtime testing." Don't attempt `podman-compose up` — it requires networking, volume mounts, and possibly privileged access that CI runners don't have for the full DarkPipe stack.
- **Modifying existing workflows** — Any change to build-custom.yml, build-prebuilt.yml, or release.yml risks breaking existing Docker CI. Keep the new validation in a separate workflow file.
- **Port 25 check failing in CI** — `check-runtime.sh` checks port 25 availability. On CI runners, port 25 might be blocked or in use. If the script fails on this check, it will cause a spurious CI failure. May need to skip this check in CI context or add a `--skip-network` flag.

## Open Risks

- **Podman version on ubuntu-latest may be below 5.3.0** — If Ubuntu 24.04 ships Podman 4.9.x, the `check-runtime.sh` version check will FAIL. Options: (a) add a step to upgrade Podman from the kubic repo, (b) accept Podman 4.x for build validation and adjust the script, or (c) use a Fedora container for Podman-native testing. Most pragmatic: install from upstream repo if version is too low.
- **podman build may fail on multi-stage Dockerfiles with Go** — `podman build` should handle standard multi-stage builds, but edge cases with build contexts or `COPY --from` may differ. Low probability since Dockerfiles are OCI-standard.
- **check-runtime.sh port 25 check in CI** — Runner may have port 25 blocked by cloud provider firewall, which could cause a false FAIL. Need to either skip network checks in CI or make the check non-fatal.
- **CI timing / flakiness** — Podman image builds are typically slightly slower than Docker on the same runner due to different storage driver defaults. Should not be a problem for validation-only workflow but worth monitoring.

## Skills Discovered

| Technology | Skill | Status |
|------------|-------|--------|
| GitHub Actions | `wshobson/agents@github-actions-templates` | available (4.9K installs) — already have `github-actions-templates` installed locally |
| Podman | `knoopx/pi@podman` | available (56 installs) — low adoption, skip |
| Podman CI | `bagelhole/devops-security-agent-skills@podman` | available (15 installs) — low adoption, skip |

No additional skill installs recommended. The locally installed `github-actions-templates` skill covers the CI workflow patterns needed.

## Sources

- Podman is pre-installed on ubuntu-latest GitHub runners (source: [oneuptime.com CI guide](https://oneuptime.com/blog/post/2026-01-27-podman-cicd/view))
- `webgtx/setup-podman-compose@v1` action installs podman-compose in CI (source: [GitHub Marketplace](https://github.com/marketplace/actions/setup-podman-compose))
- `redhat-actions/podman-login@v1` for GHCR auth with Podman (source: [byteshiva article](https://byteshiva.medium.com/automating-podman-and-ghcr-with-github-actions-0dc800db68ba))
- ubuntu-latest is Ubuntu 24.04 as of Feb 2026 (source: [Stack Overflow / runner-images](https://stackoverflow.com/questions/69840694/what-does-ubuntu-latest-mean-for-github-actions))
- macOS 26 runner available as `macos-26` (beta) — relevant for potential Apple Containers CI in future (source: [GitHub runner-images table](https://github.com/actions/runner-images))
- `podman build` with `--format docker` ensures Docker-compatible image format (source: [oneuptime.com CI guide](https://oneuptime.com/blog/post/2026-01-27-podman-cicd/view))
