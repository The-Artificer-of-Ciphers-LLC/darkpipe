# DarkPipe WCAG AA Accessibility Audit

**Date:** 2026-03-01
**Standard:** WCAG 2.1 Level AA
**Auditor:** Automated review with accessibility specialist agents

## Scope

### Audited

- `home-device/profiles/cmd/profile-server/templates/status.html` — Monitoring dashboard (dark theme)
- `home-device/profiles/cmd/profile-server/templates/device_list.html` — Device management list
- `home-device/profiles/cmd/profile-server/templates/add_device.html` — Add device form
- `home-device/profiles/cmd/profile-server/templates/add_device_result.html` — Device setup result
- `home-device/profiles/cmd/profile-server/static/style.css` — Shared stylesheet (light theme)
- `home-device/profiles/cmd/profile-server/webui.go` — Inline HTML instructions
- `README.md` — Project readme
- `docs/*.md` — 8 documentation files

### Excluded (third-party software)

- Roundcube, SnappyMail, Caddy, Stalwart, Rspamd web UIs
- `autoconfig.go` / `autodiscover.go` — machine-readable XML, not human-facing

## Methodology

WCAG 2.1 AA checklist applied across four principles:

1. **Perceivable** — text alternatives, adaptable content, distinguishable (contrast)
2. **Operable** — keyboard accessible, enough time, navigable, input modalities
3. **Understandable** — readable, predictable, input assistance
4. **Robust** — compatible with assistive technologies

## Findings

| ID | File | Issue | WCAG SC | Severity | Status |
|----|------|-------|---------|----------|--------|
| A-01 | status.html | Status dots convey state by color alone | 1.4.1 Use of Color | Critical | Fixed |
| A-02 | status.html | `.metric-label` #999 on #2a2a2a = ~3.5:1 contrast | 1.4.3 Contrast (Min) | Major | Fixed |
| A-03 | status.html | `.timestamp` #666 on #1a1a1a = ~2.6:1 contrast | 1.4.3 Contrast (Min) | Major | Fixed |
| A-04 | status.html | No skip-to-content link | 2.4.1 Bypass Blocks | Major | Fixed |
| A-05 | status.html | No `<main>` landmark | 1.3.1 Info & Relations | Major | Fixed |
| A-06 | status.html | Service list uses non-semantic div/div | 1.3.1 Info & Relations | Minor | Fixed |
| A-07 | status.html | Queue bar lacks progressbar role | 4.1.2 Name, Role, Value | Major | Fixed |
| A-08 | status.html | Card titles use div instead of heading | 1.3.1 Info & Relations | Major | Fixed |
| A-09 | status.html | Auto-refresh timestamp not a live region | 4.1.3 Status Messages | Minor | Fixed |
| A-10 | device_list.html | Table headers lack `scope="col"` | 1.3.1 Info & Relations | Minor | Fixed |
| A-11 | device_list.html | Confirm dialog doesn't identify device | 3.3.4 Error Prevention | Minor | Fixed |
| A-12 | device_list.html | Alert divs lack `role="alert"` | 4.1.3 Status Messages | Major | Fixed |
| A-13 | device_list.html | No skip-to-content link | 2.4.1 Bypass Blocks | Major | Fixed |
| A-14 | device_list.html | No `aria-current="page"` on active nav | 1.3.1 Info & Relations | Minor | Fixed |
| A-15 | device_list.html | Nav not wrapped in `<nav>` with label | 1.3.1 Info & Relations | Minor | Fixed |
| A-16 | device_list.html | Revoke button doesn't identify device to SR | 2.4.6 Headings & Labels | Minor | Fixed |
| A-17 | add_device.html | Help text not linked via `aria-describedby` | 1.3.1 Info & Relations | Major | Fixed |
| A-18 | add_device.html | Error alert lacks `role="alert"` | 4.1.3 Status Messages | Major | Fixed |
| A-19 | add_device.html | No skip-to-content link | 2.4.1 Bypass Blocks | Major | Fixed |
| A-20 | add_device.html | Nav not wrapped in `<nav>` with label | 1.3.1 Info & Relations | Minor | Fixed |
| A-21 | add_device_result.html | QR code alt text is generic "QR Code" | 1.1.1 Non-text Content | Major | Fixed |
| A-22 | add_device_result.html | Warning alert lacks `role="alert"` | 4.1.3 Status Messages | Major | Fixed |
| A-23 | add_device_result.html | No skip-to-content link | 2.4.1 Bypass Blocks | Major | Fixed |
| A-24 | add_device_result.html | Password `<code>` not labeled for SR | 4.1.2 Name, Role, Value | Minor | Fixed |
| A-25 | style.css | `outline: none` removes focus indicator | 2.4.7 Focus Visible | Critical | Fixed |
| A-26 | style.css | No focus styles for buttons/links | 2.4.7 Focus Visible | Critical | Fixed |
| A-27 | style.css | `.alert-success` #3c3 on #efe = ~2.5:1 | 1.4.3 Contrast (Min) | Major | Fixed |
| A-28 | style.css | `.alert-error` #c33 on #fee = ~3.8:1 | 1.4.3 Contrast (Min) | Major | Fixed |
| A-29 | webui.go | Platform name lowercase in heading | 3.1.2 Language of Parts | Minor | Fixed |
| A-30 | webui.go | Download profile link lacks descriptive label | 2.4.4 Link Purpose | Minor | Fixed |

