package clipboard

import (
	"context"
	"errors"
)

var (
	ErrNoImage     = errors.New("clipboard has no image")
	ErrNoFile      = errors.New("clipboard has no file")
	ErrMultiFile   = errors.New("clipboard has multiple files")
	ErrUnsupported = errors.New("clipboard operation unsupported")
)

type Image struct {
	Data      []byte
	Format    string
	Extension string
}

type FileRef struct {
	Path string
}

type Clipboard interface {
	ReadImage(ctx context.Context) (Image, error)
	ReadFiles(ctx context.Context) ([]FileRef, error)
	WriteText(ctx context.Context, text string) error
}
