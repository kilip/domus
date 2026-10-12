package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	"github.com/paroki/domus/api/ent/entry"
	entparish "github.com/paroki/domus/api/ent/parish"
	"github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/transaction"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TransactionSuite struct {
	testutil.ApiTestSuite[model.TransactionResponse]
}

func (s *TransactionSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUser(&s.ApiTestSuite)
}

func (s *TransactionSuite) grantRole(parishID int, role string) {
	GrantFinanceRole(&s.ApiTestSuite, parishID, role)
}

func (s *TransactionSuite) activate(parishID, unitID int) {
	SetFinanceCookie(&s.ApiTestSuite, parishID, unitID)
}

func (s *TransactionSuite) createAccountWithUnit(parishID int, unitID int, code string, accType string, postable bool, status string) *ent.Account {
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	acc, err := st.Ent.Account.Create().
		SetCode(code).
		SetName("Account " + code).
		SetType(account.Type(accType)).
		SetNormalBalance(account.NormalBalance(model.DeriveNormalBalance(accType))).
		SetPostable(postable).
		SetStatus(account.Status(status)).
		SetParishID(parishID).
		SetWorkspaceID(unitID).
		Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *TransactionSuite) createAccount(parishID int, code string, accType string, postable bool, status string) *ent.Account {
	unit := CreateTestUnit(s.T(), parishID, s.User.ID)
	return s.createAccountWithUnit(parishID, unit.ID, code, accType, postable, status)
}

func (s *TransactionSuite) TestCreate_Success2LegAndAutoNumbering() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	dateStr := "2026-10-15"
	req1 := model.CreateTransactionRequest{
		Date: dateStr,
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 50000},
			{AccountID: income.ID, Side: "credit", Amount: 50000},
		},
	}

	s.Request("/fin/transactions", fiber.MethodPost, req1)
	s.Created()

	res1 := s.GetResponse()
	s.Equal("TRX-202610-0001", res1.Data.Number)
	s.Equal("posted", res1.Data.Status)
	s.Equal("income", res1.Data.Kind)
	s.Equal(dateStr, res1.Data.Date)
	s.Len(res1.Data.Entries, 1)
	s.Len(res1.Data.Entries[0].Postings, 2)
	s.Equal(s.User.ID, res1.Data.CreatedBy.ID)

	// Second transaction in the same month should have sequence 0002
	req2 := model.CreateTransactionRequest{
		Date: dateStr,
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 75000},
			{AccountID: income.ID, Side: "credit", Amount: 75000},
		},
	}
	s.Request("/fin/transactions", fiber.MethodPost, req2)
	s.Created()

	res2 := s.GetResponse()
	s.Equal("TRX-202610-0002", res2.Data.Number)
}

func (s *TransactionSuite) TestCreate_SuccessSplitTransaction() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	exp1 := s.createAccountWithUnit(parish.ID, unit.ID, "5001", "expense", true, "active")
	exp2 := s.createAccountWithUnit(parish.ID, unit.ID, "5002", "expense", true, "active")

	req := model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "expense",
		Postings: []model.PostingInput{
			{AccountID: exp1.ID, Side: "debit", Amount: 30000},
			{AccountID: exp2.ID, Side: "debit", Amount: 20000},
			{AccountID: cash.ID, Side: "credit", Amount: 50000},
		},
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Equal("posted", res.Data.Status)
	s.Len(res.Data.Entries[0].Postings, 3)
}

func (s *TransactionSuite) TestCreate_ValidationErrors() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	// 1. Unbalanced: debit != credit -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 50000},
			{AccountID: income.ID, Side: "credit", Amount: 40000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)

	// 2. Only debits -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 50000},
			{AccountID: income.ID, Side: "debit", Amount: 50000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)

	// 3. Amount 0 or negative -> 422
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 0},
			{AccountID: income.ID, Side: "credit", Amount: 0},
		},
	})
	s.Require().Equal(fiber.StatusUnprocessableEntity, s.HttpResponse.StatusCode)

	// 4. Amount above max limit (1,000,000,000,000,000) -> 422
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 1000000000000001},
			{AccountID: income.ID, Side: "credit", Amount: 1000000000000001},
		},
	})
	s.Require().Equal(fiber.StatusUnprocessableEntity, s.HttpResponse.StatusCode)

	// 5. Less than 2 postings -> 422
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 50000},
		},
	})
	s.Require().Equal(fiber.StatusUnprocessableEntity, s.HttpResponse.StatusCode)

	// 6. More than 200 postings -> 422
	var manyPostings []model.PostingInput
	for i := 0; i < 201; i++ {
		manyPostings = append(manyPostings, model.PostingInput{
			AccountID: cash.ID,
			Side:      "debit",
			Amount:    100,
		})
	}
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date:     "2026-10-15",
		Kind:     "income",
		Postings: manyPostings,
	})
	s.Require().Equal(fiber.StatusUnprocessableEntity, s.HttpResponse.StatusCode)
}

