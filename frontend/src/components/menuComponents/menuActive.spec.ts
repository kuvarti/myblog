import { describe, it, expect } from 'vitest'
import { isMenuItemActive } from './menuActive'

// Mirrors the real DB menu data.
const AnaSayfa = { PageName: 'MainPage', Path: '/' } // home page
const BlogYazilari = { Path: '/lists' } //            path-only link, no PageName
const Projelerim = { Path: '/' } //                   home stub, no PageName
const Iletisim = { Path: '/' } //                     home stub, no PageName
const SoLong = { PageName: 'SoLong', Path: '' } //    real page reached via PageName
const NavDemo = { PageName: 'StyleTest', Path: '/lists' } // real page (stale Path)

describe('isMenuItemActive', () => {
	it('on home ("/", active MainPage) lights the three "/" items only', () => {
		const active = 'MainPage'
		expect(isMenuItemActive(AnaSayfa, '/', active)).toBe(true)
		expect(isMenuItemActive(Projelerim, '/', active)).toBe(true)
		expect(isMenuItemActive(Iletisim, '/', active)).toBe(true)
		expect(isMenuItemActive(BlogYazilari, '/', active)).toBe(false)
		expect(isMenuItemActive(SoLong, '/', active)).toBe(false)
		expect(isMenuItemActive(NavDemo, '/', active)).toBe(false)
	})

	it('on a PageName page ("/", active SoLong) lights only that page', () => {
		const active = 'SoLong'
		expect(isMenuItemActive(SoLong, '/', active)).toBe(true)
		// the home items must go dim when a non-home page is shown
		expect(isMenuItemActive(AnaSayfa, '/', active)).toBe(false)
		expect(isMenuItemActive(Projelerim, '/', active)).toBe(false)
		expect(isMenuItemActive(Iletisim, '/', active)).toBe(false)
		expect(isMenuItemActive(NavDemo, '/', active)).toBe(false)
	})

	it('on a path route ("/lists") lights only the matching path link', () => {
		// a stale active PageName must not leak onto a non-"/" route
		const active = 'MainPage'
		expect(isMenuItemActive(BlogYazilari, '/lists', active)).toBe(true)
		expect(isMenuItemActive(NavDemo, '/lists', active)).toBe(false) // has PageName, not a path link
		expect(isMenuItemActive(AnaSayfa, '/lists', active)).toBe(false)
		expect(isMenuItemActive(Projelerim, '/lists', active)).toBe(false)
	})
})
