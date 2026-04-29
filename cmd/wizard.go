package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"M410550-LOCAL-DEV/lite-tracer-mygo/internal/ftrace"
	"M410550-LOCAL-DEV/lite-tracer-mygo/internal/search"
	"M410550-LOCAL-DEV/lite-tracer-mygo/internal/ui"
	"M410550-LOCAL-DEV/lite-tracer-mygo/internal/wizard"

	"github.com/spf13/cobra"
)

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive wizard mode for tracing",
	Run: func(cmd *cobra.Command, args []string) {
		if !checkRoot() {
			fmt.Fprintln(os.Stderr, "🚨 Fatal: Root privileges required")
			os.Exit(1)
		}

		tracefsPath, err := ftrace.FindTracefs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		engine := ftrace.NewEngineWithDebug(tracefsPath, debugFlag)

		wizard.PrintWelcome()

		// Step 1: Choose tracer
		tracer, err := wizard.AskTracer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// Step 2: Choose filter
		filter, err := wizard.AskFilter()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// Step 3: Validate filter
		if filter != "" {
			result := search.ValidateAndNormalizeFilter(tracefsPath, filter, string(tracer))
			if result.ShouldIgnore {
				wizard.PrintWarning(result.Message)
				filter = ""
			} else if !result.Valid {
				wizard.PrintError(result.Message)
				os.Exit(1)
			} else {
				wizard.PrintSuccess(result.Message)
			}
		} else {
			wizard.PrintWarning("No filter specified. System may experience high load.")
		}

		// Step 4: Choose view mode BEFORE starting tracing
		viewMode, err := wizard.AskViewMode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// Step 5: Ask for duration based on view mode
		var duration string
		if viewMode == 2 || viewMode == 3 {
			duration, err = wizard.AskDuration()
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			if duration != "" {
				wizard.PrintSuccess(fmt.Sprintf("Tracing duration set to: %s", duration))
			} else {
				wizard.PrintWarning("Tracing will run until manually stopped")
			}
		}

		// Step 6: Confirm configuration
		confirmMsg := fmt.Sprintf("Start tracing with [%s] tracer and filter [%s]? (Y/n)", tracer, filter)
		confirmed, err := wizard.AskConfirm(confirmMsg)
		if err != nil || !confirmed {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}

		// Step 7: Start tracing
		if err := engine.StartTracing(string(tracer), filter); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		wizard.PrintSuccess("Tracing started!")

		// Step 8: Execute based on view mode
		switch viewMode {
		case 1:
			// TUI Dashboard mode
			fmt.Println("[*] Launching TUI Dashboard...")
			time.Sleep(500 * time.Millisecond)
			dashboard := ui.NewDashboard(engine)
			// tracing 已在外部启动，使用 RunWithTracingStarted
			savedFile, err := dashboard.RunWithTracingStarted()

			// 无论成功与否，先输出保存信息
			if savedFile != "" {
				fmt.Fprintf(os.Stderr, "\n[✓] Trace data saved to: %s\n", savedFile)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}

		case 2:
			// Silent mode with export
			if duration != "" {
				d, _ := time.ParseDuration(duration)
				fmt.Printf("[*] Running silently for %v...\n", d)
				time.Sleep(d)
			} else {
				fmt.Println("[*] Running silently. Press Ctrl+C to stop...")
				// Wait for interrupt
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
				<-sigChan
			}

			timestamp := time.Now().Format("20060102_150405")
			outputPath := fmt.Sprintf("/tmp/litetrace_wizard_%s.txt", timestamp)

			fmt.Printf("[*] Exporting trace to %s...\n", outputPath)
			if err := engine.StopAndExport(outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			wizard.PrintSuccess(fmt.Sprintf("Trace saved to %s", outputPath))

		case 3:
			// Background mode - 真正的后台运行
			fmt.Println("[*] Starting background tracing process...")

			// 先停止当前引擎（后台进程会自己启动）
			engine.SafeShutdown()

			// 启动真正的后台进程
			timestamp := time.Now().Format("20060102_150405")
			outputPath := fmt.Sprintf("/tmp/litetrace_wizard_%s.txt", timestamp)

			if err := ftrace.StartBackgroundProcess(string(tracer), filter, duration, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}

			// 获取后台进程PID
			_, pid, _ := ftrace.IsDaemonRunning()

			wizard.PrintSuccess(fmt.Sprintf("Background tracing started with PID %d", pid))
			fmt.Println()
			fmt.Println("[*] Tracing is now running in background.")
			if duration != "" {
				fmt.Printf("[*] Will run for %s\n", duration)
			}
			fmt.Println()
			fmt.Println("To manage the background process:")
			fmt.Println("    litetrace status     - Check tracing status")
			fmt.Println("    litetrace terminate  - Stop background tracing")
			fmt.Println()
			fmt.Printf("[*] Results will be saved to: %s\n", outputPath)
			fmt.Println()
			fmt.Println("[+] You can now continue using your shell.")
		}
	},
}
