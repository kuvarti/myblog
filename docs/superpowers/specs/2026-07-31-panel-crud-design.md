# Admin Panel — Page CRUD (with menu binding) — Design Spec

- **Date:** 2026-07-31
- **Status:** Approved (design) — pending implementation plan
- **Branch (to be created):** `panel-crud`

## Goal

Deliver the project's core purpose: manage site content through the admin panel
instead of editing the database or source files. This first slice lets an
authenticated owner **create, edit, and delete pages authored in Markdown** and,
in the same action, **bind each page to a navigation (menu) entry** so a new page
is reachable without touching MongoDB. A **live preview** renders the page exactly
as it will appear when published.

## Fixed decisions (from brainstorming)

1. **Scope = full loop:** page CRUD **plus** menu binding (create page + make it
   appear in nav as one action).
2. **Editor format = clean Markdown, `/n` hidden:** the author works with real
   newlines; the bespoke `/n` line-delimiter stays a backend storage detail.
   Conversion happens at the backend boundary. The "a line containing `<` is raw
   HTML passthrough" rule is untouched (we only translate the delimiter, never the
   content).
3. **Live preview:** a stateless backend `Preview` endpoint renders raw source
   without persisting, reusing the existing render pipeline so preview == published.
4. **Frontend delegated to Claude; keep it minimal:** a plain `<textarea>` text
   editor is enough for now; no UI polish beyond existing tokens + PrimeVue.
   Easily revised later.
5. **Auth unchanged this slice:** all CRUD endpoints sit behind the existing
   `ControlPanel` JWT middleware. JWT secret placeholder (`"SecretKey"`) and the
   1-hour token are **not** hardened here (deliberately deferred).

## Scope

**In scope**
- Backend: Page CRUD + stateless Preview + Menu upsert/delete, all under the
  JWT-protected `ControlPanel` route group.
- Frontend: authenticated panel with a page list, a plain text editor, live
  preview, page settings (ViewType + menu Caption/Path), save, and delete.
- **Minimal public-nav wiring** so a clicked menu item actually loads its page:
  `MenuItem` dispatches `SetActivePage(PageName)` and routes to `/`, and
  `contents.vue` re-fetches reactively when `ActivePage` changes. Per-page URLs /
  bookmarkable deep links remain out of scope.
- A one-time user-seeding procedure (owner runs it; the password is never handled
  by Claude).
- Focused unit tests (pure conversion + preview render; frontend state + service
  token attachment). End-to-end manual verification on desktop and mobile.

**Out of scope (future slices)**
- Auth hardening (real secret, refresh, logout UI, user management).
- Standalone Menu management section (reordering, editing menu items unrelated to a
  page).
- **Per-page URLs / bookmarkable deep links.** `ActivePage` stays in-memory (Vuex),
  so refreshing `/` returns to the default page. True per-page routing is a later
  slice. The minimal nav wiring above only makes an in-session click load the page.
- `ViewType`-driven render branching on the public site.
- Page rename (achieved via delete + create for now).
- Rich editor affordances (toolbar, image upload, syntax highlighting).

## Backend design

### New controller (composition root)

Introduce `Controllers/Panel.Controller.go` with
`InitPanelController(pageSvc, menuSvc, tokenSvc, authGroup)`. It creates
`authGroup.Group("ControlPanel")`, applies `tokenSvc.AuthenticateJWT()`, and
registers the routes below. Remove the now-empty `ControlPanel` block from
`InitUserController` so the group is owned in exactly one place. Wire the new
controller in `server.go` `InitControllers` (it already opens the `Pages` and
`Menus` collections and constructs `PageService` / `MenuService`).

### Endpoints (all under `/api/auth/ControlPanel`, JWT-protected)

| Method | Path | Body | Result | Purpose |
|--------|------|------|--------|---------|
| `GET` | `/Pages` | — | `[{ PageName, ViewType }]` | List pages for the sidebar |
| `GET` | `/Pages/:name` | — | `{ PageName, Source, ViewType, Menu: {Name, Caption, Path} \| null }` | Raw source (`/n`→`\n`) + menu binding for editing |
| `POST` | `/Pages` | `{ PageName, Source, ViewType, Menu?: {Caption, Path} }` | `201` | Create page (+ upsert menu) |
| `PUT` | `/Pages/:name` | `{ Source, ViewType, Menu?: {Caption, Path} }` | `200` | Update page (+ upsert menu) |
| `DELETE` | `/Pages/:name` | — | `200` | Delete page (+ cascade delete its menu entry) |
| `POST` | `/Preview` | `{ Source }` | `{ Html }` | Stateless render; no DB write |

