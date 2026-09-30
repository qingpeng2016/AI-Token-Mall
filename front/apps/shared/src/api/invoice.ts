import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type UserInvoiceItem = {
  id: number
  order_no: string
  title: string
  amount: number | string
  status: 'pending' | 'issued' | 'failed' | string
  created_at: string
}

export type UserInvoiceListPage = {
  items: UserInvoiceItem[]
  total: number
  page: number
  page_size: number
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createInvoiceApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(params?: { page?: number; page_size?: number }): Promise<UserInvoiceListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs ? `/api/v1/users/invoices?${qs}` : '/api/v1/users/invoices'
      const res = await http.get<ApiEnvelope<UserInvoiceListPage>>(path)
      return unwrap(res)
    },
  }
}
