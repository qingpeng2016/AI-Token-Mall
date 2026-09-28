import { ref, type Ref } from 'vue'
import type { NavBrandMenu, NavMegaColumn } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'

function mockBrandMenus(): NavBrandMenu[] {
  return navBrandDropdowns
    .filter((e): e is Extract<typeof e, { items: unknown }> => 'items' in e)
    .slice(0, 3)
    .map((e) => ({
      id: e.id,
      label: e.label,
      hot_tag_name: e.hotTagName,
      items: e.items.map((i) => ({
        label: i.label,
        price: i.price ?? '',
        href: i.href,
        featured: i.featured,
      })),
    }))
}

function applyFallback(): { mega: NavMegaColumn[]; brands: NavBrandMenu[] } {
  return { mega: navMegaMenu, brands: mockBrandMenus() }
}

async function fetchNavMenu(): Promise<{ mega: NavMegaColumn[]; brands: NavBrandMenu[] }> {
  try {
    const data = await productApi.navMenu()
    const mega = data.mega_menu ?? []
    const brands = (data.brand_menus ?? []).slice(0, 3).map((b) => ({
      id: b.id,
      label: b.label,
      hot_tag_name: b.hot_tag_name,
      items: b.items,
    }))
    if (mega.length || brands.length) {
      return { mega, brands }
    }
  } catch {
    /* fallback */
  }
  return applyFallback()
}

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  loading: Ref<boolean>
} {
  const megaMenu = ref<NavMegaColumn[]>([])
  const brandMenus = ref<NavBrandMenu[]>([])
  const loading = ref(true)

  void fetchNavMenu().then(({ mega, brands }) => {
    megaMenu.value = mega
    brandMenus.value = brands
    loading.value = false
  })

  return { megaMenu, brandMenus, loading }
}
