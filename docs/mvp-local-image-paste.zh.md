# MVP：本地图片智能粘贴

## 目的

clipterm 的第一个里程碑是让剪贴板图片能够自然地进入本地应用、终端、CLI 和 AI Agent 工作流。

用户复制截图或其他图片后，仍然按熟悉的 `Cmd+V` / `Ctrl+V`。clipterm 在后台判断当前粘贴目标：

- 如果目标应用能原生接收图片，例如 ChatGPT 网页版、聊天工具或文档编辑器，优先保留原生图片粘贴体验。
- 如果目标应用不能有效接收图片，例如终端、浏览器地址栏、记事本或普通文本框，clipterm 将图片保存为本地文件，并粘贴该文件的绝对路径。

第一阶段的核心目标是无感：用户不再为了在终端、CLI 或 AI Agent 中使用截图而手动保存图片、打开文件浏览器、寻找文件、复制路径。

## 统一产品概念

clipterm 的核心不是“截图工具”，也不是“剪贴板管理器”。它的核心概念是：

```text
clipboard object -> materialized file -> usable path in the current environment
```

本地智能粘贴和未来远程传输属于同一个产品线：

```text
本地图片粘贴:
clipboard image -> local file -> local absolute path paste

远程图片/文件传输:
clipboard image/file -> remote file -> remote absolute path paste
```

因此当前不拆成两个项目。工程上可以拆模块，但产品和命令入口保持统一。

建议 tagline：

```text
Materialize clipboard objects for terminals, CLIs, and AI agents.
```

中文定位：

```text
把剪贴板里的图片和文件，变成终端、CLI 和 AI Agent 可用的真实文件路径。
```

## 产品定义

clipterm 是一个系统级智能粘贴工具。

当用户触发粘贴时，如果剪贴板中包含图片，clipterm 会根据当前应用规则决定：

1. 放行原生图片粘贴。
2. 或将图片保存到本地缓存目录，再把绝对路径作为文本粘贴到当前焦点位置。

示例结果：

```text
/Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

Windows 示例：

```text
C:\Users\Alice\AppData\Local\clipterm\cache\clipterm-20260509-153012-a83f.png
```

## 第一阶段范围

第一阶段只实现本地图片智能粘贴。

包括：

- 读取系统剪贴板中的图片。
- 将图片保存为本机真实文件。
- 在路径粘贴目标中粘贴本机绝对路径。
- 在原生图片粘贴目标中尽量放行原生粘贴。
- 保持普通文本粘贴行为不变。

不包括：

- SSH 或远程 materialization。
- WSL 路径桥接。
- tmux 感知传输。
- 文件上传/下载。
- 目录或多文件传输。
- 大文件传输。

## 目标用户

- 需要在终端命令里引用截图的开发者。
- 需要把图片路径交给 AI coding CLI 的用户。
- 使用 ChatGPT、Codex、Claude Code、Cursor、VS Code、JetBrains、Terminal.app、iTerm2、Windows Terminal 等工具的人。
- 使用微信、QQ、macOS 截图、Windows Snipping Tool 等工具复制截图的人。
- 不想手动保存截图、查找文件、复制路径的人。

## 用户体验

### 路径粘贴流程

```text
用户复制截图图片
        |
用户聚焦终端、记事本、浏览器地址栏或其他文本输入位置
        |
用户按 Cmd+V / Ctrl+V
        |
clipterm 检测到剪贴板中有图片
        |
clipterm 判断当前目标应使用路径粘贴
        |
clipterm 将图片保存到本地缓存
        |
clipterm 将文件路径作为文本写入剪贴板
        |
clipterm 触发一次普通粘贴
        |
当前应用收到图片文件的绝对路径
```

示例：

```bash
python inspect_image.py /Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

AI CLI 示例：

```text
请查看这个截图：/Users/alice/Library/Caches/clipterm/clipterm-20260509-153012-a83f.png
```

### 原生图片粘贴流程

```text
用户复制截图图片
        |
用户聚焦支持图片粘贴的应用
        |
用户按 Cmd+V / Ctrl+V
        |
clipterm 判断当前目标应使用原生粘贴
        |
clipterm 放行系统原生粘贴
        |
目标应用直接接收图片
```

例如 ChatGPT 网页版如果能接收图片，就应该保持原生上传体验。

## 智能粘贴策略

clipterm 采用高级方案：接管普通 `Cmd+V` / `Ctrl+V`，目标是用户无感。

