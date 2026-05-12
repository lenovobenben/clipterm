//go:build (!darwin && !windows) || (darwin && !cgo)

package hotkey

import (
	"context"

	"github.com/lenovobenben/clipterm/internal/paste"
)

type unsupportedListener struct{}

func NewListener() Listener {
	return unsupportedListener{}
}

func (unsupportedListener) CanListen(ctx context.Context) bool {
	return false
}

func (unsupportedListener) RequestPermission(ctx context.Context) bool {
	return false
}

func (unsupportedListener) Run(ctx context.Context, options Options, handler Handler) error {
	return paste.ErrUnsupported
}
