# Admin Panel — Page CRUD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an authenticated owner create, edit, delete, and live-preview
Markdown pages (with a bound nav entry) from the admin panel, and make a clicked
nav item load its page.

**Architecture:** New `ControlPanel` REST endpoints (JWT-gated) extend the existing
Go/Gin three-layer backend; the bespoke `/n` line format stays a backend storage
detail (clean newlines cross the API). A minimal Vue panel (page list + plain
text editor + live preview) drives them, plus a two-file nav change so clicking a
menu item loads its page.

**Tech Stack:** Go 1.21 + Gin + mongo-driver v1.14 + gomarkdown (backend); Vue 3
`<script setup>` + TypeScript + Vuex + PrimeVue + Tailwind tokens (frontend);
Go `testing` + Vitest/jsdom (tests).

## Global Constraints

- Go module is `backend`; backend files follow `<Resource>.<Layer>.go`.
- All panel endpoints live under `/api/auth/ControlPanel` and sit behind
  `TokenService.AuthenticateJWT()`.
- The API speaks **clean newlines**; `/n` is backend-only storage. Convert with
  `services.ToStorage` / `services.FromStorage` at the boundary — never leak `/n`
  to the frontend. Do not change the render pipeline (`GetPageText`,
  `ConvertmdToHTML`) **except** the one approved bounded-loop panic fix in
  `GetPageText` (see Task 1/2 notes) — a correctness bug fix, behavior-preserving,
  covered by a regression test.
- Frontend path alias `@/` → `src/`. Reuse theme tokens (`bg-surface`, `text-fg`,
  `text-muted`, `accent`, `border`) and existing PrimeVue components.
- All docs and code comments in **English** (owner preference).
- **Commits are owner-initiated.** Each task's "Commit" step means *"ready to
  commit — pause, tell the owner, and ask for the message theme"*; do not
  auto-commit. (See `memory/git/commit-flow.md`.)
- No page rename in this slice (PageName is immutable on update).

---

## File map

**Backend**
- Modify `backend/Services/Page.Service.go` — conversion helpers, sentinel errors, and `Preview`/`List`/`GetRaw`/`Create`/`Update`/`Delete`.
- Create `backend/Services/Page.Service_test.go` — unit tests for conversion + preview.
- Modify `backend/Services/Menu.Service.go` — `Upsert`/`DeleteByPageName`/`GetByPageName`.
- Create `backend/Models/Panel.Model.go` — request/response DTOs + `PageSummary`.
- Create `backend/Controllers/Panel.Controller.go` — `InitPanelController` + handlers.
- Modify `backend/Controllers/User.Controller.go` — remove the empty `ControlPanel` block.
- Modify `backend/server.go` — construct services once, wire `InitPanelController`.

**Frontend**
- Create `frontend/src/types/PanelModels.ts` — interfaces.
- Create `frontend/src/service/Panel.service.ts` — Bearer-authed API client + 401 handling.
- Create `frontend/src/service/Panel.service.spec.ts` — unit tests.
- Create `frontend/src/global/panelState.ts` — reactive shared state.
- Create `frontend/src/global/panelState.spec.ts` — unit tests.
- Create `frontend/src/components/panelComponents/PageEditor.vue` — editor + preview + save/delete.
- Modify `frontend/src/components/sidePanelComponents/PanelMenu.vue` — page list + new.
- Modify `frontend/src/service/User.service.ts` — add `AuthChecked` ref.
- Modify `frontend/src/views/panel.vue` — loading/login/admin branches.
- Modify `frontend/src/components/menuComponents/MenuItem.vue` — load page by `PageName`.
- Modify `frontend/src/views/contents.vue` — re-fetch on `ActivePage` change.

---

## Task 1: Backend — source/storage conversion helpers

**Files:**
- Modify: `backend/Services/Page.Service.go`
- Test: `backend/Services/Page.Service_test.go` (create)

**Interfaces:**
- Produces: `services.ToStorage(clean string) string`, `services.FromStorage(stored string) string`

- [ ] **Step 1: Write the failing tests**

Create `backend/Services/Page.Service_test.go`:

```go
package services

import "testing"

func TestToStorageReplacesNewlines(t *testing.T) {
	if got := ToStorage("line1\nline2"); got != "line1/nline2" {
		t.Fatalf("ToStorage: got %q, want %q", got, "line1/nline2")
	}
}

func TestFromStorageReplacesDelimiter(t *testing.T) {
	if got := FromStorage("line1/nline2"); got != "line1\nline2" {
		t.Fatalf("FromStorage: got %q, want %q", got, "line1\nline2")
	}
}

func TestRoundTripPreservesNewlines(t *testing.T) {
	original := "# Title\n\nParagraph one\nParagraph two"
	if got := FromStorage(ToStorage(original)); got != original {
		t.Fatalf("round-trip: got %q, want %q", got, original)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./Services/ -run 'ToStorage|FromStorage|RoundTrip' -v`
Expected: FAIL — `undefined: ToStorage` / `undefined: FromStorage`.

- [ ] **Step 3: Add the helpers**

In `backend/Services/Page.Service.go`, after the imports (the `strings` package is
already imported), add:

```go
// ToStorage encodes clean newlines into the bespoke "/n" line delimiter used in
// the stored Page source. FromStorage reverses it. The render pipeline is left
// untouched — these only translate the delimiter at the API boundary.
func ToStorage(clean string) string {
	return strings.ReplaceAll(clean, "\n", "/n")
}

func FromStorage(stored string) string {
	return strings.ReplaceAll(stored, "/n", "\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./Services/ -run 'ToStorage|FromStorage|RoundTrip' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit** *(owner-initiated — pause and confirm)*

```bash
git add backend/Services/Page.Service.go backend/Services/Page.Service_test.go
git commit -m "feat(backend): add clean<->/n source conversion helpers"
```

---

## Task 2: Backend — stateless Preview render

**Files:**
- Modify: `backend/Services/Page.Service.go`
- Test: `backend/Services/Page.Service_test.go`

**Interfaces:**
- Consumes: `ToStorage` (Task 1), existing `GetPageText`
- Produces: `PageService.Preview(sourceClean string) (string, error)`

- [ ] **Step 1: Write the failing tests**

Append to `backend/Services/Page.Service_test.go`:

```go
import "strings" // add to the existing import block if not present

