package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paroki/domus/api/internal/config"
)

func main() {
	cfg := config.GetConfig()
	logger := config.GetLogger(cfg)

	config.WaitForJWKS(cfg, logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fiber := config.GetFiber(cfg, logger)
	entcli := config.GetEntClient(cfg)
	store := config.GetStorage(cfg, logger)
	state := config.State{
		Ent:      entcli,
		Authz:    config.GetAuthz(entcli),
		Config:   cfg,
		FiberApp: fiber,
		Log:      logger,
		Storage:  store,
	}
	config.Bootstrap(state)

	go func() { <-ctx.Done(); _ = fiber.Shutdown() }()
	log.Fatal(fiber.Listen(fmt.Sprintf(":%d", cfg.Port)))
}
