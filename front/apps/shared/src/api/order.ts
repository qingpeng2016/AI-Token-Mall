import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type CreateOrderBody = {
  product_id: number
  quantity: number
  channel: 'alipay' | 'wechat' | 'paypal'
  enterprise_invoice?: boolean
}

export type CreateOrderResult = {
  order_no: string
  out_trade_no: string
  order_id: number
  channel: string
  status: string
  total_amount_cents: number
  currency: string
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createOrderApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async create(body: CreateOrderBody): Promise<CreateOrderResult> {
      const res = await http.post<ApiEnvelope<CreateOrderResult>>('/api/v1/mock/orders', body)
      return unwrap(res)
    },
    async mockNotify(channel: string, outTradeNo: string): Promise<void> {
      await http.post<ApiEnvelope<unknown>>(`/api/v1/mock/payments/notify/${channel}`, {
        out_trade_no: outTradeNo,
        trade_status: 'TRADE_SUCCESS',
      })
    },
  }
}
