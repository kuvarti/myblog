import { createRouter, createWebHistory } from 'vue-router'
import * as routes from '@/router/routes'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'content',
			component: () => Promise.resolve(routes.contents)
		},
		{
			path:'/lists',
			name: 'lists',
			component: () => Promise.resolve(routes.lists)
		},
		{
			path: '/panel',
			name: 'panel',
			component: () => Promise.resolve(routes.panel)
		}
	]
})

export default router
