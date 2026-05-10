//go:build darwin && cgo

package hotkey

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>

extern int cliptermHandleKey(unsigned short keyCode, unsigned long long flags);

static CFRunLoopRef clipterm_run_loop = NULL;

static int clipterm_can_listen_keyboard(void) {
    return CGPreflightListenEventAccess() ? 1 : 0;
}

static int clipterm_request_listen_keyboard(void) {
    return CGRequestListenEventAccess() ? 1 : 0;
}

static CGEventRef clipterm_event_tap_callback(
    CGEventTapProxy proxy,
    CGEventType type,
    CGEventRef event,
    void *userInfo
) {
    if (type != kCGEventKeyDown) {
        return event;
    }

    CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    CGEventFlags flags = CGEventGetFlags(event);

    if (cliptermHandleKey((unsigned short)keyCode, (unsigned long long)flags) == 1) {
        return NULL;
    }

    return event;
}

static int clipterm_run_cmd_shift_v_event_tap(void) {
    CGEventMask mask = CGEventMaskBit(kCGEventKeyDown);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        mask,
        clipterm_event_tap_callback,
        NULL
    );
    if (eventTap == NULL) {
        return 1;
    }

    CFRunLoopSourceRef source = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    if (source == NULL) {
        CFRelease(eventTap);
        return 2;
    }

    clipterm_run_loop = CFRunLoopGetCurrent();

    CFRunLoopAddSource(clipterm_run_loop, source, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    CFRunLoopRun();

    CFRunLoopRemoveSource(clipterm_run_loop, source, kCFRunLoopCommonModes);
    clipterm_run_loop = NULL;
    CFRelease(source);
    CFRelease(eventTap);
    return 0;
}

static void clipterm_stop_event_tap(void) {
    if (clipterm_run_loop != NULL) {
        CFRunLoopStop(clipterm_run_loop);
    }
}
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

const (
	keyCodeV = 9

	eventFlagShift   = 1 << 17
	eventFlagCommand = 1 << 20
)

var hotkeyEvents chan struct{}
var debugHotkeys bool

type darwinListener struct{}

func NewListener() Listener {
	return darwinListener{}
}

func (darwinListener) CanListen(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return C.clipterm_can_listen_keyboard() == 1
}

func (darwinListener) RequestPermission(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return C.clipterm_request_listen_keyboard() == 1
}

func (darwinListener) Run(ctx context.Context, options Options, handler Handler) error {
	hotkeyEvents = make(chan struct{}, 8)
	debugHotkeys = options.Debug
	defer func() {
		hotkeyEvents = nil
		debugHotkeys = false
	}()

	var handlers sync.WaitGroup
	handlers.Add(1)
	go func() {
		defer handlers.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-hotkeyEvents:
				if !ok {
					return
				}
				handler(ctx)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		C.clipterm_stop_event_tap()
	}()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	status := C.clipterm_run_cmd_shift_v_event_tap()
	close(hotkeyEvents)
	handlers.Wait()

	if status != 0 {
		return fmt.Errorf("start event tap failed: status %d", int(status))
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

//export cliptermHandleKey
func cliptermHandleKey(keyCode C.ushort, flags C.ulonglong) C.int {
	if debugHotkeys {
		fmt.Printf("hotkey event keycode=%d flags=%#x\n", uint16(keyCode), uint64(flags))
	}

	if uint16(keyCode) != keyCodeV {
		return 0
	}

	goFlags := uint64(flags)
	hasCommand := goFlags&eventFlagCommand != 0
	hasShift := goFlags&eventFlagShift != 0
	if !hasCommand || !hasShift {
		return 0
	}

	if hotkeyEvents == nil {
		return 1
	}

	select {
	case hotkeyEvents <- struct{}{}:
	default:
	}
	return 1
}
