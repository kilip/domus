package config

import (
	"context"
	"log/slog"
	"strings"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/membership"
	"github.com/paroki/domus/api/ent/user"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/database"
)

// SeedSuperadmin idempotently ensures that the superadmin user and system:0 membership exist.
// If SUPERADMIN_EMAIL is empty, seeding is skipped.
// New users are created with status=invited; existing active users are not downgraded.
func SeedSuperadmin(ctx context.Context, client *ent.Client, cfg Config, log *slog.Logger, enforcer ...*authz.Enforcer) error {
	rawEmail := strings.TrimSpace(cfg.SuperadminEmail)
	if rawEmail == "" {
		if log != nil {
			log.Info("SUPERADMIN_EMAIL not configured, skipping superadmin seeding")
		}
		return nil
	}

	cleanEmail := strings.ToLower(rawEmail)
	name := strings.TrimSpace(cfg.SuperadminName)
	if name == "" {
		name = "Super Admin"
	}

	var seededUser *ent.User

	err := database.WithTx(ctx, client, func(tx *ent.Tx) error {
		existingUser, err := tx.User.Query().
			Where(user.EmailEqualFold(cleanEmail)).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}

		if existingUser != nil {
			seededUser = existingUser
		} else {
			newID := core.GenerateID()
			u, err := tx.User.Create().
				SetID(newID).
				SetName(name).
				SetEmail(cleanEmail).
				SetStatus(user.StatusInvited).
				Save(ctx)
			if err != nil {
				return err
			}
			seededUser = u
		}

		uidStr := seededUser.ID.String()

		exists, err := tx.Membership.Query().
			Where(
				membership.UserIDEQ(uidStr),
				membership.ScopeTypeEQ(membership.ScopeTypeSystem),
				membership.ScopeIDEQ(0),
				membership.RoleEQ(authz.RoleSuperadmin),
			).
			Exist(ctx)
		if err != nil {
			return err
		}

		if !exists {
			_, err = tx.Membership.Create().
				SetUserID(uidStr).
				SetScopeType(membership.ScopeTypeSystem).
				SetScopeID(0).
				SetRole(authz.RoleSuperadmin).
				Save(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if log != nil {
			log.Error("failed to seed superadmin", "email", cleanEmail, "error", err)
		}
		return err
	}

	if len(enforcer) > 0 && enforcer[0] != nil && seededUser != nil {
		_ = enforcer[0].AddMembership(seededUser.ID.String(), authz.RoleSuperadmin, authz.System)
	}

	if log != nil && seededUser != nil {
		log.Info("superadmin seeded successfully", "email", cleanEmail, "userId", seededUser.ID.String(), "status", seededUser.Status)
	}

	return nil
}

