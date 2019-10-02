package image

type Resolver interface {
	Get(id string) error
	Build(options []string) (string, error)
}
