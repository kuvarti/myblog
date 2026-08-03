# Panel page visibility + page-driven navigation

**Date:** 2026-08-03
**Status:** Approved
**Scope:** README backlog task 3 — a per-page visibility checkbox in the panel menu that controls whether the page appears in the public side menu. Delivered together with a switch to a **page-driven** navigation model (the page becomes the single source of truth for nav membership, order, and visibility), plus a path-uniqueness constraint on the panel editor.

## Problem

Today the public side menu (`sideMenu.vue` → `GET /api/MenuList/Menu` → `MenuController.GetMenu`) is built from the `Menus` collection: it returns every `Menus` document, then sorts them by the owning page's `Order` (`sortMenusByOrder`, added in task 2). Consequences of the menu-driven model:

- A page appears in the nav **only if a matching `Menus` document exists** (created via the editor's menu binding). There is no simple per-page toggle.
- Three hand-seeded `Menus` documents have an empty `PageName` (Projelerim / İletişim / Blog Yazıları); they are nav stubs not tied to any page and currently always show.

Task 3 asks for a checkbox on each page row in the panel: checked → the page shows in the side menu. The owner chose to satisfy this by **inverting the source of truth** (page-driven nav, "option 3"), accepting that the three page-less stubs drop out of the nav.

## Decisions (approved)

1. **Storage: `Hidden bool` on the page (inverted), not `Visible`.** Legacy and re-imported page documents have no such field; bson decodes a missing bool to `false`, i.e. `Hidden:false` = visible. This keeps every existing page visible by default with **no migration/backfill**. The API and frontend speak in terms of `Visible` (= `!Hidden`); only the storage layer uses `Hidden`.
2. **Path conflicts return HTTP 422.** The existing duplicate-name conflict on create is `409`, which the frontend already renders as "name exists". Using `422 Unprocessable Entity` for a path collision keeps the two cases cleanly separable with a single new frontend branch.
3. **Path uniqueness reserves all non-empty paths, including stubs.** The check is "another menu document (any `PageName`, excluding the page being edited) already uses this `Path`". Dormant stub paths such as `/lists` are therefore reserved. This can be narrowed to page-bound menus later if desired.

## Navigation model (page-driven)

`GetMenu` is rewritten around a pure, unit-testable builder:

```
buildNav(pages []PageSummary, menus []*MenuModel) []*MenuModel
    byPage := index menus by PageName, skipping empty PageName
    nav := []
    for p in pages:              // List() already sorts by Order asc
        if !p.Visible: continue
        if m, ok := byPage[p.PageName]; ok:
            nav.append(m)         // page's own menu doc → carries Caption/Path
        else:
            nav.append(&MenuModel{Name: p.PageName, Caption: p.PageName,
                                  Path: "", PageName: p.PageName})
    return nav
```

- **Source of truth is `Pages`.** Nav membership = every page that is visible; order is inherited from `List()`'s existing `Order` sort; the label/path come from the page's own `Menus` document when present, otherwise fall back to `PageName`.
- **The old `sortMenusByOrder` is removed** — sorting is now implicit in `List()`. Its tests are replaced by `buildNav` tests.
- **Page-less stubs drop out** (they are absent from `byPage` and no page references them). Their `Menus` documents are left in place (not deleted); they simply stop appearing in the nav.
- `GetMenu` fetches all menus once (`MenuService.GetMenu()`) and the page list once (`PageService.List()`), then calls `buildNav` — no per-page query.

`List()` gains `Visible` so both the panel and `buildNav` can read it (see below).

## Backend changes

### Models
- `Models/Pages.Model.go` — `PageModel` gains `Hidden bool \`json:"-" bson:"Hidden"\``.
- `Models/Panel.Model.go`:
  - `PageSummary` gains `Hidden bool \`json:"-" bson:"Hidden"\`` and `Visible bool \`json:"Visible" bson:"-"\``. `List()` sets `Visible = !Hidden` after decode.
  - New `type VisibilityRequest struct { PageName string \`json:"PageName"\`; Visible bool \`json:"Visible"\` }`.

### Page service (`Services/Page.Service.go`)
- Interface gains `SetVisibility(name string, visible bool) error`.
- `List()` projection adds `{Key:"Hidden",Value:1}`; after `cursor.All`, loop `summaries[i].Visible = !summaries[i].Hidden`. (Sort stays `Order` asc.)
- `Create` inserts `{Key:"Hidden",Value:false}` so new pages default to visible (explicit, though a missing field would also read as visible).
- New:
  ```go
  func (psi *PageServiceImplementation) SetVisibility(name string, visible bool) error {
      res, err := psi.collection.UpdateOne(psi.ctx,
          bson.D{{Key: "PageName", Value: name}},
          bson.D{{Key: "$set", Value: bson.D{{Key: "Hidden", Value: !visible}}}})
      if err != nil { return err }
      if res.MatchedCount == 0 { return ErrPageNotFound }
      return nil
  }
  ```

### Menu service (`Services/Menu.Service.go`)
- Interface gains `PathTakenByOther(path, excludePageName string) (bool, error)`.
  ```go
  func (msi *MenuServiceImplementation) PathTakenByOther(path, excludePageName string) (bool, error) {
      if path == "" { return false, nil }
      count, err := msi.collection.CountDocuments(msi.ctx, bson.D{
          {Key: "Path", Value: path},
          {Key: "PageName", Value: bson.D{{Key: "$ne", Value: excludePageName}}},
      })
      return count > 0, err
  }
  ```
  Empty path is never a conflict (many pages may have no path).

### Menu controller (`Controllers/Menu.Controller.go`)
- Replace `sortMenusByOrder` with pure `buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel`.
- `GetMenu`: fetch `menus` and `pages`, return `buildNav(pages, menus)`. If `PageService.List()` errors, fall back to returning the raw menus (best-effort, matching the current tolerant behavior).

### Panel controller (`Controllers/Panel.Controller.go`)
- New route `cp.PUT("/PageVisibility", pc.SetPageVisibility)`; handler binds `VisibilityRequest`, calls `PageService.SetVisibility`, returns `200 {"message":"updated"}` (or `404` on `ErrPageNotFound`).
- **Path validation** in `CreatePage` and `UpdatePage`, performed **before** any write:
  - If `req.Menu != nil && req.Menu.Path != ""`, call `MenuService.PathTakenByOther(req.Menu.Path, <name>)` — `<name>` is `req.PageName` on create, the route `:name` on update.
  - On conflict: `ctx.JSON(422, gin.H{"error": "path already used by another page"})` and return before creating/updating.
  - On create, this check runs before `PageService.Create` so a rejected path never leaves an orphan page.

### Wiring
No composition-root change (`InitMenuController` already receives `PageService`).

## Frontend changes

### Types (`types/PanelModels.ts`)
`PanelPageSummary` gains `Visible: boolean`.

### Service (`service/Panel.service.ts`)
```ts
public async setVisibility(pageName: string, visible: boolean): Promise<void> {
    return this.apiClient.put('/PageVisibility', { PageName: pageName, Visible: visible }, this.authConfig())
        .then(() => undefined)
        .catch((e) => this.handleAuthError(e))
}
```

### Panel state (`global/panelState.ts`)
```ts
export async function setVisibility(name: string, visible: boolean): Promise<void> {
    const page = state.pages.find((p) => p.PageName === name)
    if (page) page.Visible = visible          // optimistic
    try {
        await PanelService.setVisibility(name, visible)
    } catch {
        await refresh()                        // resync on failure
    }
}
```

### Panel menu (`components/sidePanelComponents/PanelMenu.vue`)
Restructure each row from a `<button>` into a draggable `<div>` (a checkbox may not be nested inside a `<button>`):

```
<div draggable="true" @dragstart/@dragover.prevent/@drop.prevent/@dragend
     class="flex items-center gap-2 ... cursor-move" :class="{ ...selected/drag states... }">
    <v-icon name="hi-menu" class="text-muted shrink-0" aria-hidden="true" />
    <button class="flex-1 text-left truncate" @click="select(p.PageName)">{{ p.PageName }}</button>
    <input type="checkbox" :checked="p.Visible"
           @change="onToggle(p, $event)" @click.stop @mousedown.stop
           class="accent-accent shrink-0" :title="p.Visible ? 'Visible in menu' : 'Hidden from menu'" />
</div>
```

- Select stays a real `<button>` (keeps keyboard access); the checkbox and drag live on separate elements.
- `@click.stop`/`@mousedown.stop` on the checkbox stop it from selecting the row or starting a drag.
- `onToggle(p, e)` calls `setVisibility(p.PageName, (e.target as HTMLInputElement).checked)`.
- Drag/selected visual states from task 2 are preserved on the `<div>`.

### Editor (`components/panelComponents/PageEditor.vue`)
Add one branch to the existing `save()` error handling: `e?.response?.status === 422` → `error.value = 'That path is already used by another page.'` (the `409` name branch is unchanged).

## Testing

- **Go (`Controllers/Menu.Controller_test.go`, rewritten):** `buildNav` unit tests —
  - visible pages only (a `Visible:false` page is excluded);
  - page-less stub menus are dropped;
  - a page with no menu doc falls back to `Caption == PageName`;
  - a page with a menu doc keeps its `Caption`/`Path`;
  - order follows the (already-sorted) `pages` slice.
- **Manual E2E:** toggle a page's checkbox off → it disappears from the public side menu after `getMenu` refetch, and back on → reappears; the three stubs are gone; saving two pages with the same non-empty path is rejected with the 422 message; saving a page keeping its own unchanged path still succeeds.
- **Toggle feedback (added post-approval):** each checkbox toggle shows a small toast confirming the change and refreshes the public side menu live. See the plan's Task 5 — a reactive `notify` store + `AppToast`, and a shared `menuVersion` signal that `sideMenu.vue` watches to refetch `getMenu`.
- Path uniqueness and `SetVisibility` touch Mongo directly (the project has no service-layer test harness), so they are covered by the controller path plus manual E2E rather than unit tests, consistent with the existing codebase.

## Out of scope

- Deleting or migrating the dormant stub `Menus` documents (they remain, simply unshown).
- Per-page URL routing that would consume the now-reserved unique paths (a separate, known WIP slice).
- Reverting the optimistic visibility toggle field-by-field beyond the `refresh()` resync.
