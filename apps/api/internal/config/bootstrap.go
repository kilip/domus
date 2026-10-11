package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/auth"
	"github.com/paroki/domus/api/internal/modules/unit"
	"github.com/paroki/domus/api/internal/platform/authz"
)

type State struct {
	Ent      *ent.Client
	Authz    *authz.Enforcer
	Config   Config
	FiberApp *fiber.App
	Log      *slog.Logger
}

func initAuth(state State) {
	auth := auth.New(auth.Config{
		RestIssuer:   state.Config.AuthUrl,
		RestAudience: state.Config.AuthUrl,
		EntCli:       state.Ent,
		Log:          state.Log,
	})
	auth.InitMiddleware(state.FiberApp)
	auth.InitRoutes(state.FiberApp)
}

func initUnit(state State) {
	unitModule := unit.New(unit.Config{
		EntCli: state.Ent,
		Authz:  state.Authz,
		Log:    state.Log,
	})
	unitModule.InitRoutes(state.FiberApp)
}

func Bootstrap(state State) {
	initAuth(state)
	initUnit(state)
}
