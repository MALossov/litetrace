# Litetrace 交互式录制指南

## 录制目标

展示完整的 ftrace 追踪流程，包括状态变化阶段和 bonus 功能。

## 录制步骤

### 准备阶段

```bash
cd /root/wkSpc/litetrace/litetrace
asciinema rec /tmp/litetrace_full_demo.cast -y
```

***

## 第一部分：初始状态检查

### 1.1 显示欢迎信息

```bash
echo "========================================"
echo "   Litetrace 完整功能演示"
echo "========================================"
echo ""
```

### 1.2 检查初始状态（追踪前）

```bash
echo "【阶段1】追踪前的初始状态"
echo "----------------------------------------"
sudo ./litetrace status
echo ""
```

**预期输出：**

- Engine Status: 🔴 STOPPED (tracing\_on = 0)
- Current Tracer: nop
- Active Filters: #### all functions enabled ####

***

## 第二部分：启动后台追踪（30秒）

### 2.1 启动后台追踪任务

```bash
echo "【阶段2】启动30秒后台追踪任务"
echo "----------------------------------------"
echo "执行: sudo ./litetrace background --tracer function --filter vfs_read --duration 30s --output /tmp/trace_30s.txt"
sudo ./litetrace background --tracer function --filter vfs_read --duration 30s --output /tmp/trace_30s.txt
echo ""
sleep 2
```

### 2.2 检查追踪中的状态（关键！）

```bash
echo "【阶段3】追踪进行中的状态"
echo "----------------------------------------"
echo "注意观察 current_tracer、tracing_on、set_ftrace_filter 的变化！"
sudo ./litetrace status
echo ""
```

**预期输出：**

- Engine Status: 🟢 RUNNING (tracing\_on = 1)
- Current Tracer: function
- Active Filters: vfs\_read
- Background Process Status: 🟢 RUNNING
- Process ID: <PID>

### 2.3 等待追踪完成

```bash
echo "等待追踪完成（约25秒）..."
sleep 25
echo ""
```

**说明：** 使用 `background` 命令启动的后台追踪会在指定时间后自动停止并保存结果，无需手动等待。

***

## 第三部分：追踪完成后状态

### 3.1 检查追踪结束后的状态

```bash
echo "【阶段4】追踪完成后的状态"
echo "----------------------------------------"
sudo ./litetrace status
echo ""
```

**预期输出：**

- Engine Status: 🔴 STOPPED (tracing\_on = 0)
- Current Tracer: nop
- Active Filters: #### all functions enabled ####

### 3.2 展示输出文件

```bash
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
```

***

## 第四部分：Bonus 功能展示

```bash
echo "========================================"
echo "   Bonus 功能展示"
echo "========================================"
echo ""
```

### 4.1 Bonus 1: Search 函数搜索

```bash
echo "【Bonus 1】Search 函数搜索功能"
echo "----------------------------------------"
echo "执行: sudo ./litetrace search \"^vfs_\""
echo "说明: 使用正则表达式搜索可追踪的内核函数"
echo ""
sudo ./litetrace search "^vfs_"
```

**预期输出：**
- 列出所有以 vfs_ 开头的内核函数
- 显示匹配函数数量

```bash
echo ""
echo "执行: sudo ./litetrace search \"tcp_send\""
echo "说明: 搜索 TCP 发送相关函数"
echo ""
sudo ./litetrace search "tcp_send"
echo ""
```

### 4.2 Bonus 2: Wizard 交互式向导

```bash
echo "【Bonus 2】Wizard 交互式向导模式"
echo "----------------------------------------"
echo "执行: sudo ./litetrace wizard"
echo "说明: 启动交互式向导，引导用户完成追踪配置"
echo ""
echo "操作步骤:"
echo "  1. 选择 tracer: function"
echo "  2. 输入 filter: vfs_write"
echo "  3. 确认: Y"
echo "  4. 选择查看模式: 2 (Run silently and Export)"
echo "  5. 等待完成"
echo ""
sudo ./litetrace wizard
```

### 4.3 Bonus 3: TUI 实时动态监控

```bash
echo ""
echo "【Bonus 3】TUI 实时动态监控"
echo "----------------------------------------"
echo "执行: sudo ./litetrace tui"
echo "说明: 启动终端图形界面，实时查看追踪数据流"
echo "快捷键:"
echo "  w         打开配置向导"
echo "  p         暂停/恢复追踪"
echo "  a         切换自动滚动"
echo "  c         清空追踪缓冲区"
echo "  s         保存当前追踪到文件"
echo "  q/Ctrl+C  退出"
echo ""
echo "注意: TUI 会全屏显示，按 'q' 退出"
sleep 2
sudo ./litetrace tui
```

### 4.4 Bonus 4: Debug 调试模式

```bash
echo ""
echo "【Bonus 4】Debug 调试模式"
echo "----------------------------------------"
echo "执行: sudo ./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/trace_debug.txt"
echo "说明: 显示详细的 tracefs 文件操作过程（黄色输出）"
echo ""
sudo ./litetrace --debug run --tracer function --filter vfs_open --duration 3s --output /tmp/trace_debug.txt
echo ""
```

### 4.5 Bonus 5: 信号终止处理

```bash
echo "【Bonus 5】信号终止处理 (Ctrl+C)"
echo "----------------------------------------"
echo "执行: sudo ./litetrace run --tracer function --filter vfs_read --duration 60s --output /tmp/trace_terminate.txt"
echo "说明: 演示在追踪过程中按 Ctrl+C 终止时的清理行为"
echo "      系统会自动恢复 ftrace 状态，确保不留下垃圾配置"
echo ""
echo "开始追踪（请在3秒后按 Ctrl+C 终止）..."
sleep 3
sudo ./litetrace run --tracer function --filter vfs_read --duration 60s --output /tmp/trace_terminate.txt
```

**操作说明：**
1. 追踪启动后，状态变为 RUNNING
2. 按 **Ctrl+C** 发送终止信号
3. 观察程序捕获信号并执行清理
4. 最后检查状态确认已恢复

```bash
echo ""
echo "检查终止后的状态（确认已清理）:"
sudo ./litetrace status
echo ""
```

***

## 第五部分：总结

### 5.1 功能总结

```bash
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
echo "   1. Search 函数搜索（支持正则表达式）"
echo "   2. Wizard 交互式向导模式"
echo "   3. TUI 实时动态监控界面（支持暂停/自动滚动/保存）"
echo "   4. --debug 调试模式（黄色输出 tracefs 操作）"
echo "   5. 信号终止处理（Ctrl+C 自动清理）"
echo ""
```

***

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

