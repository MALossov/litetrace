// Package ftrace provides an interface to the Linux ftrace subsystem via tracefs.
// It offers a high-level API for configuring and controlling kernel function tracing,
// including setting tracers, filters, and managing trace output.
//
// The Engine struct is the core component that provides thread-safe access to tracefs.
// All operations are performed by reading from and writing to files in the tracefs
// filesystem, typically mounted at /sys/kernel/tracing or /sys/kernel/debug/tracing.
package ftrace

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Engine 提供对 tracefs 的操作接口
// 封装了所有与 ftrace 子系统交互的功能
type Engine struct {
	tracefsPath string // tracefs 挂载路径
	debug       bool   // 是否启用调试模式
}

// NewEngine 创建一个新的 Engine 实例
// tracefsPath: tracefs 的挂载路径，如 /sys/kernel/tracing
func NewEngine(tracefsPath string) *Engine {
	return &Engine{
		tracefsPath: tracefsPath,
		debug:       false,
	}
}

// NewEngineWithDebug 创建带调试模式的 Engine 实例
// tracefsPath: tracefs 的挂载路径
// debug: 是否启用调试输出
func NewEngineWithDebug(tracefsPath string, debug bool) *Engine {
	return &Engine{
		tracefsPath: tracefsPath,
		debug:       debug,
	}
}

func (e *Engine) SetDebug(debug bool) {
	e.debug = debug
}

func (e *Engine) debugLog(format string, args ...interface{}) {
	if e.debug {
		fmt.Printf("\033[33m[DEBUG] "+format+"\033[0m\n", args...)
	}
}

func (e *Engine) TracefsPath() string {
	return e.tracefsPath
}

func (e *Engine) WriteTraceFile(filename string, content string, appendMode bool) error {
	filePath := filepath.Join(e.tracefsPath, filename)

	flag := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	if appendMode {
		flag = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	}

	file, err := os.OpenFile(filePath, flag, 0644)
	if err != nil {
		e.debugLog("WRITE FAILED: %s -> '%s' (error: %v)", filename, content, err)
		return fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		e.debugLog("WRITE FAILED: %s -> '%s' (error: %v)", filename, content, err)
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	e.debugLog("WRITE: %s -> '%s'", filename, content)
	return nil
}

func (e *Engine) ReadTraceFile(filename string) (string, error) {
	filePath := filepath.Join(e.tracefsPath, filename)

	file, err := os.Open(filePath)
	if err != nil {
		e.debugLog("READ FAILED: %s (error: %v)", filename, err)
		return "", fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		e.debugLog("READ FAILED: %s (error: %v)", filename, err)
		return "", fmt.Errorf("failed to read %s: %w", filename, err)
	}

	result := strings.TrimSpace(string(content))
	// Limit debug output length for large files like trace
	debugContent := result
	if len(debugContent) > 100 {
		debugContent = debugContent[:100] + "..."
	}
	e.debugLog("READ: %s <- '%s'", filename, debugContent)
	return result, nil
}

func (e *Engine) SetTracer(tracer string) error {
	if tracer == "" {
		tracer = "nop"
	}
	if tracer != "nop" && tracer != "function" && tracer != "function_graph" {
		return fmt.Errorf("invalid tracer: %s (must be nop, function, or function_graph)", tracer)
	}
	return e.WriteTraceFile("current_tracer", tracer, false)
}

func (e *Engine) GetTracer() (string, error) {
	return e.ReadTraceFile("current_tracer")
}

func (e *Engine) SetFilter(filter string) error {
	if filter == "" {
		return e.WriteTraceFile("set_ftrace_filter", "", false)
	}
	filterContent := strings.ReplaceAll(filter, ",", "\n")
	return e.WriteTraceFile("set_ftrace_filter", filterContent, false)
}

func (e *Engine) GetFilter() (string, error) {
	return e.ReadTraceFile("set_ftrace_filter")
}

func (e *Engine) Enable() error {
	return e.WriteTraceFile("tracing_on", "1", false)
}

func (e *Engine) Disable() error {
	return e.WriteTraceFile("tracing_on", "0", false)
}

func (e *Engine) IsEnabled() (bool, error) {
	content, err := e.ReadTraceFile("tracing_on")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(content) == "1", nil
}

func (e *Engine) Clear() error {
	return e.WriteTraceFile("trace", "", false)
}

func (e *Engine) GetAvailableTracers() ([]string, error) {
	content, err := e.ReadTraceFile("available_tracers")
	if err != nil {
		return nil, err
	}
	return strings.Fields(content), nil
}

func (e *Engine) GetBufferSize() (int, error) {
	content, err := e.ReadTraceFile("buffer_size_kb")
	if err != nil {
		return 0, err
	}
	var size int
	fmt.Sscanf(content, "%d", &size)
	return size, nil
}

func (e *Engine) GetTraceClock() (string, error) {
	return e.ReadTraceFile("trace_clock")
}

type Status struct {
	Tracer     string
	Filter     string
	Enabled    bool
	BufferSize int
	TraceClock string
}

func (e *Engine) GetStatus() (*Status, error) {
	tracer, err := e.GetTracer()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracer: %w", err)
	}

	filter, _ := e.GetFilter()
	enabled, _ := e.IsEnabled()
	bufferSize, _ := e.GetBufferSize()
	traceClock, _ := e.GetTraceClock()

	return &Status{
		Tracer:     tracer,
		Filter:     filter,
		Enabled:    enabled,
		BufferSize: bufferSize,
		TraceClock: traceClock,
	}, nil
}