### Service additions

`PageService` interface gains:
- `List() ([]PageSummary, error)` — `Find` with a `{PageName, ViewType}` projection.
- `GetRaw(name string) (PageModel, error)` — `FindOne` only; returns the doc
  **without** re-rendering or mutating `Hash`/`Text`. The controller converts
  `Page` (`/n`) → clean newlines for the response `Source`.
- `Create(name, sourceClean, viewType string) error` — 409 if `PageName` already
  exists (explicit `FindOne` check; Mongo has no unique index here); stores
  `Page = toStorage(sourceClean)`, `Hash = []`, `Text = ""` (lazy render on first
  public `GET /api/Page`).
- `Update(name, sourceClean, viewType string) error` — 404 if missing; `$set`
  `Page`/`ViewType` and reset `Hash`/`Text` to force re-render on next fetch.
- `Delete(name string) error` — 404 if nothing deleted.
- `Preview(sourceClean string) (string, error)` — mirrors `GetPage`'s render path
  exactly: `GetPageText(toStorage(sourceClean))` then strip `\n`. Reuses the
  existing `GetPageText`/`ConvertmdToHTML`.
  - **Render-pipeline exception (added during implementation):** `GetPageText` had
    a latent index-out-of-range panic for any source whose final line is Markdown
    (no trailing `<` line or blank separator). Preview surfaced it, and it would
    also crash public `GetPage` for such pages. A minimal, behavior-preserving fix
    bounds the inner accumulation loop (`for i < len(lines) && !contains(lines[i], "<")`),
    covered by a regression test. This narrowly relaxes the original "do not touch
    the render pipeline" constraint — for a correctness bug fix only.

`MenuService` interface gains:
- `Upsert(m MenuModel) error` — upsert by `PageName` (`Name` defaults to
  `PageName` when omitted; `Caption`/`Path` from the request).
- `DeleteByPageName(name string) error` — used by the page delete cascade.

### Conversion helpers (backend, in the page service package)

```
toStorage(clean)   = strings.ReplaceAll(clean,  "\n", "/n")
fromStorage(stored)= strings.ReplaceAll(stored, "/n", "\n")
```

Round-trips are newline-lossless. **Known inherited limitation:** content
containing a literal `/n` substring (e.g. `and/next`) is mangled by the existing
format and remains so — this conversion does not fix or worsen it beyond today's
behavior. Documented, not solved in this slice.

### Status codes / error contract

Panel endpoints return proper codes: `400` (missing/empty `PageName` or `Source`),
`404` (page not found on update/delete/get), `409` (duplicate `PageName` on
create), `401` (missing/invalid token — emitted by the existing middleware). Error
bodies use `gin.H{"error": <message>}`. (This is stricter than some existing
handlers that return `200` with an error string; new code uses correct codes.)

## Frontend design

Minimal and delegated to Claude. Reuses the existing token system and PrimeVue
components; no new design language.

### Layout (uses the existing `/panel` shell)

`sidePanel.vue` already swaps the profile for `<Menu/>` (= `PanelMenu.vue`) on the
`/panel` route. We use that:
- **`PanelMenu.vue`** (teleported side panel): the **page list + "New page"**
  action, replacing today's dummy File/Edit items. Selecting a page drives the
  editor. Delete affordance lives here or in the editor toolbar (implementer's
  choice); delete always confirms first.
- **`panel.vue`** (main `RouterView`, shown when authenticated): the **editor +
  live preview** for the selected/new page.

### Shared state — `global/panelState.ts`

A lightweight reactive singleton (no Vuex boilerplate) shared across the teleport
boundary: `pages`, `selected`, `dirty`, `select(name)`, `startNew()`,
`refresh()`. (Alternative: a Vuex `panel` module — more consistent with existing
patterns but more verbose; chosen against for simplicity.)

### `Panel.service.ts`

