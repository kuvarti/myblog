# ViewType combobox + card/list system — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `ViewType` a real render switch: a reusable "page → card" engine renders one page as a linked card (title + summary + image), driven either by a `<card path="…">` shortcode in any page or by a `ViewType = List` page that auto-lists pages by tag.

**Architecture:** The Markdown→HTML render/cache path is untouched. A new backend `CardService` runs a request-time expansion pass over the rendered HTML (never cached, so referenced-page edits show immediately): it replaces `<card path>` shortcodes and, for `ViewType = List` pages, appends a card grid of tag-matched pages. Card data is read from each referenced page's raw document (auto-extracted, with `Summary`/`Image` overrides). The frontend keeps rendering with `v-html`; it only adds SPA navigation for internal links.

**Tech Stack:** Go + Gin + MongoDB (backend), Vue 3 `<script setup>` + TypeScript + Vue Router 4 + PrimeVue **v3** + Vite (frontend). Backend tests: stdlib `testing` (pure, no Mongo). Frontend tests: Vitest (jsdom).

## Global Constraints

- All docs and code comments are written in **English** (proper nouns excepted); pre-existing Turkish comments are grandfathered.
- The `ViewType` combobox offers exactly two options for now: `PlainHTML` and `List`. `DynamicList` is cancelled.
- PrimeVue is **v3** — the dropdown component is `primevue/dropdown` (`Dropdown`), **not** `Select`.
- Backend follows the Model → Service (interface + `Impl`) → Controller → wire-in-`server.go` layering. New pure logic goes in package-level functions tested without Mongo/HTTP (mirror `buildNav`).
- Frontend testable logic is extracted into plain `.ts` modules with a sibling `*.spec.ts` (mirror `menuActive.ts` / `scrollSync.ts`).
- New `PageModel` fields are optional/back-compat (absent → zero value → old behaviour); no data migration is required.
- Commit messages use conventional style and end with:
  `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`

---

## File Structure

**Backend**
- `backend/Models/Pages.Model.go` — modify: add `Tags`, `Summary`, `Image`, `ListTags`.
- `backend/Models/Panel.Model.go` — modify: add `PageWrite` struct; extend `PageDetail`, `CreatePageRequest`, `UpdatePageRequest`.
- `backend/Services/Card.Service.go` — create: pure helpers + `CardService` interface/impl.
- `backend/Services/Card.Service_test.go` — create: pure unit tests.
- `backend/Services/Page.Service.go` — modify: `Create`/`Update` take `PageWrite`; add `GetRawByPath`, `FindByTags`.
- `backend/Controllers/Page.Controller.go` — modify: inject `CardService`, expand in `GetPage`.
- `backend/Controllers/Panel.Controller.go` — modify: inject `CardService`; pass new fields; expand shortcodes in `Preview`.
- `backend/server.go` — modify: build `CardService`, update controller wiring.

**Frontend**
- `frontend/src/components/panelComponents/tags.ts` (+ `.spec.ts`) — create: comma-string ↔ `string[]`.
- `frontend/src/components/contentLinks.ts` (+ `.spec.ts`) — create: internal-link SPA target resolver.
- `frontend/src/types/PanelModels.ts` — modify: `PanelPageDetail` gains 4 fields.
- `frontend/src/service/Panel.service.ts` — modify: `SavePagePayload` gains 4 fields.
- `frontend/src/global/panelState.ts` — modify: `startNew()` seeds the 4 fields + `ViewType` default.
- `frontend/src/components/panelComponents/PageEditor.vue` — modify: `Dropdown` + tag/summary/image/list-tag inputs, wired to `save()`.
- `frontend/src/views/contents.vue` — modify: delegated click handler for SPA nav.
- `frontend/src/assets/content.css` — modify: card styles.

**Docs**
- `README.md`, `CLAUDE.md` — modify: mark backlog item done; document the card/list system.

---

## Task 1: Backend model fields + CardService pure helpers

**Files:**
- Modify: `backend/Models/Pages.Model.go`
- Create: `backend/Services/Card.Service.go`
- Test: `backend/Services/Card.Service_test.go`

**Interfaces:**
- Consumes: `models.PageModel` (extended here).
- Produces (package-level pure funcs in `services`):
  - `extractSummary(source string) string`
  - `extractImage(source string) string`
  - `cardTitle(caption, pageName string) string`
  - `buildCardHTML(path, title, summary, image string) string`
  - `selectByTags(pages []models.PageModel, listTags []string, selfName string) []models.PageModel`
  - `expandShortcodes(htmlStr string, resolve func(path string) (string, bool)) string`

- [ ] **Step 1: Add the new PageModel fields**

