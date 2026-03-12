# Litetrace 究极工程架构与实施契约 (V-Final Ultimate Edition)

## 📜 文档定位与作业需求溯源

### 文档定位
这是 `litetrace` 项目的**唯一真理版本（Single Source of Truth）**。它融合了 V4 的完整工程架构与 V5 的深度防御设计，确保交付的不仅是一个课程作业，而是一个可以上生产环境的艺术品。

### 原始作业要求（来自《linux 调试技术_ftrace 课后作业.docx》）

| 作业要求 | 技术映射 | 验收标准 |
|----------|----------|----------|
| **实现一个简易版 trace-cmd 工具** | 纯 Go 实现，零外部二进制依赖 | 不调用 `trace-cmd`，所有操作通过读写 tracefs 完成 |
| **支持打开 function 跟踪** | 向 `current_tracer` 写入 `function` | `cat current_tracer` 显示 `function` |
| **支持 function 过滤** | 向 `set_ftrace_filter` 写入函数名 | `cat set_ftrace_filter` 显示过滤函数 |
| **动态开启和关闭跟踪** | 向 `tracing_on` 写入 `1` 或 `0` | 可控制追踪启停 |
| **支持查看当前配置状态** | 读取 `current_tracer`、`tracing_on`、`set_ftrace_filter` | 格式化输出当前状态 |
| **支持导出跟踪结果** | 读取 `trace` 文件落地到本地 | 导出文件包含真实内核追踪日志 |
| **打印 trace 中的跟踪内容** | 读取 `trace_pipe` 实时流 | TUI 或终端实时显示 |

---

## 0. 核心原理认知与深度引申

### 0.1 ftrace 本质
ftrace 并不是一个独立的程序，而是**深植于 Linux 内核态的追踪框架**。我们操作的 `/sys/kernel/debug/tracing` 实际上是基于内存的虚拟文件系统（Tracefs/Debugfs）。对这些文件的读写，本质上是**触发内核的 file_operations 回调**，直接操控内核的 **Ring Buffer（环形缓冲区）**和**函数探针（Ftrace Probes）**。

### 0.2 关键引申
| 认知点 | 工程含义 |
|--------|----------|
| **不加过滤的 function 追踪每秒产生数百万次事件** | 过滤不是可选项，而是**保命符**！必须默认要求用户设置 filter |
| **并发场景可能读到脏数据** | 状态解析必须具备极强的容错机制，不能因为读到未知 tracer 就 panic |
| **trace vs trace_pipe 的本质区别** | `trace`=一次性快照（不消耗 buffer）；`trace_pipe`=消费型管道（读取后出队） |
| **纯原生文件 I/O 的优势** | 实现亚毫秒级的配置下发延迟，无需 `os/exec` 包装 shell 命令 |

---

## 1. 严格的目录架构契约与领域驱动设计

采用标准 Go 工程规范，强制将底层逻辑封装在 `internal/` 目录下。这种设计防止了内部业务逻辑被外部意外引用（Go 编译器的原生隔离限制），保证命令行 UI 层与内核读写层彻底解耦。

```text
litetrace/
├── go.mod                          # 模块定义：github.com/malossov/lite-tracer-mygo
├── go.sum                          # 依赖校验和
├── main.go                         # 【生命周期总闸】拦截非 Root、捕获 Ctrl+C、路由分发
├── Makefile                        # 【构建契约】CGO_ENABLED=0 静态编译
├── README.md                       # 项目说明与快速开始
├── docs/
│   └── cheatsheet.txt              # tldr 命令的速查内容
├── cmd/                            # 【接口/路由层】基于 Cobra 框架
│   ├── root.go                     # 根命令：解析全局 Flag (--debug)，初始化 Logger
│   ├── run.go                      # CLI 模式：一键自动化执行 (CI/CD 友好)
│   ├── wizard.go                   # Wizard 模式：分步交互式引导 (人类用户友好)
│   ├── search.go                   # 辅助模式：模糊查询内核函数
│   ├── status.go                   # 状态模式：打印当前 tracefs 状态
│   ├── tui.go                      # 监控模式：启动 TUI 实时数据流大屏
│   └── tldr.go                     # 帮助模式：打印常见用法速查
└── internal/                       # 【领域核心层】所有的血肉和灵魂
    ├── ftrace/                     # 负责所有 /sys/kernel/debug/tracing 的原子化 I/O
    │   ├── discover.go             # 自动探测挂载点（优先 tracing，降级 debug/tracing）
    │   ├── engine.go               # Engine 结构体：SetTracer/SetFilter/Enable/Clear
    │   └── pipe.go                 # ReadTracePipe()：返回 <-chan string 并发读取
    ├── search/                     # 负责 OOM 防御级的流式匹配
    │   └── scanner.go              # bufio.Scanner 流式读取 available_filter_functions
    ├── wizard/                     # 负责 Step-by-Step 交互流向导
    │   └── prompter.go             # 封装 promptui，实现带校验的回调函数
    └── ui/                         # 负责终端图形化渲染
        └── dashboard.go            # tview 布局 + Channel 消费更新逻辑
```

