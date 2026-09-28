import { ref, type Ref } from 'vue'
import type { NavBrandMenu, NavMegaColumn, NavMenuResponse } from '@ai-token-mall/shared'
import { readSessionCache, writeSessionCache } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { navBrandDropdowns, navMegaMenu } from '@/mocks/nav'

const SESSION_KEY = 'atm:nav-menu:v1'

type NavSessionCache = NavMenuResponse & { fromApi?: boolean }

const megaMenu = ref<NavMegaColumn[]>([])
const brandMenus = ref<NavBrandMenu[]>([])
const revalidating = ref(false)

let inflight: Promise<void> | null = null

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

function applyNavData(data: NavMenuResponse) {
  if (data.mega_menu?.length) {
    megaMenu.value = data.mega_menu
  }
  if (data.brand_menus?.length) {
    brandMenus.value = data.brand_menus
  }
}

function persistNavFromApi(data: NavMenuResponse) {
  const payload: NavSessionCache = {
    mega_menu: data.mega_menu ?? [],
    brand_menus: data.brand_menus ?? [],
    fromApi: true,
  }
  applyNavData(payload)
  writeSessionCache(SESSION_KEY, payload)
}

function hydrateNavFromSession(): boolean {
  const cached = readSessionCache<NavSessionCache>(SESSION_KEY)
  if (!cached?.fromApi) return false
  applyNavData(cached)
  return megaMenu.value.length > 0 || brandMenus.value.length > 0
}

function seedNavDisplay() {
  if (hydrateNavFromSession()) return
  if (!megaMenu.value.length) {
    megaMenu.value = navMegaMenu
  }
  if (!brandMenus.value.length) {
    brandMenus.value = mockBrandMenus()
  }
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

async function revalidateNav(): Promise<void> {
  try {
    const data = mapNavApiResponse(await productApi.navMenu())
    persistNavFromApi(data)
  } catch {
    /* 保留 session / mock */
  }
}

function revalidateInBackground(): Promise<void> {
  if (inflight) return inflight
  revalidating.value = true
  inflight = revalidateNav().finally(() => {
    revalidating.value = false
    inflight = null
  })
  return inflight
}

/** 模块加载时同步读 session，避免刷新首帧先闪 mock 再被接口替换 */
seedNavDisplay()

export function useNavMenu(): {
  megaMenu: Ref<NavMegaColumn[]>
  brandMenus: Ref<NavBrandMenu[]>
  revalidating: Ref<boolean>
} {
  if (!megaMenu.value.length && !brandMenus.value.length) {
    seedNavDisplay()
  }
  void revalidateInBackground()
  return { megaMenu, brandMenus, revalidating }
}
