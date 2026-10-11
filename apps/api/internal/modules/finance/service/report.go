package service

import (
	"context"
	"log/slog"
	"time"

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
		return model.LedgerResponse{}, "", model.ErrAccountIdRequired
	}

	acc, err := s.repo.GetAccount(ctx, query.AccountID)
	if err != nil {
		if ent.IsNotFound(err) {
			return model.LedgerResponse{}, "", model.ErrAccountNotFoundInParish
		}
		return model.LedgerResponse{}, "", err
	}

	if query.From == "" || query.To == "" {
		return model.LedgerResponse{}, "", model.ErrDatesRequired
	}

	fromDate, err := model.ParseDate(query.From)
	if err != nil {
		return model.LedgerResponse{}, "", model.ErrInvalidFromDate
	}

	toDate, err := model.ParseDate(query.To)
	if err != nil {
		return model.LedgerResponse{}, "", model.ErrInvalidToDate
	}

	if fromDate.After(toDate) {
		return model.LedgerResponse{}, "", model.ErrFromDateAfterToDate
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
			return model.LedgerResponse{}, "", model.ErrInvalidCursor
		}
		cursorDate = &cd
		cursorPostingID = &cid
	}

	return s.repo.GetLedger(ctx, acc, fromDate, toDate, cursorDate, cursorPostingID, limit)
}

func (s ReportService) GetTrialBalance(ctx context.Context, query model.TrialBalanceQuery) (model.TrialBalanceResponse, error) {
	if query.AsOf == "" {
		return model.TrialBalanceResponse{}, model.ErrAsOfDateRequired
	}

	asOfDate, err := model.ParseDate(query.AsOf)
	if err != nil {
		return model.TrialBalanceResponse{}, model.ErrInvalidAsOfDate
	}

	return s.repo.GetTrialBalance(ctx, asOfDate)
}

func (s ReportService) GetCollections(ctx context.Context, query model.CollectionsQuery) (model.CollectionsResponse, error) {
	if query.From == "" || query.To == "" {
		return model.CollectionsResponse{}, model.ErrDatesRequired
	}

	fromDate, err := model.ParseDate(query.From)
	if err != nil {
		return model.CollectionsResponse{}, model.ErrInvalidFromDate
	}

	toDate, err := model.ParseDate(query.To)
	if err != nil {
		return model.CollectionsResponse{}, model.ErrInvalidToDate
	}

	if fromDate.After(toDate) {
		return model.CollectionsResponse{}, model.ErrFromDateAfterToDate
	}

	// Range <= 366 days
	if toDate.Sub(fromDate) > 366*24*time.Hour {
		return model.CollectionsResponse{}, model.ErrDateRangeExceedsLimit
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
			return model.CollectionsResponse{}, model.ErrParentAccountNotFound
		}

		if parentAcc.Postable {
			return model.CollectionsResponse{}, model.ErrParentMustBeHeader
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
