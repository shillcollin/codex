package protocol

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGeneratedBindingsAreUpToDate(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	moduleRoot := filepath.Dir(cwd)
	currentPath := filepath.Join(cwd, "generated.go")
	current, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("read generated.go: %v", err)
	}

	tempDir := t.TempDir()
	tempOut := filepath.Join(tempDir, "generated.go")
	cmd := exec.Command("go", "run", "./cmd/generate", "-out", tempOut)
	cmd.Dir = moduleRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("regenerate bindings: %v\n%s", err, output)
	}

	regenerated, err := os.ReadFile(tempOut)
	if err != nil {
		t.Fatalf("read regenerated output: %v", err)
	}

	if !bytes.Equal(current, regenerated) {
		t.Fatal("generated.go drifted after regeneration")
	}
}