### 文件职责详解

| 文件 | 职责 | 关键函数 |
|------|------|----------|
| `main.go` | 入口：权限检查 + 信号拦截 + 命令路由 | `main()`, `SetupCloseHandler()` |
| `cmd/root.go` | 根命令注册，全局参数解析 | `Execute()`, `init()` |
| `cmd/run.go` | One-Shot 模式实现 | `runCmd` |
| `cmd/wizard.go` | 五步交互向导实现 | `wizardCmd` |
| `cmd/search.go` | 函数搜索命令 | `searchCmd` |
| `cmd/status.go` | 状态查询命令 | `statusCmd` |
| `cmd/tui.go` | TUI 启动命令 | `tuiCmd` |
| `cmd/tldr.go` | 速查帮助命令 | `tldrCmd` |
| `internal/ftrace/discover.go` | tracefs 路径探测 | `FindTracefs()` |
| `internal/ftrace/engine.go` | 核心引擎：所有 tracefs 读写 | `Engine`, `SetTracer()`, `SetFilter()`, `Enable()`, `Clear()`, `WriteTraceFile()`, `ReadTraceFile()` |
| `internal/ftrace/pipe.go` | trace_pipe 流式读取 | `ReadTracePipe(ctx)` |
| `internal/search/scanner.go` | 流式函数搜索 | `FastScan(pattern, maxLimit)` |
| `internal/wizard/prompter.go` | 交互式问答 | `AskTracer()`, `AskFilter()`, `AskConfirm()` |
| `internal/ui/dashboard.go` | TUI 大屏渲染 | `RunDashboard(engine)` |

---

## 2. 严格的输入与输出规范及错误边界防护

### 2.1 核心命令 I/O 定义与边界控制

#### `litetrace run` (一键执行模式)

**输入 (Flags)**：
| Flag | 类型 | 必填 | 默认值 | 校验规则 |
|------|------|------|--------|----------|
| `--tracer` | string | ✅ | - | 枚举值严格校验：`function`, `function_graph` |
| `--filter` | string | ⚠️ | 空 | 若为空，打印 Warning 提示系统可能负载尖峰 |
| `--duration` | string | ❌ | 10s | `time.ParseDuration` 解析，拒绝负数或超大值 (如 999h) |
| `--output` | filepath | ✅ | - | 必须预检目录是否有写权限 |

**输出 (Stdout/Stderr)**：
- 标准输出：进度提示，如 `[*] Tracing started for 10s...`, `[+] Trace saved to /tmp/trace.txt`
- 错误输出：所有错误必须通过 `fmt.Fprintf(os.Stderr, ...)` 输出
- 副作用：以 `0644` 权限覆盖写入 `--output` 文件，若磁盘空间不足必须捕获 `syscall.ENOSPC`

#### `litetrace status` (状态查询)

**输入**：无

**输出 (Stdout)**：必须以极具可读性的 ASCII 表格输出：
```
=========================================
[ Ftrace Kernel Subsystem Status ]
=========================================
- Engine Status : 🟢 RUNNING (tracing_on = 1)
- Current Tracer: function_graph
- Active Filters: vfs_read, vfs_write
- Buffer Size   : 1408 KB (Per CPU)
=========================================
```