func TestPreviewRendersMarkdownHeading(t *testing.T) {
	psi := &PageServiceImplementation{}
	html, err := psi.Preview("# Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "Hello") {
		t.Fatalf("expected an <h1> containing Hello, got %q", html)
	}
}

func TestPreviewPassesThroughRawHTMLLine(t *testing.T) {
	psi := &PageServiceImplementation{}
	raw := `<div class="x">raw</div>`
	html, err := psi.Preview(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, raw) {
		t.Fatalf("expected raw HTML passthrough, got %q", html)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./Services/ -run 'Preview' -v`
Expected: FAIL — `psi.Preview undefined`.

- [ ] **Step 3: Add `Preview` to the interface and implementation**

In `backend/Services/Page.Service.go`, add to the `PageService` interface:

```go
	Preview(sourceClean string) (string, error)
```

And add the method:

```go
// Preview renders raw clean-newline source to HTML without persisting, mirroring
// GetPage's render path exactly (storage-encode, GetPageText, strip newlines).
func (psi *PageServiceImplementation) Preview(sourceClean string) (string, error) {
	text, err := psi.GetPageText(ToStorage(sourceClean))
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(text, "\n", ""), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./Services/ -run 'Preview' -v`
Expected: PASS (2 tests). Then `go test ./Services/ -v` — all 5 pass.

- [ ] **Step 5: Commit** *(owner-initiated)*

```bash
git add backend/Services/Page.Service.go backend/Services/Page.Service_test.go
git commit -m "feat(backend): add stateless Preview render for the panel"
```

---

## Task 3: Backend — Page & Menu CRUD service methods + DTOs

**Files:**
- Create: `backend/Models/Panel.Model.go`
- Modify: `backend/Services/Page.Service.go`
- Modify: `backend/Services/Menu.Service.go`

**Interfaces:**
- Consumes: `ToStorage` (Task 1); `models.MenuModel`, `models.PageModel` (existing)
- Produces:
  - `models.PageSummary{PageName, ViewType}`, `models.MenuBinding{Name, Caption, Path}`,
    `models.PageDetail{PageName, Source, ViewType, Menu *MenuBinding}`,
    `models.CreatePageRequest{PageName, Source, ViewType, Menu *MenuBinding}`,
    `models.UpdatePageRequest{Source, ViewType, Menu *MenuBinding}`,
    `models.PreviewRequest{Source}`
  - `services.ErrPageExists`, `services.ErrPageNotFound`
  - `PageService.List() ([]models.PageSummary, error)`, `GetRaw(name string) (models.PageModel, error)`,
    `Create(name, sourceClean, viewType string) error`, `Update(name, sourceClean, viewType string) error`,
    `Delete(name string) error`
  - `MenuService.Upsert(m models.MenuModel) error`, `DeleteByPageName(pageName string) error`,
    `GetByPageName(pageName string) (models.MenuModel, error)`

> These methods touch MongoDB. They are **compile-gated** here (`go build`) and
> **behavior-verified end-to-end in Task 5** with curl against the running stack
> (the project has no Mongo test double; per the spec, CRUD unit tests are out of
> scope for this slice).

- [ ] **Step 1: Create the DTO models**

Create `backend/Models/Panel.Model.go`:

```go
package models

type PageSummary struct {
	PageName string `json:"PageName" bson:"PageName"`
	ViewType string `json:"ViewType" bson:"ViewType"`
}

type MenuBinding struct {
	Name    string `json:"Name"`
	Caption string `json:"Caption"`
	Path    string `json:"Path"`
}

type PageDetail struct {
	PageName string       `json:"PageName"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type CreatePageRequest struct {
	PageName string       `json:"PageName"`
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type UpdatePageRequest struct {
	Source   string       `json:"Source"`
	ViewType string       `json:"ViewType"`
	Menu     *MenuBinding `json:"Menu"`
}

type PreviewRequest struct {
	Source string `json:"Source"`
}
```

- [ ] **Step 2: Add sentinel errors + CRUD methods to the page service**

In `backend/Services/Page.Service.go`, add `errors` and the options package to the
import block:

```go
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
```

Add the sentinels (top-level, after imports):

```go
var (
	ErrPageExists   = errors.New("page already exists")
	ErrPageNotFound = errors.New("page not found")
)
```

Extend the `PageService` interface:

```go
	List() ([]models.PageSummary, error)
	GetRaw(name string) (models.PageModel, error)
	Create(name, sourceClean, viewType string) error
	Update(name, sourceClean, viewType string) error
	Delete(name string) error
```

Add the implementations:

```go
func (psi *PageServiceImplementation) List() ([]models.PageSummary, error) {
	opts := options.Find().SetProjection(bson.D{
		{Key: "PageName", Value: 1},
		{Key: "ViewType", Value: 1},
	})
	cursor, err := psi.collection.Find(psi.ctx, bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var summaries []models.PageSummary
	if err := cursor.All(psi.ctx, &summaries); err != nil {
		return nil, err
	}
	return summaries, nil
}

func (psi *PageServiceImplementation) GetRaw(name string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "PageName", Value: name}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return page, nil
}

func (psi *PageServiceImplementation) Create(name, sourceClean, viewType string) error {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{{Key: "PageName", Value: name}})
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPageExists
	}
	_, err = psi.collection.InsertOne(psi.ctx, bson.D{
		{Key: "PageName", Value: name},
		{Key: "Page", Value: ToStorage(sourceClean)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: viewType},
	})
	return err
}

func (psi *PageServiceImplementation) Update(name, sourceClean, viewType string) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{
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

func (psi *PageServiceImplementation) Delete(name string) error {
	res, err := psi.collection.DeleteOne(psi.ctx, bson.D{{Key: "PageName", Value: name}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPageNotFound
	}
	return nil
}
```

- [ ] **Step 3: Add methods to the menu service**

In `backend/Services/Menu.Service.go`, add `"go.mongodb.org/mongo-driver/mongo/options"`
to the imports, extend the `MenuService` interface:

```go
	Upsert(m models.MenuModel) error
	DeleteByPageName(pageName string) error
	GetByPageName(pageName string) (models.MenuModel, error)
```

Add the implementations:

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
			{Key: "Path", Value: m.Path},
			{Key: "PageName", Value: m.PageName},
		}}},
		opts,
	)
	return err
}

