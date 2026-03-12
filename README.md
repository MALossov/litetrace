# Litetrace

一个使用纯 Go 语言实现的轻量级 Linux ftrace 内核追踪工具。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 简介

Litetrace 是一个零依赖、单二进制文件的 Linux 内核追踪工具，直接操作 tracefs 文件系统，无需依赖外部二进制文件（如 trace-cmd）。它提供了内核函数追踪、实时监控、交互式向导等功能，适合开发者和系统管理员进行内核性能分析和调试。

## 特性

- **纯 Go 实现**：单二进制文件，零外部依赖
- **直接操作 tracefs**：无需额外内核模块
- **多种追踪器支持**：支持 `function` 和 `function_graph` 两种追踪器
- **实时监控**：TUI 界面实时显示追踪数据
- **交互式向导**：引导式操作，适合新手使用
- **函数搜索**：支持正则表达式搜索内核函数
- **调试模式**：详细的 tracefs 文件操作日志
- **优雅关闭**：完善的错误处理和信号处理机制

## 安装

### 前置要求

- Linux 系统（内核版本 2.6.27+）
- Go 1.21 或更高版本
- root 权限

### 从源码编译

```bash
# 克隆仓库
git clone ssh://git@server.malossov.top:2233/MALoCourse/litetrace.git
cd litetrace/litetrace

# 编译
go build -o litetrace .

# 或者使用 Makefile
make build
```

## 快速开始

```bash
# 查看帮助
sudo ./litetrace --help

# 查看当前 ftrace 状态
sudo ./litetrace status

# 追踪 vfs_read 函数 5 秒
sudo ./litetrace run --tracer function --filter vfs_read --duration 5s --output /tmp/trace.txt

# 后台追踪（不阻塞终端）
sudo ./litetrace background --tracer function --filter vfs_read --duration 30s --output /tmp/trace.txt

# 搜索内核函数
sudo ./litetrace search "^vfs_"

# 启动交互式向导
sudo ./litetrace wizard

# 启动 TUI 实时监控
sudo ./litetrace tui

# 停止后台追踪
sudo ./litetrace terminate
```

## 命令参考

### 全局选项

| 选项 | 说明 |
|------|------|
| `--debug` | 启用调试模式，显示 tracefs 文件操作详情 |

### 子命令

#### `status` - 查看 ftrace 状态

```bash
sudo ./litetrace status
```

显示当前 ftrace 子系统的状态，包括追踪器类型、过滤器、缓冲区大小等。

#### `search` - 搜索内核函数

```bash
sudo ./litetrace search <正则表达式>

# 示例
sudo ./litetrace search "^vfs_"        # 搜索所有 VFS 函数
sudo ./litetrace search "tcp_send"     # 搜索 TCP 发送相关函数
```

#### `run` - 执行追踪

```bash
sudo ./litetrace run [选项]

# 选项
--tracer <类型>      # 追踪器类型: function, function_graph, nop (默认: function)
--filter <函数名>    # 要追踪的函数名，支持通配符
--duration <时长>    # 追踪持续时间，如 5s, 1m, 2h (默认: 10s)
--output <文件路径>  # 输出文件路径 (必需)

# 示例
sudo ./litetrace run --tracer function --filter vfs_read --duration 5s --output /tmp/trace.txt
sudo ./litetrace run --tracer function_graph --filter "tcp_*" --duration 30s --output /tmp/tcp_trace.txt
```

#### `wizard` - 交互式向导

```bash
sudo ./litetrace wizard
```

启动交互式向导，引导用户完成追踪配置，适合新手使用。

#### `tui` - 实时 TUI 监控

```bash
sudo ./litetrace tui
```

启动终端图形界面，实时查看追踪数据。

**快捷键：**
- `q` / `Ctrl+C` - 退出并清理
- `w` - 打开配置向导
- `p` - 暂停/恢复追踪
- `a` - 切换自动滚动（开启时自动跟踪最新内容）
- `s` - 保存当前追踪到文件
- `c` - 清空追踪缓冲区
- `↑/↓` - 滚动查看

#### `background` - 后台追踪

