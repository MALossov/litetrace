# Litetrace — Vibe Coding / AI Agent 构建实践

## 概述

Litetrace 是一个纯 Go 实现的 Linux ftrace 内核追踪工具。本项目**几乎全部由 AI Agent（基于 Trae IDE）构建**，是人类与 AI 深度协作的典型案例。本文档记录整个 Vibe Coding 流程、人机分工模式与具体成果。

---

## 1. 项目解决的核心痛点

### 痛点一：Linux ftrace 工具链碎片化

Linux 内核自带的 ftrace 子系统功能强大，但原生的 `/sys/kernel/tracing/` 接口是裸文件操作——每次追踪都需要手动 `echo` 写入控制文件、`cat` 读取 trace 输出，过程极其繁琐。现有工具 `trace-cmd` 虽然封装了操作，但依赖 `libtraceevent`、`libtracefs` 等 C 库，编译部署复杂，不适合快速诊断场景。

**Litetrace 的解法**：单个 Go 二进制文件，零外部依赖，一个 `scp` 即可部署到任意 Linux 服务器，`sudo ./litetrace wizard` 即可开始追踪。

### 痛点二：内核调试门槛高

内核函数调用链分析通常需要:
- 理解 ftrace 的各种 tracer 类型差异
- 知道如何查找可追踪的内核函数名
- 手动管理 tracing_on / current_tracer / set_ftrace_filter / trace_pipe 等文件的状态机

新手面对这些概念往往无从下手。

**Litetrace 的解法**：
- `wizard` 交互式向导：5 步引导完成追踪配置，自动验证函数名有效性
- `search` 正则搜索：快速发现可追踪的内核函数
- `tldr` 快速帮助：一页纸说清所有功能

### 痛点三：追踪过程阻塞终端

传统 ftrace 的前台追踪（读取 `trace_pipe`）会阻塞终端，无法在追踪期间进行其他操作。我在独立操作时，容易忘记关闭追踪，导致资源浪费。同时小容量、断网场景下，联网的ai-agent难以部署。

**Litetrace 的解法**：
- `background` 后台追踪模式：使用 `setsid` 启动独立进程，PID 文件管理
- `terminate` 优雅停止：随时停止后台追踪并导出结果
- `status` 实时查看：包括后台进程运行状态

### 痛点四：缺乏直观的实时监控

命令行文本输出缺乏洞察力，无法快速感知系统调用热度。我缺乏一个直观的实时监控界面，无法在追踪过程中及时发现异常情况。同时不擅长tui编程，需要依赖AI Agent。

**Litetrace 的解法**：
- 使用`ai-agent`生成实时tui界面代码，同时自动帮我选型，我参考项目之后固定选型。
- `tui` 实时终端图形界面：状态面板 + 追踪数据面板 + 日志面板三栏布局
- 支持暂停/恢复、自动滚动、保存、清空等交互操作
- 快捷键系统完整（w/p/a/s/c/q/方向键）

---

## 2. 核心逻辑流与 Agent 协作模式

### 2.1 整体构建流水线

```
┌──────────────────────────────────────────────────────────────────┐
│                     Phase 1: Spec/Plan 规划                       │
│                                                                  │
│  人类描述需求 ──► AI Agent 输出 spec 文档 ──► 人类审查确认         │
│                                                                  │
│  产出: 项目大图、模块划分、数据流设计、CLI 接口定义                  │
├──────────────────────────────────────────────────────────────────┤
│                     Phase 2: AI Agent 全量构建                     │
│                                                                  │
│  AI Agent 独立完成:                                               │
│  · 库选型 (cobra + tview + promptui)                             │
│  · 项目脚手架搭建                                                 │
│  · 所有模块代码实现                                               │
│  · 单元测试编写                                                   │
│  · 文档撰写                                                       │
├──────────────────────────────────────────────────────────────────┤
│                     Phase 3: 人机协同调试                          │
│                                                                  │
│  人类介入 (AI 权限不可及的领域):                                   │
│  · Linux Kernel TraceFS 的实际挂载点探测                          │
│  · root 权限下的 ftrace 状态机验证                                │
│  · 内核版本差异导致的 tracefs 行为不同                              │
│                                                                  │
│  人类提出交互建议:                                                │
│  · TUI 快捷键设计                                                 │
│  · Wizard 向导流程优化                                            │
│  · --debug 模式输出格式                                           │
│  · background/terminate 子命令拆分                                │
│                                                                  │
│  AI Agent 接收反馈后完成修改 ──► 人类辅助验证                      │
├──────────────────────────────────────────────────────────────────┤
│                     Phase 4: 演示与交付                            │
│                                                                  │
│  AI Agent 独立撰写演示脚本 demo_script.sh (303行)                  │
│  人类执行脚本并录制 asciinema 视频                                 │
│  产出: litetrace_demo.gif                                        │
└──────────────────────────────────────────────────────────────────┘
```

