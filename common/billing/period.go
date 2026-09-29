package billing

import "time"

// AddPeriod 按商品 billing_period 推进一个计费周期（与 infrastructure/mysql 履约一致）。
func AddPeriod(from time.Time, billingPeriod string) time.Time {
	switch billingPeriod {
	case "year":
		return from.AddDate(1, 0, 0)
	case "once":
		return from.AddDate(0, 0, 30)
	default:
		return from.AddDate(0, 1, 0)
	}
}

// CountUpgradeBillingCycles 升档计价周期数：当前周期计 1；若 expires_at 晚于 period_end，自 period_end 起按周期累加直至覆盖 expires_at。
func CountUpgradeBillingCycles(periodEnd, expiresAt time.Time, billingPeriod string) int {
	const maxQty = 99
	qty := 1
	if !expiresAt.After(periodEnd) {
		return qty
	}
	cursor := periodEnd
	for cursor.Before(expiresAt) && qty < maxQty {
		cursor = AddPeriod(cursor, billingPeriod)
		qty++
	}
	return qty
}
