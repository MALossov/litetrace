#!/bin/bash
# Litetrace 完整功能演示脚本
# 用法: asciinema rec /tmp/litetrace_demo.cast -c "sudo ./demo_script.sh"
# 本脚本展示 litetrace 的所有核心功能，包括状态查看、函数搜索、多种追踪模式、TUI界面等

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

PROJECT_DIR="/root/wkSpc/litetrace/litetrace"
cd "$PROJECT_DIR"

# 检查编译
if [ ! -f "./litetrace" ]; then
    echo -e "${YELLOW}编译 litetrace...${NC}"
    go build -o litetrace .
fi

# ==================== 第一部分：欢迎与介绍 ====================
clear
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}║${NC}   ${GREEN}🌶️  Litetrace - 纯 Go 实现的 Linux ftrace 追踪工具${NC}          ${BLUE}║${NC}"
echo -e "${BLUE}║                                                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}特性：${NC}"
echo "  • 零依赖，单二进制文件"
echo "  • 直接操作 tracefs，无需内核模块"
echo "  • 支持 function / function_graph 追踪器"
echo "  • 实时 TUI 监控界面"
echo "  • 交互式向导，适合新手"
echo "  • 后台追踪，不阻塞终端"
echo ""
echo -e "${YELLOW}按 Enter 开始演示...${NC}"
read

# ==================== 第二部分：基础命令 ====================
clear
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  第一部分：基础命令 - status / search / tldr${NC}"
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 1

# 2.1 tldr 快速帮助
echo -e "${CYAN}【1/9】tldr - 快速帮助${NC}"
echo -e "${CYAN}$ ./litetrace tldr${NC}"
sleep 0.5
./litetrace tldr
echo ""
sleep 2

# 2.2 status 查看初始状态
echo -e "${CYAN}【2/9】status - 查看 ftrace 初始状态${NC}"
echo -e "${CYAN}$ ./litetrace status${NC}"
sleep 0.5
./litetrace status
echo ""
echo -e "${YELLOW}说明：显示当前追踪器为 nop（未启动），追踪已禁用${NC}"
echo ""
sleep 3

# 2.3 search 搜索函数 - vfs
echo -e "${CYAN}【3/9】search - 搜索内核函数（VFS 相关）${NC}"
echo -e "${CYAN}$ ./litetrace search "^vfs_" | head -20${NC}"
sleep 0.5
./litetrace search "^vfs_" | head -20
echo ""
echo -e "${YELLOW}说明：使用正则表达式搜索可追踪的内核函数${NC}"
echo ""
sleep 2

# 2.4 search 搜索函数 - tcp
echo -e "${CYAN}【4/9】search - 搜索 TCP 相关函数${NC}"
echo -e "${CYAN}$ ./litetrace search "tcp_send" | head -15${NC}"
sleep 0.5
./litetrace search "tcp_send" | head -15
echo ""
sleep 2

# ==================== 第三部分：前台追踪模式 ====================
clear
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  第二部分：前台追踪模式 - run 命令${NC}"
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 1

# 3.1 run 基本用法
echo -e "${CYAN}【5/9】run - 前台追踪（5秒）${NC}"
echo -e "${CYAN}$ ./litetrace run --tracer function --filter vfs_read --duration 5s --output /tmp/run_trace.txt${NC}"
echo -e "${YELLOW}说明：前台执行，阻塞终端直到完成${NC}"
sleep 0.5
./litetrace run --tracer function --filter vfs_read --duration 5s --output /tmp/run_trace.txt
echo ""
sleep 2

# 3.2 查看结果
echo -e "${CYAN}查看追踪结果：${NC}"
echo -e "${CYAN}$ head -15 /tmp/run_trace.txt${NC}"
head -15 /tmp/run_trace.txt
echo ""
echo -e "${YELLOW}说明：包含文件头信息和实际的 vfs_read 调用记录${NC}"
echo ""
sleep 3

# 3.3 Debug 模式
echo -e "${CYAN}【6/9】--debug - 调试模式（显示 tracefs 操作详情）${NC}"
echo -e "${CYAN}$ ./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/debug_trace.txt 2>&1 | head -30${NC}"
echo -e "${YELLOW}说明：黄色输出显示详细的 tracefs 文件读写操作${NC}"
sleep 0.5
./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/debug_trace.txt 2>&1 | head -30
echo ""
sleep 3

# ==================== 第四部分：后台追踪模式 ====================
clear
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  第三部分：后台追踪模式 - background / terminate${NC}"
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 1

# 4.1 background 启动后台追踪
echo -e "${CYAN}【7/9】background - 启动后台追踪（不阻塞终端）${NC}"
echo -e "${CYAN}$ ./litetrace background --tracer function --filter vfs_write --duration 8s --output /tmp/bg_trace.txt${NC}"
echo -e "${YELLOW}说明：立即返回，追踪在后台运行${NC}"
./litetrace background --tracer function --filter vfs_write --duration 8s --output /tmp/bg_trace.txt
echo ""
sleep 2

# 4.2 查看后台状态
echo -e "${CYAN}查看后台进程状态：${NC}"
echo -e "${CYAN}$ ./litetrace status${NC}"
./litetrace status
echo ""
echo -e "${YELLOW}说明：可以看到 Background Process Status 显示为 RUNNING${NC}"
echo ""
sleep 3

# 4.3 等待完成
echo "等待后台追踪完成（6秒）..."
sleep 6

# 4.4 查看完成后状态
echo -e "${CYAN}追踪完成后状态：${NC}"
echo -e "${CYAN}$ ./litetrace status${NC}"
./litetrace status
echo ""
sleep 2

