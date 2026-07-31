# Editorial CSS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the definitive theme-aware editorial stylesheet: design tokens (light + dark), Markdown content typography, fonts, a theme toggle, and a token-based restyle of the existing components.

**Architecture:** CSS custom properties drive all color (light on `:root`, dark on `html[data-theme="dark"]`, `prefers-color-scheme` default). Tailwind is kept; its color keys point at the CSS variables so utility classes are theme-aware. Rendered Markdown (raw `v-html`) is styled via a `.content` wrapper. A tiny theme module toggles `data-theme` and persists it with the existing `LocalStorageService`.

**Tech Stack:** Vue 3 + Vite + TypeScript, Tailwind CSS v3, PrimeVue, Vitest. Fonts: Fraunces (headings), Inter (body), Fira Code (code).

**Spec:** `docs/superpowers/specs/2026-07-31-editorial-css-design.md`

## Global Constraints

- **Commits are owner-initiated** (`memory/git/commit-flow`). Do NOT auto-commit. Each task ends with a "Commit point" — pause and ask the owner whether to commit; only commit on an explicit "commit" instruction, then follow the flow (ask theme, write English message).
- **All docs and comments in English** (`memory/documentation/language`).
- **Verification is visual, not unit tests** for CSS tasks — this repo has no test suite and CSS correctness is visual. Each CSS task is verified by viewing the running app against the `StyleTest` fixture in **both** light and dark, plus `npm run type-check`. The one genuinely unit-testable piece (theme-resolution logic, Task 5) does get a Vitest test.
- **Single source of color:** the tokens in `theme.css`. Never introduce a new hardcoded color; every color resolves to a token.
- **Markdown is raw `v-html`** — style it only through the `.content` wrapper, never by adding classes to individual rendered elements.
- **Fonts:** Fraunces (`--serif`), Inter (`--sans`), Fira Code (`--mono`).
- **Design tokens (verbatim):**

  | Token | Light | Dark |
  |-------|-------|------|
  | `--bg` | `#F6F4FB` | `#1C1726` |
  | `--surface` | `#FDFCFF` | `#2A2338` |
  | `--surface-2` | `#EFEBF8` | `#211B2E` |
  | `--fg` | `#262036` | `#ECE7F4` |
  | `--muted` | `#6C6480` | `#A79FB8` |
  | `--accent` | `#6A46C9` | `#B3A4F5` |
  | `--border` | `#E4DCF2` | `#3A3350` |

- **Dev verification trick:** to view a specific page in the app, temporarily change `App.vue`'s `onMounted` dispatch `SetActivePage('MainPage')` → `SetActivePage('StyleTest')`, verify, then revert. (The `StyleTest` page is already seeded in the `Pages` collection.)

---

### Task 1: Load the web fonts

**Files:**
- Modify: `frontend/index.html` (the Google Fonts `<link>` lines)

**Interfaces:**
- Produces: Fraunces, Inter, Fira Code available to CSS via `font-family`.

- [ ] **Step 1: Replace the Rubik font link**

In `frontend/index.html`, replace the existing `<link href="https://fonts.googleapis.com/css2?family=Rubik&display=swap" ...>` line with:

```html
<link href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,400;9..144,600;9..144,700&family=Inter:wght@400;500;600;700&family=Fira+Code:wght@400;500&display=swap" rel="stylesheet">
```

Keep the two `preconnect` lines above it.

- [ ] **Step 2: Verify the fonts load**

Run `npm run dev` (if not running). In the browser dev tools Network tab, filter for `fonts.g` and confirm Fraunces / Inter / Fira Code stylesheets load with no 404. (No visible layout change yet — the CSS that uses them lands in later tasks.)

- [ ] **Step 3: Commit point** — ask the owner whether to commit (per Global Constraints).

---

### Task 2: Design tokens + base styles (`theme.css`)

**Files:**
- Create: `frontend/src/assets/theme.css`
- Modify: `frontend/src/assets/main.css` (body background/font)
- Modify: `frontend/src/main.ts` (import order)

**Interfaces:**
- Produces: CSS variables (tokens + `--serif`/`--sans`/`--mono`) on `:root` and `html[data-theme="dark"]`; base `body` typography.

- [ ] **Step 1: Create `frontend/src/assets/theme.css`**

