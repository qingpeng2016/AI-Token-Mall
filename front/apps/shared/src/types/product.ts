/** 与表 products 字段对齐（展示/下单用） */

export type UpstreamName = 'openai' | 'anthropic' | 'xai' | 'gemini' | 'perplexity'

export type ProductType = 'subscription' | 'token_topup'
export type BillingPeriod = 'month' | 'year' | 'once'
export type ProductStatus = 'on_sale' | 'off_sale'

export interface Product {
  id: number
  sku_code: string
  marketing_tier: string
  upstream_name: UpstreamName
  upstream_product: string
  limit_tokens: number
  rpm_limit: number
  tpm_limit?: number | null
  allowed_models: string[]
  product_type: ProductType
  billing_period: BillingPeriod
  price_cents: number
  currency: string
  compare_at_price_cents?: number | null
  highlights: string[]
  is_hot: boolean
  is_api_enabled: boolean
  topup_token_amount?: number | null
  sort_order: number
  status: ProductStatus
}

export interface HomeBanner {
  id: string
  title: string
  subtitle: string
  ctaText: string
  tone: 'primary' | 'success' | 'warning'
}
