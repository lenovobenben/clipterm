package materialize

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lenovobenben/clipterm/internal/clipboard"
)

const appDir = "clipterm"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Image(ctx context.Context, image clipboard.Image) (string, error) {
	if len(image.Data) == 0 {
		return "", errors.New("image data is empty")
	}

	cacheDir, err := s.CacheDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}

	extension := image.Extension
	if extension == "" {
		extension = ".png"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	path := filepath.Join(cacheDir, "clipterm-"+time.Now().Format("20060102-150405")+"-"+shortID()+extension)
	if err := os.WriteFile(path, image.Data, 0o644); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return path, nil
	}
}

func (s *Service) CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDir), nil
}

func shortID() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b[:])
}
