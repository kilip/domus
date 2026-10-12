package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/user/model"
	"github.com/paroki/domus/api/internal/platform/authz"
)

// UserRepository defines the persistence methods required by UserService.
type UserRepository interface {
	FindParishDioceseID(ctx context.Context, parishID int) (int, error)
	FindWorkspaceParishID(ctx context.Context, workspaceID int) (int, error)
	UpsertInvite(ctx context.Context, email, name string, memberships []model.MembershipInput) (*model.UserInviteResponse, bool, []model.MembershipInput, error)
	FindResolveUserByEmail(ctx context.Context, email string) (*model.ResolveUserResponse, error)
}

// UserService coordinates user invitation and membership management with authorization enforcement.
type UserService struct {
	repo  UserRepository
	authz *authz.Enforcer
	log   *slog.Logger
}

// NewUserService constructs a new UserService instance.
func NewUserService(repo UserRepository, authzEnforcer *authz.Enforcer, log *slog.Logger) UserService {
	return UserService{
		repo:  repo,
		authz: authzEnforcer,
		log:   log,
	}
}

// Invite verifies caller authorization for every requested membership, creates or updates the user,
// and registers any new memberships in the Casbin authorization enforcer.
func (s UserService) Invite(ctx context.Context, caller core.AuthenticatedUser, req model.InviteUserRequest) (*model.UserInviteResponse, bool, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	callerID := caller.ID.String()

	// 1. Authorize every membership. If any membership is not allowed, reject the whole request.
	for _, m := range req.Memberships {
		// Never allow granting the superadmin role through this endpoint.
		if m.Role == authz.RoleSuperadmin {
			return nil, false, model.ErrSuperadminNotAllowed
		}

		allowed, err := s.canManageMembership(ctx, callerID, m)
		if err != nil {
			return nil, false, err
		}
		if !allowed {
			return nil, false, core.ErrForbidden
		}
	}

	// 2. Perform database upsert in a transaction
	resp, isNew, newMemberships, err := s.repo.UpsertInvite(ctx, cleanEmail, strings.TrimSpace(req.Name), req.Memberships)
	if err != nil {
		return nil, false, err
	}

	// 3. Register newly added memberships into Casbin enforcer
	for _, nm := range newMemberships {
		scope := authz.Scope{
			Type: authz.ScopeType(nm.ScopeType),
			ID:   nm.ScopeID,
		}
		if err := s.authz.AddMembership(resp.ID.String(), nm.Role, scope); err != nil {
			s.log.Error("failed to register membership in casbin enforcer",
				"userId", resp.ID.String(),
				"role", nm.Role,
				"scope", scope.Domain(),
				"error", err,
			)
		}
	}

	return resp, isNew, nil
}

// canManageMembership checks whether the caller has permission to grant roles within the target scope.
func (s UserService) canManageMembership(ctx context.Context, callerID string, m model.MembershipInput) (bool, error) {
	// 1. Global superadmin at system:0 can grant memberships at any scope.
	canSystem, err := s.authz.Can(ctx, callerID, authz.System, authz.ResourceUser, authz.ActionWrite)
	if err != nil {
		return false, err
	}
	if canSystem {
		return true, nil
	}

	switch authz.ScopeType(m.ScopeType) {
	case authz.ScopeSystem:
		// Only superadmin can manage system scope, already checked above.
		return false, nil

	case authz.ScopeDiocese:
		scope := authz.Scope{Type: authz.ScopeDiocese, ID: m.ScopeID}
		return s.authz.Can(ctx, callerID, scope, authz.ResourceUser, authz.ActionWrite)

	case authz.ScopeParish:
		// Allowed if caller can write users at parish:X
		parishScope := authz.Scope{Type: authz.ScopeParish, ID: m.ScopeID}
		canParish, err := s.authz.Can(ctx, callerID, parishScope, authz.ResourceUser, authz.ActionWrite)
		if err != nil {
			return false, err
		}
		if canParish {
			return true, nil
		}

		// Or allowed if caller can write users at the diocese that owns parish X
		dioceseID, err := s.repo.FindParishDioceseID(ctx, int(m.ScopeID))
		if err != nil {
			return false, err
		}
		dioceseScope := authz.Scope{Type: authz.ScopeDiocese, ID: int64(dioceseID)}
		return s.authz.Can(ctx, callerID, dioceseScope, authz.ResourceUser, authz.ActionWrite)

	case authz.ScopeWorkspace:
		// Resolve the workspace's parish first
		parishID, err := s.repo.FindWorkspaceParishID(ctx, int(m.ScopeID))
		if err != nil {
			return false, err
		}

		// Check authority using the parish rule
		parishScope := authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)}
		canParish, err := s.authz.Can(ctx, callerID, parishScope, authz.ResourceUser, authz.ActionWrite)
		if err != nil {
			return false, err
		}
		if canParish {
			return true, nil
		}

		dioceseID, err := s.repo.FindParishDioceseID(ctx, parishID)
		if err != nil {
			return false, err
		}
		dioceseScope := authz.Scope{Type: authz.ScopeDiocese, ID: int64(dioceseID)}
		return s.authz.Can(ctx, callerID, dioceseScope, authz.ResourceUser, authz.ActionWrite)

	default:
		return false, core.ErrForbidden
	}
}

// Resolve finds an invited or active non-deleted user by email and returns their resolve details.
func (s UserService) Resolve(ctx context.Context, email string) (*model.ResolveUserResponse, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	return s.repo.FindResolveUserByEmail(ctx, cleanEmail)
}
