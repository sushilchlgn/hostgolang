package builder

import (
	"os/exec"
	"path/filepath"
)

type GoBuilder struct{}

func (b GoBuilder) Build(projectPath string) error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(projectPath, "app"))
	cmd.Dir = projectPath
	return cmd.Run()
}

func (b GoBuilder) Run(projectPath string) error {
	cmd := exec.Command(filepath.Join(projectPath, "app"))
	cmd.Dir = projectPath
	return cmd.Start()
}