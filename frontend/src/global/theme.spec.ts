import { describe, it, expect, beforeEach, vi } from 'vitest'
import { resolveInitialTheme, toggleTheme, getTheme } from './theme'

// jsdom's localStorage is unavailable on its default opaque origin, so provide
// a simple in-memory stand-in for the tests (the real browser has a real one).
const store: Record<string, string> = {}
vi.stubGlobal('localStorage', {
  getItem: (k: string) => (k in store ? store[k] : null),
  setItem: (k: string, v: string) => { store[k] = String(v) },
  removeItem: (k: string) => { delete store[k] },
  clear: () => { for (const k of Object.keys(store)) delete store[k] },
})

function mockMatchMedia(prefersDark: boolean) {
  window.matchMedia = vi.fn().mockImplementation((q: string) => ({
    matches: prefersDark, media: q, addEventListener: vi.fn(), removeEventListener: vi.fn(),
  })) as unknown as typeof window.matchMedia
}

describe('theme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
  })

  it('uses the stored theme when present, ignoring system preference', () => {
    localStorage.setItem('theme', 'light')
    mockMatchMedia(true) // system prefers dark
    expect(resolveInitialTheme()).toBe('light')
  })

  it('falls back to system preference when nothing is stored', () => {
    mockMatchMedia(true)
    expect(resolveInitialTheme()).toBe('dark')
  })

  it('toggle flips the attribute and persists the choice', () => {
    document.documentElement.setAttribute('data-theme', 'light')
    const next = toggleTheme()
    expect(next).toBe('dark')
    expect(getTheme()).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')
  })
})
