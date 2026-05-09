# MVP: Local Image Smart Paste

## Purpose

The first clipterm milestone is to make clipboard images flow naturally into local apps, terminals, CLIs, and AI agent workflows.

After copying a screenshot or another clipboard image, users still press the familiar `Cmd+V` / `Ctrl+V`. clipterm decides what the current paste target should receive:

- If the target app can natively accept image paste, such as ChatGPT web, chat apps, or document editors, clipterm should preserve the native image paste experience.
- If the target app cannot use image data effectively, such as terminals, browser address bars, Notepad, or plain text fields, clipterm saves the image as a local file and pastes the absolute file path.

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

When the user triggers paste and the clipboard contains an image, clipterm uses the current application rules to choose one of two behaviors:

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

The first stage implements local image smart paste only.

Included:

- Read image data from the system clipboard.
- Save clipboard images as real local files.
- Paste local absolute paths into path-paste targets.
- Preserve native image paste where appropriate.
- Preserve normal text paste behavior when the clipboard does not contain an image.

Excluded:

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
User presses Cmd+V / Ctrl+V
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

clipterm uses the advanced strategy: intercept normal `Cmd+V` / `Ctrl+V`, with the goal of making the behavior invisible to users.

However, a system-level program cannot always know whether a specific input field accepts images. In the same browser, a ChatGPT message box may accept images while the address bar only accepts text. The first version needs app rules and conservative fallback behavior.

### Rule Types

- `native`: allow native paste. Suitable for ChatGPT web, WeChat, Slack, document editors, and other image-aware targets.
- `path`: convert the image into a path paste. Suitable for terminals, AI CLIs, Notepad, browser address bars, and plain text fields.
- `auto`: default strategy. Decide based on app identity, window title, focused context, and future detection capabilities.

### Conservative Principles

- Convert to path for known path-paste targets.
- Allow native paste for known native image targets.
- If the target is unknown or risky, prefer not breaking native paste.
- Later versions can provide a force-path-paste fallback shortcut, such as `Cmd+Shift+V` / `Ctrl+Shift+V`.

## Initial Platform Scope

### macOS

First-class target for MVP.

- Normal paste shortcut: `Cmd+V`
- Optional force path paste shortcut: `Cmd+Shift+V`
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

Second first-class target after the macOS proof of concept.

- Normal paste shortcut: `Ctrl+V`
- Optional force path paste shortcut: `Ctrl+Shift+V`
- Cache directory: `%LOCALAPPDATA%\clipterm\cache\`
- Important permissions: clipboard read/write, keyboard hook, foreground window detection, synthetic keyboard events.

Important path-paste targets:

- Windows Terminal
- PowerShell
- cmd
- VS Code / Cursor / JetBrains terminals
- Notepad
- Browser address bars
- Plain text inputs

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

Background process that provides the final MVP user experience.

Behavior:

1. Listen for the platform normal paste shortcut.
2. Check whether the clipboard contains an image.
3. Identify the foreground application and available context.
4. Use rules to allow native paste or perform path paste.
5. Preserve normal paste behavior for non-image clipboards.

### `clipterm rules`

Manage application-level paste rules.

The first version can ship with built-in rules. Later versions can expose a config file or CLI management.

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
- Current app rule match result.

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

The first stage still only implements local image smart paste.

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

- Should the first implementation target macOS only, or macOS and Windows together?
- Which `native` / `path` app rules should ship in the first version?
- How should clipterm minimize user impact when intercepting normal `Cmd+V` / `Ctrl+V` fails?
- Should clipterm provide a force-path-paste fallback shortcut?
- Should clipboard restoration be enabled by default, opt-in, or deferred entirely?
- Should generated paths be shell-escaped when pasted, or pasted as raw absolute paths?
- Should fine-grained browser input detection be deferred to a later version?

Recommended initial answer for path escaping: avoid spaces in generated file paths and paste raw absolute paths. Shell-specific escaping can be added later.
