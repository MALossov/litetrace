package wizard

import (
	"fmt"
	"os"
	"runtime"

	"github.com/manifoldco/promptui"
)

type TracerType string

const (
	TracerFunction      TracerType = "function"
	TracerFunctionGraph TracerType = "function_graph"
	TracerNop           TracerType = "nop"
)

func AskTracer() (TracerType, error) {
	prompt := promptui.Select{
		Label: "Choose your tracer",
		Items: []string{
			"function        - Trace kernel function entries (lightweight)",
			"function_graph  - Trace function calls and returns (detailed)",
			"nop             - Disable tracing (safe mode)",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	switch result {
	case "function        - Trace kernel function entries (lightweight)":
		return TracerFunction, nil
	case "function_graph  - Trace function calls and returns (detailed)":
		return TracerFunctionGraph, nil
	default:
		return TracerNop, nil
	}
}

func AskFilter() (string, error) {
	prompt := promptui.Prompt{
		Label:   "Enter a kernel function to filter (e.g., vfs_read)",
		Default: "",
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func AskConfirm(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:   message,
		Default: "Y",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return false, nil
		}
		return false, err
	}

	return result == "y" || result == "Y", nil
}

func AskViewMode() (int, error) {
	prompt := promptui.Select{
		Label: "How do you want to view the kernel data?",
		Items: []string{
			"[1] Enter TUI Dashboard      - Open real-time monitoring screen",
			"[2] Run silently and Export  - Run 10s and save to file",
			"[3] Run in Background        - Detach and continue in background",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	switch result {
	case "[1] Enter TUI Dashboard      - Open real-time monitoring screen":
		return 1, nil
	case "[2] Run silently and Export  - Run 10s and save to file":
		return 2, nil
	default:
		return 3, nil
	}
}

func PrintWelcome() {
	_ = runtime.GOOS
	fmt.Println(`
  _                   _       ____             _       
 | |    ___  __ _  __| | ___ |  _ \ ___  _ __ | |_ ___ 
 | |   / _ \/ _` + "`" + ` |/ _` + "`" + ` |/ _ \| |_) / _ \| '_ \| __/ __|
 | |__|  __/ (_| | (_| |  __/|  __/ (_) | | | | |_\__ \
 |_____\___|\__,_|\__,_|\___||_|   \___/|_| |_|\__|___/
                                                       
Kernel: Linux | Tracefs: /sys/kernel/tracing`)
}

func PrintSuccess(message string) {
	fmt.Printf("[+] %s\n", message)
}

func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "[-] Error: %s\n", message)
}

func PrintWarning(message string) {
	fmt.Printf("[!] Warning: %s\n", message)
}
