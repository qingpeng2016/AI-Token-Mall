import { ref, type Ref } from 'vue'
import type { NavBrandMenu, NavMegaColumn, NavMenuResponse } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'

const megaMenu = ref<NavMegaColumn[]>([])
const brandMenus = ref<NavBrandMenu[]>([])
const loading = ref(true)

let loadSeq = 0

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

function applyMockNav() {
  megaMenu.value = navMegaMenu
  brandMenus.value = mockBrandMenus()
}

function mapNavApiResponse(data: NavMenuResponse): NavMenuResponse {
  return {
    mega_menu: data.mega_menu ?? [],
    brand_menus: (data.brand_menus ?? []).map((b) => ({
      id: b.id,
      label: b.label,
      hot_tag_name: b.hot_tag_name,
      items: b.items,
    })),
  }
}

export async function reloadNavMenu(): Promise<void> {
  const seq = ++loadSeq
  loading.value = true
  megaMenu.value = []
  brandMenus.value = []
  try {
    const data = mapNavApiResponse(await productApi.navMenu())
    if (seq !== loadSeq) return
    megaMenu.value = data.mega_menu
    brandMenus.value = data.brand_menus
  } catch {
    if (seq !== loadSeq) return
    applyMockNav()
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  void reloadNavMenu()
  return { megaMenu, brandMenus, loading, reload: reloadNavMenu }
}
