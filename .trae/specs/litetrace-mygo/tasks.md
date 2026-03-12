# Litetrace-MyGo 开发任务清单

## 任务总览
- [x] Task 1: 初始化项目结构与依赖配置
  - [x] SubTask 1.1: 创建 go.mod 并初始化模块
  - [x] SubTask 1.2: 安装核心依赖 (cobra, tview, promptui)
  - [x] SubTask 1.3: 创建目录结构 (cmd/, internal/)
- [x] Task 2: 实现 internal/ftrace 核心引擎
  - [x] SubTask 2.1: 实现 discover.go - tracefs 路径探测
  - [x] SubTask 2.2: 实现 engine.go - 核心引擎 (SetTracer/SetFilter/Enable/Clear)
  - [x] SubTask 2.3: 实现 pipe.go - trace_pipe 流式读取
- [x] Task 3: 实现 internal/search 流式搜索
  - [x] SubTask 3.1: 实现 scanner.go - bufio.Scanner 流式读取
- [x] Task 4: 实现 internal/wizard 交互向导
  - [x] SubTask 4.1: 实现 prompter.go - 封装 promptui
- [x] Task 5: 实现 internal/ui TUI 大屏
  - [x] SubTask 5.1: 实现 dashboard.go - tview 布局
- [x] Task 6: 实现 cmd/ 命令行接口层
  - [x] SubTask 6.1: 实现 root.go - 根命令
  - [x] SubTask 6.2: 实现 run.go - CLI 一键执行模式
  - [x] SubTask 6.3: 实现 wizard.go - 交互式向导模式
  - [x] SubTask 6.4: 实现 search.go - 函数搜索命令
  - [x] SubTask 6.5: 实现 status.go - 状态查询命令
  - [x] SubTask 6.6: 实现 tui.go - TUI 监控模式
  - [x] SubTask 6.7: 实现 tldr.go - 速查帮助命令
- [x] Task 7: 实现 main.go 入口文件
  - [x] SubTask 7.1: 实现权限检查
  - [x] SubTask 7.2: 实现信号拦截 (SetupCloseHandler)
  - [x] SubTask 7.3: 实现命令路由
- [x] Task 8: 创建 Makefile 构建文件
- [x] Task 9: 运行测试验证

## 任务依赖
- Task 2 依赖 Task 1 (项目初始化)
- Task 3, 4, 5 依赖 Task 2 (ftrace 引擎)
- Task 6 依赖 Task 2, 3, 4, 5 (所有内部模块)
- Task 7 依赖 Task 6 (cmd 层)
- Task 8 依赖 Task 7
- Task 9 依赖 Task 8
