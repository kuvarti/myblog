# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Owner preferences live in `memory/`.** Standing preferences (documentation language, upkeep rules, etc.) are kept as individual editable documents under `memory/`, organized by category — see `memory/MemoryArchitectures.md` for the scheme. Read them and follow them. When the owner states a new preference, record it there rather than inflating this file. Two that always apply: **this document is never final** — keep it updated in the same session a change invalidates it; and **all documentation is written in English**, proper nouns excepted.

## Overview

Personal blog by Ahmet Eryilmaz. A three-part project:

- `backend/` — Go (Gin) REST API backed by MongoDB
- `frontend/` — Vue 3 + Vite + TypeScript SPA
- `database/` — MongoDB `docker-compose.yml` plus JSON seed data in `SimpleBackup/`

The frontend talks to the backend at `http://localhost:8080/api`; the backend talks to MongoDB at `localhost:27017`. All three must run together for the app to work. Note the README's "no database planned" line is stale — the project uses MongoDB.

## Purpose & design intent

The site is built to be **maintained without ever touching HTML or code again after the initial build**. This goal drives the whole architecture:

- **Every page is authored in Markdown**, stored in the DB (`Page` field of the `Pages` collection) and rendered to HTML on demand (see Page rendering). Raw HTML tags may be embedded in the Markdown when strictly necessary, but the intended primary structure of every page is always Markdown, not HTML.
- **Content is managed through the (planned) admin panel** — pages are edited, added, and removed there, never by editing source files. The control-panel scaffolding in the WIP section exists to fulfill this.
- **Consequence for styling:** because arbitrary Markdown pages are poured into one fixed layout, the project needs a single solid, definitive global stylesheet that makes any rendered `.md` content look good on its own — with no per-page CSS. Frontend styling work should target "style the rendered Markdown output (and the shell) well," not "style individual pages." Getting this stylesheet right is a primary, deliberate goal, not an afterthought.

## Project map

Folder-level layout only (individual files follow the naming conventions below and are discoverable with `ls`). Keep in sync when directories are added or renamed.

```
myblog/
├── memory/                    — owner preferences; index in memory/MemoryArchitectures.md
├── database/
│   └── SimpleBackup/          — JSON seed for the Menus and Pages collections (imported manually)
├── backend/                   — Go + Gin API (:8080); server.go is the entrypoint + composition root
│   ├── Models/                — data structs (json/bson tags)
│   ├── Services/              — business logic + all Mongo access (interface + Impl per resource)
│   └── Controllers/           — Gin handlers (one InitXController per resource)
└── frontend/                  — Vue 3 + Vite + TS SPA (:5173)
    └── src/
        ├── router/            — routes + view barrel
        ├── global/            — Vuex store, icon registration
        ├── service/           — axios API layer (BaseAPI + User + LocalStorage)
        ├── views/             — one component per route + the teleported side menu/panel
        ├── components/        — feature components (menu, content, panel, sidePanel, utils)
        └── types/             — TypeScript interfaces
```

## Commands

### Database (run first)
```sh
cd database && docker compose up -d      # MongoDB on :27017, mongo-express UI on :8081
```
Credentials are hardcoded as `root:example`. The database is named `KuvartiBlog`. Seed data lives in `database/SimpleBackup/*.json` (collections `Menus`, `Pages`) and must be imported manually — there is no automated seeding. The files are `$oid`-style extended-JSON arrays; import with `--jsonArray`:
```sh
cd database
mongoimport --uri "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --collection Menus --file SimpleBackup/Menus.json --jsonArray
mongoimport --uri "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --collection Pages --file SimpleBackup/Pages.json --jsonArray
```
Users are not seeded — create one in the `Users` collection with a bcrypt-hashed `Password` (fields: `Username`, `Password`, `UserType`) before login will work.

