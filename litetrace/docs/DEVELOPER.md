================================================================================
                         LITETRACE 开发者文档
================================================================================

目录
----
1. 项目概述
2. 项目架构
3. 模块说明
4. 开发规范
5. API 参考
6. 贡献指南

================================================================================
1. 项目概述
================================================================================

1.1 项目简介
------------
Litetrace 是一个使用纯 Go 语言实现的 Linux ftrace 内核追踪工具。它直接操作
tracefs 文件系统，无需依赖外部二进制文件（如 trace-cmd），提供轻量级、零依赖
的内核函数追踪能力。

1.2 核心特性
------------
- 纯 Go 实现，单二进制文件，零外部依赖
- 直接操作 tracefs，无需额外内核模块
- 支持 function 和 function_graph 两种追踪器
- 提供 TUI 实时界面和交互式向导
- 完善的错误处理和优雅关闭机制
- 支持调试模式，便于问题排查

1.3 技术栈
----------
- 语言: Go 1.21+
- CLI 框架: spf13/cobra
- TUI 框架: rivo/tview + gdamore/tcell
- 交互式提示: manifoldco/promptui
- 测试: Go 标准测试框架

================================================================================
2. 项目架构
================================================================================

2.1 目录结构
------------
```
litetrace/
├── main.go                 # 程序入口，信号处理
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖校验
├── cmd/                    # 命令实现
│   ├── root.go            # 根命令和全局标志
│   ├── run.go             # run 子命令
│   ├── wizard.go          # wizard 子命令
│   ├── search.go          # search 子命令
│   ├── status.go          # status 子命令
│   ├── tui.go             # tui 子命令
│   └── tldr.go            # tldr 子命令
├── internal/              # 内部包
│   ├── ftrace/           # ftrace 核心操作
│   │   ├── engine.go     # 引擎实现
│   │   ├── engine_test.go
│   │   ├── discover.go   # tracefs 发现
│   │   ├── discover_test.go
│   │   └── pipe.go       # trace_pipe 读取
│   ├── search/           # 函数搜索
│   │   ├── scanner.go    # 扫描实现
│   │   └── scanner_test.go
│   ├── ui/               # TUI 界面
│   │   └── dashboard.go  # 仪表盘实现
│   └── wizard/           # 交互式向导
│       └── prompter.go   # 提示实现
├── docs/                 # 文档
│   ├── cheatsheet.txt    # 用户手册
│   └── DEVELOPER.md      # 开发者文档
└── README.md             # 项目说明
```

2.2 架构图
----------
```
┌─────────────────────────────────────────────────────────────┐
│                        用户界面层                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │   CLI    │ │  Wizard  │ │   TUI    │ │  Search  │       │
│  │  (cobra) │ │(promptui)│ │ (tview)  │ │(scanner) │       │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘       │
└───────┼────────────┼────────────┼────────────┼─────────────┘
        │            │            │            │
        └────────────┴──────┬─────┴────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────┐
│                      业务逻辑层                              │
│                           │                                   │
│                    ┌──────▼──────┐                           │
│                    │   Engine    │                           │
│                    │  (ftrace)   │                           │
│                    └──────┬──────┘                           │
└───────────────────────────┼─────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────┐
│                      系统接口层                              │
│                           │                                   │
│              ┌────────────┼────────────┐                     │
│              ▼            ▼            ▼                     │
│        ┌─────────┐  ┌─────────┐  ┌─────────┐                │
│        │tracefs  │  │trace_pipe│  │ signals │                │
│        │(sysfs)  │  │(stream) │  │(os/signal)│               │
│        └─────────┘  └─────────┘  └─────────┘                │
└─────────────────────────────────────────────────────────────┘
```

================================================================================
3. 模块说明
================================================================================

3.1 cmd 包
-----------
位置: `cmd/`
职责: 实现所有 CLI 子命令

主要文件:
- `root.go`: 定义根命令，注册全局标志（--debug）和所有子命令
- `run.go`: 实现一次性追踪执行，支持 --tracer, --filter, --duration, --output
- `wizard.go`: 实现交互式向导，引导用户完成配置
- `search.go`: 实现函数搜索，支持正则表达式
- `status.go`: 显示当前 ftrace 状态
- `tui.go`: 启动 TUI 实时监控界面
- `tldr.go`: 显示快速帮助

3.2 internal/ftrace 包
----------------------
位置: `internal/ftrace/`
职责: 核心 ftrace 操作，与 tracefs 交互

