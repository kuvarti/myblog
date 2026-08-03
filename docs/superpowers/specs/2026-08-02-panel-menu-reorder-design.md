# Drag-to-reorder pages in the admin panel

**Date:** 2026-08-02
**Status:** Approved design, ready for implementation
**Area:** `backend` (Pages/Menus) + `frontend` (panel menu)

## Problem

The panel's page list (`PanelMenu.vue`) and the public side menu render in
arbitrary order — neither the `Pages` nor the `Menus` collection is sorted, and
there is no order field. The owner wants to drag pages up/down in the panel to
control the order they appear in the public side menu, and have that order
persist.

## Decision: order lives on the page (page-centric)

`Order` is stored on the **page** (`Pages` collection) — a single source of
truth. The panel lists all pages sorted by `Order`; dragging rewrites `Order`;
the public side menu renders its menu-bound pages in that page order. No
denormalized copy of `Order` on the menu entry.

Rationale: the panel is the CMS surface for all pages, and the next task (task 3,
a "show in side menu" checkbox) maps cleanly onto this — a menu entry means
"visible", and the page's `Order` already gives its position.

## Backend

### Model
- `PageModel` gains `Order int` (`bson:"Order" json:"Order"`), default `0`.
- `PageSummary` gains `Order int` so the panel list carries it.
- New request model `ReorderRequest { PageNames []string }`.

### Page service (`Page.Service.go`)
- `List()` — add `Order` to the projection and `SetSort({Order: 1})`, so the
  panel list comes back in order.
- `SetOrder(names []string) error` — for each name at index `i`, set that page's
  `Order = i` (one `UpdateOne` per name, matched by `PageName`). Names not found
  are skipped.
- `Create(...)` — assign the new page `Order = <current page count>` so a new
  page appends at the end instead of jumping to the top (default `0`).

### Reorder endpoint (`Panel.Controller.go`)
- `PUT /api/auth/ControlPanel/PageOrder` (JWT-protected, in the ControlPanel
  group) — body `{ "PageNames": ["a","b",...] }` → `PageService.SetOrder`.
- A separate top segment (`/PageOrder`), **not** `/Pages/...`, to avoid colliding
  with the existing `/Pages/:name` param route in Gin.

### Public nav sorting (`Menu.Controller.go`, single source of truth)
The public nav (`GET /api/MenuList/Menu`) reads the `Menus` collection but must
order by `page.Order`. Inject `PageService` into `MenuController` (both services
already exist in `server.go`'s `InitControllers`). `GetMenu` then:
1. `menus = MenuService.GetMenu()`
2. `orders = PageService.List()` → `map[PageName]Order`
3. `sortMenusByOrder(menus, orderMap)` → sorted result.

`sortMenusByOrder(menus, orderMap)` is a **pure** function (stable sort by the
looked-up order; entries whose page is missing from the map sort last, keeping
their relative order) — unit-tested in Go without a DB.

`server.go` passes `pageService` to `InitMenuController`.

## Frontend

- `PanelPageSummary` gains `Order: number`.
- `Panel.service.ts` — `reorderPages(pageNames: string[])` → `PUT /PageOrder`
  with the bearer token, same error handling as the other panel calls.
- `menuReorder.ts` (new) — pure `moveItem(arr, from, to)` returning a new array
  with the element moved; unit-tested.
- `panelState.ts` — `reorder(from, to)`: optimistically reorder `state.pages`
  via `moveItem`, then persist with `PanelService.reorderPages(pages.map(PageName))`.
- `PanelMenu.vue` — list rows become `draggable`. Native HTML5 DnD
  (`dragstart` records the source index; `dragover` `preventDefault` to allow
  drop + shows a simple drop indicator; `drop` computes the target index and
  calls `reorder`). No third-party library.

## Testing

- **Go:** `sortMenusByOrder` — ordering by the map, stable ties, and
  missing-page-sorts-last.
- **Vitest:** `moveItem` — move down, move up, no-op (same index), and bounds.
- **DB-backed paths** (`SetOrder`, `List` sort, `Create` append) and the DnD
  wiring — verified against the running server + in the browser (jsdom does no
  layout/DnD; the codebase verifies these paths via E2E, matching `scrollSync`
  and `menuActive`).

## Files

Backend: `Models/Pages.Model.go`, `Models/Panel.Model.go`, `Services/Page.Service.go`,
`Controllers/Panel.Controller.go`, `Controllers/Menu.Controller.go`, `server.go`,
and a Go test for `sortMenusByOrder`.

Frontend: `types/PanelModels.ts`, `service/Panel.service.ts`, `global/panelState.ts`,
`components/sidePanelComponents/PanelMenu.vue`, new `components/sidePanelComponents/menuReorder.ts`
+ its `.spec.ts`.

## Known limitations (accepted)

- The teleported public side menu in the *same* panel session does not live-update
  after a reorder — it refetches on reload. The panel does not display the public
  nav, and the live site reflects the new order on next load.
- New pages created after a reorder get `Order = count` (appended); fine for MVP.
- Reordering a page with no menu entry sets its `Order` but has no visible nav
  effect until it gains a menu entry (task 3).

## Out of scope

- The task-3 visibility checkbox, the task-4 ViewType combobox.
- A drag-and-drop library / animated reordering (native DnD only).
- Live cross-component nav refresh after reorder.
