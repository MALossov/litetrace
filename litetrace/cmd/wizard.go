package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/search"
	"github.com/malossov/lite-tracer-mygo/internal/ui"
	"github.com/malossov/lite-tracer-mygo/internal/wizard"
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
			if err := dashboard.Run(); err != nil {
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
			// Background mode
			fmt.Println("[*] Tracing is running in background.")

			if duration != "" {
				d, _ := time.ParseDuration(duration)
				fmt.Printf("[*] Will run for %v...\n", d)
				fmt.Println("[*] Press Ctrl+C to stop early")

				// Setup signal handling
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

				// Wait for either duration or signal
				select {
				case <-sigChan:
					fmt.Println("\n[*] Stopping early...")
				case <-time.After(d):
					fmt.Println("\n[*] Duration completed.")
				}

				// Export results
				timestamp := time.Now().Format("20060102_150405")
				outputPath := fmt.Sprintf("/tmp/litetrace_wizard_%s.txt", timestamp)

				fmt.Printf("[*] Exporting trace to %s...\n", outputPath)
				if err := engine.StopAndExport(outputPath); err != nil {
					fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
					os.Exit(1)
				}
				wizard.PrintSuccess(fmt.Sprintf("Trace saved to %s", outputPath))
			} else {
				fmt.Println("[*] To stop tracing and view results:")
				fmt.Println("    1. Run 'litetrace status' to check status")
				fmt.Println("    2. Run 'litetrace run --output /path/to/file' to export results")
				fmt.Println("\n[!] Press Ctrl+C to stop and cleanup immediately")

				// Wait for interrupt with cleanup
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
				<-sigChan

				fmt.Println("\n[*] Cleaning up...")
				engine.SafeShutdown()
				wizard.PrintSuccess("Tracing stopped and cleaned up")
			}
		}
	},
}
