package entity

// BeTrust 回调 end_reason 枚举（与 PRD 5.2 对齐）
const (
	EndReasonDebtCleared          = "debt_cleared"           // 债务已覆盖
	EndReasonCollateralExhausted  = "collateral_exhausted"   // 抵押卖尽仍有负债
	EndReasonCancelled            = "cancelled"              // 业务取消
	EndReasonFailed               = "failed"                 // 执行失败
	EndReasonHardBoundaryBreached = "hard_boundary_breached" // 触及硬边界（如价格/风控）
	EndReasonManual               = "manual"                 // 人工结束
)

func NormalizeEndReason(s string) string {
	switch s {
	case EndReasonDebtCleared, EndReasonCollateralExhausted, EndReasonCancelled,
		EndReasonFailed, EndReasonHardBoundaryBreached, EndReasonManual:
		return s
	default:
		return s
	}
}
