# clipterm

Copy screenshot. Paste path anywhere.

clipterm solves a small but common workflow problem: screenshots and clipboard images usually have to be saved manually before a terminal, CLI, AI agent, or text input can reference them. clipterm saves the clipboard image as a real local file and pastes the file path for you.

The same smart paste shortcut also handles a single copied file by pasting its absolute path. Plain text still uses native paste.

## Current Status

The current prototype provides a dedicated smart paste shortcut:

| Platform | Smart paste shortcut | Normal paste |
| --- | --- | --- |
| macOS | `Cmd+Shift+V` | `Cmd+V` is not intercepted |
| Windows x86_64 | `Ctrl+Shift+V` | `Ctrl+V` is not intercepted |

Smart paste behavior:

| Clipboard content | Behavior |
| --- | --- |
| Plain text and ordinary clipboard content | Send native paste without modifying the clipboard. |
| Single copied file | Paste the original absolute file path. |
| Screenshot or image stream | Save a PNG under the clipterm cache directory, then paste the generated path. |
| Multiple copied files | Not supported yet; clipterm refuses it and does not modify the clipboard. |

After a successful file or image path paste, the system clipboard contains that path text. Pressing normal paste again repeats the same path until the user copies something else.

## Windows Notes

Windows support runs as a native Windows `.exe`. This is required for `RegisterHotKey`, Windows clipboard access, and `SendInput`.

Windows image support currently reads common `CF_DIB` / `CF_DIBV5` clipboard images and encodes them as PNG. The implementation supports the common 24-bit and 32-bit uncompressed DIB cases used by screenshots and many clipboard image sources. More exotic palette, compressed, or color-managed bitmap variants can be added later.

Windows path output is selected when the daemon starts:

```powershell
clipterm.exe daemon --path-style windows
clipterm.exe daemon --path-style wsl
```

Use `windows` for native Windows shells such as PowerShell and `wsl` when the target expects paths like `/mnt/c/Users/Alice/...`. Changing path style requires restarting the daemon.

Known limitation: Notepad++ is not supported by the first Windows prototype. clipterm successfully writes the path text to the system clipboard, but Notepad++ handles the synthetic paste event during `Ctrl+Shift+V` differently from Notepad, browser address bars, PowerShell, Windows Terminal, and WSL terminals. This is a target-application compatibility issue, not a path conversion or clipboard write failure.

## macOS Notes

On macOS, `clipterm daemon` listens for `Cmd+Shift+V`.

Required permissions:

- Input Monitoring / keyboard event access for the daemon hotkey.
- Accessibility for synthetic paste.

Request permissions with:

```bash
clipterm doctor --request-permissions
```

## Commands

```bash
clipterm paste
clipterm paste --copy-path
clipterm paste --copy-path --send-paste
clipterm daemon
clipterm daemon --foreground --debug-hotkeys
clipterm daemon --status
clipterm daemon --stop
clipterm doctor --request-permissions
clipterm rules
clipterm clean
clipterm version
```

Useful Windows examples:

```powershell
clipterm.exe daemon --foreground --path-style wsl --debug-hotkeys
clipterm.exe daemon --path-style windows
clipterm.exe daemon --stop
```

Useful macOS examples:

```bash
clipterm doctor --request-permissions
clipterm daemon
clipterm daemon --foreground --debug-hotkeys
```

## Cache

Generated image files are stored in the user cache directory:

```text
macOS:   ~/Library/Caches/clipterm/
Windows: %LOCALAPPDATA%\clipterm\cache\
```

Clean old generated images:

```bash
clipterm clean --dry-run
clipterm clean --days 7
```

`clipterm clean` only removes managed cache images named `clipterm-*.png`.

## Development

Build and test:

```bash
make build
make test
```

Install to `~/.local/bin/clipterm`:

```bash
make install
```

Windows cross-build from WSL/Linux:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/clipterm-windows-amd64/clipterm.exe ./cmd/clipterm
```

Validation commands used for this prototype:

```bash
go test ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/clipterm
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/clipterm
go vet ./...
```

More design detail lives in:

- `docs/mvp-local-image-paste.zh.md`
- `docs/mvp-local-image-paste.en.md`

## Later Directions

These are not commitments for the current prototype:

- Harden Windows image format coverage beyond common DIB/DIBV5 screenshots.
- Improve installation/startup guidance for Windows.
- Explore optional clipboard restoration after path paste.
- Keep normal `Cmd+V` / `Ctrl+V` interception as a future research direction, not a current goal.

## License

MIT License. See [LICENSE](LICENSE).