### Backend (`cd backend`)
```sh
air            # hot-reload dev server (config in .air.toml, builds to ./tmp/main)
go run .       # run once without hot reload
go build -o ./tmp/main .
go test ./...  # run the unit-test suite
```
The server listens on `:8080` (Gin default). The unit tests are pure logic tests — stdlib `testing`, no MongoDB or HTTP server needed: `buildNav` filtering/order/caption in `Controllers/Menu.Controller_test.go`, path validation in `Controllers/Panel.Controller_test.go`, the render + `ToStorage`/`FromStorage` helpers in `Services/Page.Service_test.go`, and the card helpers (`extractSummary`, `extractImage`, `cardTitle`, `selectByTags`, `expandShortcodes`) in `Services/Card.Service_test.go`. Run one test with e.g. `go test ./Services -run TestPreviewRendersMarkdownHeading`.

### Frontend (`cd frontend`)
```sh
npm install
npm run dev          # Vite dev server on :5173 (--host enabled)
npm run build        # type-check (vue-tsc) + production build
npm run type-check
npm run lint         # eslint --fix across the repo
npm run format       # prettier on src/
npm run test:unit    # vitest (jsdom); starts in watch mode — append `-- run` for a single pass
```
Spec files live next to the code they cover (`*.spec.ts`): `global/` (theme, panelState, notify), `service/Panel.service.spec.ts`, and the component specs under `components/` (menuActive, scrollSync, menuReorder, contentLinks, `panelComponents/tags`). Run a single test: `npm run test:unit -- <path-or-name-pattern>`.

## Configuration touchpoints

There is no `.env` file; most environment values are hardcoded. The one exception is the MongoDB URI, read from the `MONGODB_URI` env var (e.g. an Atlas `mongodb+srv` string) with the local URI as a fallback when unset. Changing an environment, port, or credential means editing all of the relevant spots below by hand (there is no single source of truth, so it is easy to miss one):

| Value | Where | Current |
|-------|-------|---------|
| Backend API base URL | `frontend/src/service/BaseAPI.service.ts` | `http://localhost:8080/api` |
| MongoDB URI + credentials | `backend/server.go` (`InitDB`) — reads `MONGODB_URI` env var, else fallback | `mongodb://root:example@localhost:27017` (fallback) |
| Database name | `backend/server.go` (`main`) | `KuvartiBlog` |
| CORS allowed origins | `backend/server.go` (`InitServer`) | `http://localhost:5173`, `https://kuvarti.github.io` (+ `FRONTEND_ORIGIN` env if set) |
| JWT signing secret | `backend/Services/Token.Service.go` | `"SecretKey"` (TODO placeholder) |
| Mongo container creds / ports / mongo-express | `database/docker-compose.yml` | `root` / `example`, `27017`, UI on `8081` |

Ports that are **not** set explicitly (rely on framework defaults): backend Gin server on `:8080` (`server.Run()` with no argument) and Vite dev server on `:5173`. `InitServer` registers a single CORS middleware with an explicit origin allowlist (`http://localhost:5173`, `https://kuvarti.github.io`, plus the `FRONTEND_ORIGIN` env var when set) and `AllowCredentials: true`. Cross-origin is real in production (GitHub Pages frontend → onrender backend are different origins), so a deployed frontend origin must be in this list or the browser blocks it; add new frontends via `FRONTEND_ORIGIN` rather than reopening to all origins.

## Backend architecture

Strict three-layer separation, wired manually in `server.go`:

