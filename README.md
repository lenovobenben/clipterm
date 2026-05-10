# clipterm

Copy screenshot. Paste path anywhere.

## Overview

clipterm is a cross-platform tool that brings modern clipboard object semantics to terminals, CLIs, AI agents, and text-focused workflows.

The current macOS prototype uses `Cmd+Shift+V` as a CLI/agent-friendly smart paste shortcut: clipboard images and single copied files become absolute paths, while normal text falls back to native paste. Fully replacing normal `Cmd+V` / `Ctrl+V` is only a possible future direction, not a committed first-stage goal.

Instead of treating paste as raw text only, clipterm materializes clipboard objects as real files inside the target environment and pastes the resulting absolute file path when that is more useful than native paste.

The goal is to make local shells, WSL, SSH sessions, tmux environments, and AI coding CLIs feel as seamless as desktop environments or virtual machines with shared clipboard support.

---

# Current Prototype

The current macOS prototype supports:

```bash
clipterm paste
clipterm paste --copy-path
clipterm paste --copy-path --send-paste
clipterm daemon
clipterm doctor --request-permissions
clipterm rules
clipterm version
```

On macOS, `clipterm daemon` listens for `Cmd+Shift+V`. When the clipboard contains an image, it saves the image to the user cache directory, copies the generated path, and sends a normal paste event to the focused app. When the clipboard contains a single copied file, it pastes the original absolute file path. For normal text and other ordinary clipboard content, it sends native `Cmd+V` without modifying the clipboard.

## macOS Paste Behavior

macOS copy and paste is object-based, not just text-based. When Finder copies a file, the system pasteboard can contain several representations of the same object, such as a file URL, a display name, and preview metadata. Each target app decides which representation to consume. This is why pasting the same copied file into a browser address bar may produce a file name, while pasting into iTerm2 may produce an absolute path.

clipterm's current `Cmd+Shift+V` behavior is CLI/agent smart paste: convert clipboard objects into local absolute paths when that is useful, and otherwise fall back to native paste.

| Clipboard content | Native macOS `Cmd+V` | clipterm `Cmd+Shift+V` |
| --- | --- | --- |
| Screenshot or image stream | Depends on the target app. Image-aware apps may accept the image; plain text fields may do nothing or use another representation. | Save the image as a PNG under the clipterm cache directory, copy the absolute path, then paste the path. |
| Single copied file in Finder | Depends on the target app. Some apps paste a file name, some paste a file path, and some consume the file object directly. | Paste the original absolute file path. The file is not copied. |
| Multiple copied files | Depends on the target app. | Not supported yet. clipterm refuses it and does not modify the clipboard. |
| Plain text and ordinary clipboard content | Paste the content normally. | Send native `Cmd+V` without modifying the clipboard. This preserves normal text and rich paste behavior. |

After a successful image or file path paste, the system clipboard contains the path text. You can press normal `Cmd+V` repeatedly to paste that same path until you copy something else. The prototype does not yet restore the original image or file clipboard object after path paste. Text fallback does not modify the clipboard.

Required macOS permissions:

- Input Monitoring / keyboard event access for the daemon hotkey.
- Accessibility for synthetic paste.

Run:

```bash
clipterm doctor --request-permissions
```

to request the required permissions.

## Local Development

Build and test:

```bash
make build
make test
```

Install to `~/.local/bin/clipterm`:

```bash
make install
```

Make sure `~/.local/bin` is on your `PATH`.

Run the daemon:

```bash
clipterm doctor --request-permissions
clipterm daemon
```

`clipterm daemon` starts a background process and returns immediately. Then copy a screenshot, a file in Finder, or normal text and press `Cmd+Shift+V` in any focused text input.

For day-to-day use during the prototype stage, start the daemon from your shell startup file. `clipterm daemon` is idempotent: if the daemon is already running, it returns without starting another copy.

For zsh:

```bash
printf '\n# Start clipterm background daemon\ncommand -v clipterm >/dev/null 2>&1 && clipterm daemon >/dev/null 2>&1 || true\n' >> ~/.zshrc
```

For bash:

```bash
printf '\n# Start clipterm background daemon\ncommand -v clipterm >/dev/null 2>&1 && clipterm daemon >/dev/null 2>&1 || true\n' >> ~/.bash_profile
```

If `clipterm` is not on your `PATH`, use the installed absolute path instead:

```bash
$HOME/.local/bin/clipterm daemon >/dev/null 2>&1 || true
```

This starts clipterm when a new shell opens. A future macOS installer may use a LaunchAgent for login startup, but the prototype intentionally keeps startup explicit and easy to remove.

If the daemon process is killed manually, the next `clipterm daemon` run will detect the stale PID file and start a fresh background process.

Check or stop the background daemon:

```bash
clipterm daemon --status
clipterm daemon --stop
```

Show the current paste strategy:

```bash
clipterm rules
```

Clean generated screenshot files:

```bash
clipterm clean --dry-run
clipterm clean --days 7
```

`clipterm clean` only removes managed cache images matching `clipterm-*.png` under the clipterm cache directory.

Run in the foreground for debugging:

```bash
clipterm daemon --foreground --debug-hotkeys
```

Daemon logs are written to:

```text
~/Library/Logs/clipterm/daemon.log
~/Library/Logs/clipterm/daemon.err.log
```

---

# Problem Statement

Terminal environments still operate mostly on plain text semantics.

Modern workflows frequently involve:

