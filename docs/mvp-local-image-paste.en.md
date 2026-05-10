# MVP: Local CLI/Agent Smart Paste

## Purpose

The first clipterm milestone is to make clipboard images, single copied files, and normal text flow naturally into local terminals, CLIs, and AI agent workflows.

A possible future direction is that after copying a screenshot or another clipboard image, users still press the familiar `Cmd+V` / `Ctrl+V`, and clipterm decides what the current paste target should receive:

- If the target app can natively accept image paste, such as ChatGPT web, chat apps, or document editors, clipterm should preserve the native image paste experience.
- If the target app cannot use image data effectively, such as terminals, browser address bars, Notepad, or plain text fields, clipterm saves the image as a local file and pastes the absolute file path.

The current macOS prototype starts with a safer shortcut: `Cmd+Shift+V` means CLI/agent smart paste. Images and single copied files become absolute paths; normal text and ordinary clipboard content fall back to native `Cmd+V`. Normal `Cmd+V` is not intercepted yet, which avoids breaking native system paste behavior. Fully replacing normal paste is only a future exploration direction, not a first-stage commitment.

The first milestone should feel invisible: users should not need to save screenshots manually, open a file browser, find the file, and copy its path just to use an image in a terminal, CLI, or AI agent.

## Unified Product Concept

clipterm is not primarily a screenshot tool or a general clipboard manager. Its core concept is:

```text
clipboard object -> materialized file -> usable path in the current environment
```

Local smart paste and future remote transfer are the same product line:

```text
local image paste:
clipboard image -> local file -> local absolute path paste

remote image/file transfer:
clipboard image/file -> remote file -> remote absolute path paste
```

Therefore this should remain one project. The implementation can be modular, but the product and CLI entry point should stay unified.

Recommended tagline:

```text
Materialize clipboard objects for terminals, CLIs, and AI agents.
```

## Product Definition

clipterm is a system-level smart paste tool.

When the user triggers clipterm smart paste and the clipboard contains an image or a single copied file, clipterm generates or reads a local absolute path and pastes that path into the focused input. If the clipboard contains normal text or ordinary content, clipterm does not rewrite the clipboard and sends one native `Cmd+V`.

If normal `Cmd+V` interception is later proven safe enough, clipterm can use application rules to choose one of two behaviors:

1. Allow native image paste.
2. Save the image to a local cache directory, then paste the absolute path as text into the current focused input.

Example result:

```text
/Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

Windows example:

```text
C:\Users\Alice\AppData\Local\clipterm\cache\clipterm-20260509-153012-a83f.png
```

## First-Stage Scope

The first stage implements local CLI/agent smart paste.

Included:

- Read image data from the system clipboard.
- Read the path of a single file copied in Finder.
- Save clipboard images as real local files.
- Use `Cmd+Shift+V` to paste local absolute paths into the focused input.
- Fall back to native `Cmd+V` for normal text and ordinary clipboard content without rewriting the clipboard.

Excluded:

- Normal `Cmd+V` interception.
- App-specific native image paste allow rules.
- SSH or remote materialization.
- WSL path bridging.
- tmux-aware transport.
- File upload or download.
- Directory or multi-file transfer.
- Large file transfer.

## Target Users

- Developers who need to reference screenshots in terminal commands.
- Users pasting image paths into AI coding CLIs.
- Users working with ChatGPT, Codex, Claude Code, Cursor, VS Code, JetBrains, Terminal.app, iTerm2, Windows Terminal, and similar tools.
- Users taking screenshots with WeChat, QQ, macOS screenshots, Windows Snipping Tool, or similar utilities.
- Users who do not want to manually save screenshots, locate files, and copy paths.

## User Experience

### Path Paste Flow

```text
User copies screenshot image
        |
User focuses a terminal, Notepad, browser address bar, or another text input
        |
User presses Cmd+Shift+V
        |
clipterm detects image clipboard content
        |
clipterm decides the target should receive a path
        |
clipterm saves image to local cache
        |
clipterm writes the file path to clipboard as text
        |
clipterm triggers normal paste
        |
Current app receives the absolute image file path
```

Example:

```bash
python inspect_image.py /Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

AI CLI example:

```text
Please inspect this screenshot: /Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

### Text Fallback Flow

```text
User copies normal text
        |
User focuses a terminal, AI CLI, or another text input
        |
