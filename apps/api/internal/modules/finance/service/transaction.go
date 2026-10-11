package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, req model.CreateTransactionRequest, number string, txDate time.Time) (*model.TransactionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.TransactionResponse, error)
	Search(ctx context.Context, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error)
	Void(ctx context.Context, id uuid.UUID, reversalDate *time.Time, reason *string) (*model.TransactionResponse, error)
}

type AccountLookupRepository interface {
	GetEntsByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Account, error)
}

type TransactionService struct {
	txRepo      TransactionRepository
	accountRepo AccountLookupRepository
	log         *slog.Logger
}

func NewTransactionService(
	txRepo TransactionRepository,
	accountRepo AccountLookupRepository,
	log *slog.Logger,
) TransactionService {
	return TransactionService{
		txRepo:      txRepo,
		accountRepo: accountRepo,
		log:         log,
	}
}

func (s TransactionService) Create(ctx context.Context, req model.CreateTransactionRequest) (*model.TransactionResponse, error) {
	// Rule 1: Parse date and validate number pattern if supplied.
	txDate, err := model.ParseDate(req.Date)
	if err != nil {
		return nil, model.ErrInvalidDateFormat
	}

	if req.Number != nil && *req.Number != "" {
		if !model.TransactionNumberRegex.MatchString(*req.Number) {
			return nil, model.ErrInvalidTxNumberFormat
		}
	}

	// Rule 2: sum(debit) == sum(credit) and both sums > 0. At least one debit posting and one credit posting.
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
			return nil, model.ErrInvalidPostingSide
		}
	}

	if debitCount == 0 || creditCount == 0 {
		return nil, model.ErrTxMissingPostings
	}

	if debitSum != creditSum || debitSum <= 0 {
		return nil, model.ErrTxUnbalanced
	}

	// Rule 3: Load distinct accounts via GetEntsByIDs(ids). Any missing account -> 400.
	accountIDsMap := make(map[uuid.UUID]struct{})
	var distinctAccountIDs []uuid.UUID
	for _, p := range req.Postings {
		if _, exists := accountIDsMap[p.AccountID]; !exists {
			accountIDsMap[p.AccountID] = struct{}{}
			distinctAccountIDs = append(distinctAccountIDs, p.AccountID)
		}
	}

	accounts, err := s.accountRepo.GetEntsByIDs(ctx, distinctAccountIDs)
	if err != nil {
		return nil, err
	}
	if len(accounts) != len(distinctAccountIDs) {
		return nil, model.ErrAccountsNotFound
	}

	// Rule 4: Every referenced account must have status = active and postable = true -> else 400.
	for _, acc := range accounts {
		if acc.Status != account.StatusActive {
			return nil, model.ErrAccountNotActive
		}
		if !acc.Postable {
			return nil, model.ErrAccountIsHeader
		}
	}

	// Rule 5: Call the repository Create. Duplicate supplied number -> 409.
	number := ""
	if req.Number != nil {
		number = *req.Number
	}

	return s.txRepo.Create(ctx, req, number, txDate)
}

func (s TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*model.TransactionResponse, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s TransactionService) Search(ctx context.Context, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error) {
	if req.From != nil && *req.From != "" {
		if _, err := model.ParseDate(*req.From); err != nil {
			return nil, model.ErrInvalidFromDate
		}
	}
	if req.To != nil && *req.To != "" {
		if _, err := model.ParseDate(*req.To); err != nil {
			return nil, model.ErrInvalidToDate
		}
	}
	return s.txRepo.Search(ctx, req)
}

func (s TransactionService) Void(ctx context.Context, id uuid.UUID, req model.VoidTransactionRequest) (*model.TransactionResponse, error) {
	var reversalDate *time.Time
	if req.Date != nil && *req.Date != "" {
		parsedDate, err := model.ParseDate(*req.Date)
		if err != nil {
			return nil, model.ErrInvalidDateFormat
		}
		reversalDate = &parsedDate
	}

	return s.txRepo.Void(ctx, id, reversalDate, req.Reason)
}
