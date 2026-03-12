package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/ui"
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

		if err := dashboard.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}
	},
}
