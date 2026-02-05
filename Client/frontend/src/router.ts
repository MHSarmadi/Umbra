import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
	{
		path: '/hello-umbra',
		name: 'HelloUmbra',
		component: () => import('./pages/HelloUmbra.vue'),
	}
]

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes,
})

export default router