#### `litetrace search <pattern>` (极速搜索)

**输入 (Args)**：字符串 `<pattern>`（必须，支持简单正则如 `^vfs_` 或精确匹配）

**输出 (Stdout)**：
- 最多输出匹配的前 **100 行**（防刷屏）
- 每行一个函数名
- 如果查询耗时超过 **2 秒**，自动提示 "The regex is too complex..."

---

### 2.2 内部方法 I/O 契约与重试策略

#### `func (e *Engine) WriteTraceFile(filename string, content string, appendMode bool) error`

**输入**：
- 目标文件名（如 `set_ftrace_filter`）
- 欲写入的内容
- 是否追加模式

**底层逻辑**：
- 针对 `set_ftrace_filter`，内核解析器对格式有要求
- 如果是多个函数，必须用换行符 `\n` 分隔

**输出**：
- 成功返回 `nil`
- 若遇到 `syscall.EINVAL`（无效参数，通常是因为写入了不存在的函数名），必须返回包装后的详尽错误：
  ```go
  fmt.Errorf("invalid filter function '%s': %w", content, err)
  ```

#### `func ReadTracePipe(ctx context.Context) <-chan string`

**输入**：注入一个生命周期控制的 `context.Context`，用于随时优雅终止后台 Goroutine 的阻塞式读取

**输出**：一个只读的字符串 Channel `<-chan string`

**极度重要**：
- 此函数内部必须使用 Goroutine 死循环读取
- 当 `ctx` 被 Cancel 或进程收到中断信号时，必须主动关闭文件描述符以解除阻塞
- 必须 `close(channel)` 以通知消费端

---

## 3. 核心状态机运转：配置落地、生效与结果导出

### 3.1 配置的生效逻辑与安全时序（⚠️ 不可乱序！）

**6 步严格写入时序**：以下顺序绝对不能颠倒，否则可能导致内核 Panic 或数据污染！

当用户通过 Wizard 向导或 CLI 完成了配置后，程序**必须**执行以下严格的写入时序：

| 步骤 | 操作 | 文件 | 写入内容 | 目的 |
|------|------|------|----------|------|
| **1. 绝对关停** | 首先关停追踪 | `tracing_on` | `0` | 确保修改配置前追踪器处于静默状态 |
| **2. 清空旧怨** | 清空旧过滤器 | `set_ftrace_filter` | `""` (空字符串) | 使用 `os.O_TRUNC` 覆盖模式，防止脏过滤器残留 |
| **3. 清空缓冲** | 刷新 Ring Buffer | `trace` | `""` 或 `\n` | 强制清空残留的旧日志快照 |
| **4. 设定范围** | 应用过滤器 | `set_ftrace_filter` | 用户指定的函数名 | 如果写入失败必须立刻回滚，不可带着空过滤器进入下一步 |
| **5. 注入灵魂** | 设置追踪器 | `current_tracer` | `function` 或 `function_graph` | 选择追踪模式 |
| **6. 拉下电闸** | 开始追踪 | `tracing_on` | `1` | **不归路**：内核立刻开始产生中断日志 |

### 3.2 跟踪结果的导出落地与清场（⚠️ 安全撤离顺序）

当跟踪结束（时间到达，或用户按下 Ctrl+C），程序**必须**按以下顺序安全落地结果并撤离：

| 步骤 | 操作 | 文件 | 写入内容 | 目的 |
|------|------|------|----------|------|
| **1. 推上电闸** | 停止内核写入 | `tracing_on` | `0` | **第一时间**让内核停止向 Ring Buffer 灌入新日志 |
| **2. 重置追踪器** | 中和探针 | `current_tracer` | `nop` | 彻底解除内核探针开销 |
| **3. 读取缓冲区** | 快照导出 | `trace` → 用户文件 | `io.Copy(dst, src)` | 将 trace 文件内容原封不动灌入用户指定的本地文件 |
| **4. 彻底清场** | 清空过滤器 | `set_ftrace_filter` | `""` | 做到"挥一挥衣袖，不带走一片云彩" |

---

## 4. 极致防御设计与核心防爆骨架

