# Editorial CSS — Implementation Progress (T1–T8)

**Branch:** `editorial-css`
**Spec:** `docs/superpowers/specs/2026-07-31-editorial-css-design.md`
**Plan:** `docs/superpowers/plans/2026-07-31-editorial-css.md`
**Date:** 2026-07-31

Summary of what each stage did and the files it touched. Verification for CSS
tasks is visual (running app against the `StyleTest` fixture, light + dark);
`type-check` is noted separately because the repo had pre-existing failures.

## T1 — Load web fonts
Replaced the single Rubik Google-Fonts link with **Fraunces** (headings),
**Inter** (body), and **Fira Code** (code).
- `frontend/index.html`

## T2 — Design tokens + base
Created the theme-aware token layer: light values on `:root`, dark on
`html[data-theme="dark"]`, plus a `prefers-color-scheme: dark` default for
users who have made no explicit choice. Pointed the base `body` at the tokens
(background, color, Inter, 18px/1.7) and removed the old `darkslategray`.
- Create `frontend/src/assets/theme.css`
- `frontend/src/assets/main.css` (body)
- `frontend/src/main.ts` (import theme.css first)

## T3 — Wire Tailwind to tokens
Mapped Tailwind color keys to the CSS variables so utility classes are
theme-aware (`bg-surface`, `text-fg`, `text-muted`, `text-accent`,
`border-border`, `bg-surface-2`, `bg-bg`). Remapped the legacy custom names
(`midnightPurple`, `activePageColor`, `deActivePageColor`) onto tokens so
existing component classes became theme-aware with no template change. Added
`serif`/`sans`/`mono` font families.
- `frontend/tailwind.config.js`

## T4 — Markdown content stylesheet (`.content`)
Wrote the full editorial typography for rendered Markdown, scoped to a
`.content` wrapper: Fraunces heading scale (h1 2.75rem → h6 uppercase eyebrow),
Inter body at ~65ch, muted lead paragraph, violet links, list markers,
blockquotes (nested), inline/block code, tables, rule, images. Added the
`.content` class to the `v-html` container.
- Create `frontend/src/assets/content.css`
- `frontend/src/views/contents.vue`
- `frontend/src/main.ts` (import content.css)
- **Bug fixed during review:** Tailwind Preflight strips list markers, so
  `content.css` restores `list-style` explicitly (disc / decimal / circle).
- **Verified:** StyleTest renders with full hierarchy and styled elements in
  both themes.

## T5 — Theme resolution + no-flash init
Added the theme logic module and a no-flash inline script. `resolveInitialTheme`
prefers the stored choice, else system preference; `toggleTheme` flips
`data-theme` and persists via the existing `LocalStorageService`. An inline
script in `<head>` sets `data-theme` before first paint (no flash). This is the
first unit-tested code in the repo (3 Vitest tests, passing).
- Create `frontend/src/global/theme.ts`
- Create `frontend/src/global/theme.spec.ts`
- `frontend/index.html` (inline script)
- **Test-env note:** jsdom's `localStorage` is unavailable on its default
  opaque origin, so the test stubs an in-memory `localStorage`.

## T6 — Theme toggle control
Added a small pill toggle button (☀ / ☾) wired to `getTheme`/`toggleTheme`,
mounted in the side panel.
- Create `frontend/src/components/ThemeToggle.vue`
- `frontend/src/views/sidePanel.vue`
- **Verified:** clicking flips the whole app (content + shell) between themes.

## T7 — Tokenize raw CSS + shell surfaces
Replaced the hardcoded colors in the slide menu (`#382D40`, `white`, `aqua`)
with tokens, and put the border token on the App shell panels.
- `frontend/src/assets/sliderMenu.css`
- `frontend/src/App.vue`

## T8 — Tokenize remaining components
Swapped hardcoded colors for tokens across the remaining components.
- `MenuItem.vue` — hover uses `text-accent`
- `sidePanel.vue` — `hr` uses `var(--border)`
- `upperProfile.vue` — profile name → `text-fg`, title → `text-muted`
  (**plan gap found during execution:** this file was missing from the original
  task list; the name/title were `text-slate-50`, invisible on a light surface)
- `PanelMenu.vue` — `text-fg`, `hover:bg-surface-2`, icon `text-accent`
- `LoginComponent.vue` — gray palette → `bg-surface`/`bg-surface-2`/`text-fg`/
  `border-border`, submit button → `bg-accent text-surface`

## T9 — Full verification + cleanup
- **Menu-strip polish:** the collapsed slide-menu strip now uses `--surface-2`
  plus a soft shadow so it reads as a distinct panel in light mode (it opens to
  `--surface` on hover).
- Deleted the throwaway `frontend/public/__preview.html`.
- **Pre-existing build blocker fixed** (owner approved): `contents.vue` and
  `sideMenu.vue` imported `@/service/service`, which does not exist. The
  production build (rollup) failed to resolve it — the app could not build
  before this work either. Changed both to `@/service/BaseAPI.service`. This
  also resolved the dependent implicit-`any` errors, so the previously-red
  `type-check` is now green.
- `frontend/src/views/contents.vue`, `frontend/src/views/sideMenu.vue`

## Final status
- Content typography and the whole shell are theme-aware; the toggle flips the
  full app between light and dark with the purple identity preserved. Verified
  visually in both themes; theme choice persists with no flash.
- **`npm run type-check` → exit 0. `npm run build` → exit 0** (production bundle
  builds). `npm run test:unit` theme suite → 3 passing.
- The `@/service/service` gotcha in `CLAUDE.md` is now fixed and that note has
  been removed.
- **Commits:** none yet — commits are owner-initiated. Work is on branch
  `editorial-css`.
