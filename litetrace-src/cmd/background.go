package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/M410550/lite-tracer-mygo/internal/ftrace"
	"github.com/spf13/cobra"
)

var (
	bgTracer   string
	bgFilter   string
	bgDuration string
	bgOutput   string
)

var backgroundCmd = &cobra.Command{
	Use:    "background",
	Short:  "Run tracing in background (internal use)",
	Long:   "This command is used internally to run tracing in background. Use 'litetrace wizard' to start background tracing interactively.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if !checkRoot() {
			fmt.Fprintln(os.Stderr, "🚨 Fatal: Root privileges required")
			os.Exit(1)
		}

		// 确定输出路径
		outputPath := bgOutput
		if outputPath == "" {
			timestamp := time.Now().Format("20060102_150405")
			outputPath = fmt.Sprintf("/tmp/litetrace_background_%s.txt", timestamp)
		}

		// 启动后台进程
		if err := ftrace.StartBackgroundProcess(bgTracer, bgFilter, bgDuration, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// 等待一小段时间确保进程启动成功
		time.Sleep(100 * time.Millisecond)

		// 获取后台进程PID
		running, pid, err := ftrace.IsDaemonRunning()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: failed to check daemon status: %v\n", err)
			os.Exit(1)
		}

		if !running {
			fmt.Fprintln(os.Stderr, "🚨 Fatal: background process failed to start")
			os.Exit(1)
		}

		fmt.Printf("[+] Background tracing started with PID %d\n", pid)
		if bgDuration != "" {
			fmt.Printf("[*] Will run for %s\n", bgDuration)
		}
		fmt.Printf("[*] Results will be saved to: %s\n", outputPath)
		fmt.Println("[*] You can now continue using your shell.")
	},
}

func init() {
	rootCmd.AddCommand(backgroundCmd)

	backgroundCmd.Flags().StringVar(&bgTracer, "tracer", "function", "Tracer type")
	backgroundCmd.Flags().StringVar(&bgFilter, "filter", "", "Function filter")
	backgroundCmd.Flags().StringVar(&bgDuration, "duration", "", "Tracing duration")
	backgroundCmd.Flags().StringVar(&bgOutput, "output", "", "Output file path")
}
