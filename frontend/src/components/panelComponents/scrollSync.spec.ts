import { describe, it, expect } from 'vitest'
import { splitBlocks, interpolateScroll, buildAnchorPair } from './scrollSync'

describe('splitBlocks', () => {
	it('groups a run of consecutive non-blank Markdown lines into one block', () => {
		expect(splitBlocks('a\nb\nc')).toEqual([0])
	})

	it('treats a blank line as a separator', () => {
		expect(splitBlocks('a\n\nb')).toEqual([0, 2])
	})

	it('makes each `<`-line its own block and breaks the run on both sides', () => {
		expect(splitBlocks('a\n<img>\nb')).toEqual([0, 1, 2])
		expect(splitBlocks('<img>\n<img>')).toEqual([0, 1])
	})

	it('returns no blocks for empty or blank-only source', () => {
		expect(splitBlocks('')).toEqual([])
		expect(splitBlocks('\n  \n\t')).toEqual([])
	})

	it('handles a mixed document', () => {
		const src = '# Title\n\npara one\npara two\n\n<img src=x>\nafter'
		//            0         1  2         3         4  5             6
		expect(splitBlocks(src)).toEqual([0, 2, 5, 6])
	})
})

describe('interpolateScroll', () => {
	const from = [0, 100, 200]
	const to = [0, 50, 300]

	it('returns the aligned anchor on an exact hit', () => {
		expect(interpolateScroll(from, to, 100)).toBe(50)
	})

	it('interpolates linearly inside a segment', () => {
		expect(interpolateScroll(from, to, 150)).toBe(175) // 50 + 0.5*(300-50)
	})

	it('clamps below the first and above the last anchor', () => {
		expect(interpolateScroll(from, to, -10)).toBe(0)
		expect(interpolateScroll(from, to, 999)).toBe(300)
	})

	it('does not divide by zero across coincident anchors', () => {
		const r = interpolateScroll([0, 50, 50, 100], [0, 10, 90, 100], 50)
		expect(Number.isFinite(r)).toBe(true)
		expect(r).toBe(90)
	})

	it('degenerate arrays are safe', () => {
		expect(interpolateScroll([], [], 40)).toBe(0)
		expect(interpolateScroll([5], [7], 40)).toBe(7)
	})
})

describe('buildAnchorPair', () => {
	it('brackets block anchors with a leading 0 and a trailing max sentinel', () => {
		const { editorTops, previewTops } = buildAnchorPair([12, 80], [0, 140], 300, 500)
		expect(editorTops).toEqual([0, 12, 80, 300])
		expect(previewTops).toEqual([0, 0, 140, 500])
	})

	it('truncates to the shorter anchor list so indices stay aligned', () => {
		const { editorTops, previewTops } = buildAnchorPair([10, 20, 30], [5], 100, 200)
		expect(editorTops).toEqual([0, 10, 100])
		expect(previewTops).toEqual([0, 5, 200])
	})

	it('with no blocks degrades to a plain top-to-bottom (proportional) mapping', () => {
		const { editorTops, previewTops } = buildAnchorPair([], [], 300, 500)
		expect(editorTops).toEqual([0, 300])
		expect(previewTops).toEqual([0, 500])
		// proportional midpoint
		expect(interpolateScroll(editorTops, previewTops, 150)).toBe(250)
	})

	it('clamps anchors into [0, max] so arrays stay non-decreasing', () => {
		const { editorTops } = buildAnchorPair([10, 999], [1, 2], 100, 100)
		expect(editorTops).toEqual([0, 10, 100, 100])
	})
})