In `backend/Models/Pages.Model.go`, extend the struct:
```go
type PageModel struct {
	PageName string   `json:"PageName" gorm:"unique"`
	Path     string   `json:"Path" bson:"Path"`
	Page     string   `json:"Page"`
	Hash     []byte   `json:"Hash"`
	Text     string   `json:"Text"`
	ViewType string   `json:"ViewType"`
	Order    int      `json:"Order"`
	Hidden   bool     `json:"-" bson:"Hidden"`
	Tags     []string `json:"Tags" bson:"Tags"`
	Summary  string   `json:"Summary" bson:"Summary"`
	Image    string   `json:"Image" bson:"Image"`
	ListTags []string `json:"ListTags" bson:"ListTags"`
}
```

- [ ] **Step 2: Write the failing pure tests**

Create `backend/Services/Card.Service_test.go`:
```go
package services

import (
	models "backend/Models"
	"strings"
	"testing"
)

func TestCardTitleFallsBackToPageName(t *testing.T) {
	if got := cardTitle("Alpha", "A"); got != "Alpha" {
		t.Fatalf("want caption, got %q", got)
	}
	if got := cardTitle("", "A"); got != "A" {
		t.Fatalf("want PageName fallback, got %q", got)
	}
}

func TestExtractSummarySkipsHeadingsAndHTML(t *testing.T) {
	src := "# Title\n<img src=\"x.jpg\">\nThe first real paragraph."
	if got := extractSummary(src); got != "The first real paragraph." {
		t.Fatalf("got %q", got)
	}
}

func TestExtractSummaryStripsMarkersAndEmphasis(t *testing.T) {
	if got := extractSummary("- **Bold** item"); got != "Bold item" {
		t.Fatalf("got %q", got)
	}
	// A leading number that is part of the text must survive.
	if got := extractSummary("2024 was a good year"); got != "2024 was a good year" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractSummaryTruncates(t *testing.T) {
	long := strings.Repeat("a", 200)
	got := extractSummary(long)
	if len([]rune(got)) != 161 || !strings.HasSuffix(got, "…") {
		t.Fatalf("want 160 runes + ellipsis, got %d runes", len([]rune(got)))
	}
}

func TestExtractImagePrefersHTMLThenMarkdown(t *testing.T) {
	if got := extractImage("intro\n<img alt=\"\" src=\"/a.jpg\">"); got != "/a.jpg" {
		t.Fatalf("html img: got %q", got)
	}
	if got := extractImage("intro\n![alt](/b.png \"t\")"); got != "/b.png" {
		t.Fatalf("md img: got %q", got)
	}
	if got := extractImage("no images here"); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestBuildCardHTML(t *testing.T) {
	withImg := buildCardHTML("/p", "Title", "Sum", "/i.jpg")
	if !strings.Contains(withImg, `href="/p"`) ||
		!strings.Contains(withImg, `src="/i.jpg"`) ||
		!strings.Contains(withImg, "Title") || !strings.Contains(withImg, "Sum") {
		t.Fatalf("missing parts: %q", withImg)
	}
	noImg := buildCardHTML("/p", "T", "S", "")
	if strings.Contains(noImg, "<img") {
		t.Fatalf("expected no <img>, got %q", noImg)
	}
}

func TestSelectByTagsOrMatchAndSelfExclude(t *testing.T) {
	pages := []models.PageModel{
		{PageName: "self", Tags: []string{"blog"}},
		{PageName: "A", Tags: []string{"blog", "go"}},
		{PageName: "B", Tags: []string{"go"}},
		{PageName: "C", Tags: []string{"news"}},
	}
	got := selectByTags(pages, []string{"blog", "go"}, "self")
	if len(got) != 2 || got[0].PageName != "A" || got[1].PageName != "B" {
		t.Fatalf("want A,B got %+v", got)
	}
	if len(selectByTags(pages, nil, "self")) != 0 {
		t.Fatalf("empty listTags must match nothing")
	}
}

func TestExpandShortcodesReplacesAndDrops(t *testing.T) {
	resolve := func(path string) (string, bool) {
		if path == "/ok" {
			return "<CARD>", true
		}
		return "", false
	}
	in := `intro <card path="/ok"> mid <card path="/missing"> end`
	got := expandShortcodes(in, resolve)
	if !strings.Contains(got, "<CARD>") || strings.Contains(got, "/missing") ||
		strings.Contains(got, "<card") {
		t.Fatalf("got %q", got)
	}
}
```
(Delete the stray `TestBuildCardHTMLOmitsImageWhenEmpty` line — it is only here to remind you the omit-image case is covered inside `TestBuildCardHTML`.)

The `TestBuildCardHTML` case covers both the with-image and omit-image (`image == ""`) branches, so no separate omit-image test is needed.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./Services -run TestCard -v` (plus the others by name)
Expected: compile error / FAIL — the helpers do not exist yet.

- [ ] **Step 4: Implement the pure helpers**

Create `backend/Services/Card.Service.go` (interface/impl come in Task 2; only the pure funcs + regexes now):
```go
package services

