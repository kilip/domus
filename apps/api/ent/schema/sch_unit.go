package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Unit holds the schema definition for the Unit entity.
type Unit struct {
	ent.Schema
}

func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixins{},
	}
}

// Fields of the Unit.
func (Unit) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").NotEmpty(),
		field.String("description").Nillable(),
	}
}

// Edges of the Unit.
func (Unit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Unit.Type).
			From("parent").
			Unique(),
	}
}