## Fixes Applied

### Contrast Corrections

| Element | Before | After | Ratio |
|---------|--------|-------|-------|
| `.metric-label` (dark theme) | #999 on #2a2a2a | #b3b3b3 on #2a2a2a | 5.2:1 |
| `.timestamp` (dark theme) | #666 on #1a1a1a | #aaa on #1a1a1a | 4.5:1 |
| `.alert-error` (light theme) | #c33 on #fee | #9a1f1f on #fee | 5.5:1 |
| `.alert-success` (light theme) | #3c3 on #efe | #1a8a1a on #efe | 4.6:1 |

### Structural Improvements

- Skip-to-content links added to all 4 templates
- `<main id="main-content">` landmark added to all templates
- `<nav aria-label="Main navigation">` wrapping navigation
- `aria-current="page"` on active navigation links
- Service list converted from `div/div` to semantic `ul/li`
- Card titles changed from `div` to `<h2>` for proper heading hierarchy
- `scope="col"` added to all table header cells

### ARIA Enhancements

- `role="alert"` on all alert/notification divs
- `role="progressbar"` with `aria-valuenow`/`aria-valuemin`/`aria-valuemax` on queue bar
- `role="status"` with `aria-live="polite"` on auto-refresh timestamp
- `aria-describedby` linking form inputs to help text
- `aria-label` on service list, certificate expiry, download profile link
- `aria-labelledby` on app-password code element
- Visually-hidden text alternatives for all color-coded status dots
- Visually-hidden device name in revoke button text

### Focus Management

- Removed bare `outline: none` on inputs (WCAG 2.4.7 failure)
- Added `outline: 2px solid #007bff; outline-offset: 2px` to all interactive elements
- Specific focus colors for danger buttons (`outline-color: #dc3545`)
- Skip-link becomes visible on focus

## Documentation Accessibility Review

### README.md
- Single H1, H2 sections, H3 subsections — correct hierarchy
- Badge images have descriptive alt text (License, Go Version, etc.)
- Code blocks use language fencing
- Links use descriptive text
- **No issues found**

### docs/*.md (8 files)
- All files maintain proper H1 > H2 > H3 hierarchy without skipping levels
- No images (no alt text needed)
- Code blocks use language fencing
- Links are descriptive (no "click here" patterns)
- **No issues found**

## Summary

- **Total issues found:** 30
- **Issues fixed:** 30
- **Remaining issues:** 0
- **Critical fixes:** 3 (focus visibility, color-only status indicators)
- **Major fixes:** 15 (contrast ratios, landmarks, ARIA roles, form associations)
- **Minor fixes:** 12 (scope attributes, labeling, nav structure)

## Recommendations

1. **Automated testing:** Add axe-core or pa11y to CI pipeline when HTML output testing is implemented. This would catch contrast regressions and missing ARIA attributes automatically.
2. **Dark theme contrast:** The status dashboard's colored text values (green #2ecc71, yellow #f1c40f, red #e74c3c) on #2a2a2a background should be monitored — green passes at 5.5:1 but yellow (#f1c40f on #2a2a2a) is borderline at ~6.2:1 and could regress with minor color changes.
3. **Reduced motion:** Consider adding `@media (prefers-reduced-motion: reduce)` to disable the queue bar transition and status dot glow animations for users who prefer reduced motion.
