package compound

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/view"
	"github.com/wagoodman/dive/utils"
)

type LayerDetailsCompoundLayout struct {
	layer               *view.Layer
	details             *view.Details
	constrainRealEstate bool
}

func NewLayerDetailsCompoundLayout(layer *view.Layer, details *view.Details) *LayerDetailsCompoundLayout {
	return &LayerDetailsCompoundLayout{
		layer:   layer,
		details: details,
	}
}

func (cl *LayerDetailsCompoundLayout) Name() string {
	return "layer-details-compound-column"
}

// OnLayoutChange is called whenever the screen dimensions are changed
func (cl *LayerDetailsCompoundLayout) OnLayoutChange() error {
	err := cl.layer.OnLayoutChange()
	if err != nil {
		logrus.Error("unable to setup layer controller onLayoutChange", err)
		return err
	}

	err = cl.details.OnLayoutChange()
	if err != nil {
		logrus.Error("unable to setup details controller onLayoutChange", err)
		return err
	}
	return nil
}

func (cl *LayerDetailsCompoundLayout) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("view.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, cl.Name())

	////////////////////////////////////////////////////////////////////////////////////
	// Layers View

	// header + border
	layerHeaderHeight := 2

	layersHeight := cl.layer.LayerCount() + layerHeaderHeight + 1 // layers + header + base image layer row
	maxLayerHeight := int(0.75 * float64(maxY))
	if layersHeight > maxLayerHeight {
		layersHeight = maxLayerHeight
	}

	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(cl.layer.Name()+"header", minX, minY, maxX, minY+layerHeaderHeight+1)

	// we are going to overlap the view over the (invisible) border (so minY will be one less than expected)
	main, viewErr := g.SetView(cl.layer.Name(), minX, minY+layerHeaderHeight, maxX, minY+layerHeaderHeight+layersHeight)

	if utils.IsNewView(viewErr, headerErr) {
		err := cl.layer.Setup(main, header)
		if err != nil {
			logrus.Error("unable to setup layer layout", err)
			return err
		}

		if _, err = g.SetCurrentView(cl.layer.Name()); err != nil {
			logrus.Error("unable to set view to layer", err)
			return err
		}
	}

	////////////////////////////////////////////////////////////////////////////////////
	// Details
	detailsMinY := minY + layersHeight

	// header + border
	detailsHeaderHeight := 2

	v, _ := g.View(cl.details.Name())
	if v != nil {
		// the view exists already!

		// don't show the details pane when there isn't enough room on the screen
		if cl.constrainRealEstate {
			// take note: deleting a view will invoke layout again, so ensure this call is protected from an infinite loop
			err := g.DeleteView(cl.details.Name())
			if err != nil {
				return err
			}
			// take note: deleting a view will invoke layout again, so ensure this call is protected from an infinite loop
			err = g.DeleteView(cl.details.Name() + "header")
			if err != nil {
				return err
			}

			return nil
		}

	}

	header, headerErr = g.SetView(cl.details.Name()+"header", minX, detailsMinY, maxX, detailsMinY+detailsHeaderHeight)
	main, viewErr = g.SetView(cl.details.Name(), minX, detailsMinY+detailsHeaderHeight, maxX, maxY)

	if utils.IsNewView(viewErr, headerErr) {
		err := cl.details.Setup(main, header)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cl *LayerDetailsCompoundLayout) RequestedSize(available int) *int {
	// "available" is the entire screen real estate, so we can guess when its a bit too small and take action.
	// This isn't perfect, but it gets the job done for now without complicated layout constraint solvers
	if available < 90 {
		cl.layer.ConstrainLayout()
		cl.constrainRealEstate = true
		size := 8
		return &size
	}
	cl.layer.ExpandLayout()
	cl.constrainRealEstate = false
	return nil
}

// todo: make this variable based on the nested views
func (cl *LayerDetailsCompoundLayout) IsVisible() bool {
	return true
}
