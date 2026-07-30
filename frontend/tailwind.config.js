/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}",],
	theme: {
		extend: {
			colors: {
				// semantic tokens (theme-aware via CSS variables)
				'bg': 'var(--bg)',
				'surface': 'var(--surface)',
				'surface-2': 'var(--surface-2)',
				'fg': 'var(--fg)',
				'muted': 'var(--muted)',
				'accent': 'var(--accent)',
				'border': 'var(--border)',
				// legacy names remapped to tokens so existing classes become theme-aware
				'midnightPurple': 'var(--surface)',
				'mainComponentBackground': 'var(--surface)',
				'activePageColor': 'var(--accent)',
				'deActivePageColor': 'var(--muted)',
			},
			fontFamily: {
				serif: ["Fraunces", "Georgia", "serif"],
				sans: ["Inter", "system-ui", "sans-serif"],
				mono: ["Fira Code", "monospace"],
				FiraCode: ["Fira Code", "monospace"],
			},
		},
	},
	plugins: [],
}

