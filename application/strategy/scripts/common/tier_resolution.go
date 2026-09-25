package common

import (
	"encoding/json"
	"fmt"
	"strings"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/shopspring/decimal"
)

// TierRuleHit 单条档位规则的判定结果（含当时实际值与阈值说明）。
type TierRuleHit struct {
	Tier      string `json:"tier"`
	Rule      string `json:"rule"`
	Matched   bool   `json:"matched"`
	PrdLine   string `json:"prd_line,omitempty"` // 与产品表一致的条目描述（仅 matched=true 时有值）
	Actual    string `json:"actual,omitempty"`
	Threshold string `json:"threshold,omitempty"`
}

// TierPRDEntry 对应产品表「档位 → 进入条件」一行，并标明本次命中了哪几条（或/且）。
type TierPRDEntry struct {
	EntryCondition string   `json:"entry_condition"`       // 产品表原文
	Matched        bool     `json:"matched"`               // 本档任一进入条件是否成立
	MatchedItems   []string `json:"matched_items,omitempty"` // 命中的子条件（中文）
}

// TierResolutionExplain 档位重算说明，写入 liquidation_task.current_tier_reason_json。
type TierResolutionExplain struct {
	ResolvedTier          string                  `json:"resolved_tier"`
	PreviousTier          string                  `json:"previous_tier"`
	ResolvedTierWhy       string                  `json:"resolved_tier_why"`        // 人话：为何是这个档位
	MatchedPRDConditions  []string                `json:"matched_prd_conditions"`   // 本次满足的产品条目（可多条、跨 P0/P1）
	TierEntryByPRD        map[string]TierPRDEntry `json:"tier_entry_by_prd"`        // P0/P1/P2 对照产品表
	Inputs                map[string]string       `json:"inputs"`
	RuleHits              []TierRuleHit           `json:"rule_hits"`
	MatchedTiers          []string                `json:"matched_tiers"`
	Note                  string                  `json:"note"`
}

var prdEntryConditionText = map[string]string{
	appcommon.LiquidationTierP0: "安全缓冲<8%，或执行总时间超30秒",
	appcommon.LiquidationTierP1: "安全缓冲8%-15%，或执行总时间超15秒，或连续3次IOC未成交，或近5秒价格跌幅>0.5%",
	appcommon.LiquidationTierP2: "安全缓冲>15%，且执行总时间<15秒",
}

// ResolveTierForTaskWithExplain 重算档位（只升不降）并返回判定明细。
func ResolveTierForTaskWithExplain(
	task entity.LiquidationTask,
	tiers []entity.LiquidationTierConfig,
	mid decimal.Decimal,
	consecutiveIOCZero int,
) (string, TierResolutionExplain) {
	previous := appcommon.LiquidationTierP2
	if task.CurrentTier != nil && strings.TrimSpace(*task.CurrentTier) != "" {
		previous = strings.TrimSpace(*task.CurrentTier)
	}
	tc := TierContext{
		SafetyBuffer:       ComputeSafetyBuffer(mid, task.TriggerCollateralDueAmount, task.TriggerCollateralQty),
		ExecSec:            ExecSeconds(task),
		ConsecutiveIOCZero: consecutiveIOCZero,
		PriceDropPct:       PriceDropPct(task, mid),
	}
	explain := buildTierExplain(previous, tiers, tc, mid, task)
	tier := ResolveTier(previous, tiers, tc)
	explain.ResolvedTier = tier
	finalizeTierExplain(&explain)
	return tier, explain
}