func (s *TransactionSuite) TestCreate_AccountAndUnitBusinessRules() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	headerAcc := s.createAccountWithUnit(parish.ID, unit.ID, "1000", "asset", false, "active")
	inactiveAcc := s.createAccountWithUnit(parish.ID, unit.ID, "1002", "asset", true, "inactive")

	otherParish := CreateTestParish(s.T())
	otherUnit := CreateTestUnit(s.T(), otherParish.ID, s.User.ID)
	otherAcc := s.createAccountWithUnit(otherParish.ID, otherUnit.ID, "1003", "asset", true, "active")

	// 1. Non-postable (header) account -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "transfer",
		Postings: []model.PostingInput{
			{AccountID: headerAcc.ID, Side: "debit", Amount: 10000},
			{AccountID: cash.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)

	// 2. Inactive account -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "transfer",
		Postings: []model.PostingInput{
			{AccountID: inactiveAcc.ID, Side: "debit", Amount: 10000},
			{AccountID: cash.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)

	// 3. Account from another parish -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "transfer",
		Postings: []model.PostingInput{
			{AccountID: otherAcc.ID, Side: "debit", Amount: 10000},
			{AccountID: cash.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)

	// 4. Non-existent account -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "transfer",
		Postings: []model.PostingInput{
			{AccountID: uuid.New(), Side: "debit", Amount: 10000},
			{AccountID: cash.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)
}

func (s *TransactionSuite) TestCreate_NumberingAndDateValidation() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	// 1. User-supplied number success
	num := "CUSTOM-TRX-001"
	req := model.CreateTransactionRequest{
		Number: &num,
		Date:   "2026-10-15",
		Kind:   "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()
	s.Equal(num, s.GetResponse().Data.Number)

	// 2. Duplicate user-supplied number in same parish -> 409
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Require().Equal(fiber.StatusConflict, s.HttpResponse.StatusCode)

	// 3. Invalid number pattern -> 422
	badNum := "INVALID NUMBER WITH SPACES!"
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Number: &badNum,
		Date:   "2026-10-15",
		Kind:   "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusUnprocessableEntity, s.HttpResponse.StatusCode)

	// 4. Invalid date -> 400
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-13-01",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(fiber.StatusBadRequest, s.HttpResponse.StatusCode)
}

func (s *TransactionSuite) TestCreate_AtomicityRollback() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	st := testutil.GetState()
	repo := repository.NewTransactionRepository(st.Ent, st.Log)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	fakeAccountID := uuid.New()

	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	ctx = model.WithFinanceContext(ctx, model.FinanceContext{ParishID: parish.ID, WorkspaceID: unit.ID})
	txDate, err := model.ParseDate("2026-10-15")
	require.NoError(s.T(), err)

	// Force failure on second posting by pointing to non-existent account
	req := model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 50000},
			{AccountID: fakeAccountID, Side: "credit", Amount: 50000},
		},
	}

	_, err = repo.Create(ctx, req, "", txDate)
	require.Error(s.T(), err)

	// Assert 0 transaction rows remain for this parish
	txCount, err := st.Ent.Transaction.Query().
		Where(transaction.HasParishWith(entparish.IDEQ(parish.ID))).
		Count(ctx)
	require.NoError(s.T(), err)
	s.Equal(0, txCount)

	// Assert 0 entry rows remain for transactions in this parish
	entryCount, err := st.Ent.Entry.Query().
		Where(entry.HasTransactionWith(transaction.HasParishWith(entparish.IDEQ(parish.ID)))).
		Count(ctx)
	require.NoError(s.T(), err)
	s.Equal(0, entryCount)

	// Assert 0 posting rows remain for entries in this parish
	postingCount, err := st.Ent.Posting.Query().
		Where(posting.HasEntryWith(entry.HasTransactionWith(transaction.HasParishWith(entparish.IDEQ(parish.ID))))).
		Count(ctx)
	require.NoError(s.T(), err)
	s.Equal(0, postingCount)
}

func (s *TransactionSuite) TestGetByID_SuccessAndCrossParish404() {
	parish1 := CreateTestParish(s.T())
	unit1 := CreateTestUnit(s.T(), parish1.ID, s.User.ID)
	parish2 := CreateTestParish(s.T())
	unit2 := CreateTestUnit(s.T(), parish2.ID, s.User.ID)
	s.grantRole(parish1.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parish2.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	cash1 := s.createAccountWithUnit(parish1.ID, unit1.ID, "1001", "asset", true, "active")
	income1 := s.createAccountWithUnit(parish1.ID, unit1.ID, "4001", "income", true, "active")

	// Create transaction in parish1
	s.activate(parish1.ID, unit1.ID)
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash1.ID, Side: "debit", Amount: 50000},
			{AccountID: income1.ID, Side: "credit", Amount: 50000},
		},
	})
	s.Created()
	txID := s.GetResponse().Data.ID

	// 1. Get in parish1 -> 200
	s.Request(fmt.Sprintf("/fin/transactions/%s", txID), fiber.MethodGet, nil)
	s.OK()
	s.Equal(txID, s.GetResponse().Data.ID)

	// 2. Get from parish2 -> 404 (tenant isolation)
	s.activate(parish2.ID, unit2.ID)
	s.Request(fmt.Sprintf("/fin/transactions/%s", txID), fiber.MethodGet, nil)
	s.Require().Equal(fiber.StatusNotFound, s.HttpResponse.StatusCode)

	// 3. Non-existent ID -> 404
	s.activate(parish1.ID, unit1.ID)
	s.Request(fmt.Sprintf("/fin/transactions/%s", uuid.New()), fiber.MethodGet, nil)
	s.Require().Equal(fiber.StatusNotFound, s.HttpResponse.StatusCode)
}

