package ui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/M410550/lite-tracer-mygo/internal/ftrace"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Dashboard struct {
	app        *tview.Application
	engine     *ftrace.Engine
	statusView *tview.TextView
	traceView  *tview.TextView
	logView    *tview.TextView // 指令输出日志区域
	mainFlex   *tview.Flex
	ctx        context.Context
	cancel     context.CancelFunc

	// 并发安全缓冲
	mu           sync.Mutex
	traceBuffer  []string // 用于保存和历史查看的全量缓冲区
	renderBuffer []string // 用于平滑 UI 渲染的零时缓冲区

	paused         bool
	tracingStarted bool
	autoScroll     bool      // 自动滚动到底部开关
	savedFile      string    // 最后保存的文件路径
	shutdownOnce   sync.Once // 确保 Shutdown 只执行一次
}

func NewDashboard(engine *ftrace.Engine) *Dashboard {
	ctx, cancel := context.WithCancel(context.Background())
	return &Dashboard{
		engine:         engine,
		ctx:            ctx,
		cancel:         cancel,
		traceBuffer:    make([]string, 0, 2000),
		renderBuffer:   make([]string, 0, 100),
		tracingStarted: false,
		paused:         false,
		autoScroll:     true, // 默认开启自动滚动
	}
}

// Run 启动 TUI，返回保存的文件路径（如果有）
func (d *Dashboard) Run() (string, error) {
	return d.runInternal(false)
}

// RunWithTracingStarted 用于外部已启动 tracing 的情况（如 wizard），返回保存的文件路径（如果有）
func (d *Dashboard) RunWithTracingStarted() (string, error) {
	return d.runInternal(true)
}

func (d *Dashboard) runInternal(tracingAlreadyStarted bool) (string, error) {
	d.app = tview.NewApplication()

	// --- UI 组件初始化 ---
	d.statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(false)
	d.statusView.SetBorder(true).SetTitle(" System Status (Kernel) ")

	d.traceView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true)
	d.traceView.SetBorder(true).SetTitle(" Live Trace Output ")

	// 指令输出日志区域
	d.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	d.logView.SetBorder(true).SetTitle(" Command Output ")

	helpView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	helpView.SetBorder(true).SetTitle(" Controls ")
	helpView.SetText(` [green]w[white]: Wizard | [green]p[white]: Pause/Resume | [green]c[white]: Clear | [green]s[white]: Save | [green]q[white]: Quit `)

	// --- 布局构建 ---
	// 状态栏固定宽度 48，追踪视图弹性占据剩余空间
	topFlex := tview.NewFlex().
		AddItem(d.statusView, 48, 1, false).
		AddItem(d.traceView, 0, 1, true)

	// 底部区域：日志视图 + 帮助视图
	bottomFlex := tview.NewFlex().
		AddItem(d.logView, 0, 2, false).
		AddItem(helpView, 48, 1, false)

	d.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 0, 3, true).
		AddItem(bottomFlex, 6, 0, false)

	// --- 关键修复：输入劫持处理 ---

	// 1. 主界面业务快捷键：绑定在 mainFlex 上
	// 这样当 Root 切换到 Wizard 弹窗时，这些监听会自动“离线”，不再拦截输入框字符
	d.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			d.Shutdown()
			return nil
		case 'w', 'W':
			d.showStartWizard()
			return nil
		case 'p', 'P':
			d.togglePause()
			return nil
		// case 'a', 'A':
		// 	d.toggleAutoScroll()
		// 	return nil
		case 'c', 'C':
			d.clearBuffer()
			return nil
		case 's', 'S':
			d.saveTrace()
			return nil
		}
		return event
	})

	// 2. 全局系统快捷键：仅处理强退等强制操作
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			d.Shutdown()
			return nil
		}
		// 其他按键（如下层输入框的 s、w 等）原样传递
		return event
	})

	// 设置根节点并聚焦
	d.app.SetRoot(d.mainFlex, true).SetFocus(d.traceView)

	// --- 启动后台处理协程 ---
	go func() {
		// 稍微延迟确保 Run() 已接管终端
		time.Sleep(100 * time.Millisecond)

		if tracingAlreadyStarted {
			// 外部已启动 tracing，直接开始读取 trace_pipe
			d.tracingStarted = true
			d.logPrint("green", "Tracing already started, launching dashboard...")
			go d.startTraceStreamLoop()
		} else {
			d.app.QueueUpdateDraw(func() {
				fmt.Fprintln(d.traceView, "[yellow]>> Litetrace Ready. Press 'w' to configure tracing...[white]")
			})
		}

		go d.updateStatusLoop()
		go d.handleSignals()
	}()

	runErr := d.app.Run()
	// 无论正常退出还是出错，都调用 Shutdown 并返回 savedFile
	d.Shutdown()
	if runErr != nil {
		return d.savedFile, fmt.Errorf("tview run error: %v", runErr)
	}
	return d.savedFile, nil
}

