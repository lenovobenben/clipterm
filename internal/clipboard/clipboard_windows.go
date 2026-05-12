//go:build windows

package clipboard

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"time"
	"unsafe"
)

const (
	cfUnicodeText = 13
	cfDIB         = 8
	cfDIBV5       = 17
	cfHDrop       = 15

	gmemMoveable = 0x0002

	biRGB       = 0
	biBitFields = 3
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procEmptyClipboard             = user32.NewProc("EmptyClipboard")
	procSetClipboardData           = user32.NewProc("SetClipboardData")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procIsClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")

	procGlobalAlloc  = kernel32.NewProc("GlobalAlloc")
	procGlobalFree   = kernel32.NewProc("GlobalFree")
	procGlobalLock   = kernel32.NewProc("GlobalLock")
	procGlobalSize   = kernel32.NewProc("GlobalSize")
	procGlobalUnlock = kernel32.NewProc("GlobalUnlock")

	procDragQueryFileW = shell32.NewProc("DragQueryFileW")
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

	format := uint32(cfDIBV5)
	if !isClipboardFormatAvailable(format) {
		format = cfDIB
	}
	if !isClipboardFormatAvailable(format) {
		return Image{}, ErrNoImage
	}

	if err := openClipboard(ctx); err != nil {
		debugf("windows clipboard: open clipboard failed: %v", err)
		return Image{}, err
	}
	defer closeClipboard()

	dib, err := readClipboardData(format)
	if err != nil {
		return Image{}, err
	}

	pngData, err := dibToPNG(dib)
	if err != nil {
		debugf("windows clipboard: read image failed: %v", err)
		return Image{}, ErrUnsupported
	}
	debugf("windows clipboard: read image format=CF_DIB len=%d png=%d", len(dib), len(pngData))
	return Image{Data: pngData, Format: "png", Extension: ".png"}, nil
}

func (systemClipboard) ReadFiles(ctx context.Context) ([]FileRef, error) {
	files, err := readFilesOnce(ctx)
	if err != nil {
		return nil, err
	}
	debugf("windows clipboard: files=%v", files)
	return files, nil
}

func readFilesOnce(ctx context.Context) ([]FileRef, error) {
	if !isClipboardFormatAvailable(cfHDrop) {
		return nil, ErrNoFile
	}
	if err := openClipboard(ctx); err != nil {
		debugf("windows clipboard: open clipboard failed: %v", err)
		return nil, err
	}
	defer closeClipboard()

	return readOpenClipboardFiles()
}

func (systemClipboard) CanWriteText(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	if err := openClipboard(ctx); err != nil {
		return false
	}
	closeClipboard()
	return true
}

func (systemClipboard) WriteText(ctx context.Context, text string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	utf16Text, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	if err := openClipboard(ctx); err != nil {
		return err
	}

	if ok := emptyClipboard(); !ok {
		closeClipboard()
		return ErrUnsupported
	}

	size := uintptr(len(utf16Text) * 2)
	handle, _, _ := procGlobalAlloc.Call(gmemMoveable, size)
	if handle == 0 {
		closeClipboard()
		return ErrUnsupported
	}

	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		procGlobalFree.Call(handle)
		closeClipboard()
		return ErrUnsupported
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16Text)), utf16Text)
	procGlobalUnlock.Call(handle)

	result, _, _ := procSetClipboardData.Call(uintptr(cfUnicodeText), handle)
	if result == 0 {
		procGlobalFree.Call(handle)
		closeClipboard()
		debugf("windows clipboard: write text failed")
		return ErrUnsupported
	}

	closeClipboard()
	debugf("windows clipboard: wrote text=%q", text)
	return nil
}

func openClipboard(ctx context.Context) error {
	deadline := time.NewTimer(500 * time.Millisecond)
	defer deadline.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		result, _, _ := procOpenClipboard.Call(0)
		if result != 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return ErrUnsupported
		case <-ticker.C:
		}
	}
}

func closeClipboard() {
	procCloseClipboard.Call()
}

func emptyClipboard() bool {
	result, _, _ := procEmptyClipboard.Call()
	return result != 0
}

func isClipboardFormatAvailable(format uint32) bool {
	result, _, _ := procIsClipboardFormatAvailable.Call(uintptr(format))
	return result != 0
}

func dragQueryFileCount(handle uintptr) uint32 {
	count, _, _ := procDragQueryFileW.Call(handle, ^uintptr(0), 0, 0)
	return uint32(count)
}

