import { createRouter, createWebHistory } from 'vue-router'
import * as routes from '@/router/routes'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'content',
			component: () => Promise.resolve(routes.viewProfile)
		}
	]
})

export default router
