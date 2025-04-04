package view

// Controller defines the a renderable terminal screen pane.
type Renderer interface {
	Update() error
	Render() error
	IsVisible() bool
}

type Helper interface {
	KeyHelp() string
}
