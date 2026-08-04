// Convert between the panel's comma-separated tag input and the string[] the API uses.
export function parseTags(input: string): string[] {
	return input.split(',').map((t) => t.trim()).filter(Boolean)
}

export function formatTags(tags: string[]): string {
	return (tags ?? []).join(', ')
}
