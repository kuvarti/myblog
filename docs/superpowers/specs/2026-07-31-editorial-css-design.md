# Editorial CSS Design — myblog

**Date:** 2026-07-31
**Status:** Approved (design), pending implementation plan

## Goal

Establish the definitive, theme-aware global stylesheet for the blog. Every page is
authored in Markdown, stored in the DB, and rendered to HTML on demand (see
`CLAUDE.md` → Purpose & design intent and Page rendering). This stylesheet must make
any rendered Markdown page look good on its own — with no per-page CSS — and give the
existing shell a coherent light/dark identity. Getting this right is a deliberate,
one-time investment; future pages inherit it.

The starting point is an unstyled prototype: rendered Markdown has almost no
typographic styling (no heading hierarchy, no list markers, no blockquote/code/table
treatment, cramped spacing). This design fixes that.

## Design direction (decisions)

| Decision | Choice |
|----------|--------|
| Aesthetic | Editorial / minimal — reading-first, generous whitespace, strong hierarchy, restrained color |
| Palette | Theme-aware (light + dark); purple identity kept in **both** themes, not just as an accent |
| Typography | Fraunces (headings, serif display) · Inter (body, sans) · Fira Code (code, mono) |
| Approach | Token-driven hand-written CSS for Markdown content; Tailwind retained for components |

## Scope

**In scope**
- Full editorial typography for rendered Markdown content.
- A design-token system (CSS custom properties) for light + dark, driving both content
  and shell.
- Restyling the **existing** components so the whole app is theme-aware and coherent:
  `App.vue`, `views/sideMenu.vue`, `views/sidePanel.vue`, `views/contents.vue`,
  `components/menuComponents/MenuItem.vue`, `components/sidePanelComponents/PanelMenu.vue`,
  and the existing PrimeVue usages. This is a **visual** rework (swap hardcoded colors
  for tokens, apply the design system) — not new behavior.
- A theme toggle (light/dark) with persistence.
- Loading the three web fonts.

**Out of scope**
- New admin **panel functionality** (login-gated create/edit/delete of pages). The panel
  components are restyled visually only; their behavior is a separate later project.
- Any change to the backend render pipeline or data model.

## Architecture / Approach (B)

Rendered Markdown arrives via `v-html` as raw HTML (plain `<h1>`, `<p>`, `<table>`…),
so it cannot carry Tailwind utility classes. Styling therefore targets a wrapper class
with descendant CSS. Tailwind stays for everything else.

1. **Design tokens** — CSS custom properties defined once (e.g. `assets/theme.css`),
   light values on `:root`, dark values on `html[data-theme="dark"]`, with a
   `prefers-color-scheme: dark` default before any explicit choice.
2. **Tailwind integration** — `tailwind.config.js` color keys reference the CSS
   variables (e.g. `surface: 'var(--surface)'`, `fg: 'var(--fg)'`). Component rework then
   swaps hardcoded classes (`bg-midnightPurple`) for token classes (`bg-surface`), and
   themes flip automatically.
3. **Markdown content CSS** — a hand-written stylesheet scoped to a `.content` wrapper on
   the `v-html` container in `contents.vue`, implementing the full element set below.
4. **Fonts** — Fraunces + Inter + Fira Code loaded via Google Fonts in `index.html`
   (replacing the current Rubik link).
5. **Theme toggle** — a small control in the shell that flips `html[data-theme]`, persisted
   with the existing `LocalStorageService` (no new infrastructure). Initial theme:
   stored value if present, else `prefers-color-scheme`.

The values below were validated visually against the `StyleTest` fixture (light + dark)
before approval.

## Design tokens (light + dark)

| Token | Light (lavender paper) | Dark (purple charcoal) | Use |
|-------|------------------------|------------------------|-----|
| `--bg` | `#F6F4FB` | `#1C1726` | page background |
| `--surface` | `#FDFCFF` | `#2A2338` | panel / card |
| `--surface-2` | `#EFEBF8` | `#211B2E` | code block, secondary surface |
| `--fg` | `#262036` | `#ECE7F4` | primary text |
| `--muted` | `#6C6480` | `#A79FB8` | secondary text, captions |
| `--accent` | `#6A46C9` | `#B3A4F5` | links, emphasis, active state |
| `--border` | `#E4DCF2` | `#3A3350` | dividers, table/code borders |

Font family tokens: `--serif: 'Fraunces', Georgia, serif` · `--sans: 'Inter', system-ui,
sans-serif` · `--mono: 'Fira Code', monospace`.

Both themes stay within the purple family (light is violet-tinted off-white with
purple-charcoal text; dark evolves the current `#382D40`). Reading contrast is high in
both.

