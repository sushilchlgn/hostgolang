package builder

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// GoBuilder implements Builder for Go projects
type GoBuilder struct{}

// Build builds the Go project at the given path
func (b *GoBuilder) Build(projectPath string) error {
	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mod tidy failed: %s\n%s", err, string(out))
	}

	// Build the project binary
	output := "app"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}
	cmd = exec.Command("go", "build", "-o", output)
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %s\n%s", err, string(out))
	}

	return nil
}

// Run executes the built Go binary safely with timeout
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

	// 🔹 Timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, absBinary)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := out.String() + stderr.String()

	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("execution timed out")
	}

	if err != nil {
		return output, fmt.Errorf("execution failed: %v", err)
	}

	return output, nil
}