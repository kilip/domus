package core

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperadmin UserRole = "superadmin"
	UserRoleUser       UserRole = "user"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuthenticatedUser struct {
	ID            uuid.UUID  `json:"id"`
	AuthID        *uuid.UUID `json:"authId,omitempty"`
	EmailVerified bool       `json:"emailVerified,omitempty"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Avatar        string     `json:"avatar,omitempty"`
	Roles         []UserRole `json:"roles"`
	Scope         string     `json:"scope,omitempty"`
}

func (u AuthenticatedUser) HasScope(requiredScope string) bool {
	if u.Scope == "" || requiredScope == "" {
		return true
	}
	fields := strings.Fields(u.Scope)
	return slices.Contains(fields, requiredScope)
}

const AUTH_USER_CONTEXT_KEY = "user"

type localGetter interface {
	Locals(key any, value ...any) any
}

func UserFromContext(ctx context.Context) AuthenticatedUser {
	if g, ok := ctx.(localGetter); ok {
		if u, ok := g.Locals(AUTH_USER_CONTEXT_KEY).(AuthenticatedUser); ok {
			return u
		}
	}
	u, _ := ctx.Value(AUTH_USER_CONTEXT_KEY).(AuthenticatedUser)
	return u
}

func ContextWithUser(ctx context.Context, u AuthenticatedUser) context.Context {
	return context.WithValue(ctx, AUTH_USER_CONTEXT_KEY, u)
}
