package betrustsign

import (
	"testing"
	"time"
)

func TestComputeSign_deterministic(t *testing.T) {
	params := map[string]string{
		"timestamp":      "1757923200123",
		"liquidation_id": "LQ-20260915-000001",
		"loan_order_id":  "LO-987654",
	}
	secret := "test_app_secret"
	base := BuildSignBase(params)
	s1 := ComputeSign(secret, base)
	s2 := ComputeSign(secret, base)
	if s1 != s2 || len(s1) != 64 {
		t.Fatalf("unexpected sign: %s", s1)
	}
	if !VerifySign(secret, params, s1) {
		t.Fatal("verify failed")
	}
}

func TestMarshalAndAttachSign_roundTrip(t *testing.T) {
	secret := "test_app_secret"
	now := time.UnixMilli(1_750_000_000_123)
	type payload struct {
		Timestamp     int64  `json:"timestamp"`
		LiquidationID string `json:"liquidation_id"`
	}
	body, err := MarshalAndAttachSign(secret, payload{Timestamp: now.UnixMilli(), LiquidationID: "LQ-1"})
	if err != nil {
		t.Fatal(err)
	}
	params, err := FlattenJSONObject(body)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := VerifyIncomingParams(secret, params, now); !ok {
		t.Fatal("signed outbound body should verify")
	}
}

func TestVerifyIncomingParams_queryStyle(t *testing.T) {
	secret := "test_app_secret"
	now := time.UnixMilli(1_750_000_000_123)
	params := map[string]string{
		"liquidation_id": "LQ-1",
		"timestamp":      "1750000000123",
	}
	params["sign"] = ComputeSign(secret, BuildSignBase(params))
	if _, ok := VerifyIncomingParams(secret, params, now); !ok {
		t.Fatal("GET-style params should verify")
	}
}

func TestVerifyTimestamp(t *testing.T) {
	now := time.UnixMilli(1_000_000)
	if !VerifyTimestamp(1_000_000, now) {
		t.Fatal("same ms should pass")
	}
	if VerifyTimestamp(1_000_000-20_000, now) {
		t.Fatal("20s skew should fail")
	}
}
