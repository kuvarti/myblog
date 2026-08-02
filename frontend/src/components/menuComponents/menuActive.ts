import { HOME_PAGE } from '@/global/constants'

export interface MenuItemLike {
	PageName?: string
	Path?: string
}

/**
 * Decide whether a side-menu item should be highlighted as the active page.
 *
 * Every content page is displayed at the "/" route and is distinguished only by
 * the active PageName (see MenuItem's RouterRedirect + the ActivePage store), so
 * route.path alone cannot tell two pages apart. Therefore:
 *   - On "/", match by PageName. A "/"-link without its own PageName represents
 *     the home page, so it is active only while the home page is shown.
 *   - On any other route (e.g. /lists), match plain navigation links by Path;
 *     PageName-carrying items never light up here (they live at "/").
 */
export function isMenuItemActive(item: MenuItemLike, routePath: string, activePage: string): boolean {
	if (routePath === '/') {
		if (item.PageName) return item.PageName === activePage
		return (item.Path || '/') === '/' && activePage === HOME_PAGE
	}
	return !item.PageName && item.Path === routePath
}
