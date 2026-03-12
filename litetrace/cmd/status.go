package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current ftrace status",
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

		engine := ftrace.NewEngine(tracefsPath)

		status, err := engine.GetStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		engineStatus := "🔴 STOPPED"
		if status.Enabled {
			engineStatus = "🟢 RUNNING"
		}

		fmt.Println("=========================================")
		fmt.Println("[ Ftrace Kernel Subsystem Status ]")
		fmt.Println("=========================================")
		fmt.Printf("- Engine Status : %s (tracing_on = %d)\n", engineStatus, func() int { if status.Enabled { return 1 }; return 0 }())
		fmt.Printf("- Current Tracer: %s\n", status.Tracer)
		fmt.Printf("- Active Filters: %s\n", status.Filter)
		fmt.Printf("- Buffer Size   : %d KB (Per CPU)\n", status.BufferSize)
		fmt.Printf("- Trace Clock   : %s\n", status.TraceClock)
		fmt.Println("=========================================")
	},
}
