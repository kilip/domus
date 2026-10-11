package finance

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
	"github.com/paroki/domus/api/internal/platform/storage"
)

type Config struct {
	EntCli  *ent.Client
	Authz   *authz.Enforcer
	Storage storage.Storage
	Log     *slog.Logger
}

type Module struct {
	config Config
}

func New(config Config) Module {
	return Module{config: config}
}

func (m Module) InitRoutes(r fiber.Router) {
	fin := r.Group("/parishes/:parishId/finance", httpx.RequireIntParam("parishId"))
	_ = fin
}
