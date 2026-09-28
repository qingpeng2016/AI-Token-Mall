import { ref, type Ref } from 'vue'
import type { NavBrandMenu, NavMegaColumn, NavMenuResponse } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'
import type { ReloadOptions } from '@/composables/reloadOptions'

const megaMenu = ref<NavMegaColumn[]>([])
const brandMenus = ref<NavBrandMenu[]>([])
const loading = ref(false)

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

function mapNavApiResponse(data: NavMenuResponse | null | undefined): NavMenuResponse {
  if (!data) {
    return { mega_menu: [], brand_menus: [] }
  }
  return {
    mega_menu: data.mega_menu ?? [],
    brand_menus: (data.brand_menus ?? []).map((b) => ({
      id: b.id,
      label: b.label,
      hot_tag_name: b.hot_tag_name,
      items: b.items ?? [],
    })),
  }
}

function applyNavData(data: NavMenuResponse) {
  if (data.brand_menus.length === 0 && data.mega_menu.length === 0) {
    applyMockNav()
    return
  }
  megaMenu.value = data.mega_menu
  brandMenus.value = data.brand_menus
}

export async function reloadNavMenu(options?: ReloadOptions): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  const hasNav = brandMenus.value.length > 0 || megaMenu.value.length > 0
  loading.value = !soft || !hasNav
  if (!soft) {
    megaMenu.value = []
    brandMenus.value = []
    loading.value = true
  }
  try {
    const raw = await productApi.navMenu()
    if (seq !== loadSeq) return
    applyNavData(mapNavApiResponse(raw))
  } catch {
    if (seq !== loadSeq) return
    applyMockNav()
  } finally {
    if (seq === loadSeq) {
      loading.value = false
    }
  }
}

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  return { megaMenu, brandMenus, loading, reload: reloadNavMenu }
}
