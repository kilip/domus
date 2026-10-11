package test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/ent/account"
	ententry "github.com/paroki/domus/api/ent/entry"
	entparish "github.com/paroki/domus/api/ent/parish"
	entposting "github.com/paroki/domus/api/ent/posting"
	"github.com/paroki/domus/api/ent/transaction"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VoidSuite struct {
	testutil.ApiTestSuite[model.TransactionResponse]
}

func (s *VoidSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *VoidSuite) grantRole(parishID int, role string) {
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parishID)}
	s.GrantRole(role, scope)
}

func (s *VoidSuite) activate(parishID, unitID int) {
	SetFinanceCookie(&s.ApiTestSuite, parishID, unitID)
}

func (s *VoidSuite) createAccountWithUnit(parishID int, unitID int, code string, accType string, postable bool, status string) *ent.Account {
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
		SetUnitID(unitID).
		Save(ctx)
	require.NoError(s.T(), err)
	return acc
}

func (s *VoidSuite) createPostedTransaction(cashID, incomeID uuid.UUID, dateStr string, amount int64) model.TransactionResponse {
	req := model.CreateTransactionRequest{
		Date: dateStr,
		Kind: "income",
		Postings: []model.PostingInput{
			{AccountID: cashID, Side: "debit", Amount: amount},
			{AccountID: incomeID, Side: "credit", Amount: amount},
		},
	}
	s.Request("/fin/transactions", fiber.MethodPost, req)
	s.Created()
	return s.GetResponse().Data
}

func (s *VoidSuite) TestVoid_SuccessDefaultDateAndReversalMirror() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 50000)
	s.Equal("posted", tx.Status)
	s.Len(tx.Entries, 1)

	// Void without body -> default date = original entry date ("2026-10-15"), reason empty
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.OK()

	voidResp := s.GetResponse().Data
	s.Equal("void", voidResp.Status)
	s.Len(voidResp.Entries, 2)

	// Entry 0: original entry
	origEntry := voidResp.Entries[0]
	// Entry 1: reversal entry
	revEntry := voidResp.Entries[1]

	s.Equal(origEntry.Date, revEntry.Date)
	s.Require().NotNil(revEntry.ReversesID)
	s.Equal(origEntry.ID, *revEntry.ReversesID)
	s.Equal(fmt.Sprintf("Reversal of %s", tx.Number), *revEntry.Description)

	// Assert reversal postings mirror original with swapped sides
	s.Len(revEntry.Postings, 2)
	for _, p := range revEntry.Postings {
		if p.AccountID == cash.ID {
			s.Equal("credit", p.Side)
			s.Equal(int64(50000), p.Amount)
		} else if p.AccountID == income.ID {
			s.Equal("debit", p.Side)
			s.Equal(int64(50000), p.Amount)
		} else {
			s.Fail("unexpected account in reversal posting: " + p.AccountID.String())
		}
	}

	// Verify net balance per account is zero
	st := testutil.GetState()
	ctx := context.Background()
	cashPostings, err := st.Ent.Posting.Query().
		Where(
			entposting.HasAccountWith(account.IDEQ(cash.ID)),
			entposting.HasEntryWith(ententry.HasTransactionWith(transaction.HasParishWith(entparish.IDEQ(parish.ID)))),
		).
		All(ctx)
	require.NoError(s.T(), err)
	var cashBalance int64
	for _, cp := range cashPostings {
		if cp.Side == entposting.SideDebit {
			cashBalance += cp.Amount
		} else {
			cashBalance -= cp.Amount
		}
	}
	s.Equal(int64(0), cashBalance)
}

func (s *VoidSuite) TestVoid_SuccessCustomDateAndReason() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 75000)

	customDate := "2026-10-18"
	reason := "Salah catat akun kolekte"
	req := model.VoidTransactionRequest{
		Date:   &customDate,
		Reason: &reason,
	}

	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, req)
	s.OK()

	voidResp := s.GetResponse().Data
	s.Equal("void", voidResp.Status)
	s.Len(voidResp.Entries, 2)

	revEntry := voidResp.Entries[1]
	s.Equal(customDate, revEntry.Date)
	s.Equal(fmt.Sprintf("Reversal of %s: %s", tx.Number, reason), *revEntry.Description)
}