但系统级程序很难可靠判断“某个具体输入框是否接受图片”。同一个浏览器中，ChatGPT 对话框可能能接收图片，地址栏只能接收文本。因此第一版需要应用规则和保守 fallback。

### 规则类型

- `native`：放行原生粘贴。适合 ChatGPT 网页版、微信、Slack、文档编辑器等支持图片粘贴的场景。
- `path`：转换为路径粘贴。适合终端、AI CLI、记事本、浏览器地址栏、纯文本输入框。
- `auto`：默认策略。根据应用、窗口标题、焦点上下文和未来可用的检测能力决定。

### 保守原则

- 对明确路径粘贴目标，转换为路径。
- 对明确原生图片粘贴目标，放行原生粘贴。
- 对无法识别或风险较高的目标，优先不破坏原生粘贴。
- 后续可以提供“强制路径粘贴”快捷键作为兜底，例如 `Cmd+Shift+V` / `Ctrl+Shift+V`。

## 初始平台范围

### macOS

MVP 的第一优先级平台。

- 普通粘贴快捷键：`Cmd+V`
- 可选强制路径粘贴快捷键：`Cmd+Shift+V`
- 缓存目录：`~/Library/Caches/clipterm/`
- 需要关注的权限：Accessibility、剪贴板读写、合成键盘事件。

重点路径粘贴目标：

- Terminal.app
- iTerm2
- WezTerm
- Ghostty
- Alacritty
- VS Code / Cursor / JetBrains 终端
- TextEdit 纯文本模式
- 浏览器地址栏
- 普通文本输入框

重点原生粘贴目标：

- ChatGPT 网页版的图片输入场景
- 微信、Slack 等聊天工具
- 支持图片粘贴的文档编辑器

### Windows

macOS 原型验证后作为第二个一等平台支持。

- 普通粘贴快捷键：`Ctrl+V`
- 可选强制路径粘贴快捷键：`Ctrl+Shift+V`
- 缓存目录：`%LOCALAPPDATA%\clipterm\cache\`
- 需要关注的权限：剪贴板读写、键盘 hook、前台窗口识别、合成键盘事件。

重点路径粘贴目标：

- Windows Terminal
- PowerShell
- cmd
- VS Code / Cursor / JetBrains 终端
- Notepad
- 浏览器地址栏
- 普通文本输入框

### Linux Desktop

暂缓。Linux 桌面剪贴板在 X11、Wayland、不同桌面环境和终端模拟器之间差异较大，不应阻塞初始 MVP。

## 命令形态

项目应同时提供直接 CLI 命令和后台 daemon。

```bash
clipterm paste
clipterm paste --copy-path
clipterm daemon
clipterm rules
clipterm clean
clipterm doctor
```

### `clipterm paste`

可调试的核心原语。

行为：

1. 从系统剪贴板读取图片数据。
2. 将图片保存到 clipterm 缓存目录。
3. 将绝对路径输出到 stdout。

### `clipterm paste --copy-path`

行为与 `clipterm paste` 相同，但会额外把生成的路径作为文本写回系统剪贴板。

这是实现系统级智能粘贴之前的重要中间阶段。

### `clipterm daemon`

提供最终 MVP 用户体验的后台进程。

行为：

1. 监听平台普通粘贴快捷键。
2. 检查剪贴板是否包含图片。
3. 识别前台应用和可用上下文。
4. 根据规则放行原生粘贴或执行路径粘贴。
5. 对非图片剪贴板保持普通粘贴行为不变。

### `clipterm rules`

管理应用级粘贴规则。

第一版可以先内置规则，后续再暴露配置文件或命令行管理。

### `clipterm clean`

删除旧缓存文件。

默认清理策略可以保守一些：

- 保留最近 7 天的文件。
- 后续可配置总缓存大小上限。

### `clipterm doctor`

检查平台集成状态。

有用的检查项：

- 剪贴板读取权限。
- 剪贴板写入权限。
- 缓存目录写入权限。
- 全局快捷键或键盘 hook 权限。
- macOS Accessibility 权限。
- 前台应用检测。
- 合成粘贴事件。
- 当前应用规则匹配结果。

## 缓存策略

使用用户级缓存目录，而不是通用临时目录。

推荐目录：

```text
macOS:   ~/Library/Caches/clipterm/
Windows: %LOCALAPPDATA%\clipterm\cache\
```

推荐文件名格式：

```text
clipterm-YYYYMMDD-HHMMSS-<short-random>.png
```

示例：

```text
clipterm-20260509-153012-a83f.png
```

MVP 应将剪贴板图片保存为 PNG。如果后续需要保留原始来源格式，再增加其他格式支持。

## 剪贴板恢复

路径粘贴时，daemon 需要临时把剪贴板中的图片替换成文本路径，目标应用才能通过普通粘贴收到路径。

第一版建议实现简单模式：

1. 保存图片。
2. 将路径写入剪贴板。
3. 触发粘贴。
4. 暂不恢复原始图片。

后续可增加完整体验模式：

1. 保存原始剪贴板图片。
2. 将剪贴板替换为路径文本。
3. 触发粘贴。
4. 延迟 `300ms` 到 `800ms` 后恢复原始图片。

完整体验更符合用户预期，但存在平台复杂度和竞态风险。如果用户在恢复窗口内复制了其他内容，clipterm 不能覆盖用户的新剪贴板。

## 远程能力可行性结论

远程图片粘贴和远程单文件传输是可实现的，但不属于第一阶段。

未来远程能力可以设计为：

```text
local clipboard image/file
 -> stream/chunk encode
 -> terminal-safe text channel
 -> remote receive helper
 -> remote materialized file
 -> remote absolute path paste
