package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/auth"
	"github.com/paroki/domus/api/internal/modules/finance"
	"github.com/paroki/domus/api/internal/modules/system"
	"github.com/paroki/domus/api/internal/modules/user"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/storage"
)

type State struct {
	Ent      *ent.Client
	Authz    *authz.Enforcer
	Config   Config
	FiberApp *fiber.App
	Log      *slog.Logger
	Storage  storage.Storage
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

func initSystem(state State) {
	sysModule := system.New(system.Config{
		EntCli: state.Ent,
		Authz:  state.Authz,
		Log:    state.Log,
	})
	sysModule.InitRoutes(state.FiberApp)
}

func initFinance(state State) {
	finModule := finance.New(finance.Config{
		EntCli:  state.Ent,
		Authz:   state.Authz,
		Storage: state.Storage,
		Bucket:  state.Config.S3Bucket,
		Log:     state.Log,
	})
	finModule.InitRoutes(state.FiberApp)
	finModule.InitEntCli(state.Ent)
}

func initUser(state State) {
	userModule := user.New(user.Config{
		EntCli: state.Ent,
		Authz:  state.Authz,
		Log:    state.Log,
	})
	userModule.InitRoutes(state.FiberApp)
}

func Bootstrap(state State) {
	initAuth(state)
	initSystem(state)
	initFinance(state)
	initUser(state)
}

