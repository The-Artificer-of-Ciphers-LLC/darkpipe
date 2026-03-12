# GSD State

**Active Milestone:** None
**Active Slice:** None
**Phase:** idle
**Requirements Status:** 0 active · 0 validated · 0 deferred · 0 out of scope

## Milestone Registry
- ✅ **M001:** MVP (Phases 1-10) — SHIPPED 2026-02-15
- ✅ **M002:** Post-Launch Hardening
- ✅ **M003:** Container Runtime Compatibility
- ✅ **M004:** DarkPipe Website (darkpipe.org)
- ⚠️ **M005:** Design Validation — External Access & Device Connectivity — PARTIAL (tooling complete, live validation blocked by DNS NXDOMAIN)

## Recent Decisions
- Domain-specific transport map entries instead of wildcard (prevents routing loop)
- Rspamd DKIM signing for all mail server profiles via single shared config
- Outbound relay via cloud relay WireGuard IP 10.8.0.1:25 for all three profiles
- M005 documented as CANNOT FULLY VALIDATE — tooling complete, DNS blocks live verification

## Blockers
- `darkpipe.email` DNS zone has no records (NXDOMAIN) — blocks all external validation

## Next Action
Restore DNS records for darkpipe.email, then re-run validation scripts to complete M005 live verification.
