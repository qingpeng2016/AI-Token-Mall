import type { CatalogProduct } from '../types/product'
import { formatTokenCount } from './format'

/** card_features 内占位符，与后端 expandCardFeatures 一致 */
export function expandCardFeaturePlaceholders(
  features: string[],
  ctx: Pick<CatalogProduct, 'limit_tokens' | 'rpm_limit'>,
): string[] {
  return features.map((line) => {
    if (!line.includes('{')) return line
    return line
      .replaceAll('{limit_tokens}', formatTokenCount(ctx.limit_tokens))
      .replaceAll('{rpm_limit}', String(ctx.rpm_limit))
  })
}

export function subKeyShareFeature(seats: number): string {
  return `支持${seats}人共用`
}

export function appendShareFeature(features: string[], seats: number): string[] {
  const line = subKeyShareFeature(seats)
  if (features.some((f) => f.includes('人共用'))) return features
  return [...features, line]
}

export function normalizeCatalogProduct(p: CatalogProduct): CatalogProduct {
  const card_features = expandCardFeaturePlaceholders(p.card_features, p)
  return {
    ...p,
    card_features: appendShareFeature(card_features, p.share_seats),
  }
}

export function heroSubKeyShareLine(products: { share_seats: number }[]): string {
  if (products.length === 0) return ''
  const seats = products.map((p) => p.share_seats)
  const min = Math.min(...seats)
  const max = Math.max(...seats)
  if (min === max) return `支持开通子 Key，${subKeyShareFeature(min)}`
  return `支持开通子 Key，${min}–${max} 人按档共用`
}
