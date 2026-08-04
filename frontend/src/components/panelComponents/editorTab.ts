// Pure text transform behind the editor's Tab / Shift+Tab indent behaviour.
// Kept out of the component so it can be unit-tested without a DOM.

export const INDENT = '\t'

export interface TabResult {
	value: string
	selStart: number
	selEnd: number
}

// applyTab mimics an editor's Tab/Shift+Tab on a textarea's value + selection:
// - Tab with no selection: insert one INDENT at the cursor.
// - Tab with a selection, or Shift+Tab (outdent): indent/outdent every line the
//   selection touches. Outdent removes a leading tab, else up to four spaces.
export function applyTab(value: string, start: number, end: number, outdent: boolean): TabResult {
	if (!outdent && start === end) {
		const next = value.slice(0, start) + INDENT + value.slice(start)
		const pos = start + INDENT.length
		return { value: next, selStart: pos, selEnd: pos }
	}

	const lineStart = value.lastIndexOf('\n', start - 1) + 1
	const lines = value.slice(lineStart, end).split('\n')
	let firstDelta = 0
	let totalDelta = 0
	const out = lines.map((line, i) => {
		if (outdent) {
			let removed = 0
			if (line.startsWith(INDENT)) {
				removed = INDENT.length
			} else {
				const m = line.match(/^ {1,4}/)
				removed = m ? m[0].length : 0
			}
			if (i === 0) firstDelta = -removed
			totalDelta -= removed
			return line.slice(removed)
		}
		if (i === 0) firstDelta = INDENT.length
		totalDelta += INDENT.length
		return INDENT + line
	})

	return {
		value: value.slice(0, lineStart) + out.join('\n') + value.slice(end),
		selStart: Math.max(lineStart, start + firstDelta),
		selEnd: end + totalDelta,
	}
}
