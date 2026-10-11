package config

import (
	"context"
	"time"

	entgo "entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/intercept"
	"github.com/paroki/domus/api/internal/core"
)

type softDeleteMutation interface {
	SetOp(entgo.Op)
	SetDeletedAt(time.Time)
	WhereP(...func(*sql.Selector))
	Client() *ent.Client
}

var softDeletableQueryTypes = map[string]string{
	ent.TypeAccount:     "deleted_at",
	ent.TypeDiocese:     "deleted_at",
	ent.TypeParish:      "deleted_at",
	ent.TypeTerritory:   "deleted_at",
	ent.TypeUnit:        "deleted_at",
	ent.TypeTransaction: "deleted_at",
	ent.TypeUser:        "deleted_at",
}

func initEntCliSoftDelete(cli *ent.Client) {
	// Query interceptor: automatically filter out soft-deleted records unless SkipSoftDelete is set.
	cli.Intercept(intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if core.HasSkipSoftDelete(ctx) {
			return nil
		}
		if col, ok := softDeletableQueryTypes[q.Type()]; ok {
			q.WhereP(sql.FieldIsNull(col))
		}
		return nil
	}))

	// Mutation hook: convert Delete and DeleteOne operations to soft deletes.
	cli.Use(func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			if m.Op().Is(entgo.OpDelete | entgo.OpDeleteOne) {
				if core.HasSkipSoftDelete(ctx) {
					return next.Mutate(ctx, m)
				}
				if sm, ok := m.(softDeleteMutation); ok {
					sm.WhereP(sql.FieldIsNull("deleted_at"))
					sm.SetOp(entgo.OpUpdate)
					sm.SetDeletedAt(time.Now())
					user := core.UserFromContext(ctx)
					if user.ID != uuid.Nil {
						if u, ok := m.(updaterMutator); ok {
							u.SetUpdatedBy(user.ID)
						} else {
							_ = m.SetField("updatedBy", user.ID)
						}
					}
					return sm.Client().Mutate(ctx, m)
				}
			}
			return next.Mutate(ctx, m)
		})
	})
}
