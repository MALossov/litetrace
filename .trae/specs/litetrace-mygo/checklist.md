# Litetrace-MyGo 开发验收清单

## 项目初始化
- [x] go.mod 文件创建正确，模块名为 github.com/malossov/lite-tracer-mygo
- [x] 核心依赖已安装 (cobra, tview, tcell, promptui)
- [x] 目录结构符合工程规范

## internal/ftrace 核心引擎
- [x] discover.go 实现 tracefs 路径探测功能
- [x] engine.go 实现 SetTracer 方法
- [x] engine.go 实现 SetFilter 方法
- [x] engine.go 实现 Enable/Disable 方法
- [x] engine.go 实现 Clear 方法
- [x] engine.go 实现 ReadTraceFile 方法
- [x] engine.go 实现 WriteTraceFile 方法
- [x] pipe.go 实现 ReadTracePipe 方法，支持 context 取消

## internal/search 流式搜索
- [x] scanner.go 实现 FastScan 方法
- [x] 支持流式读取大文件
- [x] 支持正则匹配
- [x] 限制返回结果数量（最多 100 条）

## internal/wizard 交互向导
- [x] prompter.go 实现 AskTracer 方法
- [x] prompter.go 实现 AskFilter 方法
- [x] prompter.go 实现 AskConfirm 方法

## internal/ui TUI 大屏
- [x] dashboard.go 实现 RunDashboard 方法
- [x] 正确处理并发更新
- [x] 缓冲区大小限制

## cmd/ 命令行接口
- [x] root.go 正确注册所有子命令
- [x] run.go 实现 CLI 一键执行模式
- [x] wizard.go 实现交互式向导
- [x] search.go 实现函数搜索
- [x] status.go 实现状态查询
- [x] tui.go 启动 TUI 监控
- [x] tldr.go 打印速查帮助

## main.go 入口
- [x] 实现 root 权限检查
- [x] 实现信号拦截 (Ctrl+C, SIGTERM)
- [x] 实现 Graceful Shutdown
- [x] 命令路由正确

## 构建与测试
- [x] Makefile 构建成功
- [x] 编译产物大小在 10-15MB 之间
- [x] 无动态链接库依赖
- [x] 单元测试通过
