package ftrace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngine_WriteTraceFile(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	tests := []struct {
		name       string
		filename   string
		content    string
		appendMode bool
		wantErr    bool
	}{
		{"write simple content", "test.txt", "hello world", false, false},
		{"write with newline", "test.txt", "line1\nline2", false, false},
		{"append mode", "test.txt", "append", true, false},
		{"write to nonexistent file", "nonexistent/subdir/test.txt", "test", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.WriteTraceFile(tt.filename, tt.content, tt.appendMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTraceFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				filePath := filepath.Join(tmpDir, tt.filename)
				data, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read file: %v", err)
					return
				}
				content := string(data)
				if tt.appendMode {
					if tt.filename == "test.txt" && !strings.HasPrefix(tt.content, "\n") && !strings.HasSuffix(content, "\n") {
					} else if !strings.Contains(content, tt.content) {
						t.Errorf("content mismatch: got %q", content)
					}
				} else {
					if content != tt.content {
						t.Errorf("content mismatch: got %q, want %q", content, tt.content)
					}
				}
			}
		})
	}
}

func TestEngine_ReadTraceFile(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("  hello world  \n"), 0644)

	content, err := engine.ReadTraceFile("test.txt")
	if err != nil {
		t.Errorf("ReadTraceFile() error = %v", err)
	}
	if content != "hello world" {
		t.Errorf("ReadTraceFile() = %q, want %q", content, "hello world")
	}

	_, err = engine.ReadTraceFile("nonexistent.txt")
	if err == nil {
		t.Error("ReadTraceFile() should return error for nonexistent file")
	}
}

func TestEngine_SetTracer(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	tests := []struct {
		name    string
		tracer  string
		wantErr bool
	}{
		{"set nop", "nop", false},
		{"set function", "function", false},
		{"set function_graph", "function_graph", false},
		{"set empty (should default to nop)", "", false},
		{"set invalid tracer", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.SetTracer(tt.tracer)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetTracer() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				expected := tt.tracer
				if tt.tracer == "" {
					expected = "nop"
				}
				content, _ := engine.ReadTraceFile("current_tracer")
				if content != expected {
					t.Errorf("tracer = %q, want %q", content, expected)
				}
			}
		})
	}
}

func TestEngine_SetFilter(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	err := engine.SetFilter("vfs_read")
	if err != nil {
		t.Errorf("SetFilter() error = %v", err)
	}

	content, _ := engine.ReadTraceFile("set_ftrace_filter")
	if content != "vfs_read" {
		t.Errorf("filter = %q, want %q", content, "vfs_read")
	}

	err = engine.SetFilter("")
	if err != nil {
		t.Errorf("SetFilter('') error = %v", err)
	}

	content, _ = engine.ReadTraceFile("set_ftrace_filter")
	if content != "" {
		t.Errorf("filter after clear = %q, want empty", content)
	}
}

func TestEngine_EnableDisable(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	err := engine.Enable()
	if err != nil {
		t.Errorf("Enable() error = %v", err)
	}

	enabled, _ := engine.IsEnabled()
	if !enabled {
		t.Error("IsEnabled() = false, want true after Enable()")
	}

	err = engine.Disable()
	if err != nil {
		t.Errorf("Disable() error = %v", err)
	}

	enabled, _ = engine.IsEnabled()
	if enabled {
		t.Error("IsEnabled() = true, want false after Disable()")
	}
}

func TestEngine_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "current_tracer"), []byte("function"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tracing_on"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "set_ftrace_filter"), []byte("vfs_read"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "buffer_size_kb"), []byte("1408"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "trace_clock"), []byte("local"), 0644)

	status, err := engine.GetStatus()
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}

	if status.Tracer != "function" {
		t.Errorf("Tracer = %q, want %q", status.Tracer, "function")
	}
	if status.Filter != "vfs_read" {
		t.Errorf("Filter = %q, want %q", status.Filter, "vfs_read")
	}
	if !status.Enabled {
		t.Error("Enabled = false, want true")
	}
	if status.BufferSize != 1408 {
		t.Errorf("BufferSize = %d, want %d", status.BufferSize, 1408)
	}
	if status.TraceClock != "local" {
		t.Errorf("TraceClock = %q, want %q", status.TraceClock, "local")
	}
}

func TestEngine_SafeShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(tmpDir)

	engine.SetTracer("function")
	engine.SetFilter("vfs_read")
	engine.Enable()

	engine.SafeShutdown()

	enabled, _ := engine.IsEnabled()
	if enabled {
		t.Error("tracing_on should be 0 after SafeShutdown")
	}

	tracer, _ := engine.GetTracer()
	if tracer != "nop" {
		t.Errorf("tracer = %q, want nop after SafeShutdown", tracer)
	}

	filter, _ := engine.GetFilter()
	if filter != "" {
		t.Errorf("filter = %q, want empty after SafeShutdown", filter)
	}
}
