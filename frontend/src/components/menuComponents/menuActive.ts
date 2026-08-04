export interface MenuItemLike {
	Path?: string
}

/**
 * Decide whether a side-menu item should be highlighted as the active page.
 *
 * With per-page routing every page owns a real URL (its Path), so the active
 * item is simply the one whose Path equals the current route. An item without a
 * Path defaults to "/" (the home page).
 */
export function isMenuItemActive(item: MenuItemLike, routePath: string): boolean {
	return (item.Path || '/') === routePath
}
