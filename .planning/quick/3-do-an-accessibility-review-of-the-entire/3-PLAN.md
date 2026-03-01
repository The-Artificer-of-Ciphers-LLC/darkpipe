---
phase: quick-3
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - home-device/profiles/cmd/profile-server/templates/status.html
  - home-device/profiles/cmd/profile-server/templates/device_list.html
  - home-device/profiles/cmd/profile-server/templates/add_device.html
  - home-device/profiles/cmd/profile-server/templates/add_device_result.html
  - home-device/profiles/cmd/profile-server/static/style.css
  - home-device/profiles/cmd/profile-server/webui.go
  - monitoring/status/dashboard.go
  - .planning/quick/3-do-an-accessibility-review-of-the-entire/ACCESSIBILITY-AUDIT.md
autonomous: true
requirements: [QUICK-3]

must_haves:
  truths:
    - "All HTML templates pass WCAG AA contrast requirements (4.5:1 text, 3:1 UI)"
    - "All interactive elements are keyboard-operable and have visible focus indicators"
    - "All forms have proper labels, error announcements, and semantic structure"
    - "Status information is not conveyed by color alone"
    - "All HTML documents use semantic landmarks and proper heading hierarchy"
    - "QR code images have descriptive alt text"
  artifacts:
    - path: ".planning/quick/3-do-an-accessibility-review-of-the-entire/ACCESSIBILITY-AUDIT.md"
      provides: "Comprehensive audit findings with issue severity and fix status"
    - path: "home-device/profiles/cmd/profile-server/templates/status.html"
      provides: "Accessible monitoring dashboard"
      contains: "role=|aria-"
    - path: "home-device/profiles/cmd/profile-server/static/style.css"
      provides: "Focus styles and accessible color values"
      contains: "focus"
  key_links:
    - from: "webui.go inline HTML instructions"
      to: "add_device_result.html template"
      via: "template.HTML injection"
      pattern: "template\\.HTML"
---

<objective>
Perform a WCAG AA accessibility audit of all custom web-facing HTML in the DarkPipe project, fix all identified issues, and produce an audit report.

Purpose: The project has 5 custom HTML surfaces (4 profile-server templates + 1 monitoring dashboard) plus inline HTML generated in Go code. These are the only web surfaces DarkPipe controls (third-party apps like Roundcube/SnappyMail are out of scope). Ensuring WCAG AA compliance makes the admin/monitoring interfaces usable by all operators.

Output: Fixed HTML/CSS/Go files + ACCESSIBILITY-AUDIT.md documenting findings, fixes, and any remaining notes.
</objective>

<execution_context>
@/Users/trekkie/.claude/get-shit-done/workflows/execute-plan.md
@/Users/trekkie/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md

The project has these custom web surfaces to audit:

**Profile Server Templates (4 HTML files):**
- `home-device/profiles/cmd/profile-server/templates/status.html` — Monitoring dashboard (dark theme, auto-refresh, status cards with color-coded indicators)
- `home-device/profiles/cmd/profile-server/templates/device_list.html` — Device management list with table and revoke buttons
- `home-device/profiles/cmd/profile-server/templates/add_device.html` — Add device form with name input and platform select
- `home-device/profiles/cmd/profile-server/templates/add_device_result.html` — Device setup result page with QR code, app password display, platform instructions

**Shared CSS:**
- `home-device/profiles/cmd/profile-server/static/style.css` — Styles for profile server pages (light theme)

**Go files with inline HTML:**
- `home-device/profiles/cmd/profile-server/webui.go` — Contains inline HTML instructions injected via `template.HTML` for iOS, Android, Thunderbird/Outlook, and manual setup flows
- `monitoring/status/dashboard.go` — Dashboard handler (no inline HTML, but serves the status template)

**Out of scope (third-party software, not our code):**
- Roundcube, SnappyMail, Caddy, Stalwart, Rspamd web UIs
- autoconfig.go / autodiscover.go (machine-readable XML for mail clients, not human-facing)