### 🛡️ 三根防御钢钉

**第一根钢钉**：生命周期防爆栈与资源泄漏兜底（Graceful Shutdown）：生命周期防爆栈与资源泄漏兜底（Graceful Shutdown）

在 `main.go` 中，**绝对不能省略**这段代码。如果省略，一旦程序崩溃退出，内核将永远处于 tracing 状态，服务器 I/O 性能将断崖式下跌。

```go
// internal/ftrace/engine.go & main.go
func SetupCloseHandler(ctx context.Context, cancel context.CancelFunc, engine *ftrace.Engine) {
    c := make(chan os.Signal, 2)
    // 同时捕获 Ctrl+C (SIGINT) 和 kill -15 (SIGTERM)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c // 阻塞等待操作系统的死神信号
        fmt.Println("\n[!] Emergency Signal detected (Ctrl+C). Safely shutting down ftrace...")
        
        // 1. 通知所有后台 Goroutine 停止工作
        cancel() 
        
        // 2. 救命稻草：强制重置内核态，屏蔽一切错误坚决执行
        _ = engine.WriteTraceFile("tracing_on", "0", false)
        _ = engine.WriteTraceFile("current_tracer", "nop", false)
        _ = engine.WriteTraceFile("set_ftrace_filter", "", false)
        
        fmt.Println("[+] Kernel state restored to safety. Exiting.")
        os.Exit(0)
    }()
}
```

---

第二根钢钉：OOM 防御级大文件检索策略（internal/search）

`available_filter_functions` 文件通常高达几兆字节、包含五万甚至十万行以上。如果使用 `ioutil.ReadFile` 将其一次性读入内存再进行 `strings.Split`，在内存仅有 512MB 的嵌入式板子上将直接触发 OOM Killer。

```go
// internal/search/scanner.go
// FastScan 提供零内存泄漏的流式高压匹配
func FastScan(tracefsPath string, pattern string, maxLimit int) ([]string, error) {
    file, err := os.Open(filepath.Join(tracefsPath, "available_filter_functions"))
    if err != nil {
        return nil, fmt.Errorf("failed to open available_filter_functions: %w", err)
    }
    defer file.Close()

    var results []string
    scanner := bufio.NewScanner(file)
    // 适当增大 Scanner buffer 以防个别长符号越界
    buf := make([]byte, 0, 64*1024)
    scanner.Buffer(buf, 1024*1024)

    // 强制流式读取，按行扫描内存常驻仅在 KB 级别
    for scanner.Scan() {
        line := scanner.Text()
        // 内核格式陷阱：很多函数带有模块名，如 "vfs_read [vmlinux]" 或 "e1000_xmit [e1000e]"
        // 如果带着模块名写入 set_ftrace_filter 会导致解析失败，必须用 Split 剥离
        funcName := strings.SplitN(line, " ", 2)[0] 
        
        if strings.Contains(funcName, pattern) {
            results = append(results, funcName)
            if len(results) >= maxLimit { // 触发断路器，防止刷屏
                break
            }
        }
    }
    return results, scanner.Err()
}
```

---

第三根钢钉：TUI 的跨 Goroutine 并发安全与管道阻塞攻防（internal/ui）

TUI 右侧面板要实时显示内核流。`trace_pipe` 的读取是永远阻塞的（Blocking I/O），如果放在主线程会直接把 GUI 卡死。必须新开线程，并且必须通过 GUI 框架的调度代理（如 tview 的 `QueueUpdateDraw`）进行跨线程刷新，否则会触发竞态条件引发 Panic。

```go
// internal/ui/dashboard.go
func startStream(ctx context.Context, app *tview.Application, textView *tview.TextView, pipeChan <-chan string) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                // 上下文被取消，安全撤离协程
                return
            case line, ok := <-pipeChan:
                if !ok {
                    // 管道已被关闭
                    return
                }
                // 绝对不能直接操作 textView.Write! 必须放入 UI 主线程调度队列
                app.QueueUpdateDraw(func() {
                    // 维持缓冲区大小限制，防止内存爆掉
                    if textView.GetTextLength() > 50000 {
                        textView.Clear()
                    }
                    fmt.Fprint(textView, line+"\n")
                })
            }
        }
    }()
}
```

