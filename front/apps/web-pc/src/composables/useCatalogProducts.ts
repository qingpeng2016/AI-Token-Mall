import { computed, ref, type ComputedRef, type Ref } from 'vue'
import {
  catalogFilterPillsFromCategories,
  flattenCatalogProducts,
  normalizeCatalogProduct,
  type CatalogFilterPill,
  type CatalogProduct,
  type ProductCatalogCategory,
  type ProductCatalogResponse,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { catalogFilterPills, mockCategorySlugById, mockProducts } from '@/mocks/home'

let cachedCatalog: ProductCatalogResponse | null = null
let loadingPromise: Promise<ProductCatalogResponse> | null = null

function mockCatalogResponse(): ProductCatalogResponse {
  const pills = catalogFilterPills.filter((p) => p.value !== 'all')
  const categories: ProductCatalogCategory[] = pills.map((pill) => {
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
  return { categories }
}

export async function loadProductCatalog(): Promise<ProductCatalogResponse> {
  if (cachedCatalog) return cachedCatalog
  if (!loadingPromise) {
    loadingPromise = (async () => {
      try {
        const data = await productApi.list()
        if (data.categories?.length) {
          cachedCatalog = {
            categories: data.categories.map((c) => ({
              ...c,
              products: c.products.map((p) => normalizeCatalogProduct(p)),
            })),
          }
          return cachedCatalog
        }
      } catch {
        /* 无后端或迁移未执行时回退 mock */
      }
      cachedCatalog = mockCatalogResponse()
      return cachedCatalog
    })()
  }
  return loadingPromise
}

export function useCatalogProducts(): {
  categories: Ref<ProductCatalogCategory[]>
  catalogFilterPills: ComputedRef<CatalogFilterPill[]>
  products: ComputedRef<CatalogProduct[]>
  loading: Ref<boolean>
} {
  const categories = ref<ProductCatalogCategory[]>(cachedCatalog?.categories ?? [])
  const loading = ref(!cachedCatalog)

  const catalogFilterPillsComputed = computed(() =>
    categories.value.length
      ? catalogFilterPillsFromCategories(categories.value)
      : catalogFilterPills,
  )

  const products = computed(() => flattenCatalogProducts({ categories: categories.value }))

  if (!cachedCatalog) {
    void loadProductCatalog().then((data) => {
      categories.value = data.categories
      loading.value = false
    })
  }

  return {
    categories,
    catalogFilterPills: catalogFilterPillsComputed,
    products,
    loading,
  }
}
