# Litetrace 交互式录制指南

## 录制目标
展示完整的 ftrace 追踪流程，包括状态变化阶段和 bonus 功能。

## 录制步骤

### 准备阶段
```bash
cd /root/wkSpc/litetrace/litetrace
asciinema rec /tmp/litetrace_full_demo.cast -y
```

---

## 第一部分：初始状态检查

### 1.1 显示欢迎信息
echo "========================================"
echo "   Litetrace 完整功能演示"
echo "========================================"
echo ""

### 1.2 检查初始状态（追踪前）
echo "【阶段1】追踪前的初始状态"
echo "----------------------------------------"
sudo ./litetrace_test status
echo ""

**预期输出：**
- Engine Status: 🔴 STOPPED (tracing_on = 0)
- Current Tracer: nop
- Active Filters: #### all functions enabled ####

---

## 第二部分：启动后台追踪（30秒）

### 2.1 启动后台追踪任务
echo "【阶段2】启动30秒后台追踪任务"
echo "----------------------------------------"
echo "执行: sudo ./litetrace_test run --tracer function --filter vfs_read --duration 30s --output /tmp/trace_30s.txt &"
sudo ./litetrace_test run --tracer function --filter vfs_read --duration 30s --output /tmp/trace_30s.txt &
TRACING_PID=$!
echo "后台追踪PID: $TRACING_PID"
echo ""
sleep 2

### 2.2 检查追踪中的状态（关键！）
echo "【阶段3】追踪进行中的状态"
echo "----------------------------------------"
echo "注意观察 current_tracer、tracing_on、set_ftrace_filter 的变化！"
sudo ./litetrace_test status
echo ""

**预期输出：**
- Engine Status: 🟢 RUNNING (tracing_on = 1)
- Current Tracer: function
- Active Filters: vfs_read

### 2.3 等待追踪完成
echo "等待追踪完成（约25秒）..."
sleep 25
echo ""

---

## 第三部分：追踪完成后状态

### 3.1 检查追踪结束后的状态
echo "【阶段4】追踪完成后的状态"
echo "----------------------------------------"
sudo ./litetrace_test status
echo ""

**预期输出：**
- Engine Status: 🔴 STOPPED (tracing_on = 0)
- Current Tracer: nop
- Active Filters: #### all functions enabled ####

### 3.2 展示输出文件
echo "【阶段5】输出文件详情"
echo "----------------------------------------"
echo "文件大小: $(ls -lh /tmp/trace_30s.txt | awk '{print $5}')"
echo "文件行数: $(wc -l < /tmp/trace_30s.txt)"
echo ""
echo "前15行内容（包含文件头和追踪数据）:"
head -15 /tmp/trace_30s.txt
echo ""
echo "实际追踪记录示例:"
grep "vfs_read" /tmp/trace_30s.txt | head -5
echo ""

---

## 第四部分：Bonus 功能展示

### 4.1 Bonus 1: Wizard 模式
echo "========================================"
echo "   Bonus 功能展示"
echo "========================================"
echo ""
echo "【Bonus 1】Wizard 交互式向导模式"
echo "----------------------------------------"
echo "执行: sudo ./litetrace_test wizard"
echo "说明: 启动交互式向导，引导用户完成追踪配置"
echo ""

**这里实际执行：**
sudo ./litetrace_test wizard

**操作步骤：**
1. 选择 tracer: function
2. 输入 filter: vfs_write
3. 确认: Y
4. 选择查看模式: 2 (Run silently and Export)
5. 等待完成

### 4.2 Bonus 2: Debug 模式
echo ""
echo "【Bonus 2】Debug 调试模式"
echo "----------------------------------------"
echo "执行: sudo ./litetrace_test --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/trace_debug.txt"
echo "说明: 显示详细的 tracefs 文件操作过程（黄色输出）"
echo ""
sudo ./litetrace_test --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/trace_debug.txt
echo ""

### 4.3 Bonus 3: TUI 动态监控
echo "【Bonus 3】TUI 实时动态监控"
echo "----------------------------------------"
echo "执行: sudo ./litetrace_test tui"
echo "说明: 启动终端图形界面，实时查看追踪数据流"
echo "快捷键: q=退出, p=暂停/恢复, s=保存, c=清空"
echo ""
echo "注意: TUI 会全屏显示，按 'q' 退出"
sleep 2
sudo ./litetrace_test tui

---

## 第五部分：总结

### 5.1 功能总结
echo ""
echo "========================================"
echo "   演示完成 - 功能总结"
echo "========================================"
echo ""
echo "✅ 核心功能:"
echo "   1. 支持 function 跟踪和 function 过滤"
echo "   2. 支持动态开启和关闭跟踪 (tracing_on 0→1→0)"
echo "   3. 支持查看 current_tracer 状态变化 (nop→function→nop)"
echo "   4. 支持查看 set_ftrace_filter 状态变化"
echo "   5. 支持导出跟踪结果到文件"
echo "   6. 支持打印 trace 中的跟踪内容"
echo ""
echo "✅ Bonus 功能:"
echo "   1. Wizard 交互式向导模式"
echo "   2. --debug 调试模式（黄色输出 tracefs 操作）"
echo "   3. TUI 实时动态监控界面"
echo ""

---

## 结束录制
按 Ctrl+D 或输入 exit 结束录制

## 播放录制
```bash
asciinema play /tmp/litetrace_full_demo.cast
```

## 上传到 asciinema.org 分享
```bash
asciinema upload /tmp/litetrace_full_demo.cast
```