func (msi *MenuServiceImplementation) DeleteByPageName(pageName string) error {
	_, err := msi.collection.DeleteMany(msi.ctx, bson.D{{Key: "PageName", Value: pageName}})
	return err
}

func (msi *MenuServiceImplementation) GetByPageName(pageName string) (models.MenuModel, error) {
	var m models.MenuModel
	err := msi.collection.FindOne(msi.ctx, bson.D{{Key: "PageName", Value: pageName}}).Decode(&m)
	return m, err
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: exit 0, no output.

- [ ] **Step 5: Commit** *(owner-initiated)*

```bash
git add backend/Models/Panel.Model.go backend/Services/Page.Service.go backend/Services/Menu.Service.go
git commit -m "feat(backend): add page & menu CRUD service methods"
```

---

## Task 4: Backend — Panel controller + composition wiring

**Files:**
- Create: `backend/Controllers/Panel.Controller.go`
- Modify: `backend/Controllers/User.Controller.go`
- Modify: `backend/server.go`

**Interfaces:**
- Consumes: `PageService` (Tasks 2–3), `MenuService` (Task 3), `TokenService.AuthenticateJWT()`, `services.FromStorage`, `services.ErrPageExists/ErrPageNotFound`, the DTOs
- Produces: routes `GET/POST/PUT/DELETE /api/auth/ControlPanel/Pages[...]`, `POST /api/auth/ControlPanel/Preview`

- [ ] **Step 1: Create the controller**

Create `backend/Controllers/Panel.Controller.go`:

```go
package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PanelController struct {
	PageService services.PageService
	MenuService services.MenuService
}

func InitPanelController(pageService services.PageService, menuService services.MenuService, tokenService services.TokenService, apiGroup *gin.RouterGroup) PanelController {
	pc := PanelController{PageService: pageService, MenuService: menuService}
	cp := apiGroup.Group("/auth/ControlPanel")
	cp.Use(tokenService.AuthenticateJWT())
	{
		cp.GET("/Pages", pc.ListPages)
		cp.GET("/Pages/:name", pc.GetPage)
		cp.POST("/Pages", pc.CreatePage)
		cp.PUT("/Pages/:name", pc.UpdatePage)
		cp.DELETE("/Pages/:name", pc.DeletePage)
		cp.POST("/Preview", pc.Preview)
	}
	return pc
}

func (pc *PanelController) ListPages(ctx *gin.Context) {
	pages, err := pc.PageService.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, pages)
}

func (pc *PanelController) GetPage(ctx *gin.Context) {
	name := ctx.Param("name")
	page, err := pc.PageService.GetRaw(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	detail := models.PageDetail{
		PageName: page.PageName,
		Source:   services.FromStorage(page.Page),
		ViewType: page.ViewType,
	}
	if menu, err := pc.MenuService.GetByPageName(name); err == nil {
		detail.Menu = &models.MenuBinding{Name: menu.Name, Caption: menu.Caption, Path: menu.Path}
	}
	ctx.JSON(http.StatusOK, detail)
}

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
	if err := pc.PageService.Create(req.PageName, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "a page with that name already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, Path: req.Menu.Path, PageName: req.PageName}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "created"})
}

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
	if err := pc.PageService.Update(name, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, Path: req.Menu.Path, PageName: name}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (pc *PanelController) DeletePage(ctx *gin.Context) {
	name := ctx.Param("name")
	if err := pc.PageService.Delete(name); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := pc.MenuService.DeleteByPageName(name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (pc *PanelController) Preview(ctx *gin.Context) {
	var req models.PreviewRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	html, err := pc.PageService.Preview(req.Source)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"Html": html})
}
```

- [ ] **Step 2: Remove the empty ControlPanel block from the user controller**

In `backend/Controllers/User.Controller.go`, delete these two now-redundant lines
from `InitUserController` (the panel controller owns the group now):

```go
	controlpanel := authGroup.Group("ControlPanel");
```
and the trailing block:
```go
	controlpanel.Use(uc.TokenService.AuthenticateJWT())
	{
	}
```

Leave the `/login` and `/AmIAuth` registrations untouched.

- [ ] **Step 3: Wire the controller in the composition root**

In `backend/server.go`, replace the body of `InitControllers` with:

```go
func InitControllers(db *mongo.Database, apiGroup *gin.RouterGroup) {
	userService, tokenService := services.NewUserService(ctx, db.Collection("Users"))
	menuService := services.NewMenuService(ctx, db.Collection("Menus"))
	pageService := services.NewPageService(ctx, db.Collection("Pages"))

	controllers.InitUserController(userService, tokenService, apiGroup)
	controllers.InitMenuController(menuService, apiGroup)
	controllers.InitPageController(pageService, apiGroup)
	controllers.InitPanelController(pageService, menuService, tokenService, apiGroup)
}
```

- [ ] **Step 4: Verify it compiles and boots**

Run: `cd backend && go build ./...` → exit 0.
Run: `cd backend && go run .` and check the startup log lists the new routes:
```
[GIN-debug] GET    /api/auth/ControlPanel/Pages    --> ...
[GIN-debug] POST   /api/auth/ControlPanel/Pages    --> ...
[GIN-debug] GET    /api/auth/ControlPanel/Pages/:name --> ...
[GIN-debug] PUT    /api/auth/ControlPanel/Pages/:name --> ...
[GIN-debug] DELETE /api/auth/ControlPanel/Pages/:name --> ...
[GIN-debug] POST   /api/auth/ControlPanel/Preview  --> ...
```
Expected: no panic, no route conflict. Stop the server (Ctrl-C) after confirming.

- [ ] **Step 5: Commit** *(owner-initiated)*

```bash
git add backend/Controllers/Panel.Controller.go backend/Controllers/User.Controller.go backend/server.go
git commit -m "feat(backend): wire ControlPanel CRUD + preview endpoints"
```

---

## Task 5: Backend — seed a test user + end-to-end curl verification

**Files:** none committed (a throwaway hash generator is created then deleted).

> This validates Tasks 3–4 against the running stack. It seeds a **disposable test
> user** (throwaway password, clearly a fixture). The **owner creates their own
> real account separately and types their own password** — Claude never handles the
> owner's password. Requires MongoDB (`database-mongo-1`) and the backend running.

- [ ] **Step 1: Generate a bcrypt hash for the test user**

Create `backend/tools/genhash/main.go`:

```go
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	h, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), 10)
	fmt.Println(string(h))
}
```

Run: `cd backend && go run ./tools/genhash` → copy the printed `$2a$...` hash.

- [ ] **Step 2: Insert the test user into MongoDB**

```bash
docker exec database-mongo-1 mongosh "mongodb://root:example@localhost:27017/KuvartiBlog?authSource=admin" \
  --quiet --eval 'db.Users.insertOne({Username:"testadmin", Password:"PASTE_HASH_HERE", UserType:"admin"})'
```

- [ ] **Step 3: Log in and capture the token**

```bash
curl -s -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"userName":"testadmin","passWord":"testpass123"}'
```
Expected: `{"Message":"Login Complated!","token":"<JWT>"}`. Export it:
`TOKEN=<JWT>`

- [ ] **Step 4: Exercise every endpoint**

```bash
# Create a page + menu binding
curl -s -X POST localhost:8080/api/auth/ControlPanel/Pages -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"PageName":"TestPage","Source":"# Hello\n\nBody line.","ViewType":"","Menu":{"Caption":"Test","Path":"/"}}'
# Expected: HTTP 201 {"message":"created"}

# List
curl -s localhost:8080/api/auth/ControlPanel/Pages -H "Authorization: Bearer $TOKEN"
# Expected: array including {"PageName":"TestPage","ViewType":""}

# Get raw (clean newlines back, menu present)
curl -s localhost:8080/api/auth/ControlPanel/Pages/TestPage -H "Authorization: Bearer $TOKEN"
# Expected: {"PageName":"TestPage","Source":"# Hello\n\nBody line.","ViewType":"","Menu":{"Name":"TestPage","Caption":"Test","Path":"/"}}

# Preview (no persistence)
curl -s -X POST localhost:8080/api/auth/ControlPanel/Preview -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"Source":"# Hi"}'
# Expected: {"Html":"<h1 ...>Hi</h1>"}

# Update
curl -s -X PUT localhost:8080/api/auth/ControlPanel/Pages/TestPage -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"Source":"# Changed","ViewType":"","Menu":{"Caption":"Test2","Path":"/"}}'
# Expected: {"message":"updated"}

# Public render reflects the change (lazy re-render)
curl -s "localhost:8080/api/Page?PageName=TestPage"
# Expected: {"ViewType":"","Page":"<h1 ...>Changed</h1>"}

# Duplicate create -> 409
curl -s -o /dev/null -w "%{http_code}\n" -X POST localhost:8080/api/auth/ControlPanel/Pages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"PageName":"TestPage","Source":"x","ViewType":""}'
# Expected: 409

# Missing token -> 401
curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/api/auth/ControlPanel/Pages
# Expected: 401

# Delete (cascades menu)
curl -s -X DELETE localhost:8080/api/auth/ControlPanel/Pages/TestPage -H "Authorization: Bearer $TOKEN"
# Expected: {"message":"deleted"}
```

- [ ] **Step 5: Clean up the throwaway generator**

```bash
rm -rf backend/tools/genhash
```
(The `backend/tools/` dir is not committed. The test user may stay for later manual
testing or be removed via `db.Users.deleteOne({Username:"testadmin"})`.)

No commit — this task is verification only.

---

## Task 6: Frontend — types + Panel.service (Bearer + 401)

**Files:**
- Create: `frontend/src/types/PanelModels.ts`
- Create: `frontend/src/service/Panel.service.ts`
- Test: `frontend/src/service/Panel.service.spec.ts`

**Interfaces:**
- Consumes: `serviceClass` (base), `LocalStorageService`, `UserService.IsLogin`, the backend endpoints
- Produces:
  - types `MenuBinding`, `PanelPageSummary`, `PanelPageDetail`
  - `PanelService` default export with `listPages()`, `getPage(name)`, `createPage(payload)`,
    `updatePage(name, payload)`, `deletePage(name)`, `preview(source)`; and
    `SavePagePayload{PageName, Source, ViewType, Menu}`

- [ ] **Step 1: Create the types**

Create `frontend/src/types/PanelModels.ts`:

```ts
export interface MenuBinding {
	Name: string
	Caption: string
	Path: string
}

export interface PanelPageSummary {
	PageName: string
	ViewType: string
}

export interface PanelPageDetail {
	PageName: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}
```

- [ ] **Step 2: Write the failing tests**

Create `frontend/src/service/Panel.service.spec.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'

// Shared axios spies (both UserService and PanelService call axios.create).
// Default resolutions keep UserService's AmIAuth (fired in its constructor) happy.
const getMock = vi.fn(() => Promise.resolve({ data: [] }))
const postMock = vi.fn(() => Promise.resolve({ data: {} }))
const putMock = vi.fn(() => Promise.resolve({ data: {} }))
const deleteMock = vi.fn(() => Promise.resolve({ data: {} }))
vi.mock('axios', () => ({
	default: { create: () => ({ get: getMock, post: postMock, put: putMock, delete: deleteMock }) },
}))

// jsdom runs on an opaque origin where localStorage is unavailable — stub it.
const store: Record<string, string> = {}
vi.stubGlobal('localStorage', {
	getItem: (k: string) => (k in store ? store[k] : null),
	setItem: (k: string, v: string) => { store[k] = v },
	clear: () => { for (const k in store) delete store[k] },
})

import PanelService from '@/service/Panel.service'
import UserService from '@/service/User.service'

describe('Panel.service', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		store['AuthToken'] = 'tok123'
	})

	it('attaches the bearer token from local storage', async () => {
		getMock.mockResolvedValueOnce({ data: [] })
		await PanelService.listPages()
		expect(getMock).toHaveBeenCalledWith('/Pages', {
			headers: { Authorization: 'Bearer tok123' },
		})
	})

	it('flips IsLogin to false on a 401', async () => {
		UserService.IsLogin.value = true
		getMock.mockRejectedValueOnce({ response: { status: 401 } })
		await PanelService.listPages()
		expect(UserService.IsLogin.value).toBe(false)
	})
})
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd frontend && npm run test:unit -- Panel.service`
Expected: FAIL — cannot resolve `@/service/Panel.service`.

- [ ] **Step 4: Create the service**

Create `frontend/src/service/Panel.service.ts`:

```ts
import { serviceClass } from '@/service/BaseAPI.service'
import { LocalStorageService } from '@/service/LocalStorage.service'
import UserService from '@/service/User.service'
import type { PanelPageSummary, PanelPageDetail, MenuBinding } from '@/types/PanelModels'

export interface SavePagePayload {
	PageName: string
	Source: string
	ViewType: string
	Menu: MenuBinding | null
}

class PanelService extends serviceClass {
	private localStorage: LocalStorageService
	constructor() {
		super('/auth/ControlPanel')
		this.localStorage = new LocalStorageService()
	}
	private authConfig() {
		const token = this.localStorage.GetData('AuthToken')
		return { headers: { Authorization: `Bearer ${token ? token : 'nullvalue'}` } }
	}
	private handleAuthError(err: any): never {
		if (err?.response?.status === 401) {
			UserService.IsLogin.value = false
		}
		throw err
	}
	public async listPages(): Promise<PanelPageSummary[]> {
		return this.apiClient.get('/Pages', this.authConfig())
			.then((r) => r.data as PanelPageSummary[])
			.catch((e) => { try { this.handleAuthError(e) } catch { /* swallow */ } return [] })
	}
	public async getPage(name: string): Promise<PanelPageDetail> {
		return this.apiClient.get(`/Pages/${encodeURIComponent(name)}`, this.authConfig())
			.then((r) => r.data as PanelPageDetail)
			.catch((e) => this.handleAuthError(e))
	}
	public async createPage(payload: SavePagePayload): Promise<void> {
		return this.apiClient.post('/Pages', payload, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async updatePage(name: string, payload: Omit<SavePagePayload, 'PageName'>): Promise<void> {
		return this.apiClient.put(`/Pages/${encodeURIComponent(name)}`, payload, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async deletePage(name: string): Promise<void> {
		return this.apiClient.delete(`/Pages/${encodeURIComponent(name)}`, this.authConfig())
			.then(() => undefined)
			.catch((e) => this.handleAuthError(e))
	}
	public async preview(source: string): Promise<string> {
		return this.apiClient.post('/Preview', { Source: source }, this.authConfig())
			.then((r) => r.data.Html as string)
			.catch((e) => { try { this.handleAuthError(e) } catch { /* swallow */ } return '' })
	}
}

export type PanelServiceType = PanelService
export default new PanelService()
```

> Note the base URL math: `serviceClass` sets `baseURL = 'http://localhost:8080/api' + '/auth/ControlPanel'`, so `apiClient.get('/Pages')` hits `/api/auth/ControlPanel/Pages` — same pattern `User.service.ts` uses with `super('/auth')`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd frontend && npm run test:unit -- Panel.service`
Expected: PASS (2 tests).

- [ ] **Step 6: Commit** *(owner-initiated)*

```bash
git add frontend/src/types/PanelModels.ts frontend/src/service/Panel.service.ts frontend/src/service/Panel.service.spec.ts
git commit -m "feat(frontend): add Panel API service with bearer auth + 401 handling"
```

---

## Task 7: Frontend — panelState reactive store

**Files:**
- Create: `frontend/src/global/panelState.ts`
- Test: `frontend/src/global/panelState.spec.ts`

**Interfaces:**
- Consumes: `PanelService` (Task 6), types (Task 6)
- Produces: `usePanelState()` → `{ pages, selected, isNew, dirty }`; `refresh()`, `select(name)`, `startNew()`

- [ ] **Step 1: Write the failing tests**

Create `frontend/src/global/panelState.spec.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/service/Panel.service', () => ({
	default: { listPages: vi.fn(), getPage: vi.fn() },
}))

