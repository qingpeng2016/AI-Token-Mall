import { ref, type Ref } from 'vue'
import {
  normalizeCatalogProduct,
  type CatalogProduct,
} from '@ai-token-mall/shared'
import { productApi } from '@/api'
import { mockProducts } from '@/mocks/home'

let cached: CatalogProduct[] | null = null
let loadingPromise: Promise<CatalogProduct[]> | null = null

export async function loadCatalogProducts(): Promise<CatalogProduct[]> {
  if (cached) return cached
  if (!loadingPromise) {
    loadingPromise = (async () => {
      try {
        const rows = await productApi.list()
        if (rows.length > 0) {
          cached = rows.map((p) => normalizeCatalogProduct(p))
          return cached
        }
      } catch {
        /* 无后端或迁移未执行时回退 mock */
      }
      cached = mockProducts
      return cached
    })()
  }
  return loadingPromise
}

export function useCatalogProducts(): {
  products: Ref<CatalogProduct[]>
  loading: Ref<boolean>
} {
  const products = ref<CatalogProduct[]>(cached ?? [])
  const loading = ref(!cached)

  if (!cached) {
    void loadCatalogProducts().then((list) => {
      products.value = list
      loading.value = false
    })
  }

  return { products, loading }
}
