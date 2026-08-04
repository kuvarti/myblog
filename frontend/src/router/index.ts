import { createRouter, createWebHistory } from 'vue-router'
import * as routes from '@/router/routes'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/panel',
			name: 'panel',
			component: () => Promise.resolve(routes.panel)
		},
		{
			path: '/lists',
			name: 'lists',
			component: () => Promise.resolve(routes.lists)
		},
		// Catch-all: every other path (including "/") renders a content page,
		// resolved by route.path against each page's own Path (per-page routing).
		// Kept last so the reserved /panel and /lists routes win.
		{
			path: '/:pathMatch(.*)*',
			name: 'content',
			component: () => Promise.resolve(routes.contents)
		}
	]
})

export default router
