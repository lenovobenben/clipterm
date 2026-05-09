package paste

import (
	"context"
	"errors"
)

var ErrUnsupported = errors.New("paste operation unsupported")

type Sender interface {
	CanSendPaste(ctx context.Context) bool
	RequestPastePermission(ctx context.Context) bool
	SendPaste(ctx context.Context) error
}
