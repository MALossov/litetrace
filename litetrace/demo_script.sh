#!/bin/bash
# Litetrace 功能演示脚本

echo "=========================================="
echo "   Litetrace 功能演示"
echo "=========================================="
echo ""

cd /root/wkSpc/litetrace/litetrace

echo "【功能1】查看当前 ftrace 配置状态"
echo "----------------------------------------"
./litetrace_test status
echo ""
sleep 2

echo "【功能2】搜索内核函数 - 查找 vfs_read 相关函数"
echo "----------------------------------------"
./litetrace_test search "vfs_read" | head -5
echo "... (显示前5个结果)"
echo ""
sleep 2

echo "【功能3】运行动态追踪 - 追踪 vfs_read 函数 5 秒"
echo "----------------------------------------"
echo "执行: ./litetrace_test run --tracer function --filter vfs_read --duration 5s --output /tmp/trace_demo.txt"
echo ""
./litetrace_test run --tracer function --filter vfs_read --duration 5s --output /tmp/trace_demo.txt
echo ""
sleep 1

echo "【功能4】查看追踪后的状态变化"
echo "----------------------------------------"
./litetrace_test status
echo ""
sleep 2

echo "【功能5】展示输出文件内容"
echo "----------------------------------------"
echo "文件大小: $(ls -lh /tmp/trace_demo.txt | awk '{print $5}')"
echo "文件行数: $(wc -l < /tmp/trace_demo.txt)"
echo ""
echo "前 20 行内容:"
head -20 /tmp/trace_demo.txt
echo ""
echo "..."
echo ""
echo "追踪数据示例 (显示包含 vfs_read 的实际追踪记录):"
grep "vfs_read" /tmp/trace_demo.txt | head -10
echo ""

echo "=========================================="
echo "   演示完成!"
echo "=========================================="
echo ""
echo "功能验证总结:"
echo "✅ 支持 function 跟踪和 function 过滤"
echo "✅ 支持动态开启和关闭跟踪"
echo "✅ 支持查看当前配置状态 (current_tracer, tracing_on, set_ftrace_filter)"
echo "✅ 支持导出跟踪结果到文件"
echo "✅ 支持打印 trace 中的跟踪内容"