// StopAndExport 停止追踪并导出结果，执行 4 步关闭流程
// outputPath: 导出文件路径
// 流程: 禁用追踪 -> 读取并导出 -> 重置追踪器 -> 清空过滤器
// 注意: 必须先读取 trace 再重置追踪器，否则 trace 内容会被清空！
func (e *Engine) StopAndExport(outputPath string) error {
	e.debugLog("STOPPING: Beginning trace shutdown sequence...")

	// Step 1: Disable tracing
	e.debugLog("STOPPING: Step 1/4 - Disabling tracing (tracing_on = 0)")
	if err := e.Disable(); err != nil {
		e.debugLog("STOPPING FAILED: Could not disable tracing: %v", err)
		return fmt.Errorf("failed to disable tracing: %w", err)
	}
	e.debugLog("STOPPING: Tracing disabled successfully")

	// Step 2: Read and export trace (必须在重置追踪器之前读取！)
	e.debugLog("STOPPING: Step 2/4 - Reading trace buffer")
	traceContent, err := e.ReadTraceFile("trace")
	if err != nil {
		e.debugLog("STOPPING FAILED: Could not read trace: %v", err)
		return fmt.Errorf("failed to read trace: %w", err)
	}
	e.debugLog("STOPPING: Trace buffer read successfully (%d bytes)", len(traceContent))

	// Check for empty or minimal trace content (only header, no actual trace entries)
	lines := strings.Split(traceContent, "\n")
	hasData := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment lines
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			hasData = true
			break
		}
	}
	if !hasData {
		fmt.Println("[!] WARNING: Trace buffer is empty! No data was captured.")
		fmt.Println("    Possible reasons:")
		fmt.Println("    - Tracer was set to 'nop'")
		fmt.Println("    - Filter did not match any functions")
		fmt.Println("    - Tracing duration was too short")
		fmt.Println("    - No kernel activity matched the filter")
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		e.debugLog("STOPPING FAILED: Could not create output file: %v", err)
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.WriteString(traceContent); err != nil {
		e.debugLog("STOPPING FAILED: Could not write to output file: %v", err)
		return fmt.Errorf("failed to write trace to file: %w", err)
	}
	e.debugLog("STOPPING: Trace exported to %s successfully", outputPath)
	// 计算文件大小和行数
	fileSize, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fmt.Printf("[+] Trace file size: %d bytes\n", fileSize.Size())
	content, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read output file for line count: %w", err)
	}
	fmt.Printf("[+] Trace file line count: %d\n", len(strings.Split(string(content), "\n")))

	// Step 3: Reset tracer to nop (必须在读取 trace 之后！)
	e.debugLog("STOPPING: Step 3/4 - Resetting tracer to nop")
	if err := e.SetTracer("nop"); err != nil {
		e.debugLog("STOPPING FAILED: Could not reset tracer: %v", err)
		return fmt.Errorf("failed to reset tracer: %w", err)
	}
	e.debugLog("STOPPING: Tracer reset to nop successfully")

	// Step 4: Clear filter
	e.debugLog("STOPPING: Step 4/4 - Clearing filter")
	if err := e.SetFilter(""); err != nil {
		e.debugLog("STOPPING FAILED: Could not clear filter: %v", err)
		return fmt.Errorf("failed to clear filter: %w", err)
	}
	e.debugLog("STOPPING: Filter cleared successfully")
	e.debugLog("STOPPING: Trace shutdown sequence completed successfully")

	return nil
}

