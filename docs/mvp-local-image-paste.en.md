# MVP: Local CLI/Agent Smart Paste

This document records the current clipterm MVP product semantics, platform behavior, known limits, and later directions. The README is the quick entry point; this document is for maintainers and future development.

## Goal

clipterm solves one concrete workflow problem: before a screenshot or clipboard image can be referenced by a terminal, CLI, AI agent, or plain text input, the user usually has to save it as a file and then copy the file path. clipterm turns that sequence into one system-level smart paste action.

Core model:

```text
clipboard image/file -> local file path -> paste into current input
```

The current MVP focuses on local smart paste as the core capability:

- When the clipboard contains an image or screenshot, smart paste writes a local PNG file and pastes its path.
- When the clipboard contains one copied file, smart paste pastes the original absolute file path.
- When the clipboard contains normal text, smart paste behaves like native paste.
- Normal `Cmd+V` / `Ctrl+V` is not intercepted.

This MVP only turns local clipboard objects into locally accessible paths.

## Current Platform Status

| Platform | Smart paste shortcut | Normal paste | Status |
| --- | --- | --- | --- |
| macOS | `Cmd+Shift+V` | `Cmd+V` is not intercepted | Prototype implemented |
| Windows x86_64 | `Ctrl+Shift+V` | `Ctrl+V` is not intercepted | Prototype implemented |
| Windows ARM | - | - | Not supported |
| Linux Desktop | - | - | Deferred |

## Unified Behavior

| Clipboard content | Smart paste behavior |
| --- | --- |
| Plain text and ordinary clipboard content | Send one native paste event without modifying the clipboard. |
| Single copied file | Paste the file's absolute path. The file content is not copied. |
| Screenshot or image stream | Save a PNG under the clipterm cache directory, then paste the generated path. |
| Multiple copied files | Not supported yet; return an error and do not modify the clipboard. |

After a successful file or image path paste, the system clipboard intentionally contains that path text. Users can press normal `Cmd+V` / `Ctrl+V` repeatedly to paste the same path until they copy something else. Text fallback does not modify the clipboard.

## Non-Goals

The current MVP does not include:

- Normal `Cmd+V` / `Ctrl+V` interception.
- App-specific native image paste allow rules.
- Multi-file, directory, or large-file transfer.
- Linux desktop support.

These directions belong under "Later Directions (Not Guaranteed)." They should not change the stable semantics of the current smart paste shortcut.

## macOS Design

macOS uses `Cmd+Shift+V` as the CLI/agent smart paste shortcut.

Cache directory:

```text
~/Library/Caches/clipterm/
```

Required permissions:

- Input Monitoring / keyboard event access.
- Accessibility / synthetic paste events.

macOS paste is object-based, not just text-based. When Finder copies a file, the pasteboard can expose several representations such as file URL, display name, icon, and preview metadata. Each target app chooses the representation it understands. For that reason, clipterm does not intercept normal `Cmd+V`; `Cmd+Shift+V` explicitly means "paste a CLI/agent-friendly path."

## Windows Design

Windows support runs as a native Windows `.exe` and listens for `Ctrl+Shift+V`. A Linux process running inside WSL cannot directly register Windows global hotkeys, read the Windows clipboard, or call `SendInput`, so the Windows build must run as a Windows process.

Cache directory:

```text
%LOCALAPPDATA%\clipterm\cache\
```

Main platform APIs:

- `RegisterHotKey`: register `Ctrl+Shift+V`.
- `CF_HDROP`: read file paths copied from Explorer.
- `CF_DIBV5` / `CF_DIB`: read common image clipboard data.
- `CF_UNICODETEXT`: write path text.
- `SendInput`: send native `Ctrl+V`.

Windows image flow:

```text
CF_DIB/CF_DIBV5 -> RGBA image -> PNG -> cache file -> output path
```

The current implementation supports common 24-bit and 32-bit uncompressed DIB images. PNG encoding is lossless. 32-bit DIB alpha is preserved when present; if all alpha values are zero, the image is treated as opaque, which matches common Windows screenshot behavior and avoids producing a fully transparent PNG. Palette bitmaps, compressed bitmaps, and more complex color-space variants can be added later.

### Windows Path Style

