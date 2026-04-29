package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFastScan(t *testing.T) {
	tmpDir := t.TempDir()
	functions := []string{
		"vfs_read [vmlinux]",
		"vfs_write [vmlinux]",
		"sys_read [vmlinux]",
		"sys_write [vmlinux]",
		"tcp_sendmsg [kernel]",
		"udp_recvmsg [kernel]",
	}
	content := []byte(strings.Join(functions, "\n") + "\n")
	os.WriteFile(filepath.Join(tmpDir, "available_filter_functions"), content, 0644)

	tests := []struct {
		name      string
		pattern   string
		maxLimit  int
		wantCount int
	}{
		{"match vfs_", "vfs_", 100, 2},
		{"match sys_", "sys_", 100, 2},
		{"match tcp_", "tcp_", 100, 1},
		{"match nothing", "nonexistent", 100, 0},
		{"match with limit", "vfs_", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := FastScan(tmpDir, tt.pattern, tt.maxLimit)
			if err != nil {
				t.Errorf("FastScan() error = %v", err)
				return
			}
			if len(results) != tt.wantCount {
				t.Errorf("FastScan() got %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestFastScan_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	
	largeContent := make([]string, 10000)
	for i := range largeContent {
		largeContent[i] = "func_test"
	}
	content := []byte(strings.Join(largeContent, "\n") + "\n")
	os.WriteFile(filepath.Join(tmpDir, "available_filter_functions"), content, 0644)

	results, err := FastScan(tmpDir, "func_", 100)
	if err != nil {
		t.Errorf("FastScan() error = %v", err)
	}
	
	if len(results) > 100 {
		t.Errorf("FastScan() got %d results, expected max 100", len(results))
	}
}

func TestValidateFunction(t *testing.T) {
	tmpDir := t.TempDir()
	functions := []string{
		"vfs_read [vmlinux]",
		"sys_read [vmlinux]",
	}
	content := []byte(strings.Join(functions, "\n") + "\n")
	os.WriteFile(filepath.Join(tmpDir, "available_filter_functions"), content, 0644)

	tests := []struct {
		name      string
		funcName  string
		wantValid bool
	}{
		{"valid function", "vfs_read", true},
		{"valid function 2", "sys_read", true},
		{"invalid function", "nonexistent", false},
		{"partial match should fail", "vfs_", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateFunction(tmpDir, tt.funcName)
			if err != nil {
				t.Errorf("ValidateFunction() error = %v", err)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("ValidateFunction() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}
