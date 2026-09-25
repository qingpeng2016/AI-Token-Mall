package taskexec

import (
	"encoding/json"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
)

type execSnapshotQuote struct {
	Mid                    string `json:"mid"`
	FeeRate                string `json:"fee_rate"`
	RemainingDebtAmount    string `json:"remaining_debt_amount"`
	RemainingCollateralQty string `json:"remaining_collateral_qty"`
	MinQty                 string `json:"min_qty"`
	Step                   string `json:"step"`
	Tick                   string `json:"tick"`
}

type execSnapshotPlan struct {
	Floor      string `json:"floor"`
	NeedSell   string `json:"need_sell"`
	LimitPrice string `json:"limit_price"`
}

type execSnapshotPayload struct {
	CapturedAt                 string             `json:"captured_at"`
	LiquidationID              string             `json:"liquidation_id"`
	BatchNo                    string             `json:"batch_no"`
	BatchSeq                   uint               `json:"batch_seq"`
	Tier                       string             `json:"tier"`
	Symbol                     string             `json:"symbol"`
	TriggerCollateralAsset     string             `json:"trigger_collateral_asset"`
	TriggerCollateralQty       string             `json:"trigger_collateral_qty"`
	TriggerCollateralDueAmount string             `json:"trigger_collateral_due_amount"`
	TriggerMarkPrice           string             `json:"trigger_mark_price"`
	CurrentTier                string             `json:"current_tier"`
	LastMidPrice               string             `json:"last_mid_price,omitempty"`
	Quote                      execSnapshotQuote  `json:"quote"`
	Plan                       execSnapshotPlan   `json:"plan"`
}

func buildExecSnapshotJSON(
	task entity.LiquidationTask,
	tier, symbol, batchNo string,
	batchSeq uint,
	quote *execQuoteState,
	plan *execOrderPlan,
) (string, error) {
	currentTier := common.CurrentTier(task)
	lastMid := ""
	if task.LastMidPrice != nil {
		lastMid = common.FormatDecimal(*task.LastMidPrice)
	}
	payload := execSnapshotPayload{
		CapturedAt:                 time.Now().UTC().Format(time.RFC3339Nano),
		LiquidationID:              task.LiquidationID,
		BatchNo:                    batchNo,
		BatchSeq:                   batchSeq,
		Tier:                       tier,
		Symbol:                     symbol,
		TriggerCollateralAsset:     task.TriggerCollateralAsset,
		TriggerCollateralQty:       common.FormatDecimal(task.TriggerCollateralQty),
		TriggerCollateralDueAmount: common.FormatDecimal(task.TriggerCollateralDueAmount),
		TriggerMarkPrice:           common.FormatDecimal(task.TriggerMarkPrice),
		CurrentTier:                currentTier,
		LastMidPrice:               lastMid,
		Quote: execSnapshotQuote{
			Mid:                    common.FormatDecimal(quote.Mid),
			FeeRate:                quote.FeeRate.String(),
			RemainingDebtAmount:    common.FormatDecimal(quote.RemainingDebtAmount),
			RemainingCollateralQty: common.FormatDecimal(quote.RemainingCollateralQty),
			MinQty:                 common.FormatDecimal(quote.MinQty),
			Step:                   common.FormatDecimal(quote.Step),
			Tick:                   common.FormatDecimal(quote.Tick),
		},
		Plan: execSnapshotPlan{
			Floor:      common.FormatDecimal(plan.Floor),
			NeedSell:   common.FormatDecimal(plan.NeedSell),
			LimitPrice: common.FormatDecimal(plan.LimitPrice),
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
