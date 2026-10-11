package repository

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/ent/entry"
	"github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/predicate"
	"github.com/paroki/domus/api/ent/transaction"
	entunit "github.com/paroki/domus/api/ent/unit"
	"github.com/paroki/domus/api/internal/modules/finance/model"
)

type ReportRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewReportRepository(entcli *ent.Client, log *slog.Logger) ReportRepository {
	return ReportRepository{
		entcli: entcli,
		log:    log,
	}
}

func (r ReportRepository) GetAccount(ctx context.Context, parishID int, accountID uuid.UUID) (*ent.Account, error) {
	return r.entcli.Account.Query().
		Where(
			account.IDEQ(accountID),
			account.HasParishWith(parish.IDEQ(parishID)),
		).
		Only(ctx)
}

func (r ReportRepository) ListParishAccounts(ctx context.Context, parishID int) ([]*ent.Account, error) {
	return r.entcli.Account.Query().
		Where(account.HasParishWith(parish.IDEQ(parishID))).
		WithParent().
		All(ctx)
}

func (r ReportRepository) CheckUnitExists(ctx context.Context, parishID int, unitID int) (bool, error) {
	return r.entcli.Unit.Query().
		Where(
			entunit.IDEQ(unitID),
			entunit.HasParishWith(parish.IDEQ(parishID)),
		).
		Exist(ctx)
}

func (r ReportRepository) forEachPosting(ctx context.Context, parishID int, pred predicate.Posting, fn func(*ent.Posting) error) error {
	var lastID *uuid.UUID
	for {
		q := r.entcli.Posting.Query().
			Where(
				posting.HasAccountWith(account.HasParishWith(parish.IDEQ(parishID))),
				posting.HasEntryWith(entry.HasTransactionWith(transaction.HasParishWith(parish.IDEQ(parishID)))),
			)
		if pred != nil {
			q = q.Where(pred)
		}
		if lastID != nil {
			q = q.Where(posting.IDGT(*lastID))
		}
		batch, err := q.
			WithAccount().
			WithEntry(func(eq *ent.EntryQuery) {
				eq.WithTransaction(func(tq *ent.TransactionQuery) {
					tq.WithUnit()
				})
			}).
			Order(ent.Asc(posting.FieldID)).
			Limit(1000).
			All(ctx)
		if err != nil {
			return err
		}
		for _, p := range batch {
			if err := fn(p); err != nil {
				return err
			}
		}
		if len(batch) < 1000 {
			break
		}
		lastID = &batch[len(batch)-1].ID
	}
	return nil
}

func (r ReportRepository) forEachPostingForAccount(ctx context.Context, parishID int, accountID uuid.UUID, pred predicate.Posting, fn func(*ent.Posting) error) error {
	var lastID *uuid.UUID
	for {
		q := r.entcli.Posting.Query().
			Where(
				posting.HasAccountWith(
					account.IDEQ(accountID),
					account.HasParishWith(parish.IDEQ(parishID)),
				),
				posting.HasEntryWith(entry.HasTransactionWith(transaction.HasParishWith(parish.IDEQ(parishID)))),
			)
		if pred != nil {
			q = q.Where(pred)
		}
		if lastID != nil {
			q = q.Where(posting.IDGT(*lastID))
		}
		batch, err := q.
			WithAccount().
			WithEntry(func(eq *ent.EntryQuery) {
				eq.WithTransaction()
			}).
			Order(ent.Asc(posting.FieldID)).
			Limit(1000).
			All(ctx)
		if err != nil {
			return err
		}
		for _, p := range batch {
			if err := fn(p); err != nil {
				return err
			}
		}
		if len(batch) < 1000 {
			break
		}
		lastID = &batch[len(batch)-1].ID
	}
	return nil
}