## Typography — type scale

Modular scale ≈ 1.25 (major third), body base 18px.

| Element | Font | Size | Weight | Line-height | Notes |
|---------|------|------|--------|-------------|-------|
| h1 | serif | 2.75rem | 600 | 1.1 | letter-spacing −0.01em |
| h2 | serif | 2rem | 600 | 1.15 | |
| h3 | serif | 1.5rem | 600 | 1.25 | |
| h4 | serif | 1.25rem | 600 | 1.3 | |
| h5 | sans | 1.0625rem | 700 | 1.4 | |
| h6 | sans | 0.875rem | 600 | 1.4 | UPPERCASE, tracking 0.06em, `--muted` (eyebrow) |
| body (`p`,`li`) | sans | 1.125rem | 400 | 1.7 | |
| lead (first `p`) | sans | 1.25rem | 400 | 1.6 | `--muted` |
| small / caption | sans | 0.875rem | 400 | 1.5 | `--muted` |
| code (inline + block) | mono | 0.9em | 400 | 1.6 | |

Reading measure: content max-width ≈ 65ch.

## Element specifications

- **Headings:** serif h1–h4, sans h5, sans uppercase-eyebrow h6. Large top margin to
  separate sections, small bottom margin to bind to following text.
- **Paragraph:** margin `0 0 1.4em`. First paragraph of a page renders as a muted lead.
- **Links:** `--accent`, underline with `2px` offset and `1px` thickness; thicken to `2px`
  on hover.
- **Inline emphasis:** `strong` 600; `em` italic; `del` struck through in `--muted`.
- **Lists:** `padding-left 1.4em`, item margin `0.4em`, marker color `--accent`; nested
  lists tighten vertical margin.
- **Blockquote:** `3px` `--accent` left border, left padding, italic, `--muted`; nested
  blockquote uses a `--border`-colored border.
- **Inline code:** `--surface-2` background, small padding, `5px` radius, mono `0.9em`.
- **Code block:** `--surface-2` background, `1px --border`, `10px` radius, padding
  `~1.1em 1.3em`, `overflow-x: auto`, line-height 1.6; inner `code` has no background.
- **Table:** full width within measure, `border-collapse`, cell padding `~0.6em 0.8em`,
  `1px --border` row separators, `2px` under the header row, bold header cells.
- **Horizontal rule:** borderless except `1px --border` top, margin `~2.6em 0`.
- **Image:** `max-width 100%`, `10px` radius, block, vertical margin.
- **Content container:** the `.content` sits in a rounded `--surface` panel with generous
  padding, centered, matching the app's existing content-area framing.

## Spacing & layout rhythm

Vertical rhythm derived from the 1.7 body line-height: paragraphs and blocks separated by
`~1.4em`; section headings get larger top margins (`h2 ~2.4em`, `h3 ~1.9em`) to create
clear visual grouping. Content is centered and capped at the ~65ch measure so long-form
text never sprawls.

## Theme system

- `html[data-theme]` attribute drives the active theme; default before any choice follows
  `prefers-color-scheme`.
- Toggle control lives in the shell; on click it flips the attribute and saves the choice
  via `LocalStorageService` (key e.g. `theme`).
- On app load, read the stored value (if any) and apply it before first paint to avoid a
  flash.

## Component rework (visual only)

Swap hardcoded colors/utility classes for token-based ones and apply the design system to:
`App.vue`, `views/sideMenu.vue`, `views/sidePanel.vue`, `views/contents.vue`,
`components/menuComponents/MenuItem.vue`, `components/sidePanelComponents/PanelMenu.vue`,
`components/panelComponents/LoginComponent.vue`, and existing PrimeVue usages. No behavior
changes.

## Constraints & gotchas

- Rendered Markdown is raw `v-html` — style only via the `.content` wrapper, never by
  adding classes to individual rendered elements.
- The backend's page-authoring format treats any source line containing `<` as raw-HTML
  passthrough; embedded HTML in a page is emitted verbatim and inherits the same `.content`
  styles. (Authoring detail, not a CSS concern, but relevant when testing.)
- Fira Code was referenced in `tailwind.config.js` but never actually loaded; this design
  loads it for real.

## Success criteria

- The `StyleTest` kitchen-sink page renders with clear hierarchy and fully styled elements
  in both light and dark.
- Switching themes flips the entire app (content + shell) coherently, with the purple
  identity preserved in both.
- No per-page CSS is required for any Markdown page.

## Test fixture

`StyleTest` (seeded in the `Pages` collection; source at
`database/SimpleBackup/StyleTest.json`) is the canonical fixture — a single page
exercising every Markdown element. Tune and verify the CSS against it.
