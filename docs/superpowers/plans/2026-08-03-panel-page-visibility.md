# Panel Page Visibility + Page-Driven Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a per-page visibility checkbox to the panel menu that controls whether a page appears in the public side menu, by making the page the single source of truth for nav membership/order/visibility, plus a path-uniqueness rule in the panel editor.

**Architecture:** Store an inverted `Hidden` flag on the page (missing field = visible, so no migration). Rewrite `GetMenu` to build the nav from the page list (visible, Order-sorted) joined to menu documents for display captions/paths; page-less stub menus drop out. Toggle visibility via a small `PUT /PageVisibility` endpoint, mirroring the existing `PUT /PageOrder`. Reject duplicate non-empty paths on create/update with HTTP 422.

**Tech Stack:** Go 1.x + Gin + MongoDB (backend, `go run .` — no hot reload), Vue 3 `<script setup>` + TypeScript + Vite + PrimeVue (frontend), Vitest (frontend unit), `go test` (backend unit).

**Spec:** `docs/superpowers/specs/2026-08-03-panel-page-visibility-design.md`

## Global Constraints

- **Commit discipline (overrides the skill's per-task commit step):** Do NOT commit after individual tasks. Commits are owner-initiated and made only when the whole feature is complete. Each task below ends with a **green checkpoint** (tests/build pass), not a commit. A single feature commit happens in the Completion section, and only after the owner approves. The uncommitted spec file (`docs/superpowers/specs/2026-08-03-panel-page-visibility-design.md`, already staged) is part of that final commit.
- **Documentation is English only** (proper nouns excepted).
- **Backend runs via `go run .`, not air** — after any Go change, restart the backend before manual/E2E verification (Go changes are not hot-reloaded).
- **Inverted storage:** the page stores `Hidden bool`; the API and frontend speak `Visible` (= `!Hidden`). A missing `Hidden` field decodes to `false` = visible.
- **Auth for manual verification:** obtain a token from the throwaway dev account, then send it as a Bearer header:
  ```bash
  TOKEN=$(curl -s -X POST localhost:8080/api/auth/login -H 'Content-Type: application/json' \
    -d '{"userName":"testadmin","passWord":"testpass123"}' | jq -r .token)
  ```

---

### Task 1: Visibility data model + toggle endpoint (backend)

Adds the `Hidden` field, surfaces `Visible` in the page list, and exposes `PUT /PageVisibility`. This task is Mongo-touching, so it is verified by build + curl rather than a unit test (the codebase has no service-layer test harness).

**Files:**
- Modify: `backend/Models/Pages.Model.go`
- Modify: `backend/Models/Panel.Model.go`
- Modify: `backend/Services/Page.Service.go`
- Modify: `backend/Controllers/Panel.Controller.go`

**Interfaces:**
- Produces:
  - `models.PageModel.Hidden bool` (bson `Hidden`, json `-`)
  - `models.PageSummary.Hidden bool` (bson `Hidden`, json `-`) and `models.PageSummary.Visible bool` (bson `-`, json `Visible`)
  - `models.VisibilityRequest{ PageName string; Visible bool }`
  - `services.PageService.SetVisibility(name string, visible bool) error`
  - `List()` return values now have `Visible` populated (= `!Hidden`)
  - Route `PUT /api/auth/ControlPanel/PageVisibility`

- [ ] **Step 1: Add `Hidden` to `PageModel`**

In `backend/Models/Pages.Model.go`, add the field to the struct:

```go
type PageModel struct {
	PageName	string	`json:"PageName" gorm:"unique"`
	Page		string	`json:"Page"`
	Hash		[]byte	`json:"Hash"`
	Text		string	`json:"Text"`
	ViewType	string	`json:"ViewType"`
	Order		int	`json:"Order"`
	Hidden		bool	`json:"-" bson:"Hidden"`
}
```

- [ ] **Step 2: Add `Hidden`/`Visible` to `PageSummary` and add `VisibilityRequest`**

In `backend/Models/Panel.Model.go`, update `PageSummary` and append the request type:

```go
type PageSummary struct {
	PageName string `json:"PageName" bson:"PageName"`
	ViewType string `json:"ViewType" bson:"ViewType"`
	Order    int    `json:"Order" bson:"Order"`
	Hidden   bool   `json:"-" bson:"Hidden"`
	Visible  bool   `json:"Visible" bson:"-"`
}
```

```go
type VisibilityRequest struct {
	PageName string `json:"PageName"`
	Visible  bool   `json:"Visible"`
}
```

- [ ] **Step 3: Project `Hidden` and invert to `Visible` in `List()`**

In `backend/Services/Page.Service.go`, replace the `List()` body's projection and add the inversion loop:

```go
func (psi *PageServiceImplementation) List() ([]models.PageSummary, error) {
	opts := options.Find().SetProjection(bson.D{
		{Key: "PageName", Value: 1},
		{Key: "ViewType", Value: 1},
		{Key: "Order", Value: 1},
		{Key: "Hidden", Value: 1},
	}).SetSort(bson.D{{Key: "Order", Value: 1}})
	cursor, err := psi.collection.Find(psi.ctx, bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var summaries []models.PageSummary
	if err := cursor.All(psi.ctx, &summaries); err != nil {
		return nil, err
	}
	for i := range summaries {
		summaries[i].Visible = !summaries[i].Hidden
	}
	return summaries, nil
}
```

- [ ] **Step 4: Default new pages to visible in `Create`**

In `backend/Services/Page.Service.go`, add `Hidden:false` to the inserted document in `Create` (right after the `Order` entry):

```go
	_, err = psi.collection.InsertOne(psi.ctx, bson.D{
		{Key: "PageName", Value: name},
		{Key: "Page", Value: ToStorage(sourceClean)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: viewType},
		{Key: "Order", Value: total},
		{Key: "Hidden", Value: false},
	})
```

- [ ] **Step 5: Add `SetVisibility` to the interface and implementation**

In `backend/Services/Page.Service.go`, add to the `PageService` interface (after `SetOrder`):

```go
	SetVisibility(name string, visible bool) error
```

And add the method (place it next to `SetOrder`):

```go
// SetVisibility flips a page's Hidden flag. Storage is inverted (Hidden), so a
// visible page is Hidden:false; missing field also reads as visible.
func (psi *PageServiceImplementation) SetVisibility(name string, visible bool) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "Hidden", Value: !visible}}}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPageNotFound
	}
	return nil
}
```

- [ ] **Step 6: Add the `PUT /PageVisibility` route and handler**

In `backend/Controllers/Panel.Controller.go`, register the route inside the `{ ... }` block (after the `PageOrder` line):

```go
		cp.PUT("/PageVisibility", pc.SetPageVisibility)
```

And add the handler (place it next to `ReorderPages`):

```go
func (pc *PanelController) SetPageVisibility(ctx *gin.Context) {
	var req models.VisibilityRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := pc.PageService.SetVisibility(req.PageName, req.Visible); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}
```

(`errors` and `services` are already imported in this file.)

- [ ] **Step 7: Build**

Run: `cd backend && go build ./...`
Expected: no output (clean build). Fix any compile errors before continuing.

- [ ] **Step 8: Restart backend and verify the toggle end-to-end**

Restart the backend (kill the running `go run .` and start it again), then:

```bash
TOKEN=$(curl -s -X POST localhost:8080/api/auth/login -H 'Content-Type: application/json' \
  -d '{"userName":"testadmin","passWord":"testpass123"}' | jq -r .token)
# List shows Visible for every page (all true initially):
curl -s localhost:8080/api/auth/ControlPanel/Pages -H "Authorization: Bearer $TOKEN" | jq '.[] | {PageName, Visible}'
# Hide MainPage, confirm Visible flips to false, then restore:
curl -s -X PUT localhost:8080/api/auth/ControlPanel/PageVisibility -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"PageName":"MainPage","Visible":false}'
curl -s localhost:8080/api/auth/ControlPanel/Pages -H "Authorization: Bearer $TOKEN" | jq '.[] | select(.PageName=="MainPage") | {PageName, Visible}'
curl -s -X PUT localhost:8080/api/auth/ControlPanel/PageVisibility -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"PageName":"MainPage","Visible":true}'
```
Expected: the first list shows `Visible: true` for all; after hiding, MainPage shows `Visible: false`; the restore returns it to true. **Green checkpoint — do not commit.**

---

### Task 2: Page-driven navigation via `buildNav` (backend, TDD)

Replaces the `sortMenusByOrder` approach with a pure `buildNav` that generates the nav from the page list joined to menus, and rewrites `GetMenu` to use it. This is the unit-tested core of the feature.

**Files:**
- Modify: `backend/Controllers/Menu.Controller.go`
- Test: `backend/Controllers/Menu.Controller_test.go` (replace existing `sortMenusByOrder` tests)

**Interfaces:**
- Consumes: `models.PageSummary.Visible` (Task 1), `models.MenuModel`, `services.PageService.List()`, `services.MenuService.GetMenu()`
- Produces: `buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel`

- [ ] **Step 1: Replace the test file with `buildNav` tests**

Overwrite `backend/Controllers/Menu.Controller_test.go` with:

```go
package controllers

import (
	models "backend/Models"
	"testing"
)

func TestBuildNavFiltersHidden(t *testing.T) {
	pages := []models.PageSummary{
		{PageName: "A", Visible: true},
		{PageName: "B", Visible: false},
		{PageName: "C", Visible: true},
	}
	nav := buildNav(pages, nil)
	if len(nav) != 2 {
		t.Fatalf("expected 2 visible entries, got %d", len(nav))
	}
	if nav[0].PageName != "A" || nav[1].PageName != "C" {
		t.Fatalf("expected A,C got %s,%s", nav[0].PageName, nav[1].PageName)
	}
}

func TestBuildNavDropsStubs(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Visible: true}}
	menus := []*models.MenuModel{
		{PageName: "A", Caption: "Alpha"},
		{PageName: "", Caption: "Stub", Path: "/lists"}, // page-less seed stub
	}
	nav := buildNav(pages, menus)
	if len(nav) != 1 || nav[0].PageName != "A" {
		t.Fatalf("expected only page A, got %d entries", len(nav))
	}
}

func TestBuildNavCaptionFallback(t *testing.T) {
	pages := []models.PageSummary{{PageName: "Solo", Visible: true}}
	nav := buildNav(pages, nil)
	if len(nav) != 1 || nav[0].Caption != "Solo" {
		t.Fatalf("expected caption fallback 'Solo', got %+v", nav)
	}
}

func TestBuildNavKeepsMenuCaption(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Visible: true}}
	menus := []*models.MenuModel{{PageName: "A", Caption: "Alpha", Path: "/a"}}
	nav := buildNav(pages, menus)
	if nav[0].Caption != "Alpha" || nav[0].Path != "/a" {
		t.Fatalf("expected menu caption/path, got %+v", nav[0])
	}
}

func TestBuildNavPreservesOrder(t *testing.T) {
	pages := []models.PageSummary{
		{PageName: "First", Visible: true},
		{PageName: "Second", Visible: true},
		{PageName: "Third", Visible: true},
	}
	nav := buildNav(pages, nil)
	got := []string{nav[0].PageName, nav[1].PageName, nav[2].PageName}
	want := []string{"First", "Second", "Third"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v", i, got, want)
		}
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./Controllers/`
Expected: compile failure / FAIL — `buildNav` is undefined.

- [ ] **Step 3: Implement `buildNav` and rewrite `GetMenu`**

In `backend/Controllers/Menu.Controller.go`: remove the `"sort"` import, delete `sortMenusByOrder`, rewrite `GetMenu`, and add `buildNav`. The full file becomes:

```go
package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MenuController struct {
	MenuService services.MenuService
	PageService services.PageService
}

func InitMenuController(MenuService services.MenuService, PageService services.PageService, server *gin.RouterGroup) MenuController {
	group := server.Group("/MenuList")
	mc := MenuController{
		MenuService: MenuService,
		PageService: PageService,
	}
	group.GET("/Menu", mc.GetMenu)
	return mc
}

func (mc *MenuController) GetMenu(ctx *gin.Context) {
	menus, err := mc.MenuService.GetMenu()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	// The page is the source of truth for nav membership/order/visibility; the
	// menu docs only supply display captions/paths. If the page list is
	// unavailable, fall back to returning the raw menus.
	pages, err := mc.PageService.List()
	if err != nil {
		ctx.JSON(http.StatusOK, menus)
		return
	}
	ctx.JSON(http.StatusOK, buildNav(pages, menus))
}

// buildNav produces the public navigation from the pages (the source of truth
// for membership, order, and visibility) joined to their menu documents for
// display captions/paths. Pages arrive already Order-sorted from
// PageService.List(). A visible page with no menu document falls back to its
// PageName as the caption; menu documents with no matching page (hand-seeded
// stubs) are dropped.
func buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel {
	byPage := make(map[string]*models.MenuModel, len(menus))
	for _, m := range menus {
		if m.PageName != "" {
			byPage[m.PageName] = m
		}
	}
	nav := make([]*models.MenuModel, 0, len(pages))
	for _, p := range pages {
		if !p.Visible {
			continue
		}
		if m, ok := byPage[p.PageName]; ok {
			nav = append(nav, m)
		} else {
			nav = append(nav, &models.MenuModel{
				Name:     p.PageName,
				Caption:  p.PageName,
				Path:     "",
				PageName: p.PageName,
			})
		}
	}
	return nav
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./Controllers/`
Expected: PASS (5 tests).

- [ ] **Step 5: Restart backend and verify the nav**

Restart the backend, then:
```bash
curl -s localhost:8080/api/MenuList/Menu | jq '.[] | {Caption, PageName}'
```
Expected: only pages appear (each with a `PageName`); the three page-less stubs (Projelerim / İletişim / Blog Yazıları) are gone. Hide a page (Task 1 curl) and re-run — it disappears; restore it — it returns. **Green checkpoint — do not commit.**

---

### Task 3: Path uniqueness validation (backend)

Rejects creating/updating a page whose non-empty menu path is already used by another menu document, with HTTP 422. Mongo-touching → verified by build + curl.

**Files:**
- Modify: `backend/Services/Menu.Service.go`
- Modify: `backend/Controllers/Panel.Controller.go`

**Interfaces:**
- Produces: `services.MenuService.PathTakenByOther(path, excludePageName string) (bool, error)`
- Consumes: existing `CreatePage` / `UpdatePage` handlers

- [ ] **Step 1: Add `PathTakenByOther` to the interface and implementation**

In `backend/Services/Menu.Service.go`, add to the `MenuService` interface:

```go
	PathTakenByOther(path, excludePageName string) (bool, error)
```

And add the method:

```go
// PathTakenByOther reports whether some other menu document already uses this
// path. Empty paths never conflict (many pages may have no path). The current
// page is excluded by PageName so re-saving an unchanged path is allowed.
func (msi *MenuServiceImplementation) PathTakenByOther(path, excludePageName string) (bool, error) {
	if path == "" {
		return false, nil
	}
	count, err := msi.collection.CountDocuments(msi.ctx, bson.D{
		{Key: "Path", Value: path},
		{Key: "PageName", Value: bson.D{{Key: "$ne", Value: excludePageName}}},
	})
	return count > 0, err
}
```

- [ ] **Step 2: Enforce uniqueness in `CreatePage` (before creating)**

In `backend/Controllers/Panel.Controller.go`, in `CreatePage`, insert this block **after** the `PageName`/`Source` required check and **before** `pc.PageService.Create(...)`:

```go
	if req.Menu != nil && req.Menu.Path != "" {
		taken, err := pc.MenuService.PathTakenByOther(req.Menu.Path, req.PageName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if taken {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
			return
		}
	}
```

- [ ] **Step 3: Enforce uniqueness in `UpdatePage` (before updating)**

In `backend/Controllers/Panel.Controller.go`, in `UpdatePage`, insert this block **after** the `Source` required check and **before** `pc.PageService.Update(...)` (note it excludes the page's own name via the route param `name`):

```go
	if req.Menu != nil && req.Menu.Path != "" {
		taken, err := pc.MenuService.PathTakenByOther(req.Menu.Path, name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if taken {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
			return
		}
	}
```

- [ ] **Step 4: Build**

Run: `cd backend && go build ./...`
Expected: clean build.

- [ ] **Step 5: Restart backend and verify path rejection**

Restart the backend, then (reuse `$TOKEN` from the Global Constraints helper):
```bash
# Create a scratch page taking a path:
curl -s -o /dev/null -w "%{http_code}\n" -X POST localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"PathA","Source":"a","ViewType":"","Menu":{"Name":"","Caption":"A","Path":"/dup"}}'   # 201
# A second page with the SAME path must be rejected 422:
curl -s -o /dev/null -w "%{http_code}\n" -X POST localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"PathB","Source":"b","ViewType":"","Menu":{"Name":"","Caption":"B","Path":"/dup"}}'   # 422
# Re-saving PathA with its own unchanged path still succeeds (excludes self):
curl -s -o /dev/null -w "%{http_code}\n" -X PUT localhost:8080/api/auth/ControlPanel/Pages/PathA \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"Source":"a2","ViewType":"","Menu":{"Name":"","Caption":"A","Path":"/dup"}}'   # 200
# Cleanup:
curl -s -o /dev/null -X DELETE localhost:8080/api/auth/ControlPanel/Pages/PathA -H "Authorization: Bearer $TOKEN"
```
Expected status codes: `201`, `422`, `200`. **Green checkpoint — do not commit.**

---

### Task 4: Frontend — visibility checkbox + path-conflict error

Wires the panel list checkbox to the toggle endpoint (page-driven, optimistic) and adds the 422 error message in the editor.

**Files:**
- Modify: `frontend/src/types/PanelModels.ts`
- Modify: `frontend/src/service/Panel.service.ts`
- Modify: `frontend/src/global/panelState.ts`
- Modify: `frontend/src/components/sidePanelComponents/PanelMenu.vue`
- Modify: `frontend/src/components/panelComponents/PageEditor.vue`

**Interfaces:**
- Consumes: `PUT /PageVisibility` (Task 1), `models.PageSummary.Visible` shape (arrives as JSON `Visible`)
- Produces: `PanelService.setVisibility(pageName, visible)`, `panelState.setVisibility(name, visible)`

- [ ] **Step 1: Add `Visible` to the summary type**

In `frontend/src/types/PanelModels.ts`, add to `PanelPageSummary`:

```ts
export interface PanelPageSummary {
	PageName: string
	ViewType: string
	Order: number
	Visible: boolean
}
```

- [ ] **Step 2: Add `setVisibility` to the panel service**

In `frontend/src/service/Panel.service.ts`, add this method (next to `reorderPages`):

```ts
	public async setVisibility(pageName: string, visible: boolean): Promise<void> {
		return this.apiClient.put('/PageVisibility', { PageName: pageName, Visible: visible }, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
```

- [ ] **Step 3: Add `setVisibility` to panel state (optimistic + resync)**

In `frontend/src/global/panelState.ts`, add (next to `reorder`):

```ts
export async function setVisibility(name: string, visible: boolean): Promise<void> {
	const page = state.pages.find((p) => p.PageName === name)
	if (page) page.Visible = visible // optimistic
	try {
		await PanelService.setVisibility(name, visible)
	} catch {
		await refresh() // resync on failure
	}
}
```

- [ ] **Step 4: Restructure the panel row and add the checkbox**

In `frontend/src/components/sidePanelComponents/PanelMenu.vue`, replace the `<button v-for ...>...</button>` row with a draggable `<div>` (a checkbox may not be nested in a `<button>`):

```html
				<div v-for="(p, i) in state.pages" :key="p.PageName"
					draggable="true"
					@dragstart="onDragStart(i)"
					@dragover.prevent="onDragOver(i)"
					@drop.prevent="onDrop(i)"
					@dragend="onDragEnd"
					class="flex items-center gap-2 w-full px-3 py-2 rounded hover:bg-surface-2 text-fg cursor-move"
					:class="{
						'bg-surface-2 font-semibold': state.selected?.PageName === p.PageName,
						'opacity-40': dragIndex === i,
						'border-t-2 border-accent': overIndex === i && dragIndex !== null && dragIndex !== i,
					}">
					<v-icon name="hi-menu" class="text-muted shrink-0" aria-hidden="true" />
					<button type="button" class="flex-1 text-left truncate" @click="select(p.PageName)">{{ p.PageName }}</button>
					<input type="checkbox" class="accent-accent shrink-0 cursor-pointer"
						:checked="p.Visible"
						:title="p.Visible ? 'Visible in menu' : 'Hidden from menu'"
						@change="onToggle(p, $event)" @click.stop @mousedown.stop />
				</div>
```

- [ ] **Step 5: Wire the toggle handler in the script**

In `frontend/src/components/sidePanelComponents/PanelMenu.vue`, update the import to include `setVisibility`, add the `PanelPageSummary` type import, and add the handler:

```ts
import { usePanelState, refresh, select, startNew, reorder, setVisibility } from '@/global/panelState'
import type { PanelPageSummary } from '@/types/PanelModels'
```

```ts
function onToggle(p: PanelPageSummary, e: Event) {
	setVisibility(p.PageName, (e.target as HTMLInputElement).checked)
}
```

- [ ] **Step 6: Add the 422 error branch in the editor**

In `frontend/src/components/panelComponents/PageEditor.vue`, in `save()`'s `catch`, replace the error assignment with:

```ts
		} catch (e: any) {
			error.value = e?.response?.status === 409
				? 'A page with that name already exists.'
				: e?.response?.status === 422
					? 'That path is already used by another page.'
					: 'Save failed.'
		}
```

- [ ] **Step 7: Type-check the frontend**

Run: `cd frontend && npm run type-check`
Expected: no type errors.

- [ ] **Step 8: Browser E2E**

With backend restarted and the frontend dev server running, open the panel (`/panel`, logged in). Verify:
1. Each page row shows a checkbox on the right; all checked initially.
2. Unchecking a page removes it from the public side menu (open the side menu / refetch); re-checking restores it.
3. Clicking the checkbox does not select the row or start a drag; clicking the page name still selects it; drag-reorder still works.
4. In the editor, set a page's Menu path to one already used by another page and Save → the message "That path is already used by another page." appears; a unique path saves fine.

**Green checkpoint — do not commit.**

---

### Task 5: Toast on visibility toggle + live side-menu refresh (frontend)

Added after Task 4 at the owner's request: toggling the checkbox shows a small toast confirming the change, and the public side menu refreshes live to reflect it.

**Files:**
- Create: `frontend/src/global/notify.ts` — reactive toast store: `notify(message, kind?, durationMs?)`, `dismiss(id)`, `useToasts()`.
- Create: `frontend/src/global/notify.spec.ts` — 4 tests (add, error kind, dismiss, auto-dismiss with fake timers).
- Create: `frontend/src/components/AppToast.vue` — minimal fixed bottom-right toast list, token-styled (`bg-surface-2`, `border-accent`/`border-red-500`).
- Create: `frontend/src/global/menuRefresh.ts` — shared `menuVersion` ref + `refreshMenu()`.
- Modify: `frontend/src/App.vue` — mount `<AppToast />`.
- Modify: `frontend/src/global/panelState.ts` — `setVisibility` calls `notify(...)` + `refreshMenu()` on success, `notify(error)` on failure.
- Modify: `frontend/src/views/sideMenu.vue` — read `menuVersion.value` inside the `watchEffect` so `refreshMenu()` triggers a `getMenu()` refetch.

**Verification:** `notify.spec.ts` passes; browser E2E — toggling a page's checkbox shows "<name> is now visible in/hidden from the menu" and the public side menu drops/returns that page live (verified against `GET /MenuList/Menu`).

---

## Completion: full verification + owner-gated commit

- [ ] **Step 1: Run all automated checks**

```bash
cd backend && go build ./... && go test ./Controllers/
cd ../frontend && npm run type-check && npm run test:unit -- run
```
Expected: backend builds, 5 `buildNav` tests pass; frontend type-checks; existing Vitest suites (incl. `menuReorder`) still pass.

- [ ] **Step 2: Clean up any scratch data**

Ensure no leftover scratch pages from verification remain (e.g. `PathA`/`PathB`); delete via the DELETE endpoint if present.

- [ ] **Step 3: Mark README task 3 done**

In `README.md`, change the task 3 line from `[ ]` to `[x]`:
`[x] Panel Menusunde sayfalarin sag kenarinda checkbox olacak. Isaretliyse slider menude gozuken bir sayfa olacak.`

- [ ] **Step 4: Ask the owner to commit**

Per commit discipline, do NOT commit automatically. Present the completed, verified change set and ask the owner for approval. On approval, make ONE feature commit including: the backend changes (Tasks 1–3), the frontend changes (Task 4), the rewritten `Menu.Controller_test.go`, the README task 3 checkbox, and the already-staged spec file `docs/superpowers/specs/2026-08-03-panel-page-visibility-design.md`. Suggested message:

```
feat(panel): per-page visibility with page-driven navigation

Add a visibility checkbox to each panel page row and make the page the
single source of truth for the public nav (membership, order,
visibility). GetMenu now builds the nav from visible, Order-sorted pages
joined to menu docs for captions/paths (buildNav); page-less seed stubs
drop out. Store an inverted Hidden flag (missing = visible, no
migration). Toggle via PUT /PageVisibility. Reject duplicate non-empty
menu paths on create/update with 422.
```

After committing, update the `panel-crud-review-backlog` memory (task 3 done, branch commit count).

---

## Self-Review

**Spec coverage:**
- Inverted `Hidden` storage, `Visible = !Hidden` → Task 1 (Steps 1–3).
- `VisibilityRequest`, `SetVisibility`, `PUT /PageVisibility` → Task 1 (Steps 2, 5, 6).
- New pages default visible → Task 1 (Step 4).
- Page-driven `buildNav`, `GetMenu` rewrite, remove `sortMenusByOrder`, stubs drop, caption fallback, Order preserved → Task 2.
- `PathTakenByOther`, create/update validation before write, 422 → Task 3.
- Frontend `Visible` type, service, optimistic state, row/checkbox restructure, 422 editor branch → Task 4.
- Go `buildNav` tests + manual E2E → Task 2 tests, Tasks 1/3/4 verification steps.
- README task 3 checkbox + single owner-gated commit incl. spec → Completion.

**Placeholder scan:** none — every code step has concrete code; every run step has an exact command and expected result.

**Type consistency:** `buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel`, `SetVisibility(name string, visible bool) error`, `PathTakenByOther(path, excludePageName string) (bool, error)`, `PanelService.setVisibility(pageName, visible)`, `panelState.setVisibility(name, visible)`, `PanelPageSummary.Visible` — names and signatures match across the tasks that define and consume them. `Hidden`/`Visible` inversion is applied consistently (storage `Hidden`, API/UI `Visible`).
