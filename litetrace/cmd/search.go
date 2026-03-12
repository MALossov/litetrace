package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
	"github.com/malossov/lite-tracer-mygo/internal/search"
)

var searchCmd = &cobra.Command{
	Use:   "search <pattern>",
	Short: "Search kernel functions",
	Args:  cobra.ExactArgs(1),
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

		pattern := args[0]

		results, err := search.FastScan(tracefsPath, pattern, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
			os.Exit(1)
		}

		for _, result := range results {
			fmt.Println(result)
		}

		if len(results) == 0 {
			fmt.Println("No matching functions found.")
		}
	},
}
