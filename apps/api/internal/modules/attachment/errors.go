package attachment

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

var (
	ErrFileRequired        = errors.New("FILE_REQUIRED")
	ErrFileHeaderRequired  = errors.New("FILE_HEADER_REQUIRED")
	ErrFileTooLarge        = errors.New("FILE_TOO_LARGE")
	ErrFileOpenFailed      = errors.New("FILE_OPEN_FAILED")
	ErrFileReadFailed      = errors.New("FILE_READ_FAILED")
	ErrUnsupportedFileType = errors.New("UNSUPPORTED_FILE_TYPE")
)

func init() {
	core.RegisterError(ErrFileRequired, fiber.StatusBadRequest, "file field is required")
	core.RegisterError(ErrFileHeaderRequired, fiber.StatusBadRequest, "file is required")
	core.RegisterError(ErrFileTooLarge, fiber.StatusRequestEntityTooLarge, "attachment exceeds maximum size of 10 MiB")
	core.RegisterError(ErrFileOpenFailed, fiber.StatusBadRequest, "failed to open uploaded file")
	core.RegisterError(ErrFileReadFailed, fiber.StatusBadRequest, "failed to read file content")
	core.RegisterError(ErrUnsupportedFileType, fiber.StatusBadRequest, "unsupported file type")
}
