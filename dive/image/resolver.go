package image

type Resolver interface {
	Resolve(id string) (Analyzer, error)
	Build(options []string) (string, error)
}
