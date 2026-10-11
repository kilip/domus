package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Parish holds the schema definition for the Parish entity.
type Parish struct {
	ent.Schema
}

// Fields of the Parish.
func (Parish) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name"),
		field.String("church").Optional().Nillable(),
		field.String("address").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
	}
}

// Edges of the Parish.
func (Parish) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("diocese", Diocese.Type).
			Ref("parishes").
			Unique(),
	}
}
