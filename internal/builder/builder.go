package builder

type Builder interface {
    Build(projectPath string) error
    Run(projectPath string) error
}