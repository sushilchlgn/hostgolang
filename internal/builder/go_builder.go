package builder

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// GoBuilder handles building and running Go projects
type GoBuilder struct{}

// Build builds the Go project at the given path
func (b *GoBuilder) Build(projectPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mod tidy failed: %s\n%s", err, string(cmdOut))
	}

	cmd = exec.Command("go", "build", "-o", "app")
	cmd.Dir = projectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %s\n%s", err, string(out))
	}

	log.Println("Build successful:", projectPath)
	return nil
}

// Run executes the built Go binary
func (b *GoBuilder) Run(projectPath string) error {
	binary := filepath.Join(projectPath, "app")
	cmd := exec.Command(binary)
	cmd.Dir = projectPath

	// Optional: log output of the running app
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start app: %s", err)
	}

	log.Println("App started:", binary)
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

// FindGoProjectRoot attempts to locate the folder containing main.go or go.mod
func FindGoProjectRoot(base string) string {
	files, err := os.ReadDir(base)
	if err != nil {
		return base
	}

	for _, f := range files {
		if f.IsDir() {
			sub := filepath.Join(base, f.Name())
			if _, err := os.Stat(filepath.Join(sub, "go.mod")); err == nil {
				return sub
			}
			if _, err := os.Stat(filepath.Join(sub, "main.go")); err == nil {
				return sub
			}
		}
	}
	return base
}