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
	config        Config
	ctrl          controller.DioceseController
	parishCtrl    controller.ParishController
	territoryCtrl controller.TerritoryController
	unitCtrl      controller.UnitController
}

func New(config Config) Module {
	repo := repository.NewDioceseRepository(config.EntCli, config.Log)
	svc := service.NewDioceseService(repo, config.Log)
	ctrl := controller.NewDioceseController(svc)

	parishRepo := repository.NewParishRepository(config.EntCli, config.Log)
	parishSvc := service.NewParishService(parishRepo, config.Log)
	parishCtrl := controller.NewParishController(parishSvc)

	territoryRepo := repository.NewTerritoryRepository(config.EntCli, config.Log)
	territorySvc := service.NewTerritoryService(territoryRepo, config.Log)
	territoryCtrl := controller.NewTerritoryController(territorySvc)

	unitRepo := repository.NewUnitRepository(config.EntCli, config.Log)
	unitSvc := service.NewUnitService(unitRepo, territoryRepo, config.Log)
	unitCtrl := controller.NewUnitController(unitSvc)

	return Module{config, ctrl, parishCtrl, territoryCtrl, unitCtrl}
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

	readTerritory := perm(authz.ResourceTerritory, authz.ActionRead)
	writeTerritory := perm(authz.ResourceTerritory, authz.ActionWrite)
	r.Post("/territories/search", readTerritory, m.territoryCtrl.Search)
	r.Get("/territories/:id", readTerritory, m.territoryCtrl.GetByID)
	r.Post("/territories", writeTerritory, m.territoryCtrl.Create)
	r.Put("/territories/:id", writeTerritory, m.territoryCtrl.Update)
	r.Delete("/territories/:id", writeTerritory, m.territoryCtrl.Delete)

	readUnit := perm(authz.ResourceUnit, authz.ActionRead)
	writeUnit := perm(authz.ResourceUnit, authz.ActionWrite)
	r.Post("/units/search", readUnit, m.unitCtrl.Search)
	r.Get("/units/:id", readUnit, m.unitCtrl.GetByID)
	r.Post("/units", writeUnit, m.unitCtrl.Create)
	r.Put("/units/:id", writeUnit, m.unitCtrl.Update)
	r.Delete("/units/:id", writeUnit, m.unitCtrl.Delete)
}
