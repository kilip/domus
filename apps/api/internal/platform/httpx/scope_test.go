package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParishScope(t *testing.T) {
	app := fiber.New()
	var capturedScope authz.Scope

	scopeFunc := httpx.ParishScope("parishId")
	app.Get("/parishes/:parishId/test", func(c fiber.Ctx) error {
		capturedScope = scopeFunc(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/parishes/12/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, authz.ScopeParish, capturedScope.Type)
	assert.Equal(t, int64(12), capturedScope.ID)
	assert.Equal(t, "parish:12", capturedScope.Domain())

	// Non-numeric or invalid ID defaults to ID 0
	req = httptest.NewRequest(http.MethodGet, "/parishes/abc/test", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, int64(0), capturedScope.ID)
}

func TestRequireIntParam(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			status, body := httpx.MapError(err)
			return c.Status(status).JSON(body)
		},
	})

	app.Get("/parishes/:parishId/items",
		httpx.RequireIntParam("parishId"),
		func(c fiber.Ctx) error {
			id, err := httpx.GetIntParam(c, "parishId")
			if err != nil {
				return err
			}
			return c.JSON(fiber.Map{"id": id})
		},
	)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid positive int",
			path:       "/parishes/12/items",
			wantStatus: http.StatusOK,
		},
		{
			name:       "zero",
			path:       "/parishes/0/items",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ID",
		},
		{
			name:       "negative",
			path:       "/parishes/-1/items",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ID",
		},
		{
			name:       "non numeric string",
			path:       "/parishes/abc/items",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantCode != "" {
				var errBody httpx.ErrorBody
				err = json.NewDecoder(resp.Body).Decode(&errBody)
				require.NoError(t, err)
				assert.Equal(t, tt.wantCode, errBody.Code)
			}
		})
	}
}

func TestGetIntParam_Errors(t *testing.T) {
	app := fiber.New()
	app.Get("/test/:id", func(c fiber.Ctx) error {
		val, err := httpx.GetIntParam(c, "id")
		assert.Equal(t, 0, val)
		assert.ErrorIs(t, err, core.ErrInvalidID)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test/invalid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
