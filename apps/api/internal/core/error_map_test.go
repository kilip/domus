package core_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleStruct struct {
	Name string `validate:"required"`
}

func TestErrorMap_BasicAndFallback(t *testing.T) {
	em := core.NewErrorMap()

	testErr := errors.New("CUSTOM_TEST_ERROR")
	em.RegisterItem(testErr, fiber.StatusBadRequest, "custom test error message")

	// 1. Direct match
	item := em.Map(testErr)
	assert.Equal(t, fiber.StatusBadRequest, item.Status)
	assert.Equal(t, "CUSTOM_TEST_ERROR", item.Code)
	assert.Equal(t, "custom test error message", item.Message)

	// 2. Unwrapped match
	wrappedErr := fmt.Errorf("wrapped: %w", testErr)
	item = em.Map(wrappedErr)
	assert.Equal(t, fiber.StatusBadRequest, item.Status)
	assert.Equal(t, "CUSTOM_TEST_ERROR", item.Code)
	assert.Equal(t, "custom test error message", item.Message)

	// 3. Validator errors
	v := validator.New()
	valErr := v.Struct(sampleStruct{})
	require.Error(t, valErr)
	item = em.Map(valErr)
	assert.Equal(t, fiber.StatusUnprocessableEntity, item.Status)
	assert.Equal(t, "VALIDATION_FAILED", item.Code)
	assert.Equal(t, "validation failed", item.Message)
	require.Len(t, item.Fields, 1)
	assert.Equal(t, "Name", item.Fields[0].Field)
	assert.Equal(t, "required", item.Fields[0].Rule)

	// 4. Fiber error
	fe := fiber.NewError(fiber.StatusNotFound, "route not found")
	item = em.Map(fe)
	assert.Equal(t, fiber.StatusNotFound, item.Status)
	assert.Equal(t, "HTTP_ERROR", item.Code)
	assert.Equal(t, "route not found", item.Message)

	// 5. Unhandled error fallback
	unhandled := errors.New("something totally unexpected")
	item = em.Map(unhandled)
	assert.Equal(t, fiber.StatusInternalServerError, item.Status)
	assert.Equal(t, "UNKNOWN", item.Code)
	assert.Equal(t, "internal server error", item.Message)

	// 6. Nil error
	item = em.Map(nil)
	assert.Equal(t, fiber.StatusOK, item.Status)
	assert.Equal(t, "OK", item.Code)
	assert.Empty(t, item.Message)
}

func TestGlobalErrorMap_CoreErrorsRegistered(t *testing.T) {
	// Verify that core Sentinel errors are registered in GlobalErrorMap
	item := core.GlobalErrorMap.Map(core.ErrItemNotFound)
	assert.Equal(t, fiber.StatusNotFound, item.Status)
	assert.Equal(t, "NOT_FOUND", item.Code)

	item = core.GlobalErrorMap.Map(core.ErrInvalidID)
	assert.Equal(t, fiber.StatusBadRequest, item.Status)
	assert.Equal(t, "INVALID_ID", item.Code)

	item = core.GlobalErrorMap.Map(core.ErrForbidden)
	assert.Equal(t, fiber.StatusForbidden, item.Status)
	assert.Equal(t, "FORBIDDEN", item.Code)

	item = core.GlobalErrorMap.Map(core.ErrUnauthorized)
	assert.Equal(t, fiber.StatusUnauthorized, item.Status)
	assert.Equal(t, "UNAUTHORIZED", item.Code)

	item = core.GlobalErrorMap.Map(core.ErrInternal)
	assert.Equal(t, fiber.StatusInternalServerError, item.Status)
	assert.Equal(t, "INTERNAL", item.Code)
	assert.Equal(t, "internal server error", item.Message)

	item = core.GlobalErrorMap.Map(core.ErrUnknown)
	assert.Equal(t, fiber.StatusInternalServerError, item.Status)
	assert.Equal(t, "UNKNOWN", item.Code)
	assert.Equal(t, "internal server error", item.Message)
}

func TestErrorMap_ConcurrentAccess(t *testing.T) {
	em := core.NewErrorMap()
	var wg sync.WaitGroup

	// Concurrently register and read
	for i := 0; i < 50; i++ {
		wg.Add(2)
		idx := i
		go func() {
			defer wg.Done()
			err := fmt.Errorf("ERR_%d", idx)
			em.RegisterItem(err, fiber.StatusBadRequest, fmt.Sprintf("msg %d", idx))
		}()
		go func() {
			defer wg.Done()
			err := fmt.Errorf("ERR_%d", idx)
			_ = em.Map(err)
		}()
	}
	wg.Wait()
}
