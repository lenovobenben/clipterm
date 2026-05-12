# MVP：本地 CLI/Agent 智能粘贴

本文档记录 clipterm 当前 MVP 的产品语义、平台行为、已知限制和后续路线。README 面向快速使用；本文档面向维护者和后续开发者。

## 目标

clipterm 的核心目标是解决一个具体痛点：图片或截图在进入终端、CLI、AI Agent 或普通文本输入框之前，通常要先手动保存成文件，再复制文件路径。clipterm 把这一步变成一个系统级智能粘贴动作。

核心模型：

```text
clipboard image/file -> local file path -> paste into current input
```

当前 MVP 聚焦本机智能粘贴，并把这件事作为核心能力：

- 复制图片或截图后，智能粘贴生成本机 PNG 文件并粘贴路径。
- 复制单个文件后，智能粘贴粘贴原文件绝对路径。
- 复制普通文本后，智能粘贴等价于系统原生粘贴。
- 普通 `Cmd+V` / `Ctrl+V` 不接管。

这个 MVP 只把本地剪贴板对象变成本地可访问路径。

## 当前平台状态

| 平台 | 智能粘贴快捷键 | 普通粘贴 | 状态 |
| --- | --- | --- | --- |
| macOS | `Cmd+Shift+V` | 不接管 `Cmd+V` | 已实现原型 |
| Windows x86_64 | `Ctrl+Shift+V` | 不接管 `Ctrl+V` | 已实现原型 |
| Windows ARM | - | - | 暂不支持 |
| Linux Desktop | - | - | 暂缓 |

## 统一行为

| 剪贴板内容 | 智能粘贴行为 |
| --- | --- |
| 普通文本和其他普通内容 | 发送一次原生粘贴，不修改剪贴板。 |
| 单个复制文件 | 粘贴该文件的绝对路径，不复制文件内容。 |
| 截图或图片流 | 保存为 clipterm 缓存目录下的 PNG，再粘贴生成路径。 |
| 多个复制文件 | 暂不支持，返回错误，不污染剪贴板。 |

一次成功的图片或文件路径粘贴后，系统剪贴板会变成路径文本。因此用户可以继续按普通 `Cmd+V` / `Ctrl+V` 多次粘贴同一个路径，直到下一次复制其他内容。文本 fallback 不会修改剪贴板。

## 非目标

当前 MVP 不做：

- 普通 `Cmd+V` / `Ctrl+V` 接管。
- 应用级“原生图片粘贴”放行规则。
- 多文件、目录、大文件传输。
- 剪贴板恢复。
- Linux 桌面支持。

这些方向统一放入“后续路线（不保证实现）”。它们不能改变当前智能粘贴的稳定语义。

## macOS 设计

macOS 使用 `Cmd+Shift+V` 作为 CLI/Agent 智能粘贴快捷键。

缓存目录：

```text
~/Library/Caches/clipterm/
```

需要权限：

- Input Monitoring / 键盘事件访问。
- Accessibility / 合成粘贴事件。

macOS 粘贴是对象语义，不只是文本字符串。Finder 复制文件时，pasteboard 里可能同时有文件 URL、展示名、图标和预览等表示。不同目标应用会选择不同表示。因此 clipterm 不接管普通 `Cmd+V`，只通过 `Cmd+Shift+V` 明确表达“我要面向 CLI/Agent 的路径粘贴”。

## Windows 设计

Windows 使用原生 `.exe` 运行，监听 `Ctrl+Shift+V`。WSL 中运行的 Linux 进程不能直接注册 Windows 全局热键、读取 Windows 剪贴板或调用 `SendInput`，所以 Windows 版必须作为 Windows 进程启动。

缓存目录：

```text
%LOCALAPPDATA%\clipterm\cache\
```

主要平台 API：

- `RegisterHotKey`：注册 `Ctrl+Shift+V`。
- `CF_HDROP`：读取 Explorer 复制的文件路径。
- `CF_DIBV5` / `CF_DIB`：读取常见图片剪贴板数据。
- `CF_UNICODETEXT`：写入路径文本。
- `SendInput`：发送原生 `Ctrl+V`。

Windows 图片处理：

```text
CF_DIB/CF_DIBV5 -> RGBA image -> PNG -> cache file -> output path
```

当前支持常见 24-bit / 32-bit 未压缩 DIB。PNG 编码是无损的；32-bit DIB 会尽量保留 alpha。如果 32-bit DIB 的 alpha 全为 0，会按 Windows 常见截图语义当作不透明图片处理，避免生成全透明 PNG。调色板位图、压缩位图和更复杂色彩空间后续再扩展。

