/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}",],
	theme: {
		extend: {
			colors: {
				'midnightPurple': '#382D40',
				'mainComponentBackground': '#382D40',
				'activePageColor': '#eef2ff',
				'deActivePageColor': '#a5b4fc',
			},
			fontFamily: {
				FiraCode: ["Fira Code", "monospace"],
			},
		},
	},
	plugins: [],
}

