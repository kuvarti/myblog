import { describe, it, expect } from 'vitest'
import { internalNavTarget, interceptNavTarget } from '@/components/contentLinks'

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

const plainClick = { button: 0, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false, defaultPrevented: false }

describe('interceptNavTarget', () => {
  it('intercepts a plain left-click on an internal link', () => {
    expect(interceptNavTarget(plainClick, anchor({ href: '/x' }))).toBe('/x')
  })
  it('ignores modifier clicks', () => {
    expect(interceptNavTarget({ ...plainClick, metaKey: true }, anchor({ href: '/x' }))).toBeNull()
    expect(interceptNavTarget({ ...plainClick, ctrlKey: true }, anchor({ href: '/x' }))).toBeNull()
  })
  it('ignores middle-click', () => {
    expect(interceptNavTarget({ ...plainClick, button: 1 }, anchor({ href: '/x' }))).toBeNull()
  })
  it('ignores already-prevented events', () => {
    expect(interceptNavTarget({ ...plainClick, defaultPrevented: true }, anchor({ href: '/x' }))).toBeNull()
  })
  it('still ignores external/target links on a plain click', () => {
    expect(interceptNavTarget(plainClick, anchor({ href: '/x', target: '_blank' }))).toBeNull()
    expect(interceptNavTarget(plainClick, anchor({ href: 'https://x.com' }))).toBeNull()
  })
})
