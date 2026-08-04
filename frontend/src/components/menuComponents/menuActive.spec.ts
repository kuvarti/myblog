import { describe, it, expect } from 'vitest'
import { isMenuItemActive } from './menuActive'

// Mirrors the real DB menu data: with per-page routing each item carries its
// own Path, and the active item is the one whose Path matches the route.
const AnaSayfa = { Path: '/' } //          home page
const BlogYazilari = { Path: '/lists' } // reserved list route
const SoLong = { Path: '/solong' } //      a normal content page
const StyleTest = { Path: '/styletest' } //another content page

describe('isMenuItemActive', () => {
	it('on home ("/") lights only the "/" item', () => {
		expect(isMenuItemActive(AnaSayfa, '/')).toBe(true)
		expect(isMenuItemActive(BlogYazilari, '/')).toBe(false)
		expect(isMenuItemActive(SoLong, '/')).toBe(false)
		expect(isMenuItemActive(StyleTest, '/')).toBe(false)
	})

	it('on a content route ("/solong") lights only that page', () => {
		expect(isMenuItemActive(SoLong, '/solong')).toBe(true)
		expect(isMenuItemActive(AnaSayfa, '/solong')).toBe(false)
		expect(isMenuItemActive(StyleTest, '/solong')).toBe(false)
	})

	it('on the reserved list route ("/lists") lights the matching link', () => {
		expect(isMenuItemActive(BlogYazilari, '/lists')).toBe(true)
		expect(isMenuItemActive(AnaSayfa, '/lists')).toBe(false)
		expect(isMenuItemActive(SoLong, '/lists')).toBe(false)
	})

	it('treats a Path-less item as the home page', () => {
		expect(isMenuItemActive({}, '/')).toBe(true)
		expect(isMenuItemActive({}, '/lists')).toBe(false)
	})
})
