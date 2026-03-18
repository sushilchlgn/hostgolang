package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GoBuilder
type GoBuilder struct{}

// Build step
func (b *GoBuilder) Build(projectPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mod tidy failed:\n%s", string(out))
	}

	output := "app"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	cmd = exec.Command("go", "build", "-o", output)
	cmd.Dir = projectPath

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed:\n%s", string(out))
	}

	return nil
}

// Run step (capture output)
func (b *GoBuilder) Run(projectPath string) (string, error) {
	binary := filepath.Join(projectPath, "app")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return "", fmt.Errorf("binary not found: %s", binary)
	}

	absBinary, err := filepath.Abs(binary)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %s", err)
	}

	cmd := exec.Command(absBinary)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()

	output := out.String() + stderr.String()

	if err != nil {
		return output, fmt.Errorf("execution failed: %v", err)
	}

	return output, nil
}