```css
:root {
  --bg: #F6F4FB;
  --surface: #FDFCFF;
  --surface-2: #EFEBF8;
  --fg: #262036;
  --muted: #6C6480;
  --accent: #6A46C9;
  --border: #E4DCF2;

  --serif: 'Fraunces', Georgia, serif;
  --sans: 'Inter', system-ui, sans-serif;
  --mono: 'Fira Code', monospace;
}

html[data-theme="dark"] {
  --bg: #1C1726;
  --surface: #2A2338;
  --surface-2: #211B2E;
  --fg: #ECE7F4;
  --muted: #A79FB8;
  --accent: #B3A4F5;
  --border: #3A3350;
}

/* Default to dark when the user has expressed no explicit choice. */
@media (prefers-color-scheme: dark) {
  html:not([data-theme]) {
    --bg: #1C1726;
    --surface: #2A2338;
    --surface-2: #211B2E;
    --fg: #ECE7F4;
    --muted: #A79FB8;
    --accent: #B3A4F5;
    --border: #3A3350;
  }
}
```

- [ ] **Step 2: Update `main.css` body**

In `frontend/src/assets/main.css`, replace `background-color: darkslategray;` in the `body` rule with token-based base styles:

```css
body {
	height: 100%;
	background-color: var(--bg);
	color: var(--fg);
	font-family: var(--sans);
	font-size: 18px;
	line-height: 1.7;
	-webkit-font-smoothing: antialiased;
}
```

(Leave the `* { margin:0; padding:0; }` and `html { height:100% }` rules as they are. Remove the commented-out `background-image` line.)

- [ ] **Step 3: Import `theme.css` first in `main.ts`**

In `frontend/src/main.ts`, add the theme import as the **first** import so the variables exist before any other stylesheet:

```ts
import './assets/theme.css'
import './assets/main.css'
import './assets/sliderMenu.css'
```

- [ ] **Step 4: Verify**

Reload the app. The page background should be the light lavender `#F6F4FB` (or the dark tone if your OS is in dark mode and no `data-theme` is set yet). In dev tools, set `<html data-theme="dark">` and confirm the background flips to `#1C1726` and text to `#ECE7F4`. Run `npm run type-check` — passes.

- [ ] **Step 5: Commit point** — ask the owner whether to commit.

---

### Task 3: Wire Tailwind colors to the tokens

**Files:**
- Modify: `frontend/tailwind.config.js`

**Interfaces:**
- Produces: Tailwind color utilities (`bg-surface`, `text-fg`, `text-muted`, `text-accent`, `border-border`, `bg-bg`, `bg-surface-2`) resolving to the CSS variables; existing custom color names remapped to tokens; `font-serif`/`font-sans`/`font-mono` families.

- [ ] **Step 1: Update `tailwind.config.js` theme.extend**

Replace the `colors` and `fontFamily` blocks with:

```js
      colors: {
        // semantic tokens (theme-aware via CSS variables)
        bg: 'var(--bg)',
        surface: 'var(--surface)',
        'surface-2': 'var(--surface-2)',
        fg: 'var(--fg)',
        muted: 'var(--muted)',
        accent: 'var(--accent)',
        border: 'var(--border)',
        // legacy names remapped to tokens so existing classes become theme-aware
        midnightPurple: 'var(--surface)',
        mainComponentBackground: 'var(--surface)',
        activePageColor: 'var(--accent)',
        deActivePageColor: 'var(--muted)',
      },
      fontFamily: {
        serif: ['Fraunces', 'Georgia', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['Fira Code', 'monospace'],
        FiraCode: ['Fira Code', 'monospace'],
      },
```

- [ ] **Step 2: Verify build**

Run `npm run type-check` and confirm the dev server recompiles with no Tailwind/config errors. Components that already use `bg-midnightPurple` / `text-activePageColor` / `text-deActivePageColor` now resolve to tokens — the shell should already look different (surfaces follow `--surface`). Glance at the app to confirm it still renders (deeper shell polish is Tasks 7–8).

- [ ] **Step 3: Commit point** — ask the owner whether to commit.

---

### Task 4: Markdown content stylesheet (`.content`)

**Files:**
- Create: `frontend/src/assets/content.css`
- Modify: `frontend/src/views/contents.vue` (wrapper class)
- Modify: `frontend/src/main.ts` (import)

**Interfaces:**
- Consumes: tokens from Task 2.
- Produces: full editorial typography under `.content`.

- [ ] **Step 1: Create `frontend/src/assets/content.css`**

