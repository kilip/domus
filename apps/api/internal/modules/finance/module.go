package finance

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/attachment"
	"github.com/paroki/domus/api/internal/modules/finance/controller"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/paroki/domus/api/internal/modules/finance/service"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
	"github.com/paroki/domus/api/internal/platform/storage"
)

type Config struct {
	EntCli     *ent.Client
	Authz      *authz.Enforcer
	Storage    storage.Storage
	Bucket     string
	CookieName string
	Log        *slog.Logger
}

type Module struct {
	config       Config
	unitCtrl     controller.FinanceUnitController
	activateCtrl controller.ActivateController
	accountCtrl  controller.AccountController
	txCtrl       controller.TransactionController
	attMod       attachment.Module
	reportCtrl   controller.ReportController
}

func New(config Config) Module {
	if config.Bucket == "" {
		config.Bucket = "domus-finance"
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCookieName
	}

	unitRepo := repository.NewFinanceUnitRepository(config.EntCli, config.Log)
	unitSvc := service.NewFinanceUnitService(unitRepo, config.Log)
	unitCtrl := controller.NewFinanceUnitController(unitSvc)

	activateRepo := repository.NewActivateRepository(config.EntCli, config.Log)
	activateSvc := service.NewActivateService(activateRepo, config.Authz, config.Log)
	activateCtrl := controller.NewActivateController(activateSvc, config.CookieName)

	accountRepo := repository.NewAccountRepository(config.EntCli, config.Log)
	accountSvc := service.NewAccountService(accountRepo, config.Log)
	accountCtrl := controller.NewAccountController(accountSvc)

	txRepo := repository.NewTransactionRepository(config.EntCli, config.Log)
	txSvc := service.NewTransactionService(txRepo, accountRepo, config.Log)
	txCtrl := controller.NewTransactionController(txSvc)

	attMod := attachment.New(attachment.Config{
		EntCli:  config.EntCli,
		Authz:   config.Authz,
		Storage: config.Storage,
		Bucket:  config.Bucket,
		Log:     config.Log,
	})

	reportRepo := repository.NewReportRepository(config.EntCli, config.Log)
	reportSvc := service.NewReportService(reportRepo, config.Log)
	reportCtrl := controller.NewReportController(reportSvc)

	return Module{
		config:       config,
		unitCtrl:     unitCtrl,
		activateCtrl: activateCtrl,
		accountCtrl:  accountCtrl,
		txCtrl:       txCtrl,
		attMod:       attMod,
		reportCtrl:   reportCtrl,
	}
}

func (m Module) InitRoutes(r fiber.Router) {
	fin := r.Group("/fin")

	// Activation and unit listing do not require active context cookie
	fin.Get("/units", m.unitCtrl.GetAccessibleUnits)
	fin.Post("/activate", m.activateCtrl.Activate)

	// All other /fin routes require active finance context cookie
	fin.Use(RequireFinanceContext(m.config.CookieName))

	az := m.config.Authz
	perm := func(act authz.Action) fiber.Handler {
		return httpx.RequirePermission(az, FinanceUnitScope, authz.ResourceFinance, act)
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

	// Attachments are authorized at the transaction's unit scope, not parish.
	m.attMod.Mount(fin, attachment.Spec{
		OwnerType: repository.AttachmentOwnerTransaction,
		Path:      "/transactions/:id/attachments",
		Resource:  authz.ResourceFinance,
		Resolver:  repository.NewTransactionOwnerResolver(m.config.EntCli),
		ParishFunc: func(c fiber.Ctx) (int, error) {
			fc, ok := model.FinanceContextFromContext(c)
			if !ok || fc.ParishID <= 0 {
				return 0, ErrContextNotActive
			}
			return fc.ParishID, nil
		},
	})

	fin.Get("/reports/ledger", read, m.reportCtrl.GetLedger)
	fin.Get("/reports/trial-balance", read, m.reportCtrl.GetTrialBalance)
	fin.Get("/reports/collections", read, m.reportCtrl.GetCollections)
}

type parishMutator interface {
	SetParishID(int)
	ParishID() (int, bool)
}

type unitMutator interface {
	SetUnitID(int)
	UnitID() (int, bool)
}

func (m Module) InitEntCli(cli *ent.Client) {
	cli.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mut ent.Mutation) (ent.Value, error) {
			if fc, ok := model.FinanceContextFromContext(ctx); ok {
				if mut.Op().Is(ent.OpCreate) {
					if p, ok := mut.(parishMutator); ok {
						if _, exists := p.ParishID(); !exists && fc.ParishID > 0 {
							p.SetParishID(fc.ParishID)
						}
					}
					if u, ok := mut.(unitMutator); ok {
						if _, exists := u.UnitID(); !exists && fc.UnitID > 0 {
							u.SetUnitID(fc.UnitID)
						}
					}
				}
			}
			return next.Mutate(ctx, mut)
		})
	})
}
