#!/bin/bash
# Litetrace asciinema 录制脚本
# 用法: asciinema rec /tmp/litetrace_demo.cast -c "sudo ./demo_script.sh"

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_DIR="/root/wkSpc/litetrace/litetrace"
cd "$PROJECT_DIR"

# 检查编译
if [ ! -f "./litetrace" ]; then
    echo -e "${YELLOW}编译 litetrace...${NC}"
    go build -o litetrace .
fi

clear
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Litetrace 功能演示${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
sleep 1

# 1. 查看状态
echo -e "${CYAN}$ ./litetrace status${NC}"
sleep 0.5
./litetrace status
echo ""
sleep 2

# 2. Search 搜索函数
echo -e "${CYAN}$ ./litetrace search "^vfs_"${NC}"
sleep 0.5
./litetrace search "^vfs_" | head -15
echo ""
sleep 2

# 3. 启动追踪（后台）
echo -e "${CYAN}$ ./litetrace run --tracer function --filter vfs_read --duration 10s --output /tmp/trace.txt &${NC}"
./litetrace run --tracer function --filter vfs_read --duration 10s --output /tmp/trace.txt &
TRACING_PID=$!
echo "后台追踪PID: $TRACING_PID"
sleep 2

# 4. 查看运行中状态
echo -e "${CYAN}$ ./litetrace status${NC}"
sleep 0.5
./litetrace status
echo ""
sleep 2

# 5. 等待完成
echo "等待追踪完成..."
wait $TRACING_PID
sleep 1

# 6. 查看结果
echo -e "${CYAN}$ head -10 /tmp/trace.txt${NC}"
head -10 /tmp/trace.txt
echo ""
sleep 2

# 7. Debug 模式
echo -e "${CYAN}$ ./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/debug.txt${NC}"
sleep 0.5
./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/debug.txt
echo ""
sleep 2

# 8. TUI 演示提示
clear
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}   TUI 实时演示${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "即将启动 TUI 界面，请手动演示以下操作："
echo ""
echo -e "  ${GREEN}w${NC} - 打开配置向导，选择 tracer 和 filter"
echo -e "  ${GREEN}p${NC} - 暂停/恢复追踪"
echo -e "  ${GREEN}a${NC} - 切换自动滚动"
echo -e "  ${GREEN}s${NC} - 保存追踪数据"
echo -e "  ${GREEN}c${NC} - 清空缓冲区"
echo -e "  ${GREEN}q${NC} - 退出 TUI"
echo ""
echo -e "${YELLOW}按 Enter 启动 TUI...${NC}"
read

# 启动 TUI
echo -e "${CYAN}$ ./litetrace tui${NC}"
sleep 0.5
./litetrace tui

# TUI 退出后
clear
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   演示完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "已展示功能："
echo "  ✓ status - 查看 ftrace 状态"
echo "  ✓ search - 搜索内核函数"
echo "  ✓ run - 执行追踪任务"
echo "  ✓ --debug - 调试模式"
echo "  ✓ tui - 实时图形界面"
echo ""
sleep 2
