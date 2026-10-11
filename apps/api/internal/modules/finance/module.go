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
	Bucket  string
	Log     *slog.Logger
}

type Module struct {
	config      Config
	accountCtrl controller.AccountController
	txCtrl      controller.TransactionController
	attCtrl     controller.AttachmentController
	reportCtrl  controller.ReportController
}

func New(config Config) Module {
	if config.Bucket == "" {
		config.Bucket = "domus-finance"
	}

	accountRepo := repository.NewAccountRepository(config.EntCli, config.Log)
	accountSvc := service.NewAccountService(accountRepo, config.EntCli, config.Log)
	accountCtrl := controller.NewAccountController(accountSvc)

	txRepo := repository.NewTransactionRepository(config.EntCli, config.Log)
	txSvc := service.NewTransactionService(txRepo, accountRepo, config.EntCli, config.Log)
	txCtrl := controller.NewTransactionController(txSvc)

	attRepo := repository.NewAttachmentRepository(config.EntCli, config.Log)
	attSvc := service.NewAttachmentService(attRepo, config.Storage, config.Bucket, config.Log)
	attCtrl := controller.NewAttachmentController(attSvc)

	reportRepo := repository.NewReportRepository(config.EntCli, config.Log)
	reportSvc := service.NewReportService(reportRepo, config.Log)
	reportCtrl := controller.NewReportController(reportSvc)

	return Module{
		config:      config,
		accountCtrl: accountCtrl,
		txCtrl:      txCtrl,
		attCtrl:     attCtrl,
		reportCtrl:  reportCtrl,
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

	fin.Post("/transactions", write, m.txCtrl.Create)
	fin.Post("/transactions/search", read, m.txCtrl.Search)
	fin.Get("/transactions/:id", read, m.txCtrl.GetByID)
	fin.Post("/transactions/:id/void", write, m.txCtrl.Void)

	fin.Post("/transactions/:id/attachments", write, m.attCtrl.Upload)
	fin.Get("/transactions/:id/attachments", read, m.attCtrl.List)
	fin.Get("/transactions/:id/attachments/:attachmentId/download", read, m.attCtrl.Download)

	fin.Get("/reports/ledger", read, m.reportCtrl.GetLedger)
	fin.Get("/reports/trial-balance", read, m.reportCtrl.GetTrialBalance)
	fin.Get("/reports/collections", read, m.reportCtrl.GetCollections)
}
