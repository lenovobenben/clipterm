//go:build windows

package hotkey

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	hotkeyID = 1

	modShift    = 0x0004
	modControl  = 0x0002
	modNoRepeat = 0x4000

	vkV = 0x56

	wmHotkey = 0x0312
	wmQuit   = 0x0012
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")

	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

type windowsListener struct{}

type point struct {
	X int32
	Y int32
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

func NewListener() Listener {
	return windowsListener{}
}

func (windowsListener) CanListen(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	if !registerHotKey(hotkeyID + 1) {
		return false
	}
	unregisterHotKey(hotkeyID + 1)
	return true
}

func (windowsListener) RequestPermission(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func (windowsListener) Run(ctx context.Context, options Options, handler Handler) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if !registerHotKey(hotkeyID) {
		return fmt.Errorf("register Ctrl+Shift+V hotkey failed")
	}
	defer unregisterHotKey(hotkeyID)

	threadID := currentThreadID()
	go func() {
		<-ctx.Done()
		postThreadMessage(threadID, wmQuit)
	}()

	var message msg
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(result) == -1 {
			return fmt.Errorf("get message failed")
		}
		if result == 0 || message.Message == wmQuit {
			break
		}
		if message.Message != wmHotkey || message.WParam != hotkeyID {
			continue
		}
		if options.Debug {
			fmt.Println("hotkey event Ctrl+Shift+V")
		}
		handler(ctx)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func registerHotKey(id int) bool {
	result, _, _ := procRegisterHotKey.Call(0, uintptr(id), modControl|modShift|modNoRepeat, vkV)
	return result != 0
}

func unregisterHotKey(id int) {
	procUnregisterHotKey.Call(0, uintptr(id))
}

func currentThreadID() uint32 {
	result, _, _ := procGetCurrentThreadId.Call()
	return uint32(result)
}

func postThreadMessage(threadID uint32, message uint32) {
	procPostThreadMessageW.Call(uintptr(threadID), uintptr(message), 0, 0)
}
