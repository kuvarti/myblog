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
```
The server listens on `:8080` (Gin default). There is no test suite.

### Frontend (`cd frontend`)
```sh
npm install
npm run dev          # Vite dev server on :5173 (--host enabled)
npm run build        # type-check (vue-tsc) + production build
npm run type-check
npm run lint         # eslint --fix across the repo
npm run format       # prettier on src/
npm run test:unit    # vitest (jsdom); no test files exist yet
```
Run a single test once suites exist: `npm run test:unit -- <path-or-name-pattern>`.

## Configuration touchpoints

There is no `.env` or central config — every environment value is hardcoded. Changing an environment, port, or credential means editing all of the relevant spots below by hand (there is no single source of truth, so it is easy to miss one):

| Value | Where | Current |
|-------|-------|---------|
| Backend API base URL | `frontend/src/service/BaseAPI.service.ts` | `http://localhost:8080/api` |
| MongoDB URI + credentials | `backend/server.go` (`InitDB`) | `mongodb://root:example@localhost:27017` |
| Database name | `backend/server.go` (`main`) | `KuvartiBlog` |
| CORS allowed origin | `backend/server.go` (`InitServer`) | `http://localhost:5173` (listed twice) |
| JWT signing secret | `backend/Services/Token.Service.go` | `"SecretKey"` (TODO placeholder) |
| Mongo container creds / ports / mongo-express | `database/docker-compose.yml` | `root` / `example`, `27017`, UI on `8081` |

Ports that are **not** set explicitly (rely on framework defaults): backend Gin server on `:8080` (`server.Run()` with no argument) and Vite dev server on `:5173`. Note `InitServer` registers CORS twice — a restrictive `localhost:5173` config followed by a second `AllowAllOrigins = true` config that overrides it; treat CORS as effectively open in development.

## Backend architecture

Strict three-layer separation, wired manually in `server.go`:

