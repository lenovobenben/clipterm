package clipterm

import (
	"context"
	"errors"
	"time"

	"github.com/lenovobenben/clipterm/internal/clipboard"
	"github.com/lenovobenben/clipterm/internal/materialize"
	"github.com/lenovobenben/clipterm/internal/paste"
)

type PasteOptions struct {
	CopyPath  bool
	SendPaste bool
}

type DoctorReport struct {
	CacheDir              string
	CanWriteClipboardText bool
	ClipboardImageRead    string
	CanSendPaste          bool
}

type Service struct {
	clipboard   clipboard.Clipboard
	materialize *materialize.Service
	paste       paste.Sender
}

func NewService() *Service {
	return &Service{
		clipboard:   clipboard.NewSystemClipboard(),
		materialize: materialize.NewService(),
		paste:       paste.NewSystemSender(),
	}
}

func (s *Service) Paste(ctx context.Context, options PasteOptions) (string, error) {
	image, err := s.clipboard.ReadImage(ctx)
	if err != nil {
		return "", err
	}

	path, err := s.materialize.Image(ctx, image)
	if err != nil {
		return "", err
	}

	if options.CopyPath || options.SendPaste {
		if err := s.clipboard.WriteText(ctx, path); err != nil {
			return "", err
		}
	}

	if options.SendPaste {
		if err := waitForClipboard(ctx); err != nil {
			return "", err
		}
		if err := s.paste.SendPaste(ctx); err != nil {
			return "", err
		}
	}

	return path, nil
}

func (s *Service) Doctor(ctx context.Context) DoctorReport {
	cacheDir, _ := s.materialize.CacheDir()

	report := DoctorReport{
		CacheDir: cacheDir,
	}

	if checker, ok := s.clipboard.(interface {
		CanWriteText(context.Context) bool
	}); ok {
		report.CanWriteClipboardText = checker.CanWriteText(ctx)
	}

	_, err := s.clipboard.ReadImage(ctx)
	switch {
	case err == nil:
		report.ClipboardImageRead = "ok"
	case errors.Is(err, clipboard.ErrNoImage):
		report.ClipboardImageRead = "no_image"
	case errors.Is(err, clipboard.ErrUnsupported):
		report.ClipboardImageRead = "unavailable"
	default:
		report.ClipboardImageRead = "error"
	}
	report.CanSendPaste = s.paste.CanSendPaste(ctx)

	return report
}

func (s *Service) RequestPastePermission(ctx context.Context) bool {
	return s.paste.RequestPastePermission(ctx)
}

func waitForClipboard(ctx context.Context) error {
	timer := time.NewTimer(80 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
