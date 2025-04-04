package view

import (
	"fmt"
	"github.com/anchore/go-logger"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"github.com/wagoodman/dive/internal/log"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/image"
)

type LayerDetails struct {
	gui          *gocui.Gui
	header       *gocui.View
	body         *gocui.View
	CurrentLayer *image.Layer
	kb           key.Bindings
	logger       logger.Logger
}

func (v *LayerDetails) Name() string {
	return "layerDetails"
}

func (v *LayerDetails) Setup(body, header *gocui.View) error {
	v.logger = log.Nested("ui", "layerDetails")
	v.logger.Trace("setup()")

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
	}

	_, err := key.GenerateBindings(v.gui, v.Name(), infos)
	if err != nil {
		return err
	}
	return nil
}

// Render flushes the state objects to the screen.
// The details pane reports the currently selected layer's:
// 1. tags
// 2. ID
// 3. digest
// 4. command
func (v *LayerDetails) Render() error {
	v.gui.Update(func(g *gocui.Gui) error {
		v.header.Clear()
		width, _ := v.body.Size()

		layerHeaderStr := format.RenderHeader("Layer Details", width, v.gui.CurrentView() == v.body)

		_, err := fmt.Fprintln(v.header, layerHeaderStr)
		if err != nil {
			return err
		}

		// this is for layer details
		var lines = make([]string, 0)

		tags := "(none)"
		if len(v.CurrentLayer.Names) > 0 {
			tags = strings.Join(v.CurrentLayer.Names, ", ")
		}

		lines = append(lines, []string{
			format.Header("Tags:   ") + tags,
			format.Header("Id:     ") + v.CurrentLayer.Id,
			format.Header("Size:   ") + humanize.Bytes(v.CurrentLayer.Size),
			format.Header("Digest: ") + v.CurrentLayer.Digest,
			format.Header("Command:"),
			v.CurrentLayer.Command,
		}...)

		v.body.Clear()
		if _, err = fmt.Fprintln(v.body, strings.Join(lines, "\n")); err != nil {
			log.WithFields("layer", v.CurrentLayer.Id, "error", err).Debug("unable to write to buffer")
		}
		return nil
	})
	return nil
}

func (v *LayerDetails) OnLayoutChange() error {
	if err := v.Update(); err != nil {
		return err
	}
	return v.Render()
}

// IsVisible indicates if the details view pane is currently initialized.
func (v *LayerDetails) IsVisible() bool {
	return v.body != nil
}

// CursorUp moves the cursor up in the details pane
func (v *LayerDetails) CursorUp() error {
	if err := CursorUp(v.gui, v.body); err != nil {
		v.logger.WithFields("error", err).Debug("couldn't move the cursor up")
	}
	return nil
}

// CursorDown moves the cursor up in the details pane
func (v *LayerDetails) CursorDown() error {
	if err := CursorDown(v.gui, v.body); err != nil {
		v.logger.WithFields("error", err).Debug("couldn't move the cursor down")
	}
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (v *LayerDetails) KeyHelp() string {
	return ""
}

// Update refreshes the state objects for future rendering.
func (v *LayerDetails) Update() error {
	return nil
}

func (v *LayerDetails) SetCursor(x, y int) error {
	return v.body.SetCursor(x, y)
}
