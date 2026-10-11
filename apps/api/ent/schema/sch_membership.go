// ent/schema/membership.go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Membership struct{ ent.Schema }

func (Membership) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty(), // = auth.user.id (text), tanpa FK
		field.Enum("scope_type").Values("system", "diocese", "parish", "unit"),
		field.Int64("scope_id").NonNegative(), // 0 for scope_type=system
		field.String("role").NotEmpty(), // e.g. superadmin, finance-writer
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Membership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "scope_type", "scope_id", "role").Unique(),
		index.Fields("user_id"),                // /me
		index.Fields("scope_type", "scope_id"), // daftar admin/anggota per scope
	}
}
