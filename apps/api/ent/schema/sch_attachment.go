package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Attachment is metadata of a file stored in S3/RustFS (or compatible).
// The file itself is never stored in the database.
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
		edge.From("transaction", Transaction.Type).
			Ref("attachments").
			Unique().
			Required(),
	}
}

func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bucket", "object_key").Unique(),
	}
}
