package view

import (
	"fmt"
	"github.com/anchore/go-logger"
	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/utils"
)

// Debug is just for me :)
type Debug struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	logger logger.Logger

	selectedView Helper
}

// newDebugView creates a new view object attached the global [gocui] screen object.
func newDebugView(gui *gocui.Gui) *Debug {
	c := new(Debug)

	// populate main fields
	c.name = "debug"
	c.gui = gui
	c.logger = log.Nested("ui", "debug")

	return c
}

func (v *Debug) SetCurrentView(r Helper) {
	v.selectedView = r
}

func (v *Debug) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *Debug) Setup(view *gocui.View, header *gocui.View) error {
	v.logger.Trace("setup()")

	// set controller options
	v.view = view
	v.view.Editable = false
	v.view.Wrap = false
	v.view.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	return v.Render()
}

// IsVisible indicates if the status view pane is currently initialized.
func (v *Debug) IsVisible() bool {
	return v != nil
}

// Update refreshes the state objects for future rendering (currently does nothing).
func (v *Debug) Update() error {
	return nil
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (v *Debug) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Render flushes the state objects to the screen.
func (v *Debug) Render() error {
	v.logger.Trace("render()")

	v.gui.Update(func(g *gocui.Gui) error {
		// update header...
		v.header.Clear()
		width, _ := g.Size()
		headerStr := format.RenderHeader("Debug", width, false)
		_, _ = fmt.Fprintln(v.header, headerStr)

		// update view...
		v.view.Clear()
		_, err := fmt.Fprintln(v.view, "blerg")
		if err != nil {
			v.logger.WithFields("error", err).Debug("unable to write to buffer")
		}

		return nil
	})
	return nil
}

func (v *Debug) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	v.logger.Tracef("layout(minX: %d, minY: %d, maxX: %d, maxY: %d)", minX, minY, maxX, maxY)

	// header
	headerSize := 1
	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(v.Name()+"header", minX, minY, maxX, minY+headerSize+1, 0)
	// we are going to overlap the view over the (invisible) border (so minY will be one less than expected).
	// additionally, maxY will be bumped by one to include the border
	view, viewErr := g.SetView(v.Name(), minX, minY+headerSize, maxX, maxY+1, 0)
	if utils.IsNewView(viewErr, headerErr) {
		err := v.Setup(view, header)
		if err != nil {
			return fmt.Errorf("unable to setup debug controller: %w", err)
		}
	}
	return nil
}

func (v *Debug) RequestedSize(available int) *int {
	return nil
}
