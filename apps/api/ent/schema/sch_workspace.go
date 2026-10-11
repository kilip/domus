package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workspace holds the schema definition for the Workspace entity.
//
// A Workspace is an organizational body or working space of a parish.
type Workspace struct {
	ent.Schema
}

func (Workspace) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

// Fields of the Workspace.
func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional().Nillable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

// Edges of the Workspace.
func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		// Every workspace belongs to exactly one parish.
		edge.From("parish", Parish.Type).
			Ref("workspaces").
			Unique().
			Required(),
	}
}

