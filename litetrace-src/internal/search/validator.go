package search

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ValidationResult struct {
	Valid        bool
	Filter       string
	Message      string
	ShouldIgnore bool
}

func ValidateAndNormalizeFilter(tracefsPath string, filter string, tracer string) ValidationResult {
	if filter == "" {
		if tracer != "nop" {
			return ValidationResult{
				Valid:        false,
				Filter:       "",
				Message:      "No filter specified. System may experience high load.",
				ShouldIgnore: false,
			}
		}
		return ValidationResult{
			Valid:        true,
			Filter:       "",
			Message:      "No filter (tracer is nop)",
			ShouldIgnore: true,
		}
	}

	availablePath := filepath.Join(tracefsPath, "available_filter_functions")
	if _, err := os.Stat(availablePath); err != nil {
		return ValidationResult{
			Valid:        true,
			Filter:       filter,
			Message:      fmt.Sprintf("Cannot verify filter '%s': available_filter_functions not accessible", filter),
			ShouldIgnore: false,
		}
	}

	valid, err := ValidateFunction(tracefsPath, filter)
	if err != nil {
		return ValidationResult{
			Valid:        true,
			Filter:       filter,
			Message:      fmt.Sprintf("Filter verification error: %v", err),
			ShouldIgnore: false,
		}
	}

	if !valid {
		return ValidationResult{
			Valid:        true,
			Filter:       filter,
			Message:      fmt.Sprintf("Function '%s' not found in kernel symbols. Filter will not ignore but take your attention!", filter),
			ShouldIgnore: false,
		}
	}

	return ValidationResult{
		Valid:        true,
		Filter:       filter,
		Message:      fmt.Sprintf("Filter '%s' validated successfully", filter),
		ShouldIgnore: false,
	}
}

func NormalizeFilter(tracefsPath string, filter string) string {
	if filter == "" {
		return ""
	}

	filter = strings.TrimSpace(filter)

	parts := strings.SplitN(filter, "*", 2)
	if len(parts) == 2 && parts[0] == "" {
		prefix := parts[1]
		matches, err := FastScan(tracefsPath, "^"+prefix, 1)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}

	valid, err := ValidateFunction(tracefsPath, filter)
	if err == nil && valid {
		return filter
	}

	return ""
}
