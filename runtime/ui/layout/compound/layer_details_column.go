package compound

import (
	"github.com/awesome-gocui/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/view"
	"github.com/wagoodman/dive/utils"
)

type LayerDetailsCompoundLayout struct {
	layer               *view.Layer
	layerDetails        *view.LayerDetails
	imageDetails        *view.ImageDetails
	constrainRealEstate bool
}

func NewLayerDetailsCompoundLayout(layer *view.Layer, layerDetails *view.LayerDetails, imageDetails *view.ImageDetails) *LayerDetailsCompoundLayout {
	return &LayerDetailsCompoundLayout{
		layer:        layer,
		layerDetails: layerDetails,
		imageDetails: imageDetails,
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

	err = cl.layerDetails.OnLayoutChange()
	if err != nil {
		logrus.Error("unable to setup layer details controller onLayoutChange", err)
		return err
	}

	err = cl.imageDetails.OnLayoutChange()
	if err != nil {
		logrus.Error("unable to setup image details controller onLayoutChange", err)
		return err
	}
	return nil
}

func (cl *LayerDetailsCompoundLayout) layoutRow(g *gocui.Gui, minX, minY, maxX, maxY int, viewName string, setup func(*gocui.View, *gocui.View) error) error {
	logrus.Tracef("layoutRow(g, minX: %d, minY: %d, maxX: %d, maxY: %d, viewName: %s, <setup func>)", minX, minY, maxX, maxY, viewName)
	// header + border
	headerHeight := 2

	// TODO: investigate overlap
	// note: maxY needs to account for the (invisible) border, thus a +1
	headerView, headerErr := g.SetView(viewName+"Header", minX, minY, maxX, minY+headerHeight+1, 0)

	// we are going to overlap the view over the (invisible) border (so minY will be one less than expected)
	bodyView, bodyErr := g.SetView(viewName, minX, minY+headerHeight, maxX, maxY, 0)

	if utils.IsNewView(bodyErr, headerErr) {
		err := setup(bodyView, headerView)
		if err != nil {
			logrus.Debug("unable to setup row layout for ", viewName, err)
			return err
		}
	}
	return nil
}

func (cl *LayerDetailsCompoundLayout) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("LayerDetailsCompountLayout.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, cl.Name())

	layouts := []view.IView{
		cl.layer,
		cl.layerDetails,
		cl.imageDetails,
	}

	rowHeight := maxY / 3
	for i := 0; i < 3; i++ {
		if err := cl.layoutRow(g, minX, i*rowHeight, maxX, (i+1)*rowHeight, layouts[i].Name(), layouts[i].Setup); err != nil {
			logrus.Debug("Laying out layers view errored!")
			return err
		}
	}

	if g.CurrentView() == nil {
		if _, err := g.SetCurrentView(cl.layer.Name()); err != nil {
			logrus.Error("unable to set view to layer", err)
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
