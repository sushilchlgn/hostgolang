package builder

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"path/filepath"
)

// GoBuilder handles building and running Go projects
type GoBuilder struct{}

// Build builds the Go project at the given path
func (b *GoBuilder) Build(projectPath string) error {
    // Run go mod tidy
    cmd := exec.Command("go", "mod", "tidy")
    cmd.Dir = projectPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("mod tidy failed: %s\n%s", err, string(out))
    }

    // Determine output name for Windows vs others
    output := "app"
    if runtime.GOOS == "windows" {
        output += ".exe"
    }

    cmd = exec.Command("go", "build", "-o", output)
    cmd.Dir = projectPath
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("go build failed: %s\n%s", err, string(out))
    }

    log.Println("Build successful:", projectPath, "binary:", output)
    return nil
}

// Run executes the built Go binary
func (b *GoBuilder) Run(projectPath string) error {
    binary := filepath.Join(projectPath, "app")

    // Append .exe on Windows
    if runtime.GOOS == "windows" {
        binary += ".exe"
    }

    if _, err := os.Stat(binary); os.IsNotExist(err) {
        return fmt.Errorf("binary not found: %s", binary)
    }

    absBinary, err := filepath.Abs(binary)
    if err != nil {
        return fmt.Errorf("failed to get absolute path of binary: %s", err)
    }

    cmd := exec.Command(absBinary)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err = cmd.Start()
    if err != nil {
        return fmt.Errorf("failed to start app: %s", err)
    }

    return nil
}

// Unzip extracts a zip archive to the specified destination
func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// FindGoProjectRoot recursively finds the folder containing main.go or go.mod
func FindGoProjectRoot(base string) string {
	// Check if base itself has main.go or go.mod
	if _, err := os.Stat(filepath.Join(base, "main.go")); err == nil {
		return base
	}
	if _, err := os.Stat(filepath.Join(base, "go.mod")); err == nil {
		return base
	}

	// Check first-level subfolders
	files, err := os.ReadDir(base)
	if err != nil {
		return base
	}

	for _, f := range files {
		if f.IsDir() {
			sub := filepath.Join(base, f.Name())
			if _, err := os.Stat(filepath.Join(sub, "main.go")); err == nil {
				return sub
			}
			if _, err := os.Stat(filepath.Join(sub, "go.mod")); err == nil {
				return sub
			}
		}
	}

	// Default to base
	return base
}