### 2.2 长链推理特征

Litetrace 的构建包含多条长链推理路径：

**链路 A：ftrace 状态机推理**
```
挂载点发现 → 可用追踪器枚举 → 可用函数列表加载 
→ 占位 filter 设置 → nop 安全重置 → 目标 tracer 设置 
→ 实际 filter 设置 → tracing_on 使能 → trace_pipe 读取
→ 信号捕获 → tracing_on 关闭 → nop 重置 → filter 清空 → 结果导出
```
这是一个 14 步的有序操作链，任一步失败需回滚。AI Agent 在实现 `StartTracing()` 的 7 步启动流程和 `StopAndExport()` 的 4 步关闭流程时，需要推理每一步的依赖关系和错误传播路径。

**链路 B：后台进程生命周期推理**
```
用户发起 background → setsid 创建独立进程 
→ PID 写入 /var/run/litetrace/litetrace.pid 
→ 守护进程开始追踪 
→ status 读取 PID 文件查询状态 
→ terminate 发送 SIGTERM → 守护进程捕获信号 → 安全关闭 → 清理 PID 文件
```
这涉及进程管理、信号处理、文件锁、超时控制等多个维度的交叉推理。

**链路 C：TUI 事件驱动推理**
```
tview 应用初始化 → Flex 布局构建 
→ trace_pipe goroutine 启动 → 数据 channel 传递 
→ app.QueueUpdateDraw 线程安全刷新 
→ 键盘事件路由 (q/p/a/s/c/↑/↓/w) 
→ Wizard 模态框嵌套 
→ 退出时 goroutine 取消 + ftrace 状态清理
```

### 2.3 链式 Agent 调用模式

虽然本项目使用单一 Agent（Trae IDE 内置），但构建过程体现了类 Multi-Agent 的链式调用特征：

1. **Plan Agent 角色**：首轮对话中 AI Agent 输出完整 spec，包括模块树、接口签名、数据流方向，经人类审查后固化
2. **Build Agent 角色**：基于 spec 逐模块实现，从 `internal/ftrace` 核心引擎 → `internal/search` → `internal/ui` → `cmd/*`，自底向上构建
3. **Review Agent 角色**：每次代码提交后，AI Agent 自动进行 lint/typecheck 验证，发现错误后自行修复
4. **Doc Agent 角色**：独立撰写 README、cheatsheet、DEVELOPER.md、demo_script.sh 全部文档

---

## 3. AI Agent 独立完成的成果

### 3.1 库选型决策

AI Agent 在无人工干预下完成了技术选型：

| 功能 | 选型 | 理由 (由 AI Agent 判断) |
|------|------|------------------------|
| CLI 框架 | `spf13/cobra` | Go CLI 事实标准，子命令管理成熟 |
| TUI 框架 | `rivo/tview` + `gdamore/tcell` | 纯 Go 实现，与零依赖目标一致 |
| 交互提示 | `manifoldco/promptui` | 轻量级交互式选择/输入组件 |
| 测试 | Go 标准 `testing` | 零额外依赖 |
| 正则搜索 | Go 标准 `regexp` | 无需求第三方库 |
| 构建 | Go 原生 `go build` + Makefile | 最小化工具链 |
| 后台进程 | 标准库 `os/exec` + `setsid` | 利用 Linux 原生能力 |

全部依赖仅 3 个外部 Go 库，编译产物为单一静态二进制文件 ~8MB。

### 3.2 代码实现统计

| 模块 | 文件 | 核心职责 |
|------|------|---------|
| `cmd/` | 10 个文件 | CLI 子命令实现 (root/run/background/background_daemon/wizard/search/status/terminate/tui/tldr) |
| `internal/ftrace/` | 6 个文件 | ftrace 引擎、tracefs 发现、trace_pipe 流式读取、后台守护进程管理 |
| `internal/search/` | 2 个文件 | 正则搜索扫描器、过滤器验证与标准化 |
| `internal/ui/` | 1 个文件 | TUI Dashboard (三栏布局 + 事件驱动) |
| `internal/wizard/` | 1 个文件 | 5 步交互式向导 |
| `docs/` | 3 个文件 | 用户手册 + 开发者文档（本文档为第 4 个） |
| `scripts/` | 1 个文件 | 303 行完整功能演示脚本 |
| 根目录 | 3 个文件 | main.go / go.mod / Makefile |

### 3.3 关键算法由 AI Agent 设计

