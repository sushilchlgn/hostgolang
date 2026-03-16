package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// Run executes the built Go binary
func (b *GoBuilder) Run(projectPath string) error {
	binary := filepath.Join(projectPath, "app")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return fmt.Errorf("binary not found: %s", binary)
	}

	absBinary, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %s", err)
	}

	cmd := exec.Command(absBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start app: %s", err)
	}

	return nil
}