- **Models/** — plain structs with `json`/`bson` tags (`UserModel`, `PageModel`, `MenuModel`, `JWTToken`, `LoginModel`).
- **Services/** — business logic + all MongoDB access. Each service is an interface (`FooService`) with an `Impl` struct constructed via `NewFooService(ctx, collection)`. Controllers depend on the interface, not the implementation.
- **Controllers/** — Gin handlers. Each has `InitFooController(service, apiGroup)` that registers routes on the shared `/api` router group.

**Naming convention.** Backend files are `<Resource>.<Layer>.go` (e.g. `Menu.Controller.go`, `Page.Service.go`, `User.Model.go`); each new resource adds one file per layer with the same prefix. Frontend service classes follow `<Name>.service.ts`.

`server.go` `InitControllers` is the composition root: it opens each MongoDB collection (`Users`, `Menus`, `Pages`) and injects it into the matching service, then hands the service to the controller. To add a resource, follow this Model → Service (interface + impl) → Controller → wire-in-`InitControllers` chain.

`Services/Card.Service.go` (`CardService`) is a fourth service that doesn't own a collection: it's a "page → card" expansion layer built on top of `PageService` (`GetRawByPath`, `FindByTags`) and `MenuService` (captions for card titles) — see Card rendering below. `InitControllers` constructs it from the already-built `pageService`/`menuService` and injects it into both `PageController` and `PanelController`. Relatedly, `PageService.Create`/`Update` no longer take a growing positional-argument list — they take one `models.PageWrite` struct (`PageName, Path, Source, ViewType, Tags, Summary, Image, ListTags`).

### Database collections
MongoDB is schemaless — document shape is defined by the Go structs (bson tags), not by the database. Field-by-field shapes live in the model files; this table only maps each collection to its authoritative model and owning service, plus the non-obvious bits:

| Collection | Shape (authoritative) | Owner service | Notes |
|-----------|-----------------------|---------------|-------|
| `Users` | `Models/User.Model.go` (`UserModel`) | `UserService` | No seed — create manually with a bcrypt-hashed password |
| `Menus` | `Models/Menu.Model.go` (`MenuModel`) | `MenuService` | Supplies the nav **caption** only; page membership/order/visibility/`Path` are owned by `Pages`. DB docs have `_id` (unmapped) and a legacy `Path` field that is no longer read |
| `Pages` | `Models/Pages.Model.go` (`PageModel`) | `PageService` | `Path` is the unique route key (see Routing); `Hash`/`Text` are machine-managed render cache, not hand-authored (see Page rendering below); `Tags []string`/`Summary string`/`Image string`/`ListTags []string` are optional fields (no migration needed) that back the card/List system — written by `PageService`, read by `CardService` at request time (see Card rendering below) |

### API endpoints
All routes are under the `/api` group. Auth column = requires `Authorization: Bearer <token>` header.

| Method | Path | Auth | Body / Query | Purpose |
|--------|------|------|--------------|---------|
| POST | `/api/auth/login` | — | `{ "userName", "passWord" }` | Verify credentials, return `{ Message, token }` |
| GET | `/api/auth/AmIAuth` | Bearer | — | Token check (returns `{ "Yes": "YES" }`) |
| GET | `/api/MenuList/Menu` | — | — | All menu items (`Menus` collection) |
| GET | `/api/Page` | — | `?Path=<path>` or `?PageName=<name>` | Rendered page HTML, **card-expanded** (see Card rendering), + `ViewType`; resolves by `Path` (per-page routing) or `PageName`. Missing both → `400` |
| GET | `/api/auth/ControlPanel/Pages` | Bearer | — | List `[{PageName, Path, ViewType, Order, Visible}]` |
| GET | `/api/auth/ControlPanel/Pages/:name` | Bearer | — | Raw source (clean newlines) + `Path` + `Tags`/`Summary`/`Image`/`ListTags` + menu binding for editing |
| POST | `/api/auth/ControlPanel/Pages` | Bearer | `{PageName, Path, Source, ViewType, Tags?, Summary?, Image?, ListTags?, Menu?}` | Create page (+ upsert menu); `409` duplicate name, `400` bad path, `422` reserved/taken path |
| PUT | `/api/auth/ControlPanel/Pages/:name` | Bearer | `{Path, Source, ViewType, Tags?, Summary?, Image?, ListTags?, Menu?}` | Update page (+ upsert menu); PageName immutable; `400` bad path, `422` reserved/taken path |
| DELETE | `/api/auth/ControlPanel/Pages/:name` | Bearer | — | Delete page (+ cascade its menu entry) |
| POST | `/api/auth/ControlPanel/Preview` | Bearer | `{Source}` | Stateless render → `{Html}`, with `<card>` shortcodes expanded (not the `List` grid; no DB write) |

The `ControlPanel` endpoints power the admin panel (`Controllers/Panel.Controller.go`). They speak **clean newlines**; the bespoke `/n` line delimiter stays backend-only, converted with `services.ToStorage` / `services.FromStorage`.

### Auth flow
- `POST /api/auth/login` verifies credentials with bcrypt against the `Users` collection, then `TokenService.GenerateToken` issues an HS256 JWT (1-hour expiry).
- Protected routes use the `TokenService.AuthenticateJWT()` Gin middleware, which expects an `Authorization: Bearer <token>` header. `/api/auth/AmIAuth` sits behind it.
- `UserService` and `TokenService` are both returned from `NewUserService` (they share the request context).
- **Gotcha:** the JWT `secretKey` in `Services/Token.Service.go` is the literal string `"SecretKey"` (marked `TODO`). Do not treat auth as production-grade.

### Page rendering (`Services/Page.Service.go`)
A page document in the `Pages` collection has: `Page` (the hand-authored source), `Text` (the cached rendered HTML — don't edit by hand), `Hash` (SHA-1 of `Page`, the cache key), plus `PageName`, `Path` (the unique route key — see Routing), and `ViewType`.

On `GetPage`/`GetPageByPath`, the service SHA-1-hashes the raw `Page` and compares it to the stored `Hash`; if the content changed (or `Text` is empty) it re-renders via `GetPageText` + `gomarkdown` and caches the new `Text`/`Hash` back into MongoDB (keyed on `PageName`), otherwise it returns the cached `Text`. So editing `Page` triggers a re-render on next fetch; editing `Text` directly is pointless.

**Source authoring format** (`Page` field) — this is a bespoke line-based mix of Markdown and HTML, not standard Markdown:
- Lines are separated by the literal token `/n` (slash-n), which `GetPageText` converts to real newlines before splitting; `/n/n` is a blank separator.
- A line **containing `<`** is emitted as **raw HTML, verbatim** — Markdown is not applied to it.
- A run of **consecutive lines with no `<`** is treated as one **Markdown block** and rendered together.
- **Blank lines are dropped** (they produce no output on their own).
- **Gotcha:** any `<` anywhere on a line switches that whole line to HTML passthrough, so a Markdown line like `a < b` or `List<T>` won't render as Markdown — keep `<` out of Markdown lines. Also, the final rendered HTML has all newlines stripped (`ReplaceAll(Text, "\n", "")`), and Markdown links open in a new tab (`HrefTargetBlank`).

**Authoring a new page:** the intended way is the **admin panel** (create/edit/delete + live preview, with a bound menu entry — see the `ControlPanel` endpoints and Frontend architecture). Under the hood a page is a `Pages` document with `PageName`, a unique `Path` (its route), a `Page` string in the format above, and empty `Hash`/`Text` (both fill in on the first `GET /api/Page?Path=<path>`), plus a matching `Menus` entry to surface it in the nav.

Editing this rendering/caching logic is the trickiest part of the backend.

### Card rendering (`Services/Card.Service.go`)
A second, separate expansion pass turns pages into linked cards. It layers on top of the Markdown `Text` cache above but is **never itself cached** — it re-runs on every request, so edits to a *referenced* page's card data show up immediately:

- **`<card path="/some/path">` shortcode** — placeable on its own line in any page's Markdown source. A line containing `<` is passed through raw by `GetPageText` (see above), so the shortcode survives rendering unchanged; `CardService` then string-replaces it over the already-rendered `Text` (matched by `` <card\s+path="([^"]*)"\s*/?> ``). An unresolved `path` (no such page) is dropped so no raw shortcode leaks to the browser.
- **Card data is auto-extracted from the referenced page**, read via `PageService.GetRawByPath` — an un-rendered, raw-source lookup, so resolving a card never triggers or recurses into that page's own card expansion. Title = the referenced page's menu `Caption` (via `MenuService.GetByPageName`), falling back to its `PageName`. Summary = its `Summary` field if set, else the first Markdown paragraph of its source — the first non-blank line that isn't raw HTML, a heading, or a Markdown image, with list/blockquote markers and emphasis characters (`*`, `_`, `` ` ``) stripped, truncated to ~160 chars with an ellipsis. Image = its `Image` field if set, else the first `<img src>` or Markdown `![]()` found in its source.
- **`ViewType = "List"`** turns a page into a dynamic card grid: after the page's own content renders (intro text is allowed and renders normally), `CardService.ExpandCards` calls `PageService.FindByTags(page.ListTags)` and appends a `<div class="card-grid">` of every page sharing **at least one** tag with `ListTags` (OR match across tags), excluding the page itself, sorted by `Order`. Menu `Hidden` is not consulted — list membership is independent of nav visibility. `DynamicList` (an earlier, separate planned `ViewType`) was dropped; its tag-driven behaviour folded entirely into `List`.
- **`PageModel` fields backing this**: `Tags []string` (the page's own tags), `Summary string` / `Image string` (per-page card overrides), `ListTags []string` (meaningful only when `ViewType == "List"`). All four are optional — absent means old behaviour, so existing pages need no migration.
- Card expansion runs on `GET /api/Page` (`PageController.GetPage` → `CardService.ExpandCards`, falling back to the un-expanded `Text` on error rather than failing the request) and, shortcode-only, on `POST /Preview` (`PanelController.Preview` → `CardService.ExpandShortcodes`) so the editor's live preview shows real cards; the `List` grid itself is never previewed — it depends on the saved `ViewType`/`ListTags`, not the in-progress editor source.
- Out of scope for now: an inline `<cardlist tag="…">` shortcode for a hand-placed dynamic subset inside a page (today only whole-page `List` and the static `<card path>` shortcode exist).

## Frontend architecture

Vue 3 `<script setup>` + TypeScript, Vuex 4, Vue Router 4, PrimeVue, Tailwind (+ SCSS).

- **Two mount points** (`index.html`): `#sidemenu` and `#app`. `App.vue` `<teleport>`s the sliding side menu into `#sidemenu`, keeping it outside the routed content grid.
- **Routing** (`router/index.ts`): two reserved routes — `/panel` → `panel`, `/lists` → `lists` — followed by a catch-all `/:pathMatch(.*)*` → `contents` (kept last so the reserved routes win). Every content page lives at its own URL: `contents.vue` reads `route.path` and fetches `GET /api/Page?Path=<route.path>`, re-fetching on `route.path` change. The home page is the one whose `Path` is `/`. View components are re-exported as a barrel in `router/routes.ts`. `contents.vue` also delegates a click handler on the `.content` wrapper (`onContentClick` → `internalNavTarget` in `components/contentLinks.ts`) that intercepts internal links — `href` starting with `/` and no `target`, which covers `<card>` links and the `List` grid's card links — and routes them client-side via `router.push` instead of a full reload; ordinary Markdown body links keep `target="_blank"` (`HrefTargetBlank`, see Page rendering) and are left alone.
- **Admin panel** (`/panel`): `views/panel.vue` gates on `UserService.IsLogin` / `AuthChecked` — shows `LoginComponent` when signed out, else `PageEditor.vue` (plain-textarea Markdown editor + debounced live preview via `POST /Preview`, rendered into a `.content` wrapper; fields include the page's route **Path**, **View type** — a native `<select>` offering `PlainHTML`/`List` (native rather than PrimeVue's `Dropdown` because the project loads no PrimeVue theme CSS, so multi-element PrimeVue widgets don't lay out — single-element `InputText`/`Textarea` are fine) — comma-separated **Tags** and, only when View type is `List`, **List tags** (parsed via `components/panelComponents/tags.ts`), **Card summary**/**Card image** overrides, and the menu **Caption**). The teleported side panel shows `PanelMenu.vue` (page list + "New page"). Cross-component state lives in `global/panelState.ts` (a reactive singleton: `pages`, `selected`, `dirty`, `refresh`/`select`/`startNew`). `service/Panel.service.ts` extends `serviceClass` (base path `/auth/ControlPanel`), attaches the Bearer token per call, and flips `IsLogin` to false on a 401.
- **State** (`global/store.ts`): Vuex store tracks responsive `ScreenLevel` (see `MediaEnum`) and a derived `GetIsMobile` getter (`ScreenLevel < 2`). `App.vue` dispatches `SetScreenLevel` on mount and window resize to drive the responsive layout. The active page is no longer store state — it is the URL (`route.path`); `MenuItem` navigates with `router.push(Path)` and highlights via `isMenuItemActive(item, route.path)` (`components/menuComponents/menuActive.ts`).
- **API layer**: `service/BaseAPI.service.ts` exports a `serviceClass` (axios wrapper, base URL `http://localhost:8080/api`) and a default singleton, with `getPageByPath(path)` (per-page routing) and the legacy `getPage(pageName)`. `App.vue` `provide()`s this singleton as `'Service'`; components `inject('Service')` it. `service/User.service.ts` extends `serviceClass` (base path `/auth`) and persists the JWT via `LocalStorage.service.ts`.
- **Path alias**: `@/` → `src/` (configured in both `vite.config.ts` and tsconfig).
- **Theme & styling** (theme-aware design-token system): Design tokens live as CSS custom properties in `src/assets/theme.css` — colors (`--bg`, `--surface`, `--surface-2`, `--fg`, `--muted`, `--accent`, `--border`) and type (`--serif` Fraunces, `--sans` Inter, `--mono` Fira Code). Three scopes define them: `:root` (light, the default), `html[data-theme="dark"]` (dark), and a `prefers-color-scheme: dark` block that supplies the dark values when no explicit `data-theme` is set. `tailwind.config.js` maps its color keys to these `var(--…)` tokens, so utility classes like `bg-surface`, `text-fg`, `text-accent`, `border-border` are automatically theme-aware; the legacy names (`midnightPurple` / `mainComponentBackground` → surface, `activePageColor` → accent, `deActivePageColor` → muted) are remapped to tokens for back-compat. Theme switching lives in `global/theme.ts` (`resolveInitialTheme` / `applyTheme` / `getTheme` / `toggleTheme`), which flips `data-theme` on `<html>` and persists the choice via `LocalStorage.service.ts` under the key `theme`; `components/ThemeToggle.vue` is the toggle UI, and a no-flash inline script in `index.html`'s `<head>` sets `data-theme` before first paint.
- **Styling the rendered Markdown**: the `v-html` output in `views/contents.vue` is styled exclusively through the `.content` wrapper in `src/assets/content.css` using descendant selectors — Tailwind utility classes do not reach that injected HTML. This is the "single definitive stylesheet for arbitrary `.md` content" goal from Purpose & design intent. Global stylesheets are imported in `main.ts` in order: `theme.css`, `main.css`, `sliderMenu.css`, `content.css`.

## Work in progress / incomplete areas

These are intentionally unfinished, not bugs — don't treat them as breakage:

- **Reserved-route views.** `/panel` is the admin panel; `/lists` is a reserved client-side route that still renders the `lists` view stub — a leftover placeholder, **unrelated** to the now-shipped `ViewType = "List"` per-page card grid (see Card rendering), which lives at each list page's own `Path`, not at `/lists`. Every other path is a content page resolved by `Path`. Path uniqueness is enforced panel-side (`422` on collision); there is no DB unique index yet, so a direct DB write could still create a duplicate `Path`.
- **`ViewType` render mode.** `PageModel.ViewType` is a panel combobox (a native `<select>`) with two shipped values: `PlainHTML` (default) and `List` (a tag-driven card grid — see Card rendering). The earlier `DynamicList` plan was dropped and folded entirely into `List`. The frontend still does **not** branch on `ViewType` — that's by design, not an unfinished hook: `CardService.ExpandCards` does all the mode-specific work backend-side and `/api/Page` returns final HTML for every mode, so `contents.vue` just injects it via `v-html`. The combobox is intentionally extensible — a future `ViewType` needing genuinely different frontend handling (not just more server-side HTML) would still require a branch — but only `PlainHTML`/`List` exist today.
