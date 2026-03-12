# 提供以下文件
- litetrace-src: 源代码目录
- litetrace: 可执行文件，静态编译，无需依赖任何库
- litetrace_demo.cast: 演示 cast 文件,使用`asciinema play ./litetrace_demo.cast`播放
- litetrace_demo.gif: 演示截图，展示 litetrace 的界面和功能

## 源码目录litetrace-src关键文档

```bash
litetrace-src/
|-- Makefile     # 编译脚本
|-- README.md    # 项目介绍
|-- cmd/         # go命令行工具目录，包含 litetrace 命令子命令
|-- docs/        # 文档目录
|   |-- DEVELOPER.md     # 开发者文档
|   |-- cheatsheet.txt   # 快速参考
|-- go.mod
|-- go.sum
|-- internal/     # 内部代码目录，包含核心功能实现
|-- litetrace     # 可执行文件
|-- main.go     # 主程序入口
`-- scripts/      # 脚本目录，包含演示脚本
    `-- demo_script.sh  # 演示脚本
```