---

## 5. 交互流极深解析与用户心理学防护（The Wizard Flow）

当用户执行 `litetrace wizard` 时，我们不仅要实现功能，还要实现极佳的开发者体验 (DX)。通过精美的终端交互流，抹平普通开发者和内核调试之间的鸿沟。

### 完整状态机剧本

#### 【Welcome Screen】
清理终端屏幕，打印酷炫的 ASCII Art Logo 和当前的系统内核版本：
```
  _                   _       ____             _       
 | |    ___  __ _  __| | ___ |  _ \ ___  _ __ | |_ ___ 
 | |   / _ \/ _` |/ _` |/ _ \| |_) / _ \| '_ \| __/ __|
 | |__|  __/ (_| | (_| |  __/|  __/ (_) | | | | |_\__ \
 |_____\___|\__,_|\__,_|\___||_|   \___/|_| |_|\__|___/
                                                       
Kernel: 6.1.0-12-amd64 | Tracefs: /sys/kernel/tracing
```

#### 【问答 1：追踪器选择】
用 `promptui.Select` 提供选项：
```
[?] Choose your tracer:
  ❯ function        - Trace kernel function entries (lightweight)
    function_graph  - Trace function calls and returns (detailed)
    nop             - Disable tracing (safe mode)
```

#### 【问答 2：高风险过滤器确认】
用 `promptui.Prompt` 要求用户输入关键字：
```
[?] Enter a kernel function to filter (e.g., vfs_read), or leave empty for all: _
```

#### 【隐式闭环校验机制】⚠️ 重中之重
拿着用户输入的关键字去调用 `search.FastScan()`。如果在内核函数表里找不到（比如用户手误把 `vfs_read` 打成了 `vfc_read`），**必须拦截该流程**并用红色字体提示用户：
```
❌ Function 'vfc_read' not found in kernel symbols. Please try again.
```
并重新弹回输入框，直到输入合法值为止。这**阻断了底层写入崩溃的可能性**。

#### 【问答 3：战前确认】
提示：
```
Target set to monitor [vfs_read] using [function_graph]. Ready to start tracing? (Y/n)
```

#### 【最终分支：表现层分流】
一旦用户按下 Y，底层立刻向 `tracing_on` 写 1。随后立刻出现最后一道选择题：
```
How do you want to view the kernel data?
  ❯ [1] Enter TUI Dashboard      - Open real-time monitoring screen
    [2] Run silently and Export  - Run 10s and save to file
    [3] Run in Background        - Detach and continue in background
```

---

## 6. 五阶段物理机严格验收指南

代码研发完成后，在一台未经任何特殊配置的纯净版 Debian 12 裸机（建议使用虚拟机或云服务器以免内核卡死影响物理机）上，**必须能不差分毫地通过以下全链路测试**：

### Phase 1：权限阻断与引擎自检

| 项目 | 内容 |
|------|------|
| **执行** | 作为普通用户（非 root）在终端运行 `go run main.go status` |
| **预期** | 程序在 **10 毫秒内**立刻输出红色的 `🚨 Fatal: Root privileges required` 并以**状态码 1**退出，绝不接触任何 VFS 系统 |

---

### Phase 2：流式搜索与内存稳定性

| 项目 | 内容 |
|------|------|
| **执行** | `sudo ./litetrace search "^sys_"` |
| **预期** | 在**不到一秒**的时间内打印出系统调用函数名（如 `sys_read`, `sys_write` 等），列表被平滑裁剪至 **100 以内** |
| **内存限制** | 使用 `top` 或 `htop` 观察，执行期间 `litetrace` 的内存占用**不得超过 15MB** |

---

### Phase 3：向导写入与状态机一致性

| 项目 | 内容 |
|------|------|
| **执行** | 运行 `sudo ./litetrace wizard`，按照向导选择 `function_graph` 并搜索 `tcp_recvmsg`，确认开启追踪然后按 Ctrl+C 退出 |
| **预期 1** | 另开一个终端执行 `cat /sys/kernel/debug/tracing/set_ftrace_filter`，必须看到正确写入了 `tcp_recvmsg` |
| **预期 2** | 再执行 `cat /sys/kernel/debug/tracing/tracing_on`，必须看到值为 `0`（安全关闭） |

