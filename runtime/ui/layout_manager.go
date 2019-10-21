package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var lastY, lastX int

type LayoutManager struct {
	fileTreeSplitRatio float64
	controller         *Controller
}

// todo: this needs a major refactor (derive layout from view obj info, which should not live here)
func NewLayoutManager(c *Controller) *LayoutManager {

	fileTreeSplitRatio := viper.GetFloat64("filetree.pane-width")
	if fileTreeSplitRatio >= 1 || fileTreeSplitRatio <= 0 {
		logrus.Errorf("invalid config value: 'filetree.pane-width' should be 0 < value < 1, given '%v'", fileTreeSplitRatio)
		fileTreeSplitRatio = 0.5
	}

	return &LayoutManager{
		fileTreeSplitRatio: fileTreeSplitRatio,
		controller:         c,
	}
}

// isNewView determines if a view has already been created based on the set of errors given (a bit hokie)
func isNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if err != gocui.ErrUnknownView {
			return false
		}
	}
	return true
}

// layout defines the definition of the window pane size and placement relations to one another. This
// is invoked at application start and whenever the screen dimensions change.
func (lm *LayoutManager) Layout(g *gocui.Gui) error {
	// TODO: this logic should be refactored into an abstraction that takes care of the math for us

	maxX, maxY := g.Size()
	var resized bool
	if maxX != lastX {
		resized = true
	}
	if maxY != lastY {
		resized = true
	}
	lastX, lastY = maxX, maxY

	splitCols := int(float64(maxX) * (1.0 - lm.fileTreeSplitRatio))
	debugWidth := 0
	if debug {
		debugWidth = maxX / 4
	}
	debugCols := maxX - debugWidth
	bottomRows := 1
	headerRows := 2

	filterBarHeight := 1
	helpBarHeight := 1

	helpBarIndex := 1
	filterBarIndex := 2

	layersHeight := len(lm.controller.Layer.Layers) + headerRows + 1 // layers + header + base image layer row
	maxLayerHeight := int(0.75 * float64(maxY))
	if layersHeight > maxLayerHeight {
		layersHeight = maxLayerHeight
	}

	var view, header *gocui.View
	var viewErr, headerErr, err error

	if !lm.controller.Filter.IsVisible() {
		bottomRows--
		filterBarHeight = 0
	}

	// Debug pane
	if debug {
		if _, err := g.SetView("debug", debugCols, -1, maxX, maxY-bottomRows); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}

	// Layers
	view, viewErr = g.SetView(lm.controller.Layer.Name(), -1, -1+headerRows, splitCols, layersHeight)
	header, headerErr = g.SetView(lm.controller.Layer.Name()+"header", -1, -1, splitCols, headerRows)
	if isNewView(viewErr, headerErr) {
		err = lm.controller.Layer.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup layer controller", err)
			return err
		}

		if _, err = g.SetCurrentView(lm.controller.Layer.Name()); err != nil {
			logrus.Error("unable to set view to layer", err)
			return err
		}
		// since we are selecting the view, we should rerender to indicate it is selected
		err = lm.controller.Layer.Render()
		if err != nil {
			logrus.Error("unable to render layer view", err)
			return err
		}
	}

	// Details
	view, viewErr = g.SetView(lm.controller.Details.Name(), -1, -1+layersHeight+headerRows, splitCols, maxY-bottomRows)
	header, headerErr = g.SetView(lm.controller.Details.Name()+"header", -1, -1+layersHeight, splitCols, layersHeight+headerRows)
	if isNewView(viewErr, headerErr) {
		err = lm.controller.Details.Setup(view, header)
		if err != nil {
			return err
		}
	}

	// Filetree
	offset := 0
	if !lm.controller.Tree.AreAttributesVisible() {
		offset = 1
	}
	view, viewErr = g.SetView(lm.controller.Tree.Name(), splitCols, -1+headerRows-offset, debugCols, maxY-bottomRows)
	header, headerErr = g.SetView(lm.controller.Tree.Name()+"header", splitCols, -1, debugCols, headerRows-offset)
	if isNewView(viewErr, headerErr) {
		err = lm.controller.Tree.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup tree controller", err)
			return err
		}
	}
	err = lm.controller.Tree.OnLayoutChange(resized)
	if err != nil {
		logrus.Error("unable to setup layer controller onLayoutChange", err)
		return err
	}

	// Help Bar
	view, viewErr = g.SetView(lm.controller.Help.Name(), -1, maxY-helpBarHeight-helpBarIndex, maxX, maxY-(helpBarIndex-1))
	if isNewView(viewErr, headerErr) {
		err = lm.controller.Help.Setup(view, nil)
		if err != nil {
			logrus.Error("unable to setup help controller", err)
			return err
		}
	}

	// Filter Bar
	view, viewErr = g.SetView(lm.controller.Filter.Name(), len(lm.controller.Filter.HeaderStr())-1, maxY-filterBarHeight-filterBarIndex, maxX, maxY-(filterBarIndex-1))
	header, headerErr = g.SetView(lm.controller.Filter.Name()+"header", -1, maxY-filterBarHeight-filterBarIndex, len(lm.controller.Filter.HeaderStr()), maxY-(filterBarIndex-1))
	if isNewView(viewErr, headerErr) {
		err = lm.controller.Filter.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup filter controller", err)
			return err
		}
	}

	return nil
}
