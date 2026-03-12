package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/wizard"
	"github.com/malossov/lite-tracer-mygo/internal/search"
	"github.com/malossov/lite-tracer-mygo/internal/ui"
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
			valid, err := search.ValidateFunction(tracefsPath, filter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			if !valid {
				wizard.PrintError(fmt.Sprintf("Function '%s' not found in kernel symbols. Please try again.", filter))
				os.Exit(1)
			}
			wizard.PrintSuccess(fmt.Sprintf("Function '%s' validated successfully", filter))
		} else {
			wizard.PrintWarning("No filter specified. System may experience high load.")
		}

		// Step 4: Confirm configuration
		confirmMsg := fmt.Sprintf("Start tracing with [%s] tracer and filter [%s]? (Y/n)", tracer, filter)
		confirmed, err := wizard.AskConfirm(confirmMsg)
		if err != nil || !confirmed {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}

		// Step 5: Choose view mode BEFORE starting tracing
		viewMode, err := wizard.AskViewMode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		// Step 6: Start tracing
		if err := engine.StartTracing(string(tracer), filter); err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		wizard.PrintSuccess("Tracing started!")

		// Step 7: Execute based on view mode
		switch viewMode {
		case 1:
			// TUI Dashboard mode
			fmt.Println("[*] Launching TUI Dashboard...")
			time.Sleep(500 * time.Millisecond)
			dashboard := ui.NewDashboard(engine)
			if err := dashboard.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}

		case 2:
			// Silent mode with export
			fmt.Println("[*] Running silently for 10 seconds...")
			time.Sleep(10 * time.Second)
			
			timestamp := time.Now().Format("20060102_150405")
			outputPath := fmt.Sprintf("/tmp/litetrace_wizard_%s.txt", timestamp)
			
			fmt.Printf("[*] Exporting trace to %s...\n", outputPath)
			if err := engine.StopAndExport(outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
				os.Exit(1)
			}
			wizard.PrintSuccess(fmt.Sprintf("Trace saved to %s", outputPath))

		case 3:
			// Background mode - just print instructions
			fmt.Println("[*] Tracing is running in background.")
			fmt.Println("[*] To stop tracing and view results:")
			fmt.Println("    1. Run 'litetrace status' to check status")
			fmt.Println("    2. Run 'litetrace run --output /path/to/file' to export results")
			fmt.Println("\n[!] Press Ctrl+C to stop and cleanup immediately")
			
			// Wait for interrupt
			select {}
		}
	},
}