主要组件:
- `Engine`: 核心结构体，封装所有 tracefs 操作
  - `WriteTraceFile()`: 写入 tracefs 文件
  - `ReadTraceFile()`: 读取 tracefs 文件
  - `StartTracing()`: 7步启动流程
  - `StopAndExport()`: 4步关闭流程
  - `SafeShutdown()`: 紧急安全关闭
  - `RunWithDuration()`: 定时追踪

- `FindTracefs()`: 自动发现 tracefs 挂载点

关键流程:
```
启动流程 (7步):
1. Disable tracing (tracing_on = 0)
2. Reset tracer to nop
3. Set placeholder filter
4. Clear trace buffer
5. Set actual tracer
6. Set actual filter
7. Enable tracing (tracing_on = 1)

关闭流程 (4步):
1. Disable tracing (tracing_on = 0)
2. Reset tracer to nop
3. Read and export trace
4. Clear filter
```

3.3 internal/search 包
----------------------
位置: `internal/search/`
职责: 内核函数搜索

主要函数:
- `FastScan()`: 快速扫描 available_filter_functions，支持正则匹配
- `ValidateFunction()`: 验证函数名是否有效

特性:
- 使用 bufio.Scanner 流式读取，防止 OOM
- 2秒超时保护
- 最大返回 100 个结果

3.4 internal/ui 包
------------------
位置: `internal/ui/`
职责: TUI 界面实现

主要组件:
- `Dashboard`: TUI 仪表盘
  - 状态面板: 显示当前追踪状态
  - 追踪面板: 实时显示追踪数据
  - 控制面板: 显示键盘快捷键

交互:
- `q/Ctrl+C`: 退出
- `p`: 暂停/恢复
- `s`: 保存到文件
- `c`: 清空缓冲区
- `↑/↓`: 滚动

3.5 internal/wizard 包
----------------------
位置: `internal/wizard/`
职责: 交互式向导

主要函数:
- `AskTracer()`: 选择追踪器
- `AskFilter()`: 输入过滤器
- `AskConfirm()`: 确认操作
- `AskViewMode()`: 选择查看模式

================================================================================
4. 开发规范
================================================================================

4.1 代码风格
------------
- 遵循 Go 官方代码规范 (gofmt, golint)
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查潜在问题
- 注释使用中文，符合项目要求

4.2 命名规范
------------
- 包名: 小写，简短，无下划线 (ftrace, search, ui)
- 接口名: 动词+名词 (Reader, Writer, Scanner)
- 结构体名: 名词 (Engine, Dashboard, Status)
- 方法名: 动词开头 (StartTracing, StopAndExport)
- 私有函数: 小写开头 (debugLog, checkRoot)
- 常量: 驼峰命名 (MaxBufferSize)

4.3 注释规范
------------
- 包注释: 说明包的用途和主要功能
- 类型注释: 说明类型的用途和字段含义
- 函数注释: 说明功能、参数、返回值
- 关键逻辑: 解释复杂算法的思路

示例:
```go
// Engine 提供对 tracefs 的操作接口
// 封装了所有与 ftrace 子系统交互的功能
type Engine struct {
    tracefsPath string // tracefs 挂载路径
    debug       bool   // 是否启用调试模式
}

// StartTracing 启动追踪，执行完整的 7 步初始化流程
// tracer: 追踪器类型 (function/function_graph/nop)
// filter: 函数过滤器，支持通配符
func (e *Engine) StartTracing(tracer, filter string) error {
    // ...
}
```

4.4 错误处理
------------
- 使用 `fmt.Errorf()` 包装错误，添加上下文
- 使用 `%w` 动词保留原始错误
- 关键错误需要记录 debug 日志
- 用户友好的错误信息输出到 stderr

示例:
```go
if err := e.Disable(); err != nil {
    e.debugLog("STOPPING FAILED: Could not disable tracing: %v", err)
    return fmt.Errorf("failed to disable tracing: %w", err)
}
```

4.5 测试规范
------------
- 使用表格驱动测试
- 测试函数名: `TestXxx` 或 `TestEngine_Xxx`
- 使用子测试 (`t.Run()`) 组织相关用例
- 测试覆盖核心功能路径

