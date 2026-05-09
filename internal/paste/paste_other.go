//go:build !darwin

package paste

import "context"

type systemSender struct{}

func NewSystemSender() Sender {
	return systemSender{}
}

func (systemSender) CanSendPaste(ctx context.Context) bool {
	return false
}

func (systemSender) RequestPastePermission(ctx context.Context) bool {
	return false
}

func (systemSender) SendPaste(ctx context.Context) error {
	return ErrUnsupported
}