---

### Phase 4：监控并发与 GUI 防抖

| 项目 | 内容 |
|------|------|
| **执行** | `sudo ./litetrace tui` |
| **预期** | 界面必须分为左右两块。右侧面板应如瀑布般刷出 `trace_pipe` 的实时流。此时在系统里另开终端执行 `curl https://www.google.com` 或大量的网络请求，TUI 界面应展现出大量网络底层的调用树（如 `tcp_v4_rcv`），并且即使数据流再大，**键盘响应和退出指令依然非常灵敏**（证明 Goroutine 未阻塞主线程） |

---

### Phase 5：导出落地与洁癖级恢复

| 项目 | 内容 |
|------|------|
| **执行** | `sudo ./litetrace run --tracer function --filter "vfs_write" --duration 5s --output my_trace_dump.txt` |
| **预期 1** | 终端显示倒计时进度条，静默 5 秒后安全退出 |
| **预期 2** | 执行 `head -n 20 my_trace_dump.txt`，里面必须全部包含真实的内核追踪日志格式 |
| **预期 3** | 随后检查 `current_tracer` 是否变回了 `nop`，`set_ftrace_filter` 是否已被清空 |

---

## 7. 致 AI 执行系统的最终封印与死线契约

如果你是执行代码生成的 AI，请在你的上下文中锁定以下铁律。**违反任意一条，都将被判定为任务失败**：

### 🔒 铁律一：绝对原生，严禁取巧
- **不准**自行发明外部依赖库（只允许 cobra、tview、promptui）
- **严禁**使用 `os/exec` 包去执行如 `bash -c "echo 1 > tracing_on"` 这类命令
- **一切**必须使用 Go 原生库 `os.OpenFile` 配合 `WriteString`
- 我们是在开发工具，不是在写运维 Shell 脚本的包裹器！

### 🔒 铁律二：极端的错误链路追踪
- 所有的内部错误必须被层层包裹返回
- 比如遇到写入失败，不能只打印 `"error writing"`，必须使用 `%w`：
  ```go
  fmt.Errorf("failed to apply filter %s: %w", filter, err)
  ```

### 🔒 铁律三：注意文件描述符泄露与权限掩码
- 在任何时候操作 tracefs 里的文件，必须在同一函数体内紧跟 `defer file.Close()`
- 切勿给系统暴露安全风险

### 🔒 铁律四：强行注释与可维护性
- 所有位于 `internal/ftrace` 下对 `/sys/...` 文件进行的原子化操作，其函数上方**必须**拥有清晰的中文注释
- 明确说明它映射到哪个底层内核动作

---

## 8. 依赖管理与构建契约

### 8.1 Go 模块依赖

```bash
go mod init github.com/malossov/lite-tracer-mygo

# 核心依赖（仅 3 个外部库）
go get github.com/spf13/cobra@latest        # CLI 命令树框架
go get github.com/rivo/tview@latest         # TUI 组件库
go get github.com/gdamore/tcell/v2@latest   # tview 底层终端控制
go get github.com/manifoldco/promptui@latest # 交互式问答（Wizard 用）
```

### 8.2 Makefile 构建契约

```makefile
.PHONY: build clean test run

# 强制静态编译，确保脱离 Glibc 依赖
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o litetrace .

clean:
	rm -f litetrace
	go clean

test:
	go test ./internal/... -v

run:
	go run main.go --help
```

### 8.3 编译产物要求

| 指标 | 要求 |
|------|------|
| **编译命令** | `CGO_ENABLED=0 go build -ldflags="-s -w"` |
| **文件大小** | 10MB - 15MB 之间 |
| **依赖** | 零动态链接库依赖（`ldd litetrace` 应显示 `not a dynamic executable`） |
| **跨平台** | 可在任意 Linux 发行版（Debian/Ubuntu/CentOS/Alpine）直接运行 |

---

## 9. 项目完备性复核（对照原始作业）

