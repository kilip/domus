package test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ActivateSuite struct {
	testutil.ApiTestSuite[model.ActivateFinanceResponse]
}

func (s *ActivateSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	EnsureUserSnapshot(&s.ApiTestSuite)
}

func (s *ActivateSuite) TestActivate_SuccessAndCookie() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID)}
	s.GrantRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelReader), scope)

	req := model.ActivateFinanceRequest{
		ParishID: parish.ID,
		UnitID:   unit.ID,
	}

	s.Request("/fin/activate", fiber.MethodPost, req)
	s.OK()

	res := s.GetResponse()
	assert.Equal(s.T(), parish.ID, res.Data.ActiveParishID)
	assert.Equal(s.T(), unit.ID, res.Data.ActiveUnitID)

	// Verify Set-Cookie header
	rawCookies := s.GetHttpResponse().Header["Set-Cookie"]
	assert.NotEmpty(s.T(), rawCookies, "Set-Cookie header should be present")

	cookies := s.GetHttpResponse().Cookies()
	var finCookie string
	for _, c := range cookies {
		if c.Name == "domus.fin" {
			finCookie = c.Value
			assert.True(s.T(), c.MaxAge > 0, "cookie should be persistent with positive max-age")
		}
	}
	assert.NotEmpty(s.T(), finCookie, fmt.Sprintf("domus.fin cookie should be present in response. Raw Set-Cookie: %v", rawCookies))

	unescapedCookie, err := url.QueryUnescape(finCookie)
	assert.NoError(s.T(), err)
	expectedJSON := fmt.Sprintf(`{"activeParishId":%d,"activeUnitId":%d}`, parish.ID, unit.ID)
	assert.JSONEq(s.T(), expectedJSON, unescapedCookie)
}

func (s *ActivateSuite) TestActivate_ParishNotFound() {
	req := model.ActivateFinanceRequest{
		ParishID: 999999,
		UnitID:   1,
	}

	s.Request("/fin/activate", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *ActivateSuite) TestActivate_UnitBelongsToDifferentParish() {
	parish1 := CreateTestParish(s.T())
	parish2 := CreateTestParish(s.T())
	unit2 := CreateTestUnit(s.T(), parish2.ID, s.User.ID)

	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parish1.ID)}
	s.GrantRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelReader), scope)

	req := model.ActivateFinanceRequest{
		ParishID: parish1.ID,
		UnitID:   unit2.ID,
	}

	s.Request("/fin/activate", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *ActivateSuite) TestActivate_ForbiddenIfNoRole() {
	parish := CreateTestParish(s.T())
	unit := CreateTestUnit(s.T(), parish.ID, s.User.ID)

	req := model.ActivateFinanceRequest{
		ParishID: parish.ID,
		UnitID:   unit.ID,
	}

	s.Request("/fin/activate", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusForbidden)
}

func (s *ActivateSuite) TestRequireFinanceContext_MissingCookie() {
	parish := CreateTestParish(s.T())
	scope := authz.Scope{Type: authz.ScopeParish, ID: int64(parish.ID)}
	s.GrantRole(authz.ModuleRole(authz.ResourceFinance, authz.LevelReader), scope)

	// Call /fin/accounts without setting cookie
	s.Request("/fin/accounts", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func TestActivateSuite(t *testing.T) {
	suite.Run(t, new(ActivateSuite))
}
