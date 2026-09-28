import type { RouteLocationNormalized, Router } from 'vue-router'
import { reloadCatalogProducts } from '@/composables/useCatalogProducts'
import { reloadNavMenu } from '@/composables/useNavMenu'
import {
  prefetchBlogArticle,
  reloadTutorialBlog,
} from '@/composables/useTutorialBlog'
import { setRouteNavigationLoading } from '@/composables/useRouteNavigationLoading'

async function prepareRouteData(to: RouteLocationNormalized): Promise<void> {
  const name = to.name

  if (name === 'login' || name === 'register') {
    return
  }

  await reloadNavMenu({ soft: true })

  if (name === 'home' || name === 'product') {
    await reloadCatalogProducts({ soft: true })
  }
  if (name === 'blog') {
    await reloadTutorialBlog({ soft: true })
  }
  if (name === 'blog-article') {
    const slug = String(to.params.slug ?? '')
    await reloadTutorialBlog({ soft: true })
    if (slug) {
      await prefetchBlogArticle(slug)
    }
  }
}

/** 接口返回后再 next()：停留上一页，顶部 loading，新页一次展示内容 */
export function installRouteDataGuards(router: Router): void {
  router.beforeEach(async (to, from) => {
    if (to.fullPath === from.fullPath) {
      return true
    }

    setRouteNavigationLoading(true)
    try {
      await prepareRouteData(to)
      return true
    } finally {
      setRouteNavigationLoading(false)
    }
  })
}
