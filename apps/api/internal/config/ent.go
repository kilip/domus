package config

import (
	"context"
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // Registers the "pgx" driver
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
)

type creatorMutator interface {
	SetCreatedBy(uuid.UUID)
}

type updaterMutator interface {
	SetUpdatedBy(uuid.UUID)
}

func initEntCliHook(cli *ent.Client) {
	cli.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			user := core.UserFromContext(ctx)
			if user.ID != uuid.Nil {
				if m.Op().Is(ent.OpCreate) {
					if c, ok := m.(creatorMutator); ok {
						if _, exists := m.Field("createdBy"); !exists {
							c.SetCreatedBy(user.ID)
						}
					} else if _, exists := m.Field("createdBy"); !exists {
						_ = m.SetField("createdBy", user.ID)
					}

					if u, ok := m.(updaterMutator); ok {
						if _, exists := m.Field("updatedBy"); !exists {
							u.SetUpdatedBy(user.ID)
						}
					} else if _, exists := m.Field("updatedBy"); !exists {
						_ = m.SetField("updatedBy", user.ID)
					}
				} else if m.Op().Is(ent.OpUpdate | ent.OpUpdateOne) {
					if u, ok := m.(updaterMutator); ok {
						u.SetUpdatedBy(user.ID)
					} else {
						_ = m.SetField("updatedBy", user.ID)
					}
				}
			}
			return next.Mutate(ctx, m)
		})
	})
}

func initEntCliInjector(cli *ent.Client) {

}

func ConfigureEntCli(cli *ent.Client) {
	initEntCliHook(cli)
	initEntCliInjector(cli)
	initEntCliSoftDelete(cli)
}

func GetEntClient(cfg Config) *ent.Client {
	db, err := sql.Open("pgx", cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Error while connecting to database %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed to creating schema resources: %v", err)
	}

	ConfigureEntCli(client)
	return client
}
