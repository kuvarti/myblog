# Per-page routing

**Date:** 2026-08-03
**Status:** Approved
**Scope:** Give every page its own URL so pages are bookmarkable/shareable and the menu navigates to a real route, while staying a single-page application. The page's `Path` becomes the route. This supersedes the forward-looking path-uniqueness added in the panel-page-visibility feature (that check moves from the menu binding to the page).

## Problem

Today the app is an SPA (Vue Router, `createWebHistory`) with only three routes: `/` (contents), `/lists`, `/panel`. Content pages have no URL of their own — they all render at `/`, distinguished by `ActivePage` held in Vuex (`SetActivePage` + `router.push('/')` in `MenuItem`). So a page is not bookmarkable, and refreshing `/` returns to the default page. The `Path` field that exists on the menu binding is only used to navigate the (now-removed) page-less stubs; for real pages it is dead.

We want each page to live at its own URL (e.g. `/about`), navigated to when its menu item is clicked, with no two pages sharing a URL — all without leaving the SPA model (client-side routing via the History API; only a server `index.html` fallback is needed for deep-links in production).

## Decisions (approved)

1. **The route key is the custom `Path`**, not `PageName`. The `Path` a user sets in the panel is the page's real URL.
2. **`Path` is mandatory** for every page. Saving without a valid path is rejected.
3. **`Path` is a first-class `PageModel` field** (required, unique). It moves off the menu binding. The `Menus` collection keeps only the optional display `Caption`. The uniqueness check moves from `MenuService` to `PageService`.
4. **Visibility is menu-presence, not access control.** A hidden page is absent from the menu but **still reachable by its direct URL** (the resolver keys on `Path`, ignoring `Hidden`).
5. **`/panel` and `/lists` are reserved** and cannot be assigned as a page path.
6. **Home `/` is just the page whose `Path` is `/`** (currently `MainPage`). No special-casing; `/` is an ordinary path a page owns.
7. **Production deep-link server config is out of scope** (documented note only); the Vite dev server already serves the SPA fallback.

## Data model

- `PageModel` gains `Path string` (`json:"Path" bson:"Path"`), required and unique across pages.
- `Menus` collection now holds only display metadata (`Caption`, and the legacy `Name`/`PageName`). Its `Path` is no longer written or read; `MenuModel` keeps the `Path` struct field because `buildNav` still emits `Path` in the nav DTO — but that value now comes from `page.Path`, not the menu document.
- Existing pages must be backfilled with a valid unique `Path` (see Migration).

## Backend

### Models
- `Models/Pages.Model.go` — add `Path string \`json:"Path" bson:"Path"\``.
- `Models/Panel.Model.go`:
  - `PageSummary` gains `Path string \`json:"Path" bson:"Path"\``.
  - `PageDetail` gains `Path string \`json:"Path"\`` (the editor reads it).
  - `MenuBinding` drops `Path` → `{ Name, Caption }` only.
  - `CreatePageRequest` and `UpdatePageRequest` gain `Path string \`json:"Path"\``.

### Page service (`Services/Page.Service.go`)
- Interface changes:
  - `Create(name, path, sourceClean, viewType string) error`
  - `Update(name, path, sourceClean, viewType string) error`
  - add `GetPageByPath(path string) (models.PageModel, error)`
  - add `PathTaken(path, excludePageName string) (bool, error)`
  - `List()` projects `Path`.
- `Create`/`Update` store `Path` alongside the other fields.
- **Refactor render/cache:** extract the hash/render/cache body of `GetPage(name)` into a helper that operates on an already-fetched `PageModel` and persists `Text`/`Hash`. `GetPage(name)` finds by `PageName` then calls it; new `GetPageByPath(path)` finds by `Path` then calls the same helper. Both return the rendered `PageModel` (newlines stripped as today).
  ```go
  func (psi *PageServiceImplementation) GetPageByPath(path string) (models.PageModel, error) {
      var page models.PageModel
      err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "Path", Value: path}}).Decode(&page)
      if err != nil { return models.PageModel{}, err }
      return psi.renderAndCache(page)
  }
  ```
- `PathTaken` mirrors the old menu check but on the pages collection:
  ```go
  func (psi *PageServiceImplementation) PathTaken(path, excludePageName string) (bool, error) {
      count, err := psi.collection.CountDocuments(psi.ctx, bson.D{
          {Key: "Path", Value: path},
          {Key: "PageName", Value: bson.D{{Key: "$ne", Value: excludePageName}}},
      })
      return count > 0, err
  }
  ```

### Menu service (`Services/Menu.Service.go`)
- Remove `PathTakenByOther` (moved to `PageService`).
- `Upsert` no longer sets `Path` (writes `Name`, `Caption`, `PageName` only). Callers stop passing a menu path.

### Menu controller (`Controllers/Menu.Controller.go`)
- `buildNav` now sources each nav item's `Path` from `page.Path` and its `Caption` from the page's menu document (fallback to `PageName`):
  ```go
  func buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel {
      capByPage := make(map[string]string, len(menus))
      for _, m := range menus {
          if m.PageName != "" { capByPage[m.PageName] = m.Caption }
      }
      nav := make([]*models.MenuModel, 0, len(pages))
      for _, p := range pages {
          if !p.Visible { continue }
          caption := p.PageName
          if c, ok := capByPage[p.PageName]; ok && c != "" { caption = c }
          nav = append(nav, &models.MenuModel{
              Name: p.PageName, Caption: caption, Path: p.Path, PageName: p.PageName,
          })
      }
      return nav
  }
  ```

