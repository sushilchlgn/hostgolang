package builder

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Builder interface
type Builder interface {
	Build(projectPath string) error
	Run(projectPath string) (string, error)
}

// Secure Unzip (Zip Slip protection)
func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// 🔒 Zip Slip protection
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

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
			outFile.Close()
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

// Find project root
func FindGoProjectRoot(base string) string {
	if _, err := os.Stat(filepath.Join(base, "main.go")); err == nil {
		return base
	}
	if _, err := os.Stat(filepath.Join(base, "go.mod")); err == nil {
		return base
	}

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
	return base
}