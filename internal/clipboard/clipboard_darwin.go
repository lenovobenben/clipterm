//go:build darwin && cgo

package clipboard

/*
#cgo LDFLAGS: -framework AppKit
#include <stdlib.h>
#include "pasteboard_darwin.h"
*/
import "C"

import (
	"context"
	"io"
	"os/exec"
	"unsafe"
)

type systemClipboard struct{}

func NewSystemClipboard() Clipboard {
	return systemClipboard{}
}

func (systemClipboard) ReadImage(ctx context.Context) (Image, error) {
	select {
	case <-ctx.Done():
		return Image{}, ctx.Err()
	default:
	}

	result := C.clipterm_read_clipboard_image_png()
	defer C.clipterm_free_image_result(result)

	if result.err != nil {
		err := C.GoString(result.err)
		if err == "no image in clipboard" {
			return Image{}, ErrNoImage
		}
		return Image{}, ErrUnsupported
	}

	if result.data == nil || result.len <= 0 {
		return Image{}, ErrNoImage
	}

	data := C.GoBytes(unsafe.Pointer(result.data), C.int(result.len))
	return Image{
		Data:      data,
		Format:    "png",
		Extension: ".png",
	}, nil
}

func (systemClipboard) CanWriteText(ctx context.Context) bool {
	_, err := exec.LookPath("pbcopy")
	return err == nil
}

func (systemClipboard) WriteText(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, "pbcopy")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.WriteString(stdin, text); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return err
	}

	if err := stdin.Close(); err != nil {
		_ = cmd.Wait()
		return err
	}

	return cmd.Wait()
}
