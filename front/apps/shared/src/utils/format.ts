export function formatCnyFromCents(cents: number): string {
  return `¥${(cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)}`
}

export function formatTokenCount(n: number): string {
  if (n >= 1_000_000) {
    return `${(n / 1_000_000).toFixed(n % 1_000_000 === 0 ? 0 : 1)}M`
  }
  if (n >= 1_000) {
    return `${(n / 1_000).toFixed(n % 1_000 === 0 ? 0 : 1)}K`
  }
  return String(n)
}

export function billingPeriodLabel(period: 'month' | 'year' | 'once'): string {
  switch (period) {
    case 'month':
      return '/ 月'
    case 'year':
      return '/ 年'
    case 'once':
      return '一次性'
    default:
      return ''
  }
}
