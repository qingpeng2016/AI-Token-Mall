import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    // 会员中心等同一路径只改 ?tab= 时不滚回顶部，避免侧栏/整页上下跳
    if (from.path && to.path === from.path) return false
    if (to.hash) {
      return { el: to.hash, top: 80, behavior: 'smooth' }
    }
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/views/HomeView.vue') },
        {
          path: 'p/:slug',
          name: 'product',
          component: () => import('@/views/ProductDetailView.vue'),
        },
        {
          path: 'member',
          name: 'member',
          component: () => import('@/views/MemberCenterView.vue'),
        },
        {
          path: 'enterprise',
          name: 'enterprise',
          component: () => import('@/views/EnterpriseView.vue'),
        },
        {
          path: 'blog',
          name: 'blog',
          component: () => import('@/views/BlogView.vue'),
        },
        {
          path: 'blog/:slug',
          name: 'blog-article',
          component: () => import('@/views/BlogArticleView.vue'),
        },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/AuthLayout.vue'),
      children: [
        { path: 'login', name: 'login', component: () => import('@/views/LoginView.vue') },
        { path: 'register', name: 'register', component: () => import('@/views/RegisterView.vue') },
      ],
    },
  ],
})

export default router
