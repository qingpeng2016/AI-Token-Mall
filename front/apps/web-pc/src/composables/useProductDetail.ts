import { ref, type Ref } from 'vue'
import type { CatalogProduct, ProductDetailResponse } from '@ai-token-mall/shared'
import { normalizeCatalogProduct } from '@ai-token-mall/shared'
import { productApi } from '@/api'
import {
  getProductDetail,
  type ProductDetailContent as MockProductDetailContent,
} from '@/mocks/productDetails'
import { findProductBySlug } from '@/mocks/productRoutes'
import { mockProducts } from '@/mocks/home'

export type ProductDetailViewContent = MockProductDetailContent

const product = ref<CatalogProduct | null>(null)
const detail = ref<ProductDetailViewContent | null>(null)
const loading = ref(false)

let loadSeq = 0
let prefetchedSlug: string | null = null

function mapApiDetail(data: ProductDetailResponse): {
  product: CatalogProduct
  detail: ProductDetailViewContent
} {
  const p = normalizeCatalogProduct(data.product)
  const d = data.detail
  return {
    product: p,
    detail: {
      eyebrow: d.eyebrow,
      heroTitle: d.hero_title,
      heroLead: d.hero_lead,
      heroBullets: d.hero_bullets ?? [],
      heroTags: d.hero_tags ?? [],
      audiences: d.audiences ?? [],
      compareTitle: d.compare_title,
      compareBody: d.compare_body,
      steps: d.steps ?? [],
      ctaTitle: d.cta_title,
      ctaSubtitle: d.cta_subtitle,
      faqs: d.faqs ?? [],
      related: d.related ?? [],
    },
  }
}

function applyMock(slug: string) {
  const fromList = findProductBySlug(slug, mockProducts)
  if (!fromList) {
    product.value = null
    detail.value = null
    return
  }
  product.value = fromList
  detail.value = getProductDetail(slug, fromList)
}

export async function prefetchProductDetail(slug: string): Promise<void> {
  const s = slug.trim()
  if (!s) return
  const seq = ++loadSeq
  loading.value = true
  try {
    const data = await productApi.detailBySlug(s)
    if (seq !== loadSeq) return
    const mapped = mapApiDetail(data)
    product.value = mapped.product
    detail.value = mapped.detail
    prefetchedSlug = s
  } catch {
    if (seq !== loadSeq) return
    applyMock(s)
    prefetchedSlug = product.value ? s : null
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function takePrefetchedProductDetail(slug: string): boolean {
  if (prefetchedSlug === slug && product.value && detail.value) {
    prefetchedSlug = null
    return true
  }
  return false
}

export function useProductDetail(): {
  product: Ref<CatalogProduct | null>
  detail: Ref<ProductDetailViewContent | null>
  loading: Ref<boolean>
} {
  return { product, detail, loading }
}
