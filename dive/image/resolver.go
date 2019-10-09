package image

type Resolver interface {
	Fetch(id string) (*Image, error)
	Build(options []string) (*Image, error)
}
