import { describe, it, expect } from 'vitest'
import { moveItem } from './menuReorder'

describe('moveItem', () => {
	it('moves an element down', () => {
		expect(moveItem(['a', 'b', 'c', 'd'], 0, 2)).toEqual(['b', 'c', 'a', 'd'])
	})

	it('moves an element up', () => {
		expect(moveItem(['a', 'b', 'c', 'd'], 3, 1)).toEqual(['a', 'd', 'b', 'c'])
	})

	it('is a no-op when from === to', () => {
		expect(moveItem(['a', 'b', 'c'], 1, 1)).toEqual(['a', 'b', 'c'])
	})

	it('returns a copy unchanged for out-of-range indices', () => {
		const src = ['a', 'b', 'c']
		expect(moveItem(src, -1, 2)).toEqual(['a', 'b', 'c'])
		expect(moveItem(src, 0, 9)).toEqual(['a', 'b', 'c'])
	})

	it('does not mutate the input array', () => {
		const src = ['a', 'b', 'c']
		moveItem(src, 0, 2)
		expect(src).toEqual(['a', 'b', 'c'])
	})
})
