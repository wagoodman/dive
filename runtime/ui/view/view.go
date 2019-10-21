package view

const (
	HeightFull = -1
	WidthFull = -1
	IdentityNone = ""
)


type Identifiable interface {
	Name()    string
}

type Dimensional interface {
	IsVisible() bool
	Height()    int
	Width()     int
}


// View defines the an element with state that can be updated, queried if visible, and render elements to the screen
type View interface {
	Identifiable
	Dimensional
	Update() error
	Render() error
	KeyHelp() string
}
