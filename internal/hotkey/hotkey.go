package hotkey

import "context"

type Handler func(context.Context)

type Options struct {
	Debug bool
}

type Listener interface {
	CanListen(context.Context) bool
	RequestPermission(context.Context) bool
	Run(ctx context.Context, options Options, handler Handler) error
}
