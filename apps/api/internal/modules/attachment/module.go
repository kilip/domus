// Package attachment provides reusable file attachments for any module.
//
// An owning module mounts the routes with Mount, supplying an OwnerResolver
// that maps (parishId, ownerId) to the owner's workspace. Permission is checked
// at the owner's workspace scope through Casbin.
package attachment

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/storage"
)

type Config struct {
	EntCli  *ent.Client
	Authz   *authz.Enforcer
	Storage storage.Storage
	Bucket  string
	Log     *slog.Logger
}

// Spec describes how an owning module exposes attachments.
type Spec struct {
	// OwnerType is stored on each attachment, e.g. "transaction".
	OwnerType string
	// Path is the route (relative to the router passed to Mount), e.g.
	// "/transactions/:id/attachments". Download is Path + "/:attachmentId/download".
	Path string
	// Resource is the Casbin object checked (e.g. authz.ResourceFinance).
	Resource authz.Resource
	Resolver OwnerResolver
	// ParishParam / OwnerParam default to "parishId" / "id".
	ParishParam string
	OwnerParam  string
	// ParishFunc optionally resolves parish ID (e.g. from context cookie) instead of route param.
	ParishFunc func(c fiber.Ctx) (int, error)
}

type Module struct {
	config Config
	ctl    Controller
}

func New(config Config) Module {
	repo := NewRepository(config.EntCli)
	svc := NewService(repo, config.Storage, config.Bucket, config.Log)
	return Module{config: config, ctl: NewController(svc)}
}

// Mount registers upload/list/download routes on r.
func (m Module) Mount(r fiber.Router, spec Spec) {
	if spec.ParishParam == "" {
		spec.ParishParam = "parishId"
	}
	if spec.OwnerParam == "" {
		spec.OwnerParam = "id"
	}
	read := m.guard(spec, authz.ActionRead)
	write := m.guard(spec, authz.ActionWrite)

	r.Post(spec.Path, write, m.ctl.Upload(spec.OwnerType))
	r.Get(spec.Path, read, m.ctl.List(spec.OwnerType))
	r.Get(spec.Path+"/:attachmentId/download", read, m.ctl.Download(spec.OwnerType))
}

type ownerKey struct{}

func ownerFrom(c fiber.Ctx) Owner {
	o, _ := c.Locals(ownerKey{}).(Owner)
	return o
}

// guard resolves the owner and checks the permission at the owner's workspace scope.
func (m Module) guard(spec Spec, act authz.Action) fiber.Handler {
	return func(c fiber.Ctx) error {
		var parishID int
		var err error
		if spec.ParishFunc != nil {
			parishID, err = spec.ParishFunc(c)
		} else {
			parishID, err = parseParish(c, spec.ParishParam)
		}
		if err != nil {
			return err
		}
		ownerID, err := uuid.Parse(c.Params(spec.OwnerParam))
		if err != nil {
			return core.ErrInvalidID
		}

		owner, err := spec.Resolver(c.Context(), parishID, ownerID)
		if err != nil {
			return err
		}

		user := core.UserFromContext(c)
		scope := authz.Scope{Type: authz.ScopeWorkspace, ID: int64(owner.WorkspaceID)}
		ok, err := m.config.Authz.Can(c.Context(), user.ID.String(), scope, spec.Resource, act)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrForbidden
		}

		c.Locals(ownerKey{}, *owner)
		return c.Next()
	}
}
