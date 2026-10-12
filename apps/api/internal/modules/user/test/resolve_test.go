package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paroki/domus/api/ent/membership"
	"github.com/paroki/domus/api/ent/user"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/user/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type ResolveSuite struct {
	suite.Suite
	internalKey string
}

func (s *ResolveSuite) SetupSuite() {
	s.internalKey = testutil.GetState().Config.InternalKey
}

func (s *ResolveSuite) requestResolve(body any, keyHeader *string) *http.Response {
	s.T().Helper()
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		s.Require().NoError(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/internal/users/resolve", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if keyHeader != nil {
		req.Header.Set("X-Internal-Key", *keyHeader)
	}

	resp, err := testutil.GetState().FiberApp.Test(req)
	s.Require().NoError(err)
	return resp
}

func (s *ResolveSuite) TestWrongOrMissingKey() {
	body := model.ResolveUserRequest{Email: "test@example.com"}

	// Missing header
	respMissing := s.requestResolve(body, nil)
	s.Equal(http.StatusUnauthorized, respMissing.StatusCode)

	// Wrong key
	wrongKey := "wrong-secret-key-that-does-not-match-at-all"
	respWrong := s.requestResolve(body, &wrongKey)
	s.Equal(http.StatusUnauthorized, respWrong.StatusCode)

	// Empty key
	emptyKey := ""
	respEmpty := s.requestResolve(body, &emptyKey)
	s.Equal(http.StatusUnauthorized, respEmpty.StatusCode)
}

func (s *ResolveSuite) TestUnknownEmail() {
	body := model.ResolveUserRequest{Email: "unknown-person@example.com"}
	resp := s.requestResolve(body, &s.internalKey)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *ResolveSuite) TestSoftDeletedUser() {
	email := "deleted-" + core.GenerateID().String() + "@example.com"
	u, err := testutil.GetState().Ent.User.Create().
		SetID(core.GenerateID()).
		SetName("Deleted User").
		SetEmail(email).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	// Soft delete the user
	err = testutil.GetState().Ent.User.DeleteOneID(u.ID).Exec(context.Background())
	s.Require().NoError(err)

	body := model.ResolveUserRequest{Email: email}
	resp := s.requestResolve(body, &s.internalKey)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *ResolveSuite) TestInvitedAndActiveUsers() {
	// 1. Invited user
	invitedEmail := "invited-" + core.GenerateID().String() + "@example.com"
	invitedUser, err := testutil.GetState().Ent.User.Create().
		SetID(core.GenerateID()).
		SetName("Invited User").
		SetEmail(invitedEmail).
		SetStatus(user.StatusInvited).
		Save(context.Background())
	s.Require().NoError(err)

	// Resolve with uppercase email to test case-insensitivity
	respInvited := s.requestResolve(model.ResolveUserRequest{Email: strings.ToUpper(invitedEmail)}, &s.internalKey)
	s.Equal(http.StatusOK, respInvited.StatusCode)

	var resolvedInvited model.ResolveUserResponse
	defer respInvited.Body.Close()
	s.Require().NoError(json.NewDecoder(respInvited.Body).Decode(&resolvedInvited))
	s.Equal(invitedUser.ID, resolvedInvited.ID)
	s.Equal("invited", resolvedInvited.Status)
	s.Equal("user", resolvedInvited.Role)

	// 2. Active user
	activeEmail := "active-" + core.GenerateID().String() + "@example.com"
	activeUser, err := testutil.GetState().Ent.User.Create().
		SetID(core.GenerateID()).
		SetName("Active User").
		SetEmail(activeEmail).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	respActive := s.requestResolve(model.ResolveUserRequest{Email: activeEmail}, &s.internalKey)
	s.Equal(http.StatusOK, respActive.StatusCode)

	var resolvedActive model.ResolveUserResponse
	defer respActive.Body.Close()
	s.Require().NoError(json.NewDecoder(respActive.Body).Decode(&resolvedActive))
	s.Equal(activeUser.ID, resolvedActive.ID)
	s.Equal("active", resolvedActive.Status)
	s.Equal("user", resolvedActive.Role)
}

func (s *ResolveSuite) TestSuperadminUser() {
	saEmail := "superadmin-" + core.GenerateID().String() + "@example.com"
	saUser, err := testutil.GetState().Ent.User.Create().
		SetID(core.GenerateID()).
		SetName("Super Admin").
		SetEmail(saEmail).
		SetStatus(user.StatusActive).
		Save(context.Background())
	s.Require().NoError(err)

	// Add superadmin membership at system:0
	_, err = testutil.GetState().Ent.Membership.Create().
		SetUserID(saUser.ID.String()).
		SetScopeType(membership.ScopeTypeSystem).
		SetScopeID(0).
		SetRole(authz.RoleSuperadmin).
		Save(context.Background())
	s.Require().NoError(err)

	resp := s.requestResolve(model.ResolveUserRequest{Email: saEmail}, &s.internalKey)
	s.Equal(http.StatusOK, resp.StatusCode)

	var resolved model.ResolveUserResponse
	defer resp.Body.Close()
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&resolved))
	s.Equal(saUser.ID, resolved.ID)
	s.Equal("active", resolved.Status)
	s.Equal("superadmin", resolved.Role)
}

func (s *ResolveSuite) TestNonInternalRouteWithoutJWT() {
	// A request to a non-internal route (e.g., /ping) without JWT must return 401
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := testutil.GetState().FiberApp.Test(req)
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	// Even if X-Internal-Key is provided on a non-internal route without JWT, it must still be 401
	reqWithKey := httptest.NewRequest(http.MethodGet, "/ping", nil)
	reqWithKey.Header.Set("X-Internal-Key", s.internalKey)
	respWithKey, err := testutil.GetState().FiberApp.Test(reqWithKey)
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, respWithKey.StatusCode)
}

func TestResolveSuite(t *testing.T) {
	suite.Run(t, new(ResolveSuite))
}
