package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var UploadDir = func() string {
	if v := os.Getenv("UPLOADS_DIR"); v != "" {
		return v
	}
	return "./uploads"
}()

const (
	PhotosDir   = "photos"
	MaxFileSize = 100 << 20
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
	_, ok := AllowedImageTypes[contentType]
	if !ok {
		return "", fmt.Errorf("file type not allowed: %s", contentType)
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	filename := hex.EncodeToString(buf) + ".jpg"
	filePath := filepath.Join(s.baseDir, filename)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 80}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, &out); err != nil {
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