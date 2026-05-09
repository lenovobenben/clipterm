package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/lenovobenben/clipterm/internal/clipboard"
	"github.com/lenovobenben/clipterm/internal/clipterm"
	"github.com/lenovobenben/clipterm/internal/paste"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	ctx := context.Background()
	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "paste":
		return runPaste(ctx, commandArgs, stdout, stderr)
	case "daemon":
		fmt.Fprintln(stderr, "clipterm daemon is not implemented yet")
		return 1
	case "clean":
		fmt.Fprintln(stderr, "clipterm clean is not implemented yet")
		return 1
	case "doctor":
		return runDoctor(ctx, commandArgs, stdout, stderr)
	case "rules":
		fmt.Fprintln(stderr, "clipterm rules is not implemented yet")
		return 1
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", command)
		printUsage(stderr)
		return 2
	}
}

func runPaste(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("paste", flag.ContinueOnError)
	flags.SetOutput(stderr)

	copyPath := flags.Bool("copy-path", false, "copy generated path to the clipboard")
	sendPaste := flags.Bool("send-paste", false, "send a synthetic paste event after copying the path")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	service := clipterm.NewService()
	path, err := service.Paste(ctx, clipterm.PasteOptions{
		CopyPath:  *copyPath,
		SendPaste: *sendPaste,
	})
	if err != nil {
		printCommandError(stderr, err)
		return 1
	}

	fmt.Fprintln(stdout, path)
	return 0
}

func runDoctor(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)

	requestPermissions := flags.Bool("request-permissions", false, "request macOS permissions needed for synthetic paste")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	service := clipterm.NewService()
	if *requestPermissions {
		if service.RequestPastePermission(ctx) {
			fmt.Fprintln(stdout, "synthetic paste permission is already available")
		} else {
			fmt.Fprintln(stdout, "synthetic paste permission was requested; approve clipterm in System Settings if prompted")
		}
	}

	report := service.Doctor(ctx)
	fmt.Fprintf(stdout, "cache_dir: %s\n", report.CacheDir)
	fmt.Fprintf(stdout, "clipboard_text_write: %s\n", statusString(report.CanWriteClipboardText))
	fmt.Fprintf(stdout, "clipboard_image_read: %s\n", report.ClipboardImageRead)
	fmt.Fprintf(stdout, "synthetic_paste: %s\n", statusString(report.CanSendPaste))
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, strings.TrimSpace(`
clipterm

Copy screenshot. Paste path anywhere.

Usage:
  clipterm paste [--copy-path] [--send-paste]
  clipterm daemon
  clipterm clean
  clipterm doctor [--request-permissions]
  clipterm rules
`)+"\n")
}

func printCommandError(w io.Writer, err error) {
	switch {
	case errors.Is(err, clipboard.ErrNoImage):
		fmt.Fprintln(w, "clipboard does not contain a supported image")
	case errors.Is(err, clipboard.ErrUnsupported):
		fmt.Fprintln(w, "clipboard image reading is not supported by this build yet")
	case errors.Is(err, paste.ErrUnsupported):
		fmt.Fprintln(w, "synthetic paste is not supported or not permitted by this build")
	default:
		fmt.Fprintf(w, "error: %v\n", err)
	}
}

func statusString(ok bool) string {
	if ok {
		return "ok"
	}
	return "unavailable"
}
