package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/ent/entry"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/transaction"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/attachment"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

var errConflictRetry = errors.New("conflict retry")

type TransactionRepository struct {
	entcli *ent.Client
	log    *slog.Logger
	voidMu *sync.Mutex
}

func NewTransactionRepository(entcli *ent.Client, log *slog.Logger) TransactionRepository {
	return TransactionRepository{
		entcli: entcli,
		log:    log,
		voidMu: &sync.Mutex{},
	}
}

func toTransactionResponse(tx *ent.Transaction) model.TransactionResponse {
	resp := model.TransactionResponse{
		ID:          tx.ID,
		Number:      tx.Number,
		Date:        model.FormatDate(tx.Date),
		Kind:        string(tx.Kind),
		Status:      string(tx.Status),
		Description: tx.Description,
		CreatedAt:   tx.CreatedAt,
		Entries:     make([]model.EntryResponse, 0),
		Attachments: make([]attachment.Response, 0),
	}
	if tx.Edges.Unit != nil {
		resp.UnitID = &tx.Edges.Unit.ID
	}
	if tx.Edges.Creator != nil {
		resp.CreatedBy = toCoreUser(tx.Edges.Creator)
	}

	for _, e := range tx.Edges.Entries {
		entryResp := model.EntryResponse{
			ID:          e.ID,
			Date:        model.FormatDate(e.Date),
			Description: e.Description,
			Postings:    make([]model.PostingResponse, 0),
		}
		if e.Edges.Reverses != nil {
			entryResp.ReversesID = &e.Edges.Reverses.ID
		}
		for _, p := range e.Edges.Postings {
			postResp := model.PostingResponse{
				ID:     p.ID,
				Side:   string(p.Side),
				Amount: p.Amount,
				Memo:   p.Memo,
			}
			if p.Edges.Account != nil {
				postResp.AccountID = p.Edges.Account.ID
				postResp.AccountCode = p.Edges.Account.Code
				postResp.AccountName = p.Edges.Account.Name
			}
			entryResp.Postings = append(entryResp.Postings, postResp)
		}
		resp.Entries = append(resp.Entries, entryResp)
	}

	return resp
}

func toTransactionSummary(tx *ent.Transaction) model.TransactionSummary {
	var totalAmount int64
	for _, e := range tx.Edges.Entries {
		if e.Edges.Reverses == nil {
			for _, p := range e.Edges.Postings {
				if p.Side == posting.SideDebit {
					totalAmount += p.Amount
				}
			}
			break
		}
	}
	if totalAmount == 0 && len(tx.Edges.Entries) > 0 {
		for _, p := range tx.Edges.Entries[0].Edges.Postings {
			if p.Side == posting.SideDebit {
				totalAmount += p.Amount
			}
		}
	}

	summary := model.TransactionSummary{
		ID:          tx.ID,
		Number:      tx.Number,
		Date:        model.FormatDate(tx.Date),
		Kind:        string(tx.Kind),
		Status:      string(tx.Status),
		Description: tx.Description,
		TotalAmount: totalAmount,
	}
	if tx.Edges.Unit != nil {
		summary.UnitID = &tx.Edges.Unit.ID
	}
	return summary
}

func (r TransactionRepository) NextNumber(ctx context.Context, tx *ent.Tx, parishID int, date time.Time) (string, error) {
	prefix := fmt.Sprintf("TRX-%s-", date.UTC().Format("200601"))

	rows, err := tx.Transaction.Query().
		Where(
			transaction.HasParishWith(parish.IDEQ(parishID)),
			transaction.NumberHasPrefix(prefix),
		).
		Select(transaction.FieldNumber).
		All(core.SkipSoftDelete(ctx))
	if err != nil {
		return "", err
	}

	maxSeq := 0
	for _, row := range rows {
		suffix := strings.TrimPrefix(row.Number, prefix)
		if n, err := strconv.Atoi(suffix); err == nil && n > maxSeq {
			maxSeq = n
		}
	}

	nextSeq := maxSeq + 1
	return fmt.Sprintf("%s%04d", prefix, nextSeq), nil
}