```css
.content { max-width: 65ch; margin: 0 auto; }

.content h1, .content h2, .content h3, .content h4 {
  font-family: var(--serif); font-weight: 600; color: var(--fg); line-height: 1.15;
}
.content h1 { font-size: 2.75rem; line-height: 1.1; letter-spacing: -.01em; margin: 0 0 .3em; }
.content h2 { font-size: 2rem; margin: 2.4em 0 .5em; }
.content h3 { font-size: 1.5rem; margin: 1.9em 0 .45em; }
.content h4 { font-size: 1.25rem; margin: 1.7em 0 .4em; }
.content h5 { font-family: var(--sans); font-size: 1.0625rem; font-weight: 700; margin: 1.6em 0 .4em; color: var(--fg); }
.content h6 {
  font-family: var(--sans); font-size: .875rem; font-weight: 600; text-transform: uppercase;
  letter-spacing: .06em; color: var(--muted); margin: 1.6em 0 .4em;
}

.content p { margin: 0 0 1.4em; }
.content > p:first-of-type { font-size: 1.25rem; line-height: 1.6; color: var(--muted); }
.content a { color: var(--accent); text-decoration: underline; text-underline-offset: 2px; text-decoration-thickness: 1px; }
.content a:hover { text-decoration-thickness: 2px; }
.content strong { font-weight: 600; }
.content em { font-style: italic; }
.content del { color: var(--muted); }

/* Tailwind Preflight removes list markers, so restore them explicitly. */
.content ul, .content ol { margin: 0 0 1.4em; padding-left: 1.4em; }
.content ul { list-style: disc; }
.content ol { list-style: decimal; }
.content ul ul { list-style: circle; }
.content li { margin: .4em 0; }
.content li::marker { color: var(--accent); }
.content ul ul, .content ol ol { margin: .4em 0; }

.content blockquote {
  margin: 1.6em 0; padding: .2em 0 .2em 1.1em; border-left: 3px solid var(--accent);
  color: var(--muted); font-style: italic;
}
.content blockquote blockquote { margin: .6em 0; border-left-color: var(--border); }

.content code {
  font-family: var(--mono); font-size: .9em; background: var(--surface-2);
  padding: .15em .4em; border-radius: 5px;
}
.content pre {
  background: var(--surface-2); border: 1px solid var(--border); border-radius: 10px;
  padding: 1.1em 1.3em; overflow-x: auto; margin: 0 0 1.4em; line-height: 1.6;
}
.content pre code { background: none; padding: 0; font-size: .9rem; }

.content table { width: 100%; border-collapse: collapse; margin: 0 0 1.4em; font-size: .95rem; }
.content th, .content td { padding: .6em .8em; text-align: left; border-bottom: 1px solid var(--border); }
.content th { font-weight: 600; color: var(--fg); }
.content thead th { border-bottom: 2px solid var(--border); }
.content tbody tr:last-child td { border-bottom: none; }

.content hr { border: 0; border-top: 1px solid var(--border); margin: 2.6em 0; }
.content img { max-width: 100%; height: auto; border-radius: 10px; display: block; margin: 1.4em 0; }
```

- [ ] **Step 2: Add the wrapper class in `contents.vue`**

In `frontend/src/views/contents.vue`, add `content` to the class of the div that carries `v-html`:

```html
<template>
	<div class="py-4 pl-4 pr-4">
		<div class="content" v-html="returnedHTML"></div>
	</div>
</template>
```

(Remove the redundant middle `<div>` wrapper; the `.content` div holds the rendered HTML.)

- [ ] **Step 3: Import `content.css` in `main.ts`**

Add after the other asset imports in `frontend/src/main.ts`:

```ts
import './assets/content.css'
```

- [ ] **Step 4: Verify against StyleTest (light + dark)**

Temporarily set `App.vue` `onMounted` to `GlobalStore.dispatch('SetActivePage', 'StyleTest')`. Reload the app and confirm: h1 large serif, clear heading hierarchy h1→h6, muted lead paragraph, violet list markers with nested indentation, violet blockquote border (nested lighter), lavender inline-code chips, panelled code block, row-separated table, rule, rounded image. Toggle `<html data-theme="dark">` in dev tools and confirm all elements re-theme with purple preserved. **Revert** the `App.vue` change to `'MainPage'`. Run `npm run type-check`.

- [ ] **Step 5: Commit point** — ask the owner whether to commit.

---

### Task 5: Theme resolution + no-flash init

**Files:**
- Create: `frontend/src/global/theme.ts`
- Create: `frontend/src/global/theme.spec.ts`
- Modify: `frontend/index.html` (inline no-flash script)

**Interfaces:**
- Consumes: `LocalStorageService` (`SaveData`, `GetData`) from `@/service/LocalStorage.service`.
- Produces: `resolveInitialTheme(): 'light'|'dark'`, `applyTheme(t)`, `getTheme(): 'light'|'dark'`, `toggleTheme(): 'light'|'dark'`. Storage key: `'theme'`.

- [ ] **Step 1: Write the failing test `frontend/src/global/theme.spec.ts`**

```ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { resolveInitialTheme, toggleTheme, getTheme } from './theme'

function mockMatchMedia(prefersDark: boolean) {
  window.matchMedia = vi.fn().mockImplementation((q: string) => ({
    matches: prefersDark, media: q, addEventListener: vi.fn(), removeEventListener: vi.fn(),
  })) as unknown as typeof window.matchMedia
}

describe('theme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
  })

  it('uses the stored theme when present, ignoring system preference', () => {
    localStorage.setItem('theme', 'light')
    mockMatchMedia(true) // system prefers dark
    expect(resolveInitialTheme()).toBe('light')
  })

  it('falls back to system preference when nothing is stored', () => {
    mockMatchMedia(true)
    expect(resolveInitialTheme()).toBe('dark')
  })

  it('toggle flips the attribute and persists the choice', () => {
    document.documentElement.setAttribute('data-theme', 'light')
    const next = toggleTheme()
    expect(next).toBe('dark')
    expect(getTheme()).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')
  })
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `npm run test:unit -- theme`
Expected: FAIL (`./theme` has no such exports / module not found).

- [ ] **Step 3: Implement `frontend/src/global/theme.ts`**

```ts
import { LocalStorageService } from '@/service/LocalStorage.service'

export type Theme = 'light' | 'dark'
const KEY = 'theme'
const ls = new LocalStorageService()

