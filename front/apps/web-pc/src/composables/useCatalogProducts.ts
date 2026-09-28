import { computed, ref, type ComputedRef, type Ref } from 'vue'
import {
  catalogFilterPillsFromCategories,
  flattenCatalogProducts,
  normalizeCatalogProduct,
  readSessionCache,
  writeSessionCache,
  type CatalogFilterPill,
  type CatalogProduct,
  type ProductCatalogCategory,
  type ProductCatalogResponse,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { catalogFilterPills, mockCategorySlugById, mockProducts } from '@/mocks/home'

const SESSION_KEY = 'atm:product-catalog:v3'

type CatalogSession = ProductCatalogResponse & { fromApi?: boolean }

const categories = ref<ProductCatalogCategory[]>([])
const revalidating = ref(false)

let inflight: Promise<void> | null = null

function mockCatalogResponse(): ProductCatalogResponse {
  const pills = catalogFilterPills.filter((p) => p.value !== 'all')
  const cats: ProductCatalogCategory[] = pills.map((pill) => {
    const id = Number(pill.value)
    const slug = mockCategorySlugById[id]
    const products = mockProducts
      .filter((p) => p.products_category_name === slug)
      .map((p) =>
        normalizeCatalogProduct({ ...p, products_category_id: id }),
      )
    return {
      id,
      name: pill.label,
      dot_color: pill.dot,
      active_bg: pill.activeBg,
      sort: id * 10,
      products,
    }
  })
  return { categories: cats }
}

function seedDisplayCategories(): ProductCatalogCategory[] {
  const fromSession = readSessionCache<CatalogSession>(SESSION_KEY)
  if (fromSession?.fromApi && fromSession.categories?.length) {
    return fromSession.categories
  }
  return mockCatalogResponse().categories
}

function setCategories(cats: ProductCatalogCategory[]) {
  categories.value = cats
}

function persistApiCatalog(data: ProductCatalogResponse) {
  const payload: CatalogSession = { ...data, fromApi: true }
  setCategories(data.categories ?? [])
  writeSessionCache(SESSION_KEY, payload)
}

async function revalidateCatalog(): Promise<void> {
  try {
    const data = await productApi.list()
    persistApiCatalog({
      categories: (data.categories ?? []).map((c) => ({
        ...c,
        products: (c.products ?? []).map((p) => normalizeCatalogProduct(p)),
      })),
    })
  } catch {
    /* 保留当前 categories（mock / session / 上次 API） */
  }
}

function revalidateInBackground(): Promise<void> {
  if (inflight) return inflight
  revalidating.value = true
  inflight = revalidateCatalog().finally(() => {
    revalidating.value = false
    inflight = null
  })
  return inflight
}

export function useCatalogProducts(): {
  categories: Ref<ProductCatalogCategory[]>
  catalogFilterPills: ComputedRef<CatalogFilterPill[]>
  products: ComputedRef<CatalogProduct[]>
  initialLoading: Ref<boolean>
  revalidating: Ref<boolean>
} {
  if (!categories.value.length) {
    setCategories(seedDisplayCategories())
  }
  void revalidateInBackground()

  const initialLoading = computed(
    () => revalidating.value && categories.value.length === 0,
  )

  const catalogFilterPillsComputed = computed(() => {
    if (categories.value.length) {
      return catalogFilterPillsFromCategories(categories.value)
    }
    return catalogFilterPills
  })

  const products = computed(() =>
    flattenCatalogProducts({ categories: categories.value }),
  )

  return {
    categories,
    catalogFilterPills: catalogFilterPillsComputed,
    products,
    initialLoading,
    revalidating,
  }
}