import PanelService from '@/service/Panel.service'
import { usePanelState, refresh, select, startNew } from '@/global/panelState'

describe('panelState', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		const s = usePanelState()
		s.pages = []
		s.selected = null
		s.dirty = false
		s.isNew = false
	})

	it('refresh populates pages from the service', async () => {
		;(PanelService.listPages as any).mockResolvedValue([{ PageName: 'A', ViewType: '' }])
		await refresh()
		expect(usePanelState().pages).toEqual([{ PageName: 'A', ViewType: '' }])
	})

	it('select loads a page and clears dirty/new', async () => {
		;(PanelService.getPage as any).mockResolvedValue({ PageName: 'A', Source: 'x', ViewType: '', Menu: null })
		usePanelState().dirty = true
		await select('A')
		const s = usePanelState()
		expect(s.selected?.PageName).toBe('A')
		expect(s.dirty).toBe(false)
		expect(s.isNew).toBe(false)
	})

	it('startNew creates an empty editable page', () => {
		startNew()
		const s = usePanelState()
		expect(s.isNew).toBe(true)
		expect(s.selected?.PageName).toBe('')
	})
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npm run test:unit -- panelState`
Expected: FAIL — cannot resolve `@/global/panelState`.

- [ ] **Step 3: Create the store**

Create `frontend/src/global/panelState.ts`:

```ts
import { reactive } from 'vue'
import PanelService from '@/service/Panel.service'
import type { PanelPageSummary, PanelPageDetail } from '@/types/PanelModels'