User presses Cmd+Shift+V
        |
clipterm does not find an image or single file that should become a path
        |
clipterm does not modify the clipboard
        |
clipterm sends one native Cmd+V
        |
The current app receives the system-native paste content
```

### Native Image Paste Flow

```text
User copies screenshot image
        |
User focuses an app that supports image paste
        |
User presses Cmd+V / Ctrl+V
        |
clipterm decides the target should receive native paste
        |
clipterm allows the system paste to continue
        |
The target app receives the image directly
```

For example, if ChatGPT web can accept an image, clipterm should preserve that upload flow.

## Smart Paste Strategy

The current prototype uses a low-risk strategy: do not intercept normal `Cmd+V` / `Ctrl+V`; provide `Cmd+Shift+V` as CLI/agent smart paste.

Future versions may explore more advanced normal `Cmd+V` / `Ctrl+V` interception, but that depends on stability and negative-impact assessment.

A system-level program cannot always know whether a specific input field accepts images. In the same browser, a ChatGPT message box may accept images while the address bar only accepts text. Therefore, if future versions explore normal `Cmd+V` smart interception, they will need app rules and conservative fallback behavior. The current first stage does not do that yet.

### Rule Types

- `native`: allow native paste. Suitable for ChatGPT web, WeChat, Slack, document editors, and other image-aware targets.
- `path`: convert images, screenshots, and single copied files into path paste.
- `native-text`: keep normal text and ordinary clipboard content native by sending paste without rewriting the clipboard.
- `auto`: default strategy. Decide based on app identity, window title, focused context, and future detection capabilities.

### Conservative Principles

- Convert to path for known path-paste targets.
- Allow native paste for known native image targets.
- Preserve native paste for normal text without reading or rewriting the text content.
- If the target is unknown or risky, prefer not breaking native paste.
- The current prototype already provides `Cmd+Shift+V` as CLI/agent smart paste. If future versions intercept normal `Cmd+V`, this shortcut can remain the stable fallback.

## Initial Platform Scope

### macOS

First-class target for MVP.

- Normal paste shortcut: `Cmd+V`, not intercepted by the current prototype.
- CLI/agent smart paste shortcut: `Cmd+Shift+V`
- Cache directory: `~/Library/Caches/clipterm/`
- Important permissions: Accessibility, clipboard read/write, synthetic keyboard events.

Important path-paste targets:

- Terminal.app
- iTerm2
- WezTerm
- Ghostty
- Alacritty
- VS Code / Cursor / JetBrains terminals
- TextEdit plain text mode
- Browser address bars
- Plain text inputs

Important native-paste targets:

- ChatGPT web image input contexts
- WeChat, Slack, and similar chat apps
- Document editors that support image paste

### Windows

Second first-class target after the macOS proof of concept. The Windows version is a multi-platform implementation of the same product semantics; it does not change the `Cmd+Shift+V` / `Ctrl+Shift+V` smart paste model.

Scope:

- Support Windows x86_64 only.
- Do not support Windows ARM.
- Use `Ctrl+Shift+V` as the fixed CLI/agent smart paste shortcut.
- Match the current macOS `Cmd+Shift+V` behavior.

- Normal paste shortcut: `Ctrl+V`
- CLI/agent smart paste shortcut: `Ctrl+Shift+V`
- Cache directory: `%LOCALAPPDATA%\clipterm\cache\`
- Important permissions: clipboard read/write, keyboard hook, foreground window detection, synthetic keyboard events.

Windows `Ctrl+Shift+V` behavior:

- Screenshot or image stream: save PNG and paste the absolute path.
- Single copied file: paste the original absolute file path.
- Normal text and ordinary clipboard content: fall back to native `Ctrl+V` without modifying the clipboard.
- Multiple files: not supported yet; do not modify the clipboard.

Main platform-specific additions:

- Windows clipboard: read images, read single copied file paths, write Unicode text.
- Windows global hotkey: register `Ctrl+Shift+V`.
- Windows synthetic paste: send native `Ctrl+V`.
- Windows release package: publish x86_64 only.

Important path-paste targets:

- Windows Terminal
- PowerShell
- cmd
- VS Code / Cursor / JetBrains terminals
- Notepad
- Browser address bars
- Plain text inputs

### Windows Development Handoff

The first instruction for a future developer or AI agent working in Windows can be:

```text
Please understand this project first, especially the design documents under docs. The current goal is to complete the native Windows x86_64 implementation, matching the already completed and released macOS version.
```

When taking over the Windows version, do not redefine the product. Inherit the user model and engineering boundaries already validated by macOS v0.1.0:

- `Ctrl+Shift+V` is the Windows CLI/agent smart paste shortcut, corresponding to macOS `Cmd+Shift+V`.
- Do not intercept normal `Ctrl+V`. Fully replacing normal paste remains only a future exploration direction, not the current Windows goal.
- Image, screenshot, single-file, and normal-text behavior must match the current macOS version.
- Multiple files remain unsupported; do not expand scope for the first Windows release.
- Windows ARM is not supported; release packages target Windows x86_64 only.
- The first Windows release is platform-adapter work, not product redesign.

Existing code and design to reuse:

- `internal/clipterm`: core smart paste semantics.
- `internal/materialize`: cache directory, image materialization, cleanup policy.
- `internal/daemon`: background process, PID file, idempotent startup, and stale PID handling.
- `internal/cli`: command entry points and `daemon` / `doctor` / `rules` / `clean` shape.
- The docs around clipboard objects, path materialization, and remote capability boundaries.

Windows-specific platform layers to add or replace:

- `internal/clipboard/clipboard_windows.go`
- `internal/hotkey/hotkey_windows.go`
- `internal/paste/paste_windows.go`
- Windows x86_64 release packaging target.

Recommended implementation order:

1. Implement the `Ctrl+Shift+V` hotkey and native `Ctrl+V` fallback first, then verify lossless normal text paste.
2. Implement `CF_HDROP` single-file path reading, then verify Explorer single-file copy into terminals, Notepad, and browser address bars.
3. Implement image clipboard reading and PNG materialization. Windows clipboard images are commonly DIB/DIBV5, so this is the riskiest part and should be validated separately.
4. Finish `doctor` capability checks, README installation notes, and the Windows x86_64 release package.

Minimum acceptance criteria for the Windows version:

- Copy normal text, press `Ctrl+Shift+V`, and get native `Ctrl+V` behavior without modifying the clipboard.
- Copy one file in Explorer, press `Ctrl+Shift+V`, and paste the original absolute file path.
- Copy a screenshot or image, press `Ctrl+Shift+V`, and paste the generated PNG absolute path.
- Multiple files are ignored without polluting the clipboard.
- Starting the daemon repeatedly does not create multiple background processes.
- The Windows x86_64 package name is user-facing and clear; do not publish a Windows ARM package.

### Windows + WSL Development Notes

When using Codex through WSL on a Windows machine, separate the development environment from the runtime environment under test:

- WSL is a Linux userland and Linux kernel environment, not a Windows process environment.
- A Go program running inside WSL cannot directly register Windows global hotkeys, use the Windows clipboard APIs, or call `SendInput` as a normal Windows GUI process.
- The Windows clipterm build must run as a native Windows `.exe` to listen for `Ctrl+Shift+V`, read the Windows clipboard, and send paste events to the focused Windows application.
- WSL is still useful for editing code, running ordinary Go tests, and testing code paths that do not depend on Win32 GUI integration.
- Windows platform integration tests need to run in the native Windows environment, such as PowerShell, cmd, Windows Terminal, or by invoking the Windows exe and observing Windows-side behavior.

Recommended workflow:

1. Use WSL/Codex for code editing and most non-GUI logic development.
2. Build or run `clipterm.exe` in the native Windows environment.
3. Test `Ctrl+Shift+V` in real Windows focus targets such as Windows Terminal, PowerShell, cmd, Notepad, and browser address bars.

WSL can later become a supported target environment, for example by materializing Windows clipboard objects into WSL paths. That is WSL path bridging and is separate from the first native Windows implementation.

### Linux Desktop

Deferred. Linux clipboard handling differs across X11, Wayland, desktop environments, and terminal emulators, so it should not block the initial MVP.

## Command Shape

The project should expose both direct CLI commands and a background daemon.

```bash
clipterm paste
clipterm paste --copy-path
clipterm daemon
clipterm rules
clipterm clean
clipterm doctor
```

### `clipterm paste`

Debuggable core primitive.

Behavior:

1. Read image data from the system clipboard.
2. Save it to the clipterm cache directory.
3. Print the absolute file path to stdout.

### `clipterm paste --copy-path`

Same as `clipterm paste`, but also writes the generated path back to the system clipboard as text.

This is an important intermediate step before system-level smart paste.

### `clipterm daemon`

Background process that provides the current macOS prototype interaction.

Behavior:

1. Listen for `Cmd+Shift+V`.
2. First check whether the clipboard contains a single copied file.
3. If it does, copy and paste the original absolute file path.
4. Otherwise check whether the clipboard contains an image stream.
5. If it does, save the image as a cached PNG, then copy and paste the generated absolute path.
6. If neither applies, send one native `Cmd+V` without modifying the clipboard.

### Daemon Startup Recommendation

For the current prototype, users should start the daemon from a shell startup file, such as `~/.zshrc` or `~/.bash_profile`.

```bash
command -v clipterm >/dev/null 2>&1 && clipterm daemon >/dev/null 2>&1 || true
```

If `clipterm` is not on `PATH`, use the installed absolute path:

```bash
$HOME/.local/bin/clipterm daemon >/dev/null 2>&1 || true
```

This is a good prototype-stage default:

- `clipterm daemon` is idempotent; if it is already running, it does not start another background process.
- If the daemon is killed manually, the next `clipterm daemon` run detects the stale PID file and starts a fresh background process.
- Opening a new terminal automatically attempts startup, and the command returns immediately without occupying the terminal window.
- It avoids launchd code and installer complexity, and users can remove it by deleting one shell config line.

The more formal macOS approach is a LaunchAgent that starts at login. The current stage should not implement that yet. It can be added later as an optional installer or release-package feature after the project stabilizes.

### `clipterm rules`

Show the current paste strategy.

The current prototype does not manage application-level rules yet. If future versions explore normal `Cmd+V` smart interception, built-in rules, config files, or CLI management can be revisited.

### `clipterm clean`

Remove old cache files.

Default cleanup policy can be conservative:

- Keep files from the last 7 days.
- Keep total cache size under a future configurable limit.

### `clipterm doctor`

Check platform integration status.

Useful checks:

- Clipboard read access.
- Clipboard write access.
- Cache directory write access.
- Global shortcut or keyboard hook permission.
- macOS Accessibility permission.
- Foreground app detection.
- Synthetic paste event.
- Current paste strategy.

## Cache Strategy

Use a user-level cache directory instead of a generic temporary directory.

Recommended directories:

```text
macOS:   ~/Library/Caches/clipterm/
Windows: %LOCALAPPDATA%\clipterm\cache\
```

Recommended filename format:

```text
clipterm-YYYYMMDD-HHMMSS-<short-random>.png
```

Example:

```text
clipterm-20260509-153012-a83f.png
```

The MVP should save clipboard images as PNG. Other formats can be added later if preserving the original source format becomes important.

## Native macOS Paste vs clipterm Paste

macOS copy and paste is not just a fixed text string. It places a clipboard object on the system pasteboard. A Finder file object can expose several representations at the same time, such as a file URL, display name, icon, preview, and type metadata. The target app chooses which representation it can consume.

This is why the same Finder file copied with `Cmd+C` can behave differently:

- Pasting into a browser address bar may produce the file name.
- Pasting into iTerm2 may produce the absolute path.
- Pasting into an app that supports file objects may import or upload the file directly.

The current clipterm prototype does not intercept normal `Cmd+V`. The current `Cmd+Shift+V` is CLI/agent smart paste: convert path-capable clipboard objects into local absolute paths, and preserve native paste for normal text or ordinary content.

| Clipboard content | Native macOS `Cmd+V` | clipterm `Cmd+Shift+V` |
| --- | --- | --- |
| Screenshot or image stream | Decided by the target app. Image-aware apps may accept the image; plain text fields may do nothing or read another representation. | Save the image as a PNG under the clipterm cache directory, copy the generated absolute path, and paste that path. |
| Single file copied in Finder | Decided by the target app. It may paste a file name, paste an absolute path, or consume the file object directly. | Paste the original absolute file path. The file content is not copied. |
| Multiple files copied in Finder | Decided by the target app. | Not supported yet. clipterm refuses it and does not modify the clipboard. |
| Plain text and ordinary clipboard content | Paste the content normally. | Send one native `Cmd+V` without modifying the clipboard, preserving system-native text and rich paste behavior. |

After a successful image or file path paste, the system clipboard contains path text. Users can press normal `Cmd+V` repeatedly to paste the same path until they copy something else. Text fallback does not modify the clipboard.

## Clipboard Restoration

For path paste, the daemon needs to temporarily replace the clipboard image with text so the target application can receive the generated path through normal paste.

The first version should use simple mode:

1. Save the image.
2. Write the path to the clipboard.
3. Trigger paste.
4. Do not restore the original image yet.

Later versions can add polished mode:

1. Save the original clipboard image.
2. Replace the clipboard with path text.
3. Trigger paste.
4. Restore the original image after `300ms` to `800ms`.

Polished mode better matches user expectations, but it introduces platform complexity and race conditions. If the user copies something else during the restore window, clipterm must not overwrite the user's new clipboard.

## Remote Feasibility Conclusion

Remote image paste and remote single-file transfer are feasible, but they are not part of the first stage.

Future remote capability can be designed as:

```text
local clipboard image/file
 -> stream/chunk encode
 -> terminal-safe text channel
 -> remote receive helper
 -> remote materialized file
 -> remote absolute path paste
