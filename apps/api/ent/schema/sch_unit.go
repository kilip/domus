package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Unit holds the schema definition for the Unit entity.
//
// A Unit is an organizational body of a parish: a parish council, a stasi or
// lingkungan board, a categorical group (OMK, WKRI, ...), a section or a
// temporary committee. It is NOT a geographic structure; geography is modeled
// by Territory. Officers and their terms of office are modeled separately.
//
// Rules that cannot be expressed in the database are enforced in the service
// layer:
//   - kind = territorial requires a territory; every other kind must not have one
//   - the parent and the territory must belong to the same parish as the unit
//   - the tree must not contain cycles
//   - a unit that still has children must not be deleted; deactivate it
//     (status = inactive) instead
type Unit struct {
	ent.Schema
}

func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

// Fields of the Unit.
func (Unit) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").NotEmpty(),
		// Short parish-defined code, unique per parish when set.
		field.String("code").Optional().Nillable(),
		field.Enum("kind").Values(
			"council",
			"territorial",
			"category",
			"section",
			"committee",
		),
		field.String("description").Optional().Nillable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

// Edges of the Unit.
func (Unit) Edges() []ent.Edge {
	return []ent.Edge{
		// Every unit belongs to exactly one parish.
		edge.From("parish", Parish.Type).
			Ref("units").
			Unique().
			Required(),
		// Organizational hierarchy, e.g. a section under a parish council.
		edge.To("children", Unit.Type).
			From("parent").
			Unique(),
		// Only set when kind = territorial; one unit per territory.
		edge.To("territory", Territory.Type).
			Unique(),
	}
}

// Indexes of the Unit.
func (Unit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Edges("parish").Unique(),
	}
}
