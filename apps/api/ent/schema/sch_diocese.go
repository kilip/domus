package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Diocese holds the schema definition for the Diocese entity.
type Diocese struct {
	ent.Schema
}

func (Diocese) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

// Fields of the Diocese.
func (Diocese) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").NotEmpty(),
		field.Bool("active").Default(true),
	}
}

// Edges of the Diocese.
func (Diocese) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parishes", Parish.Type),
	}
}
