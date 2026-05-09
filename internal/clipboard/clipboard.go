package clipboard

import (
	"context"
	"errors"
)

var (
	ErrNoImage     = errors.New("clipboard has no image")
	ErrUnsupported = errors.New("clipboard operation unsupported")
)

type Image struct {
	Data      []byte
	Format    string
	Extension string
}

type Clipboard interface {
	ReadImage(ctx context.Context) (Image, error)
	WriteText(ctx context.Context, text string) error
}
