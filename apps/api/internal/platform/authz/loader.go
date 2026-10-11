package authz

import (
	"context"
	"fmt"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/membership"
)

// LoadMemberships loads every Membership row as a casbin "g" rule.
func LoadMemberships(ctx context.Context, cli *ent.Client, e *Enforcer) error {
	items, err := cli.Membership.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("load memberships: %w", err)
	}

	rules := make([][]string, 0, len(items))
	for _, m := range items {
		scope := Scope{Type: ScopeType(m.ScopeType), ID: m.ScopeID}
		rules = append(rules, []string{m.UserID, m.Role, scope.Domain()})
	}
	if len(rules) == 0 {
		return nil
	}

	if _, err := e.e.AddGroupingPolicies(rules); err != nil {
		return fmt.Errorf("load memberships: %w", err)
	}
	return nil
}

// LoadUserMemberships loads Membership rows for a specific user as casbin "g" rules.
func LoadUserMemberships(ctx context.Context, cli *ent.Client, e *Enforcer, userID string) error {
	items, err := cli.Membership.Query().
		Where(membership.UserID(userID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("load user memberships: %w", err)
	}

	for _, m := range items {
		if m.UserID == userID {
			scope := Scope{Type: ScopeType(m.ScopeType), ID: m.ScopeID}
			_ = e.AddMembership(m.UserID, m.Role, scope)
		}
	}
	return nil
}
