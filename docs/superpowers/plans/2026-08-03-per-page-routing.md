# Per-page Routing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give every page its own bookmarkable URL (its `Path`) so the side menu navigates to a real route and refresh/deep-link works, while the app stays a single-page application.

**Architecture:** `Path` becomes a first-class field on the page document (required, unique). The public menu is still page-driven (`buildNav`), but each nav item's `Path` now comes from the page, not the menu doc. The frontend router gains a catch-all that renders `contents`, which fetches the page by `route.path`. The Vuex `ActivePage` machinery is deleted — `route.path` is the single source of truth.

**Tech Stack:** Go + Gin + MongoDB (backend, `:8080`), Vue 3 `<script setup>` + TypeScript + Vue Router 4 + Vuex 4 + Vite (frontend, `:5173`), Vitest (jsdom) for frontend units, Go `testing` for backend pure units.

## Global Constraints

- **Documentation is written in English** (proper nouns excepted).
- **Do NOT commit per task.** This repo commits only when a whole feature/fix is done, and only when the owner initiates it (see `memory/` → commit-discipline). Every task below ends by leaving the tree green (tests/build pass) — **not** by committing. The approved design spec (`docs/superpowers/specs/2026-08-03-per-page-routing-design.md`) and this plan fold into the single, owner-initiated feature commit at the very end.
- **Backend is run as a compiled binary, not `air`.** Any backend behaviour check requires: `cd backend && go build -o ./tmp/main .`, then kill the running `./tmp/main` process and start it again. A struct/field change is not live until the binary is rebuilt and restarted.
- **Dev MongoDB:** URI `mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin`, database `KuvartiBlog`, container `database-mongo-1`. Collections `Pages`, `Menus`.
- **Reserved routes** (cannot be a page path): `/panel`, `/lists`.
- **Backend gates:** `cd backend && go test ./...` and `go build ./...` must pass.
- **Frontend gates:** `cd frontend && npm run test:unit` and `npm run type-check` must pass.
- **Dev admin login (for panel curl checks):** `POST /api/auth/login` with `{"userName":"testadmin","passWord":"testpass123"}` returns `{ ..., "token": "<jwt>" }`. These are throwaway dev credentials.

---

## File Structure

**Backend (Go)**
- `backend/Models/Pages.Model.go` — `PageModel` gains `Path`.
- `backend/Models/Panel.Model.go` — `PageSummary`/`PageDetail`/`CreatePageRequest`/`UpdatePageRequest` gain `Path`; `MenuBinding` drops `Path`.
- `backend/Services/Page.Service.go` — `Path` in `List`/`Create`/`Update`; extract `renderAndCache`; add `GetPageByPath`, `PathTaken`.
- `backend/Services/Menu.Service.go` — remove `PathTakenByOther`; `Upsert` stops writing `Path`.
- `backend/Controllers/Page.Controller.go` — `GET /api/Page` accepts `?Path=`.
- `backend/Controllers/Panel.Controller.go` — `Create`/`Update` bind + validate `Path`; add pure `validatePagePathFormat`.
- `backend/Controllers/Panel.Controller_test.go` — **new**, pure unit test for `validatePagePathFormat`.
- `backend/Controllers/Menu.Controller.go` — `buildNav` sources `Path` from the page.
- `backend/Controllers/Menu.Controller_test.go` — update for page-sourced `Path`.

**Frontend (Vue/TS)**
- `frontend/src/router/index.ts` — catch-all route → `contents`.
- `frontend/src/views/contents.vue` — fetch by `route.path`.
- `frontend/src/service/BaseAPI.service.ts` — add `getPageByPath`.
- `frontend/src/components/menuComponents/menuActive.ts` — `isMenuItemActive(item, routePath)`.
- `frontend/src/components/menuComponents/menuActive.spec.ts` — rewrite for new signature.
- `frontend/src/components/menuComponents/MenuItem.vue` — navigate to `Path`; drop `ActivePage`.
- `frontend/src/global/store.ts` — remove `ActivePage` state/getter/mutation/action.
- `frontend/src/App.vue` — drop `SetActivePage(HOME_PAGE)` + `HOME_PAGE` import.
- `frontend/src/global/constants.ts` — **delete** (only held `HOME_PAGE`).
- `frontend/src/types/PanelModels.ts` — `Path` on summary/detail; drop from `MenuBinding`.
- `frontend/src/service/Panel.service.ts` — `SavePagePayload` gains `Path`.
- `frontend/src/global/panelState.ts` — `startNew()` seeds `Path`.
- `frontend/src/components/panelComponents/PageEditor.vue` — "Page path" field bound to the page.

**Docs**
- `CLAUDE.md` — routing/state/endpoint sections refreshed.

---

## Task 1: Backend — page `Path` on the model and service

Adds the `Path` field end-to-end in the data layer: stored on create/update, projected in the list, and queryable by path. The old menu-path uniqueness check in `Panel.Controller` is left untouched here (it still compiles because `MenuBinding.Path` is not removed until Task 4) — this task only widens the happy path.

**Files:**
- Modify: `backend/Models/Pages.Model.go`
- Modify: `backend/Models/Panel.Model.go:3-35`
- Modify: `backend/Services/Page.Service.go:24-35` (interface), `82-102` (List), `113-137` (Create), `171-188` (Update), `227-257` (GetPage → renderAndCache)
- Modify: `backend/Controllers/Panel.Controller.go:82` and `:122` (pass `Path` into `Create`/`Update`, bind `req.Path`)

**Interfaces:**
- Produces (service): `GetPageByPath(path string) (models.PageModel, error)`, `PathTaken(path, excludePageName string) (bool, error)`, `Create(name, path, sourceClean, viewType string) error`, `Update(name, path, sourceClean, viewType string) error`. `List()` now projects `Path` into each `PageSummary`.
- Produces (models): `PageModel.Path`, `PageSummary.Path`, `PageDetail.Path`, `CreatePageRequest.Path`, `UpdatePageRequest.Path` — all `string`.

- [ ] **Step 1: Add `Path` to `PageModel`**

In `backend/Models/Pages.Model.go`, add the field right after `PageName`:

```go
package models

type PageModel struct {
	PageName	string	`json:"PageName" gorm:"unique"`
	Path		string	`json:"Path" bson:"Path"`
	Page		string	`json:"Page"`
	Hash		[]byte	`json:"Hash"`
	Text		string	`json:"Text"`
	ViewType	string	`json:"ViewType"`
	Order		int	`json:"Order"`
	Hidden		bool	`json:"-" bson:"Hidden"`
}
```

- [ ] **Step 2: Add `Path` to the panel DTOs (leave `MenuBinding` alone)**

In `backend/Models/Panel.Model.go`, add `Path` to `PageSummary`, `PageDetail`, `CreatePageRequest`, and `UpdatePageRequest`. Do **not** touch `MenuBinding` yet:

```go
type PageSummary struct {
	PageName string `json:"PageName" bson:"PageName"`
	Path     string `json:"Path" bson:"Path"`
	ViewType string `json:"ViewType" bson:"ViewType"`
	Order    int    `json:"Order" bson:"Order"`
	Hidden   bool   `json:"-" bson:"Hidden"`
	Visible  bool   `json:"Visible" bson:"-"`
}

type PageDetail struct {
	PageName string       `json:"PageName"`
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type CreatePageRequest struct {
	PageName string       `json:"PageName"`
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type UpdatePageRequest struct {
	Path     string       `json:"Path"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}
```

- [ ] **Step 3: Extend the `PageService` interface**

In `backend/Services/Page.Service.go`, update the interface (add `GetPageByPath`, `PathTaken`; change `Create`/`Update` signatures):

```go
type PageService interface {
	GetPage(string) (models.PageModel, error)
	GetPageByPath(path string) (models.PageModel, error)
	ConvertmdToHTML(md []byte) []byte
	Preview(sourceClean string) (string, error)
	List() ([]models.PageSummary, error)
	GetRaw(name string) (models.PageModel, error)
	Create(name, path, sourceClean, viewType string) error
	Update(name, path, sourceClean, viewType string) error
	Delete(name string) error
	SetOrder(names []string) error
	SetVisibility(name string, visible bool) error
	PathTaken(path, excludePageName string) (bool, error)
}
```

- [ ] **Step 4: Project `Path` in `List`**

Add `Path` to the projection in `List()`:

```go
func (psi *PageServiceImplementation) List() ([]models.PageSummary, error) {
	opts := options.Find().SetProjection(bson.D{
		{Key: "PageName", Value: 1},
		{Key: "Path", Value: 1},
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

- [ ] **Step 5: Store `Path` in `Create`**

Add the `path` parameter and the `Path` insert field:

```go
func (psi *PageServiceImplementation) Create(name, path, sourceClean, viewType string) error {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{{Key: "PageName", Value: name}})
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPageExists
	}
	total, err := psi.collection.CountDocuments(psi.ctx, bson.D{{}})
	if err != nil {
		return err
	}
	_, err = psi.collection.InsertOne(psi.ctx, bson.D{
		{Key: "PageName", Value: name},
		{Key: "Path", Value: path},
		{Key: "Page", Value: ToStorage(sourceClean)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: viewType},
		{Key: "Order", Value: total},
		{Key: "Hidden", Value: false},
	})
	return err
}
```

- [ ] **Step 6: Store `Path` in `Update`**

Add the `path` parameter and the `Path` `$set` field:

```go
func (psi *PageServiceImplementation) Update(name, path, sourceClean, viewType string) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Path", Value: path},
			{Key: "Page", Value: ToStorage(sourceClean)},
			{Key: "ViewType", Value: viewType},
			{Key: "Hash", Value: []byte{}},
			{Key: "Text", Value: ""},
		}}},
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

- [ ] **Step 7: Extract `renderAndCache` and add `GetPageByPath`**

Replace the existing `GetPage` (lines ~227-257) with a thin finder that delegates to a shared render/cache helper, and add `GetPageByPath` that finds by `Path` and reuses the same helper. Behaviour (hash check, re-render, cache write, newline strip) is preserved exactly; the cache `UpdateOne` now keys on the page's own `PageName`:

```go
func (psi *PageServiceImplementation) renderAndCache(page models.PageModel) (models.PageModel, error) {
	hasher := sha1.New()
	if _, err := io.WriteString(hasher, page.Page); err != nil {
		return models.PageModel{}, err
	}
	if page.Text == "" || !testEq(page.Hash, hasher.Sum(nil)) {
		text, err := psi.GetPageText(page.Page)
		if err != nil {
			return models.PageModel{}, err
		}
		page.Text = text
		page.Hash = hasher.Sum(nil)
		_, err = psi.collection.UpdateOne(psi.ctx,
			bson.D{{Key: "PageName", Value: page.PageName}},
			bson.D{{Key: "$set", Value: bson.D{
				{Key: "Hash", Value: page.Hash},
				{Key: "Text", Value: page.Text},
			}}},
		)
		if err != nil {
			return models.PageModel{}, err
		}
	}
	page.Text = strings.ReplaceAll(page.Text, "\n", "")
	return page, nil
}

func (psi *PageServiceImplementation) GetPage(pn string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "PageName", Value: pn}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return psi.renderAndCache(page)
}

func (psi *PageServiceImplementation) GetPageByPath(path string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "Path", Value: path}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return psi.renderAndCache(page)
}
```

- [ ] **Step 8: Add `PathTaken`**

Append to `backend/Services/Page.Service.go` (near the other `PageServiceImplementation` methods):

```go
// PathTaken reports whether some other page already owns this path. The current
// page is excluded by PageName so re-saving an unchanged path is allowed.
func (psi *PageServiceImplementation) PathTaken(path, excludePageName string) (bool, error) {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{
		{Key: "Path", Value: path},
		{Key: "PageName", Value: bson.D{{Key: "$ne", Value: excludePageName}}},
	})
	return count > 0, err
}
```

- [ ] **Step 9: Update the two `Panel.Controller` call sites so the build compiles**

`Create`/`Update` now require a `path`. In `backend/Controllers/Panel.Controller.go`, change only the two service calls to thread `req.Path` through (no validation yet — that is Task 4). Line ~82:

```go
	if err := pc.PageService.Create(req.PageName, req.Path, req.Source, req.ViewType); err != nil {
```

Line ~122:

```go
	if err := pc.PageService.Update(name, req.Path, req.Source, req.ViewType); err != nil {
```

Leave the existing `req.Menu.Path` / `MenuService.PathTakenByOther` block and the menu `Upsert` exactly as they are for now.

- [ ] **Step 10: Build the backend and confirm it compiles**

Run: `cd backend && go build ./... && go vet ./...`
Expected: no errors. (No new unit tests here — the new service methods are Mongo-backed and are exercised by curl E2E in Tasks 3–5.)

- [ ] **Step 11: Leave the tree green (no commit)**

Confirm `go build ./...` and existing `go test ./...` pass. Do **not** commit.

---

## Task 2: Migration — backfill `Path` on existing pages

`Path` is now required and unique going forward, but the seeded pages have none. Backfill once against the dev database so existing pages resolve by path. This is a one-time dev data operation, not code.

**Files:** none (data only).

- [ ] **Step 1: Inspect current pages**

Run:

```sh
mongosh "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --quiet --eval 'db.Pages.find({}, {PageName:1, Path:1, _id:0}).forEach(p => print(JSON.stringify(p)))'
```

Expected: each page's `PageName` and (currently absent) `Path`. Note the exact `PageName` set (expected: `MainPage`, `StyleTest`, `SoLong`, `42MainWorks`).

- [ ] **Step 2: Backfill `Path` for every page**

`MainPage` becomes home (`/`); every other page gets a lowercased slug of its `PageName`. This is idempotent:

```sh
mongosh "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --quiet --eval 'db.Pages.find({}, {PageName:1}).forEach(function(p){
    var path = p.PageName === "MainPage" ? "/" : "/" + p.PageName.toLowerCase();
    db.Pages.updateOne({_id: p._id}, {$set: {Path: path}});
    print(p.PageName + " -> " + path);
  })'
