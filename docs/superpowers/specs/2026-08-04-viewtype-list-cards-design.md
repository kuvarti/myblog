# ViewType combobox + card/list system

**Date:** 2026-08-04
**Status:** Approved
**Scope:** Turn `ViewType` from a free-text field into a combobox and give it real meaning by adding a reusable "page → card" engine. A card renders another page as a linked rectangle (title + summary + image). Two consumers share the one engine: a `<card path="…">` shortcode for hand-curated (static) lists inside any page, and a `ViewType = List` page that auto-lists pages by tag. This also introduces per-page tags, folding the backlog's separate "DynamicList needs tags" item into `List`.

## Problem

`PageModel.ViewType` is returned by `/api/Page` and seeded with values like `"PlainHTML"`, but the frontend never branches on it — every page renders identically (`contents.vue` dumps the rendered HTML into a `.content` wrapper via `v-html`). In the panel it is an ordinary text input. The README backlog wants `ViewType` to become a combobox and to support a `List` render mode (originally `List` + `DynamicList`).

We want a page to be able to surface *other* pages as cards: a title, a short summary, and an image, each linking to the referenced page — with that per-card data living on the referenced page itself, so a list "just pulls the pages' info". Two authoring needs: a **static** curated set (author names specific pages) and a **dynamic** set (author names tags, matching pages appear automatically).

## Decisions (approved)

