package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Transaction is the business document (collection, expense receipt, transfer).
// Its accounting effect lives in Entry/Posting.
//
// Rules enforced in the service layer:
//   - status flow: draft -> submitted -> approved -> posted -> void
//   - only draft transactions can be edited or soft-deleted
//   - posting creates one Entry; void creates a reversal Entry
//   - at most one active (non-reversed) Entry per transaction
//   - workspace must belong to the same parish
type Transaction struct {
	ent.Schema
}

func (Transaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		// Document number, unique per parish.
		field.String("number").NotEmpty(),
		field.Time("date"),
		field.Enum("kind").Values("income", "expense", "transfer", "adjustment"),
		field.String("description").Optional().Nillable(),
		field.Enum("status").
			Values("draft", "submitted", "approved", "posted", "void").
			Default("draft"),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parish", Parish.Type).Unique().Required(),
		edge.To("workspace", Workspace.Type).Unique(),
		edge.To("entries", Entry.Type),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("number").Edges("parish").Unique(),
		index.Fields("date"),
		index.Fields("status"),
	}
}
