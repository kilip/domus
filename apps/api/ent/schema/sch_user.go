package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
	}
}

