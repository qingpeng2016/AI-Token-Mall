export function parseSubscriptionDate(s: string): Date | null {
  const raw = s?.trim()
  if (!raw) return null
  const d = new Date(`${raw}T12:00:00`)
  return Number.isNaN(d.getTime()) ? null : d
}

export function addBillingPeriod(from: Date, billingPeriod: string): Date {
  const d = new Date(from.getTime())
  if (billingPeriod === 'year') {
    d.setFullYear(d.getFullYear() + 1)
  } else if (billingPeriod === 'once') {
    d.setDate(d.getDate() + 30)
  } else {
    d.setMonth(d.getMonth() + 1)
  }
  return d
}

/** 升档计价周期：当前周期 1；expires_at 晚于 period_end 时自 period_end 按周期累加。 */
export function countUpgradeBillingCycles(
  periodEndStr: string,
  expiresAtStr: string,
  billingPeriod: string,
): number {
  const maxQty = 99
  let qty = 1
  const periodEnd = parseSubscriptionDate(periodEndStr)
  const expiresAt = parseSubscriptionDate(expiresAtStr)
  if (!periodEnd || !expiresAt || expiresAt.getTime() <= periodEnd.getTime()) {
    return qty
  }
  let cursor = periodEnd
  while (cursor.getTime() < expiresAt.getTime() && qty < maxQty) {
    cursor = addBillingPeriod(cursor, billingPeriod)
    qty++
  }
  return qty
}

export function formatSubscriptionDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}