import (
	models "backend/Models"
	"html"
	"regexp"
	"strings"
)

const cardSummaryMax = 160

var (
	cardShortcodeRe = regexp.MustCompile(`<card\s+path="([^"]*)"\s*/?>`)
	htmlImgRe       = regexp.MustCompile(`<img[^>]*\ssrc="([^"]*)"`)
	mdImgRe         = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)`)
	listMarkerRe    = regexp.MustCompile(`^\s*([-*+]\s+|\d+\.\s+|>\s+)`)
	emphasisStrip   = strings.NewReplacer("*", "", "_", "", "`", "")
)

// cardTitle prefers the page's menu caption, falling back to its PageName.
func cardTitle(caption, pageName string) string {
	if strings.TrimSpace(caption) != "" {
		return caption
	}
	return pageName
}

// extractSummary returns the first real paragraph of a page's (clean-newline)
// source: the first non-blank line that is neither a heading, a raw-HTML line,
// nor a Markdown image, with list markers and emphasis characters stripped,
// truncated to cardSummaryMax runes.
func extractSummary(source string) string {
	for _, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.Contains(line, "<") {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "![") {
			continue
		}
		line = listMarkerRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(emphasisStrip.Replace(line))
		if line == "" {
			continue
		}
		if r := []rune(line); len(r) > cardSummaryMax {
			return strings.TrimSpace(string(r[:cardSummaryMax])) + "…"
		}
		return line
	}
	return ""
}

// extractImage returns the first image URL in the source: a raw <img src>
// first, else a Markdown ![alt](url), else "".
func extractImage(source string) string {
	if m := htmlImgRe.FindStringSubmatch(source); m != nil {
		return m[1]
	}
	if m := mdImgRe.FindStringSubmatch(source); m != nil {
		return m[1]
	}
	return ""
}

// buildCardHTML renders one linked card; the <img> is omitted when image == "".
func buildCardHTML(path, title, summary, image string) string {
	img := ""
	if image != "" {
		img = `<img class="card-img" src="` + html.EscapeString(image) + `" alt="">`
	}
	return `<a class="card" href="` + html.EscapeString(path) + `">` + img +
		`<div class="card-body">` +
		`<h3 class="card-title">` + html.EscapeString(title) + `</h3>` +
		`<p class="card-summary">` + html.EscapeString(summary) + `</p>` +
		`</div></a>`
}