Extends `serviceClass` with base path `/auth/ControlPanel`. Reads `AuthToken` from
`LocalStorage.service.ts` and attaches `Authorization: Bearer <token>` on every
call (same pattern as `AmIAuth`). Methods: `listPages`, `getPage`, `createPage`,
`updatePage`, `deletePage`, `preview`. A **401 interceptor** flips
`UserService.IsLogin` to `false` (session expired → login screen). Types live in
`types/` (`PanelPageSummary`, `PanelPageDetail`, `MenuBinding`).

### Editor (`PageEditor.vue` inside `panel.vue`)

- **Settings row:** `PageName` (editable only when creating; read-only when
  editing — immutable key), `ViewType`, menu `Caption` + `Path`.
- **Split area:** plain `<textarea>` (clean newlines) on the left; **live preview**
  on the right. Preview is a **debounced (~400 ms)** `preview(source)` call whose
  returned HTML is injected into a `.content` wrapper — so it looks identical to a
  published page (reuses `content.css`).
- **Toolbar:** Save (create vs update by mode) and Delete (with confirm).
- **Dirty guard:** `dirty` flag; switching pages with unsaved edits prompts to
  confirm. Save/create clears `dirty` and refreshes the list.

### Auth guard

Reactive `UserService.IsLogin`: valid token → admin UI; otherwise
`LoginComponent`. The login flow already works. A small `loading` state suppresses
the brief login flash while `AmIAuth` resolves on first load.

### Public nav wiring (closes the create → click → open loop)

Today `MenuItem.vue` only does `router.push(Path)` and ignores `PageName`;
`contents.vue` fetches only in `onMounted`. Two small changes make a clicked menu
item load its page:
- **`MenuItem.vue`:** on click, if `PageName` is set, dispatch
  `SetActivePage(PageName)` and `router.push('/')`; otherwise keep the existing
  `router.push(Path)` (so `/lists`-style items are unchanged).
- **`contents.vue`:** extract the fetch into a function called from `onMounted`
  **and** a `watch` on `GetActivePage`, so navigating while on `/` re-renders.

This closes the loop for the current single-content-route model. It does **not**
add per-page URLs (see Out of scope). The active-item highlight in `MenuItem`
(currently Path-segment based) may not track the selected page — cosmetic, left as
a known limitation.

## Data flow (happy path)

1. Open `/panel` → `AmIAuth` → authed → `panelState.refresh()` → sidebar list.
2. Select page → `getPage(name)` → editor populated (clean source, ViewType, menu),
   `dirty=false`.
3. Edit → `dirty=true` → debounced `preview` → preview pane updates.
4. New → `startNew()` → empty editor, `PageName` editable.
5. Save → create (`POST`) or update (`PUT`) → success → `dirty=false`, refresh,
   select the saved page. `409` on create → inline "name already exists".
6. Delete → confirm → `DELETE` → refresh, clear editor.

## Testing strategy

- **Backend unit:** pure `toStorage`/`fromStorage` round-trip; `Preview` render
  (no DB) producing expected HTML for a small Markdown + raw-HTML mixed input.
  Full CRUD service tests need a Mongo test double → optional/stretch this slice.
- **Frontend unit (Vitest/jsdom):** `panelState` transitions; `Panel.service`
  attaches the Bearer header (axios mocked) and the 401 interceptor flips
  `IsLogin`.
- **End-to-end (manual, primary):** seed user → login → create page + menu → see it
  in nav → open it → edit → preview matches → save → delete. Verified on desktop
  and on mobile via DevTools device emulation.

## Prerequisite — seed a user (owner runs; password never handled by Claude)

Login compares the submitted password (bcrypt) against a stored bcrypt hash in the
`Users` collection (`{ Username, Password, UserType }`). No user is seeded. The
implementation plan will include a one-time procedure that generates a bcrypt hash
and inserts the user; **the owner supplies and types the password**, Claude does
not see or enter it.

## Known limitations (accepted for this slice)

- Literal `/n` in page content is mangled (inherited from the existing format).
- Auth is not production-grade: placeholder JWT secret, no logout/refresh, single
  gate. Deliberately deferred.
- No page rename (delete + create instead).
- Clicked pages are not bookmarkable; refreshing `/` returns to the default page
  (`ActivePage` is in-memory, not in the URL).
- The active menu-item highlight may not track the selected page (cosmetic).
