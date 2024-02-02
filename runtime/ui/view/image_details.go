package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
)

type ImageDetails struct {
	gui            *gocui.Gui
	body           *gocui.View
	header         *gocui.View
	imageName      string
	imageSize      uint64
	efficiency     float64
	inefficiencies filetree.EfficiencySlice
}

func (v *ImageDetails) Name() string {
	return "imageDetails"
}

func (v *ImageDetails) Setup(body, header *gocui.View) error {
	logrus.Tracef("ImageDetails setup()")
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
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   v.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   v.PageDown,
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
	analysisTemplate := "%5s  %12s  %-s\n"
	inefficiencyReport := fmt.Sprintf(format.Header(analysisTemplate), "Count", "Total Space", "Path")

	var wastedSpace int64
	for idx := 0; idx < len(v.inefficiencies); idx++ {
		data := v.inefficiencies[len(v.inefficiencies)-1-idx]
		wastedSpace += data.CumulativeSize

		inefficiencyReport += fmt.Sprintf(analysisTemplate, strconv.Itoa(len(data.Nodes)), humanize.Bytes(uint64(data.CumulativeSize)), data.Path)
	}

	imageNameStr := fmt.Sprintf("%s %s", format.Header("Image name:"), v.imageName)
	imageSizeStr := fmt.Sprintf("%s %s", format.Header("Total Image size:"), humanize.Bytes(v.imageSize))
	efficiencyStr := fmt.Sprintf("%s %d %%", format.Header("Image efficiency score:"), int(100.0*v.efficiency))
	wastedSpaceStr := fmt.Sprintf("%s %s", format.Header("Potential wasted space:"), humanize.Bytes(uint64(wastedSpace)))

	v.gui.Update(func(g *gocui.Gui) error {
		width, _ := v.body.Size()

		imageHeaderStr := format.RenderHeader("Image Details", width, v.gui.CurrentView() == v.body)

		v.header.Clear()
		_, err := fmt.Fprintln(v.header, imageHeaderStr)
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
		}

		var lines = []string{
			imageNameStr,
			imageSizeStr,
			wastedSpaceStr,
			efficiencyStr,
			" ", // to avoid an empty line so CursorDown can work as expected
			inefficiencyReport,
		}

		v.body.Clear()
		_, err = fmt.Fprintln(v.body, strings.Join(lines, "\n"))
		if err != nil {
			logrus.Debug("unable to write to buffer: ", err)
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
	return CursorUp(v.body, uint(height))
}

func (v *ImageDetails) PageDown() error {
	_, height := v.body.Size()
	return CursorDown(v.body, uint(height))
}

func (v *ImageDetails) CursorUp() error {
	return CursorUp(v.body, 1)
}

func (v *ImageDetails) CursorDown() error {
	return CursorDown(v.body, 1)
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected (currently does nothing).
func (v *ImageDetails) KeyHelp() string {
	return ""
}

// Update refreshes the state objects for future rendering.
func (v *ImageDetails) Update() error {
	return nil
}
