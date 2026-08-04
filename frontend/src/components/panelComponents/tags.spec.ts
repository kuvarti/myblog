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
