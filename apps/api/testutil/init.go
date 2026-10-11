package testutil

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/config"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=cmocks/mock_slog_handler.go -package=cmocks log/slog Handler

var jwksMock JWKSMock
var cfg config.Config
var state config.State
var api *fiber.App
var logger *slog.Logger
var entClient *ent.Client

func loadEnvForTest() {
	// _, b, _, _ returns the absolute path of the current file
	_, b, _, _ := runtime.Caller(0)

	// Get the directory of this file, then navigate up to the project root
	// Adjust "../" depending on how deep this file is nested from the root
	basepath := filepath.Join(filepath.Dir(b), "../.env")

	_ = godotenv.Load(basepath)
}

func init() {
	loadEnvForTest()

	cfg = config.GetConfig()

	mock := NewJWKSMock()
	jwksMock = *mock

	// force test environment config
	cfg.JWKSUrl = mock.URL

	_, b, _, _ := runtime.Caller(0)
	logPath := filepath.Join(filepath.Dir(b), "../tmp/test.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	var logWriter io.Writer = logFile
	if err != nil {
		logWriter = os.Stdout
	}

	logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	api = config.GetFiber(cfg, logger)
	entClient = createTestDB()

	config.ConfigureEntCli(entClient)
	//mockStorage := service.NewMockStorageService()
	state = config.State{
		FiberApp: api,
		Config:   cfg,
		Log:      logger,
		Ent:      entClient,
		Authz:    config.GetAuthz(entClient),
	}

	config.Bootstrap(state)
}

func GetState() config.State {
	return state
}
