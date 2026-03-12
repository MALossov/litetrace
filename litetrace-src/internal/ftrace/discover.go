package ftrace

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	tracefsPaths = []string{
		"/sys/kernel/tracing",
		"/sys/kernel/debug/tracing",
	}
)

func FindTracefs() (string, error) {
	for _, path := range tracefsPaths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			testFile := filepath.Join(path, "available_tracers")
			if _, err := os.Stat(testFile); err == nil {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("tracefs not found. Please mount with: mount -t tracefs nodev /sys/kernel/tracing")
}
