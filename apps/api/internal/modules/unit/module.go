package unit

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/unit/controller"
	"github.com/paroki/domus/api/internal/modules/unit/repository"
	"github.com/paroki/domus/api/internal/modules/unit/service"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
)

type Config struct {
	EntCli *ent.Client
	Authz  *authz.Enforcer
	Log    *slog.Logger
}

type Module struct {
	config     Config
	ctrl       controller.DioceseController
	parishCtrl controller.ParishController
}

func New(config Config) Module {
	repo := repository.NewDioceseRepository(config.EntCli, config.Log)
	svc := service.NewDioceseService(repo, config.Log)
	ctrl := controller.NewDioceseController(svc)

	parishRepo := repository.NewParishRepository(config.EntCli, config.Log)
	parishSvc := service.NewParishService(parishRepo, config.Log)
	parishCtrl := controller.NewParishController(parishSvc)

	return Module{config, ctrl, parishCtrl}
}

func (m Module) InitRoutes(r fiber.Router) {
	az := m.config.Authz
	perm := func(obj authz.Resource, act authz.Action) fiber.Handler {
		return httpx.RequirePermission(az, httpx.SystemScope, obj, act)
	}

	read := perm(authz.ResourceDiocese, authz.ActionRead)
	write := perm(authz.ResourceDiocese, authz.ActionWrite)
	r.Get("/dioceses", read, m.ctrl.GetAll)
	r.Get("/dioceses/:id", read, m.ctrl.GetByID)
	r.Post("/dioceses", write, m.ctrl.Create)
	r.Put("/dioceses/:id", write, m.ctrl.Update)
	r.Delete("/dioceses/:id", write, m.ctrl.Delete)

	readParish := perm(authz.ResourceParish, authz.ActionRead)
	writeParish := perm(authz.ResourceParish, authz.ActionWrite)
	r.Get("/parishes", readParish, m.parishCtrl.GetAll)
	r.Get("/parishes/:id", readParish, m.parishCtrl.GetByID)
	r.Post("/parishes", writeParish, m.parishCtrl.Create)
	r.Put("/parishes/:id", writeParish, m.parishCtrl.Update)
	r.Delete("/parishes/:id", writeParish, m.parishCtrl.Delete)
}
