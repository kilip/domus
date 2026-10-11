package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
)

type ReportRepository interface {
	GetAccount(ctx context.Context, accountID uuid.UUID) (*ent.Account, error)
	ListAccounts(ctx context.Context) ([]*ent.Account, error)
	GetLedger(ctx context.Context, acc *ent.Account, from, to time.Time, cursorDate *time.Time, cursorPostingID *uuid.UUID, limit int) (model.LedgerResponse, string, error)
	GetTrialBalance(ctx context.Context, asOf time.Time) (model.TrialBalanceResponse, error)
	GetCollections(ctx context.Context, from, to time.Time, allowedAccountIDs map[uuid.UUID]bool) (model.CollectionsResponse, error)
}

type ReportService struct {
	repo ReportRepository
	log  *slog.Logger
}

func NewReportService(repo ReportRepository, log *slog.Logger) ReportService {
	return ReportService{
		repo: repo,
		log:  log,
	}
}

func (s ReportService) GetLedger(ctx context.Context, query model.LedgerQuery) (model.LedgerResponse, string, error) {
	if query.AccountID == uuid.Nil {
		return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "accountId is required")
	}

	acc, err := s.repo.GetAccount(ctx, query.AccountID)
	if err != nil {
		if ent.IsNotFound(err) {
			return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusNotFound, "account not found in this parish")
		}
		return model.LedgerResponse{}, "", err
	}

	if query.From == "" || query.To == "" {
		return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "from and to dates are required")
	}

	fromDate, err := model.ParseDate(query.From)
	if err != nil {
		return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "invalid from date: "+err.Error())
	}

	toDate, err := model.ParseDate(query.To)
	if err != nil {
		return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "invalid to date: "+err.Error())
	}

	if fromDate.After(toDate) {
		return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "from date must be before or equal to to date")
	}

	limit := 50
	if query.Limit > 0 {
		limit = query.Limit
		if limit > 200 {
			limit = 200
		}
	}

	var cursorDate *time.Time
	var cursorPostingID *uuid.UUID
	if query.Cursor != "" {
		cd, cid, err := repository.DecodeCursor(query.Cursor)
		if err != nil {
			return model.LedgerResponse{}, "", fiber.NewError(fiber.StatusBadRequest, "invalid cursor: "+err.Error())
		}
		cursorDate = &cd
		cursorPostingID = &cid
	}

	return s.repo.GetLedger(ctx, acc, fromDate, toDate, cursorDate, cursorPostingID, limit)
}

func (s ReportService) GetTrialBalance(ctx context.Context, query model.TrialBalanceQuery) (model.TrialBalanceResponse, error) {
	if query.AsOf == "" {
		return model.TrialBalanceResponse{}, fiber.NewError(fiber.StatusBadRequest, "asOf date is required")
	}

	asOfDate, err := model.ParseDate(query.AsOf)
	if err != nil {
		return model.TrialBalanceResponse{}, fiber.NewError(fiber.StatusBadRequest, "invalid asOf date: "+err.Error())
	}

	return s.repo.GetTrialBalance(ctx, asOfDate)
}

func (s ReportService) GetCollections(ctx context.Context, query model.CollectionsQuery) (model.CollectionsResponse, error) {
	if query.From == "" || query.To == "" {
		return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "from and to dates are required")
	}

	fromDate, err := model.ParseDate(query.From)
	if err != nil {
		return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "invalid from date: "+err.Error())
	}

	toDate, err := model.ParseDate(query.To)
	if err != nil {
		return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "invalid to date: "+err.Error())
	}

	if fromDate.After(toDate) {
		return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "from date must be before or equal to to date")
	}

	// Range <= 366 days
	if toDate.Sub(fromDate) > 366*24*time.Hour {
		return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "date range exceeds 366 days")
	}

	var allowedAccountIDs map[uuid.UUID]bool
	if query.ParentAccountID != nil {
		accounts, err := s.repo.ListAccounts(ctx)
		if err != nil {
			return model.CollectionsResponse{}, err
		}

		var parentAcc *ent.Account
		childrenMap := make(map[uuid.UUID][]*ent.Account)
		for _, a := range accounts {
			if a.ID == *query.ParentAccountID {
				parentAcc = a
			}
			if a.Edges.Parent != nil {
				childrenMap[a.Edges.Parent.ID] = append(childrenMap[a.Edges.Parent.ID], a)
			}
		}

		if parentAcc == nil {
			return model.CollectionsResponse{}, fiber.NewError(fiber.StatusNotFound, "parent account not found in this parish")
		}

		if parentAcc.Postable {
			return model.CollectionsResponse{}, fiber.NewError(fiber.StatusBadRequest, "parent account must be a header account")
		}

		allowedAccountIDs = make(map[uuid.UUID]bool)
		queue := []uuid.UUID{parentAcc.ID}
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			for _, child := range childrenMap[curr] {
				allowedAccountIDs[child.ID] = true
				queue = append(queue, child.ID)
			}
		}
	}

	return s.repo.GetCollections(ctx, fromDate, toDate, allowedAccountIDs)
}