1. **`DynamicList` is dropped.** The `ViewType` combobox offers exactly `PlainHTML` and `List` for now (extensible later). The backlog's separate tag item folds into `List`.
2. **One "page → card" engine, two entry points.** A card is always "render page X as a linked card". Static: a `<card path="/x">` shortcode placed in any page. Dynamic: a `ViewType = List` page whose tag filter selects the pages.
3. **Card data is auto-extracted with optional overrides.** Title = the page's menu `Caption` (fallback `PageName`). Summary = the page's `Summary` override if set, else the first text block of its source (~160 chars). Image = the page's `Image` override if set, else the first `<img src>` in its source.
4. **Tags are a first-class per-page field** (`Tags []string`). A `List` page carries a **separate** filter field (`ListTags []string`) — chosen over overloading `Tags` so "a page's tags" means the same thing everywhere.
5. **List match is OR across tags.** A page is listed if `Tags` shares **at least one** tag with the List page's `ListTags`. The List page excludes itself. Results are ordered by `Order`. Menu `Hidden` is ignored (visibility governs the nav, not list membership).
6. **Cards render on the backend at request time, and are not cached.** The Markdown→HTML cache (`Text`/`Hash`) is unchanged; a separate expansion pass runs on every `GET /api/Page`, so edits to a *referenced* page show up immediately (no stale-card cache). Card metadata is read from the referenced page's **raw** document, so a card never expands the referenced page's own cards — no recursion.
7. **The frontend stays `v-html`.** It does not branch on `ViewType`; the backend returns final HTML. The only frontend addition is SPA navigation for internal links (so cards don't full-reload).
8. **Preview expands `<card>` shortcodes** (read-only lookups) for WYSIWYG; it does **not** render the dynamic List grid (that depends on the saved `ViewType`/`ListTags`).

## Data model

`PageModel` gains four optional fields (absent → zero value → old behaviour, so no migration is required):

- `Tags []string` `bson:"Tags"` — the page's own tags.
- `Summary string` `bson:"Summary"` — card-summary override; empty → auto-extract.
- `Image string` `bson:"Image"` — card-image override; empty → auto-extract.
- `ListTags []string` `bson:"ListTags"` — only meaningful when `ViewType == "List"`; the tags this page lists.

## Authoring model

- **Static list** — a `PlainHTML` page whose source contains one `<card path="/blog/yazi1">` per card (own line). Each is replaced by that page's card at request time. An unresolved path (no such page) is dropped (replaced with empty) so no raw shortcode leaks.
- **Dynamic list** — a page with `ViewType = List` and one or more `ListTags`. Its own source renders normally (intro text allowed); a card grid for the tag-matched pages is appended after it.
- **Card contents** come from the referenced page (Decision 3). The author controls a card by editing the referenced page (its caption/first paragraph/first image), or by setting that page's `Summary`/`Image` overrides.

### Shortcode syntax

`<card path="/some/path">` — matched by `<card\s+path="([^"]*)"\s*/?>`. It survives Markdown rendering unchanged (a line containing `<` is passed through raw by `GetPageText`), so expansion is a string replace over the rendered `Text`.

### Card HTML (styled by `content.css`)

```html
<a class="card" href="/blog/yazi1">
  <img class="card-img" src="/kapak.jpg" alt="">
  <div class="card-body">
    <h3 class="card-title">Yazı Başlığı</h3>
    <p class="card-summary">özet metni…</p>
  </div>
</a>
```

The image element is omitted when there is no override and no image in the source. A List grid wraps its cards in `<div class="card-grid">…</div>`.

## Backend

### Models
- `Models/Pages.Model.go` — add `Tags []string`, `Summary string`, `Image string`, `ListTags []string` (with `bson` tags as above).
- `Models/Panel.Model.go`:
  - `PageDetail` gains `Tags`, `Summary`, `Image`, `ListTags` (the editor reads them).
  - `CreatePageRequest` and `UpdatePageRequest` gain the same four fields.
  - `PageSummary` is unchanged (the panel list view does not need them).

### Page service (`Services/Page.Service.go`)
- `Create` / `Update` persist the four new fields alongside the existing ones. Their signatures grow (e.g. `Create(name, path, sourceClean, viewType string, tags []string, summary, image string, listTags []string) error`) — or take a small `PageWrite` struct to avoid a long argument list (decided in the plan).
- Add `GetRawByPath(path string) (models.PageModel, error)` — an un-rendered lookup by `Path` (for card metadata; does **not** go through `renderAndCache`).
- Add `FindByTags(tags []string) ([]models.PageModel, error)` — candidate pages whose `Tags` intersect `tags` (Mongo `$in`), ordered by `Order`, returned raw (source included) for metadata extraction. This narrows at the DB; the authoritative membership rule lives in the pure `selectByTags` helper below.
- The existing render/cache path is untouched.

### Card service (`Services/Card.Service.go`, new)
Orchestrates lookups around a set of **pure, unit-testable helpers**, mirroring how `buildNav` is pure and the controller wires the DB:

- Pure helpers:
  - `extractSummary(source string) string` — first non-empty text block of the raw source with HTML tags and heading markers stripped, truncated to ~160 chars with an ellipsis.
  - `extractImage(source string) string` — `src` of the first `<img …>` in the source, else `""`.
  - `cardTitle(caption, pageName string) string` — caption if non-empty, else pageName.
  - `buildCardHTML(path, title, summary, image string) string` — the anchor markup above (omits `<img>` when image is `""`).
  - `selectByTags(pages []models.PageModel, listTags []string, selfName string) []models.PageModel` — the authoritative List-membership rule as a pure function: OR intersection with `listTags`, excludes `selfName`, preserves incoming `Order` sort. The service applies it to `FindByTags` results, so membership is unit-testable without Mongo.
  - `expandShortcodes(html string, resolve func(path string) (cardHTML string, ok bool)) string` — regex-replace each `<card path="…">`; drop on `!ok`.
- Interface (both methods used by controllers):
  - `ExpandShortcodes(html string) (string, error)` — resolves each `<card path>` (via `GetRawByPath` + `MenuService.GetByPageName` for the caption + overrides/extraction) into a card. Used by both the page fetch and Preview.
  - `ExpandCards(html string, page models.PageModel) (string, error)` — calls `ExpandShortcodes`, then, when `page.ViewType == "List"`, builds a `card-grid` from `FindByTags(page.ListTags)` minus `page.PageName` and appends it.
- Depends on `PageService` (for `GetRawByPath` / `FindByTags`) and `MenuService` (for captions).

### Page controller (`Controllers/Page.Controller.go`)
- Gains a `CardService` dependency (updated `InitPageController` signature + `server.go` wiring).
- `GetPage`: after `PageService` returns the page, call `CardService.ExpandCards(page.Text, page)` and return the expanded HTML as `Page` (DTO shape `{ViewType, Page}` unchanged). On expansion error, fall back to the un-expanded `Text` rather than 502.

### Panel controller (`Controllers/Panel.Controller.go`)
- `CreatePage` / `UpdatePage`: bind and pass the four new fields through to the service. Path validation is unchanged.
- `GetPage` detail returns the four new fields.
- `Preview`: after `PageService.Preview`, call `CardService.ExpandShortcodes(html)` so the editor preview shows real cards. The List grid is not previewed.

### Wiring (`server.go`)
`InitControllers` constructs `CardService` from the page + menu services (both already built there) and injects it into `PageController` and `PanelController`.

## Frontend

### Types & service
- `types/PanelModels.ts` — `PanelPageDetail` gains `Tags: string[]`, `Summary: string`, `Image: string`, `ListTags: string[]`. The save payload type gains the same.
- `service/Panel.service.ts` — `createPage`/`updatePage` send the new fields.
- `types/PageResponseModal.ts` — unchanged (still `{ Page, ViewType }`; `Page` is now the card-expanded HTML).

### Contents (`views/contents.vue`)
Add a delegated click handler on the `.content` wrapper for SPA navigation of internal links (cards and any internal Markdown link that has no `target`):
```ts
function onContentClick(e: MouseEvent) {
  const a = (e.target as HTMLElement).closest('a')
  if (!a) return
  const href = a.getAttribute('href') ?? ''
  if (href.startsWith('/') && !a.target) { e.preventDefault(); router.push(href) }
}
```
Markdown links keep `target="_blank"` (`HrefTargetBlank`) and are left alone; card links have no `target`, so they route client-side.

### Editor (`components/panelComponents/PageEditor.vue`)
- Replace the `ViewType` `InputText` with a PrimeVue `Select` (options `PlainHTML`, `List`).
- Add a **Tags** input (comma-separated → `string[]`), a **Summary override** input, and an **Image override** input.
- Show a **List tags** input only when `ViewType === 'List'`.
- `save()` includes the four new fields in the create/update payload.

### Styling (`assets/content.css`)
Style the card system under `.content` with descendant selectors (the single-definitive-stylesheet rule): `.card` (flex row, bordered, hover), `.card-img` (fixed-width left thumb, `object-fit: cover`), `.card-body`, `.card-title`, `.card-summary` (clamped), `.card-grid` (vertical stack of horizontal cards). Theme-aware via the existing `var(--…)` tokens.

## Testing

- **Go unit (`Services/Card.Service_test.go`, pure — no Mongo/HTTP):**
  - `extractSummary` — plain first block, HTML/heading stripping, truncation + ellipsis, empty source.
  - `extractImage` — first `<img>` found, none, multiple (first wins).
  - `cardTitle` — caption vs `PageName` fallback.
  - `selectByTags` — OR match, self-exclusion, order preserved, empty `listTags` → empty.
  - `expandShortcodes` — single/multiple replacements, unresolved path dropped, no shortcode → unchanged.
- **Frontend unit (Vitest):**
  - `contents` click interception — internal `href` (no target) → `router.push` called + `preventDefault`; `target="_blank"` and external hrefs → not intercepted.
  - `PageEditor` — `List` selection reveals the List-tags input; payload carries the new fields.
- **Manual E2E:** a `PlainHTML` page with two `<card path>` lines renders two cards linking to those pages, image/summary pulled from each; editing a referenced page's first paragraph/image updates its card on next load; a `ViewType = List` page with `ListTags: [blog]` lists every page tagged `blog` (and only those), excluding itself, in `Order`; overrides beat auto-extract; clicking a card navigates within the SPA (no full reload) and the URL is bookmarkable.

## Out of scope

- Inline `<cardlist tag="…">` shortcode (dynamic lists are `ViewType = List` only for now).
- `DynamicList` view type (cancelled).
- List pagination, tag index/landing pages, and filtering list membership by menu `Hidden`.
- Custom per-list ordering beyond the global `Order`; caching the card-expanded HTML.
- Previewing the dynamic List grid inside the editor.
