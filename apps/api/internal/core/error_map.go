package core

import (
	"errors"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type ErrorMap struct {
	mu    sync.RWMutex
	items map[string]ErrorItem
}

func NewErrorMap() *ErrorMap {
	return &ErrorMap{
		items: make(map[string]ErrorItem),
	}
}

var GlobalErrorMap = NewErrorMap()

func RegisterErrors(items map[string]ErrorItem) {
	GlobalErrorMap.Register(items)
}

func RegisterError(err error, status int, message string) {
	GlobalErrorMap.RegisterItem(err, status, message)
}

func (m *ErrorMap) Register(items map[string]ErrorItem) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range items {
		m.items[k] = v
	}
}

func (m *ErrorMap) RegisterItem(err error, status int, message string) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[err.Error()] = ErrorItem{
		Status:  status,
		Code:    err.Error(),
		Message: message,
	}
}

func (m *ErrorMap) Get(key string) (ErrorItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.items[key]
	return item, ok
}

func (m *ErrorMap) Map(err error) ErrorItem {
	if err == nil {
		return ErrorItem{
			Status:  fiber.StatusOK,
			Code:    "OK",
			Message: "",
		}
	}

	m.mu.RLock()
	if item, ok := m.items[err.Error()]; ok {
		m.mu.RUnlock()
		return item
	}

	curr := err
	for {
		unwrapped := errors.Unwrap(curr)
		if unwrapped == nil {
			break
		}
		if item, ok := m.items[unwrapped.Error()]; ok {
			m.mu.RUnlock()
			return item
		}
		curr = unwrapped
	}
	m.mu.RUnlock()

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]FieldError, 0, len(ve))
		for _, f := range ve {
			fields = append(fields, FieldError{Field: f.Field(), Rule: f.Tag()})
		}
		return ErrorItem{
			Status:  fiber.StatusUnprocessableEntity,
			Code:    "VALIDATION_FAILED",
			Message: "validation failed",
			Fields:  fields,
		}
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return ErrorItem{
			Status:  fe.Code,
			Code:    "HTTP_ERROR",
			Message: fe.Message,
		}
	}

	m.mu.RLock()
	if item, ok := m.items[ErrUnknown.Error()]; ok {
		m.mu.RUnlock()
		return item
	}
	m.mu.RUnlock()

	return ErrorItem{
		Status:  fiber.StatusInternalServerError,
		Code:    "UNKNOWN",
		Message: "internal server error",
	}
}

func init() {
	RegisterError(ErrItemNotFound, fiber.StatusNotFound, "resource not found")
	RegisterError(ErrInvalidID, fiber.StatusBadRequest, "invalid id parameter")
	RegisterError(ErrForbidden, fiber.StatusForbidden, "forbidden")
	RegisterError(ErrUnauthorized, fiber.StatusUnauthorized, "unauthorized")
	RegisterError(ErrInternal, fiber.StatusInternalServerError, "internal server error")
	RegisterError(ErrUnknown, fiber.StatusInternalServerError, "internal server error")
}
