export interface AnchorLike {
	getAttribute(name: string): string | null
}

// The path to SPA-navigate to for a clicked anchor, or null to let the browser
// handle it. Internal links (href starting with "/") that have no target are
// routed client-side; Markdown links carry target="_blank" and are left alone.
export function internalNavTarget(a: AnchorLike): string | null {
	const href = a.getAttribute('href') ?? ''
	const target = a.getAttribute('target') ?? ''
	if (href.startsWith('/') && target === '') return href
	return null
}

export interface ClickLike {
	button: number
	metaKey: boolean
	ctrlKey: boolean
	shiftKey: boolean
	altKey: boolean
	defaultPrevented: boolean
}

// The internal path to SPA-navigate to for a click on an anchor, or null to let
// the browser handle it. Only a plain left-click on an internal link is
// intercepted; modifier / middle clicks (open-in-new-tab etc.) pass through.
export function interceptNavTarget(e: ClickLike, a: AnchorLike): string | null {
	if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
		return null
	}
	return internalNavTarget(a)
}