- screenshots
- images from clipboard streams
- small files
- logs
- PDFs
- patches
- AI coding assistants

Existing solutions have major limitations:

- clipboard screenshots are not real files
- SSH/scp/rz/sz may not work in restricted environments
- tunnels are often disabled
- WSL path translation is inconsistent
- remote terminals cannot access local filesystem paths
- AI coding CLIs need actual file paths

Users constantly switch between GUI clipboard workflows and terminal workflows.

clipterm aims to bridge this gap.

---

# Core Concept

Clipboard object → terminal-safe transfer → real file in target environment → absolute path paste

The current environment should receive a usable file path to a real file it can access, regardless of where the clipboard data originated.

Examples:

- clipboard screenshot → `/home/user/.clipterm/shot.png`
- copied local file → `/tmp/.clipterm/demo.zip`
- WSL paste → `/mnt/c/...`
- remote SSH paste → remote materialized file

---

# Goals

## Main Goals

- Make paste work naturally for clipboard objects in terminals, CLIs, AI agents, and text-focused workflows
- Support screenshots and copied files
- Support local shells, WSL, SSH, tmux, and AI coding CLIs
- Work even in restricted environments where scp/rz/sz/tunnels fail
- Require only terminal text transport
- Materialize real files in the target environment
- Paste usable absolute file paths
- Provide a lightweight rz/sz-style fallback for clipboard-scale image and file transfer

## Non-Goals

- Not a replacement for scp/rsync
- Not optimized for large file transfer
- Not a general cloud sync tool
- Not a terminal emulator

---

# Initial Scope

## Supported Clipboard Types

### Clipboard image streams

Examples:

- WeChat screenshots
- QQ screenshots
- macOS screenshots
- system snipping tools

### Copied files

Examples:

- Finder copied files
- Explorer copied files

---

# Initial Target Environments

- Local shell
- WSL
- SSH sessions
- tmux

---

# Transfer Model

clipterm transfers data using terminal-safe text streams.

For remote environments:

1. Read clipboard object
2. Convert to real file
3. Compress/resize if needed
4. Encode as Base64
5. Send through terminal text channel
6. Decode remotely
7. Materialize file in target environment
8. Paste resulting absolute path

This allows operation in environments where traditional binary transfer methods fail.

---

# Roadmap

## Phase 1: Local CLI/Agent Smart Paste

Copy a screenshot, clipboard image, single file, or text, then press `Cmd+Shift+V` on macOS. Images and files paste as local absolute paths; text falls back to native paste without changing the clipboard. Future versions may explore normal `Ctrl+V` / `Cmd+V` interception after the risk is better understood.

## Phase 2: Remote Image Paste

Copy a local screenshot or clipboard image, paste into an SSH, tmux, or similar remote terminal session, transfer the image through the terminal text channel, materialize it on the remote host, and paste the remote absolute path.

This is the first remote materialization target.

## Phase 3: Clipboard File Paste

Copy one or more local files from Finder, Explorer, or the desktop environment, paste into a terminal session, transfer the file data, materialize it in the target environment, and paste the resulting path.

For remote sessions, this becomes a lightweight rz/sz-style upload flow that works even when scp, rz/sz, tunnels, or shared folders are unavailable.

## Phase 4: Reverse Transfer

Support remote-to-local transfer for clipboard-scale files, allowing users to bring generated files, logs, patches, screenshots, and artifacts back from a terminal environment to the local machine.

This is the download side of the rz/sz-style workflow.

---

# File Size Philosophy

clipterm is optimized for clipboard-scale objects.

Recommended limits:

- ideal: under 2 MB
- supported: under 10 MB
- larger files may require chunked transfer mode

Large files are intentionally out of scope.

---

# Example UX

## Screenshot to Remote AI CLI

1. Take screenshot with WeChat
2. Copy screenshot
3. Focus remote Codex terminal
4. Press `Ctrl+V`
5. clipterm transfers image
6. Remote file is created:
   `/home/user/.clipterm/shot.png`
7. Absolute path is pasted automatically

---

# Technical Direction

## Language

Go

Reasons:

- cross-platform
- fast startup
- native binaries
- strong filesystem support
- strong standard library for streaming, encoding, checksums, and file IO
- good CLI ecosystem
- easier long-term maintenance for the current project owner

---

# Development Model

This project is intended to be developed 100% with AI agents.

The human maintainer focuses on product design, direction, review, testing, and release decisions. Implementation work is delegated to AI coding agents, making clipterm both a utility project and a practical showcase for AI-agent-driven software development.

---

# Planned Architecture

## Components

### clipterm daemon

Clipboard monitoring and integration layer.

### clipterm CLI

Transfer, materialization, and terminal integration.

### platform adapters

- Windows clipboard
- macOS clipboard
- WSL integration
- SSH integration

### transport layer

- Base64 streaming
- chunked transfer
- terminal-safe protocols

---

# Future Ideas

- reverse transfer (remote → local clipboard)
- OSC52 integration
- automatic terminal environment detection
- shell plugins
- tmux integration
- AI CLI integrations
- transfer progress UI
- encrypted transfer mode
- multi-file packaging
- directory transfer

---

# Design Philosophy

clipterm is not just a screenshot tool.

It aims to modernize clipboard semantics inside terminal and CLI-centered workflows.

The ultimate goal is to make terminal workflows feel as seamless as desktop clipboard workflows.

---

# License

MIT License. See [LICENSE](LICENSE).