func (r ReportRepository) GetLedger(
	ctx context.Context,
	parishID int,
	acc *ent.Account,
	from, to time.Time,
	cursorDate *time.Time,
	cursorPostingID *uuid.UUID,
	limit int,
) (model.LedgerResponse, string, error) {
	isDebitNormal := acc.NormalBalance == account.NormalBalanceDebit
	var openingBalance int64

	var opPred predicate.Posting
	if cursorDate == nil {
		opPred = posting.HasEntryWith(entry.DateLT(from))
	} else {
		opPred = posting.Or(
			posting.HasEntryWith(entry.DateLT(*cursorDate)),
			posting.And(
				posting.HasEntryWith(entry.DateEQ(*cursorDate)),
				posting.IDLTE(*cursorPostingID),
			),
		)
	}

	err := r.forEachPostingForAccount(ctx, parishID, acc.ID, opPred, func(p *ent.Posting) error {
		if isDebitNormal {
			if p.Side == posting.SideDebit {
				openingBalance += p.Amount
			} else {
				openingBalance -= p.Amount
			}
		} else {
			if p.Side == posting.SideCredit {
				openingBalance += p.Amount
			} else {
				openingBalance -= p.Amount
			}
		}
		return nil
	})
	if err != nil {
		return model.LedgerResponse{}, "", err
	}

	q := r.entcli.Posting.Query().
		Where(
			posting.HasAccountWith(
				account.IDEQ(acc.ID),
				account.HasParishWith(parish.IDEQ(parishID)),
			),
			posting.HasEntryWith(
				entry.HasTransactionWith(transaction.HasParishWith(parish.IDEQ(parishID))),
				entry.DateGTE(from),
				entry.DateLTE(to),
			),
		)

	if cursorDate != nil {
		q = q.Where(
			posting.Or(
				posting.HasEntryWith(entry.DateGT(*cursorDate)),
				posting.And(
					posting.HasEntryWith(entry.DateEQ(*cursorDate)),
					posting.IDGT(*cursorPostingID),
				),
			),
		)
	}

	pagePostings, err := q.
		WithEntry(func(eq *ent.EntryQuery) {
			eq.WithTransaction()
		}).
		Order(
			posting.ByEntryField(entry.FieldDate),
			ent.Asc(posting.FieldID),
		).
		Limit(limit + 1).
		All(ctx)
	if err != nil {
		return model.LedgerResponse{}, "", err
	}

	hasNext := len(pagePostings) > limit
	if hasNext {
		pagePostings = pagePostings[:limit]
	}

	runningBalance := openingBalance
	rows := make([]model.LedgerRow, 0, len(pagePostings))
	for _, p := range pagePostings {
		var debit, credit int64
		if p.Side == posting.SideDebit {
			debit = p.Amount
		} else {
			credit = p.Amount
		}

		if isDebitNormal {
			runningBalance += debit - credit
		} else {
			runningBalance += credit - debit
		}

		var desc *string
		if p.Memo != nil && *p.Memo != "" {
			desc = p.Memo
		} else if p.Edges.Entry != nil && p.Edges.Entry.Description != nil {
			desc = p.Edges.Entry.Description
		} else if p.Edges.Entry != nil && p.Edges.Entry.Edges.Transaction != nil {
			desc = p.Edges.Entry.Edges.Transaction.Description
		}

		var txID uuid.UUID
		var txNumber string
		if p.Edges.Entry != nil && p.Edges.Entry.Edges.Transaction != nil {
			txID = p.Edges.Entry.Edges.Transaction.ID
			txNumber = p.Edges.Entry.Edges.Transaction.Number
		}

		rows = append(rows, model.LedgerRow{
			Date:              model.FormatDate(p.Edges.Entry.Date),
			EntryID:           p.Edges.Entry.ID,
			PostingID:         p.ID,
			TransactionID:     txID,
			TransactionNumber: txNumber,
			Description:       desc,
			Debit:             debit,
			Credit:            credit,
			Balance:           runningBalance,
		})
	}

	closingBalance := openingBalance
	if len(rows) > 0 {
		closingBalance = runningBalance
	}

	var nextCursor string
	if hasNext && len(rows) > 0 {
		lastRow := rows[len(rows)-1]
		lastDate, _ := model.ParseDate(lastRow.Date)
		nextCursor = EncodeCursor(lastDate, lastRow.PostingID)
	}

	resp := model.LedgerResponse{
		Account: model.LedgerAccount{
			ID:            acc.ID,
			Code:          acc.Code,
			Name:          acc.Name,
			Type:          string(acc.Type),
			NormalBalance: string(acc.NormalBalance),
		},
		OpeningBalance: openingBalance,
		Rows:           rows,
		ClosingBalance: closingBalance,
	}

	return resp, nextCursor, nil
}

