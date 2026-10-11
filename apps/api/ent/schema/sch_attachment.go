package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Attachment is metadata of a file stored in S3/RustFS (or compatible).
// The file itself is never stored in the database.
//
// An attachment is generic: it belongs to an owner identified by
// (owner_type, owner_id), e.g. ("transaction", <tx uuid>). There is no
// foreign key to the owner table so any module can attach files.
// parish and unit are denormalized from the owner at creation time and are
// immutable (an owner's unit never changes); unit is the Casbin scope used
// to authorize access.
type Attachment struct {
	ent.Schema
}

func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
	}
}

func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		// Module-defined owner kind, e.g. "transaction".
		field.String("owner_type").NotEmpty().Immutable(),
		field.UUID("owner_id", uuid.UUID{}).Immutable(),
		field.String("bucket").NotEmpty(),
		field.String("object_key").NotEmpty(),
		field.String("filename").NotEmpty(),
		field.String("mime_type").NotEmpty(),
		field.Int64("size").NonNegative(),
		// Hex-encoded SHA-256 of the content.
		field.String("checksum").NotEmpty(),
	}
}

func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parish", Parish.Type).Unique().Required().Immutable(),
		edge.To("unit", Unit.Type).Unique().Required().Immutable(),
	}
}

func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bucket", "object_key").Unique(),
		index.Fields("owner_type", "owner_id"),
	}
}
