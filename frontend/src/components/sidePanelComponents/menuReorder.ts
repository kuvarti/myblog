// Pure list-reorder helper for the panel menu drag-and-drop. Returns a new array
// with the element at `from` moved to `to`; out-of-range or no-op moves return a
// copy unchanged. Kept pure so it can be unit-tested without the DOM.
export function moveItem<T>(arr: T[], from: number, to: number): T[] {
	const result = arr.slice()
	if (
		from < 0 || from >= result.length ||
		to < 0 || to >= result.length ||
		from === to
	) {
		return result
	}
	const [item] = result.splice(from, 1)
	result.splice(to, 0, item)
	return result
}
