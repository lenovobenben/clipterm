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
	"github.com/lenovobenben/clipterm/internal/daemon"
	"github.com/lenovobenben/clipterm/internal/paste"
	"github.com/lenovobenben/clipterm/internal/version"
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
		return runDaemon(ctx, commandArgs, stdout, stderr)
	case "clean":
		return runClean(ctx, commandArgs, stdout, stderr)
	case "doctor":
		return runDoctor(ctx, commandArgs, stdout, stderr)
	case "rules":
		fmt.Fprintln(stderr, "clipterm rules is not implemented yet")
		return 1
	case "version":
		fmt.Fprintln(stdout, version.Version)
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", command)
		printUsage(stderr)
		return 2
	}
}

func runClean(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("clean", flag.ContinueOnError)
	flags.SetOutput(stderr)

	days := flags.Int("days", 7, "remove cached clipterm images older than this many days")
	dryRun := flags.Bool("dry-run", false, "show files that would be removed without deleting them")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	result, err := clipterm.NewService().Clean(ctx, clipterm.CleanOptions{
		Days:   *days,
		DryRun: *dryRun,
	})
	if err != nil {
		printCommandError(stderr, err)
		return 1
	}

	action := "removed"
	if result.DryRun {
		action = "would_remove"
	}

	fmt.Fprintf(stdout, "cache_dir: %s\n", result.CacheDir)
	fmt.Fprintf(stdout, "%s_files: %d\n", action, len(result.Files))
	fmt.Fprintf(stdout, "%s_bytes: %d\n", action, result.Bytes)
	for _, path := range result.Files {
		fmt.Fprintln(stdout, path)
	}

	return 0
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

func runDaemon(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("daemon", flag.ContinueOnError)
	flags.SetOutput(stderr)

	foreground := flags.Bool("foreground", false, "run daemon in the foreground")
	stopDaemon := flags.Bool("stop", false, "stop the background daemon")
	statusDaemon := flags.Bool("status", false, "show background daemon status")
	debugHotkeys := flags.Bool("debug-hotkeys", false, "print captured key codes while daemon is running")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	switch {
	case *stopDaemon:
		if err := daemon.Stop(); err != nil {
			printCommandError(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, "daemon stopped")
		return 0
	case *statusDaemon:
		status, err := daemon.CurrentStatus()
		if err != nil {
			printCommandError(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "running: %s\n", yesNoString(status.Running))
		if status.PID > 0 {
			fmt.Fprintf(stdout, "pid: %d\n", status.PID)
		}
		return 0
	case !*foreground:
		status, err := daemon.Start(ctx, daemon.StartOptions{
			DebugHotkeys: *debugHotkeys,
		})
		if err != nil {
			printCommandError(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "daemon started pid=%d\n", status.PID)
		logDir, err := daemon.LogDir()
		if err == nil {
			fmt.Fprintf(stdout, "logs: %s\n", logDir)
		}
		return 0
	}

	service := clipterm.NewService()
	fmt.Fprintln(stderr, "clipterm daemon listening on Cmd+Shift+V")

	err := service.RunDaemon(ctx, clipterm.DaemonOptions{
		DebugHotkeys: *debugHotkeys,
	}, func(ctx context.Context) {
		path, err := service.Paste(ctx, clipterm.PasteOptions{
			CopyPath:  true,
			SendPaste: true,
		})
		if err != nil {
			if errors.Is(err, clipboard.ErrNoImage) {
				return
			}
			printCommandError(stderr, err)
			return
		}
		fmt.Fprintln(stdout, path)
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		printCommandError(stderr, err)
		return 1
	}

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
		if service.RequestHotkeyPermission(ctx) {
			fmt.Fprintln(stdout, "global hotkey permission is already available")
		} else {
			fmt.Fprintln(stdout, "global hotkey permission was requested; approve clipterm in System Settings if prompted")
		}
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
	fmt.Fprintf(stdout, "clipboard_file_read: %s\n", report.ClipboardFileRead)
	fmt.Fprintf(stdout, "global_hotkey: %s\n", statusString(report.CanListenHotkey))
	fmt.Fprintf(stdout, "synthetic_paste: %s\n", statusString(report.CanSendPaste))
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, strings.TrimSpace(`
clipterm

Copy screenshot. Paste path anywhere.

Usage:
  clipterm paste [--copy-path] [--send-paste]
  clipterm daemon [--foreground] [--debug-hotkeys]
  clipterm daemon --status
  clipterm daemon --stop
  clipterm clean [--days 7] [--dry-run]
  clipterm doctor [--request-permissions]
  clipterm rules
  clipterm version
`)+"\n")
}

func printCommandError(w io.Writer, err error) {
	switch {
	case errors.Is(err, clipboard.ErrNoImage):
		fmt.Fprintln(w, "clipboard does not contain a supported image or single file")
	case errors.Is(err, clipboard.ErrNoFile):
		fmt.Fprintln(w, "clipboard does not contain a supported image or single file")
	case errors.Is(err, clipboard.ErrMultiFile):
		fmt.Fprintln(w, "clipboard contains multiple files; only single-file path paste is supported")
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

func yesNoString(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}