- **7 步安全启动流程**：nlop 占位 → tracer 切换 → filter 设置 → tracing_on，确保任何异常都不会留下脏状态
- **Filter 标准化与降级**：当用户输入的 filter 包含无效函数时，自动排除无效项而非直接报错，支持通配符匹配后仅保留有效函数
- **流式扫描器**：使用 `bufio.Scanner` 流式读取 `available_filter_functions`（数万行级别的文件），2 秒超时保护，防止 OOM
- **SafeShutdown 紧急恢复**：注册 SIGINT/SIGTERM handler，保证任何退出路径都会将 ftrace 恢复至 nop 安全状态

---

## 4. 人机协同的关键节点

### 4.1 人类突破 AI 的能力边界

**TraceFS 实际行为调试**：
- AI Agent 无法访问真实的 Linux 内核 tracefs 环境，只能基于文档推理
- 人类在真实 Linux 服务器上运行，发现：
  - 某些内核版本中 `set_ftrace_filter` 的默认值为 "all functions enabled"，写入单个函数前必须先清空
  - `trace_pipe` 被占用时返回 `EBUSY` 的具体错误处理
  - 不同发行版 tracefs 挂载点差异 (`/sys/kernel/tracing/` vs `/sys/kernel/debug/tracing/`)
- 人类将发现反馈给 AI Agent，AI 据此完善 `discover.go` 的多路径探测逻辑和错误处理

### 4.2 人类提出的交互设计改进

| 改进项 | 原始 AI 设计 | 人类反馈 | 最终实现 |
|--------|-------------|---------|---------|
| TUI 快捷键 | 仅 q 退出 | 增加 p(暂停)/a(自动滚动)/s(保存)/c(清空)/w(向导) | 6 个快捷键 |
| background 子命令 | 合并到 run 命令中 | 拆分为独立子命令 + terminate 配对 | 语义更清晰 |
| Wizard 查看模式 | 仅 TUI | 增加 Run silently / Background 三种模式 | 覆盖全部使用场景 |
| --debug 输出 | 无 | 建议黄色高亮区分调试信息 | 增强可观测性 |
| status 展示 | 仅 ftrace 状态 | 增加后台进程状态段落 | 完整运行态视图 |

### 4.3 演示脚本的人机协作

`scripts/demo_script.sh`（303 行）由 AI Agent 独立撰写，包含：
- 5 个部分的结构化演示流程（基础命令 / 前台追踪 / 后台追踪 / TUI / Wizard）
- ANSI 颜色着色美化输出
- 分步暂停等待用户阅读
- 实时命令执行与结果展示
- 完整的 ASCII 艺术框体和交互提示

人类负责：在真实环境中执行脚本、用 asciinema 录制、转换为 GIF。AI Agent 负责：脚本的全部内容创作。

---

## 5. 量化成果

| 指标 | 数据 |
|------|------|
| AI Agent 代码贡献率 | ~98%（仅 tracefs 行为调优含人类输入） |
| 外部 Go 依赖数 | 3 个（cobra + tview/tcell + promptui） |
| 编译产物大小 | ~8MB 单二进制 |
| 支持子命令数 | 8 个（status/search/run/wizard/tui/background/terminate/tldr） |
| 支持 tracer 类型 | 2 种（function + function_graph）+ nop 安全模式 |
| 文档字数 | ~15,000 字（README + cheatsheet + DEVELOPER + 本文档） |
| 演示脚本行数 | 303 行 Shell 脚本 |
| 核心推理链路 | 3 条长链（ftrace 状态机 14 步 / 后台进程生命周期 / TUI 事件驱动） |

---

## 6. 方法论总结

Litetrace 验证了 **"Spec → Agent Build → Human Debug → Agent Fix → Deliver"** 的 Vibe Coding 工作流在系统工具类项目中的可行性：

1. **Spec 先行**：人类用自然语言描述需求，AI Agent 输出结构化 spec，确保双方对目标达成共识
2. **Agent 全量构建**：库选型、代码实现、测试编写、文档撰写全部托付 AI Agent
3. **人机能力互补**：AI 负责代码生成（快速、规范），人类负责内核环境调试（AI 不可及）和交互体验判断（AI 缺乏主观感受）
4. **快速迭代**：人类发现问题 → 自然语言反馈 → AI Agent 修改 → 人类验证，闭环周期 < 5 分钟

---

## 7. 相关文件

- [README.md](../README.md) — 项目说明与快速开始
- [cheatsheet.txt](./cheatsheet.txt) — 完整使用手册
- [DEVELOPER.md](./DEVELOPER.md) — 开发者文档（含架构设计与 API 参考）
- [demo_script.sh](../scripts/demo_script.sh) — 由 AI Agent 独立撰写的功能演示脚本
- [litetrace_demo.gif](./litetrace_demo.gif) — 人机协作录制的演示视频
