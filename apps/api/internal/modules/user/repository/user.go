package repository

import (
	"context"
	"log/slog"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/membership"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/user"
	"github.com/paroki/domus/api/ent/workspace"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/user/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

// UserRepository handles database queries and mutations for users and their memberships.
type UserRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

// NewUserRepository constructs a new UserRepository instance.
func NewUserRepository(entcli *ent.Client, log *slog.Logger) UserRepository {
	return UserRepository{entcli: entcli, log: log}
}

// FindParishDioceseID looks up the parent diocese ID for a given parish.
func (r UserRepository) FindParishDioceseID(ctx context.Context, parishID int) (int, error) {
	p, err := r.entcli.Parish.Query().
		Where(parish.IDEQ(parishID)).
		WithDiocese().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, model.ErrScopeNotFound
		}
		return 0, err
	}
	if p.Edges.Diocese == nil {
		return 0, model.ErrScopeNotFound
	}
	return p.Edges.Diocese.ID, nil
}

// FindWorkspaceParishID looks up the parent parish ID for a given workspace.
func (r UserRepository) FindWorkspaceParishID(ctx context.Context, workspaceID int) (int, error) {
	w, err := r.entcli.Workspace.Query().
		Where(workspace.IDEQ(workspaceID)).
		WithParish().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, model.ErrScopeNotFound
		}
		return 0, err
	}
	if w.Edges.Parish == nil {
		return 0, model.ErrScopeNotFound
	}
	return w.Edges.Parish.ID, nil
}

// UpsertInvite atomically finds or creates a user with status "invited" and inserts new memberships.
// It returns the response model, a boolean indicating if a new user was created, and the slice of newly created memberships.
func (r UserRepository) UpsertInvite(
	ctx context.Context,
	email string,
	name string,
	memberships []model.MembershipInput,
) (*model.UserInviteResponse, bool, []model.MembershipInput, error) {
	var resp model.UserInviteResponse
	var isNew bool
	var newMemberships []model.MembershipInput

	err := database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
		// 1. Look for existing non-deleted user by email (lowercase)
		existingUser, err := tx.User.Query().
			Where(user.EmailEqualFold(email)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}

		var u *ent.User
		if existingUser != nil {
			u = existingUser
			isNew = false
		} else {
			isNew = true
			newID := core.GenerateID()
			u, err = tx.User.Create().
				SetID(newID).
				SetName(name).
				SetEmail(email).
				SetStatus(user.StatusInvited).
				Save(ctx)
			if err != nil {
				return err
			}
		}

		uidStr := u.ID.String()

		// 2. Insert memberships, deduplicating with existing ones
		for _, m := range memberships {
			scopeType := membership.ScopeType(m.ScopeType)
			exists, err := tx.Membership.Query().
				Where(
					membership.UserIDEQ(uidStr),
					membership.ScopeTypeEQ(scopeType),
					membership.ScopeIDEQ(m.ScopeID),
					membership.RoleEQ(m.Role),
				).
				Exist(ctx)
			if err != nil {
				return err
			}

			if !exists {
				_, err = tx.Membership.Create().
					SetUserID(uidStr).
					SetScopeType(scopeType).
					SetScopeID(m.ScopeID).
					SetRole(m.Role).
					Save(ctx)
				if err != nil {
					return err
				}
				newMemberships = append(newMemberships, m)
			}
		}

		// 3. Query all current memberships for the user to populate response
		allMemberships, err := tx.Membership.Query().
			Where(membership.UserIDEQ(uidStr)).
			All(ctx)
		if err != nil {
			return err
		}

		respMemberships := make([]model.MembershipResponse, 0, len(allMemberships))
		for _, mem := range allMemberships {
			respMemberships = append(respMemberships, model.MembershipResponse{
				ID:        mem.ID,
				ScopeType: string(mem.ScopeType),
				ScopeID:   mem.ScopeID,
				Role:      mem.Role,
				CreatedAt: mem.CreatedAt,
			})
		}

		resp = model.UserInviteResponse{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			Status:      string(u.Status),
			Memberships: respMemberships,
		}

		return nil
	})

	if err != nil {
		return nil, false, nil, err
	}

	return &resp, isNew, newMemberships, nil
}