// showStartWizard 修复了输入字符被劫持的问题
func (d *Dashboard) showStartWizard() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Configuration Wizard ")

	tracerOptions := []string{"function", "function_graph", "nop"}
	selectedTracer := "function"
	form.AddDropDown("Current Tracer", tracerOptions, 0, func(option string, _ int) {
		selectedTracer = option
	})

	filterText := ""
	form.AddInputField("Function Filter", "", 30, nil, func(text string) {
		filterText = text
	})

	form.AddButton("Apply & Start", func() {
		// 先恢复主布局，避免弹窗阻塞
		d.app.SetRoot(d.mainFlex, true).SetFocus(d.traceView)
		d.traceView.Clear()
		d.logPrint("yellow", fmt.Sprintf("Starting tracing with [%s]...", selectedTracer))

		// 异步启动 tracing，避免阻塞 TUI
		go func() {
			if err := d.engine.StartTracing(selectedTracer, filterText); err != nil {
				d.showErrorInTraceView(err)
				return
			}
			d.app.QueueUpdateDraw(func() {
				d.tracingStarted = true
				d.paused = false
				d.logPrint("green", fmt.Sprintf("Tracing started with [%s]", selectedTracer))
			})
			// tracing 启动成功后再启动 trace_pipe 读取
			go d.startTraceStreamLoop()
		}()
	})

	form.AddButton("Cancel", func() {
		d.app.SetRoot(d.mainFlex, true).SetFocus(d.traceView)
	})

	// 居中弹窗逻辑
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 12, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	d.app.SetRoot(modal, true).SetFocus(form)
}

func (d *Dashboard) showErrorInTraceView(err error) {
	d.app.QueueUpdateDraw(func() {
		fmt.Fprintf(d.logView, "[red]!! Error: %v[white]\n", err)
		d.logView.ScrollToEnd()
	})
}

// logPrint 在日志区域打印消息（带时间戳）
func (d *Dashboard) logPrint(color, msg string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(d.logView, "[grey]%s[%s] %s[white]\n", timestamp, color, msg)
	d.logView.ScrollToEnd()
}

func (d *Dashboard) updateStatusLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			status, err := d.engine.GetStatus()
			if err != nil {
				continue
			}

			onColor := "[red]OFF"
			if status.Enabled {
				onColor = "[green]ON"
			}

			statusText := fmt.Sprintf(
				"[yellow]STATUS:[white]    %s\n"+
					"[yellow]TRACER:[white]    %s\n"+
					"[yellow]BUFFER:[white]    %d KB\n"+
					"[yellow]CLOCK:[white]     %s\n"+
					"[yellow]FILTER:[white]    %s",
				onColor, status.Tracer, status.BufferSize, status.TraceClock, status.Filter,
			)

			d.app.QueueUpdateDraw(func() {
				d.statusView.SetText(statusText)
			})
		}
	}
}

