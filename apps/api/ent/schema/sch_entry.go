package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Entry is a journal entry (header). It is immutable: no soft delete,
// corrections are made with a reversal Entry (reverses -> original).
//
// Rules enforced in the service layer (single tx):
//   - sum(debit) == sum(credit) over its postings, at least 2 postings
//   - an Entry can be reversed at most once
type Entry struct {
	ent.Schema
}

func (Entry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
	}
}

func (Entry) Fields() []ent.Field {
	return []ent.Field{
		// Accounting date (may differ from the transaction date).
		field.Time("date").Immutable(),
		field.String("description").Optional().Nillable().Immutable(),
	}
}

func (Entry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("transaction", Transaction.Type).
			Ref("entries").
			Unique().
			Required().
			Immutable(),
		edge.To("postings", Posting.Type),
		// A reversal entry points to the entry it reverses.
		edge.To("reversed_by", Entry.Type).
			Unique().
			From("reverses").
			Unique(),
	}
}

func (Entry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("date"),
	}
}
