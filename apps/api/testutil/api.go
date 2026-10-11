package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/paroki/domus/api/ent/membership"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite[T any] struct {
	suite.Suite
	HttpResponse *http.Response
	User         *core.AuthenticatedUser
}

func (s *ApiTestSuite[T]) SetupTest() {
	s.User = &core.AuthenticatedUser{
		Name:  "Test User",
		Email: "test@example.com",
		ID:    core.GenerateID(),
		Roles: []core.UserRole{core.UserRoleAdmin},
	}
}

// GrantRole creates a Membership row for the current test user and
// registers it in the enforcer, mirroring what production does.
func (s *ApiTestSuite[T]) GrantRole(role string, scope authz.Scope) {
	s.T().Helper()
	uid := s.User.ID.String()

	_, err := entClient.Membership.Create().
		SetUserID(uid).
		SetScopeType(membership.ScopeType(scope.Type)).
		SetScopeID(scope.ID).
		SetRole(role).
		Save(context.Background())
	require.NoError(s.T(), err)
	require.NoError(s.T(), state.Authz.AddMembership(uid, role, scope))
}

func (s *ApiTestSuite[T]) jsonBody(v any) io.Reader {
	s.T().Helper()
	b, err := json.Marshal(v)
	require.NoError(s.T(), err)
	return bytes.NewReader(b)
}

func (s *ApiTestSuite[T]) RequestWithToken(path string, method string, requestBody any, token string) {
	s.T().Helper()

	var body io.Reader
	if requestBody != nil {
		body = s.jsonBody(requestBody)
	}

	req := httptest.NewRequest(method, path, body)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	var err error
	s.HttpResponse = nil
	s.HttpResponse, err = api.Test(req)
	require.NoError(s.T(), err)
}

func (s *ApiTestSuite[T]) RequestWithClaims(path string, method string, requestBody any, extraClaims jwt.MapClaims) {
	s.T().Helper()

	claims := jwt.MapClaims{
		"id":    s.User.ID,
		"name":  s.User.Name,
		"email": s.User.Email,
		"roles": s.User.Roles,
	}
	for k, v := range extraClaims {
		claims[k] = v
	}

	token := SignToken(s.T(), claims)
	s.RequestWithToken(path, method, requestBody, token)
}

func (s *ApiTestSuite[T]) Request(path string, method string, requestBody any) {
	s.T().Helper()
	s.RequestWithClaims(path, method, requestBody, nil)
}

func (s *ApiTestSuite[T]) GetResponse() httpx.WebResponse[T] {
	var env httpx.WebResponse[T]
	s.T().Helper()
	defer s.HttpResponse.Body.Close()
	require.NoError(s.T(), json.NewDecoder(s.HttpResponse.Body).Decode(&env))
	return env
}

func (s *ApiTestSuite[T]) PagedResponse() httpx.WebResponse[[]T] {
	var env httpx.WebResponse[[]T]
	s.T().Helper()
	defer s.HttpResponse.Body.Close()
	require.NoError(s.T(), json.NewDecoder(s.HttpResponse.Body).Decode(&env))
	return env
}

func DecodeOther[R any](s interface{ GetHttpResponse() *http.Response }) httpx.WebResponse[R] {
	var env httpx.WebResponse[R]
	resp := s.GetHttpResponse()
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&env)
	return env
}

func (s *ApiTestSuite[T]) GetHttpResponse() *http.Response {
	return s.HttpResponse
}

func (s *ApiTestSuite[T]) AssertStatus(code int) {
	s.T().Helper()
	s.Equal(code, s.HttpResponse.StatusCode)
}

func (s *ApiTestSuite[T]) OK() {
	s.T().Helper()
	s.AssertStatus(fiber.StatusOK)
}

func (s *ApiTestSuite[T]) Created() {
	s.T().Helper()
	s.AssertStatus(fiber.StatusCreated)
}

func (s *ApiTestSuite[T]) NoContent() {
	s.T().Helper()
	s.AssertStatus(fiber.StatusNoContent)
}
