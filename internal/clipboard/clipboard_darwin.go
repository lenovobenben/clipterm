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

func (systemClipboard) ReadFiles(ctx context.Context) ([]FileRef, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := C.clipterm_read_clipboard_files()
	defer C.clipterm_free_files_result(result)

	if result.err != nil {
		err := C.GoString(result.err)
		if err == "no file in clipboard" {
			return nil, ErrNoFile
		}
		return nil, ErrUnsupported
	}

	if result.paths == nil || result.count <= 0 {
		return nil, ErrNoFile
	}

	paths := unsafe.Slice(result.paths, int(result.count))
	files := make([]FileRef, 0, int(result.count))
	for _, path := range paths {
		if path == nil {
			continue
		}
		files = append(files, FileRef{Path: C.GoString(path)})
	}

	if len(files) == 0 {
		return nil, ErrNoFile
	}

	return files, nil
}

func (systemClipboard) CanWriteText(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func (systemClipboard) WriteText(ctx context.Context, text string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	errText := C.clipterm_write_clipboard_text(cText)
	defer C.clipterm_free_error(errText)
	if errText != nil {
		return ErrUnsupported
	}

	return nil
}