func dragQueryFilePath(handle uintptr, index uint32) string {
	length, _, _ := procDragQueryFileW.Call(handle, uintptr(index), 0, 0)
	if length == 0 {
		return ""
	}

	buffer := make([]uint16, int(length)+1)
	procDragQueryFileW.Call(
		handle,
		uintptr(index),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	return syscall.UTF16ToString(buffer)
}

func readOpenClipboardFiles() ([]FileRef, error) {
	handle, _, _ := procGetClipboardData.Call(uintptr(cfHDrop))
	if handle == 0 {
		return nil, ErrNoFile
	}

	count := dragQueryFileCount(handle)
	if count == 0 {
		return nil, ErrNoFile
	}

	files := make([]FileRef, 0, count)
	for i := uint32(0); i < count; i++ {
		path := dragQueryFilePath(handle, i)
		if path == "" {
			continue
		}
		files = append(files, FileRef{Path: path})
	}
	if len(files) == 0 {
		return nil, ErrNoFile
	}

	return files, nil
}

func readClipboardData(format uint32) ([]byte, error) {
	handle, _, _ := procGetClipboardData.Call(uintptr(format))
	if handle == 0 {
		return nil, ErrUnsupported
	}

	size, _, _ := procGlobalSize.Call(handle)
	if size == 0 {
		return nil, ErrUnsupported
	}

	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		return nil, ErrUnsupported
	}
	defer procGlobalUnlock.Call(handle)

	data := make([]byte, int(size))
	copy(data, unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(size)))
	return data, nil
}

func dibToPNG(dib []byte) ([]byte, error) {
	img, err := decodeDIB(dib)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func decodeDIB(dib []byte) (image.Image, error) {
	if len(dib) < 40 {
		return nil, ErrUnsupported
	}

	headerSize := int(binary.LittleEndian.Uint32(dib[0:4]))
	if headerSize < 40 || len(dib) < headerSize {
		return nil, ErrUnsupported
	}

	width := int(int32(binary.LittleEndian.Uint32(dib[4:8])))
	rawHeight := int(int32(binary.LittleEndian.Uint32(dib[8:12])))
	planes := binary.LittleEndian.Uint16(dib[12:14])
	bitCount := binary.LittleEndian.Uint16(dib[14:16])
	compression := binary.LittleEndian.Uint32(dib[16:20])
	clrUsed := uint32(0)
	if len(dib) >= 40 {
		clrUsed = binary.LittleEndian.Uint32(dib[32:36])
	}

	if width <= 0 || rawHeight == 0 || planes != 1 {
		return nil, ErrUnsupported
	}
	if compression != biRGB && compression != biBitFields {
		return nil, ErrUnsupported
	}
	if bitCount != 24 && bitCount != 32 {
		return nil, ErrUnsupported
	}

	height := rawHeight
	topDown := rawHeight < 0
	if topDown {
		height = -height
	}

	pixelOffset := headerSize + colorTableSize(bitCount, clrUsed)
	if compression == biBitFields && headerSize == 40 {
		pixelOffset += 12
	}
	if pixelOffset < 0 || pixelOffset > len(dib) {
		return nil, ErrUnsupported
	}

	rowStride := ((width*int(bitCount) + 31) / 32) * 4
	required := pixelOffset + rowStride*height
	if rowStride <= 0 || required > len(dib) {
		return nil, ErrUnsupported
	}

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	hasAlpha := false
	for y := 0; y < height; y++ {
		srcY := y
		if !topDown {
			srcY = height - 1 - y
		}
		row := dib[pixelOffset+srcY*rowStride:]
		for x := 0; x < width; x++ {
			switch bitCount {
			case 24:
				i := x * 3
				rgba.SetRGBA(x, y, color.RGBA{
					R: row[i+2],
					G: row[i+1],
					B: row[i],
					A: 0xff,
				})
			case 32:
				i := x * 4
				alpha := row[i+3]
				if alpha != 0 {
					hasAlpha = true
				}
				rgba.SetRGBA(x, y, color.RGBA{
					R: row[i+2],
					G: row[i+1],
					B: row[i],
					A: alpha,
				})
			}
		}
	}

	if bitCount == 32 && !hasAlpha {
		for i := 3; i < len(rgba.Pix); i += 4 {
			rgba.Pix[i] = 0xff
		}
	}

	return rgba, nil
}

func colorTableSize(bitCount uint16, clrUsed uint32) int {
	if bitCount > 8 {
		return 0
	}
	if clrUsed > 0 {
		return int(clrUsed) * 4
	}
	return (1 << bitCount) * 4
}