// StartTracing 启动追踪，执行完整的 7 步初始化流程
// tracer: 追踪器类型 (function/function_graph/nop)
// filter: 函数过滤器，支持通配符
// 流程: 禁用追踪 -> 重置追踪器 -> 设置占位过滤器 -> 清空缓冲区 -> 设置追踪器 -> 设置过滤器 -> 启用追踪
func (e *Engine) StartTracing(tracer, filter string) error {
	e.debugLog("STARTING: Beginning trace startup sequence...")

	// Step 1: Always disable tracing first
	fmt.Println("[*] Disabling current tracing...")
	e.debugLog("STARTING: Step 1/7 - Disabling tracing (tracing_on = 0)")
	if err := e.Disable(); err != nil {
		e.debugLog("STARTING FAILED: Could not disable tracing: %v", err)
		return fmt.Errorf("failed to disable tracing: %w", err)
	}
	e.debugLog("STARTING: Tracing disabled successfully")

	// Wait a moment to ensure tracing is fully stopped
	time.Sleep(100 * time.Millisecond)

	// Step 2: Reset tracer to nop before configuration
	fmt.Println("[*] Resetting tracer to nop...")
	e.debugLog("STARTING: Step 2/7 - Resetting tracer to nop")
	if err := e.SetTracer("nop"); err != nil {
		e.debugLog("STARTING FAILED: Could not reset tracer: %v", err)
		return fmt.Errorf("failed to reset tracer: %w", err)
	}
	e.debugLog("STARTING: Tracer reset to nop successfully")

	// Step 3: Set a dummy filter first (to prevent "all functions enabled" blocking tracer change)
	fmt.Println("[*] Setting placeholder filter...")
	e.debugLog("STARTING: Step 3/7 - Setting placeholder filter")
	if err := e.SetFilter("vfs_read"); err != nil {
		// If vfs_read doesn't exist, try another common function
		e.debugLog("STARTING: vfs_read not available, trying do_nanosleep")
		_ = e.SetFilter("do_nanosleep")
	}
	e.debugLog("STARTING: Placeholder filter set successfully")

	// Step 4: Clear buffer
	fmt.Println("[*] Clearing trace buffer...")
	e.debugLog("STARTING: Step 4/7 - Clearing trace buffer")
	if err := e.Clear(); err != nil {
		e.debugLog("STARTING FAILED: Could not clear buffer: %v", err)
		return fmt.Errorf("failed to clear buffer: %w", err)
	}
	e.debugLog("STARTING: Trace buffer cleared successfully")

	// Step 5: Set tracer (before filter for nop check)
	if tracer == "" {
		tracer = "function"
	}
	fmt.Printf("[*] Setting tracer to '%s'...\n", tracer)
	e.debugLog("STARTING: Step 5/7 - Setting tracer to '%s'", tracer)
	if err := e.SetTracer(tracer); err != nil {
		e.debugLog("STARTING FAILED: Could not set tracer: %v", err)
		return fmt.Errorf("failed to set tracer: %w", err)
	}
	e.debugLog("STARTING: Tracer set to '%s' successfully", tracer)

	// Step 6: Set actual filter (warn if nop tracer)
	if filter != "" {
		fmt.Printf("[*] Setting filter to '%s'...\n", filter)
		e.debugLog("STARTING: Step 6/7 - Setting filter to '%s'", filter)
		if err := e.SetFilter(filter); err != nil {
			e.debugLog("STARTING FAILED: Could not set filter: %v", err)
			return fmt.Errorf("failed to set filter: %w", err)
		}
		e.debugLog("STARTING: Filter set to '%s' successfully", filter)
	} else {
		// Clear the placeholder filter to enable all functions
		e.debugLog("STARTING: Step 6/7 - Clearing placeholder filter (no filter specified)")
		if err := e.SetFilter(""); err != nil {
			e.debugLog("STARTING FAILED: Could not clear filter: %v", err)
			return fmt.Errorf("failed to clear filter: %w", err)
		}
		e.debugLog("STARTING: Placeholder filter cleared successfully")
		if tracer == "nop" {
			fmt.Println("[!] WARNING: Tracer is 'nop' with no filter - no data will be captured!")
		} else {
			fmt.Println("[!] WARNING: No filter specified - capturing all functions (high overhead)!")
		}
	}

	// Step 7: Enable tracing
	fmt.Println("[*] Enabling tracing...")
	e.debugLog("STARTING: Step 7/7 - Enabling tracing (tracing_on = 1)")
	if err := e.Enable(); err != nil {
		e.debugLog("STARTING FAILED: Could not enable tracing: %v", err)
		return fmt.Errorf("failed to enable tracing: %w", err)
	}
	e.debugLog("STARTING: Tracing enabled successfully")
	e.debugLog("STARTING: Trace startup sequence completed successfully")

	fmt.Println("[+] Tracing started successfully")
	return nil
}