// selectByTags is the authoritative List-membership rule: OR intersection with
// listTags, self excluded, incoming Order preserved.
func selectByTags(pages []models.PageModel, listTags []string, selfName string) []models.PageModel {
	if len(listTags) == 0 {
		return nil
	}
	want := make(map[string]bool, len(listTags))
	for _, t := range listTags {
		want[t] = true
	}
	var out []models.PageModel
	for _, p := range pages {
		if p.PageName == selfName {
			continue
		}
		for _, t := range p.Tags {
			if want[t] {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// expandShortcodes replaces every <card path="…"> with resolve(path); an
// unresolved path is dropped so no raw shortcode leaks to the browser.
func expandShortcodes(htmlStr string, resolve func(path string) (string, bool)) string {
	return cardShortcodeRe.ReplaceAllStringFunc(htmlStr, func(match string) string {
		sub := cardShortcodeRe.FindStringSubmatch(match)
		if card, ok := resolve(sub[1]); ok {
			return card
		}
		return ""
	})
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./Services -v`
Expected: PASS (new card tests + existing page tests).

- [ ] **Step 6: Commit**

```bash
git add backend/Models/Pages.Model.go backend/Services/Card.Service.go backend/Services/Card.Service_test.go
git commit -m "feat(backend): card metadata helpers + PageModel tag/summary/image fields

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 2: CardService impl + PageService lookups + controller wiring

This task changes `Create`/`Update` signatures, so it must leave the whole module compiling and green in one pass (Go builds the module as a whole). It has DB code that is not unit-tested; the membership/expansion logic it calls is already covered by Task 1, and its end-to-end behaviour is verified with a manual curl.

**Files:**
- Modify: `backend/Models/Panel.Model.go`
- Modify: `backend/Services/Page.Service.go`
- Modify: `backend/Services/Card.Service.go`
- Modify: `backend/Controllers/Page.Controller.go`
- Modify: `backend/Controllers/Panel.Controller.go`
- Modify: `backend/server.go`

**Interfaces:**
- Consumes: Task 1 pure helpers; `MenuService.GetByPageName`; `PageService`.
- Produces:
  - `models.PageWrite{ PageName, Path, Source, ViewType string; Tags []string; Summary, Image string; ListTags []string }`
  - `PageService.Create(w models.PageWrite) error`
  - `PageService.Update(name string, w models.PageWrite) error`
  - `PageService.GetRawByPath(path string) (models.PageModel, error)`
  - `PageService.FindByTags(tags []string) ([]models.PageModel, error)`
  - `CardService` interface: `ExpandShortcodes(htmlStr string) (string, error)`, `ExpandCards(htmlStr string, page models.PageModel) (string, error)`; constructor `NewCardService(pages PageService, menus MenuService) CardService`
  - `controllers.InitPageController(pageService, cardService, apiGroup)`
  - `controllers.InitPanelController(pageService, menuService, cardService, tokenService, apiGroup)`

- [ ] **Step 1: Add PageWrite and extend the panel DTOs**

In `backend/Models/Panel.Model.go`:
```go
type PageWrite struct {
	PageName string
	Path     string
	Source   string // clean newlines
	ViewType string
	Tags     []string
	Summary  string
	Image    string
	ListTags []string
}
```
Extend the existing structs with the four fields (add to `PageDetail`, `CreatePageRequest`, `UpdatePageRequest`):
```go
	Tags     []string `json:"Tags"`
	Summary  string   `json:"Summary"`
	Image    string   `json:"Image"`
	ListTags []string `json:"ListTags"`
```

- [ ] **Step 2: Update PageService interface + Create/Update + add lookups**

In `backend/Services/Page.Service.go`, change the interface entries:
```go
	Create(w models.PageWrite) error
	Update(name string, w models.PageWrite) error
	GetRawByPath(path string) (models.PageModel, error)
	FindByTags(tags []string) ([]models.PageModel, error)
```
Rewrite `Create` to persist the new fields:
```go
func (psi *PageServiceImplementation) Create(w models.PageWrite) error {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{{Key: "PageName", Value: w.PageName}})
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
		{Key: "PageName", Value: w.PageName},
		{Key: "Path", Value: w.Path},
		{Key: "Page", Value: ToStorage(w.Source)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: w.ViewType},
		{Key: "Order", Value: total},
		{Key: "Hidden", Value: false},
		{Key: "Tags", Value: w.Tags},
		{Key: "Summary", Value: w.Summary},
		{Key: "Image", Value: w.Image},
		{Key: "ListTags", Value: w.ListTags},
	})
	return err
}
```
Rewrite `Update`'s `$set` to include the new fields (keep the existing `Path`, `Page`, `ViewType`, `Hash`, `Text` resets):
```go
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Path", Value: w.Path},
			{Key: "Page", Value: ToStorage(w.Source)},
			{Key: "ViewType", Value: w.ViewType},
			{Key: "Hash", Value: []byte{}},
			{Key: "Text", Value: ""},
			{Key: "Tags", Value: w.Tags},
			{Key: "Summary", Value: w.Summary},
			{Key: "Image", Value: w.Image},
			{Key: "ListTags", Value: w.ListTags},
		}}},
```
Add the two lookups (place near `GetPageByPath`):
```go
// GetRawByPath fetches a page by Path without rendering/caching — used to read
// card metadata off a referenced page.
func (psi *PageServiceImplementation) GetRawByPath(path string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "Path", Value: path}}).Decode(&page)
	return page, err
}

// FindByTags returns candidate pages whose Tags intersect tags (Mongo $in),
// Order-sorted, raw (source included). Authoritative membership is selectByTags.
func (psi *PageServiceImplementation) FindByTags(tags []string) ([]models.PageModel, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	opts := options.Find().SetSort(bson.D{{Key: "Order", Value: 1}})
	cursor, err := psi.collection.Find(psi.ctx,
		bson.D{{Key: "Tags", Value: bson.D{{Key: "$in", Value: tags}}}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var pages []models.PageModel
	if err := cursor.All(psi.ctx, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}
```

- [ ] **Step 3: Add the CardService interface + impl**

Append to `backend/Services/Card.Service.go` (add `strings` is already imported; no new imports needed):
```go
type CardService interface {
	ExpandShortcodes(htmlStr string) (string, error)
	ExpandCards(htmlStr string, page models.PageModel) (string, error)
}

type CardServiceImplementation struct {
	pages PageService
	menus MenuService
}

func NewCardService(pages PageService, menus MenuService) CardService {
	return &CardServiceImplementation{pages: pages, menus: menus}
}

// cardFor renders a fetched page as a card, applying overrides then auto-extract.
func (c *CardServiceImplementation) cardFor(page models.PageModel) string {
	caption := ""
	if m, err := c.menus.GetByPageName(page.PageName); err == nil {
		caption = m.Caption
	}
	source := FromStorage(page.Page)
	summary := page.Summary
	if summary == "" {
		summary = extractSummary(source)
	}
	image := page.Image
	if image == "" {
		image = extractImage(source)
	}
	return buildCardHTML(page.Path, cardTitle(caption, page.PageName), summary, image)
}

func (c *CardServiceImplementation) resolveCard(path string) (string, bool) {
	page, err := c.pages.GetRawByPath(path)
	if err != nil {
		return "", false
	}
	return c.cardFor(page), true
}

func (c *CardServiceImplementation) ExpandShortcodes(htmlStr string) (string, error) {
	return expandShortcodes(htmlStr, c.resolveCard), nil
}

func (c *CardServiceImplementation) ExpandCards(htmlStr string, page models.PageModel) (string, error) {
	out, err := c.ExpandShortcodes(htmlStr)
	if err != nil {
		return htmlStr, err
	}
	if page.ViewType != "List" {
		return out, nil
	}
	candidates, err := c.pages.FindByTags(page.ListTags)
	if err != nil {
		return out, err
	}
	selected := selectByTags(candidates, page.ListTags, page.PageName)
	if len(selected) == 0 {
		return out, nil
	}
	var grid strings.Builder
	grid.WriteString(`<div class="card-grid">`)
	for _, p := range selected {
		grid.WriteString(c.cardFor(p))
	}
	grid.WriteString(`</div>`)
	return out + grid.String(), nil
}
```

- [ ] **Step 4: Wire CardService into the page controller**

In `backend/Controllers/Page.Controller.go`, add the dependency and expand in `GetPage`:
```go
type PageController struct {
	PageService services.PageService
	CardService services.CardService
}

func InitPageController(PageService services.PageService, CardService services.CardService, server *gin.RouterGroup) PageController {
	pc := PageController{PageService: PageService, CardService: CardService}
	server.GET("/Page", pc.GetPage)
	return pc
}
```
At the end of `GetPage`, replace the final `ctx.JSON(...)` block with:
```go
	final := respons.Text
	if expanded, exErr := pc.CardService.ExpandCards(respons.Text, respons); exErr == nil {
		final = expanded
	}
	ctx.JSON(http.StatusOK, gin.H{
		"ViewType": respons.ViewType,
		"Page":     final,
	})
```

- [ ] **Step 5: Wire CardService into the panel controller**

In `backend/Controllers/Panel.Controller.go`:
- Add `CardService services.CardService` to the struct and to `InitPanelController`'s parameters + literal.
- In `CreatePage`, replace the `PageService.Create(...)` call with:
```go
	if err := pc.PageService.Create(models.PageWrite{
		PageName: req.PageName, Path: req.Path, Source: req.Source, ViewType: req.ViewType,
		Tags: req.Tags, Summary: req.Summary, Image: req.Image, ListTags: req.ListTags,
	}); err != nil {
```
- In `UpdatePage`, replace the `PageService.Update(...)` call with:
```go
	if err := pc.PageService.Update(name, models.PageWrite{
		Path: req.Path, Source: req.Source, ViewType: req.ViewType,
		Tags: req.Tags, Summary: req.Summary, Image: req.Image, ListTags: req.ListTags,
	}); err != nil {
```
- In `GetPage`, add the four fields to the returned `PageDetail`:
```go
	detail := models.PageDetail{
		PageName: page.PageName,
		Path:     page.Path,
		Source:   services.FromStorage(page.Page),
		ViewType: page.ViewType,
		Tags:     page.Tags,
		Summary:  page.Summary,
		Image:    page.Image,
		ListTags: page.ListTags,
	}
```
- In `Preview`, expand shortcodes before returning:
```go
	html, err := pc.PageService.Preview(req.Source)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if expanded, exErr := pc.CardService.ExpandShortcodes(html); exErr == nil {
		html = expanded
	}
	ctx.JSON(http.StatusOK, gin.H{"Html": html})
```

- [ ] **Step 6: Update the composition root**

In `backend/server.go` `InitControllers`, build the card service and update the two wiring calls:
```go
	cardService := services.NewCardService(pageService, menuService)

	controllers.InitUserController(userService, tokenService, apiGroup)
	controllers.InitMenuController(menuService, pageService, apiGroup)
	controllers.InitPageController(pageService, cardService, apiGroup)
	controllers.InitPanelController(pageService, menuService, cardService, tokenService, apiGroup)
```

- [ ] **Step 7: Build and run the existing suite**

Run: `cd backend && go build ./... && go test ./...`
Expected: builds clean; all tests PASS (Task 1 card tests, page tests, controller tests).

- [ ] **Step 8: Manual curl verification**

With MongoDB up and the backend running (`air` or `go run .`), and a page whose source contains a `<card path="/…">` pointing to an existing page:
```bash
curl -s 'http://localhost:8080/api/Page?Path=/your-static-list-page' | head -c 800
```
Expected: the response `Page` HTML contains `<a class="card"` (shortcode expanded), and pointing at a page that does not exist leaves no raw `<card` in the output. For a `ViewType=List` page whose `ListTags` match some tagged pages, the output ends with a `<div class="card-grid">…</div>`.

- [ ] **Step 9: Commit**

```bash
git add backend/
git commit -m "feat(backend): CardService expands <card> shortcodes and List tag grids

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 3: Frontend pure helpers (tags + internal-link nav)

**Files:**
- Create: `frontend/src/components/panelComponents/tags.ts`
- Test: `frontend/src/components/panelComponents/tags.spec.ts`
- Create: `frontend/src/components/contentLinks.ts`
- Test: `frontend/src/components/contentLinks.spec.ts`

**Interfaces:**
- Produces:
  - `parseTags(input: string): string[]`
  - `formatTags(tags: string[]): string`
  - `internalNavTarget(a: AnchorLike): string | null` where `AnchorLike = { getAttribute(name: string): string | null }`

- [ ] **Step 1: Write the failing tag tests**

Create `frontend/src/components/panelComponents/tags.spec.ts`:
```ts
import { describe, it, expect } from 'vitest'
import { parseTags, formatTags } from '@/components/panelComponents/tags'

describe('tags', () => {
  it('parses, trims, and drops empties', () => {
    expect(parseTags(' blog , go ,, ')).toEqual(['blog', 'go'])
  })
  it('formats with ", "', () => {
    expect(formatTags(['blog', 'go'])).toBe('blog, go')
  })
  it('round-trips', () => {
    expect(parseTags(formatTags(['a', 'b']))).toEqual(['a', 'b'])
  })
  it('handles empty/undefined', () => {
    expect(parseTags('')).toEqual([])
    expect(formatTags(undefined as unknown as string[])).toBe('')
  })
})
```

- [ ] **Step 2: Write the failing link tests**

Create `frontend/src/components/contentLinks.spec.ts`:
```ts
import { describe, it, expect } from 'vitest'
import { internalNavTarget } from '@/components/contentLinks'

function anchor(attrs: Record<string, string>) {
  return { getAttribute: (n: string) => (n in attrs ? attrs[n] : null) }
}

describe('internalNavTarget', () => {
  it('returns internal href with no target', () => {
    expect(internalNavTarget(anchor({ href: '/blog/x' }))).toBe('/blog/x')
  })
  it('ignores target=_blank (Markdown links)', () => {
    expect(internalNavTarget(anchor({ href: '/blog/x', target: '_blank' }))).toBeNull()
  })
  it('ignores external hrefs', () => {
    expect(internalNavTarget(anchor({ href: 'https://x.com' }))).toBeNull()
  })
  it('ignores missing href', () => {
    expect(internalNavTarget(anchor({}))).toBeNull()
  })
})
```

- [ ] **Step 3: Run to verify both fail**

Run: `cd frontend && npm run test:unit -- run tags contentLinks`
Expected: FAIL — modules not found.

- [ ] **Step 4: Implement the helpers**

Create `frontend/src/components/panelComponents/tags.ts`:
```ts
// Convert between the panel's comma-separated tag input and the string[] the API uses.
export function parseTags(input: string): string[] {
	return input.split(',').map((t) => t.trim()).filter(Boolean)
}

export function formatTags(tags: string[]): string {
	return (tags ?? []).join(', ')
}
```
Create `frontend/src/components/contentLinks.ts`:
```ts
export interface AnchorLike {
	getAttribute(name: string): string | null
}

// The path to SPA-navigate to for a clicked anchor, or null to let the browser
// handle it. Internal links (href starting with "/") that have no target are
// routed client-side; Markdown links carry target="_blank" and are left alone.
export function internalNavTarget(a: AnchorLike): string | null {
	const href = a.getAttribute('href') ?? ''
	const target = a.getAttribute('target') ?? ''
	if (href.startsWith('/') && target === '') return href
	return null
}
```

- [ ] **Step 5: Run to verify both pass**

Run: `cd frontend && npm run test:unit -- run tags contentLinks`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/panelComponents/tags.ts frontend/src/components/panelComponents/tags.spec.ts frontend/src/components/contentLinks.ts frontend/src/components/contentLinks.spec.ts
git commit -m "feat(frontend): tag parsing + internal-link nav helpers

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 4: Panel editor — ViewType dropdown + tag/summary/image/list-tag fields

**Files:**
- Modify: `frontend/src/types/PanelModels.ts`
- Modify: `frontend/src/service/Panel.service.ts`
- Modify: `frontend/src/global/panelState.ts`
- Modify: `frontend/src/components/panelComponents/PageEditor.vue`

**Interfaces:**
- Consumes: `parseTags`, `formatTags` (Task 3); backend `PageDetail`/`Create`/`Update` fields (Task 2).
- Produces: `SavePagePayload` with `Tags: string[]`, `Summary: string`, `Image: string`, `ListTags: string[]`; `PanelPageDetail` with the same four.

- [ ] **Step 1: Extend the detail type**

In `frontend/src/types/PanelModels.ts`, extend `PanelPageDetail`:
```ts
export interface PanelPageDetail {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Tags: string[]
	Summary: string
	Image: string
	ListTags: string[]
	Menu: MenuBinding | null
}
```

- [ ] **Step 2: Extend the save payload**

In `frontend/src/service/Panel.service.ts`, extend `SavePagePayload`:
```ts
export interface SavePagePayload {
	PageName: string
	Path: string
	Source: string
	ViewType: string
	Tags: string[]
	Summary: string
	Image: string
	ListTags: string[]
	Menu: MenuBinding | null
}
```

- [ ] **Step 3: Seed the new fields for a new page**

In `frontend/src/global/panelState.ts`, update `startNew()`:
```ts
export function startNew(): void {
	state.selected = {
		PageName: '', Path: '', Source: '', ViewType: 'PlainHTML',
		Tags: [], Summary: '', Image: '', ListTags: [],
		Menu: { Name: '', Caption: '' },
	}
	state.isNew = true
	state.dirty = false
}
```

- [ ] **Step 4: Wire the editor UI + save**

In `frontend/src/components/panelComponents/PageEditor.vue`:

Add imports:
```ts
import Dropdown from 'primevue/dropdown'
import { parseTags, formatTags } from '@/components/panelComponents/tags'
```
Add local refs (near `const source = ref('')`):
```ts
const tagsInput = ref('')
const listTagsInput = ref('')
```
In the `watch(() => state.selected, …)` callback, seed the inputs (add these two lines before `runPreview()`):
```ts
	tagsInput.value = formatTags(sel?.Tags ?? [])
	listTagsInput.value = formatTags(sel?.ListTags ?? [])
```
Replace the existing **View type** field block in the template:
```vue
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">View type</label>
					<Dropdown v-model="state.selected.ViewType" :options="['PlainHTML', 'List']"
						class="bg-surface-2 border border-border text-fg rounded" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Tags</label>
					<InputText v-model="tagsInput" placeholder="blog, go"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div v-if="state.selected.ViewType === 'List'" class="flex flex-col">
					<label class="text-sm text-muted mb-1">List tags</label>
					<InputText v-model="listTagsInput" placeholder="blog"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Card summary (override)</label>
					<InputText v-model="state.selected.Summary"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Card image (override)</label>
					<InputText v-model="state.selected.Image"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
```
In `save()`, build the shared field set and include it in both payloads. Replace the `createPage`/`updatePage` calls:
```ts
		const fields = {
			Path: state.selected.Path,
			Source: source.value,
			ViewType: state.selected.ViewType,
			Tags: parseTags(tagsInput.value),
			Summary: state.selected.Summary ?? '',
			Image: state.selected.Image ?? '',
			ListTags: parseTags(listTagsInput.value),
			Menu: hasMenu ? menu.value : null,
		}
		if (isNew) {
			await PanelService.createPage({ PageName: name, ...fields })
		} else {
			await PanelService.updatePage(name, fields)
		}
```

- [ ] **Step 5: Type-check**

Run: `cd frontend && npm run type-check`
Expected: no errors.

- [ ] **Step 6: Manual panel check**

Run the app (DB + backend + `npm run dev`), sign into `/panel`. Create/select a page: the View type is now a dropdown (`PlainHTML`/`List`); a **Tags** field is present; selecting `List` reveals **List tags**; **Card summary/image (override)** fields save and reload correctly (edit, Save, reselect, values persist).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/types/PanelModels.ts frontend/src/service/Panel.service.ts frontend/src/global/panelState.ts frontend/src/components/panelComponents/PageEditor.vue
git commit -m "feat(panel): ViewType dropdown + tags/summary/image/list-tag fields

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 5: Public page — SPA card links + card styles

**Files:**
- Modify: `frontend/src/views/contents.vue`
- Modify: `frontend/src/assets/content.css`

**Interfaces:**
- Consumes: `internalNavTarget` (Task 3); `useRouter` (vue-router).

- [ ] **Step 1: Intercept internal-link clicks in contents.vue**

In `frontend/src/views/contents.vue`, add to the imports/setup:
```ts
import { useRouter } from 'vue-router'
import { internalNavTarget } from '@/components/contentLinks'

const router = useRouter()

function onContentClick(e: MouseEvent) {
	const el = e.target as HTMLElement
	const a = el.closest('a')
	if (!a) return
	const to = internalNavTarget(a)
	if (to) {
		e.preventDefault()
		router.push(to)
	}
}
```
Bind the handler on the content wrapper:
```vue
		<div class="content" v-html="returnedHTML" @click="onContentClick"></div>
```

- [ ] **Step 2: Verify SPA navigation manually**

With the app running, open a page containing cards. Clicking a card navigates to the referenced page **without a full reload** (no browser refresh flash; URL updates and the view swaps). External Markdown links still open in a new tab.

- [ ] **Step 3: Add the card styles**

Append to `frontend/src/assets/content.css`:
```css

/* Card system: a page rendered as a linked horizontal card (ViewType/List). */
.content .card-grid { display: flex; flex-direction: column; gap: 1em; margin: 1.4em 0; }
.content .card {
  display: flex; gap: 1em; align-items: stretch;
  border: 1px solid var(--border); border-radius: 12px; overflow: hidden;
  background: var(--surface-2); text-decoration: none; color: var(--fg);
  transition: border-color .2s, transform .2s;
}
.content .card:hover { border-color: var(--accent); transform: translateY(-2px); }
.content .card-img {
  width: 160px; flex: 0 0 auto; object-fit: cover;
  border-radius: 0; margin: 0; align-self: stretch;
}
.content .card-body { padding: .9em 1.1em; display: flex; flex-direction: column; gap: .4em; min-width: 0; }
.content .card-title { margin: 0; font-size: 1.15rem; color: var(--fg); }
.content .card-summary {
  margin: 0; color: var(--muted); font-size: .95rem;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
@media (max-width: 640px) {
  .content .card { flex-direction: column; }
  .content .card-img { width: 100%; height: 160px; }
}
```
(The generic `.content img { display:block; margin:1.4em 0 }` rule is overridden here for `.card-img`.)

- [ ] **Step 4: Verify styling manually**

Reload a card page: cards are horizontal rectangles (image left, title + 2-line clamped summary right), hover lifts them, and they stack vertically on a narrow viewport. Check both light and dark themes.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/contents.vue frontend/src/assets/content.css
git commit -m "feat(frontend): SPA navigation for card links + card styles

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Task 6: Documentation

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update the README backlog**

In `README.md`, mark the ViewType item done and drop `DynamicList`, e.g.:
```
[x] Paneldeki ViewType Text degil combobox olacak. Secenekler 'PlainHTML' ve 'List'. (DynamicList iptal — List, tag'e gore listeler.)
[ ] (opsiyonel) Inline <cardlist tag> shortcode'u.
```
(Keep the existing Turkish tone of the backlog; this is the one grandfathered-Turkish file.)

- [ ] **Step 2: Update CLAUDE.md**

Reflect the shipped behaviour (English), touching:
- **Purpose & design intent / Work in progress:** `ViewType` is no longer just a hook — it is a combobox (`PlainHTML`/`List`); a `List` page lists tag-matched pages as cards; the `DynamicList` idea is folded into `List`.
- **Page rendering:** document the `<card path="…">` shortcode, that cards are expanded at request time on top of the cached Markdown render (not cached), card data auto-extracted (title = caption/PageName, summary = first paragraph, image = first image) with `Summary`/`Image` overrides, and the new `Tags`/`ListTags` fields.
- **API endpoints:** `GET /api/Page` now returns card-expanded HTML; `POST /Preview` expands `<card>` shortcodes (not the List grid); Create/Update carry `Tags`/`Summary`/`Image`/`ListTags`.
- **Backend architecture:** add `CardService` (new resource: `Services/Card.Service.go`) and the `PageService.GetRawByPath`/`FindByTags` lookups; note `Create`/`Update` now take `models.PageWrite`.
- **Frontend architecture:** `contents.vue` intercepts internal links for SPA nav; the editor's ViewType is a PrimeVue **v3** `Dropdown`; tag inputs use `components/panelComponents/tags.ts`.
- **Commands:** add the new pure test file to the backend test list (`Services/Card.Service_test.go`) and the two new frontend specs (`components/panelComponents/tags.spec.ts`, `components/contentLinks.spec.ts`).

- [ ] **Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: document the ViewType card/list system

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Self-review notes (for the executor)

- **Spec coverage:** ViewType combobox → Task 4; `<card>` shortcode engine → Tasks 1–2; List tag grid (OR, self-exclude, Order) → Tasks 1–2; auto-extract + overrides → Tasks 1–2; SPA links → Tasks 3, 5; card styles → Task 5; tests → Tasks 1, 3; docs → Task 6.
- **Preview** expands only `<card>` shortcodes, never the List grid (Task 2, Step 5) — matches the spec.
- **No recursion:** cards read a referenced page's raw source (`GetRawByPath`) and only extract metadata; they never expand that page's own shortcodes.
- **Back-compat:** all new `PageModel`/DTO fields are optional; existing pages need no migration (untagged pages simply never match a List and render as before).
- **Do not commit** anything until the owner authorises it — the commit steps above are executed under the owner's git flow, not on your own initiative.
