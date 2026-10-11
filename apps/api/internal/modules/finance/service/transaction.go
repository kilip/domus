package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/ent/parish"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, parishID int, req model.CreateTransactionRequest, number string, txDate time.Time) (*model.TransactionResponse, error)
	GetByID(ctx context.Context, parishID int, id uuid.UUID) (*model.TransactionResponse, error)
	Search(ctx context.Context, parishID int, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error)
}

type AccountLookupRepository interface {
	GetEntsByIDs(ctx context.Context, parishID int, ids []uuid.UUID) ([]*ent.Account, error)
}

type TransactionService struct {
	txRepo      TransactionRepository
	accountRepo AccountLookupRepository
	entcli      *ent.Client
	log         *slog.Logger
}

func NewTransactionService(
	txRepo TransactionRepository,
	accountRepo AccountLookupRepository,
	entcli *ent.Client,
	log *slog.Logger,
) TransactionService {
	return TransactionService{
		txRepo:      txRepo,
		accountRepo: accountRepo,
		entcli:      entcli,
		log:         log,
	}
}

func (s TransactionService) Create(ctx context.Context, parishID int, req model.CreateTransactionRequest) (*model.TransactionResponse, error) {
	// Rule 1: Parse date (D2) and validate number pattern if supplied.
	txDate, err := model.ParseDate(req.Date)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid date format: "+err.Error())
	}

	if req.Number != nil && *req.Number != "" {
		if !model.TransactionNumberRegex.MatchString(*req.Number) {
			return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid transaction number format")
		}
	}

	// Rule 2: Σ debit == Σ credit and both sums > 0. At least one debit posting and one credit posting.
	var debitSum, creditSum int64
	var debitCount, creditCount int
	for _, p := range req.Postings {
		switch p.Side {
		case "debit":
			debitSum += p.Amount
			debitCount++
		case "credit":
			creditSum += p.Amount
			creditCount++
		default:
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid posting side: must be debit or credit")
		}
	}

	if debitCount == 0 || creditCount == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "transaction must have at least one debit posting and one credit posting")
	}

	if debitSum != creditSum || debitSum <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "transaction is unbalanced: sum of debits must equal sum of credits and be greater than zero")
	}

	// Rule 3: Load distinct accounts via GetEntsByIDs(parishID, ids). Any missing or cross-parish account -> 400.
	accountIDsMap := make(map[uuid.UUID]struct{})
	var distinctAccountIDs []uuid.UUID
	for _, p := range req.Postings {
		if _, exists := accountIDsMap[p.AccountID]; !exists {
			accountIDsMap[p.AccountID] = struct{}{}
			distinctAccountIDs = append(distinctAccountIDs, p.AccountID)
		}
	}

	accounts, err := s.accountRepo.GetEntsByIDs(ctx, parishID, distinctAccountIDs)
	if err != nil {
		return nil, err
	}
	if len(accounts) != len(distinctAccountIDs) {
		return nil, fiber.NewError(fiber.StatusBadRequest, "one or more accounts not found in this parish")
	}

	// Rule 4: Every referenced account must have status = active and postable = true -> else 400.
	for _, acc := range accounts {
		if acc.Status != account.StatusActive {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("account %s is not active", acc.Code))
		}
		if !acc.Postable {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("account %s is a header account and cannot accept postings", acc.Code))
		}
	}

	// Rule 5: unitId, if supplied, must belong to the same parish -> else 400.
	if req.UnitID != nil {
		unitExists, err := s.entcli.Unit.Query().
			Where(
				entunit.IDEQ(*req.UnitID),
				entunit.HasParishWith(parish.IDEQ(parishID)),
			).
			Exist(ctx)
		if err != nil {
			return nil, err
		}
		if !unitExists {
			return nil, fiber.NewError(fiber.StatusBadRequest, "unit does not belong to this parish")
		}
	}

	// Rule 6: Call the repository Create. Duplicate supplied number -> 409.
	number := ""
	if req.Number != nil {
		number = *req.Number
	}

	return s.txRepo.Create(ctx, parishID, req, number, txDate)
}

func (s TransactionService) GetByID(ctx context.Context, parishID int, id uuid.UUID) (*model.TransactionResponse, error) {
	return s.txRepo.GetByID(ctx, parishID, id)
}

func (s TransactionService) Search(ctx context.Context, parishID int, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error) {
	if req.From != nil && *req.From != "" {
		if _, err := model.ParseDate(*req.From); err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid from date: "+err.Error())
		}
	}
	if req.To != nil && *req.To != "" {
		if _, err := model.ParseDate(*req.To); err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid to date: "+err.Error())
		}
	}
	return s.txRepo.Search(ctx, parishID, req)
}
