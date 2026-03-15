package builder

import (
	"os/exec"
)

type GoBuilder struct{}

func (b GoBuilder) Build(path string) error {
	cmd := exec.Command("go", "build", "-o", "app")
	cmd.Dir = path
	return cmd.Run()
}

func (b GoBuilder) Run(path string) error {
	cmd := exec.Command("./app")
	cmd.Dir = path
	return cmd.Start()
}