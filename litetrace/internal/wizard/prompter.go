package wizard

import (
	"os"
	"runtime"

	"github.com/fatih/color"
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
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F336 {{ . | red }}",
			Inactive: "  {{ . | cyan }}",
			Selected: "\U0001F336 {{ . | red | cyan }}",
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
		Label:   "Enter a kernel function to filter",
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
		Label:     message,
		Default:   "Y",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return false, nil
		}
		return false, err
	}

	return result == "y" || result == "Y" || result == "", nil
}

func AskViewMode() (int, error) {
	prompt := promptui.Select{
		Label: "How do you want to view the kernel data?",
		Items: []string{
			"[1] Enter TUI Dashboard      - Open real-time monitoring screen",
			"[2] Run silently and Export  - Run for specified duration and save to file",
			"[3] Run in Background        - Detach and continue in background",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F336 {{ . | red }}",
			Inactive: "  {{ . | cyan }}",
			Selected: "\U0001F336 {{ . | red | cyan }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	switch result {
	case "[1] Enter TUI Dashboard      - Open real-time monitoring screen":
		return 1, nil
	case "[2] Run silently and Export  - Run for specified duration and save to file":
		return 2, nil
	default:
		return 3, nil
	}
}

func AskDuration() (string, error) {
	prompt := promptui.Select{
		Label: "Select tracing duration",
		Items: []string{
			"10 seconds",
			"30 seconds",
			"1 minute",
			"5 minutes",
			"10 minutes",
			"Until I stop it",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F336 {{ . | red }}",
			Inactive: "  {{ . | cyan }}",
			Selected: "\U0001F336 {{ . | red | cyan }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	switch result {
	case "10 seconds":
		return "10s", nil
	case "30 seconds":
		return "30s", nil
	case "1 minute":
		return "1m", nil
	case "5 minutes":
		return "5m", nil
	case "10 minutes":
		return "10m", nil
	default:
		return "", nil
	}
}

func PrintWelcome() {
	_ = runtime.GOOS
	c := color.New(color.FgCyan, color.Bold)
	c.Println(`
//  _    _ _____ __________ ____ ____ ____ _____    _    ___  __________ _ 
// / \  / /__ __/  __/__ __/  __/  _ /   _/  __/   / \__/\  \//  __/  _ / \
// | |  | | / \ |  \   / \ |  \/| / \|  / |  \_____| |\/||\  /| |  | / \| |
// | |_/| | | | |  /_  | | |    | |-||  \_|  /\____| |  ||/ / | |_/| \_/\_/
// \____\_/ \_/ \____\ \_/ \_/\_\_/ \\____\____\   \_/  \/_/  \____\____(_)
//   

Kernel: Linux | Tracefs: /sys/kernel/tracing`)
}

func PrintSuccess(message string) {
	c := color.New(color.FgGreen, color.Bold)
	c.Printf("[+] %s\n", message)
}

func PrintError(message string) {
	c := color.New(color.FgRed, color.Bold)
	c.Fprintf(os.Stderr, "[-] Error: %s\n", message)
}

func PrintWarning(message string) {
	c := color.New(color.FgYellow, color.Bold)
	c.Printf("[!] Warning: %s\n", message)
}
