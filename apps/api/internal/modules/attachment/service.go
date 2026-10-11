package attachment

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

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/platform/storage"
)

const (
	maxSize    = 10 * 1024 * 1024 // 10 MiB
	presignTTL = 5 * time.Minute
)

type Service struct {
	repo    Repository
	storage storage.Storage
	bucket  string
	log     *slog.Logger
}

func NewService(repo Repository, storage storage.Storage, bucket string, log *slog.Logger) Service {
	if bucket == "" {
		bucket = "domus-attachments"
	}
	return Service{
		repo:    repo,
		storage: storage,
		bucket:  bucket,
		log:     log,
	}
}

func (s Service) Upload(ctx context.Context, ownerType string, owner Owner, fileHeader *multipart.FileHeader) (*Response, error) {
	if fileHeader == nil {
		return nil, ErrFileHeaderRequired
	}

	// 1. Size check
	if fileHeader.Size > maxSize {
		return nil, ErrFileTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, ErrFileOpenFailed
	}
	defer file.Close()

	// 2. MIME sniffing from first 512 bytes
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, ErrFileReadFailed
	}

	sniffedMIME := http.DetectContentType(buf[:n])
	baseMIME := strings.TrimSpace(strings.Split(sniffedMIME, ";")[0])
	switch baseMIME {
	case "application/pdf", "image/jpeg", "image/png":
		// Allowed
	default:
		return nil, ErrUnsupportedFileType
	}

	// 3. Filename sanitization
	sanitizedFilename := sanitizeFilename(fileHeader.Filename)

	// 4. Object key & ID generation
	attachmentID := core.GenerateID()
	objectKey := fmt.Sprintf("%s/%d/%d/%s/%s-%s", ownerType, owner.ParishID, owner.UnitID, owner.ID.String(), attachmentID.String(), sanitizedFilename)

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
	resp, err := s.repo.Create(ctx, ownerType, owner, CreateInput{
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

func (s Service) List(ctx context.Context, ownerType string, owner Owner) ([]Response, error) {
	return s.repo.ListByOwner(ctx, ownerType, owner)
}

func (s Service) GetDownloadURL(ctx context.Context, ownerType string, owner Owner, attachmentID uuid.UUID) (*DownloadResponse, error) {
	att, err := s.repo.GetByID(ctx, ownerType, owner, attachmentID)
	if err != nil {
		return nil, err
	}

	url, err := s.storage.PresignGet(ctx, att.ObjectKey, presignTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download url: %w", err)
	}

	return &DownloadResponse{
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
