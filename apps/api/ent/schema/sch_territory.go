package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Territory holds the schema definition for the Territory entity.
//
// A Territory is a geographic area of a parish (e.g. deanery, station, region,
// neighborhood). It answers "where does a parishioner live?". Organizational bodies
// (boards, sections, categorical groups) are modeled by Unit, not here.
//
// Rules that cannot be expressed in the database are enforced in the service
// layer:
//   - the parent must belong to the same parish as the child
//   - the tree must not contain cycles
//   - a territory that still has children must not be deleted; deactivate it
//     (status = inactive) instead
type Territory struct {
	ent.Schema
}

func (Territory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

// Fields of the Territory.
func (Territory) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").NotEmpty(),
		// Short parish-defined code, unique per parish when set.
		field.String("code").Optional().Nillable(),
		field.String("kind").NotEmpty(),
		field.String("description").Optional().Nillable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

// Edges of the Territory.
func (Territory) Edges() []ent.Edge {
	return []ent.Edge{
		// Every territory belongs to exactly one parish.
		edge.From("parish", Parish.Type).
			Ref("territories").
			Unique().
			Required(),
		// Geographic tree (adjacency list, at most 3 levels deep).
		edge.To("children", Territory.Type).
			From("parent").
			Unique(),
		// Organizational body of this territory (Unit.kind = territorial).
		// A territory has at most one such unit.
		edge.From("unit", Unit.Type).
			Ref("territory").
			Unique(),
	}
}

// Indexes of the Territory.
func (Territory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Edges("parish").Unique(),
		index.Fields("kind", "name").Edges("parish").Unique(),
	}
}
