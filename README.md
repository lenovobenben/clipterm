# clipterm

Copy screenshot. Paste path anywhere.

## Overview

clipterm is a cross-platform tool that brings modern clipboard object semantics to terminals, CLIs, AI agents, and text-focused workflows.

It allows users to copy clipboard images or files and paste usable file paths into the current environment using normal paste operations (`Ctrl+V` / `Cmd+V`).

Instead of treating paste as raw text only, clipterm materializes clipboard objects as real files inside the target environment and pastes the resulting absolute file path when that is more useful than native paste.

The goal is to make local shells, WSL, SSH sessions, tmux environments, and AI coding CLIs feel as seamless as desktop environments or virtual machines with shared clipboard support.

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

- Make `Ctrl+V` / `Cmd+V` work naturally for clipboard objects in terminals, CLIs, AI agents, and text-focused workflows
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

## Phase 1: Local Image Paste

Copy a screenshot or clipboard image, press `Ctrl+V` / `Cmd+V`, and either preserve native image paste where it works or paste the absolute path of a saved local image file where text paths are more useful.

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
