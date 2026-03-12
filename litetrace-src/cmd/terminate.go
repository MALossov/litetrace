package cmd

import (
	"fmt"
	"os"

	"github.com/M410550/lite-tracer-mygo/internal/ftrace"
	"github.com/spf13/cobra"
)

var terminateCmd = &cobra.Command{
	Use:   "terminate",
	Short: "Terminate tracing",
	Long:  "Terminate tracing and remove all tracefs files.",
	Run: func(cmd *cobra.Command, args []string) {
		// 首先检查是否有后台进程在运行
		running, pid, _ := ftrace.IsDaemonRunning()
		if running {
			fmt.Printf("[*] Stopping background process (PID %d)...\n", pid)
			if err := ftrace.StopDaemon(); err != nil {
				fmt.Fprintf(os.Stderr, "[!] Warning: %v\n", err)
			} else {
				fmt.Println("[+] Background process stopped successfully.")
			}
		}

		tracefsPath, err := ftrace.FindTracefs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}
		engine := ftrace.NewEngineWithDebug(tracefsPath, debugFlag)
		engine.SafeShutdown(true)
		fmt.Println("[+] Tracing terminated successfully.")
		fmt.Println()
		fmt.Println("Current status:")

		// 在后面再调用一个status命令
		statusCmd.Run(cmd, []string{"status"})
	},
}
