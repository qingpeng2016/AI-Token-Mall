import type { ApiEnvelope } from '../types/user'
import type { CatalogProduct } from '../types/product'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createProductApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(skuUpstreamName?: string): Promise<CatalogProduct[]> {
      const qs =
        skuUpstreamName && skuUpstreamName.trim()
          ? `?sku_upstream_name=${encodeURIComponent(skuUpstreamName.trim())}`
          : ''
      const res = await http.get<ApiEnvelope<{ products: CatalogProduct[] }>>(
        `/api/v1/products${qs}`,
      )
      return unwrap(res).products ?? []
    },
  }
}