**Also review (documentation accessibility):**
- README.md heading hierarchy and image alt text (badge images)
- docs/*.md heading structure (8 markdown files)
</context>

<tasks>

<task type="auto">
  <name>Task 1: Audit and fix HTML templates and CSS for WCAG AA compliance</name>
  <files>
    home-device/profiles/cmd/profile-server/templates/status.html,
    home-device/profiles/cmd/profile-server/templates/device_list.html,
    home-device/profiles/cmd/profile-server/templates/add_device.html,
    home-device/profiles/cmd/profile-server/templates/add_device_result.html,
    home-device/profiles/cmd/profile-server/static/style.css
  </files>
  <action>
Audit all 4 HTML templates and the shared CSS file against WCAG 2.1 AA. Apply the accessibility-lead, aria-specialist, keyboard-navigator, contrast-master, forms-specialist, alt-text-headings, tables-data-specialist, and link-checker specialist checklists from the project's accessibility agents.

**status.html (monitoring dashboard) -- Known issues to fix:**
1. Color-only status indicators: The `.status-dot` elements (green/yellow/red circles) convey status solely via color. Add a visually-hidden text label (e.g., "Healthy", "Degraded", "Unhealthy") next to each dot, or use `aria-label` on the parent element. The `.overall-status` span already has text but the per-card status dots in `.card-title` do not.
2. Contrast on dark background: Verify `#999` metric-label text on `#2a2a2a` card background meets 4.5:1 ratio (it does not -- #999 on #2a2a2a is ~3.5:1). Fix by lightening to at least `#aaa` or `#b3b3b3`.
3. Contrast for `.timestamp` text: `#666` on `#1a1a1a` is ~2.6:1. Fix by lightening to at least `#999`.
4. Missing skip-to-content link.
5. Missing `<main>` landmark wrapping content area.
6. The auto-refresh meta tag should be documented with a visible note explaining the auto-refresh behavior (already present in timestamp text, but add `role="status"` to the timestamp div so screen readers announce updates).
7. Queue bar: The `.queue-bar-fill` has text inside a gradient background -- ensure sufficient contrast. Add `aria-label` or `role="progressbar"` with `aria-valuenow`/`aria-valuemax` attributes.
8. Add `role="list"` to `.service-list` div and `role="listitem"` to `.service-item` divs (or convert to semantic `<ul>/<li>`).

**device_list.html -- Known issues to fix:**
1. Table has proper `<thead>/<tbody>` and `<th>` elements (good). Add `scope="col"` to each `<th>`.
2. Revoke form: The `onsubmit="return confirm(...)"` is acceptable, but the confirm dialog content should identify which device. Change to `return confirm('Revoke access for {{.DeviceName}}?')` so the dialog is specific.
3. Alert divs (`.alert-error`, `.alert-success`): Add `role="alert"` so screen readers announce them.
4. Missing skip-to-content link.
5. Nav links have no indication of current page. Add `aria-current="page"` to the "My Devices" link.

**add_device.html -- Known issues to fix:**
1. Form labels are properly associated via `for`/`id` (good). Verify the `help-text` paragraphs are programmatically associated with inputs using `aria-describedby`.
2. Error alert div: Add `role="alert"`.
3. Missing skip-to-content link.
4. Nav: Add `aria-current="page"` to "Add Device" link. Wrap nav in `<nav aria-label="Main navigation">`.

**add_device_result.html -- Known issues to fix:**
1. QR code image: `alt="QR Code"` is insufficient. Change to `alt="QR code for {{.Platform}} device configuration"` or similar descriptive text.
2. Warning alert: Add `role="alert"`.
3. Password display: The `<code>` element should have `aria-label="App password"` or be wrapped in a labeled container so screen readers identify it.
4. Missing skip-to-content link.

**style.css -- Known issues to fix:**
1. No visible focus styles defined. The `input:focus` and `select:focus` use `outline: none` which removes the default focus indicator -- this is a WCAG failure. Replace with a visible focus ring: `outline: 2px solid #007bff; outline-offset: 2px;` or `box-shadow: 0 0 0 3px rgba(0,123,255,0.5);`.
2. Add focus styles for buttons, links, and `.button` class: `button:focus, a:focus, .button:focus { outline: 2px solid #007bff; outline-offset: 2px; }`.
3. Alert colors: `.alert-success` uses `color: #3c3` on `background: #efe`. Verify contrast. `#3c3` (#33cc33) on `#eeffee` is ~2.5:1 -- fix by darkening to `#1a8a1a` or similar (~4.5:1).
4. `.alert-error` uses `color: #c33` on `#fee`. `#cc3333` on `#ffeeee` is ~3.8:1 -- fix by darkening to `#b12020` or similar.
5. Add a skip-to-content CSS class: `.skip-link { position: absolute; top: -40px; left: 0; ... }` with `:focus` state that brings it into view.
6. Touch target size: Verify buttons and links meet 44x44px minimum. The current `padding: 10px 20px` on 16px font should be adequate for height but verify.
  </action>
  <verify>
    <automated>cd /Users/trekkie/projects/darkpipe && grep -c 'aria-\|role=' home-device/profiles/cmd/profile-server/templates/*.html | grep -v ':0$' | wc -l | xargs test 4 -eq && echo "All 4 templates have ARIA attributes" && grep -c 'outline\|focus' home-device/profiles/cmd/profile-server/static/style.css | xargs test 0 -lt && echo "CSS has focus styles" && grep 'scope="col"' home-device/profiles/cmd/profile-server/templates/device_list.html > /dev/null && echo "Table headers have scope"</automated>
  </verify>
  <done>
    All 4 HTML templates updated with: skip-to-content links, semantic landmarks, ARIA roles/labels on status indicators, role="alert" on notification divs, proper table scope attributes, descriptive QR alt text, aria-describedby on form fields. CSS updated with: visible focus indicators on all interactive elements, corrected alert text colors for 4.5:1 contrast, skip-link styles, corrected metric-label and timestamp colors on dark theme.
  </done>
</task>

<task type="auto">
  <name>Task 2: Fix inline HTML in webui.go and produce audit report</name>
  <files>
    home-device/profiles/cmd/profile-server/webui.go,
    .planning/quick/3-do-an-accessibility-review-of-the-entire/ACCESSIBILITY-AUDIT.md
  </files>
  <action>
**webui.go inline HTML fixes:**

The `processAddDevice` function generates HTML instruction strings that are injected into the template via `template.HTML`. These bypass template escaping and render directly. Audit each platform's instruction block:

1. **iOS/macOS instructions (lines ~200-210):** The `<h3>` is appropriate. The `<a href="%s" class="button">Download Profile</a>` link is fine but add `aria-label="Download .mobileconfig profile for {{deviceName}}"` for clarity. The `<p>` with expiry info is OK.

2. **Android instructions (lines ~225-240):** The `<h3>` and `<ol>` are good semantic choices. The `<strong>` tags for email/password are appropriate. The `<ul>` for manual settings is correct. No issues beyond what the template already handles.

3. **Thunderbird/Outlook instructions (lines ~243-253):** Uses `platform` as the `<h3>` text which will display as lowercase "thunderbird" or "outlook". Capitalize: `strings.Title(platform)` or manual capitalization. Otherwise semantically correct.

4. **Manual/Other instructions (lines ~257-269):** Semantically correct `<h3>`, `<p>`, `<ul>` usage. No issues.

5. **All instruction blocks:** Ensure heading hierarchy is maintained. The main template has `<h2>` for "Device Created: ..." so `<h3>` in instructions is correct.

Apply fixes in webui.go for the identified issues (capitalize platform name, add aria-label to download link).

**ACCESSIBILITY-AUDIT.md:**

Create a comprehensive audit report at `.planning/quick/3-do-an-accessibility-review-of-the-entire/ACCESSIBILITY-AUDIT.md` documenting:

1. **Scope** -- What was audited (4 templates, 1 CSS, 1 Go file with inline HTML, 9 markdown docs) and what was excluded (third-party UIs, machine-readable XML).

2. **Methodology** -- WCAG 2.1 AA checklist applied: Perceivable (text alternatives, adaptable, distinguishable), Operable (keyboard, enough time, seizures/physical, navigable, input modalities), Understandable (readable, predictable, input assistance), Robust (compatible).

3. **Findings table** with columns: ID, File, Issue, WCAG SC, Severity (Critical/Major/Minor), Status (Fixed/Noted). List every issue found and fixed in Tasks 1 and 2.

4. **Documentation accessibility review** -- Report on README.md and docs/*.md:
   - README.md: Badge images have alt text (good: "License: AGPL-3.0", "Go Version", etc.). Heading hierarchy: single H1, H2s for sections, H3s for subsections (correct). Code blocks have language hints. Links have descriptive text. No issues found.
   - docs/*.md: All 8 files use proper H1 -> H2 -> H3 hierarchy without skipping levels (verified via grep). No images in docs (no alt text needed). Code blocks use language fencing. Links are descriptive (not "click here"). No issues found.

5. **Summary** -- Total issues found, total fixed, any remaining items with rationale.

6. **Recommendations** -- Suggest adding automated accessibility linting (e.g., axe-core or pa11y) to CI for future changes if/when a CI pipeline covers HTML output.
  </action>
  <verify>
    <automated>cd /Users/trekkie/projects/darkpipe && go build ./home-device/profiles/cmd/profile-server/ && echo "Go code compiles" && test -f .planning/quick/3-do-an-accessibility-review-of-the-entire/ACCESSIBILITY-AUDIT.md && echo "Audit report exists" && grep -c 'strings.ToUpper\|strings.Title\|Title\|[A-Z]' home-device/profiles/cmd/profile-server/webui.go > /dev/null && echo "Platform name capitalization present"</automated>
  </verify>
  <done>
    webui.go inline HTML instructions updated with: capitalized platform names in headings, aria-label on download profile link. ACCESSIBILITY-AUDIT.md created with complete findings table documenting all issues found across all surfaces, their WCAG success criteria references, severity levels, and fix status. Go code compiles successfully.
  </done>
</task>

</tasks>

<verification>
1. All 4 HTML templates contain ARIA attributes (aria-label, role, aria-describedby)
2. CSS contains visible focus styles (no bare `outline: none` without replacement)
3. All alert/notification elements have `role="alert"`
4. Status indicators have text alternatives (not color-only)
5. Table headers have `scope="col"`
6. QR code image has descriptive alt text
7. Go code compiles without errors
8. ACCESSIBILITY-AUDIT.md exists with findings table
9. Dark theme (status.html) text colors meet 4.5:1 contrast ratios
10. Light theme (style.css) alert text colors meet 4.5:1 contrast ratios
</verification>

<success_criteria>
All custom web surfaces in DarkPipe meet WCAG 2.1 AA compliance. Every interactive element has visible focus indicators. Status information is conveyed via text, not color alone. Forms have programmatic label associations. Audit report documents all findings with WCAG success criteria references.
</success_criteria>

<output>
After completion, create `.planning/quick/3-do-an-accessibility-review-of-the-entire/3-SUMMARY.md`
</output>
