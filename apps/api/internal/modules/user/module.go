package user

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/user/controller"
	"github.com/paroki/domus/api/internal/modules/user/repository"
	"github.com/paroki/domus/api/internal/modules/user/service"
	"github.com/paroki/domus/api/internal/platform/authz"
)

// Config provides dependencies required by the user module.
type Config struct {
	EntCli *ent.Client
	Authz  *authz.Enforcer
	Log    *slog.Logger
}

// Module encapsulates the controllers and services of the user module.
type Module struct {
	cfg        Config
	controller controller.UserController
}

// New creates and wires up a new user Module.
func New(cfg Config) Module {
	repo := repository.NewUserRepository(cfg.EntCli, cfg.Log)
	svc := service.NewUserService(repo, cfg.Authz, cfg.Log)
	ctrl := controller.NewUserController(svc)

	return Module{
		cfg:        cfg,
		controller: ctrl,
	}
}

// InitRoutes registers all user module endpoints on the Fiber router.
func (m Module) InitRoutes(r fiber.Router) {
	r.Post("/users", m.controller.Invite)
	r.Post("/internal/users/resolve", m.controller.Resolve)
}
