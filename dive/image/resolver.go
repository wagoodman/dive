package image

type Resolver interface {
	Name() string
	Fetch(id string) (*Image, error)
	Build(options []string) (*Image, error)
}