### Windows 路径风格

Windows 原生路径和 WSL 路径不是同一种绝对路径：

```text
C:\Users\Alice\Pictures\a.png
/mnt/c/Users/Alice/Pictures/a.png
```

Windows 版通过启动参数选择输出路径风格：

```powershell
clipterm.exe daemon --path-style windows
clipterm.exe daemon --path-style wsl
```

语义：

- `windows`：保持 Windows 原生路径。
- `wsl`：将盘符路径转换为 `/mnt/<drive>/...`。
- `native`：保持原路径，主要用于跨平台默认语义。

路径转换发生在写入剪贴板之前：

```text
clipboard object -> Windows path -> output path transform -> clipboard text -> Ctrl+V
```

修改路径风格需要重启 daemon。当前不做配置热加载。

### Windows 已验证目标

已验证可用：

- Notepad
- 浏览器地址栏
- PowerShell
- Windows Terminal
- 从 PowerShell 进入的 WSL 终端

已知限制：

- Notepad++ 当前不作为 Windows 首版支持目标。实测中 clipterm 已经成功把路径写入系统剪贴板，但 Notepad++ 对 `Ctrl+Shift+V` 触发期间的合成粘贴事件处理不同于 Notepad、浏览器地址栏、PowerShell 和 Windows Terminal，可能不会把路径插入编辑区。
- 这不是路径转换或剪贴板写入失败；同一轮操作后用 PowerShell `Get-Clipboard` 可以看到路径文本。后续若要支持 Notepad++，应作为单独兼容性工作处理。

## 命令形态

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

`clipterm paste` 是可调试核心原语：读取剪贴板图片或单文件，生成路径并输出到 stdout。

`clipterm daemon` 是日常交互入口：后台监听智能粘贴快捷键，执行文件/图片路径化或文本 fallback。

## 缓存策略

使用用户级缓存目录：

```text
macOS:   ~/Library/Caches/clipterm/
Windows: %LOCALAPPDATA%\clipterm\cache\
```

生成文件名：

```text
clipterm-YYYYMMDD-HHMMSS-<short-random>.png
```

`clipterm clean` 只清理 clipterm 管理的 `clipterm-*.png`。

## 代码结构

核心业务语义：

- `internal/clipterm`：文件优先、图片其次、文本 fallback、多文件拒绝。
- `internal/materialize`：图片落盘、缓存目录、清理。
- `internal/pathstyle`：Windows / WSL 输出路径转换。

平台 adapter：

- `internal/clipboard/*`
- `internal/hotkey/*`
- `internal/paste/*`
- `internal/daemon/process_*`

命令入口：

- `internal/cli`
- `cmd/clipterm`

平台代码应保持相互隔离。Windows 的热键、剪贴板和粘贴实现不应影响 macOS 已验证路径。

## 验收标准

macOS：

- 普通文本：`Cmd+Shift+V` 等价于原生 `Cmd+V`，不修改剪贴板。
- Finder 单文件：粘贴原文件绝对路径。
- 图片/截图：保存 PNG 并粘贴生成路径。
- 多文件：拒绝处理，不污染剪贴板。

Windows：

- 普通文本：`Ctrl+Shift+V` 等价于原生 `Ctrl+V`，不修改剪贴板。
- Explorer 单文件：粘贴原文件路径，支持 `windows` / `wsl` 输出风格。
- 图片/截图：保存 PNG 并粘贴生成路径，支持 `windows` / `wsl` 输出风格。
- 多文件：拒绝处理，不污染剪贴板。
- 普通 `Ctrl+V` 不接管。
- daemon 重复启动不会产生多个后台进程。

## 后续路线（不保证实现）

以下方向只作为可能性记录，不是当前 MVP 的承诺。

- 扩展 Windows 图片格式覆盖，例如调色板 DIB、压缩 DIB 或直接 PNG 剪贴板格式。
- 提供更正式的 Windows 启动方式，例如 Startup folder、Task Scheduler 或安装器。
- 支持可配置智能粘贴快捷键。
- 实现可选剪贴板恢复，避免路径粘贴后覆盖原始图片/文件对象。
- 研究路径 shell escaping，但当前先粘贴原始绝对路径。
- 研究普通 `Cmd+V` / `Ctrl+V` 接管，但这会影响用户原生粘贴，当前不做。

当前文档和代码应继续围绕已实现的本地图片/单文件路径粘贴能力演进。
