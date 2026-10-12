package auth

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/auth/controller"
	"github.com/paroki/domus/api/internal/modules/auth/middleware"
	"github.com/paroki/domus/api/internal/modules/auth/repository"
	"github.com/paroki/domus/api/internal/modules/auth/service"
)

type Config struct {
	RestIssuer   string
	RestAudience string
	EntCli       *ent.Client
	Log          *slog.Logger
}
type Module struct {
	config  Config
	userSvc service.UserService
}

func New(config Config) Module {
	userRepo := repository.NewUserRepository(config.EntCli, config.Log)
	userSvc := service.NewUserService(userRepo, config.Log)
	return Module{config, userSvc}
}

func (m Module) InitMiddleware(app *fiber.App) {
	app.Use(middleware.UserInjector(m.config.RestIssuer, m.config.RestAudience))
	app.Use(middleware.UserValidate(m.userSvc))
}

func (m Module) InitRoutes(r fiber.Router) {
	hc := controller.HealthController{}
	r.Get("/ping", hc.Ping)
}
