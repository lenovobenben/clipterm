package materialize

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lenovobenben/clipterm/internal/clipboard"
)

const appDir = "clipterm"

type Service struct{}

type CleanOptions struct {
	Days   int
	DryRun bool
}

type CleanResult struct {
	CacheDir string
	Files    []string
	Bytes    int64
	DryRun   bool
}

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
	if runtime.GOOS == "windows" {
		return filepath.Join(base, appDir, "cache"), nil
	}
	return filepath.Join(base, appDir), nil
}

func (s *Service) Clean(ctx context.Context, options CleanOptions) (CleanResult, error) {
	if options.Days < 0 {
		return CleanResult{}, errors.New("days must be non-negative")
	}

	cacheDir, err := s.CacheDir()
	if err != nil {
		return CleanResult{}, err
	}

	result := CleanResult{
		CacheDir: cacheDir,
		DryRun:   options.DryRun,
	}

	cutoff := time.Now().Add(-time.Duration(options.Days) * 24 * time.Hour)
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return CleanResult{}, err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		if entry.IsDir() || !isManagedImageName(entry.Name()) {
			continue
		}

		path := filepath.Join(cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return result, err
		}

		if info.ModTime().After(cutoff) {
			continue
		}

		result.Files = append(result.Files, path)
		result.Bytes += info.Size()

		if !options.DryRun {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return result, err
			}
		}
	}

	return result, nil
}

func isManagedImageName(name string) bool {
	return strings.HasPrefix(name, "clipterm-") && strings.HasSuffix(name, ".png")
}

func shortID() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b[:])
}
