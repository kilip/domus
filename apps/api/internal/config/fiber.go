package config

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	stdhttp "net/http"

	"github.com/MicahParks/keyfunc/v3"
	jwtware "github.com/gofiber/contrib/v3/jwt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/paroki/domus/api/internal/platform/httpx"
	slogfiber "github.com/samber/slog-fiber"
)

func initMiddlewares(cfg Config, app *fiber.App, logger *slog.Logger) error {
	// init healthcheck
	// Use the default probe on the conventional endpoints
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))

	// cors config
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.TrustedOrigins,
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// jwks config
	jwks, err := keyfunc.NewDefaultCtx(context.Background(), []string{cfg.JWKSUrl})
	if err != nil {
		return fmt.Errorf("fetch JWKS from %s: %w", cfg.JWKSUrl, err)
	}
	app.Use(jwtware.New(jwtware.Config{
		Next: func(c fiber.Ctx) bool {
			return c.Method() == fiber.MethodOptions
		},
		KeyFunc:   jwks.KeyfuncCtx(context.Background()),
		Extractor: extractors.FromAuthHeader("Bearer"),
		ErrorHandler: func(c fiber.Ctx, err error) error {
			logger.WarnContext(c, "JWKS Error", "error", err.Error())
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		},
	}))

	return nil
}

type structValidator struct{ v *validator.Validate }

func (s *structValidator) Validate(out any) error { return s.v.Struct(out) }

func WaitForJWKS(cfg Config, logger *slog.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := checkJWKS(ctx, cfg.JWKSUrl, logger); err != nil {
		log.Fatalf("auth not ready: %v", err)
	}
}

func checkJWKS(ctx context.Context, url string, log *slog.Logger) error {
	client := &stdhttp.Client{Timeout: 3 * time.Second}
	for i := 0; ; i++ {
		req, _ := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err == nil {
			ok := resp.StatusCode == stdhttp.StatusOK
			resp.Body.Close()
			if ok {
				return nil
			}
			err = fmt.Errorf("status %d", resp.StatusCode)
		}
		wait := time.Duration(1<<min(i, 4)) * time.Second
		log.Warn("waiting for auth JWKS", "url", url, "in", wait, "error", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func GetFiber(cfg Config, logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:       12 * 1024 * 1024,
		StructValidator: &structValidator{validator.New(validator.WithRequiredStructEnabled())},
		ErrorHandler:    ErrorHandler(logger),
	})

	app.Use(slogfiber.New(logger))
	app.Use(recover.New())

	if err := initMiddlewares(cfg, app, logger); err != nil {
		log.Fatalf("can't initialize fiber middleware %s", err)
	}
	return app
}

func ErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		status, body := httpx.MapError(err)
		if status >= 500 {
			log.ErrorContext(c, "unhandled error", "error", err.Error())
		}
		return c.Status(status).JSON(httpx.ErrorResponse{
			Error: body,
			Meta:  httpx.NewMeta(c, ""),
		})
	}
}
