package cmd

import (
	"fmt"
	"os"

	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/ui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch TUI dashboard for real-time monitoring",
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

		dashboard := ui.NewDashboard(engine)

		savedFile, err := dashboard.Run()

		// 无论成功与否，先输出保存信息
		if savedFile != "" {
			fmt.Fprintf(os.Stderr, "\n[✓] Trace data saved to: %s\n", savedFile)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}
	},
}
