import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'home',
			component: () => import("@/views/HomeView.vue")
		},
		{
			path: '/deneme',
			name: 'deneme',
			component: () => import("@/views/deneme.vue")
		}
	]
})

export default router
