package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		SoftDeleteMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("auth_user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("name").NotEmpty(),
		field.String("email").NotEmpty(),
		field.String("avatar").Optional(),
		field.Enum("status").
			Values("invited", "active").
			Default("active"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at IS NULL"),
			),
		index.Fields("auth_user_id").
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at IS NULL"),
			),
	}
}

