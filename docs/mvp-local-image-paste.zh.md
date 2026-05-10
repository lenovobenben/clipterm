# MVP：本地 CLI/Agent 智能粘贴

## 目的

clipterm 的第一个里程碑是让剪贴板图片、单个文件和普通文本能够自然地进入本地终端、CLI 和 AI Agent 工作流。

远景探索方向是：用户复制截图或其他图片后，仍然按熟悉的 `Cmd+V` / `Ctrl+V`，clipterm 在后台判断当前粘贴目标：

- 如果目标应用能原生接收图片，例如 ChatGPT 网页版、聊天工具或文档编辑器，优先保留原生图片粘贴体验。
- 如果目标应用不能有效接收图片，例如终端、浏览器地址栏、记事本或普通文本框，clipterm 将图片保存为本地文件，并粘贴该文件的绝对路径。

当前 macOS 原型先采用更安全的快捷键：`Cmd+Shift+V` 表示面向 CLI/Agent 的智能粘贴：图片和单个文件转成绝对路径，普通文本和其他普通内容走原生 `Cmd+V` fallback。普通 `Cmd+V` 暂不接管，避免影响系统原生复制粘贴。全面替代普通粘贴只是远景探索，不是第一阶段承诺。

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

当用户触发 clipterm 智能粘贴时，如果剪贴板中包含图片或单个文件，clipterm 会生成或读取一个本机绝对路径，并把该路径粘贴到当前焦点位置。如果剪贴板是普通文本或其他普通内容，clipterm 不改写剪贴板，只发送一次原生 `Cmd+V`。

如果未来确认可以安全接管普通 `Cmd+V`，clipterm 可以根据当前应用规则决定：

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

第一阶段先实现本地 CLI/Agent 智能粘贴。

包括：

- 读取系统剪贴板中的图片。
- 读取 Finder 中复制的单个文件路径。
- 将图片保存为本机真实文件。
- 使用 `Cmd+Shift+V` 在当前焦点位置粘贴本机绝对路径。
- 普通文本和其他普通内容走原生 `Cmd+V` fallback，不改写剪贴板。

不包括：

- 普通 `Cmd+V` 接管。
- 针对具体应用的原生图片粘贴放行规则。
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
用户按 Cmd+Shift+V
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

### 文本 fallback 流程