```bash
sudo ./litetrace background [选项]

# 选项
--tracer <类型>      # 追踪器类型: function, function_graph, nop (默认: function)
--filter <函数名>    # 要追踪的函数名，支持通配符
--duration <时长>    # 追踪持续时间，如 5s, 1m, 2h (默认: 无限期)
--output <文件路径>  # 输出文件路径 (默认: /tmp/litetrace_background_<时间戳>.txt)

# 示例
sudo ./litetrace background --tracer function --filter vfs_read --duration 30s --output /tmp/trace.txt
sudo ./litetrace background --tracer function_graph --filter "tcp_*"  # 无限期运行，手动停止
```

在后台运行追踪任务，立即返回不阻塞终端。适合长时间追踪或需要在追踪期间继续使用shell的场景。

#### `terminate` - 停止后台追踪

```bash
sudo ./litetrace terminate
```

停止正在运行的后台追踪进程，导出结果并清理ftrace状态。

#### `tldr` - 快速帮助

```bash
sudo ./litetrace tldr
```

显示简明的使用帮助。

## 追踪器类型

| 类型 | 说明 |
|------|------|
| `function` | 追踪函数入口，轻量级，性能开销小 |
| `function_graph` | 追踪函数调用和返回，显示完整的调用链，开销较大 |
| `nop` | 禁用追踪，用于安全模式或重置状态 |

## 过滤器语法

```bash
# 精确匹配
vfs_read

# 通配符匹配
vfs_*                 # 所有以 vfs_ 开头的函数
*read*                # 包含 read 的函数

# 多个函数
vfs_read,vfs_write    # 同时追踪 vfs_read 和 vfs_write

# 所有函数（不指定 --filter，高开销）
```

## 示例场景

### 场景1：追踪文件读取操作

```bash
sudo ./litetrace run --tracer function --filter vfs_read --duration 10s --output /tmp/vfs_read.txt
```

### 场景2：追踪 TCP 网络操作

```bash
sudo ./litetrace run --tracer function_graph --filter "tcp_*" --duration 30s --output /tmp/tcp.txt
```

### 场景3：实时监控所有系统调用

```bash
sudo ./litetrace tui
```

### 场景4：查找并追踪特定函数

```bash
sudo ./litetrace search "sched_"
sudo ./litetrace run --filter sched_switch --duration 5s --output /tmp/sched.txt
```

## 调试模式

使用 `--debug` 选项查看详细的 tracefs 文件操作：

```bash
sudo ./litetrace --debug run --tracer function --filter vfs_read --duration 5s --output /tmp/trace.txt
```

调试信息以黄色输出，有助于排查问题。

## 项目结构

```
litetrace/
├── main.go                 # 程序入口
├── go.mod                  # Go 模块定义
├── Makefile                # 构建脚本
├── cmd/                    # 命令实现
│   ├── root.go            # 根命令
│   ├── run.go             # run 子命令
│   ├── wizard.go          # wizard 子命令
│   ├── search.go          # search 子命令
│   ├── status.go          # status 子命令
│   ├── tui.go             # tui 子命令
│   └── tldr.go            # tldr 子命令
├── internal/              # 内部包
│   ├── ftrace/           # ftrace 核心操作
│   ├── search/           # 函数搜索
│   ├── ui/               # TUI 界面
│   └── wizard/           # 交互式向导
└── docs/                 # 文档
    ├── cheatsheet.txt    # 用户手册
    └── DEVELOPER.md      # 开发者文档
```

## 开发

```bash
# 运行测试
make test

# 构建
make build

# 清理
make clean
```

更多信息请参考 [DEVELOPER.md](litetrace/docs/DEVELOPER.md)。

## 故障排除

### Permission denied
使用 `sudo` 运行命令。

### tracefs not found
手动挂载 tracefs：
```bash
mount -t tracefs nodev /sys/kernel/tracing
```

### Invalid filter / Function not found
使用 `search` 命令查找有效的函数名：
```bash
sudo ./litetrace search "^vfs_"
```

### Trace buffer is empty
- 检查追踪器类型是否正确
- 检查过滤器是否匹配到函数
- 增加追踪时间

## 相关资源

- [Linux ftrace 文档](https://www.kernel.org/doc/Documentation/trace/ftrace.rst)
- [trace-cmd 工具](https://trace-cmd.org/)
- [KernelShark](https://kernelshark.org/)

## 许可证

MIT License