func buildTierExplain(
	previous string,
	tiers []entity.LiquidationTierConfig,
	c TierContext,
	mid decimal.Decimal,
	task entity.LiquidationTask,
) TierResolutionExplain {
	inputs := map[string]string{
		"mid":                           FormatDecimal(mid),
		"safety_buffer":                 c.SafetyBuffer.String(),
		"exec_sec":                      decimal.NewFromInt(int64(c.ExecSec)).String(),
		"consecutive_ioc_unfilled":      decimal.NewFromInt(int64(c.ConsecutiveIOCZero)).String(),
		"price_drop_pct":                c.PriceDropPct.String(),
		"trigger_collateral_due_amount": task.TriggerCollateralDueAmount.String(),
		"trigger_collateral_qty":        task.TriggerCollateralQty.String(),
		"trigger_mark_price":            task.TriggerMarkPrice.String(),
	}
	if task.LastMidPrice != nil {
		inputs["last_mid_price"] = task.LastMidPrice.String()
	} else {
		inputs["last_mid_price"] = ""
	}

	var hits []TierRuleHit
	var matchedTiers []string
	for _, cfg := range tiers {
		if cfg.IsEnabled != 1 {
			continue
		}
		tierHits := explainTierRules(cfg, c)
		hits = append(hits, tierHits...)
		if tierMatches(cfg, c) {
			matchedTiers = append(matchedTiers, cfg.Tier)
		}
	}

	return TierResolutionExplain{
		PreviousTier: previous,
		Inputs:       inputs,
		RuleHits:     hits,
		MatchedTiers: matchedTiers,
		Note:         "只升不降：在命中档位中取 tier_rank 最小；未命中任何进入条件时默认 P2",
	}
}

func finalizeTierExplain(ex *TierResolutionExplain) {
	ex.TierEntryByPRD = map[string]TierPRDEntry{
		appcommon.LiquidationTierP0: {EntryCondition: prdEntryConditionText[appcommon.LiquidationTierP0]},
		appcommon.LiquidationTierP1: {EntryCondition: prdEntryConditionText[appcommon.LiquidationTierP1]},
		appcommon.LiquidationTierP2: {EntryCondition: prdEntryConditionText[appcommon.LiquidationTierP2]},
	}

	var matchedPRD []string
	for i := range ex.RuleHits {
		h := &ex.RuleHits[i]
		if !h.Matched {
			continue
		}
		line := prdMatchedLine(*h)
		h.PrdLine = line
		matchedPRD = append(matchedPRD, line)
		entry := ex.TierEntryByPRD[h.Tier]
		entry.Matched = true
		entry.MatchedItems = append(entry.MatchedItems, prdItemShort(h.Tier, h.Rule))
		ex.TierEntryByPRD[h.Tier] = entry
	}
	ex.MatchedPRDConditions = matchedPRD
	ex.ResolvedTierWhy = buildResolvedTierWhy(ex.ResolvedTier, ex.PreviousTier, ex.MatchedTiers, matchedPRD)
}

func prdItemShort(tier, rule string) string {
	switch tier + "/" + rule {
	case appcommon.LiquidationTierP0 + "/safety_buffer_lt":
		return "安全缓冲<8%"
	case appcommon.LiquidationTierP0 + "/exec_time_exceed":
		return "执行总时间超30秒"
	case appcommon.LiquidationTierP1 + "/safety_buffer_between":
		return "安全缓冲8%-15%"
	case appcommon.LiquidationTierP1 + "/exec_time_exceed":
		return "执行总时间超15秒"
	case appcommon.LiquidationTierP1 + "/consecutive_ioc_unfilled":
		return "连续3次IOC未成交"
	case appcommon.LiquidationTierP1 + "/price_drop_pct":
		return "近5秒价格跌幅>0.5%（实现：相对上次 mid 跌幅）"
	case appcommon.LiquidationTierP2 + "/safety_buffer_gte_and_exec_under":
		return "安全缓冲>15% 且 执行总时间<15秒（须同时满足）"
	default:
		return tier + " " + rule
	}
}

func prdMatchedLine(h TierRuleHit) string {
	item := prdItemShort(h.Tier, h.Rule)
	switch h.Rule {
	case "safety_buffer_lt", "safety_buffer_between":
		return fmt.Sprintf("%s：%s（safety_buffer=%s）", h.Tier, item, h.Actual)
	case "exec_time_exceed":
		return fmt.Sprintf("%s：%s（exec_sec=%s）", h.Tier, item, h.Actual)
	case "consecutive_ioc_unfilled":
		return fmt.Sprintf("%s：%s（连续=%s）", h.Tier, item, h.Actual)
	case "price_drop_pct":
		return fmt.Sprintf("%s：%s（price_drop_pct=%s）", h.Tier, item, h.Actual)
	case "safety_buffer_gte_and_exec_under":
		return fmt.Sprintf("%s：%s（%s）", h.Tier, item, h.Actual)
	default:
		return fmt.Sprintf("%s：%s", h.Tier, item)
	}
}