```text
用户复制普通文本
        |
用户聚焦终端、AI CLI 或其他文本输入位置
        |
用户按 Cmd+Shift+V
        |
clipterm 未检测到可路径化的图片或单个文件
        |
clipterm 不修改剪贴板
        |
clipterm 发送一次原生 Cmd+V
        |
当前应用收到系统原生粘贴内容
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

当前原型采用低风险方案：不接管普通 `Cmd+V` / `Ctrl+V`，只提供 `Cmd+Shift+V` 作为 CLI/Agent 智能粘贴。

未来可以探索更高级的普通 `Cmd+V` / `Ctrl+V` 接管方案，但是否要做取决于稳定性和负面影响评估。

系统级程序很难可靠判断“某个具体输入框是否接受图片”。同一个浏览器中，ChatGPT 对话框可能能接收图片，地址栏只能接收文本。因此如果未来探索普通 `Cmd+V` 智能接管，需要应用规则和保守 fallback。当前第一阶段先不做这件事。

### 规则类型

- `native`：放行原生粘贴。适合 ChatGPT 网页版、微信、Slack、文档编辑器等支持图片粘贴的场景。
- `path`：转换为路径粘贴。适合图片、截图和单个复制文件。
- `native-text`：普通文本和其他普通内容不改写剪贴板，直接发送原生粘贴。
- `auto`：默认策略。根据应用、窗口标题、焦点上下文和未来可用的检测能力决定。

### 保守原则

- 对明确路径粘贴目标，转换为路径。
- 对明确原生图片粘贴目标，放行原生粘贴。
- 对普通文本，保持原生粘贴，不读取或重写文本内容。
- 对无法识别或风险较高的目标，优先不破坏原生粘贴。
- 当前已提供 `Cmd+Shift+V` 作为 CLI/Agent 智能粘贴快捷键。未来如果接管普通 `Cmd+V`，该快捷键仍可作为稳定兜底。

## 初始平台范围

### macOS

MVP 的第一优先级平台。

- 普通粘贴快捷键：`Cmd+V`，当前不接管。
- CLI/Agent 智能粘贴快捷键：`Cmd+Shift+V`
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
- 候选 CLI/Agent 智能粘贴快捷键：`Ctrl+Shift+V`
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

提供当前 macOS 原型交互的后台进程。

行为：

1. 监听 `Cmd+Shift+V`。
2. 优先检查剪贴板是否包含单个复制文件。
3. 如果是单个文件，复制并粘贴原文件绝对路径。
4. 否则检查剪贴板是否包含图片流。
5. 如果是图片流，将图片保存为缓存 PNG，复制并粘贴生成的绝对路径。
6. 如果都不是，直接发送一次原生 `Cmd+V`，不修改剪贴板。

### daemon 启动建议

当前原型建议用户把 daemon 启动命令放进 shell startup 文件，例如 `~/.zshrc` 或 `~/.bash_profile`。

```bash
command -v clipterm >/dev/null 2>&1 && clipterm daemon >/dev/null 2>&1 || true
```

如果 `clipterm` 没有加入 `PATH`，可以使用安装路径：

```bash
$HOME/.local/bin/clipterm daemon >/dev/null 2>&1 || true
```

这个方案适合早期原型：

- `clipterm daemon` 是幂等的；如果已经运行，不会重复启动多个后台进程。
- 如果 daemon 被人为 kill，下一次执行 `clipterm daemon` 会识别 stale PID 文件并重新启动。
- 打开新终端时会自动尝试启动，命令会立即返回，不占用终端窗口。
- 不引入 launchd 代码和安装复杂度，用户可以直接从 shell 配置中删除。

macOS 的更正式方案是 LaunchAgent 登录自启动，但当前阶段暂不实现。等项目更稳定、有安装器或发布包后，再考虑提供可选 LaunchAgent。

### `clipterm rules`

显示当前粘贴策略。

当前原型还没有应用级规则管理。后续如果探索普通 `Cmd+V` 智能接管，再考虑内置规则、配置文件或命令行管理。

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
- 当前粘贴策略。

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

## macOS 原生粘贴与 clipterm 粘贴对比

macOS 的复制粘贴不是简单地复制一个固定字符串，而是把一个剪贴板对象放进系统 pasteboard。一个 Finder 文件对象通常可以同时包含文件 URL、展示名、图标、预览和类型信息。目标应用会自行选择它能理解的表示。

因此，同一个 Finder 文件执行 `Cmd+C` 后：

- 粘贴到浏览器地址栏，可能得到文件名。
- 粘贴到 iTerm2，可能得到绝对路径。
- 粘贴到支持文件对象的应用，可能直接触发文件导入或上传。

clipterm 当前原型不接管普通 `Cmd+V`。当前的 `Cmd+Shift+V` 是面向 CLI/Agent 的智能粘贴动作：能路径化的对象转换为本机绝对路径；普通文本和其他普通内容保持原生粘贴。

| 剪贴板内容 | macOS 原生 `Cmd+V` | clipterm `Cmd+Shift+V` |
| --- | --- | --- |
| 截图或图片流 | 由目标应用决定。支持图片的应用可能直接接收图片；普通文本输入框可能没有反应或读取其他表示。 | 将图片保存为 clipterm 缓存目录下的 PNG，复制生成的绝对路径，并粘贴该路径。 |
| Finder 中复制的单个文件 | 由目标应用决定。可能粘贴文件名、绝对路径，或直接消费文件对象。 | 直接粘贴原文件的绝对路径。不复制文件内容。 |
| Finder 中复制的多个文件 | 由目标应用决定。 | 暂不支持。clipterm 拒绝处理，并且不修改剪贴板。 |
| 普通文本和其他普通内容 | 正常粘贴内容。 | 发送一次原生 `Cmd+V`，不修改剪贴板，保留系统原生文本/富文本粘贴行为。 |

一次成功的图片或文件路径粘贴后，系统剪贴板会变成路径文本。因此用户可以继续按普通 `Cmd+V` 多次粘贴同一个路径，直到下一次复制其他内容。文本 fallback 不会修改剪贴板。

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

但第一阶段只实现本地 CLI/Agent 智能粘贴。

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
- Windows 版本是否也采用 `Ctrl+Shift+V` 作为 CLI/Agent 智能粘贴快捷键？
- 如果未来探索普通 `Cmd+V` / `Ctrl+V` 接管，如何最小化对用户原生粘贴的影响？
- 是否需要把当前 `Cmd+Shift+V` 智能粘贴做成可配置快捷键？
- 剪贴板恢复应该默认开启、可选开启，还是完全延后？
- 粘贴生成路径时是否需要 shell escaping，还是直接粘贴原始绝对路径？
- 浏览器内部输入框的细粒度识别是否放到后续版本？

对路径 escaping 的初始建议：生成路径中避免空格，直接粘贴原始绝对路径。shell 特定 escaping 后续再加。