func (s *VoidSuite) TestVoid_TwiceConflict409() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 30000)

	// First void -> 200
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.OK()

	// Second void -> 409
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.Require().Equal(fiber.StatusConflict, s.HttpResponse.StatusCode)
}

func (s *VoidSuite) TestVoid_ConcurrentRequests() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 30000)

	st := testutil.GetState()
	repo := repository.NewTransactionRepository(st.Ent, st.Log)
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	ctx = model.WithFinanceContext(ctx, model.FinanceContext{ParishID: parish.ID, UnitID: unit.ID})

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := repo.Void(ctx, tx.ID, nil, nil)
			results[idx] = err
		}(i)
	}
	wg.Wait()

	// Exactly one succeeds, the other gets 409 conflict
	successCount := 0
	conflictCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		} else {
			if assert.ErrorIs(s.T(), err, model.ErrTxNotPosted) {
				conflictCount++
			}
		}
	}
	s.Equal(1, successCount)
	s.Equal(1, conflictCount)

	// Assert exactly 2 entries exist in the database for this transaction
	count, err := st.Ent.Entry.Query().
		Where(ententry.HasTransactionWith(transaction.HasParishWith(entparish.IDEQ(parish.ID)))).
		Count(ctx)
	require.NoError(s.T(), err)
	s.Equal(2, count)
}

func (s *VoidSuite) TestVoid_WithDeactivatedAccountAllowed() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 40000)

	// Now deactivate income account
	st := testutil.GetState()
	ctx := core.ContextWithUser(context.Background(), core.AuthenticatedUser{ID: s.User.ID})
	_, err := st.Ent.Account.UpdateOneID(income.ID).SetStatus(account.StatusInactive).Save(ctx)
	require.NoError(s.T(), err)

	// Void must still succeed even though an account is inactive
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.OK()
	s.Equal("void", s.GetResponse().Data.Status)
}

func (s *VoidSuite) TestVoid_CrossParish404() {
	parish1 := CreateTestParish(s.T())
	unit1 := CreateTestUnit(s.T(), parish1.ID, s.User.ID)
	parish2 := CreateTestParish(s.T())
	unit2 := CreateTestUnit(s.T(), parish2.ID, s.User.ID)
	s.grantRole(parish1.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.grantRole(parish2.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))

	s.activate(parish1.ID, unit1.ID)
	cash1 := s.createAccountWithUnit(parish1.ID, unit1.ID, "1001", "asset", true, "active")
	income1 := s.createAccountWithUnit(parish1.ID, unit1.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash1.ID, income1.ID, "2026-10-15", 20000)

	// Attempt to void transaction of parish1 in parish2 -> 404
	s.activate(parish2.ID, unit2.ID)
	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.Require().Equal(fiber.StatusNotFound, s.HttpResponse.StatusCode)
}

func (s *VoidSuite) TestVoid_AuthzMatrix() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	s.grantRole(parish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelWriter))
	s.activate(parish.ID, unit.ID)

	cash := s.createAccountWithUnit(parish.ID, unit.ID, "1001", "asset", true, "active")
	income := s.createAccountWithUnit(parish.ID, unit.ID, "4001", "income", true, "active")

	tx := s.createPostedTransaction(cash.ID, income.ID, "2026-10-15", 20000)

	// Reader cannot void -> 403
	readerParish := CreateTestParish(s.T())
	readerUnit := CreateTestUnit(s.T(), readerParish.ID, s.User.ID)
	s.grantRole(readerParish.ID, authz.ModuleRole(authz.ResourceFinance, authz.LevelReader))
	s.activate(readerParish.ID, readerUnit.ID)

	s.Request(fmt.Sprintf("/fin/transactions/%s/void", tx.ID), fiber.MethodPost, nil)
	s.Require().Equal(http.StatusForbidden, s.HttpResponse.StatusCode)
}

func TestVoidSuite(t *testing.T) {
	suite.Run(t, new(VoidSuite))
}