export function resolveInitialTheme(): Theme {
  const stored = ls.GetData(KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(t: Theme): void {
  document.documentElement.setAttribute('data-theme', t)
}

export function getTheme(): Theme {
  return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

export function toggleTheme(): Theme {
  const next: Theme = getTheme() === 'dark' ? 'light' : 'dark'
  applyTheme(next)
  ls.SaveData(KEY, next)
  return next
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `npm run test:unit -- theme`
Expected: PASS (3 tests).

- [ ] **Step 5: Add the no-flash inline script to `index.html`**

In `frontend/index.html`, add this as the first line inside `<head>` (before the font links) so `data-theme` is set before first paint:

```html
<script>
  (function () {
    try {
      var t = localStorage.getItem('theme');
      if (t !== 'light' && t !== 'dark') {
        t = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      }
      document.documentElement.setAttribute('data-theme', t);
    } catch (e) {}
  })();
</script>
```

- [ ] **Step 6: Verify**

Reload the app: no theme flash on load. In dev tools set `localStorage.theme='dark'` and reload → app is dark; set `'light'` and reload → light; `delete localStorage.theme` and reload → follows OS preference. `npm run type-check` passes.

- [ ] **Step 7: Commit point** — ask the owner whether to commit.

---

### Task 6: Theme toggle control

**Files:**
- Create: `frontend/src/components/ThemeToggle.vue`
- Modify: `frontend/src/views/sidePanel.vue` (mount the toggle)

**Interfaces:**
- Consumes: `getTheme`, `toggleTheme` from `@/global/theme`.

- [ ] **Step 1: Create `frontend/src/components/ThemeToggle.vue`**

```vue
<template>
	<button
		type="button"
		class="text-sm font-medium text-fg bg-surface border border-border rounded-full px-3 py-1 hover:border-accent"
		@click="onToggle"
	>
		{{ current === 'dark' ? '☾ Dark' : '☀ Light' }}
	</button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getTheme, toggleTheme, type Theme } from '@/global/theme'

const current = ref<Theme>(getTheme())
function onToggle() {
	current.value = toggleTheme()
}
</script>
```

- [ ] **Step 2: Mount it in the shell**

In `frontend/src/views/sidePanel.vue`, add the toggle at the top of the profile block (inside the `#profile-grid` div, before the profile component):

```html
<div class="flex justify-end p-3">
	<ThemeToggle />
</div>
```

Add the import in the `<script setup>`:

```ts
import ThemeToggle from '@/components/ThemeToggle.vue'
```

- [ ] **Step 3: Verify**

Reload the app. The toggle appears in the side panel. Clicking it flips the whole app between light and dark (content + shell), the label updates, and the choice survives a reload. `npm run type-check` passes.

- [ ] **Step 4: Commit point** — ask the owner whether to commit.

---

### Task 7: Tokenize raw CSS + shell surfaces

**Files:**
- Modify: `frontend/src/assets/sliderMenu.css`
- Modify: `frontend/src/App.vue`

**Interfaces:**
- Consumes: tokens (Task 2), Tailwind token classes (Task 3).

- [ ] **Step 1: Tokenize `sliderMenu.css`**

Replace the hardcoded colors:
- `#sliding` → `background-color: var(--surface);` and `border-color: var(--border);`
- `.topMenu:hover` → `background-color: var(--surface-2);` (replace `aqua`)
- `.leftMenu:hover` → `background-color: var(--surface-2);` (replace `aqua`)
- `.menuIcon` → `color: var(--fg);` (replace `white`)

- [ ] **Step 2: Tokenize the App.vue shell frame**

In `frontend/src/App.vue`, the `RouterView` and `sidePanel` currently use `bg-midnightPurple border-2 rounded` (now token-backed via Task 3). Change `border-2` framing to use the border token by adding `border-border`, e.g. `bg-surface border-2 border-border rounded m-4`. Apply the same `border-border` to both the `RouterView` and the `<routes.sidePanel>` elements.

- [ ] **Step 3: Verify (light + dark)**

Reload. Confirm the page background is `--bg`, the content and side panels are `--surface` with `--border` borders, the sliding menu is `--surface` with a sane hover (no aqua), and the menu icon uses `--fg`. Toggle themes and re-check. `npm run type-check` passes.

- [ ] **Step 4: Commit point** — ask the owner whether to commit.

---

### Task 8: Tokenize remaining components

**Files:**
- Modify: `frontend/src/components/menuComponents/MenuItem.vue`
- Modify: `frontend/src/views/sidePanel.vue`
- Modify: `frontend/src/components/contentViewComponents/upperProfile.vue`
- Modify: `frontend/src/components/sidePanelComponents/PanelMenu.vue`
- Modify: `frontend/src/components/panelComponents/LoginComponent.vue`

**Interfaces:**
- Consumes: Tailwind token classes (Task 3).

- [ ] **Step 1: MenuItem.vue**

Active/inactive text already map to tokens via config (`text-activePageColor` → accent, `text-deActivePageColor` → muted). Change the hover class `hover:text-activePageColor` to `hover:text-accent` for clarity. Confirm the item text is readable in both themes.

- [ ] **Step 2: sidePanel.vue hr + text**

Replace the inline `<style scoped>` `hr` border colors (`rgba(255,255,255,.3)` / `rgba(0,0,0,.08)`) with `border-top: 1px solid var(--border);` and remove the second border line. Ensure any profile text uses `text-fg` / `text-muted`.

- [ ] **Step 2b: upperProfile.vue**

The profile card uses `text-slate-50` (near-white) for the name and job title, which is invisible on a light surface. Replace the name's `text-slate-50` with `text-fg` and the job title's `text-slate-50` with `text-muted`.

- [ ] **Step 3: PanelMenu.vue**

Replace `text-white` → `text-fg`, `hover:bg-slate-900/50` → `hover:bg-surface-2`, and `text-primary-500 dark:text-primary-400` → `text-accent`.

- [ ] **Step 4: LoginComponent.vue**

Replace the gray palette with tokens: outer card `bg-gray-500 ... dark:bg-gray-800 dark:border-gray-700` → `bg-surface border border-border`; input classes `bg-gray-50 border-gray-300 text-gray-900 ...` → `bg-surface-2 border border-border text-fg`; labels `text-gray-900 dark:text-white` → `text-fg`; button `bg-gray-300 border-gray-400 hover:bg-gray-400` → `bg-accent text-surface border border-border hover:opacity-90`.

- [ ] **Step 5: Verify (light + dark)**

Navigate to `/` and `/panel`. Confirm nav active/inactive states, the profile panel divider, the panel menu, and the login form are all readable and theme-correct in both themes. `npm run type-check` passes.

- [ ] **Step 6: Commit point** — ask the owner whether to commit.

---

### Task 9: Full-app verification + cleanup

**Files:**
- Delete: `frontend/public/__preview.html`

- [ ] **Step 1: Remove the throwaway preview**

Delete `frontend/public/__preview.html` (the brainstorming mockup; it is gitignored so was never committed).

- [ ] **Step 2: Full verification pass**

With the app running, check in **both** themes: MainPage (`/`) renders cleanly; `/lists` and `/panel` render; then temporarily point `App.vue` at `StyleTest`, confirm the full kitchen-sink looks right in both themes, and revert to `MainPage`. Confirm the theme toggle persists across reload with no flash. Run `npm run type-check` and `npm run build` — both pass.

- [ ] **Step 3: Commit point** — ask the owner whether to commit.

---

## Notes for the implementer

- Do the tasks in order; Tasks 4–8 all depend on Tasks 2–3 (tokens + Tailwind wiring).
- If a token value needs nudging when seen in-app, change it in **one** place: `theme.css`. Do not hardcode a color anywhere else.
- The `StyleTest` fixture is the source of truth for "does arbitrary Markdown look good." If an element looks wrong, fix it in `content.css`, not per-page.
