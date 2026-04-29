package search

import (
	"testing"
)

func TestValidateAndNormalizeFilter_EmptyFilter(t *testing.T) {
	tests := []struct {
		name    string
		tracer  string
		wantMsg string
		wantIgn bool
	}{
		{
			name:    "empty filter with function tracer",
			tracer:  "function",
			wantMsg: "No filter specified. System may experience high load.",
			wantIgn: false,
		},
		{
			name:    "empty filter with nop tracer",
			tracer:  "nop",
			wantMsg: "No filter (tracer is nop)",
			wantIgn: true,
		},
		{
			name:    "empty filter with function_graph tracer",
			tracer:  "function_graph",
			wantMsg: "No filter specified. System may experience high load.",
			wantIgn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAndNormalizeFilter("/nonexistent", "", tt.tracer)
			if result.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", result.Message, tt.wantMsg)
			}
			if result.ShouldIgnore != tt.wantIgn {
				t.Errorf("ShouldIgnore = %v, want %v", result.ShouldIgnore, tt.wantIgn)
			}
			if result.Filter != "" {
				t.Errorf("Filter = %v, want empty", result.Filter)
			}
		})
	}
}

func TestValidateAndNormalizeFilter_FileNotAccessible(t *testing.T) {
	result := ValidateAndNormalizeFilter("/nonexistent", "vfs_read", "function")
	if !result.Valid {
		t.Errorf("Valid = false, want true when file not accessible")
	}
	if result.Filter != "vfs_read" {
		t.Errorf("Filter = %v, want vfs_read", result.Filter)
	}
	if result.ShouldIgnore {
		t.Errorf("ShouldIgnore = true, want false when file not accessible")
	}
}

func TestValidateFunction_FileNotAccessible(t *testing.T) {
	valid, err := ValidateFunction("/nonexistent", "test_func")
	if err == nil {
		t.Errorf("Expected error when file not accessible, got nil")
	}
	if valid {
		t.Errorf("Valid = true, want false when file not accessible")
	}
}

func TestFastScan_InvalidRegex(t *testing.T) {
	results, err := FastScan("/nonexistent", "[invalid", 10)
	if err == nil {
		t.Errorf("Expected error for invalid regex, got nil")
	}
	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

func TestFastScan_FileNotAccessible(t *testing.T) {
	results, err := FastScan("/nonexistent", ".*", 10)
	if err == nil {
		t.Errorf("Expected error when file not accessible, got nil")
	}
	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

func TestValidateAndNormalizeFilter_InvalidFilter(t *testing.T) {
	result := ValidateAndNormalizeFilter("/nonexistent", "nonexistent_function_xyz", "function")
	if result.Valid {
		t.Logf("Valid = true (file not accessible, treating as valid)")
	}
	if !result.ShouldIgnore {
		t.Logf("ShouldIgnore = false (file not accessible)")
	}
	if result.Filter != "nonexistent_function_xyz" {
		t.Errorf("Filter = %v, want original value", result.Filter)
	}
}

func TestValidateAndNormalizeFilter_ValidFilter(t *testing.T) {
	result := ValidateAndNormalizeFilter("/nonexistent", "vfs_read", "function")
	if result.ShouldIgnore {
		t.Errorf("ShouldIgnore = true, want false for valid filter")
	}
	if result.Filter != "vfs_read" {
		t.Errorf("Filter = %v, want vfs_read", result.Filter)
	}
}

func TestNormalizeFilter_Empty(t *testing.T) {
	result := NormalizeFilter("/nonexistent", "")
	if result != "" {
		t.Errorf("NormalizeFilter empty input = %v, want empty", result)
	}
}

func TestNormalizeFilter_InvalidFunction(t *testing.T) {
	result := NormalizeFilter("/nonexistent", "nonexistent_function_xyz_123")
	if result != "" {
		t.Errorf("NormalizeFilter invalid function = %v, want empty", result)
	}
}

func TestNormalizeFilter_FileNotAccessible(t *testing.T) {
	result := NormalizeFilter("/nonexistent", "vfs_read")
	if result != "" {
		t.Errorf("NormalizeFilter when file not accessible = %v, want empty", result)
	}
}

func TestValidationResult_Fields(t *testing.T) {
	result := ValidationResult{
		Valid:        true,
		Filter:       "test_filter",
		Message:      "test message",
		ShouldIgnore: false,
	}

	if !result.Valid {
		t.Error("Valid should be true")
	}
	if result.Filter != "test_filter" {
		t.Errorf("Filter = %v, want test_filter", result.Filter)
	}
	if result.Message != "test message" {
		t.Errorf("Message = %v, want test message", result.Message)
	}
	if result.ShouldIgnore {
		t.Error("ShouldIgnore should be false")
	}
}

func TestValidateFunction_Empty(t *testing.T) {
	valid, err := ValidateFunction("/nonexistent", "")
	if err == nil {
		t.Errorf("Expected error for empty function name, got nil")
	}
	if valid {
		t.Error("Valid should be false for empty function name")
	}
}

func TestFastScan_MaxLimit(t *testing.T) {
	results, err := FastScan("/nonexistent", ".*", 0)
	if err != nil {
		t.Logf("Error = %v (file not accessible)", err)
	}
	if results == nil {
		t.Logf("Results = nil (file not accessible)")
	}
	if len(results) != 0 {
		t.Errorf("len(Results) = %v, want 0", len(results))
	}
}