func (r TransactionRepository) Create(ctx context.Context, parishID int, req model.CreateTransactionRequest, number string, txDate time.Time) (*model.TransactionResponse, error) {
	var resp *model.TransactionResponse

	maxAttempts := 1
	if number == "" {
		maxAttempts = 3
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := database.WithTx(ctx, r.entcli, func(tx *ent.Tx) error {
			num := number
			if num == "" {
				var err error
				num, err = r.NextNumber(ctx, tx, parishID, txDate)
				if err != nil {
					return err
				}
			}

			txBuilder := tx.Transaction.Create().
				SetParishID(parishID).
				SetNumber(num).
				SetDate(txDate).
				SetKind(transaction.Kind(req.Kind)).
				SetStatus(transaction.StatusPosted)
			if req.Description != nil && *req.Description != "" {
				txBuilder.SetDescription(*req.Description)
			}
			if req.UnitID != nil {
				txBuilder.SetUnitID(*req.UnitID)
			}

			txEnt, err := txBuilder.Save(ctx)
			if err != nil {
				if ent.IsConstraintError(err) {
					if number != "" {
						return fiber.NewError(fiber.StatusConflict, "transaction number already exists in this parish")
					}
					return errConflictRetry
				}
				return err
			}

			entryBuilder := tx.Entry.Create().
				SetTransaction(txEnt).
				SetDate(txDate)
			if req.Description != nil && *req.Description != "" {
				entryBuilder.SetDescription(*req.Description)
			}
			entryEnt, err := entryBuilder.Save(ctx)
			if err != nil {
				return err
			}

			postingBuilders := make([]*ent.PostingCreate, len(req.Postings))
			for i, p := range req.Postings {
				b := tx.Posting.Create().
					SetEntry(entryEnt).
					SetAccountID(p.AccountID).
					SetSide(posting.Side(p.Side)).
					SetAmount(p.Amount)
				if p.Memo != nil && *p.Memo != "" {
					b.SetMemo(*p.Memo)
				}
				postingBuilders[i] = b
			}

			if len(postingBuilders) > 0 {
				_, err = tx.Posting.CreateBulk(postingBuilders...).Save(ctx)
				if err != nil {
					return err
				}
			}

			loaded, err := tx.Transaction.Query().
				Where(transaction.IDEQ(txEnt.ID)).
				WithParish().
				WithUnit().
				WithCreator().
				WithEntries(func(eq *ent.EntryQuery) {
					eq.WithPostings(func(pq *ent.PostingQuery) {
						pq.WithAccount()
					}).WithReverses()
				}).
				Only(ctx)
			if err != nil {
				return err
			}

			res := toTransactionResponse(loaded)
			if err := loadTransactionAttachments(ctx, tx.Client(), &res); err != nil {
				return err
			}
			resp = &res
			return nil
		})

		if err == nil {
			return resp, nil
		}
		if errors.Is(err, errConflictRetry) && number == "" {
			continue
		}
		return nil, err
	}

	return nil, fiber.NewError(fiber.StatusConflict, "could not allocate transaction number after retries")
}

func (r TransactionRepository) GetByID(ctx context.Context, parishID int, id uuid.UUID) (*model.TransactionResponse, error) {
	txEnt, err := r.entcli.Transaction.Query().
		Where(
			transaction.IDEQ(id),
			transaction.HasParishWith(parish.IDEQ(parishID)),
		).
		WithParish().
		WithUnit().
		WithCreator().
		WithEntries(func(eq *ent.EntryQuery) {
			eq.WithPostings(func(pq *ent.PostingQuery) {
				pq.WithAccount()
			}).WithReverses()
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}
	resp := toTransactionResponse(txEnt)
	if err := loadTransactionAttachments(ctx, r.entcli, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r TransactionRepository) Search(ctx context.Context, parishID int, req model.SearchTransactionRequest) (*model.PaginatedResult[model.TransactionSummary], error) {
	q := r.entcli.Transaction.Query().
		Where(transaction.HasParishWith(parish.IDEQ(parishID)))

	if req.From != nil && *req.From != "" {
		fromDate, err := model.ParseDate(*req.From)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid from date: "+err.Error())
		}
		q = q.Where(transaction.DateGTE(fromDate))
	}

	if req.To != nil && *req.To != "" {
		toDate, err := model.ParseDate(*req.To)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "invalid to date: "+err.Error())
		}
		q = q.Where(transaction.DateLTE(toDate))
	}

	if req.Kind != nil && *req.Kind != "" {
		q = q.Where(transaction.KindEQ(transaction.Kind(*req.Kind)))
	}

	if req.Status != nil && *req.Status != "" {
		q = q.Where(transaction.StatusEQ(transaction.Status(*req.Status)))
	}

	if req.UnitID != nil {
		q = q.Where(transaction.HasUnitWith(entunit.IDEQ(*req.UnitID)))
	}

	if req.AccountID != nil {
		q = q.Where(transaction.HasEntriesWith(entry.HasPostingsWith(posting.HasAccountWith(account.IDEQ(*req.AccountID)))))
	}

	if req.Q != nil && *req.Q != "" {
		searchStr := *req.Q
		q = q.Where(transaction.Or(
			transaction.NumberContainsFold(searchStr),
			transaction.DescriptionContainsFold(searchStr),
		))
	}

	if req.Cursor != nil && *req.Cursor != "" {
		cursorDate, cursorID, err := DecodeCursor(*req.Cursor)
		if err != nil {
			return nil, err
		}
		q = q.Where(transaction.Or(
			transaction.DateLT(cursorDate),
			transaction.And(
				transaction.DateEQ(cursorDate),
				transaction.IDLT(cursorID),
			),
		))
	}

	limit := 20
	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
		if limit > 100 {
			limit = 100
		}
	}

	q = q.Order(ent.Desc(transaction.FieldDate), ent.Desc(transaction.FieldID)).
		Limit(limit + 1).
		WithUnit().
		WithEntries(func(eq *ent.EntryQuery) {
			eq.WithPostings().WithReverses()
		})

	txs, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	hasNext := len(txs) > limit
	if hasNext {
		txs = txs[:limit]
	}

	summaries := make([]model.TransactionSummary, len(txs))
	for i, t := range txs {
		summaries[i] = toTransactionSummary(t)
	}

	var nextCursor string
	if hasNext && len(txs) > 0 {
		last := txs[len(txs)-1]
		nextCursor = EncodeCursor(last.Date, last.ID)
	}

	return &model.PaginatedResult[model.TransactionSummary]{
		Data:       summaries,
		NextCursor: nextCursor,
	}, nil
}
