package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/M410550/lite-tracer-mygo/internal/ftrace"
	"github.com/spf13/cobra"
)

var (
	daemonTracer   string
	daemonFilter   string
	daemonDuration string
	daemonOutput   string
)

// backgroundDaemonCmd 是实际在后台运行tracing的命令（内部使用）
var backgroundDaemonCmd = &cobra.Command{
	Use:    "background-daemon",
	Short:  "Internal daemon for background tracing",
	Long:   "This command is used internally by the background command. Do not use directly.",
	Hidden: true,
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

		// 写入PID文件
		if err := ftrace.WritePidFile(); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// 确保退出时清理PID文件
		defer ftrace.RemovePidFile()

		// 启动tracing
		if err := engine.StartTracing(daemonTracer, daemonFilter); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// 重置信号处理（覆盖main.go中的全局处理）
		signal.Reset(os.Interrupt, syscall.SIGTERM)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// 根据是否有duration决定行为
		if daemonDuration != "" {
			d, err := time.ParseDuration(daemonDuration)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: invalid duration: %v\n", err)
				engine.SafeShutdown()
				os.Exit(1)
			}

			// 等待duration或信号
			select {
			case <-sigChan:
				fmt.Println("\n[*] Stopping early...")
			case <-time.After(d):
				fmt.Println("\n[*] Duration completed.")
			}

			// 导出结果
			fmt.Printf("[*] Exporting trace to %s...\n", daemonOutput)
			if err := engine.StopAndExport(daemonOutput); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("[+] Trace saved to: %s\n", daemonOutput)
		} else {
			// 无限期运行，等待信号
			fmt.Printf("[*] Tracing is running indefinitely. Results will be saved to: %s\n", daemonOutput)
			fmt.Println("[*] Press Ctrl+C to stop and save results")
			<-sigChan
			fmt.Println("\n[*] Stopping and exporting results...")
			if err := engine.StopAndExport(daemonOutput); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("[+] Trace saved to: %s\n", daemonOutput)
		}
	},
}

func init() {
	rootCmd.AddCommand(backgroundDaemonCmd)

	backgroundDaemonCmd.Flags().StringVar(&daemonTracer, "tracer", "function", "Tracer type")
	backgroundDaemonCmd.Flags().StringVar(&daemonFilter, "filter", "", "Function filter")
	backgroundDaemonCmd.Flags().StringVar(&daemonDuration, "duration", "", "Tracing duration")
	backgroundDaemonCmd.Flags().StringVar(&daemonOutput, "output", "", "Output file path")
}