func (s *TransactionSuite) TestSearch_FiltersAndCursorPagination() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")
	exp := s.createAccountWithUnit(parish.ID, unit.ID, "5001", "expense", true, "active")

	// Create 5 transactions on the SAME date to verify stable cursor pagination across identical dates
	sameDate := "2026-10-20"
	for i := 1; i <= 5; i++ {
		desc := fmt.Sprintf("Kolekte Minggu %d", i)
		req := model.CreateTransactionRequest{
			Date:        sameDate,
			Kind:        "income",
			Description: &desc,
			Postings: []model.PostingInput{
				{AccountID: cash.ID, Side: "debit", Amount: int64(10000 * i)},
				{AccountID: income.ID, Side: "credit", Amount: int64(10000 * i)},
			},
		}
		s.Request("/fin/transactions", fiber.MethodPost, req)
		s.Created()
	}

	// Create 1 expense transaction on another date
	diffDate := "2026-10-22"
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: diffDate,
		Kind: "expense",
		Postings: []model.PostingInput{
			{AccountID: exp.ID, Side: "debit", Amount: 25000},
			{AccountID: cash.ID, Side: "credit", Amount: 25000},
		},
	})
	s.Created()

	// 1. Pagination across identical dates (limit = 2)
	limit := 2
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		From:  &sameDate,
		To:    &sameDate,
		Limit: &limit,
	})
	s.OK()
	page1 := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(page1.Data, 2)
	s.NotEmpty(page1.Meta.Cursor)
	s.Equal(int64(50000), page1.Data[0].TotalAmount) // Most recent first (DESC)

	// Page 2
	cursor1 := page1.Meta.Cursor
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		From:   &sameDate,
		To:     &sameDate,
		Cursor: &cursor1,
		Limit:  &limit,
	})
	s.OK()
	page2 := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(page2.Data, 2)
	s.NotEmpty(page2.Meta.Cursor)
	s.NotEqual(page1.Data[1].ID, page2.Data[0].ID)

	// Page 3
	cursor2 := page2.Meta.Cursor
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		From:   &sameDate,
		To:     &sameDate,
		Cursor: &cursor2,
		Limit:  &limit,
	})
	s.OK()
	page3 := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(page3.Data, 1)
	s.Empty(page3.Meta.Cursor) // No more pages

	// 2. Filter by Kind
	kind := "expense"
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		Kind: &kind,
	})
	s.OK()
	expSearch := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(expSearch.Data, 1)
	s.Equal("expense", expSearch.Data[0].Kind)
	s.Equal(int64(25000), expSearch.Data[0].TotalAmount)

	// 3. Filter by Account
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		AccountID: &exp.ID,
	})
	s.OK()
	accSearch := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(accSearch.Data, 1)

	// 4. Filter by text query Q
	query := "Minggu 3"
	s.Request("/fin/transactions/search", fiber.MethodPost, model.SearchTransactionRequest{
		Q: &query,
	})
	s.OK()
	qSearch := testutil.DecodeOther[[]model.TransactionSummary](s)
	s.Len(qSearch.Data, 1)
	s.Contains(*qSearch.Data[0].Description, "Minggu 3")
}

func (s *TransactionSuite) TestAuthzMatrix() {
	parish1 := CreateTestParish(s.T())
	unit1 := CreateTestUnit(s.T(), parish1.ID, s.User.ID)
	parish2 := CreateTestParish(s.T())
	unit2 := CreateTestUnit(s.T(), parish2.ID, s.User.ID)
	_ = unit2

	cash := s.createAccountWithUnit(parish1.ID, unit1.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish1.ID, unit1.ID, "4001", "income", true, "active")

	// 1. No membership -> 403
	s.activate(parish1.ID, unit1.ID)
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(http.StatusForbidden, s.HttpResponse.StatusCode)

	// 2. Reader can search and GET, cannot POST -> 403
	s.grantRole(parish1.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelReader))
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(http.StatusForbidden, s.HttpResponse.StatusCode)

	s.Request("/fin/transactions/search", fiber.MethodPost, nil)
	s.OK()

	// 3. Writer in parish 2 cannot access parish 1
	s.grantRole(parish2.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish1.ID, unit1.ID)
	s.Request("/fin/transactions", fiber.MethodPost, model.CreateTransactionRequest{
		Date: "2026-10-15",
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cash.ID, Side: "debit", Amount: 10000},
			{AccountID: income.ID, Side: "credit", Amount: 10000},
		},
	})
	s.Require().Equal(http.StatusForbidden, s.HttpResponse.StatusCode)
}

func TestTransactionSuite(t *testing.T) {
	suite.Run(t, new(TransactionSuite))
}
