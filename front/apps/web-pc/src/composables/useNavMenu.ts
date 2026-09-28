import { ref, type Ref } from 'vue'
import type { NavBrandMenu, NavMegaColumn } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'

let cachedMega: NavMegaColumn[] | null = null
let cachedBrands: NavBrandMenu[] | null = null
let loadingPromise: Promise<void> | null = null

function mockBrandMenus(): NavBrandMenu[] {
  return navBrandDropdowns
    .filter((e): e is Extract<typeof e, { items: unknown }> => 'items' in e)
    .map((e) => ({
      id: e.id,
      label: e.label,
      hot_sale: e.hotSale,
      items: e.items.map((i) => ({
        label: i.label,
        price: i.price ?? '',
        href: i.href,
        featured: i.featured,
      })),
    }))
}

function applyFallback() {
  cachedMega = navMegaMenu
  cachedBrands = mockBrandMenus()
}

async function loadNavMenu(): Promise<void> {
  if (cachedMega && cachedBrands) return
  if (!loadingPromise) {
    loadingPromise = (async () => {
      try {
        const data = await productApi.navMenu()
        if (data.mega_menu?.length) {
          cachedMega = data.mega_menu
          cachedBrands = (data.brand_menus ?? []).map((b) => ({
            id: b.id,
            label: b.label,
            hot_sale: b.hot_sale,
            items: b.items,
          }))
          return
        }
      } catch {
        /* fallback */
      }
      applyFallback()
    })()
  }
  await loadingPromise
}

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  loading: Ref<boolean>
} {
  const megaMenu = ref<NavMegaColumn[]>(cachedMega ?? navMegaMenu)
  const brandMenus = ref<NavBrandMenu[]>(cachedBrands ?? mockBrandMenus())
  const loading = ref(!(cachedMega && cachedBrands))

  if (!(cachedMega && cachedBrands)) {
    void loadNavMenu().then(() => {
      if (cachedMega) megaMenu.value = cachedMega
      if (cachedBrands) brandMenus.value = cachedBrands
      loading.value = false
    })
  }

  return { megaMenu, brandMenus, loading }
}