示例:
```go
func TestEngine_SetTracer(t *testing.T) {
    tests := []struct {
        name    string
        tracer  string
        wantErr bool
    }{
        {"set nop", "nop", false},
        {"set function", "function", false},
        {"set invalid", "invalid", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

4.6 Git 提交规范
----------------
- 提交信息使用中文
- 格式: `<类型>: <描述>`
- 类型: feat(新功能), fix(修复), docs(文档), refactor(重构), test(测试)

示例:
```
feat: 添加 --debug 全局标志
fix: 修复 tracer 设置时的竞态条件
docs: 更新开发者文档
```

================================================================================
5. API 参考
================================================================================

5.1 Engine API
--------------

构造函数:
```go
func NewEngine(tracefsPath string) *Engine
func NewEngineWithDebug(tracefsPath string, debug bool) *Engine
```

核心方法:
```go
// 追踪控制
func (e *Engine) StartTracing(tracer, filter string) error
func (e *Engine) StopAndExport(outputPath string) error
func (e *Engine) RunWithDuration(tracer, filter, outputPath string, duration time.Duration) error
func (e *Engine) SafeShutdown()

// 状态查询
func (e *Engine) GetStatus() (*Status, error)
func (e *Engine) IsEnabled() (bool, error)
func (e *Engine) GetTracer() (string, error)
func (e *Engine) GetFilter() (string, error)

// 配置设置
func (e *Engine) SetTracer(tracer string) error
func (e *Engine) SetFilter(filter string) error
func (e *Engine) Enable() error
func (e *Engine) Disable() error
func (e *Engine) Clear() error

// 调试
func (e *Engine) SetDebug(debug bool)
```

5.2 Search API
--------------
```go
// FastScan 快速扫描函数列表
func FastScan(tracefsPath string, pattern string, maxLimit int) ([]string, error)

// ValidateFunction 验证函数名是否有效
func ValidateFunction(tracefsPath string, funcName string) (bool, error)
```

5.3 Discover API
----------------
```go
// FindTracefs 自动发现 tracefs 挂载点
func FindTracefs() (string, error)
```

================================================================================
6. 贡献指南
================================================================================

6.1 开发环境搭建
----------------
1. 克隆仓库:
   ```bash
   git clone <repository-url>
   cd litetrace/litetrace
   ```

2. 安装依赖:
   ```bash
   go mod download
   ```

3. 验证环境:
   ```bash
   go build -o litetrace .
   ./litetrace --help
   ```

6.2 开发流程
-----------
1. 创建功能分支:
   ```bash
   git checkout -b feat/your-feature
   ```

2. 编写代码和测试

3. 运行测试:
   ```bash
   go test ./... -v
   ```

4. 构建验证:
   ```bash
   go build -o litetrace .
   ```

5. 提交代码:
   ```bash
   git add .
   git commit -m "feat: 添加新功能"
   ```

6.3 测试要求
-----------
- 新功能必须包含单元测试
- 测试覆盖率应达到 70% 以上
- 所有测试必须通过
- 手动测试关键路径

6.4 调试技巧
-----------
1. 使用 --debug 标志查看详细操作:
   ```bash
   sudo ./litetrace --debug run --tracer function --filter vfs_read --duration 5s --output /tmp/trace.txt
   ```

2. 检查 tracefs 状态:
   ```bash
   cat /sys/kernel/tracing/tracing_on
   cat /sys/kernel/tracing/current_tracer
   cat /sys/kernel/tracing/set_ftrace_filter
   ```

3. 使用 dmesg 查看内核消息:
   ```bash
   dmesg | tail -20
   ```

6.5 常见问题
-----------
Q: 编译失败，提示找不到包?
A: 确保在 litetrace/litetrace 目录下执行，且 go.mod 存在

Q: 测试失败，提示权限不足?
A: 部分测试需要 root 权限，使用 sudo 运行

Q: 如何调试 TUI 界面?
A: TUI 使用 tview，可以添加日志输出到文件进行调试

6.6 提交 PR
-----------
1. 确保代码符合规范
2. 所有测试通过
3. 更新相关文档
4. 提交 PR 并描述变更内容

================================================================================
附录
================================================================================

A. 相关资源
-----------
- Go 官方文档: https://golang.org/doc/
- Cobra 文档: https://github.com/spf13/cobra
- tview 文档: https://github.com/rivo/tview
- Linux ftrace 文档: Documentation/trace/ftrace.rst

B. 术语表
---------
- ftrace: Linux 内核内置的追踪框架
- tracefs: ftrace 的虚拟文件系统接口
- tracer: 追踪器，决定追踪的类型（function/function_graph）
- filter: 过滤器，限制追踪的函数范围
- trace_pipe: 流式追踪输出接口

================================================================================
                            文档版本: 1.0
================================================================================
