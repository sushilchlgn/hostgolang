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

type GoBuilder struct{}

func (b *GoBuilder) Build(projectPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mod tidy failed: %s\n%s", err, string(out))
	}

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

// ✅ New function to run with timeout and capture output
func (b *GoBuilder) RunWithTimeout(projectPath string, timeout time.Duration) (string, error) {
	binary := filepath.Join(projectPath, "app")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return "", fmt.Errorf("binary not found: %s", binary)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Start()
	if err != nil {
		return out.String(), fmt.Errorf("failed to start app: %s", err)
	}

	err = cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return out.String(), fmt.Errorf("execution timed out")
	}

	return out.String(), err
}