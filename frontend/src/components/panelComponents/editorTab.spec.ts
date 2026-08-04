import { describe, it, expect } from 'vitest'
import { applyTab, INDENT } from '@/components/panelComponents/editorTab'

describe('applyTab', () => {
	it('inserts an indent at the cursor when there is no selection', () => {
		const r = applyTab('ab', 1, 1, false)
		expect(r.value).toBe('a' + INDENT + 'b')
		expect(r.selStart).toBe(2)
		expect(r.selEnd).toBe(2)
	})

	it('outdents a single line by removing a leading tab', () => {
		const r = applyTab(INDENT + 'ab', 2, 2, true)
		expect(r.value).toBe('ab')
	})

	it('outdents up to four leading spaces when there is no tab', () => {
		const r = applyTab('    ab', 6, 6, true)
		expect(r.value).toBe('ab')
	})

	it('indents every line the selection spans', () => {
		const r = applyTab('a\nb', 0, 3, false)
		expect(r.value).toBe(INDENT + 'a\n' + INDENT + 'b')
	})

	it('outdents every line the selection spans', () => {
		const v = INDENT + 'a\n' + INDENT + 'b'
		const r = applyTab(v, 0, v.length, true)
		expect(r.value).toBe('a\nb')
	})
})