// RunWithDuration 按指定时长执行追踪
// tracer: 追踪器类型
// filter: 函数过滤器
// outputPath: 输出文件路径
// duration: 追踪持续时间
func (e *Engine) RunWithDuration(tracer, filter, outputPath string, duration time.Duration) error {
	fmt.Printf("[*] Tracing started for %v...\n", duration)

	if err := e.StartTracing(tracer, filter); err != nil {
		return fmt.Errorf("failed to start tracing: %w", err)
	}

	time.Sleep(duration)

	fmt.Printf("[*] Tracing completed. Exporting to %s...\n", outputPath)

	if err := e.StopAndExport(outputPath); err != nil {
		return fmt.Errorf("failed to stop and export: %w", err)
	}

	fmt.Printf("[+] Trace saved to %s\n", outputPath)

	// 计算文件大小和行数
	fileSize, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fmt.Printf("[+] Trace file size: %d bytes\n", fileSize.Size())
	content, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read output file for line count: %w", err)
	}
	fmt.Printf("[+] Trace file line count: %d\n", len(strings.Split(string(content), "\n")))

	return nil
}

// SafeShutdown 紧急安全关闭，用于信号处理
// 执行 3 步清理：禁用追踪 -> 重置追踪器 -> 清空过滤器
// 忽略错误，确保尽最大努力恢复安全状态
func (e *Engine) SafeShutdown(forced ...bool) {
	e.debugLog("SAFE SHUTDOWN: Emergency shutdown initiated")

	e.debugLog("SAFE SHUTDOWN: Disabling tracing (tracing_on = 0)")
	if err := e.WriteTraceFile("tracing_on", "0", false); err != nil {
		e.debugLog("SAFE SHUTDOWN FAILED: Could not disable tracing: %v", err)
	} else {
		e.debugLog("SAFE SHUTDOWN: Tracing disabled")
	}

	e.debugLog("SAFE SHUTDOWN: Resetting tracer to nop")
	if err := e.WriteTraceFile("current_tracer", "nop", false); err != nil {
		e.debugLog("SAFE SHUTDOWN FAILED: Could not reset tracer: %v", err)
	} else {
		e.debugLog("SAFE SHUTDOWN: Tracer reset to nop")
	}

	e.debugLog("SAFE SHUTDOWN: Clearing filter")
	if err := e.WriteTraceFile("set_ftrace_filter", "", false); err != nil {
		e.debugLog("SAFE SHUTDOWN FAILED: Could not clear filter: %v", err)
	} else {
		e.debugLog("SAFE SHUTDOWN: Filter cleared")
	}

	// 检查 trace_pipe 是否存在，如果不存在则略过
	if len(forced) > 0 && !forced[0] {
		return
	}

	tracePipePath := filepath.Join(e.tracefsPath, "trace_pipe")
	if _, err := os.Stat(tracePipePath); err == nil {
		e.debugLog("SAFE SHUTDOWN: Checking for processes using %s", tracePipePath)
		cmd := exec.Command("lsof", tracePipePath)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines[1:] {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					pid := fields[1]
					e.debugLog("SAFE SHUTDOWN: Killing process %s using trace_pipe", pid)
					killCmd := exec.Command("kill", "-9", pid)
					_ = killCmd.Run()
				}
			}
		}
	}

	e.debugLog("SAFE SHUTDOWN: Emergency shutdown completed")
}