interface PanelState {
	pages: PanelPageSummary[]
	selected: PanelPageDetail | null
	isNew: boolean
	dirty: boolean
}

const state = reactive<PanelState>({ pages: [], selected: null, isNew: false, dirty: false })

export function usePanelState() {
	return state
}

export async function refresh(): Promise<void> {
	state.pages = await PanelService.listPages()
}

export async function select(name: string): Promise<void> {
	state.selected = await PanelService.getPage(name)
	state.isNew = false
	state.dirty = false
}

export function startNew(): void {
	state.selected = { PageName: '', Source: '', ViewType: '', Menu: { Name: '', Caption: '', Path: '' } }
	state.isNew = true
	state.dirty = false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npm run test:unit -- panelState`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit** *(owner-initiated)*

```bash
git add frontend/src/global/panelState.ts frontend/src/global/panelState.spec.ts
git commit -m "feat(frontend): add shared panelState store"
```

---

## Task 8: Frontend — PageEditor component (editor + live preview + save/delete)

**Files:**
- Create: `frontend/src/components/panelComponents/PageEditor.vue`

**Interfaces:**
- Consumes: `usePanelState`, `refresh`, `select` (Task 7); `PanelService` (Task 6); `MenuBinding` (Task 6); PrimeVue `InputText`/`Button`; `.content` styling from `content.css`
- Produces: `PageEditor.vue` default component

- [ ] **Step 1: Create the component**

Create `frontend/src/components/panelComponents/PageEditor.vue`:

```vue
<template>
	<div class="flex flex-col h-full p-4 gap-3">
		<div v-if="!state.selected" class="text-muted">Select a page or create a new one.</div>
		<template v-else>
			<div class="flex flex-wrap gap-3 items-end">
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Page name</label>
					<InputText v-model="state.selected.PageName" :disabled="!state.isNew"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">View type</label>
					<InputText v-model="state.selected.ViewType"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Menu caption</label>
					<InputText v-model="menu.Caption"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Menu path</label>
					<InputText v-model="menu.Path"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
			</div>

			<div class="flex flex-1 gap-4 min-h-0">
				<textarea v-model="source" @input="onInput"
					class="flex-1 bg-surface-2 border border-border text-fg rounded p-3 font-mono resize-none"
					placeholder="Write Markdown here..."></textarea>
				<div class="flex-1 overflow-auto border border-border rounded p-3">
					<div class="content" v-html="previewHtml"></div>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<Button label="Save" class="bg-accent text-surface rounded px-4 py-2" @click="save" />
				<template v-if="!state.isNew">
					<Button v-if="!confirming" label="Delete"
						class="border border-border text-fg rounded px-4 py-2" @click="confirming = true" />
					<template v-else>
						<span class="text-fg text-sm">Really delete?</span>
						<Button label="Yes" class="bg-accent text-surface rounded px-3 py-1" @click="doDelete" />
						<Button label="No" class="border border-border text-fg rounded px-3 py-1" @click="confirming = false" />
					</template>
				</template>
				<span v-if="state.dirty" class="text-muted text-sm">unsaved changes</span>
				<span v-if="error" class="text-sm" style="color:#c0392b">{{ error }}</span>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import { ref, watch } from 'vue'
import { usePanelState, refresh, select } from '@/global/panelState'
import PanelService from '@/service/Panel.service'
import type { MenuBinding } from '@/types/PanelModels'

const state = usePanelState()
const source = ref('')
const previewHtml = ref('')
const menu = ref<MenuBinding>({ Name: '', Caption: '', Path: '' })
const confirming = ref(false)
const error = ref('')

let debounce: ReturnType<typeof setTimeout> | undefined

watch(() => state.selected, (sel) => {
	source.value = sel?.Source ?? ''
	menu.value = sel?.Menu ?? { Name: '', Caption: '', Path: '' }
	confirming.value = false
	error.value = ''
	runPreview()
}, { immediate: true })

function onInput() {
	state.dirty = true
	if (debounce) clearTimeout(debounce)
	debounce = setTimeout(runPreview, 400)
}

async function runPreview() {
	previewHtml.value = await PanelService.preview(source.value)
}

async function save() {
	if (!state.selected) return
	error.value = ''
	const hasMenu = !!(menu.value.Caption || menu.value.Path)
	try {
		if (state.isNew) {
			await PanelService.createPage({
				PageName: state.selected.PageName,
				Source: source.value,
				ViewType: state.selected.ViewType,
				Menu: hasMenu ? menu.value : null,
			})
		} else {
			await PanelService.updatePage(state.selected.PageName, {
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
			: 'Save failed.'
	}
}

async function doDelete() {
	if (!state.selected) return
	const name = state.selected.PageName
	try {
		await PanelService.deletePage(name)
		state.selected = null
		state.dirty = false
		await refresh()
	} catch {
		error.value = 'Delete failed.'
	} finally {
		confirming.value = false
	}
}
</script>
```

> Deliberate deviations from the spec's wording, both to stay minimal and to avoid
> blocking browser dialogs during verification: delete uses an **inline two-step
> confirm** (not PrimeVue `ConfirmDialog` / `window.confirm`), and the dirty flag is
> an **informational indicator** (switching pages is not hard-blocked). Both are
> trivially upgradable later.

- [ ] **Step 2: Verify type-check passes**

Run: `cd frontend && npm run type-check`
Expected: exit 0. (Visual behavior is verified in Task 11.)

- [ ] **Step 3: Commit** *(owner-initiated)*

```bash
git add frontend/src/components/panelComponents/PageEditor.vue
git commit -m "feat(frontend): add PageEditor with live preview and save/delete"
```

---

## Task 9: Frontend — PanelMenu page list + new

**Files:**
- Modify: `frontend/src/components/sidePanelComponents/PanelMenu.vue` (replace contents)

**Interfaces:**
- Consumes: `usePanelState`, `refresh`, `select`, `startNew` (Task 7); `UserService.IsLogin`
- Produces: page-list side panel that drives `panelState`

- [ ] **Step 1: Replace the component**

Replace the entire contents of `frontend/src/components/sidePanelComponents/PanelMenu.vue`:

```vue
<template>
	<div v-if="IsAuth" class="flex flex-col h-full w-full p-2 gap-1">
		<button class="text-left px-3 py-2 rounded bg-accent text-surface" @click="startNew">+ New page</button>
		<div class="flex-1 overflow-auto">
			<button v-for="p in state.pages" :key="p.PageName"
				class="block w-full text-left px-3 py-2 rounded hover:bg-surface-2 text-fg"
				:class="{ 'bg-surface-2 font-semibold': state.selected?.PageName === p.PageName }"
				@click="select(p.PageName)">
				{{ p.PageName }}
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { usePanelState, refresh, select, startNew } from '@/global/panelState'
import UserService from '@/service/User.service'

const state = usePanelState()
const IsAuth = UserService.IsLogin

function load() {
	if (IsAuth.value) refresh()
}
onMounted(load)
watch(IsAuth, load) // refresh once the login resolves
</script>
```

> This drops the old dummy PrimeVue `PanelMenu` items. The component is teleported
> into the side panel by `sidePanel.vue` whenever the route contains `panel`; the
> `IsAuth` guard hides the list until login and the `watch` refreshes it right after.

- [ ] **Step 2: Verify type-check passes**

Run: `cd frontend && npm run type-check`
Expected: exit 0.

- [ ] **Step 3: Commit** *(owner-initiated)*

```bash
git add frontend/src/components/sidePanelComponents/PanelMenu.vue
git commit -m "feat(frontend): make PanelMenu a live page list"
```

---

## Task 10: Frontend — UserService.AuthChecked

**Files:**
- Modify: `frontend/src/service/User.service.ts`

**Interfaces:**
- Produces: `UserService.AuthChecked` (a `Ref<boolean>`, true once `AmIAuth` resolves)

- [ ] **Step 1: Add the ref and set it in AmIAuth**

In `frontend/src/service/User.service.ts`, add a public field next to `IsLogin`:

```ts
	public AuthChecked = ref(false);
```

Then update `AmIAuth` to flag completion in a `finally` (keep the existing
`IsLogin` logic):

```ts
	private AmIAuth() {
		this.apiClient.get("/AmIAuth", {
			headers: {
				"Authorization": `Bearer ${this.Token ? this.Token : "nullvalue"}`
			}
		}).then(() => {
			this.IsLogin.value = true;
		}).catch(() => {
			this.IsLogin.value = false;
		}).finally(() => {
			this.AuthChecked.value = true;
		})
	}
```

- [ ] **Step 2: Verify type-check + existing tests pass**

Run: `cd frontend && npm run type-check && npm run test:unit -- theme`
Expected: type-check exit 0; theme tests still pass.

- [ ] **Step 3: Commit** *(owner-initiated)*

```bash
git add frontend/src/service/User.service.ts
git commit -m "feat(frontend): expose AuthChecked so the panel can gate on auth resolution"
```

---

## Task 11: Frontend — panel.vue authed branch

**Files:**
- Modify: `frontend/src/views/panel.vue` (replace contents)

**Interfaces:**
- Consumes: `UserService.IsLogin` + `UserService.AuthChecked` (Task 10); `panelView.Login`; `PageEditor` (Task 8)

- [ ] **Step 1: Replace the view**

Replace the entire contents of `frontend/src/views/panel.vue`:

```vue
<template>
	<div v-if="!AuthChecked" class="p-6 text-muted">Loading…</div>
	<div v-else-if="!IsAuth">
		<panelView.Login />
	</div>
	<div v-else class="h-full">
		<PageEditor />
	</div>
</template>

<script async setup lang="ts">
import * as panelView from "@/components/panelViews"
import PageEditor from "@/components/panelComponents/PageEditor.vue"
import UserService, { type UserServiceType } from "@/service/User.service"
import { provide } from "vue"
provide<UserServiceType>("UserService", UserService)

const IsAuth = UserService.IsLogin
const AuthChecked = UserService.AuthChecked
</script>
```

- [ ] **Step 2: Verify type-check passes**

Run: `cd frontend && npm run type-check`
Expected: exit 0.

- [ ] **Step 3: Manual smoke test**

With Mongo + backend + `npm run dev` running, open `http://localhost:5173/panel`:
- Not logged in → login form appears (after a brief "Loading…").
- Log in with the test user (`testadmin` / `testpass123`) → the editor appears and
  the side panel lists pages.

- [ ] **Step 4: Commit** *(owner-initiated)*

```bash
git add frontend/src/views/panel.vue
git commit -m "feat(frontend): render the authenticated panel editor"
```

---

## Task 12: Frontend — nav wiring (load page by name)

**Files:**
- Modify: `frontend/src/components/menuComponents/MenuItem.vue`
- Modify: `frontend/src/views/contents.vue`

**Interfaces:**
- Consumes: Vuex `SetActivePage` action + `GetActivePage` getter (existing)
- Produces: clicking a menu item with a `PageName` loads that page at `/`

- [ ] **Step 1: Update MenuItem click handler**

In `frontend/src/components/menuComponents/MenuItem.vue`, add the store import and
replace `RouterRedirect`:

```ts
import { useStore } from 'vuex'
// ...existing imports...
const GlobalStore = useStore()

let RouterRedirect = () => {
	if (props.PageName) {
		GlobalStore.dispatch('SetActivePage', props.PageName)
		router.push('/')
	} else {
		router.push(props.Path || '/')
	}
}
```

- [ ] **Step 2: Make contents.vue re-fetch on ActivePage change**

Replace the `<script setup>` block of `frontend/src/views/contents.vue`:

```ts
import { onMounted, ref, inject, watch } from 'vue';
import { useStore } from 'vuex';
import { type ServiceType } from '@/service/BaseAPI.service'

let service:ServiceType = inject<ServiceType>('Service');
let GlobalStore = useStore()
let returnedHTML = ref<string>("");

function fetchPage(name: string) {
	service?.getPage(name).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
}

onMounted(() => fetchPage(GlobalStore.getters.GetActivePage))
watch(() => GlobalStore.getters.GetActivePage, (name) => fetchPage(name))
```

- [ ] **Step 3: Verify type-check passes**

Run: `cd frontend && npm run type-check`
Expected: exit 0.

- [ ] **Step 4: Manual smoke test**

Reload `http://localhost:5173/`. Click a nav item that has a `PageName` (e.g. the
`TestPage` you create in Task 13) → the content area renders that page without a
full reload.

- [ ] **Step 5: Commit** *(owner-initiated)*

```bash
git add frontend/src/components/menuComponents/MenuItem.vue frontend/src/views/contents.vue
git commit -m "feat(frontend): load a page when its nav item is clicked"
```

---

## Task 13: End-to-end verification + docs

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Full end-to-end walkthrough (desktop + mobile)**

With Mongo + backend + frontend running and logged in as the test user, verify the
full loop in the browser (mobile via DevTools device emulation, per
`memory/workflow/mobile-testing.md`):
1. Panel → **New page** → set PageName `Demo`, write Markdown → preview updates live.
2. Set Menu Caption `Demo` + Path `/` → **Save** → the page appears in the panel list.
3. Go to `/` → the `Demo` item shows in the nav → click it → the page renders.
4. Back in the panel → select `Demo` → edit → Save → reload `/` → change is visible.
5. Panel → select `Demo` → **Delete** → confirm → it disappears from the list and nav.

- [ ] **Step 2: Run the whole test + build gate**

```bash
cd frontend && npm run test:unit -- --run && npm run type-check && npm run build
cd ../backend && go build ./... && go test ./...
```
Expected: all green (frontend tests pass, type-check + build exit 0; backend builds
and `Services` tests pass).

- [ ] **Step 3: Update CLAUDE.md**

- In **Work in progress / incomplete areas**, remove the "Admin control panel" bullet
  (now implemented) and narrow the "Routing coverage" bullet to note that clicking a
  nav item now loads its page, but per-page URLs are still pending.
- In **API endpoints**, add the `ControlPanel` routes (`GET/POST/PUT/DELETE /Pages`,
  `GET /Pages/:name`, `POST /Preview`) with the Bearer-auth column.
- In **Frontend architecture**, note the panel editor (`panel.vue` + `PageEditor.vue`),
  the `panelState` store, `Panel.service.ts`, and the nav-loads-page behavior.

- [ ] **Step 4: Commit** *(owner-initiated)*

```bash
git add CLAUDE.md
git commit -m "docs: record panel CRUD endpoints and behavior in CLAUDE.md"
```

---

## Self-review notes (author)

- **Spec coverage:** page CRUD (T3/T4), preview (T2/T4), menu binding + cascade (T3/T4),
  clean-newline boundary (T1), panel UI + live preview (T6–T11), nav wiring (T12),
  user-seeding prerequisite (T5), tests + E2E (T6/T7/T13) — all covered.
- **Type consistency:** `ToStorage`/`FromStorage`, `Preview`, `List`/`GetRaw`/`Create`/
  `Update`/`Delete`, `Upsert`/`DeleteByPageName`/`GetByPageName`, `PageSummary`/
  `PageDetail`/`MenuBinding`/`*Request`, `PanelService.*`, `panelState.*`,
  `UserService.AuthChecked` — names are used identically across producing and
  consuming tasks.
- **Known deviations from spec (intentional, noted in-task):** inline delete confirm
  instead of `ConfirmDialog`; dirty flag informational rather than blocking; backend
  CRUD is compile-gated + curl-verified rather than unit-tested (no Mongo double).
- **Deviations discovered during execution:**
  - `GetPageText` bounded-loop panic fix + regression test (Task 1/2), approved by the
    owner — see the render-pipeline exception in the spec.
  - `PanelService.getPage` was renamed to `getPageDetail` — the base `serviceClass`
    already declares `getPage(): Promise<PageResponseModal>`, so the override clashed.
    `panelState.select` and the panelState test call `getPageDetail` accordingly.
  - `Panel.service.spec` uses `vi.hoisted` for the axios spies and mocks
    `@/service/User.service` so UserService's import-time side effects (localStorage
    read + AmIAuth request) don't run under jsdom.
