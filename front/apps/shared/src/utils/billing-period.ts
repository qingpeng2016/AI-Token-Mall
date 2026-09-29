export function parseSubscriptionDate(s: string): Date | null {
  const raw = s?.trim()
  if (!raw) return null
  const d = new Date(`${raw}T12:00:00`)
  return Number.isNaN(d.getTime()) ? null : d
}

export function resolvePeriodDays(periodDays: number | undefined, billingPeriod: string): number {
  if (periodDays != null && periodDays > 0) return periodDays
  if (billingPeriod === 'year') return 365
  if (billingPeriod === 'once') return 30
  return 30
}

export function addPeriodDays(from: Date, periodDays: number): Date {
  const d = new Date(from.getTime())
  d.setDate(d.getDate() + periodDays)
  return d
}

/** 升档计价周期：当前周期 1；expires_at 晚于 period_end 时自 period_end 按 period_days 累加。 */
export function countUpgradeBillingCycles(
  periodEndStr: string,
  expiresAtStr: string,
  periodDays: number,
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
    cursor = addPeriodDays(cursor, periodDays)
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

/** @deprecated 使用 addPeriodDays / resolvePeriodDays */
export function addBillingPeriod(from: Date, billingPeriod: string): Date {
  return addPeriodDays(from, resolvePeriodDays(undefined, billingPeriod))
}
