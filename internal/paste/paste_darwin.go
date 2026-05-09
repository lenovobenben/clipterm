//go:build darwin && cgo

package paste

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>

static int clipterm_can_send_paste(void) {
    return AXIsProcessTrusted() ? 1 : 0;
}

static int clipterm_request_paste_permission(void) {
    const void *keys[] = { kAXTrustedCheckOptionPrompt };
    const void *values[] = { kCFBooleanTrue };
    CFDictionaryRef options = CFDictionaryCreate(
        kCFAllocatorDefault,
        keys,
        values,
        1,
        &kCFCopyStringDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );
    if (options == NULL) {
        return 0;
    }

    Boolean trusted = AXIsProcessTrustedWithOptions(options);
    CFRelease(options);
    return trusted ? 1 : 0;
}

static int clipterm_send_cmd_v(void) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    if (source == NULL) {
        return 1;
    }

    CGEventRef keyDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, true);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)9, false);
    if (keyDown == NULL || keyUp == NULL) {
        if (keyDown != NULL) {
            CFRelease(keyDown);
        }
        if (keyUp != NULL) {
            CFRelease(keyUp);
        }
        CFRelease(source);
        return 2;
    }

    CGEventSetFlags(keyDown, kCGEventFlagMaskCommand);
    CGEventSetFlags(keyUp, kCGEventFlagMaskCommand);
    CGEventPost(kCGHIDEventTap, keyDown);
    CGEventPost(kCGHIDEventTap, keyUp);

    CFRelease(keyDown);
    CFRelease(keyUp);
    CFRelease(source);
    return 0;
}
*/
import "C"

import (
	"context"
	"fmt"
)

type systemSender struct{}

func NewSystemSender() Sender {
	return systemSender{}
}

func (systemSender) CanSendPaste(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return C.clipterm_can_send_paste() == 1
}

func (systemSender) RequestPastePermission(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return C.clipterm_request_paste_permission() == 1
}

func (systemSender) SendPaste(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	code := C.clipterm_send_cmd_v()
	if code != 0 {
		return fmt.Errorf("send paste event failed: %w", ErrUnsupported)
	}

	return nil
}
