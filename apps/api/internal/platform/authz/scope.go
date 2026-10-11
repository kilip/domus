package authz

import "fmt"

type ScopeType string

const (
	ScopeSystem  ScopeType = "system"
	ScopeDiocese ScopeType = "diocese"
	ScopeParish  ScopeType = "parish"
	ScopeUnit    ScopeType = "unit"
)

// Scope is the casbin domain. It maps 1:1 to a Membership row's
// (scope_type, scope_id). There is no inheritance between scopes.
type Scope struct {
	Type ScopeType
	ID   int64
}

// System is the global scope, used for superadmin and for resources
// that are not owned by any diocese/parish/unit (e.g. diocese itself).
var System = Scope{Type: ScopeSystem, ID: 0}

func (s Scope) Domain() string {
	return fmt.Sprintf("%s:%d", s.Type, s.ID)
}
