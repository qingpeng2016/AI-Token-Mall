package billing

import (
	"testing"
	"time"
)

func TestCountUpgradeBillingCycles(t *testing.T) {
	loc := time.UTC
	d := func(y int, m time.Month, day int) time.Time {
		return time.Date(y, m, day, 0, 0, 0, 0, loc)
	}

	tests := []struct {
		name     string
		periodEnd, expiresAt time.Time
		period   string
		want     int
	}{
		{"current only", d(2026, 4, 26), d(2026, 4, 26), "month", 1},
		{"before period end", d(2026, 4, 26), d(2026, 4, 10), "month", 1},
		{"three months ahead", d(2026, 4, 26), d(2026, 7, 26), "month", 4},
		{"one month ahead", d(2026, 4, 26), d(2026, 5, 26), "month", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountUpgradeBillingCycles(tt.periodEnd, tt.expiresAt, tt.period)
			if got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}