func (d *Dashboard) startTraceStreamLoop() {
	pipeChan := d.engine.ReadTracePipe(d.ctx)

	// 渲染刷屏保护：100ms 批次处理
	renderTicker := time.NewTicker(100 * time.Millisecond)
	defer renderTicker.Stop()

	maxHistory := 5000 // 最大历史行数

	for {
		select {
		case <-d.ctx.Done():
			return
		case line, ok := <-pipeChan:
			if !ok {
				return
			}
			d.mu.Lock()
			d.traceBuffer = append(d.traceBuffer, line)
			d.renderBuffer = append(d.renderBuffer, line)

			// 内存管理：防止无限增长
			if len(d.traceBuffer) > maxHistory {
				d.traceBuffer = d.traceBuffer[len(d.traceBuffer)-maxHistory:]
			}
			d.mu.Unlock()

		case <-renderTicker.C:
			d.mu.Lock()
			if len(d.renderBuffer) == 0 {
				d.mu.Unlock()
				continue
			}

			batch := strings.Join(d.renderBuffer, "")
			d.renderBuffer = d.renderBuffer[:0]
			d.mu.Unlock()

			// 更新 UI - 根据 autoScroll 设置决定是否自动滚动到底部
			d.app.QueueUpdateDraw(func() {
				fmt.Fprint(d.traceView, batch)
				if d.autoScroll {
					d.traceView.ScrollToEnd()
				}
			})
		}
	}
}

func (d *Dashboard) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	d.Shutdown()
}

func (d *Dashboard) Shutdown() {
	d.shutdownOnce.Do(func() {
		d.cancel()
		d.app.Stop()
		d.engine.SafeShutdown()
	})
}

func (d *Dashboard) togglePause() {
	if !d.tracingStarted {
		d.logPrint("yellow", "Tracing not started yet")
		return
	}
	d.paused = !d.paused
	if d.paused {
		d.engine.Disable()
		d.logPrint("yellow", "Tracing PAUSED")
	} else {
		d.engine.Enable()
		d.logPrint("green", "Tracing RESUMED")
	}
}

// func (d *Dashboard) toggleAutoScroll() {
// 	d.autoScroll = !d.autoScroll
// 	if d.autoScroll {
// 		d.logPrint("green", "AutoScroll ON")
// 		d.traceView.ScrollToEnd()
// 	} else {
// 		d.logPrint("yellow", "AutoScroll OFF")
// 	}
// }

func (d *Dashboard) clearBuffer() {
	d.mu.Lock()
	d.traceBuffer = d.traceBuffer[:0]
	d.renderBuffer = d.renderBuffer[:0]
	d.mu.Unlock()
	d.traceView.Clear()
	d.logPrint("grey", "Buffer Cleared")
}

func (d *Dashboard) saveTrace() {
	d.mu.Lock()
	if len(d.traceBuffer) == 0 {
		d.mu.Unlock()
		d.logPrint("yellow", "No trace data to save")
		return
	}
	content := strings.Join(d.traceBuffer, "")
	d.mu.Unlock()

	filename := fmt.Sprintf("/tmp/litetrace_tui_dump_%s.txt", time.Now().Format("150405"))
	err := os.WriteFile(filename, []byte(content), 0644)

	if err != nil {
		d.logPrint("red", fmt.Sprintf("Save Failed: %v", err))
	} else {
		d.savedFile = filename
		d.logPrint("green", fmt.Sprintf("Saved to %s", filename))
		// 计算文件大小和行数
		fileSize, err := os.Stat(filename)
		if err != nil {
			d.logPrint("red", fmt.Sprintf("failed to get file size: %v", err))
		}
		d.logPrint("green", fmt.Sprintf("Trace file size: %d bytes", fileSize.Size()))
		content, err := os.ReadFile(filename)
		if err != nil {
			d.logPrint("red", fmt.Sprintf("failed to read output file for line count: %v", err))
		}
		d.logPrint("green", fmt.Sprintf("Trace file line count: %d", len(strings.Split(string(content), "\n"))))
	}
}