Windows native paths and WSL paths are different absolute-path formats:

```text
C:\Users\Alice\Pictures\a.png
/mnt/c/Users/Alice/Pictures/a.png
```

The Windows version selects output path style at daemon startup:

```powershell
clipterm.exe daemon --path-style windows
clipterm.exe daemon --path-style wsl
```

Semantics:

- `windows`: keep the native Windows path.
- `wsl`: convert drive paths to `/mnt/<drive>/...`.
- `native`: keep the original path, mainly for cross-platform default behavior.

Path conversion happens before writing text to the clipboard:

```text
clipboard object -> Windows path -> output path transform -> clipboard text -> Ctrl+V
```

Changing path style requires restarting the daemon. The current version does not hot-reload config.

### Validated Windows Targets

Validated targets:

- Notepad
- Browser address bars
- PowerShell
- Windows Terminal
- WSL terminal entered from PowerShell

Known limitation:

- Notepad++ is not a supported target for the first Windows version. In testing, clipterm successfully writes the path to the system clipboard, but Notepad++ handles the synthetic paste event during `Ctrl+Shift+V` differently from Notepad, browser address bars, PowerShell, and Windows Terminal, and may not insert the path into the editor.
- This is not a path conversion or clipboard write failure; PowerShell `Get-Clipboard` shows the path text after the same operation. Future Notepad++ support should be handled as separate compatibility work.

## Command Shape

```bash
clipterm paste
clipterm paste --copy-path
clipterm paste --copy-path --send-paste
clipterm daemon
clipterm daemon --foreground --debug-hotkeys
clipterm daemon --status
clipterm daemon --stop
clipterm rules
clipterm clean
clipterm doctor
clipterm version
```

`clipterm paste` is the debuggable core primitive: read a clipboard image or single file, produce a path, and print it to stdout.

`clipterm daemon` is the normal interaction entry point: listen for the smart paste shortcut, convert files/images into paths, or fall back to native text paste.

## Cache Strategy

Use the user cache directory:

```text
macOS:   ~/Library/Caches/clipterm/
Windows: %LOCALAPPDATA%\clipterm\cache\
```

Generated filename format:

```text
clipterm-YYYYMMDD-HHMMSS-<short-random>.png
```

`clipterm clean` only removes managed `clipterm-*.png` files.

## Code Structure

Core behavior:

- `internal/clipterm`: file-first, image-second, text fallback, multi-file refusal.
- `internal/materialize`: image materialization, cache directory, cleanup.
- `internal/pathstyle`: Windows / WSL output path conversion.

Platform adapters:

- `internal/clipboard/*`
- `internal/hotkey/*`
- `internal/paste/*`
- `internal/daemon/process_*`

CLI entry:

- `internal/cli`
- `cmd/clipterm`

Platform code should stay isolated. Windows hotkey, clipboard, and paste behavior should not affect the already validated macOS paths.

## Acceptance Criteria

macOS:

- Plain text: `Cmd+Shift+V` behaves like native `Cmd+V` and does not modify the clipboard.
- Finder single file: paste the original absolute file path.
- Image/screenshot: save PNG and paste the generated path.
- Multiple files: refuse and do not modify the clipboard.

Windows:

- Plain text: `Ctrl+Shift+V` behaves like native `Ctrl+V` and does not modify the clipboard.
- Explorer single file: paste the file path, with `windows` / `wsl` output styles.
- Image/screenshot: save PNG and paste the generated path, with `windows` / `wsl` output styles.
- Multiple files: refuse and do not modify the clipboard.
- Normal `Ctrl+V` is not intercepted.
- Starting the daemon repeatedly does not create multiple background processes.

## Later Directions (Not Guaranteed)

The following items are possibilities, not commitments for the current MVP:

- Extend Windows image format coverage, such as palette DIB, compressed DIB, or direct PNG clipboard formats.
- Provide a more formal Windows startup path, such as Startup folder, Task Scheduler, or an installer.
- Support a configurable smart paste shortcut.
- Study path shell escaping; the current behavior pastes raw absolute paths.
- Study normal `Cmd+V` / `Ctrl+V` interception, but this affects native paste behavior and is not part of the current scope.

The documentation and code should continue to evolve around the already implemented local image/single-file path paste capability.
