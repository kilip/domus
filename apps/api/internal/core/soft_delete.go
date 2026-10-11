package core

import "context"

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that tells soft-delete interceptors and hooks
// to bypass filtering and soft-delete logic (e.g. for hard deletes or querying soft-deleted entities).
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// HasSkipSoftDelete reports whether the context has soft-delete skipping enabled.
func HasSkipSoftDelete(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if g, ok := ctx.(localGetter); ok {
		if skip, ok := g.Locals(softDeleteKey{}).(bool); ok {
			return skip
		}
	}
	skip, _ := ctx.Value(softDeleteKey{}).(bool)
	return skip
}
