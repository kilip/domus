package system

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/system/controller"
	"github.com/paroki/domus/api/internal/modules/system/repository"
	"github.com/paroki/domus/api/internal/modules/system/service"
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
	workspaceCtrl controller.WorkspaceController
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

	workspaceRepo := repository.NewWorkspaceRepository(config.EntCli, config.Log)
	workspaceSvc := service.NewWorkspaceService(workspaceRepo, config.Log)
	workspaceCtrl := controller.NewWorkspaceController(workspaceSvc)

	return Module{config, ctrl, parishCtrl, territoryCtrl, workspaceCtrl}
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

	readWorkspace := perm(authz.ResourceWorkspace, authz.ActionRead)
	writeWorkspace := perm(authz.ResourceWorkspace, authz.ActionWrite)
	r.Post("/workspaces/search", readWorkspace, m.workspaceCtrl.Search)
	r.Get("/workspaces/:id", readWorkspace, m.workspaceCtrl.GetByID)
	r.Post("/workspaces", writeWorkspace, m.workspaceCtrl.Create)
	r.Put("/workspaces/:id", writeWorkspace, m.workspaceCtrl.Update)
	r.Delete("/workspaces/:id", writeWorkspace, m.workspaceCtrl.Delete)
}
