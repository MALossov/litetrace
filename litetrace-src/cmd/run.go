package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/search"
	"github.com/malossov/lite-tracer-mygo/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	tracerFlag   string
	filterFlag   string
	durationFlag string
	outputFlag   string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tracing with specified parameters",
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

		duration, err := time.ParseDuration(durationFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: Invalid duration: %v\n", err)
			os.Exit(1)
		}

		if duration < 0 || duration > 24*time.Hour {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: Duration must be between 0 and 24 hours\n")
			os.Exit(1)
		}

		if outputFlag == "" {
			fmt.Fprintln(os.Stderr, "🚨 Fatal: --output is required")
			os.Exit(1)
		}

		if tracerFlag == "nop" {
			fmt.Println("[!] WARNING: Tracer is 'nop' - no data will be captured!")
			fmt.Println("    Use 'function' or 'function_graph' tracer to capture data.")
		}

		result := search.ValidateAndNormalizeFilter(tracefsPath, filterFlag, tracerFlag)
		if result.ShouldIgnore {
			fmt.Printf("[!] %s\n", result.Message)
			filterFlag = ""
		} else if !result.Valid {
			fmt.Printf("[!] %s\n", result.Message)
			wizard.PrintWarning("No filter specified. System may experience high load.")
		} else if filterFlag != "" {
			fmt.Printf("[+] %s\n", result.Message)
		}

		if err := engine.RunWithDuration(tracerFlag, filterFlag, outputFlag, duration); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().StringVar(&tracerFlag, "tracer", "function", "Tracer type: function, function_graph")
	runCmd.Flags().StringVar(&filterFlag, "filter", "", "Function filter")
	runCmd.Flags().StringVar(&durationFlag, "duration", "10s", "Duration (e.g., 5s, 1m)")
	runCmd.Flags().StringVar(&outputFlag, "output", "", "Output file path (required)")
	runCmd.MarkFlagRequired("output")
}
