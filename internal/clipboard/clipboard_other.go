//go:build !darwin && !windows

package clipboard

import "context"

type systemClipboard struct{}

func NewSystemClipboard() Clipboard {
	return systemClipboard{}
}

func (systemClipboard) ReadImage(ctx context.Context) (Image, error) {
	return Image{}, ErrUnsupported
}

func (systemClipboard) ReadFiles(ctx context.Context) ([]FileRef, error) {
	return nil, ErrUnsupported
}

func (systemClipboard) CanWriteText(ctx context.Context) bool {
	return false
}

func (systemClipboard) WriteText(ctx context.Context, text string) error {
	return ErrUnsupported
}