# 4.5 查看后台追踪结果
echo -e "${CYAN}查看后台追踪结果：${NC}"
echo -e "${CYAN}$ wc -l /tmp/bg_trace.txt && head -10 /tmp/bg_trace.txt${NC}"
wc -l /tmp/bg_trace.txt
head -10 /tmp/bg_trace.txt
echo ""
sleep 3

# 4.6 演示 terminate
echo -e "${CYAN}【8/9】terminate - 停止后台追踪${NC}"
echo -e "${CYAN}$ ./litetrace background --tracer function --filter vfs_read${NC}"
./litetrace background --tracer function --filter vfs_read
echo ""
sleep 1

echo -e "${CYAN}启动另一个后台追踪（无限期运行）：${NC}"
echo -e "${CYAN}$ ./litetrace status${NC}"
./litetrace status
echo ""
sleep 2

echo -e "${CYAN}使用 terminate 停止：${NC}"
echo -e "${CYAN}$ ./litetrace terminate${NC}"
./litetrace terminate
echo ""
echo -e "${YELLOW}说明：terminate 会停止后台进程并清理 ftrace 状态${NC}"
echo ""
sleep 3

# ==================== 第五部分：TUI 实时界面 ====================
clear
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  第四部分：TUI 实时图形界面${NC}"
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 1

# 5.1 TUI 介绍
echo -e "${CYAN}【9/9】tui - 实时图形界面${NC}"
echo ""
echo -e "${YELLOW}TUI 界面功能：${NC}"
echo "  ┌─────────────────────────────────────────┐"
echo "  │  状态面板    │    追踪数据面板          │"
echo "  ├──────────────┼──────────────────────────┤"
echo "  │              │                          │"
echo "  │  实时显示    │    实时显示内核函数      │"
echo "  │  追踪状态    │    调用记录              │"
echo "  │              │                          │"
echo "  ├──────────────┴──────────────────────────┤"
echo "  │  日志面板 / 快捷键提示                  │"
echo "  └─────────────────────────────────────────┘"
echo ""
echo -e "${GREEN}快捷键：${NC}"
echo "  w - 打开配置向导，选择 tracer 和 filter"
echo "  p - 暂停/恢复追踪"
echo "  a - 切换自动滚动"
echo "  s - 保存追踪数据到文件"
echo "  c - 清空追踪缓冲区"
echo "  q - 退出 TUI"
echo ""
echo -e "${YELLOW}按 Enter 启动 TUI（体验实时追踪）...${NC}"
read

# 启动 TUI（使用子 shell 防止退出影响主脚本）
echo -e "${CYAN}$ ./litetrace tui${NC}"
sleep 0.5
(./litetrace tui) || true
echo ""
echo -e "${YELLOW}TUI 已退出，继续演示...${NC}"
sleep 2

# ==================== 第六部分：Wizard 交互式向导 ====================
clear
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  第五部分：Wizard 交互式向导${NC}"
echo -e "${MAGENTA}══════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 1

echo -e "${CYAN}wizard - 交互式向导${NC}"
echo ""
echo -e "${YELLOW}向导流程：${NC}"
echo "  Step 1: 选择追踪器（function / function_graph / nop）"
echo "  Step 2: 输入函数过滤器（支持通配符）"
echo "  Step 3: 选择查看模式"
echo "          [1] TUI Dashboard - 实时图形界面"
echo "          [2] Run silently  - 静默运行并导出"
echo "          [3] Background    - 后台运行（推荐！）"
echo "  Step 4: 设置追踪时长（可选）"
echo "  Step 5: 确认配置并开始追踪"
echo ""
echo -e "${GREEN}特点：${NC}"
echo "  • 自动验证函数名有效性"
echo "  • 引导式操作，适合新手"
echo "  • 支持三种查看模式"
echo "  • 后台模式不阻塞终端"
echo ""
echo -e "${YELLOW}按 Enter 启动 Wizard...${NC}"
read

echo -e "${CYAN}$ ./litetrace wizard${NC}"
sleep 0.5
(./litetrace wizard) || true
echo ""
echo -e "${YELLOW}Wizard 已退出...${NC}"
sleep 1

# ==================== 结束部分 ====================
clear
echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                                ║${NC}"
echo -e "${GREEN}║${NC}              🎉  Litetrace 功能演示完成！ 🎉                    ${GREEN}║${NC}"
echo -e "${GREEN}║                                                                ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}已展示的所有功能：${NC}"
echo ""
echo -e "${GREEN}基础命令：${NC}"
echo "  ✓ tldr    - 快速帮助"
echo "  ✓ status  - 查看 ftrace 状态"
echo "  ✓ search  - 搜索内核函数（支持正则）"
echo ""
echo -e "${GREEN}追踪模式：${NC}"
echo "  ✓ run       - 前台追踪（阻塞终端）"
echo "  ✓ background- 后台追踪（不阻塞）"
echo "  ✓ terminate - 停止后台追踪"
echo ""
echo -e "${GREEN}高级功能：${NC}"
echo "  ✓ --debug   - 调试模式（显示 tracefs 操作）"
echo "  ✓ tui       - 实时图形界面"
echo "  ✓ wizard    - 交互式向导"
echo ""
echo -e "${GREEN}追踪器类型：${NC}"
echo "  • function       - 追踪函数入口（轻量）"
echo "  • function_graph - 追踪调用链（详细）"
echo "  • nop            - 禁用追踪"
echo ""
echo -e "${YELLOW}更多信息请查看：${NC}"
echo "  • README.md          - 项目说明"
echo "  • docs/cheatsheet.txt - 使用手册"
echo "  • docs/DEVELOPER.md   - 开发者文档"
echo ""
sleep 2