- **Models/** — plain structs with `json`/`bson` tags (`UserModel`, `PageModel`, `MenuModel`, `JWTToken`, `LoginModel`).
- **Services/** — business logic + all MongoDB access. Each service is an interface (`FooService`) with an `Impl` struct constructed via `NewFooService(ctx, collection)`. Controllers depend on the interface, not the implementation.
- **Controllers/** — Gin handlers. Each has `InitFooController(service, apiGroup)` that registers routes on the shared `/api` router group.

**Naming convention.** Backend files are `<Resource>.<Layer>.go` (e.g. `Menu.Controller.go`, `Page.Service.go`, `User.Model.go`); each new resource adds one file per layer with the same prefix. Frontend service classes follow `<Name>.service.ts`.

`server.go` `InitControllers` is the composition root: it opens each MongoDB collection (`Users`, `Menus`, `Pages`) and injects it into the matching service, then hands the service to the controller. To add a resource, follow this Model → Service (interface + impl) → Controller → wire-in-`InitControllers` chain.

### Database collections
MongoDB is schemaless — document shape is defined by the Go structs (bson tags), not by the database. Field-by-field shapes live in the model files; this table only maps each collection to its authoritative model and owning service, plus the non-obvious bits:

| Collection | Shape (authoritative) | Owner service | Notes |
|-----------|-----------------------|---------------|-------|
| `Users` | `Models/User.Model.go` (`UserModel`) | `UserService` | No seed — create manually with a bcrypt-hashed password |
| `Menus` | `Models/Menu.Model.go` (`MenuModel`) | `MenuService` | Some seed rows omit `PageName`; DB docs have `_id`, which the model doesn't map |
| `Pages` | `Models/Pages.Model.go` (`PageModel`) | `PageService` | `Hash`/`Text` are machine-managed render cache, not hand-authored (see Page rendering below) |

### API endpoints
All routes are under the `/api` group. Auth column = requires `Authorization: Bearer <token>` header.

| Method | Path | Auth | Body / Query | Purpose |
|--------|------|------|--------------|---------|
| POST | `/api/auth/login` | — | `{ "userName", "passWord" }` | Verify credentials, return `{ Message, token }` |
| GET | `/api/auth/AmIAuth` | Bearer | — | Token check (returns `{ "Yes": "YES" }`) |
| GET | `/api/MenuList/Menu` | — | — | All menu items (`Menus` collection) |
| GET | `/api/Page` | — | `?PageName=<name>` | Rendered page HTML + `ViewType` |
| GET | `/api/auth/ControlPanel/Pages` | Bearer | — | List `[{PageName, ViewType}]` |
| GET | `/api/auth/ControlPanel/Pages/:name` | Bearer | — | Raw source (clean newlines) + menu binding for editing |
| POST | `/api/auth/ControlPanel/Pages` | Bearer | `{PageName, Source, ViewType, Menu?}` | Create page (+ upsert menu); `409` on duplicate |
| PUT | `/api/auth/ControlPanel/Pages/:name` | Bearer | `{Source, ViewType, Menu?}` | Update page (+ upsert menu); PageName immutable |
| DELETE | `/api/auth/ControlPanel/Pages/:name` | Bearer | — | Delete page (+ cascade its menu entry) |
| POST | `/api/auth/ControlPanel/Preview` | Bearer | `{Source}` | Stateless render → `{Html}` (no DB write) |

The `ControlPanel` endpoints power the admin panel (`Controllers/Panel.Controller.go`). They speak **clean newlines**; the bespoke `/n` line delimiter stays backend-only, converted with `services.ToStorage` / `services.FromStorage`.

### Auth flow
- `POST /api/auth/login` verifies credentials with bcrypt against the `Users` collection, then `TokenService.GenerateToken` issues an HS256 JWT (1-hour expiry).
- Protected routes use the `TokenService.AuthenticateJWT()` Gin middleware, which expects an `Authorization: Bearer <token>` header. `/api/auth/AmIAuth` sits behind it.
- `UserService` and `TokenService` are both returned from `NewUserService` (they share the request context).
- **Gotcha:** the JWT `secretKey` in `Services/Token.Service.go` is the literal string `"SecretKey"` (marked `TODO`). Do not treat auth as production-grade.

### Page rendering (`Services/Page.Service.go`)
A page document in the `Pages` collection has: `Page` (the hand-authored source), `Text` (the cached rendered HTML — don't edit by hand), `Hash` (SHA-1 of `Page`, the cache key), plus `PageName` and `ViewType`.

On `GetPage`, the service SHA-1-hashes the raw `Page` and compares it to the stored `Hash`; if the content changed (or `Text` is empty) it re-renders via `GetPageText` + `gomarkdown` and caches the new `Text`/`Hash` back into MongoDB, otherwise it returns the cached `Text`. So editing `Page` triggers a re-render on next fetch; editing `Text` directly is pointless.

**Source authoring format** (`Page` field) — this is a bespoke line-based mix of Markdown and HTML, not standard Markdown:
- Lines are separated by the literal token `/n` (slash-n), which `GetPageText` converts to real newlines before splitting; `/n/n` is a blank separator.
- A line **containing `<`** is emitted as **raw HTML, verbatim** — Markdown is not applied to it.
- A run of **consecutive lines with no `<`** is treated as one **Markdown block** and rendered together.
- **Blank lines are dropped** (they produce no output on their own).
- **Gotcha:** any `<` anywhere on a line switches that whole line to HTML passthrough, so a Markdown line like `a < b` or `List<T>` won't render as Markdown — keep `<` out of Markdown lines. Also, the final rendered HTML has all newlines stripped (`ReplaceAll(Text, "\n", "")`), and Markdown links open in a new tab (`HrefTargetBlank`).

**Authoring a new page:** the intended way is the **admin panel** (create/edit/delete + live preview, with a bound menu entry — see the `ControlPanel` endpoints and Frontend architecture). Under the hood a page is a `Pages` document with `PageName`, a `Page` string in the format above, and empty `Hash`/`Text` (both fill in on the first `GET /api/Page?PageName=<name>`), plus a matching `Menus` entry to surface it in the nav.

Editing this rendering/caching logic is the trickiest part of the backend.

## Frontend architecture

Vue 3 `<script setup>` + TypeScript, Vuex 4, Vue Router 4, PrimeVue, Tailwind (+ SCSS).

- **Two mount points** (`index.html`): `#sidemenu` and `#app`. `App.vue` `<teleport>`s the sliding side menu into `#sidemenu`, keeping it outside the routed content grid.
- **Routing** (`router/index.ts`): three routes — `/` → `contents`, `/lists` → `lists`, `/panel` → `panel`. View components are re-exported as a barrel in `router/routes.ts`.
- **Admin panel** (`/panel`): `views/panel.vue` gates on `UserService.IsLogin` / `AuthChecked` — shows `LoginComponent` when signed out, else `PageEditor.vue` (plain-textarea Markdown editor + debounced live preview via `POST /Preview`, rendered into a `.content` wrapper). The teleported side panel shows `PanelMenu.vue` (page list + "New page"). Cross-component state lives in `global/panelState.ts` (a reactive singleton: `pages`, `selected`, `dirty`, `refresh`/`select`/`startNew`). `service/Panel.service.ts` extends `serviceClass` (base path `/auth/ControlPanel`), attaches the Bearer token per call, and flips `IsLogin` to false on a 401.
- **State** (`global/store.ts`): Vuex store tracks responsive `ScreenLevel` (see `MediaEnum`), a derived `GetIsMobile` getter (`ScreenLevel < 2`), and `ActivePage`. `App.vue` dispatches `SetScreenLevel` on mount and window resize to drive the responsive layout. `ActivePage` names which backend page `contents.vue` fetches.
- **API layer**: `service/BaseAPI.service.ts` exports a `serviceClass` (axios wrapper, base URL `http://localhost:8080/api`) and a default singleton. `App.vue` `provide()`s this singleton as `'Service'`; components `inject('Service')` it. `service/User.service.ts` extends `serviceClass` (base path `/auth`) and persists the JWT via `LocalStorage.service.ts`.
- **Path alias**: `@/` → `src/` (configured in both `vite.config.ts` and tsconfig).
- **Theme & styling** (theme-aware design-token system): Design tokens live as CSS custom properties in `src/assets/theme.css` — colors (`--bg`, `--surface`, `--surface-2`, `--fg`, `--muted`, `--accent`, `--border`) and type (`--serif` Fraunces, `--sans` Inter, `--mono` Fira Code). Three scopes define them: `:root` (light, the default), `html[data-theme="dark"]` (dark), and a `prefers-color-scheme: dark` block that supplies the dark values when no explicit `data-theme` is set. `tailwind.config.js` maps its color keys to these `var(--…)` tokens, so utility classes like `bg-surface`, `text-fg`, `text-accent`, `border-border` are automatically theme-aware; the legacy names (`midnightPurple` / `mainComponentBackground` → surface, `activePageColor` → accent, `deActivePageColor` → muted) are remapped to tokens for back-compat. Theme switching lives in `global/theme.ts` (`resolveInitialTheme` / `applyTheme` / `getTheme` / `toggleTheme`), which flips `data-theme` on `<html>` and persists the choice via `LocalStorage.service.ts` under the key `theme`; `components/ThemeToggle.vue` is the toggle UI, and a no-flash inline script in `index.html`'s `<head>` sets `data-theme` before first paint.
- **Styling the rendered Markdown**: the `v-html` output in `views/contents.vue` is styled exclusively through the `.content` wrapper in `src/assets/content.css` using descendant selectors — Tailwind utility classes do not reach that injected HTML. This is the "single definitive stylesheet for arbitrary `.md` content" goal from Purpose & design intent. Global stylesheets are imported in `main.ts` in order: `theme.css`, `main.css`, `sliderMenu.css`, `content.css`.

## Work in progress / incomplete areas

These are intentionally unfinished, not bugs — don't treat them as breakage:

- **Routing coverage / per-page URLs.** Clicking a nav item now loads its page by `PageName` (`MenuItem` dispatches `SetActivePage`, `contents.vue` re-fetches reactively), but there are still only three routes (`/`, `/lists`, `/panel`). `ActivePage` lives in Vuex, not the URL, so a clicked page is not bookmarkable and refreshing `/` returns to the default page. Real per-page routing is a later slice.
- **`ViewType` render mode.** `PageModel.ViewType` is returned by `/api/Page` and seeded with values like `"PlainHTML"`. It is deliberately the per-page display-mode switch: the intent is that if a page ever needs to be shown differently from the default Markdown pipeline (e.g. served as plain HTML), that page carries a distinct `ViewType` and the frontend branches on it to pick the matching render mode. The frontend doesn't branch on it yet — every page renders the same way — so it is a designed-in hook, not dead code; wire the frontend to it when a page that needs a different display type actually appears.