```

Expected output:

```
MainPage -> /
StyleTest -> /styletest
SoLong -> /solong
42MainWorks -> /42mainworks
```

- [ ] **Step 3: Verify uniqueness and coverage**

Run:

```sh
mongosh "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --quiet --eval '
    print("missing: " + db.Pages.countDocuments({$or:[{Path:{$exists:false}},{Path:""}]}));
    db.Pages.aggregate([{$group:{_id:"$Path", n:{$sum:1}}},{$match:{n:{$gt:1}}}]).forEach(d => print("DUP " + d._id + " x" + d.n));
  '
```

Expected: `missing: 0` and no `DUP` lines. If a duplicate prints, rename the offending page's `Path` by hand before continuing.

---

## Task 3: Backend — resolve `GET /api/Page` by `?Path=`

The public page endpoint currently reads `?PageName=` only (and panics if it is absent). Accept `?Path=` (preferred) as well, falling back to `?PageName=`, and return `400` when neither is present.

**Files:**
- Modify: `backend/Controllers/Page.Controller.go:1-33`

**Interfaces:**
- Consumes: `PageService.GetPageByPath(path)` and `PageService.GetPage(name)` (Task 1).

- [ ] **Step 1: Rewrite `GetPage` to branch on the query parameter**

Replace the body of `backend/Controllers/Page.Controller.go`. Add the `models` import:

```go
package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageController struct {
	PageService services.PageService
}

func InitPageController(PageService services.PageService, server *gin.RouterGroup) PageController {
	pc := PageController{
		PageService: PageService,
	}
	server.GET("/Page", pc.GetPage)
	return pc
}

func (pc *PageController) GetPage(ctx *gin.Context) {
	q := ctx.Request.URL.Query()
	var (
		respons models.PageModel
		err     error
	)
	if path := q.Get("Path"); path != "" {
		respons, err = pc.PageService.GetPageByPath(path)
	} else if name := q.Get("PageName"); name != "" {
		respons, err = pc.PageService.GetPage(name)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Path or PageName query is required"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"ViewType": respons.ViewType,
		"Page":     respons.Text,
	})
}
```

- [ ] **Step 2: Rebuild and restart the backend**

Run: `cd backend && go build -o ./tmp/main .`
Then kill the running `./tmp/main` and start it again so the new handler is live.
Expected: build succeeds; server starts on `:8080`.

- [ ] **Step 3: E2E — resolve by path, by name, and the empty case**

Run (against the backfilled DB from Task 2):

```sh
curl -s 'http://localhost:8080/api/Page?Path=/' | head -c 200; echo
curl -s 'http://localhost:8080/api/Page?PageName=MainPage' | head -c 200; echo
curl -s -o /dev/null -w '%{http_code}\n' 'http://localhost:8080/api/Page'
```

Expected: the first two return the same non-empty `{"Page":...,"ViewType":...}` JSON (MainPage's rendered HTML); the third prints `400`.

- [ ] **Step 4: Leave the tree green (no commit)**

Confirm `go build ./...` passes. Do **not** commit.

---

## Task 4: Backend — validate the page path and retire the menu-path check

Move path uniqueness from the menu to the page, add format/reserved validation, and drop the now-dead `MenuBinding.Path` / `PathTakenByOther`. The reserved/format rule is extracted as a pure function so it is unit-testable (TDD).

**Files:**
- Create: `backend/Controllers/Panel.Controller_test.go`
- Modify: `backend/Controllers/Panel.Controller.go` (imports, `CreatePage`, `UpdatePage`, `GetPage` detail, new helper)
- Modify: `backend/Models/Panel.Model.go:11-15` (drop `MenuBinding.Path`)
- Modify: `backend/Services/Menu.Service.go:12-18` (interface), `46-62` (Upsert), remove `78-87` (`PathTakenByOther`)

**Interfaces:**
- Produces: `validatePagePathFormat(path string) pathValidity` with sentinels `pathOK`, `pathBadFormat`, `pathReserved`.
- Consumes: `PageService.PathTaken(path, excludePageName)` (Task 1).

- [ ] **Step 1: Write the failing pure test for path format**

Create `backend/Controllers/Panel.Controller_test.go`:

```go
package controllers

