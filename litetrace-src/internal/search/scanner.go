package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func FastScan(tracefsPath string, pattern string, maxLimit int) ([]string, error) {
	file, err := os.Open(filepath.Join(tracefsPath, "available_filter_functions"))
	if err != nil {
		return nil, fmt.Errorf("failed to open available_filter_functions: %w", err)
	}
	defer file.Close()

	var results []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	startTime := time.Now()

	for scanner.Scan() {
		if time.Since(startTime) > 2*time.Second {
			fmt.Println("The regex is too complex...")
			break
		}

		line := scanner.Text()
		funcName := strings.SplitN(line, " ", 2)[0]

		if re.MatchString(funcName) {
			results = append(results, funcName)
			if len(results) >= maxLimit {
				break
			}
		}
	}
	return results, scanner.Err()
}

func ValidateFunction(tracefsPath string, funcName string) (bool, error) {
	file, err := os.Open(filepath.Join(tracefsPath, "available_filter_functions"))
	if err != nil {
		return false, fmt.Errorf("failed to open available_filter_functions: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		name := strings.SplitN(line, " ", 2)[0]
		if name == funcName {
			return true, nil
		}
	}
	return false, scanner.Err()
}
