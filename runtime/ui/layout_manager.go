package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type layoutManager struct {
	fileTreeSplitRatio float64
	controllers        *Controller
}

// todo: this needs a major refactor (derive layout from view obj info, which should not live here)
func newLayoutManager(c *Controller) *layoutManager {

	fileTreeSplitRatio := viper.GetFloat64("filetree.pane-width")
	if fileTreeSplitRatio >= 1 || fileTreeSplitRatio <= 0 {
		logrus.Errorf("invalid config value: 'filetree.pane-width' should be 0 < value < 1, given '%v'", fileTreeSplitRatio)
		fileTreeSplitRatio = 0.5
	}

	return &layoutManager{
		fileTreeSplitRatio: fileTreeSplitRatio,
		controllers:        c,
	}
}

// IsNewView determines if a view has already been created based on the set of errors given (a bit hokie)
func IsNewView(errs ...error) bool {
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
func (lm *layoutManager) layout(g *gocui.Gui) error {
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
	statusBarHeight := 1

	statusBarIndex := 1
	filterBarIndex := 2

	layersHeight := len(lm.controllers.Layer.Layers) + headerRows + 1 // layers + header + base image layer row
	maxLayerHeight := int(0.75 * float64(maxY))
	if layersHeight > maxLayerHeight {
		layersHeight = maxLayerHeight
	}

	var view, header *gocui.View
	var viewErr, headerErr, err error

	if !lm.controllers.Filter.IsVisible() {
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
	view, viewErr = g.SetView(lm.controllers.Layer.Name(), -1, -1+headerRows, splitCols, layersHeight)
	header, headerErr = g.SetView(lm.controllers.Layer.Name()+"header", -1, -1, splitCols, headerRows)
	if IsNewView(viewErr, headerErr) {
		err = lm.controllers.Layer.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup layer controller", err)
			return err
		}

		if _, err = g.SetCurrentView(lm.controllers.Layer.Name()); err != nil {
			logrus.Error("unable to set view to layer", err)
			return err
		}
		// since we are selecting the view, we should rerender to indicate it is selected
		err = lm.controllers.Layer.Render()
		if err != nil {
			logrus.Error("unable to render layer view", err)
			return err
		}
	}

	// Details
	view, viewErr = g.SetView(lm.controllers.Details.Name(), -1, -1+layersHeight+headerRows, splitCols, maxY-bottomRows)
	header, headerErr = g.SetView(lm.controllers.Details.Name()+"header", -1, -1+layersHeight, splitCols, layersHeight+headerRows)
	if IsNewView(viewErr, headerErr) {
		err = lm.controllers.Details.Setup(view, header)
		if err != nil {
			return err
		}
	}

	// Filetree
	offset := 0
	if !lm.controllers.Tree.AreAttributesVisible() {
		offset = 1
	}
	view, viewErr = g.SetView(lm.controllers.Tree.Name(), splitCols, -1+headerRows-offset, debugCols, maxY-bottomRows)
	header, headerErr = g.SetView(lm.controllers.Tree.Name()+"header", splitCols, -1, debugCols, headerRows-offset)
	if IsNewView(viewErr, headerErr) {
		err = lm.controllers.Tree.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup tree controller", err)
			return err
		}
	}
	err = lm.controllers.Tree.OnLayoutChange(resized)
	if err != nil {
		logrus.Error("unable to setup layer controller onLayoutChange", err)
		return err
	}

	// Status Bar
	view, viewErr = g.SetView(lm.controllers.Status.Name(), -1, maxY-statusBarHeight-statusBarIndex, maxX, maxY-(statusBarIndex-1))
	if IsNewView(viewErr, headerErr) {
		err = lm.controllers.Status.Setup(view, nil)
		if err != nil {
			logrus.Error("unable to setup status controller", err)
			return err
		}
	}

	// Filter Bar
	view, viewErr = g.SetView(lm.controllers.Filter.Name(), len(lm.controllers.Filter.HeaderStr())-1, maxY-filterBarHeight-filterBarIndex, maxX, maxY-(filterBarIndex-1))
	header, headerErr = g.SetView(lm.controllers.Filter.Name()+"header", -1, maxY-filterBarHeight-filterBarIndex, len(lm.controllers.Filter.HeaderStr()), maxY-(filterBarIndex-1))
	if IsNewView(viewErr, headerErr) {
		err = lm.controllers.Filter.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup filter controller", err)
			return err
		}
	}

	return nil
}
