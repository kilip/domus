package finance

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/finance/controller"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/paroki/domus/api/internal/modules/finance/service"
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
	config      Config
	accountCtrl controller.AccountController
}

func New(config Config) Module {
	accountRepo := repository.NewAccountRepository(config.EntCli, config.Log)
	accountSvc := service.NewAccountService(accountRepo, config.EntCli, config.Log)
	accountCtrl := controller.NewAccountController(accountSvc)

	return Module{
		config:      config,
		accountCtrl: accountCtrl,
	}
}

func (m Module) InitRoutes(r fiber.Router) {
	fin := r.Group("/parishes/:parishId/finance", httpx.RequireIntParam("parishId"))

	az := m.config.Authz
	perm := func(act authz.Action) fiber.Handler {
		return httpx.RequirePermission(az, httpx.ParishScope("parishId"), authz.ResourceFinance, act)
	}

	read := perm(authz.ActionRead)
	write := perm(authz.ActionWrite)

	fin.Post("/accounts", write, m.accountCtrl.Create)
	fin.Get("/accounts", read, m.accountCtrl.List)
	fin.Get("/accounts/:id", read, m.accountCtrl.GetByID)
	fin.Put("/accounts/:id", write, m.accountCtrl.Update)
	fin.Delete("/accounts/:id", write, m.accountCtrl.Delete)
}
