# Preview / editor scroll sync (admin panel)

**Date:** 2026-08-02
**Status:** Approved design, ready for implementation planning
**Area:** `frontend` — admin panel page editor (`PageEditor.vue`)

## Problem

In the panel page editor the Markdown source (`<textarea>`) and the live
preview (`.content` div) scroll independently. While editing a longer page you
lose your place: the source you are looking at and the rendered output you are
looking at drift apart.

We want them synchronized so that **the block at the top of the editor viewport
lines up with the same block at the top of the preview viewport** (and the
reverse when the preview is scrolled).

## Why not plain proportional (ratio) sync

The obvious approach — mirror `scrollTop / (scrollHeight - clientHeight)` between
the panes — was rejected. Source and rendered heights do not correspond: a
single source line can render to something much taller (an image, a code block,
a long paragraph that soft-wraps). With a pure ratio the tops match only by
luck and drift as soon as such an element appears. The user explicitly asked for
content-aligned behavior ("top text of the editor at the top of the preview"),
not a ratio.

## Why not backend line-tagging

The most precise approach would be to tag rendered output with source line
numbers (`data-line`) and map exactly. We are **not** doing this:

- It touches the trickiest, recently-reviewed part of the backend (the
  render + SHA-1 cache pipeline in `Page.Service.go`) and the `/Preview`
  endpoint.
- The bespoke authoring format is block-based, not line-based, so per-line
  tagging is awkward.
- Lines containing `<` are emitted as **raw HTML verbatim**; wrapping or
  injecting attributes around user-authored HTML fragments risks corrupting it.

The content-aligned behavior we want is achievable entirely on the frontend
without this risk.

## Approach: anchored, piecewise-linear sync (frontend only)

Compute the vertical pixel offset of each **block boundary** in both panes, then
map a scroll position in one pane to the other by **piecewise-linear
interpolation** between corresponding anchors. Block `i`'s top in the editor maps
to block `i`'s top in the preview; positions between anchors interpolate
linearly, so scrolling stays smooth.

### Blocks

Blocks are the units we anchor on: each should correspond to **one rendered
top-level element** in the preview. The split is applied to the editor's
clean-newline source (the `/n` storage delimiter is backend-only; the textarea
and `/Preview` speak real newlines):

- Split the source on `\n`.
- A blank line is a separator (produces no block).
- A line **containing `<`** is its own block (raw-HTML passthrough) and breaks
  any current run on both sides.
- A maximal run of consecutive **non-blank** lines that contain no `<` is **one**
  block (gomarkdown joins its soft line breaks into a single element).

**Why this matches the preview but is not verbatim the backend rule.**
`GetPageText` batches a Markdown run and *all following blank lines* into a
single `ConvertmdToHTML` call, so one "backend block" can span several
paragraphs. But gomarkdown still emits **one element per blank-separated
paragraph** (a blank line is a Markdown paragraph break), so the rendered
`.content` children line up one-per-paragraph. Splitting the source on blank
lines (and on `<`-lines) therefore yields one block per rendered element — which
is exactly the anchor granularity we need, even though it does not mirror the
backend's internal batching.

`splitBlocks(source)` returns, for each block, the **source line index** where it
starts. This is the list of editor-side anchor lines. The preview-side anchors
are the rendered top-level elements, in order, so block `i` ≈ the `i`-th child of
`.content`. (Inline raw-HTML `<`-lines that render to a text node rather than an
element are the main source of index drift — see mismatch handling below.)

### Anchor positions

- **Preview side (easy):** the direct children of the `.content` element are the
  rendered blocks. Collect each child's top offset within the scroll container →
  `previewTops[i]`.
- **Editor side (`<textarea>` has no per-line geometry):** measure with a hidden
  **mirror** `<div>` that copies the textarea's font, width, padding and
  `white-space: pre-wrap` (so soft-wrapping matches). Insert zero-size marker
  spans at each block-start line and read their `offsetTop` in a single layout
  pass → `editorTops[i]`.

Both arrays get a trailing **sentinel** anchor equal to that pane's maximum
scroll (`scrollHeight - clientHeight`), so the region below the last block still
maps sensibly to the bottom instead of snapping.

### Pure interpolation core

```
interpolateScroll(fromTops, toTops, fromScroll) -> targetScroll
```

- Find segment `i` with `fromTops[i] <= fromScroll <= fromTops[i+1]`.
- `t = (fromScroll - fromTops[i]) / (fromTops[i+1] - fromTops[i])`, guarding a
  zero-length segment (`t = 0`).
- `target = toTops[i] + t * (toTops[i+1] - toTops[i])`.
- Below the first anchor → `toTops[0]`; above the last → last value.

This function is **pure** (arrays + a number in, a number out) and carries all
the math, so it is fully unit-testable without a DOM — matching the existing
`menuActive.ts` pattern (pure core + thin component wiring).

### Wiring in `PageEditor.vue`

- Template refs on the textarea and the preview scroll container.
- A `@scroll` handler on each pane: on scroll of pane A, set pane B's
  `scrollTop = interpolateScroll(Atops, Btops, A.scrollTop)`.
- **Feedback-loop guard:** setting B's `scrollTop` fires B's own `scroll` event.
  A `syncing` lock cleared on the next `requestAnimationFrame` makes that echoed
  event bail, preventing an A↔B loop.

### Re-measuring anchors

Anchor arrays are cached and recomputed when they can change:

- **Source change:** after the debounced preview re-renders — measured on the
  *next tick* so the new DOM is in place before reading `offsetTop`.
- **Resize:** a `ResizeObserver` on both panes recomputes on layout changes.

### Edge cases / graceful degradation

- **Not scrollable** (content fits): a pane's max scroll ≤ 0 → sync is a no-op.
- **Empty source / no blocks:** only the sentinels exist, so
  `interpolateScroll` reduces to a single `0 → maxScroll` segment — i.e. it
  degrades to plain proportional sync. Acceptable fallback.
- **Block/element count mismatch** (a Markdown block that renders to more than
  one element, or an empty render): align by index up to the shorter length; the
  trailing sentinel bounds the drift toward the bottom. Documented approximation,
  not a correctness bug.

## Testing

- **`splitBlocks(source)`** — Vitest units: a plain Markdown run, a `<`-line
  passthrough breaking a run, blank-line separators, and a mixed document;
  assert the returned start-line indices.
- **`interpolateScroll(fromTops, toTops, fromScroll)`** — Vitest units: exact
  anchor hit, mid-segment interpolation, below-first, above-last (tail sentinel),
  zero-length-segment guard, and the empty→linear fallback.
- **DOM parts** (mirror measurement, `offsetTop` collection, `@scroll` wiring +
  rAF lock) — verified in the browser; jsdom does no layout (`scrollHeight`/
  `offsetTop` are 0), so DOM-level scroll tests are not meaningful. Same split as
  `menuActive` (pure unit tests + browser E2E).

## Files

- **NEW** `frontend/src/components/panelComponents/scrollSync.ts` — `splitBlocks`,
  `interpolateScroll`, and the DOM measurement helpers.
- **NEW** `frontend/src/components/panelComponents/scrollSync.spec.ts` — unit
  tests for the pure functions.
- **MODIFIED** `frontend/src/components/panelComponents/PageEditor.vue` — refs,
  `@scroll` handlers, rAF lock, anchor caching + re-measure triggers.

## Out of scope

- Backend line-tagging / `data-line` output.
- Caret-following (preview jumps to the cursor as you type).
- Pixel-perfect per-element mapping; block-level anchoring with linear
  interpolation between anchors is the target fidelity.