func (r ReportRepository) GetTrialBalance(ctx context.Context, parishID int, asOf time.Time) (model.TrialBalanceResponse, error) {
	type accStats struct {
		acc    *ent.Account
		debit  int64
		credit int64
	}
	statsMap := make(map[uuid.UUID]*accStats)

	pred := posting.HasEntryWith(entry.DateLTE(asOf))
	err := r.forEachPosting(ctx, parishID, pred, func(p *ent.Posting) error {
		acc := p.Edges.Account
		if acc == nil || !acc.Postable {
			return nil
		}
		s, exists := statsMap[acc.ID]
		if !exists {
			s = &accStats{acc: acc}
			statsMap[acc.ID] = s
		}
		if p.Side == posting.SideDebit {
			s.debit += p.Amount
		} else {
			s.credit += p.Amount
		}
		return nil
	})
	if err != nil {
		return model.TrialBalanceResponse{}, err
	}

	accountsList := make([]*accStats, 0, len(statsMap))
	for _, s := range statsMap {
		accountsList = append(accountsList, s)
	}
	sort.Slice(accountsList, func(i, j int) bool {
		return accountsList[i].acc.Code < accountsList[j].acc.Code
	})

	rows := make([]model.TrialBalanceRow, 0, len(accountsList))
	var totals model.TrialBalanceTotals
	for _, s := range accountsList {
		var balance int64
		normalBalance := model.DeriveNormalBalance(string(s.acc.Type))
		if normalBalance == "debit" {
			balance = s.debit - s.credit
		} else {
			balance = s.credit - s.debit
		}
		rows = append(rows, model.TrialBalanceRow{
			AccountID: s.acc.ID,
			Code:      s.acc.Code,
			Name:      s.acc.Name,
			Type:      string(s.acc.Type),
			Debit:     s.debit,
			Credit:    s.credit,
			Balance:   balance,
		})
		totals.Debit += s.debit
		totals.Credit += s.credit
	}

	return model.TrialBalanceResponse{
		AsOf:   model.FormatDate(asOf),
		Rows:   rows,
		Totals: totals,
	}, nil
}

func (r ReportRepository) GetCollections(
	ctx context.Context,
	parishID int,
	from, to time.Time,
	allowedAccountIDs map[uuid.UUID]bool,
	unitID *int,
) (model.CollectionsResponse, error) {
	type incStats struct {
		acc   *ent.Account
		total int64
	}
	statsMap := make(map[uuid.UUID]*incStats)

	pred := posting.And(
		posting.HasEntryWith(
			entry.DateGTE(from),
			entry.DateLTE(to),
		),
		posting.HasAccountWith(
			account.TypeEQ(account.TypeIncome),
		),
	)
	if unitID != nil {
		pred = posting.And(
			pred,
			posting.HasEntryWith(
				entry.HasTransactionWith(
					transaction.HasUnitWith(entunit.IDEQ(*unitID)),
				),
			),
		)
	}

	err := r.forEachPosting(ctx, parishID, pred, func(p *ent.Posting) error {
		acc := p.Edges.Account
		if acc == nil || acc.Type != account.TypeIncome {
			return nil
		}
		if allowedAccountIDs != nil && !allowedAccountIDs[acc.ID] {
			return nil
		}

		s, exists := statsMap[acc.ID]
		if !exists {
			s = &incStats{acc: acc}
			statsMap[acc.ID] = s
		}
		if p.Side == posting.SideCredit {
			s.total += p.Amount
		} else {
			s.total -= p.Amount
		}
		return nil
	})
	if err != nil {
		return model.CollectionsResponse{}, err
	}

	accountsList := make([]*incStats, 0)
	for _, s := range statsMap {
		if s.total != 0 {
			accountsList = append(accountsList, s)
		}
	}
	sort.Slice(accountsList, func(i, j int) bool {
		return accountsList[i].acc.Code < accountsList[j].acc.Code
	})

	rows := make([]model.CollectionsRow, 0, len(accountsList))
	var grandTotal int64
	for _, s := range accountsList {
		rows = append(rows, model.CollectionsRow{
			AccountID: s.acc.ID,
			Code:      s.acc.Code,
			Name:      s.acc.Name,
			Total:     s.total,
		})
		grandTotal += s.total
	}

	return model.CollectionsResponse{
		From:       model.FormatDate(from),
		To:         model.FormatDate(to),
		Rows:       rows,
		GrandTotal: grandTotal,
	}, nil
}
