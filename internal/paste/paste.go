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

var debugLogger func(format string, args ...any)

func SetDebugLogger(logger func(format string, args ...any)) {
	debugLogger = logger
}

func debugf(format string, args ...any) {
	if debugLogger != nil {
		debugLogger(format, args...)
	}
}
