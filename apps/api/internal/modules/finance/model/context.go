package model

import "context"

const FinanceContextKey = "financeContext"

type FinanceContext struct {
	ParishID    int `json:"activeParishId"`
	WorkspaceID int `json:"activeWorkspaceId"`
}

type localGetter interface {
	Locals(key any, value ...any) any
}

func FinanceContextFromContext(ctx context.Context) (FinanceContext, bool) {
	if ctx == nil {
		return FinanceContext{}, false
	}
	if g, ok := ctx.(localGetter); ok {
		if fc, ok := g.Locals(FinanceContextKey).(FinanceContext); ok {
			return fc, true
		}
	}
	fc, ok := ctx.Value(FinanceContextKey).(FinanceContext)
	return fc, ok
}

func WithFinanceContext(ctx context.Context, fc FinanceContext) context.Context {
	return context.WithValue(ctx, FinanceContextKey, fc)
}
