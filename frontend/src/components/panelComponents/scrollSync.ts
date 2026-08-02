// Content-aligned scroll sync between the panel editor (<textarea>) and the live
// preview (.content div). See docs/superpowers/specs/2026-08-02-preview-editor-
// scroll-sync-design.md. The pure functions (splitBlocks, interpolateScroll)
// carry all the logic and are unit-tested; the DOM helpers are browser-verified
// (jsdom does no layout, so scrollHeight/offsetTop are 0 there).

/**
 * Split source into the blocks we anchor on — one block per rendered top-level
 * element. Mirrors the preview's element structure (not the backend's internal
 * batching): a line containing `<` is its own block (raw-HTML passthrough); a
 * maximal run of consecutive non-blank lines with no `<` is one block; blank
 * lines separate blocks. Returns each block's start line index in the source.
 */
export function splitBlocks(source: string): number[] {
	const lines = source.split('\n')
	const starts: number[] = []
	let inRun = false
	for (let i = 0; i < lines.length; i++) {
		const line = lines[i]
		if (line.includes('<')) {
			starts.push(i) // each `<`-line is its own block
			inRun = false
		} else if (line.trim() === '') {
			inRun = false // blank line separates blocks
		} else if (!inRun) {
			starts.push(i) // first line of a Markdown run
			inRun = true
		}
	}
	return starts
}

/**
 * Map a scroll position in one pane to the other by piecewise-linear
 * interpolation between corresponding anchor offsets. `fromTops`/`toTops` must be
 * the same length, index-aligned, and non-decreasing. Positions before the first
 * / after the last anchor clamp to the endpoints.
 */
export function interpolateScroll(fromTops: number[], toTops: number[], fromScroll: number): number {
	const n = Math.min(fromTops.length, toTops.length)
	if (n === 0) return 0
	if (n === 1) return toTops[0]
	if (fromScroll <= fromTops[0]) return toTops[0]
	if (fromScroll >= fromTops[n - 1]) return toTops[n - 1]
	let i = 0
	while (i < n - 1 && fromTops[i + 1] <= fromScroll) i++
	const span = fromTops[i + 1] - fromTops[i]
	const t = span > 0 ? (fromScroll - fromTops[i]) / span : 0
	return toTops[i] + t * (toTops[i + 1] - toTops[i])
}

function escapeHtml(s: string): string {
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

const MIRROR_COPY_PROPS = [
	'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft',
	'fontFamily', 'fontSize', 'fontWeight', 'fontStyle',
	'lineHeight', 'letterSpacing', 'textTransform', 'wordSpacing', 'tabSize',
] as const

/**
 * Measure the top pixel offset (in the textarea's scrollTop coordinate space) of
 * each given start line, using a hidden mirror div that reproduces the
 * textarea's font, width and wrapping. Markers are inserted at each start line
 * and read in a single layout pass.
 */
export function measureEditorTops(textarea: HTMLTextAreaElement, startLines: number[]): number[] {
	if (startLines.length === 0) return []
	const cs = getComputedStyle(textarea)
	const padL = parseFloat(cs.paddingLeft) || 0
	const padR = parseFloat(cs.paddingRight) || 0

	const mirror = document.createElement('div')
	const s = mirror.style
	s.position = 'absolute'
	s.visibility = 'hidden'
	s.top = '0'
	s.left = '-9999px'
	s.boxSizing = 'content-box'
	s.width = Math.max(0, textarea.clientWidth - padL - padR) + 'px'
	s.whiteSpace = 'pre-wrap'
	s.overflowWrap = 'break-word'
	s.wordWrap = 'break-word'
	for (const p of MIRROR_COPY_PROPS) s[p as any] = cs[p as any]

	const startSet = new Set(startLines)
	const lines = textarea.value.split('\n')
	const parts: string[] = []
	for (let i = 0; i < lines.length; i++) {
		if (startSet.has(i)) parts.push(`<span data-anchor="${i}"></span>`)
		parts.push(escapeHtml(lines[i]))
		if (i < lines.length - 1) parts.push('\n')
	}
	mirror.innerHTML = parts.join('')

	document.body.appendChild(mirror)
	const base = mirror.getBoundingClientRect().top
	const tops = Array.from(mirror.querySelectorAll('span[data-anchor]')).map(
		(m) => (m as HTMLElement).getBoundingClientRect().top - base,
	)
	document.body.removeChild(mirror)
	return tops
}

/**
 * Top pixel offset (in the container's scrollTop coordinate space) of each
 * rendered top-level element inside the container's `.content` child.
 */
export function measurePreviewTops(container: HTMLElement): number[] {
	const content = container.querySelector('.content')
	if (!content) return []
	const cTop = container.getBoundingClientRect().top
	return Array.from(content.children).map(
		(el) => el.getBoundingClientRect().top - cTop + container.scrollTop,
	)
}

/**
 * Build index-aligned anchor arrays for interpolateScroll: a leading 0 (top),
 * the per-block anchors truncated to the shorter side, and a trailing sentinel
 * at each pane's max scroll. Anchors are clamped into [0, max] so both arrays
 * stay non-decreasing even when a pane is barely scrollable.
 */
export function buildAnchorPair(
	editorAnchors: number[],
	previewAnchors: number[],
	editorMax: number,
	previewMax: number,
): { editorTops: number[]; previewTops: number[] } {
	const m = Math.min(editorAnchors.length, previewAnchors.length)
	const clamp = (v: number, max: number) => Math.min(Math.max(v, 0), Math.max(max, 0))
	const editorTops = [0, ...editorAnchors.slice(0, m).map((v) => clamp(v, editorMax)), Math.max(editorMax, 0)]
	const previewTops = [0, ...previewAnchors.slice(0, m).map((v) => clamp(v, previewMax)), Math.max(previewMax, 0)]
	return { editorTops, previewTops }
}