```

关键判断：

- 默认上限 10MB 单文件是合理承诺。
- Base64 支持流式编码/解码，可以边读、边编码、边发送，降低内存峰值。
- 终端文本通道仍需要 chunk、节流、checksum 和失败处理。
- 远端安装 helper 时，可以做到接近无感；无 helper 的 shell bootstrap fallback 不能保证完全无感。
- 这可以成为 clipboard-scale 的 rz/sz 风格上传/下载 fallback，但不是 scp/rsync 的大文件或高性能替代品。

远程方向暂时只保留为路线图和架构边界，不在本 MVP 中展开协议细节。

## 项目边界

不拆项目。

理由：

- 本地智能粘贴和远程 materialization 共用同一个核心模型。
- 用户心智一致：剪贴板对象变成当前环境可用路径。
- 工程上可以拆模块，产品上保持一个名字和一个 CLI。

建议内部模块：

```text
clipboard/
materialize/
paste/
rules/
platform/
transport/
remote/
```

未来命令可以扩展为：

```bash
clipterm send
clipterm receive
```

但第一阶段只实现本地图片智能粘贴。

## 技术栈方向

第一实现语言改为 Go。

原因：

- 当前维护者熟悉 Go，有利于长期迭代。
- Go 可以发布单个原生 binary。
- Go 标准库适合文件 IO、流式 Base64、chunk、checksum 和远程传输。
- macOS / Windows 平台集成可以通过平台 API、cgo、syscall 或外部命令 fallback 分层实现。
- 第一阶段的主要难点是系统集成和产品体验，不是语言层面的性能极限。

## 建议 Go 依赖和接口

这些只是候选项，不是最终承诺：

- `cobra` 或标准库 `flag`：CLI 参数解析。
- Go 标准库 `os.UserCacheDir`：平台缓存目录。
- Go 标准库 `image/png`：PNG 编码。
- Go 标准库 `time`：带时间戳的文件名。
- Go 标准库 `crypto/rand` 或 `math/rand/v2`：短随机文件名后缀。
- Go 标准库 `encoding/base64`：未来远程流式 Base64 编码/解码。
- Go 标准库 `crypto/sha256`：未来远程传输校验。
- 平台专用剪贴板 API：可靠处理图片剪贴板。
- 平台专用键盘和前台窗口 API：支持 daemon 行为。
- 平台专用应用识别 API：支持应用级粘贴规则。

剪贴板、应用规则、粘贴事件和平台集成应抽象到内部 interface 后面，让 macOS 和 Windows 可以独立演进。

## 开放问题

- 第一版应该只做 macOS，还是 macOS 和 Windows 一起做？
- 第一版内置哪些 `native` / `path` 应用规则？
- 普通 `Cmd+V` / `Ctrl+V` 接管失败时，如何最小化对用户原生粘贴的影响？
- 是否提供强制路径粘贴快捷键作为 fallback？
- 剪贴板恢复应该默认开启、可选开启，还是完全延后？
- 粘贴生成路径时是否需要 shell escaping，还是直接粘贴原始绝对路径？
- 浏览器内部输入框的细粒度识别是否放到后续版本？

对路径 escaping 的初始建议：生成路径中避免空格，直接粘贴原始绝对路径。shell 特定 escaping 后续再加。
