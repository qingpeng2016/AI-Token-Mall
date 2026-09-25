package trace

import (
	"context"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/tool"
)

const (
	TraceID       = "trace_id"
	HeaderTraceID = "X-Trace-Id"
)

func GetTraceIdByCtx(ctx context.Context) string {
	data := ctx.Value(TraceID)
	traceId, ok := data.(string)
	if ok {
		return traceId
	}
	return ""
}

func GenerateTraceId() string {
	return tool.Random()
}
