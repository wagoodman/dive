package view

import (
	"fmt"
	"github.com/anchore/go-logger"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"github.com/wagoodman/dive/internal/log"

	"github.com/awesome-gocui/gocui"
	"github.com/wagoodman/dive/dive/filetree"
)

type ImageDetails struct {
	gui    *gocui.Gui
	body   *gocui.View
	header *gocui.View
	logger logger.Logger

	imageName      string
	imageSize      uint64
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
	kb             key.Bindings
}

func (v *ImageDetails) Name() string {
	return "imageDetails"
}

func (v *ImageDetails) Setup(body, header *gocui.View) error {
	v.logger = log.Nested("ui", "imageDetails")
	v.logger.Trace("Setup()")

	v.body = body
	v.body.Editable = false
	v.body.Wrap = true
	v.body.Highlight = true
	v.body.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = true
	v.header.Highlight = false
	v.header.Frame = false

	var infos = []key.BindingInfo{
		{
			Config:   v.kb.Navigation.Down,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			Config:   v.kb.Navigation.Up,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			Config:   v.kb.Navigation.PageUp,
			OnAction: v.PageUp,
		},
		{
			Config:   v.kb.Navigation.PageDown,
			OnAction: v.PageDown,
		},
	}

	_, err := key.GenerateBindings(v.gui, v.Name(), infos)
	if err != nil {
		return err
	}
	return nil
}

// Render flushes the state objects to the screen. The details pane reports:
// 1. the image efficiency score
// 2. the estimated wasted image space
// 3. a list of inefficient file allocations
func (v *ImageDetails) Render() error {
	body, err := renderImageDetailsBody(v.imageName, v.imageSize, v.efficiency, v.inefficiencies)
	if err != nil {
		return err
	}

	v.gui.Update(func(g *gocui.Gui) error {
		width, _ := v.body.Size()

		imageHeaderStr := format.RenderHeader("Image Details", width, v.gui.CurrentView() == v.body)

		v.header.Clear()
		_, err := fmt.Fprintln(v.header, imageHeaderStr)
		if err != nil {
			log.WithFields("error", err).Debug("unable to write to buffer")
		}
		v.body.Clear()
		_, err = fmt.Fprintln(v.body, body)
		if err != nil {
			log.WithFields("error", err).Debug("unable to write to buffer")
		}
		return err
	})

	return nil
}

func (v *ImageDetails) OnLayoutChange() error {
	if err := v.Update(); err != nil {
		return err
	}
	return v.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (v *ImageDetails) IsVisible() bool {
	return v.body != nil
}

func (v *ImageDetails) PageUp() error {
	_, height := v.body.Size()
	if err := CursorStep(v.gui, v.body, -height); err != nil {
		v.logger.WithFields("error", err).Debugf("couldn't move the cursor up by %d steps", height)
	}
	return nil
}

func (v *ImageDetails) PageDown() error {
	_, height := v.body.Size()
	if err := CursorStep(v.gui, v.body, height); err != nil {
		v.logger.WithFields("error", err).Debugf("couldn't move the cursor down by %d steps", height)
	}
	return nil
}

func (v *ImageDetails) CursorUp() error {
	if err := CursorUp(v.gui, v.body); err != nil {
		v.logger.WithFields("error", err).Debug("couldn't move the cursor up")
	}
	return nil
}

func (v *ImageDetails) CursorDown() error {
	if err := CursorDown(v.gui, v.body); err != nil {
		v.logger.WithFields("error", err).Debug("couldn't move the cursor down")
	}
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (v *ImageDetails) KeyHelp() string {
	return ""
}

// Update refreshes the state objects for future rendering.
func (v *ImageDetails) Update() error {
	return nil
}
