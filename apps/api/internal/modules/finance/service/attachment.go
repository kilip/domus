package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/finance/model"
	"github.com/paroki/domus/api/internal/modules/finance/repository"
	"github.com/paroki/domus/api/internal/platform/storage"
)

const (
	maxAttachmentSize = 10 * 1024 * 1024 // 10 MiB
	presignTTL        = 5 * time.Minute
)

type AttachmentService struct {
	repo    repository.AttachmentRepository
	storage storage.Storage
	bucket  string
	log     *slog.Logger
}

func NewAttachmentService(repo repository.AttachmentRepository, storage storage.Storage, bucket string, log *slog.Logger) AttachmentService {
	if bucket == "" {
		bucket = "domus-finance"
	}
	return AttachmentService{
		repo:    repo,
		storage: storage,
		bucket:  bucket,
		log:     log,
	}
}

func (s AttachmentService) Upload(ctx context.Context, parishID int, transactionID uuid.UUID, fileHeader *multipart.FileHeader) (*model.AttachmentResponse, error) {
	if fileHeader == nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	// 1. Size check
	if fileHeader.Size > maxAttachmentSize {
		return nil, fiber.NewError(fiber.StatusRequestEntityTooLarge, "attachment exceeds maximum size of 10 MiB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "failed to open uploaded file")
	}
	defer file.Close()

	// 2. MIME sniffing from first 512 bytes
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fiber.NewError(fiber.StatusBadRequest, "failed to read file content")
	}

	sniffedMIME := http.DetectContentType(buf[:n])
	baseMIME := strings.TrimSpace(strings.Split(sniffedMIME, ";")[0])
	switch baseMIME {
	case "application/pdf", "image/jpeg", "image/png":
		// Allowed
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported file type: %s", baseMIME))
	}

	// 3. Filename sanitization
	sanitizedFilename := sanitizeFilename(fileHeader.Filename)

	// 4. Object key & ID generation
	attachmentID := core.GenerateID()
	objectKey := fmt.Sprintf("finance/%d/%s/%s-%s", parishID, transactionID.String(), attachmentID.String(), sanitizedFilename)

	// 5. Stream to storage and compute SHA-256
	combinedReader := io.MultiReader(bytes.NewReader(buf[:n]), file)
	hasher := sha256.New()
	teeReader := io.TeeReader(combinedReader, hasher)

	if err := s.storage.Put(ctx, objectKey, teeReader, fileHeader.Size, baseMIME); err != nil {
		if s.log != nil {
			s.log.Error("failed to put attachment to storage", "error", err, "key", objectKey)
		}
		return nil, fmt.Errorf("failed to save file to storage: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	// 6. DB insert
	resp, err := s.repo.Create(ctx, parishID, transactionID, repository.CreateAttachmentInput{
		ID:        attachmentID,
		Bucket:    s.bucket,
		ObjectKey: objectKey,
		Filename:  sanitizedFilename,
		MimeType:  baseMIME,
		Size:      fileHeader.Size,
		Checksum:  checksum,
	})
	if err != nil {
		// Rollback storage upload on DB error
		_ = s.storage.Delete(ctx, objectKey)
		return nil, err
	}

	return resp, nil
}

func (s AttachmentService) List(ctx context.Context, parishID int, transactionID uuid.UUID) ([]model.AttachmentResponse, error) {
	return s.repo.ListByTransaction(ctx, parishID, transactionID)
}

func (s AttachmentService) GetDownloadURL(ctx context.Context, parishID int, transactionID uuid.UUID, attachmentID uuid.UUID) (*model.DownloadResponse, error) {
	att, err := s.repo.GetByID(ctx, parishID, transactionID, attachmentID)
	if err != nil {
		return nil, err
	}

	url, err := s.storage.PresignGet(ctx, att.ObjectKey, presignTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download url: %w", err)
	}

	return &model.DownloadResponse{
		URL:       url,
		ExpiresAt: time.Now().Add(presignTTL),
	}, nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.ReplaceAll(base, "\\", "")
	base = strings.ReplaceAll(base, "/", "")

	var b strings.Builder
	for _, r := range base {
		if r < 32 || r == 127 {
			continue
		}
		b.WriteRune(r)
	}
	cleaned := strings.TrimSpace(b.String())
	cleaned = strings.TrimLeft(cleaned, ".")
	if cleaned == "" {
		cleaned = "file"
	}
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}
	return cleaned
}
