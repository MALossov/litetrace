package ftrace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTracefs(t *testing.T) {
	tmpDir := t.TempDir()
	
	tracefsPaths = []string{tmpDir}

	os.WriteFile(filepath.Join(tmpDir, "available_tracers"), []byte("function nop"), 0644)

	path, err := FindTracefs()
	if err != nil {
		t.Errorf("FindTracefs() error = %v", err)
	}
	if path != tmpDir {
		t.Errorf("FindTracefs() = %q, want %q", path, tmpDir)
	}

	tracefsPaths = []string{}
	_, err = FindTracefs()
	if err == nil {
		t.Error("FindTracefs() should return error when no paths available")
	}
}

func TestFindTracefs_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	
	os.WriteFile(filepath.Join(tmpDir, "file"), []byte("test"), 0644)
	tracefsPaths = []string{filepath.Join(tmpDir, "file")}

	_, err := FindTracefs()
	if err == nil {
		t.Error("FindTracefs() should return error when path is not a directory")
	}
}
