import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/views/DashboardView.vue'),
    meta: { title: '首页' }
  },
  {
    path: '/monitor',
    component: () => import('@/views/MonitorView.vue'),
    meta: { title: '系统监控' }
  },
  {
    path: '/apps',
    component: () => import('@/views/AppsView.vue'),
    meta: { title: '微应用' }
  },
  {
    path: '/about',
    component: () => import('@/views/AboutView.vue'),
    meta: { title: '系统信息' }
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