import "testing"

func TestValidatePagePathFormat(t *testing.T) {
	cases := []struct {
		path string
		want pathValidity
	}{
		{"", pathBadFormat},
		{"about", pathBadFormat},
		{"/", pathOK},
		{"/about", pathOK},
		{"/panel", pathReserved},
		{"/lists", pathReserved},
	}
	for _, c := range cases {
		if got := validatePagePathFormat(c.path); got != c.want {
			t.Errorf("validatePagePathFormat(%q) = %d, want %d", c.path, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run it to confirm it fails to compile**

Run: `cd backend && go test ./Controllers/ -run TestValidatePagePathFormat`
Expected: FAIL — `undefined: validatePagePathFormat` (and `pathValidity`/sentinels).

- [ ] **Step 3: Add the pure helper**

At the top of `backend/Controllers/Panel.Controller.go` (after the imports, before `PanelController`), add the type, sentinels, and function. Add `"strings"` to the import block:

```go
// pathValidity classifies a page path so CreatePage/UpdatePage can map each
// failure to the right HTTP status. Uniqueness needs the DB and is checked
// separately; this covers only format and reserved-route rules.
type pathValidity int

const (
	pathOK pathValidity = iota
	pathBadFormat // empty or missing leading slash
	pathReserved  // collides with a client-only route
)

func validatePagePathFormat(path string) pathValidity {
	if path == "" || !strings.HasPrefix(path, "/") {
		return pathBadFormat
	}
	if path == "/panel" || path == "/lists" {
		return pathReserved
	}
	return pathOK
}
```

- [ ] **Step 4: Run the test to confirm it passes**

Run: `cd backend && go test ./Controllers/ -run TestValidatePagePathFormat`
Expected: PASS.

- [ ] **Step 5: Enforce validation in `CreatePage`**

Replace `CreatePage` in `backend/Controllers/Panel.Controller.go` — validate format → `400`, reserved → `422`, then `PathTaken` → `422`, drop the old `req.Menu.Path` block, and drop `Path` from the menu upsert:

```go
func (pc *PanelController) CreatePage(ctx *gin.Context) {
	var req models.CreatePageRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PageName == "" || req.Source == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "PageName and Source are required"})
		return
	}
	switch validatePagePathFormat(req.Path) {
	case pathBadFormat:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path must start with /"})
		return
	case pathReserved:
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path is reserved"})
		return
	}
	taken, err := pc.PageService.PathTaken(req.Path, req.PageName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if taken {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
		return
	}
	if err := pc.PageService.Create(req.PageName, req.Path, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "a page with that name already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, PageName: req.PageName}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "created"})
}
```

- [ ] **Step 6: Enforce validation in `UpdatePage`**

Replace `UpdatePage` — same validation, excluding the page's own name in `PathTaken`:

```go
func (pc *PanelController) UpdatePage(ctx *gin.Context) {
	name := ctx.Param("name")
	var req models.UpdatePageRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Source == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Source is required"})
		return
	}
	switch validatePagePathFormat(req.Path) {
	case pathBadFormat:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path must start with /"})
		return
	case pathReserved:
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path is reserved"})
		return
	}
	taken, err := pc.PageService.PathTaken(req.Path, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if taken {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
		return
	}
	if err := pc.PageService.Update(name, req.Path, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, PageName: name}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}
```

- [ ] **Step 7: Return the page `Path` in the editor detail**

In `GetPage` (the panel detail handler), populate `Path` from the page and drop `Path` from the `MenuBinding`:

```go
	detail := models.PageDetail{
		PageName: page.PageName,
		Path:     page.Path,
		Source:   services.FromStorage(page.Page),
		ViewType: page.ViewType,
	}
	if menu, err := pc.MenuService.GetByPageName(name); err == nil {
		detail.Menu = &models.MenuBinding{Name: menu.Name, Caption: menu.Caption}
	}
```

- [ ] **Step 8: Drop `Path` from `MenuBinding`**

In `backend/Models/Panel.Model.go`:

```go
type MenuBinding struct {
	Name    string `json:"Name"`
	Caption string `json:"Caption"`
}
```

- [ ] **Step 9: Remove `PathTakenByOther` and stop writing menu `Path`**

In `backend/Services/Menu.Service.go`, remove `PathTakenByOther` from the interface and delete its implementation. In `Upsert`, drop the `Path` `$set` line:

```go
type MenuService interface {
	GetMenu() ([]*models.MenuModel, error)
	Upsert(m models.MenuModel) error
	DeleteByPageName(pageName string) error
	GetByPageName(pageName string) (models.MenuModel, error)
}
```

```go
func (msi *MenuServiceImplementation) Upsert(m models.MenuModel) error {
	if m.Name == "" {
		m.Name = m.PageName
	}
	opts := options.Update().SetUpsert(true)
	_, err := msi.collection.UpdateOne(msi.ctx,
		bson.D{{Key: "PageName", Value: m.PageName}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Name", Value: m.Name},
			{Key: "Caption", Value: m.Caption},
			{Key: "PageName", Value: m.PageName},
		}}},
		opts,
	)
	return err
}
```

(`MenuModel.Path` in `backend/Models/Menu.Model.go` stays — `buildNav` still emits `Path` in Task 5.)

- [ ] **Step 10: Build + unit test**

Run: `cd backend && go build ./... && go test ./...`
Expected: build succeeds, `TestValidatePagePathFormat` passes, existing `buildNav` tests still pass.

- [ ] **Step 11: Rebuild, restart, and E2E the validation**

Run: `cd backend && go build -o ./tmp/main .`, then restart `./tmp/main`. Obtain a token, then exercise the branches:

```sh
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"userName":"testadmin","passWord":"testpass123"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')

# reserved -> 422
curl -s -o /dev/null -w 'reserved=%{http_code}\n' -X POST http://localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"TmpA","Path":"/panel","Source":"x","ViewType":"PlainHTML"}'

# bad format -> 400
curl -s -o /dev/null -w 'badformat=%{http_code}\n' -X POST http://localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"TmpA","Path":"nolead","Source":"x","ViewType":"PlainHTML"}'

# duplicate of MainPage's "/" -> 422
curl -s -o /dev/null -w 'dup=%{http_code}\n' -X POST http://localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"TmpA","Path":"/","Source":"x","ViewType":"PlainHTML"}'
```

Expected: `reserved=422`, `badformat=400`, `dup=422`. (No page named `TmpA` should be created — verify with `db.Pages.countDocuments({PageName:"TmpA"})` = 0.)

- [ ] **Step 12: Leave the tree green (no commit)**

Confirm `go test ./...` and `go build ./...` pass. Do **not** commit.

---

## Task 5: Backend — `buildNav` sources `Path` from the page

The nav item's `Path` must be the page's real URL, not the (now-unwritten) menu-doc path. Rewrite `buildNav` to take `Path` from the page and `Caption` from the menu doc (fallback to `PageName`). TDD via the existing controller test file.

**Files:**
- Modify: `backend/Controllers/Menu.Controller.go:49-73` (`buildNav`)
- Modify: `backend/Controllers/Menu.Controller_test.go` (update `TestBuildNavKeepsMenuCaption`, add `TestBuildNavPathFromPage`)

**Interfaces:**
- Consumes: `PageSummary.Path` (Task 1).

- [ ] **Step 1: Write/adjust the failing tests**

In `backend/Controllers/Menu.Controller_test.go`, set the page `Path` in `TestBuildNavKeepsMenuCaption` and add a new test that proves the nav path comes from the page even when a stale menu path exists:

```go
func TestBuildNavKeepsMenuCaption(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Path: "/a", Visible: true}}
	menus := []*models.MenuModel{{PageName: "A", Caption: "Alpha", Path: "/a"}}
	nav := buildNav(pages, menus)
	if nav[0].Caption != "Alpha" || nav[0].Path != "/a" {
		t.Fatalf("expected menu caption/page path, got %+v", nav[0])
	}
}

func TestBuildNavPathFromPage(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Path: "/alpha", Visible: true}}
	menus := []*models.MenuModel{{PageName: "A", Caption: "Alpha", Path: "/stale"}}
	nav := buildNav(pages, menus)
	if nav[0].Path != "/alpha" {
		t.Fatalf("expected nav path from page '/alpha', got %q", nav[0].Path)
	}
	if nav[0].Caption != "Alpha" {
		t.Fatalf("expected caption 'Alpha', got %q", nav[0].Caption)
	}
}
```

- [ ] **Step 2: Run to confirm the new test fails**

Run: `cd backend && go test ./Controllers/ -run TestBuildNavPathFromPage`
Expected: FAIL — current `buildNav` returns the menu doc verbatim, so `nav[0].Path` is `/stale`, not `/alpha`.

- [ ] **Step 3: Rewrite `buildNav`**

Replace `buildNav` in `backend/Controllers/Menu.Controller.go`:

```go
// buildNav produces the public navigation from the pages (the source of truth
// for membership, order, and visibility) joined to their menu documents for the
// display caption only. Each nav item's Path comes from the page itself. Pages
// arrive already Order-sorted from PageService.List(). A visible page with no
// menu caption falls back to its PageName; menu documents with no matching page
// (hand-seeded stubs) are dropped.
func buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel {
	capByPage := make(map[string]string, len(menus))
	for _, m := range menus {
		if m.PageName != "" {
			capByPage[m.PageName] = m.Caption
		}
	}
	nav := make([]*models.MenuModel, 0, len(pages))
	for _, p := range pages {
		if !p.Visible {
			continue
		}
		caption := p.PageName
		if c, ok := capByPage[p.PageName]; ok && c != "" {
			caption = c
		}
		nav = append(nav, &models.MenuModel{
			Name:     p.PageName,
			Caption:  caption,
			Path:     p.Path,
			PageName: p.PageName,
		})
	}
	return nav
}
```

- [ ] **Step 4: Run the full controller test suite**

Run: `cd backend && go test ./Controllers/ -v`
Expected: all `TestBuildNav*` pass (`FiltersHidden`, `DropsStubs`, `CaptionFallback`, `KeepsMenuCaption`, `PreservesOrder`, `PathFromPage`) plus `TestValidatePagePathFormat`.

- [ ] **Step 5: Rebuild, restart, and confirm the live menu carries paths**

Run: `cd backend && go build -o ./tmp/main .`, restart `./tmp/main`, then:

```sh
curl -s 'http://localhost:8080/api/MenuList/Menu'
```

Expected: JSON array where each item's `Path` matches the backfilled page path (e.g. `MainPage` → `"/"`), not an empty string.

- [ ] **Step 6: Leave the tree green (no commit)**

Confirm `go test ./...` and `go build ./...` pass. Do **not** commit.

---

## Task 6: Frontend — catch-all route + fetch `contents` by `route.path`

Point the router's content route at a catch-all and make `contents.vue` fetch by the current path via a new `getPageByPath` service method. `/panel` and `/lists` keep precedence.

**Files:**
- Modify: `frontend/src/service/BaseAPI.service.ts:30-45`
- Modify: `frontend/src/router/index.ts:6-22`
- Modify: `frontend/src/views/contents.vue:13-33`

**Interfaces:**
- Produces: `serviceClass.getPageByPath(path: string): Promise<PageResponseModal>`.

- [ ] **Step 1: Add `getPageByPath` to the API service**

In `frontend/src/service/BaseAPI.service.ts`, add a method mirroring `getPage` but keyed on `?Path=` (URL-encoded), directly after `getPage`:

```ts
	public async getPageByPath(path: string) : Promise<PageResponseModal> {
		return new Promise((resolve) => {
			this.apiClient.get('Page?Path=' + encodeURIComponent(path)).catch((reason) => {
				console.log('apiget field fail:', reason);
				resolve({Page: '', ViewType: ''})
			}).then((value) => {
				if (value && value.data){
					value.data.Page = value.data.Page.replace(/\/n/g, '\n').replace(/\\n/g, '\n')
					resolve(value.data)
				}
				else
					resolve({Page: '', ViewType: ''})
			})
		})
	}
```

- [ ] **Step 2: Add the catch-all route (keep static routes first)**

Replace the `routes` array in `frontend/src/router/index.ts`:

```ts
	routes: [
		{
			path: '/panel',
			name: 'panel',
			component: () => Promise.resolve(routes.panel)
		},
		{
			path: '/lists',
			name: 'lists',
			component: () => Promise.resolve(routes.lists)
		},
		{
			path: '/:pathMatch(.*)*',
			name: 'content',
			component: () => Promise.resolve(routes.contents)
		}
	]
```

- [ ] **Step 3: Fetch by `route.path` in `contents.vue`**

Replace the `<script setup>` block of `frontend/src/views/contents.vue` — drop Vuex, use `useRoute`:

```ts
<script setup lang="ts">
import { onMounted, ref, inject, watch } from 'vue';
import { useRoute } from 'vue-router';
import { type ServiceType } from '@/service/BaseAPI.service'

let service:ServiceType = inject<ServiceType>('Service');
let route = useRoute();
let returnedHTML = ref<string>("");

function fetchPage(path: string) {
	service?.getPageByPath(path).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
}

onMounted(() => fetchPage(route.path))
watch(() => route.path, (path) => fetchPage(path))
</script>
```

- [ ] **Step 4: Type-check**

Run: `cd frontend && npm run type-check`
Expected: no errors. (This task has no unit test — routing/axios wiring is verified in the final manual E2E.)

- [ ] **Step 5: Leave the tree green (no commit)**

Confirm `npm run type-check` passes. Do **not** commit.

---

## Task 7: Frontend — simplify active-highlight and delete the `ActivePage` machinery

`route.path` is now the source of truth. Simplify `isMenuItemActive` to a path comparison (TDD), make `MenuItem` navigate straight to the page's `Path`, and remove every `ActivePage`/`HOME_PAGE` reference from the store, `App.vue`, and `constants.ts`.

**Files:**
- Modify: `frontend/src/components/menuComponents/menuActive.spec.ts` (rewrite)
- Modify: `frontend/src/components/menuComponents/menuActive.ts` (rewrite)
- Modify: `frontend/src/components/menuComponents/MenuItem.vue:13-38`
- Modify: `frontend/src/global/store.ts`
- Modify: `frontend/src/App.vue:24-43`
- Delete: `frontend/src/global/constants.ts`

**Interfaces:**
- Produces: `isMenuItemActive(item: { Path?: string }, routePath: string): boolean`.

- [ ] **Step 1: Rewrite the spec to the new signature (failing)**

Replace `frontend/src/components/menuComponents/menuActive.spec.ts`:

```ts
import { describe, it, expect } from 'vitest'
import { isMenuItemActive } from './menuActive'

const Home = { Path: '/' }
const Blog = { Path: '/lists' }
const About = { Path: '/about' }

describe('isMenuItemActive', () => {
	it('lights the item whose Path equals the route', () => {
		expect(isMenuItemActive(Home, '/')).toBe(true)
		expect(isMenuItemActive(Blog, '/')).toBe(false)
		expect(isMenuItemActive(About, '/')).toBe(false)
	})

	it('matches a non-root route by Path', () => {
		expect(isMenuItemActive(Blog, '/lists')).toBe(true)
		expect(isMenuItemActive(About, '/lists')).toBe(false)
		expect(isMenuItemActive(Home, '/lists')).toBe(false)
	})

	it('treats a missing Path as "/"', () => {
		expect(isMenuItemActive({}, '/')).toBe(true)
		expect(isMenuItemActive({}, '/about')).toBe(false)
	})
})
```

- [ ] **Step 2: Run to confirm it fails**

Run: `cd frontend && npm run test:unit -- menuActive`
Expected: FAIL — current `isMenuItemActive` takes three args and imports `HOME_PAGE`; the new two-arg calls give wrong results / type errors.

- [ ] **Step 3: Rewrite `menuActive.ts`**

Replace `frontend/src/components/menuComponents/menuActive.ts`:

```ts
export interface MenuItemLike {
	Path?: string
}

/**
 * Decide whether a side-menu item should be highlighted as the active page.
 * Each page now lives at its own route (its Path), so an item is active exactly
 * when its Path matches the current route. An item without a Path defaults to
 * the home route "/".
 */
export function isMenuItemActive(item: MenuItemLike, routePath: string): boolean {
	return (item.Path || '/') === routePath
}
```

- [ ] **Step 4: Run to confirm it passes**

Run: `cd frontend && npm run test:unit -- menuActive`
Expected: PASS (3 tests).

- [ ] **Step 5: Simplify `MenuItem.vue`**

Replace the `<script setup>` block of `frontend/src/components/menuComponents/MenuItem.vue` — drop Vuex/`activePage`, navigate to `Path`:

```ts
<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import { type MenuListModal } from '@/types/MenuListModal'
import { isMenuItemActive } from '@/components/menuComponents/menuActive'
import router from '@/router';

let route = useRoute();
let props = defineProps<MenuListModal>()
let textColorFunction = computed(() =>
	isMenuItemActive(props, route.path)
		? 'text-activePageColor'
		: 'text-deActivePageColor'
)

let RouterRedirect = () => {
	router.push(props.Path || '/')
}
</script>
```

(The template is unchanged. `text-activePageColor` / `text-deActivePageColor` are Tailwind color classes, unrelated to the removed store field.)

- [ ] **Step 6: Remove `ActivePage` from the Vuex store**

Replace `frontend/src/global/store.ts`:

```ts
import { createStore } from 'vuex'
import { MediaEnum } from '@/components/utils/utils'

export default createStore({
	state: {
		ScreenLevel: MediaEnum.Medium,
		IsMobile: true,
	},
	getters: {
		GetScreenLevel(state: any): typeof MediaEnum {
			return state.ScreenLevel
		},
		GetIsMobile(state: any): boolean {
			return state.ScreenLevel < 2
		},
	},
	mutations: {
		ScreenLevelMutation(state:any, level: typeof MediaEnum) {
			state.ScreenLevel = level;
			state.IsMobile = state.ScreenLevel < 2;
		},
	},
	actions: {
		SetScreenLevel(state:any, width: number) {
			if (width >= 1536) //* 1400 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.ExtraHigh);
			else if (width >= 1280) //* 1200 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.High);
			else if (width >= 1024) //* 992 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.Medium);
			else if (width >= 768) //* 768 Ustu
				state.commit('ScreenLevelMutation', MediaEnum.Small);
			else //* 768 alti
				state.commit('ScreenLevelMutation', MediaEnum.ExtraSmall);
		},
	}
})
```

- [ ] **Step 7: Drop the `SetActivePage` dispatch and `HOME_PAGE` import from `App.vue`**

In `frontend/src/App.vue`, replace the `<script setup>` block:

```ts
<script setup lang="ts">
import * as routes from '@/router/routes'
import { useStore } from 'vuex';
import { onMounted, provide } from 'vue'
import serviceClass from '@/service/BaseAPI.service'
import AppToast from '@/components/AppToast.vue'
provide('Service', serviceClass);

let GlobalStore = useStore();

let handleResize = function() {
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth)
}
onMounted(() => {
	window.addEventListener('resize', handleResize);
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth);
})
</script>
```

- [ ] **Step 8: Confirm `HOME_PAGE` has no remaining importers, then delete `constants.ts`**

Run: `cd frontend && grep -rn "HOME_PAGE\|global/constants" src`
Expected: no matches (App.vue and menuActive.ts were the only two, both updated). Then delete the file:

```sh
rm frontend/src/global/constants.ts
```

If `grep` still shows a match, update that importer before deleting.

- [ ] **Step 9: Type-check + unit tests**

Run: `cd frontend && npm run type-check && npm run test:unit`
Expected: no type errors; all Vitest suites pass (including the rewritten `menuActive`).

- [ ] **Step 10: Leave the tree green (no commit)**

Do **not** commit.

---

## Task 8: Frontend — "Page path" editor field + panel types/service

Move the path input from the menu binding to the page in the admin editor, and thread `Path` through the panel types and save payload.

**Files:**
- Modify: `frontend/src/types/PanelModels.ts`
- Modify: `frontend/src/service/Panel.service.ts:6-11`
- Modify: `frontend/src/global/panelState.ts:50-54`
- Modify: `frontend/src/components/panelComponents/PageEditor.vue:21-25` (template), `71` and `130` (menu default), `152-182` (save)

**Interfaces:**
- Consumes: backend `PageDetail.Path` (Task 4), `CreatePageRequest.Path` / `UpdatePageRequest.Path` (Task 1).

- [ ] **Step 1: Add `Path` to the panel types; drop it from `MenuBinding`**

Replace `frontend/src/types/PanelModels.ts`:

```ts
export interface MenuBinding {
	Name: string
	Caption: string
}

export interface PanelPageSummary {
	PageName: string
	Path: string
	ViewType: string
	Order: number
	Visible: boolean
}

export interface PanelPageDetail {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}
```

- [ ] **Step 2: Add `Path` to `SavePagePayload`**

In `frontend/src/service/Panel.service.ts`:

```ts
export interface SavePagePayload {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}
```

- [ ] **Step 3: Seed `Path` in `startNew()`**

In `frontend/src/global/panelState.ts`, update `startNew` (new page starts with an empty path and a menu binding without `Path`):

```ts
export function startNew(): void {
	state.selected = { PageName: '', Path: '', Source: '', ViewType: '', Menu: { Name: '', Caption: '' } }
	state.isNew = true
	state.dirty = false
}
```

- [ ] **Step 4: Relabel the editor field to "Page path" bound to the page**

In `frontend/src/components/panelComponents/PageEditor.vue`, replace the "Menu path" field (the fourth `flex flex-col` block) with a "Page path" field bound to the page's own `Path`:

```html
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Page path</label>
					<InputText v-model="state.selected.Path"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
```

(Keep the "Menu caption" field bound to `menu.Caption`.)

- [ ] **Step 5: Update the `menu` ref default to drop `Path`**

Two spots reference `{ Name: '', Caption: '', Path: '' }`. Change the ref initializer (line ~71):

```ts
const menu = ref<MenuBinding>({ Name: '', Caption: '' })
```

and the `watch` fallback (line ~130):

```ts
	menu.value = sel?.Menu ?? { Name: '', Caption: '' }
```

- [ ] **Step 6: Send `Path` and map the error codes in `save()`**

Replace `save()` — `hasMenu` now keys on `Caption` (path left the menu), the payload carries `Path`, and the catch maps `400`/`422` to path messages:

```ts
async function save() {
	if (!state.selected) return
	error.value = ''
	const hasMenu = !!menu.value.Caption
	try {
		if (state.isNew) {
			await PanelService.createPage({
				PageName: state.selected.PageName,
				Path: state.selected.Path,
				Source: source.value,
				ViewType: state.selected.ViewType,
				Menu: hasMenu ? menu.value : null,
			})
		} else {
			await PanelService.updatePage(state.selected.PageName, {
				Path: state.selected.Path,
				Source: source.value,
				ViewType: state.selected.ViewType,
				Menu: hasMenu ? menu.value : null,
			})
		}
		const name = state.selected.PageName
		state.dirty = false
		await refresh()
		await select(name)
	} catch (e: any) {
		error.value = e?.response?.status === 409
			? 'A page with that name already exists.'
			: e?.response?.status === 422
				? 'That path is reserved or already used.'
				: e?.response?.status === 400
					? 'Path must start with /.'
					: 'Save failed.'
	}
}
```

- [ ] **Step 7: Type-check + unit tests**

Run: `cd frontend && npm run type-check && npm run test:unit`
Expected: no type errors; all suites pass.

- [ ] **Step 8: Leave the tree green (no commit)**

Do **not** commit.

---

## Task 9: Docs — refresh `CLAUDE.md` for per-page routing

The feature invalidates three `CLAUDE.md` statements: the `/api/Page` query column, the "three routes" routing line + the `ActivePage` state description, and the "Routing coverage / per-page URLs" WIP bullet (now implemented). Update them in the same session (CLAUDE.md is never final). English only.

**Files:**
- Modify: `CLAUDE.md` (API endpoints table, Frontend architecture "Routing" + "State" bullets, WIP section)

- [ ] **Step 1: Update the `/api/Page` endpoint row**

In the API endpoints table, change the `GET | /api/Page` row's query column from `?PageName=<name>` to reflect that `?Path=` is now the primary key:

```
| GET | `/api/Page` | — | `?Path=<path>` (or legacy `?PageName=<name>`) | Rendered page HTML + `ViewType` |
```

- [ ] **Step 2: Update the "Routing" bullet**

In Frontend architecture, replace the routing description so it reflects the catch-all:

```
- **Routing** (`router/index.ts`): `/panel` → `panel` and `/lists` → `lists` are explicit; every other path (including `/`) matches a catch-all (`/:pathMatch(.*)*`) → `contents`, which fetches the page whose `Path` equals `route.path`. View components are re-exported as a barrel in `router/routes.ts`.
```

- [ ] **Step 3: Update the "State" bullet**

Remove the `ActivePage` sentence from the `global/store.ts` description:

```
- **State** (`global/store.ts`): Vuex store tracks responsive `ScreenLevel` (see `MediaEnum`) and a derived `GetIsMobile` getter (`ScreenLevel < 2`). `App.vue` dispatches `SetScreenLevel` on mount and window resize to drive the responsive layout. Which page `contents.vue` shows is driven by `route.path` (per-page routing), not the store.
```

- [ ] **Step 4: Replace the "Routing coverage / per-page URLs" WIP bullet**

That bullet described the pre-routing state as a "later slice." It is now done — remove it from the "Work in progress / incomplete areas" list. In its place (still under WIP), note only the remaining deploy concern:

```
- **Production deep-link fallback.** Client-side per-page routing is implemented (each page lives at its `Path`; the menu navigates to real URLs). Deep-linking in production still needs the web server to serve `index.html` for unknown paths — a deploy-time concern, not wired yet. The Vite dev server already does this in development.
```

- [ ] **Step 5: Sanity-check the doc**

Run: `grep -n "ActivePage\|PageName=<name>\|three routes" CLAUDE.md`
Expected: no stale references remain (no `ActivePage`; the `/api/Page` row and routing line read as updated).

- [ ] **Step 6: Leave the tree green (no commit)**

Do **not** commit.

---

## Feature completion (owner-initiated commit)

Run the full manual end-to-end pass, then hand back to the owner for the single commit — do **not** self-commit.

- [ ] **Full-stack E2E (browser):** with MongoDB, the rebuilt backend (`./tmp/main`), and `npm run dev` all running:
  - Visiting `/` renders `MainPage`; visiting `/solong` renders `SoLong`; both survive a hard refresh (Vite fallback) and are bookmarkable.
  - Clicking a menu item navigates to its `Path` (URL changes, no full reload) and highlights the active item.
  - In the panel, create a page with a fresh `Path` → it appears in the menu and renders at that URL.
  - Saving a second page with an existing `Path` shows "That path is reserved or already used."; `Path` `/panel` shows the same; an empty/no-slash `Path` shows "Path must start with /.".
  - Hiding a page removes it from the menu but its direct URL still renders (visibility is menu-presence, not access control).
- [ ] **Gates green:** `cd backend && go test ./... && go build ./...` and `cd frontend && npm run type-check && npm run test:unit` all pass.
- [ ] **Hand off to the owner.** Report that per-page routing is complete and verified, and that the design spec (`docs/superpowers/specs/2026-08-03-per-page-routing-design.md`) + this plan fold into this one feature commit. Wait for the owner to initiate the commit (commit-discipline).

---

## Out of scope (documented, not built)

- **Production SPA fallback** (server serves `index.html` for unknown paths) — deploy-time concern; noted in CLAUDE.md WIP (Task 9, Step 4).
- **Moving `Caption` onto the page / removing the `Menus` collection** — kept for the display label.
- **Nested/multi-segment paths** beyond a single `/name` segment — the catch-all accepts them, but the panel UX assumes simple paths.
- **Redirect from an old path** when a page's `Path` changes.
- **A Mongo unique index on `Path`** — uniqueness is enforced at the app layer (`PathTaken`); a DB index is optional hardening, skipped to avoid null-collision on any un-backfilled page.

---

## Self-review

**Spec coverage** — every spec section maps to a task:
- Data model (`PageModel.Path`, DTOs, `MenuBinding` drops `Path`) → Tasks 1, 4.
- Page service (`Create`/`Update` signatures, `GetPageByPath`, `PathTaken`, render refactor, `List` projection) → Task 1.
- Menu service (`PathTakenByOther` removed, `Upsert` drops `Path`) → Task 4.
- Menu controller (`buildNav` sources `Path` from page) → Task 5.
- Page controller (`?Path=` resolver, `400` on blank) → Task 3.
- Panel controller (format `400`, reserved `422`, dup `422`, wire `Create`/`Update`, detail returns `Path`) → Tasks 1, 4.
- Frontend router catch-all → Task 6; `contents` by `route.path` → Task 6; `getPageByPath` → Task 6.
- `MenuItem` navigates to `Path`; `menuActive` simplified; Vuex `ActivePage`/`HOME_PAGE` removed → Task 7.
- Panel types/service/editor `Path` → Task 8.
- Migration backfill table → Task 2.
- Testing (buildNav tests, menuActive spec, pure path-format test, manual E2E) → Tasks 4, 5, 7, Completion.
- Docs note for production fallback → Task 9.

**Placeholder scan** — no `TBD`/"add error handling"/"similar to Task N"; every code step carries the actual code.

**Type consistency** — `Create(name, path, sourceClean, viewType)` / `Update(name, path, sourceClean, viewType)`, `GetPageByPath(path)`, `PathTaken(path, excludePageName)`, `isMenuItemActive(item, routePath)`, `getPageByPath(path)`, `SavePagePayload.Path`, `PageDetail.Path`, and `validatePagePathFormat`/`pathValidity`/`pathOK`/`pathBadFormat`/`pathReserved` are named identically everywhere they appear.

**Green-at-every-task check** — Task 1 keeps `MenuBinding.Path` and the old menu-path check so the build stays green while `Create`/`Update` signatures change; Task 4 removes them together with their only call sites. No task leaves the backend or frontend uncompilable.
