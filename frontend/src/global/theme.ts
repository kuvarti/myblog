import { LocalStorageService } from '@/service/LocalStorage.service'

export type Theme = 'light' | 'dark'
const KEY = 'theme'
const ls = new LocalStorageService()

export function resolveInitialTheme(): Theme {
	const stored = ls.GetData(KEY)
	if (stored === 'light' || stored === 'dark') return stored
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(t: Theme): void {
	document.documentElement.setAttribute('data-theme', t)
}

export function getTheme(): Theme {
	return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

export function toggleTheme(): Theme {
	const next: Theme = getTheme() === 'dark' ? 'light' : 'dark'
	applyTheme(next)
	ls.SaveData(KEY, next)
	return next
}
