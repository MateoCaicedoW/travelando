package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"crypto/rand"
	"encoding/hex"
)

var UploadDir = func() string {
	if v := os.Getenv("UPLOADS_DIR"); v != "" {
		return v
	}
	return "./uploads"
}()

const (
	PhotosDir   = "photos"
	MaxFileSize = 10 << 20
)

var AllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type Service interface {
	SavePhoto(file multipart.File, header *multipart.FileHeader) (string, error)
	DeletePhoto(relPath string) error
}

type service struct{ baseDir string }

func NewService() Service {
	baseDir := filepath.Join(UploadDir, PhotosDir)
	_ = os.MkdirAll(baseDir, 0755)
	return &service{baseDir: baseDir}
}

func (s *service) SavePhoto(file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed")
	}
	contentType := header.Header.Get("Content-Type")
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return "", fmt.Errorf("file type not allowed: %s", contentType)
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	filename := hex.EncodeToString(buf) + ext
	filePath := filepath.Join(s.baseDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(filePath)
		return "", err
	}

	return filepath.ToSlash(filepath.Join(PhotosDir, filename)), nil
}

func (s *service) DeletePhoto(relPath string) error {
	if relPath == "" {
		return nil
	}
	fullPath := filepath.Join(UploadDir, relPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fullPath)
}

func PublicURL(relPath string) string {
	if relPath == "" {
		return ""
	}
	return "/uploads/" + strings.ReplaceAll(relPath, "\\", "/")
}