| 作业要求 | V-Final 实现方案 | 状态 |
|----------|------------------|------|
| 实现一个简易版 trace-cmd 工具 | 纯 Go 实现，零外部二进制依赖 | ✅ |
| 支持打开 function 跟踪 | `Tracer: function`，写入 `current_tracer` | ✅ |
| 支持 function 过滤 | `search` 预览 + `SetFilter` 写入 `set_ftrace_filter` | ✅ |
| 动态开启和关闭跟踪 | `tracing_on` 写 1/0 + defer 安全清理 | ✅ |
| 支持查看当前配置状态 | `status` 子命令 + TUI 左侧面板实时展示 | ✅ |
| 支持导出跟踪结果 | 读取 `trace` 文件落地到本地文件 | ✅ |
| 打印 trace 中的跟踪内容 | 并发读取 `trace_pipe` + TUI 实时流 | ✅ |

**所有作业要求均以最高标准（Wizard + TUI + CLI + 防御式编程）超额完成！**

---

## 附录 A：tracefs 关键文件映射表

| 文件路径 | 读写权限 | 用途 | 示例值 |
|----------|----------|------|--------|
| `current_tracer` | 读/写 | 设置/查看当前追踪器类型 | `function`, `function_graph`, `nop` |
| `tracing_on` | 读/写 | 开启/关闭追踪（总开关） | `1` (开), `0` (关) |
| `set_ftrace_filter` | 读/写 | 设置函数过滤器 | `vfs_read`, `tcp_*` |
| `trace` | 只读 | 读取追踪快照（非消费型） | 内核调用栈日志 |
| `trace_pipe` | 只读 | 实时流式读取追踪（消费型） | 持续输出的日志流 |
| `available_tracers` | 只读 | 查看系统支持的追踪器列表 | `function`, `function_graph` |
| `available_filter_functions` | 只读 | 查看所有可过滤的内核函数 | 50000+ 行函数名列表 |
| `trace_clock` | 读/写 | 设置时间戳时钟源 | `local`, `global`, `counter` |
| `buffer_size_kb` | 读/写 | 设置每 CPU 缓冲区大小（KB） | `1408` |

---

## 附录 B：常见内核函数过滤模式示例

| 场景 | 过滤模式 | 说明 |
|------|----------|------|
| 所有 VFS 调用 | `vfs_*` | 匹配所有 vfs_ 开头的函数 |
| 特定系统调用 | `sys_read`, `sys_write` | 精确匹配单个函数 |
| 网络子系统 | `tcp_*`, `udp_*` | 匹配 TCP/UDP 协议栈函数 |
| 块设备 I/O | `blk_*`, `__make_request` | 匹配块设备层函数 |
| 进程调度 | `schedule`, `try_to_wake_up` | 匹配调度器相关函数 |
| 内存管理 | `*page*`, `*alloc*` | 匹配内存分配相关函数 |

---

## 附录 C：故障排查速查表

| 现象 | 可能原因 | 解决方案 |
|------|----------|----------|
| `Permission denied` | 非 root 用户运行 | 使用 `sudo` 或以 root 身份运行 |
| `No such file or directory` | tracefs 未挂载 | 执行 `mount -t tracefs nodev /sys/kernel/tracing` |
| `Invalid argument` 写入 filter 失败 | 函数名不存在或带有模块后缀 | 使用 `search` 命令确认函数名，剥离 `[module]` 后缀 |
| TUI 卡死 | 主线程被阻塞 | 确保 `trace_pipe` 读取在独立 Goroutine 中 |
| 内存占用过高 | 使用了 `ioutil.ReadFile` | 改用 `bufio.Scanner` 流式读取 |
| 退出后系统变慢 | 未正确清理 tracing 状态 | 确保 `SetupCloseHandler` 正确拦截信号 |

---

**文档版本**：V-Final Ultimate Edition  
**最后更新**：2026 年 3 月 11 日  
**适用项目**：github.com/malossov/lite-tracer-mygo  
**状态**：🔒 冻结为最终契约，任何修改需创建新版本

---

*本工程蓝图至此宣告完毕。所有宏观建筑与微观防御手段已全部就绪。拿起这份契约，展开真正的底层编码风暴吧。*
