package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var tldrCmd = &cobra.Command{
	Use:   "tldr",
	Short: "Print quick help",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(`
===============================================
              LITETRACE QUICK HELP
===============================================

USAGE:
  litetrace <command> [flags]

COMMANDS:
  run         Run tracing with specified parameters
              Example: litetrace run --tracer function --filter vfs_read --duration 5s --output trace.txt

  wizard      Interactive wizard mode
              Example: litetrace wizard

  search      Search kernel functions
              Example: litetrace search "^vfs_"

  status      Show current ftrace status
              Example: litetrace status

  tui         Launch TUI dashboard
              Example: litetrace tui

  tldr        Print this help

FLAGS:
  --tracer    Tracer type: function, function_graph (default: function)
  --filter    Function filter (e.g., vfs_read, tcp_*)
  --duration  Tracing duration (e.g., 5s, 1m) (default: 10s)
  --output    Output file path (required for run)

EXAMPLES:
  # Search for functions
  sudo litetrace search "^sys_"

  # Check ftrace status
  sudo litetrace status

  # Run a quick trace
  sudo litetrace run --tracer function --filter vfs_read --duration 5s --output /tmp/trace.txt

  # Interactive wizard
  sudo litetrace wizard

  # Real-time TUI
  sudo litetrace tui

NOTES:
  - All commands require root privileges
  - tracer options: function (lightweight), function_graph (detailed)
  - Use search to find available kernel functions

===============================================
`)
		os.Exit(0)
	},
}
