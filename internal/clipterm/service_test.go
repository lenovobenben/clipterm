package clipterm

import (
	"context"
	"errors"
	"testing"

	"github.com/lenovobenben/clipterm/internal/clipboard"
	"github.com/lenovobenben/clipterm/internal/hotkey"
	"github.com/lenovobenben/clipterm/internal/materialize"
	"github.com/lenovobenben/clipterm/internal/paste"
)

type fakeClipboard struct {
	files       []clipboard.FileRef
	filesErr    error
	image       clipboard.Image
	imageErr    error
	writtenText string
	writeCount  int
}

func (f *fakeClipboard) ReadImage(context.Context) (clipboard.Image, error) {
	if f.imageErr != nil {
		return clipboard.Image{}, f.imageErr
	}
	return f.image, nil
}

func (f *fakeClipboard) ReadFiles(context.Context) ([]clipboard.FileRef, error) {
	if f.filesErr != nil {
		return nil, f.filesErr
	}
	return f.files, nil
}

func (f *fakeClipboard) WriteText(_ context.Context, text string) error {
	f.writtenText = text
	f.writeCount++
	return nil
}

type fakePasteSender struct {
	sendCount int
}

func (f *fakePasteSender) SendPaste(context.Context) error {
	f.sendCount++
	return nil
}

func (f *fakePasteSender) CanSendPaste(context.Context) bool {
	return true
}

func (f *fakePasteSender) RequestPastePermission(context.Context) bool {
	return true
}

func TestPastePrefersSingleFileOverImage(t *testing.T) {
	cb := &fakeClipboard{
		files: []clipboard.FileRef{{Path: "/tmp/from-file.png"}},
		image: clipboard.Image{
			Data:      []byte("image-data"),
			Format:    "png",
			Extension: ".png",
		},
	}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       paste.NewSystemSender(),
	}

	path, err := svc.Paste(context.Background(), PasteOptions{CopyPath: true})
	if err != nil {
		t.Fatalf("Paste returned error: %v", err)
	}
	if path != "/tmp/from-file.png" {
		t.Fatalf("path = %q, want file path", path)
	}
	if cb.writtenText != "/tmp/from-file.png" {
		t.Fatalf("written text = %q, want file path", cb.writtenText)
	}
}

func TestPasteIgnoresPlainTextClipboard(t *testing.T) {
	cb := &fakeClipboard{
		filesErr: clipboard.ErrNoFile,
		imageErr: clipboard.ErrNoImage,
	}
	sender := &fakePasteSender{}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       sender,
	}

	_, err := svc.Paste(context.Background(), PasteOptions{CopyPath: true, SendPaste: true})
	if !errors.Is(err, clipboard.ErrNoImage) {
		t.Fatalf("error = %v, want ErrNoImage", err)
	}
	if cb.writeCount != 0 {
		t.Fatalf("writeCount = %d, want 0", cb.writeCount)
	}
	if sender.sendCount != 0 {
		t.Fatalf("sendCount = %d, want 0", sender.sendCount)
	}
}

func TestSmartPasteFallsBackToNativePaste(t *testing.T) {
	cb := &fakeClipboard{
		filesErr: clipboard.ErrNoFile,
		imageErr: clipboard.ErrNoImage,
	}
	sender := &fakePasteSender{}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       sender,
	}

	result, err := svc.SmartPaste(context.Background(), SmartPasteOptions{})
	if err != nil {
		t.Fatalf("SmartPaste returned error: %v", err)
	}
	if !result.NativePaste {
		t.Fatal("NativePaste = false, want true")
	}
	if result.Path != "" {
		t.Fatalf("Path = %q, want empty", result.Path)
	}
	if cb.writeCount != 0 {
		t.Fatalf("writeCount = %d, want 0", cb.writeCount)
	}
	if sender.sendCount != 1 {
		t.Fatalf("sendCount = %d, want 1", sender.sendCount)
	}
}

func TestPasteRejectsMultipleFilesWithoutWritingClipboard(t *testing.T) {
	cb := &fakeClipboard{
		files: []clipboard.FileRef{
			{Path: "/tmp/a.txt"},
			{Path: "/tmp/b.txt"},
		},
	}
	sender := &fakePasteSender{}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       sender,
	}

	_, err := svc.Paste(context.Background(), PasteOptions{CopyPath: true, SendPaste: true})
	if !errors.Is(err, clipboard.ErrMultiFile) {
		t.Fatalf("error = %v, want ErrMultiFile", err)
	}
	if cb.writeCount != 0 {
		t.Fatalf("writeCount = %d, want 0", cb.writeCount)
	}
	if sender.sendCount != 0 {
		t.Fatalf("sendCount = %d, want 0", sender.sendCount)
	}
}

func TestSmartPasteRejectsMultipleFilesWithoutNativeFallback(t *testing.T) {
	cb := &fakeClipboard{
		files: []clipboard.FileRef{
			{Path: "/tmp/a.txt"},
			{Path: "/tmp/b.txt"},
		},
	}
	sender := &fakePasteSender{}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       sender,
	}

	_, err := svc.SmartPaste(context.Background(), SmartPasteOptions{})
	if !errors.Is(err, clipboard.ErrMultiFile) {
		t.Fatalf("error = %v, want ErrMultiFile", err)
	}
	if cb.writeCount != 0 {
		t.Fatalf("writeCount = %d, want 0", cb.writeCount)
	}
	if sender.sendCount != 0 {
		t.Fatalf("sendCount = %d, want 0", sender.sendCount)
	}
}

func TestPasteTransformsOutputPathForClipboard(t *testing.T) {
	cb := &fakeClipboard{
		files: []clipboard.FileRef{{Path: `C:\Users\Alice\Desktop\a.txt`}},
	}
	svc := &Service{
		clipboard:   cb,
		hotkey:      hotkey.NewListener(),
		materialize: materialize.NewService(),
		paste:       paste.NewSystemSender(),
	}

	path, err := svc.Paste(context.Background(), PasteOptions{CopyPath: true, PathStyle: "wsl"})
	if err != nil {
		t.Fatalf("Paste returned error: %v", err)
	}
	if path != "/mnt/c/Users/Alice/Desktop/a.txt" {
		t.Fatalf("path = %q, want WSL path", path)
	}
	if cb.writtenText != "/mnt/c/Users/Alice/Desktop/a.txt" {
		t.Fatalf("written text = %q, want WSL path", cb.writtenText)
	}
}