```

Key conclusions:

- A default 10MB single-file limit is a reasonable promise.
- Base64 supports streaming encode/decode, so clipterm can read, encode, and send incrementally instead of buffering the whole file.
- Terminal text transport still needs chunking, throttling, checksums, and failure handling.
- With a remote helper installed, the experience can be nearly invisible. Without a helper, shell bootstrap fallback cannot guarantee invisibility.
- This can become a clipboard-scale rz/sz-style upload/download fallback, but not a large-file or high-throughput replacement for scp/rsync.

The remote direction should stay in the roadmap and architecture boundary for now. This MVP should not define the full remote protocol yet.

## Project Boundary

Do not split the project.

Reasons:

- Local smart paste and remote materialization share the same core model.
- The user mental model is consistent: clipboard objects become usable paths in the current environment.
- The implementation can be modular while the product keeps one name and one CLI.

Suggested internal modules:

```text
clipboard/
materialize/
paste/
rules/
platform/
transport/
remote/
```

Future commands can include:

```bash
clipterm send
clipterm receive
```

The first stage still only implements local CLI/agent smart paste.

## Technical Stack Direction

The first implementation language is Go.

Reasons:

- The current maintainer is comfortable with Go, which improves long-term iteration speed.
- Go can ship as a single native binary.
- The Go standard library is strong for file IO, streaming Base64, chunking, checksums, and future remote transfer.
- macOS / Windows platform integration can be layered through platform APIs, cgo, syscall, or external-command fallbacks.
- The first-stage difficulty is system integration and product experience, not language-level performance limits.

## Suggested Go Dependencies and Interfaces

These are candidates, not final commitments:

- `cobra` or the standard library `flag` for CLI parsing.
- Go standard library `os.UserCacheDir` for platform cache directories.
- Go standard library `image/png` for PNG encoding.
- Go standard library `time` for timestamped filenames.
- Go standard library `crypto/rand` or `math/rand/v2` for short filename suffixes.
- Go standard library `encoding/base64` for future streaming Base64 encode/decode.
- Go standard library `crypto/sha256` for future remote transfer verification.
- Platform-specific clipboard APIs for reliable image handling.
- Platform-specific keyboard and foreground-window APIs for daemon behavior.
- Platform-specific application identity APIs for app-level paste rules.

Clipboard access, app rules, paste events, and platform integration should be abstracted behind internal interfaces so macOS and Windows can evolve independently.

## Open Questions

- In the native Windows implementation, should single-file path paste and text fallback ship before image DIB decoding?
- Should the Windows version recommend startup through something beyond shell startup, such as the Startup folder or Task Scheduler?
- If future versions explore normal `Cmd+V` / `Ctrl+V` interception, how should clipterm minimize user impact?
- Should the current `Cmd+Shift+V` smart paste shortcut be configurable?
- Should clipboard restoration be enabled by default, opt-in, or deferred entirely?
- Should generated paths be shell-escaped when pasted, or pasted as raw absolute paths?
- Should fine-grained browser input detection be deferred to a later version?

Recommended initial answer for path escaping: avoid spaces in generated file paths and paste raw absolute paths. Shell-specific escaping can be added later.
