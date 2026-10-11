package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Posting is one debit/credit leg of an Entry. Immutable, no soft delete.
//
// amount is a positive integer in minor units; direction is in side.
type Posting struct {
	ent.Schema
}

func (Posting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
	}
}

func (Posting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("amount").Positive().Immutable(),
		field.Enum("side").Values("debit", "credit").Immutable(),
		field.String("memo").Optional().Nillable().Immutable(),
	}
}

func (Posting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entry", Entry.Type).
			Ref("postings").
			Unique().
			Required().
			Immutable(),
		edge.To("account", Account.Type).
			Unique().
			Required().
			Immutable(),
	}
}

func (Posting) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("account"),
	}
}