func buildResolvedTierWhy(resolved, previous string, matchedTiers, matchedPRD []string) string {
	if len(matchedTiers) == 0 {
		return fmt.Sprintf("未命中 P0/P1/P2 任一进入条件 → %s（默认 P2）；上一档 %s", resolved, previous)
	}
	why := fmt.Sprintf("命中档位 %s（产品表「进入条件」或关系）；只升不降，上一档 %s → 取最急 %s。",
		strings.Join(matchedTiers, "、"),
		previous,
		resolved,
	)
	if len(matchedPRD) > 0 {
		why += " 本次具体命中：" + strings.Join(matchedPRD, "；")
	}
	return why
}

func explainTierRules(cfg entity.LiquidationTierConfig, c TierContext) []TierRuleHit {
	var out []TierRuleHit
	tier := cfg.Tier

	if cfg.SafetyBufferLtPct != nil {
		th := *cfg.SafetyBufferLtPct
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "safety_buffer_lt",
			Matched:   c.SafetyBuffer.LessThan(th),
			Actual:    c.SafetyBuffer.String(),
			Threshold: "safety_buffer < " + th.String(),
		})
	}
	if cfg.SafetyBufferBetweenLowPct != nil && cfg.SafetyBufferBetweenHighPct != nil {
		lo, hi := *cfg.SafetyBufferBetweenLowPct, *cfg.SafetyBufferBetweenHighPct
		matched := !c.SafetyBuffer.LessThan(lo) && c.SafetyBuffer.LessThan(hi)
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "safety_buffer_between",
			Matched:   matched,
			Actual:    c.SafetyBuffer.String(),
			Threshold: lo.String() + " <= safety_buffer < " + hi.String(),
		})
	}
	if cfg.SafetyBufferGtePct != nil && cfg.ExecTimeUnderSec != nil {
		sb, sec := *cfg.SafetyBufferGtePct, *cfg.ExecTimeUnderSec
		matched := c.SafetyBuffer.GreaterThanOrEqual(sb) && c.ExecSec < sec
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "safety_buffer_gte_and_exec_under",
			Matched: matched,
			Actual:    "safety_buffer=" + c.SafetyBuffer.String() + ", exec_sec=" + decimal.NewFromInt(int64(c.ExecSec)).String(),
			Threshold: "safety_buffer >= " + sb.String() + " AND exec_sec < " + decimal.NewFromInt(int64(sec)).String(),
		})
	}
	if cfg.ExecTimeExceedSec != nil {
		th := *cfg.ExecTimeExceedSec
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "exec_time_exceed",
			Matched:   c.ExecSec > th,
			Actual:    decimal.NewFromInt(int64(c.ExecSec)).String(),
			Threshold: "exec_sec > " + decimal.NewFromInt(int64(th)).String(),
		})
	}
	if cfg.ConsecutiveIOCUnfilled != nil {
		th := *cfg.ConsecutiveIOCUnfilled
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "consecutive_ioc_unfilled",
			Matched:   c.ConsecutiveIOCZero >= int(th),
			Actual:    decimal.NewFromInt(int64(c.ConsecutiveIOCZero)).String(),
			Threshold: ">= " + decimal.NewFromInt(int64(th)).String(),
		})
	}
	if cfg.PriceDropPct != nil {
		th := *cfg.PriceDropPct
		out = append(out, TierRuleHit{
			Tier: tier, Rule: "price_drop_pct",
			Matched:   c.PriceDropPct.GreaterThan(th),
			Actual:    c.PriceDropPct.String(),
			Threshold: "price_drop_pct > " + th.String() + " (相对 last_mid 与本次 mid)",
		})
	}
	return out
}

// MarshalTierResolutionExplain JSON 序列化档位说明。
func MarshalTierResolutionExplain(ex TierResolutionExplain) (string, error) {
	b, err := json.Marshal(ex)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
