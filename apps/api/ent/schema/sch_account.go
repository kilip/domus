package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Account is a node of the chart of accounts (COA) of a parish.
//
// Rules enforced in the service layer:
//   - parent and workspace must belong to the same parish as the account
//   - the tree must not contain cycles
//   - only postable accounts (postable = true) can receive postings
//   - an account that has postings or children must not be deleted;
//     deactivate it (status = inactive) instead
type Account struct {
	ent.Schema
}

func (Account) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
		SoftDeleteMixin{},
	}
}

func (Account) Fields() []ent.Field {
	return []ent.Field{
		// Short parish-defined code, unique per parish (e.g. 1-1001).
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("type").Values("asset", "liability", "equity", "income", "expense"),
		field.Enum("normal_balance").Values("debit", "credit"),
		// false for header/group accounts that only aggregate children.
		field.Bool("postable").Default(true),
		field.String("description").Optional().Nillable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

func (Account) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parish", Parish.Type).Unique().Required().Immutable(),
		edge.To("workspace", Workspace.Type).Unique().Required().Immutable(),
		edge.To("children", Account.Type).From("parent").Unique(),
		edge.From("postings", Posting.Type).Ref("account"),
	}
}

func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Edges("parish").Unique(),
	}
}
