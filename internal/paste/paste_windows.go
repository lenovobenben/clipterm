//go:build windows

package paste

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	inputKeyboard = 1

	keyeventfKeyUp = 0x0002

	vkControl  = 0x11
	vkShift    = 0x10
	vkV        = 0x56
	vkLShift   = 0xA0
	vkRShift   = 0xA1
	vkLControl = 0xA2
	vkRControl = 0xA3
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	procSendInput        = user32.NewProc("SendInput")
)

type systemSender struct{}

type keyboardInput struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type input struct {
	Type uint32
	_    uint32
	Ki   keyboardInput
	_    [8]byte
}

func NewSystemSender() Sender {
	return systemSender{}
}

func (systemSender) CanSendPaste(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func (systemSender) RequestPastePermission(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func (systemSender) SendPaste(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	shiftKey := pressedShiftKey()
	ctrlDown := keyDownNow(vkControl) || keyDownNow(vkLControl) || keyDownNow(vkRControl)

	inputs := make([]input, 0, 6)
	if shiftKey != 0 {
		inputs = append(inputs, keyUp(shiftKey))
	}
	if !ctrlDown {
		inputs = append(inputs, keyDown(vkControl))
	}
	inputs = append(inputs, keyDown(vkV), keyUp(vkV))
	if !ctrlDown {
		inputs = append(inputs, keyUp(vkControl))
	}
	if shiftKey != 0 {
		inputs = append(inputs, keyDown(shiftKey))
	}

	if err := sendInputs(inputs); err != nil {
		return err
	}

	debugf("windows paste: sent Ctrl+V shiftReleased=%v", shiftKey != 0)
	return nil
}

func sendInputs(inputs []input) error {
	if len(inputs) == 0 {
		return nil
	}

	sent, _, _ := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if sent != uintptr(len(inputs)) {
		return fmt.Errorf("send Ctrl+V failed: %w", ErrUnsupported)
	}
	return nil
}

func keyDownNow(vk uint16) bool {
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return state&0x8000 != 0
}

func pressedShiftKey() uint16 {
	if keyDownNow(vkLShift) {
		return vkLShift
	}
	if keyDownNow(vkRShift) {
		return vkRShift
	}
	if keyDownNow(vkShift) {
		return vkShift
	}
	return 0
}

func keyDown(vk uint16) input {
	return input{
		Type: inputKeyboard,
		Ki: keyboardInput{
			Vk: vk,
		},
	}
}

func keyUp(vk uint16) input {
	event := keyDown(vk)
	event.Ki.Flags = keyeventfKeyUp
	return event
}