### Page controller (`Controllers/Page.Controller.go`)
- `GET /api/Page` accepts `?Path=<path>` (preferred) as well as the existing `?PageName=<name>`. If `Path` is present, resolve via `GetPageByPath`; else fall back to `PageName`. Response DTO is unchanged (`{ViewType, Page}` where `Page` is the rendered `Text`). Missing/blank query → 400.

### Panel controller (`Controllers/Panel.Controller.go`)
- **Path validation** helper used by `CreatePage` and `UpdatePage`, run before any write:
  - empty or not starting with `/` → `400 {"error":"path must start with /"}`.
  - equals a reserved route (`/panel`, `/lists`) → `422 {"error":"path is reserved"}`.
  - already used by another page (`PageService.PathTaken(path, <name>)`) → `422 {"error":"path already used by another page"}`.
- `CreatePage`: bind `Path`, validate, `PageService.Create(PageName, Path, Source, ViewType)`; menu upsert now carries only `Caption`/`Name`.
- `UpdatePage`: bind `Path`, validate (exclude own name), `PageService.Update(name, Path, Source, ViewType)`; menu upsert `Caption`/`Name`.
- `GetPage` detail returns `Path` (from the page) and `Menu.Caption`.
- Remove the old `MenuService.PathTakenByOther` call.

## Frontend

### Router (`router/index.ts`)
Keep the static routes and add a catch-all that renders `contents`; static routes take precedence:
```ts
routes: [
  { path: '/panel', name: 'panel', component: () => Promise.resolve(routes.panel) },
  { path: '/lists', name: 'lists', component: () => Promise.resolve(routes.lists) },
  { path: '/:pathMatch(.*)*', name: 'content', component: () => Promise.resolve(routes.contents) },
]
```
`/` and `/about` both match the catch-all → `contents`, which fetches by `route.path`.

### Contents (`views/contents.vue`)
Fetch by the current route path instead of `ActivePage`:
```ts
import { useRoute } from 'vue-router'
const route = useRoute()
function load() { service?.getPageByPath(route.path).then(d => returnedHTML.value = d.Page).catch(() => returnedHTML.value = 'Error') }
onMounted(load)
watch(() => route.path, load)
```

### Service (`service/BaseAPI.service.ts`)
Add `getPageByPath(path)` → `GET Page?Path=<encoded path>`, same response handling as `getPage` (the `/n`→newline replace stays).

### Menu item (`components/menuComponents/MenuItem.vue`)
Every nav item is a page with a `Path`; navigate straight to it (drop the `SetActivePage`/`'/'` branch, stubs are gone):
```ts
const go = () => router.push(props.Path || '/')
```

### Active highlight (`components/menuComponents/menuActive.ts`)
Simplify to a path comparison (no `ActivePage`):
```ts
export function isMenuItemActive(item: { Path?: string }, routePath: string): boolean {
    return (item.Path || '/') === routePath
}
```
Update `MenuItem.vue` and the spec's tests accordingly.

### Vuex (`global/store.ts`) and `App.vue`
- Remove `ActivePage` state, `GetActivePage` getter, `ActivePageMutation`, and `SetActivePage` action — `route.path` is now the source of truth.
- `App.vue` drops the `SetActivePage(HOME_PAGE)` dispatch; `HOME_PAGE` import removed if otherwise unused.

### Panel types & service
- `types/PanelModels.ts` — `PanelPageSummary` and `PanelPageDetail` gain `Path: string`; `MenuBinding` drops `Path`.
- `service/Panel.service.ts` — `SavePagePayload` gains `Path`; `createPage`/`updatePage` send it.

### Editor (`components/panelComponents/PageEditor.vue`)
- Relabel the "Menu path" field to **"Page path"** and bind it to the page's `Path` (a page field), not the menu binding. It is required.
- Keep "Menu caption" bound to the menu binding.
- `save()` sends `Path`. Error branches: `409` → name exists; `422` → path reserved or already used (message: "That path is reserved or already used."); `400` → "Path must start with /".

## Migration

`Path` is now required and unique, but existing page documents have none. Backfill once (dev, via `mongosh` or a throwaway script) — using the page's existing menu `Path` when set, otherwise a slug of `PageName`, ensuring uniqueness and a leading `/`:

| Page | Path |
|------|------|
| MainPage | `/` |
| StyleTest | `/styletest` |
| SoLong | `/solong` |
| 42MainWorks | `/42mainworks` |

(The owner can rename any of these in the panel afterward.) The exact commands live in the implementation plan.

## Testing

- **Go unit:** `buildNav` tests updated for the new signature (nav `Path` = page `Path`, `Caption` fallback to `PageName`). Add a small pure test for the reserved/format path check if it is extracted as a pure helper.
- **Frontend unit:** update `menuActive.spec.ts` for the simplified `isMenuItemActive(item, routePath)`.
- **Manual E2E:** visiting `/about` (a page's path) renders that page; the URL is bookmarkable and survives refresh (Vite fallback); clicking a menu item navigates to its path and highlights it; a hidden page is absent from the menu but its direct URL still renders; saving two pages with the same path, or a path of `/panel`/`/lists`, or an empty path, is rejected with the right message; `/` renders the page whose path is `/`.

## Out of scope

- Production server SPA fallback (serve `index.html` for unknown paths) — a deploy concern; note it in the README/CLAUDE.md when deployment is set up.
- Moving `Caption` onto the page or removing the `Menus` collection entirely (kept for the display label).
- Multi-segment / nested paths beyond a single leading-slash segment (the catch-all accepts them, but the panel/UX assumes simple `/name` paths).
- Redirect from an old path when a page's `Path` changes.
