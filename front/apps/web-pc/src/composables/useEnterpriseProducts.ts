import { computed, ref, type ComputedRef, type Ref } from 'vue'
import type { EnterpriseProduct } from '@ai-token-mall/shared'
import { enterpriseApi } from '@/api'
import { enterprisePlans as mockPlans } from '@/mocks/enterprise'
import type { ReloadOptions } from '@/composables/reloadOptions'

export type EnterprisePlanView = {
  id: string
  name: string
  badge?: string
  priceHint: string
  seats: string
  features: string[]
  cta: string
  buttonLabel: string
  featured?: boolean
}

const products = ref<EnterpriseProduct[]>([])
const loading = ref(false)

let loadSeq = 0

function fromMock(): EnterpriseProduct[] {
  return mockPlans.map((p) => ({
    code: p.id,
    name: p.name,
    badge: p.badge,
    price_hint: p.priceHint,
    seats: p.seats,
    features: p.features,
    tagline: p.cta,
    button_label: '获取报价',
    is_featured: p.featured === true,
    sort: 0,
  }))
}

function mapToView(p: EnterpriseProduct): EnterprisePlanView {
  return {
    id: p.code,
    name: p.name,
    badge: p.badge,
    priceHint: p.price_hint,
    seats: p.seats,
    features: p.features ?? [],
    cta: p.tagline,
    buttonLabel: p.button_label || '获取报价',
    featured: p.is_featured,
  }
}

export async function reloadEnterpriseProducts(
  options?: ReloadOptions,
): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  loading.value = !soft || products.value.length === 0
  try {
    const list = await enterpriseApi.listProducts()
    if (seq !== loadSeq) return
    products.value = list.length ? list : fromMock()
  } catch {
    if (seq !== loadSeq) return
    if (!soft || products.value.length === 0) {
      products.value = fromMock()
    }
  } finally {
    if (seq === loadSeq) {
      loading.value = false
    }
  }
}

export function useEnterpriseProducts(): {
  plans: ComputedRef<EnterprisePlanView[]>
  loading: Ref<boolean>
} {
  const plans = computed(() => products.value.map(mapToView))
  return { plans, loading }
}
