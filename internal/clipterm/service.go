package clipterm

import (
	"context"
	"errors"
	"time"

	"github.com/lenovobenben/clipterm/internal/clipboard"
	"github.com/lenovobenben/clipterm/internal/hotkey"
	"github.com/lenovobenben/clipterm/internal/materialize"
	"github.com/lenovobenben/clipterm/internal/paste"
)

type PasteOptions struct {
	CopyPath  bool
	SendPaste bool
}

type SmartPasteResult struct {
	Path        string
	NativePaste bool
}

type DaemonOptions struct {
	DebugHotkeys bool
}

type CleanOptions struct {
	Days   int
	DryRun bool
}

type DoctorReport struct {
	CacheDir              string
	CanWriteClipboardText bool
	ClipboardImageRead    string
	ClipboardFileRead     string
	CanListenHotkey       bool
	CanSendPaste          bool
}

type Service struct {
	clipboard   clipboard.Clipboard
	hotkey      hotkey.Listener
	materialize *materialize.Service
	paste       paste.Sender
}

func NewService() *Service {
	return &Service{
		clipboard:   clipboard.NewSystemClipboard(),
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       paste.NewSystemSender(),
	}
}

func (s *Service) Paste(ctx context.Context, options PasteOptions) (string, error) {
	files, err := s.clipboard.ReadFiles(ctx)
	if err == nil {
		if len(files) == 0 {
			return "", clipboard.ErrNoFile
		}
		if len(files) > 1 {
			return "", clipboard.ErrMultiFile
		}

		path := files[0].Path
		if path == "" {
			return "", clipboard.ErrNoFile
		}
		if err := s.outputPath(ctx, path, options); err != nil {
			return "", err
		}

		return path, nil
	}

	if !errors.Is(err, clipboard.ErrNoFile) {
		return "", err
	}

	image, err := s.clipboard.ReadImage(ctx)
	if err != nil {
		return "", err
	}

	path, err := s.materialize.Image(ctx, image)
	if err != nil {
		return "", err
	}

	if err := s.outputPath(ctx, path, options); err != nil {
		return "", err
	}

	return path, nil
}

func (s *Service) SmartPaste(ctx context.Context) (SmartPasteResult, error) {
	path, err := s.Paste(ctx, PasteOptions{
		CopyPath:  true,
		SendPaste: true,
	})
	if err == nil {
		return SmartPasteResult{Path: path}, nil
	}
	if !errors.Is(err, clipboard.ErrNoImage) {
		return SmartPasteResult{}, err
	}

	if err := s.paste.SendPaste(ctx); err != nil {
		return SmartPasteResult{}, err
	}

	return SmartPasteResult{NativePaste: true}, nil
}

func (s *Service) outputPath(ctx context.Context, path string, options PasteOptions) error {
	if options.CopyPath || options.SendPaste {
		if err := s.clipboard.WriteText(ctx, path); err != nil {
			return err
		}
	}

	if options.SendPaste {
		if err := waitForClipboard(ctx); err != nil {
			return err
		}
		if err := s.paste.SendPaste(ctx); err != nil {
			return err
		}
	}

	return nil
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

	files, err := s.clipboard.ReadFiles(ctx)
	switch {
	case err == nil && len(files) == 1:
		report.ClipboardFileRead = "ok"
	case err == nil && len(files) > 1:
		report.ClipboardFileRead = "multiple"
	case errors.Is(err, clipboard.ErrNoFile):
		report.ClipboardFileRead = "no_file"
	case errors.Is(err, clipboard.ErrUnsupported):
		report.ClipboardFileRead = "unavailable"
	default:
		report.ClipboardFileRead = "error"
	}

	report.CanListenHotkey = s.hotkey.CanListen(ctx)
	report.CanSendPaste = s.paste.CanSendPaste(ctx)

	return report
}

func (s *Service) RequestPastePermission(ctx context.Context) bool {
	return s.paste.RequestPastePermission(ctx)
}

func (s *Service) RequestHotkeyPermission(ctx context.Context) bool {
	return s.hotkey.RequestPermission(ctx)
}

func (s *Service) Clean(ctx context.Context, options CleanOptions) (materialize.CleanResult, error) {
	return s.materialize.Clean(ctx, materialize.CleanOptions{
		Days:   options.Days,
		DryRun: options.DryRun,
	})
}

func (s *Service) RunDaemon(ctx context.Context, options DaemonOptions, handler func(context.Context)) error {
	return s.hotkey.Run(ctx, hotkey.Options{
		Debug: options.DebugHotkeys,
	}, handler)
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
